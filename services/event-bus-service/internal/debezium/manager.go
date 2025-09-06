// Package debezium provides Debezium Change Data Capture integration for the Event Bus Service
// This package implements enterprise-grade CDC functionality using Debezium Connect
// with comprehensive monitoring, health checks, and connector management.
package debezium

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/Mir00r/X-Form-Backend/services/event-bus-service/internal/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

// Manager manages Debezium connectors and Change Data Capture operations
// It provides enterprise-grade CDC functionality with monitoring and health checks
type Manager struct {
	config     *config.Config
	logger     *zap.Logger
	httpClient *http.Client
	connectors map[string]*ConnectorStatus
	mutex      sync.RWMutex
	metrics    *DebeziumMetrics
	stopCh     chan struct{}
	wg         sync.WaitGroup
}

// ConnectorStatus represents the status of a Debezium connector
type ConnectorStatus struct {
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	State        string                 `json:"state"`
	WorkerID     string                 `json:"worker_id"`
	Config       map[string]interface{} `json:"config"`
	Tasks        []TaskStatus           `json:"tasks"`
	LastUpdated  time.Time              `json:"last_updated"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	RestartCount int                    `json:"restart_count"`
	HealthScore  float64                `json:"health_score"`
}

// TaskStatus represents the status of a connector task
type TaskStatus struct {
	ID       int    `json:"id"`
	State    string `json:"state"`
	WorkerID string `json:"worker_id"`
	Trace    string `json:"trace,omitempty"`
}

// ConnectorConfig represents Debezium connector configuration
type ConnectorConfig struct {
	Name   string            `json:"name"`
	Config map[string]string `json:"config"`
}

// DebeziumMetrics contains Prometheus metrics for Debezium operations
type DebeziumMetrics struct {
	ConnectorsTotal       prometheus.Gauge
	ConnectorsRunning     prometheus.Gauge
	ConnectorsFailed      prometheus.Gauge
	ConnectorRestarts     prometheus.Counter
	SourceRecordsPolled   prometheus.Counter
	SourceRecordsFiltered prometheus.Counter
	TasksTotal            prometheus.Gauge
	TasksRunning          prometheus.Gauge
	TasksFailed           prometheus.Gauge
	OffsetCommits         prometheus.Counter
	OffsetCommitLatency   prometheus.Histogram
	HealthCheckDuration   prometheus.Histogram
	APIResponseTime       prometheus.Histogram
}

// PostgresConnectorConfig represents PostgreSQL-specific connector configuration
type PostgresConnectorConfig struct {
	ConnectorClass             string `json:"connector.class"`
	DatabaseHostname           string `json:"database.hostname"`
	DatabasePort               string `json:"database.port"`
	DatabaseUser               string `json:"database.user"`
	DatabasePassword           string `json:"database.password"`
	DatabaseDBName             string `json:"database.dbname"`
	DatabaseServerName         string `json:"database.server.name"`
	TableIncludeList           string `json:"table.include.list,omitempty"`
	TableExcludeList           string `json:"table.exclude.list,omitempty"`
	PluginName                 string `json:"plugin.name"`
	SlotName                   string `json:"slot.name"`
	PublicationName            string `json:"publication.name,omitempty"`
	TopicPrefix                string `json:"topic.prefix"`
	KeyConverter               string `json:"key.converter"`
	ValueConverter             string `json:"value.converter"`
	KeyConverterSchemas        string `json:"key.converter.schemas.enable"`
	ValueConverterSchemas      string `json:"value.converter.schemas.enable"`
	IncludeSchemaChanges       string `json:"include.schema.changes"`
	ProvideTransactionMetadata string `json:"provide.transaction.metadata"`
	SnapshotMode               string `json:"snapshot.mode"`
	HeartbeatIntervalMs        string `json:"heartbeat.interval.ms"`
	HeartbeatTopicsPrefix      string `json:"heartbeat.topics.prefix"`
	TransformationsRoute       string `json:"transforms,omitempty"`
	SMTClass                   string `json:"transforms.route.type,omitempty"`
	SMTTopicRegex              string `json:"transforms.route.topic.regex,omitempty"`
	SMTTopicReplacement        string `json:"transforms.route.topic.replacement,omitempty"`
}

// NewManager creates a new Debezium manager instance
func NewManager(cfg *config.Config, logger *zap.Logger) (*Manager, error) {
	if logger == nil {
		logger = zap.NewNop()
	}

	// Create HTTP client with timeouts and security configuration
	httpClient := createHTTPClient(cfg)

	manager := &Manager{
		config:     cfg,
		logger:     logger,
		httpClient: httpClient,
		connectors: make(map[string]*ConnectorStatus),
		metrics:    initDebeziumMetrics(),
		stopCh:     make(chan struct{}),
	}

	// Test connectivity to Debezium Connect
	if err := manager.testConnectivity(); err != nil {
		return nil, fmt.Errorf("failed to connect to Debezium Connect: %w", err)
	}

	logger.Info("Debezium manager initialized successfully",
		zap.String("connect_url", cfg.Debezium.Connect.URL))

	return manager, nil
}

// createHTTPClient creates an HTTP client with proper configuration
func createHTTPClient(cfg *config.Config) *http.Client {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		DisableKeepAlives:   false,
	}

	// Configure TLS if enabled
	if cfg.Debezium.Connect.TLSConfig.Enabled {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false, // Always verify in production
		}

		if cfg.Debezium.Connect.TLSConfig.CertFile != "" && cfg.Debezium.Connect.TLSConfig.KeyFile != "" {
			cert, err := tls.LoadX509KeyPair(cfg.Debezium.Connect.TLSConfig.CertFile, cfg.Debezium.Connect.TLSConfig.KeyFile)
			if err == nil {
				tlsConfig.Certificates = []tls.Certificate{cert}
			}
		}

		transport.TLSClientConfig = tlsConfig
	}

	return &http.Client{
		Transport: transport,
		Timeout:   cfg.Debezium.Connect.Timeout,
	}
}

// Start starts the Debezium manager and begins monitoring connectors
func (m *Manager) Start(ctx context.Context) error {
	if !m.config.Debezium.Enabled {
		m.logger.Info("Debezium is disabled, skipping startup")
		return nil
	}

	m.logger.Info("Starting Debezium manager")

	// Initialize configured connectors
	if err := m.initializeConnectors(ctx); err != nil {
		return fmt.Errorf("failed to initialize connectors: %w", err)
	}

	// Start monitoring goroutine
	if m.config.Debezium.Monitoring.Enabled {
		m.wg.Add(1)
		go m.monitorConnectors(ctx)
	}

	// Start health check goroutine
	m.wg.Add(1)
	go m.healthCheckLoop(ctx)

	return nil
}

// Stop stops the Debezium manager and all monitoring
func (m *Manager) Stop() error {
	m.logger.Info("Stopping Debezium manager")

	close(m.stopCh)
	m.wg.Wait()

	m.logger.Info("Debezium manager stopped")
	return nil
}

// CreateConnector creates a new Debezium connector
func (m *Manager) CreateConnector(ctx context.Context, connectorConfig *ConnectorConfig) error {
	start := time.Now()
	defer func() {
		m.metrics.APIResponseTime.Observe(time.Since(start).Seconds())
	}()

	// Validate connector configuration
	if err := m.validateConnectorConfig(connectorConfig); err != nil {
		return fmt.Errorf("invalid connector configuration: %w", err)
	}

	// Prepare request
	jsonData, err := json.Marshal(connectorConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal connector config: %w", err)
	}

	url := fmt.Sprintf("%s/connectors", m.config.Debezium.Connect.URL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	m.setAuthHeaders(req)

	// Execute request
	resp, err := m.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create connector: %w", err)
	}
	defer resp.Body.Close()

	// Handle response
	if resp.StatusCode == http.StatusConflict {
		return fmt.Errorf("connector %s already exists", connectorConfig.Name)
	}

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create connector, status: %d, body: %s", resp.StatusCode, string(body))
	}

	m.logger.Info("Connector created successfully",
		zap.String("connector", connectorConfig.Name))

	// Update local status
	m.mutex.Lock()
	m.connectors[connectorConfig.Name] = &ConnectorStatus{
		Name:        connectorConfig.Name,
		Type:        m.getConnectorType(connectorConfig.Config),
		State:       "RUNNING",
		Config:      convertStringMapToInterface(connectorConfig.Config),
		LastUpdated: time.Now(),
		HealthScore: 1.0,
	}
	m.mutex.Unlock()

	m.updateMetrics()
	return nil
}

// DeleteConnector deletes a Debezium connector
func (m *Manager) DeleteConnector(ctx context.Context, connectorName string) error {
	start := time.Now()
	defer func() {
		m.metrics.APIResponseTime.Observe(time.Since(start).Seconds())
	}()

	url := fmt.Sprintf("%s/connectors/%s", m.config.Debezium.Connect.URL, connectorName)
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	m.setAuthHeaders(req)

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete connector: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("connector %s not found", connectorName)
	}

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete connector, status: %d, body: %s", resp.StatusCode, string(body))
	}

	m.logger.Info("Connector deleted successfully",
		zap.String("connector", connectorName))

	// Remove from local status
	m.mutex.Lock()
	delete(m.connectors, connectorName)
	m.mutex.Unlock()

	m.updateMetrics()
	return nil
}

// RestartConnector restarts a Debezium connector
func (m *Manager) RestartConnector(ctx context.Context, connectorName string) error {
	start := time.Now()
	defer func() {
		m.metrics.APIResponseTime.Observe(time.Since(start).Seconds())
	}()

	url := fmt.Sprintf("%s/connectors/%s/restart", m.config.Debezium.Connect.URL, connectorName)
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	m.setAuthHeaders(req)

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to restart connector: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("connector %s not found", connectorName)
	}

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to restart connector, status: %d, body: %s", resp.StatusCode, string(body))
	}

	m.logger.Info("Connector restarted successfully",
		zap.String("connector", connectorName))

	// Update restart count
	m.mutex.Lock()
	if status, exists := m.connectors[connectorName]; exists {
		status.RestartCount++
		status.LastUpdated = time.Now()
	}
	m.mutex.Unlock()

	m.metrics.ConnectorRestarts.Inc()
	return nil
}

// GetConnectorStatus returns the status of a specific connector
func (m *Manager) GetConnectorStatus(ctx context.Context, connectorName string) (*ConnectorStatus, error) {
	start := time.Now()
	defer func() {
		m.metrics.APIResponseTime.Observe(time.Since(start).Seconds())
	}()

	url := fmt.Sprintf("%s/connectors/%s/status", m.config.Debezium.Connect.URL, connectorName)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	m.setAuthHeaders(req)

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get connector status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("connector %s not found", connectorName)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get connector status, status: %d, body: %s", resp.StatusCode, string(body))
	}

	var status ConnectorStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	status.LastUpdated = time.Now()
	status.HealthScore = m.calculateHealthScore(&status)

	// Update local cache
	m.mutex.Lock()
	m.connectors[connectorName] = &status
	m.mutex.Unlock()

	return &status, nil
}

// ListConnectors returns a list of all connectors
func (m *Manager) ListConnectors(ctx context.Context) ([]string, error) {
	start := time.Now()
	defer func() {
		m.metrics.APIResponseTime.Observe(time.Since(start).Seconds())
	}()

	url := fmt.Sprintf("%s/connectors", m.config.Debezium.Connect.URL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	m.setAuthHeaders(req)

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list connectors: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to list connectors, status: %d, body: %s", resp.StatusCode, string(body))
	}

	var connectors []string
	if err := json.NewDecoder(resp.Body).Decode(&connectors); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return connectors, nil
}

// CreatePostgreSQLConnector creates a PostgreSQL CDC connector with optimized configuration
func (m *Manager) CreatePostgreSQLConnector(ctx context.Context, dbConfig config.DatabaseConfig, topicPrefix string) error {
	connectorName := fmt.Sprintf("%s-postgres-connector", topicPrefix)

	// Build PostgreSQL connector configuration
	pgConfig := &PostgresConnectorConfig{
		ConnectorClass:             "io.debezium.connector.postgresql.PostgresConnector",
		DatabaseHostname:           dbConfig.Host,
		DatabasePort:               fmt.Sprintf("%d", dbConfig.Port),
		DatabaseUser:               dbConfig.Username,
		DatabasePassword:           dbConfig.Password,
		DatabaseDBName:             dbConfig.Name,
		DatabaseServerName:         topicPrefix,
		PluginName:                 "pgoutput",
		SlotName:                   fmt.Sprintf("%s_slot", topicPrefix),
		TopicPrefix:                topicPrefix,
		KeyConverter:               "org.apache.kafka.connect.json.JsonConverter",
		ValueConverter:             "org.apache.kafka.connect.json.JsonConverter",
		KeyConverterSchemas:        "false",
		ValueConverterSchemas:      "false",
		IncludeSchemaChanges:       "true",
		ProvideTransactionMetadata: "true",
		SnapshotMode:               "initial",
		HeartbeatIntervalMs:        "60000",
		HeartbeatTopicsPrefix:      fmt.Sprintf("%s.heartbeat", topicPrefix),
	}

	// Apply table filtering if specified
	if len(m.config.Debezium.Connectors) > 0 {
		for _, connector := range m.config.Debezium.Connectors {
			if connector.Database.Name == dbConfig.Name {
				if includeList, ok := connector.Config["table.include.list"]; ok {
					pgConfig.TableIncludeList = includeList
				}
				if excludeList, ok := connector.Config["table.exclude.list"]; ok {
					pgConfig.TableExcludeList = excludeList
				}
				break
			}
		}
	}

	// Convert to map for API call
	configMap := m.structToMap(pgConfig)

	connectorConfig := &ConnectorConfig{
		Name:   connectorName,
		Config: configMap,
	}

	return m.CreateConnector(ctx, connectorConfig)
}

// HealthCheck performs a comprehensive health check on Debezium Connect
func (m *Manager) HealthCheck(ctx context.Context) error {
	start := time.Now()
	defer func() {
		m.metrics.HealthCheckDuration.Observe(time.Since(start).Seconds())
	}()

	// Test basic connectivity
	if err := m.testConnectivity(); err != nil {
		return fmt.Errorf("connectivity check failed: %w", err)
	}

	// Check all configured connectors
	connectors, err := m.ListConnectors(ctx)
	if err != nil {
		return fmt.Errorf("failed to list connectors: %w", err)
	}

	failedConnectors := 0
	for _, connector := range connectors {
		status, err := m.GetConnectorStatus(ctx, connector)
		if err != nil {
			m.logger.Warn("Failed to get connector status",
				zap.String("connector", connector),
				zap.Error(err))
			failedConnectors++
			continue
		}

		if status.State != "RUNNING" {
			m.logger.Warn("Connector is not running",
				zap.String("connector", connector),
				zap.String("state", status.State))
			failedConnectors++
		}
	}

	if failedConnectors > 0 {
		return fmt.Errorf("%d out of %d connectors are not healthy", failedConnectors, len(connectors))
	}

	return nil
}

// testConnectivity tests basic connectivity to Debezium Connect
func (m *Manager) testConnectivity() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s/", m.config.Debezium.Connect.URL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	m.setAuthHeaders(req)

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// initializeConnectors initializes all configured connectors
func (m *Manager) initializeConnectors(ctx context.Context) error {
	for _, connectorConfig := range m.config.Debezium.Connectors {
		m.logger.Info("Initializing connector",
			zap.String("name", connectorConfig.Name),
			zap.String("type", connectorConfig.Type))

		switch connectorConfig.Type {
		case "postgres", "postgresql":
			if err := m.CreatePostgreSQLConnector(ctx, connectorConfig.Database, connectorConfig.Topics.Prefix); err != nil {
				m.logger.Error("Failed to create PostgreSQL connector",
					zap.String("name", connectorConfig.Name),
					zap.Error(err))
				// Don't fail startup for individual connector failures
				continue
			}
		default:
			m.logger.Warn("Unsupported connector type",
				zap.String("type", connectorConfig.Type),
				zap.String("name", connectorConfig.Name))
		}
	}

	return nil
}

// monitorConnectors periodically monitors connector health and status
func (m *Manager) monitorConnectors(ctx context.Context) {
	defer m.wg.Done()

	ticker := time.NewTicker(m.config.Debezium.Monitoring.HealthInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.performHealthCheck(ctx)
		}
	}
}

// healthCheckLoop performs periodic health checks on Debezium Connect
func (m *Manager) healthCheckLoop(ctx context.Context) {
	defer m.wg.Done()

	ticker := time.NewTicker(30 * time.Second) // Health check every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopCh:
			return
		case <-ticker.C:
			if err := m.HealthCheck(ctx); err != nil {
				m.logger.Error("Debezium health check failed", zap.Error(err))
			}
		}
	}
}

// performHealthCheck performs health check on all connectors
func (m *Manager) performHealthCheck(ctx context.Context) {
	connectors, err := m.ListConnectors(ctx)
	if err != nil {
		m.logger.Error("Failed to list connectors for health check", zap.Error(err))
		return
	}

	for _, connector := range connectors {
		status, err := m.GetConnectorStatus(ctx, connector)
		if err != nil {
			m.logger.Error("Failed to get connector status",
				zap.String("connector", connector),
				zap.Error(err))
			continue
		}

		// Log unhealthy connectors
		if status.State != "RUNNING" {
			m.logger.Warn("Connector is not running",
				zap.String("connector", connector),
				zap.String("state", status.State),
				zap.String("error", status.ErrorMessage))

			// Attempt restart if configured
			if status.RestartCount < 3 { // Limit restart attempts
				m.logger.Info("Attempting to restart connector",
					zap.String("connector", connector))
				if err := m.RestartConnector(ctx, connector); err != nil {
					m.logger.Error("Failed to restart connector",
						zap.String("connector", connector),
						zap.Error(err))
				}
			}
		}
	}

	m.updateMetrics()
}

// updateMetrics updates Prometheus metrics based on current connector status
func (m *Manager) updateMetrics() {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	totalConnectors := len(m.connectors)
	runningConnectors := 0
	failedConnectors := 0
	totalTasks := 0
	runningTasks := 0
	failedTasks := 0

	for _, status := range m.connectors {
		if status.State == "RUNNING" {
			runningConnectors++
		} else if status.State == "FAILED" {
			failedConnectors++
		}

		totalTasks += len(status.Tasks)
		for _, task := range status.Tasks {
			if task.State == "RUNNING" {
				runningTasks++
			} else if task.State == "FAILED" {
				failedTasks++
			}
		}
	}

	m.metrics.ConnectorsTotal.Set(float64(totalConnectors))
	m.metrics.ConnectorsRunning.Set(float64(runningConnectors))
	m.metrics.ConnectorsFailed.Set(float64(failedConnectors))
	m.metrics.TasksTotal.Set(float64(totalTasks))
	m.metrics.TasksRunning.Set(float64(runningTasks))
	m.metrics.TasksFailed.Set(float64(failedTasks))
}

// Helper methods

// setAuthHeaders sets authentication headers if configured
func (m *Manager) setAuthHeaders(req *http.Request) {
	if m.config.Debezium.Connect.Username != "" {
		req.SetBasicAuth(m.config.Debezium.Connect.Username, m.config.Debezium.Connect.Password)
	}
}

// validateConnectorConfig validates connector configuration
func (m *Manager) validateConnectorConfig(config *ConnectorConfig) error {
	if config.Name == "" {
		return fmt.Errorf("connector name is required")
	}

	if config.Config == nil || len(config.Config) == 0 {
		return fmt.Errorf("connector config is required")
	}

	if _, exists := config.Config["connector.class"]; !exists {
		return fmt.Errorf("connector.class is required")
	}

	return nil
}

// getConnectorType extracts connector type from configuration
func (m *Manager) getConnectorType(config map[string]string) string {
	connectorClass, exists := config["connector.class"]
	if !exists {
		return "unknown"
	}

	if contains(connectorClass, "postgresql") {
		return "postgresql"
	} else if contains(connectorClass, "mysql") {
		return "mysql"
	} else if contains(connectorClass, "mongodb") {
		return "mongodb"
	}

	return "unknown"
}

// calculateHealthScore calculates a health score for a connector
func (m *Manager) calculateHealthScore(status *ConnectorStatus) float64 {
	score := 1.0

	// Reduce score for non-running state
	if status.State != "RUNNING" {
		score -= 0.5
	}

	// Reduce score for failed tasks
	if len(status.Tasks) > 0 {
		failedTasks := 0
		for _, task := range status.Tasks {
			if task.State == "FAILED" {
				failedTasks++
			}
		}
		if failedTasks > 0 {
			score -= float64(failedTasks) / float64(len(status.Tasks)) * 0.3
		}
	}

	// Reduce score for multiple restarts
	if status.RestartCount > 0 {
		score -= float64(status.RestartCount) * 0.05
	}

	if score < 0 {
		score = 0
	}

	return score
}

// structToMap converts a struct to a map[string]string
func (m *Manager) structToMap(s interface{}) map[string]string {
	result := make(map[string]string)

	// This is a simplified implementation
	// In production, use reflection or a proper conversion library
	if pgConfig, ok := s.(*PostgresConnectorConfig); ok {
		result["connector.class"] = pgConfig.ConnectorClass
		result["database.hostname"] = pgConfig.DatabaseHostname
		result["database.port"] = pgConfig.DatabasePort
		result["database.user"] = pgConfig.DatabaseUser
		result["database.password"] = pgConfig.DatabasePassword
		result["database.dbname"] = pgConfig.DatabaseDBName
		result["database.server.name"] = pgConfig.DatabaseServerName
		result["plugin.name"] = pgConfig.PluginName
		result["slot.name"] = pgConfig.SlotName
		result["topic.prefix"] = pgConfig.TopicPrefix
		result["key.converter"] = pgConfig.KeyConverter
		result["value.converter"] = pgConfig.ValueConverter
		result["key.converter.schemas.enable"] = pgConfig.KeyConverterSchemas
		result["value.converter.schemas.enable"] = pgConfig.ValueConverterSchemas
		result["include.schema.changes"] = pgConfig.IncludeSchemaChanges
		result["provide.transaction.metadata"] = pgConfig.ProvideTransactionMetadata
		result["snapshot.mode"] = pgConfig.SnapshotMode
		result["heartbeat.interval.ms"] = pgConfig.HeartbeatIntervalMs
		result["heartbeat.topics.prefix"] = pgConfig.HeartbeatTopicsPrefix

		if pgConfig.TableIncludeList != "" {
			result["table.include.list"] = pgConfig.TableIncludeList
		}
		if pgConfig.TableExcludeList != "" {
			result["table.exclude.list"] = pgConfig.TableExcludeList
		}
		if pgConfig.PublicationName != "" {
			result["publication.name"] = pgConfig.PublicationName
		}
	}

	return result
}

// convertStringMapToInterface converts map[string]string to map[string]interface{}
func convertStringMapToInterface(m map[string]string) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range m {
		result[k] = v
	}
	return result
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			(len(s) > len(substr) && s[1:len(substr)+1] == substr))))
}

// initDebeziumMetrics initializes Prometheus metrics for Debezium operations
func initDebeziumMetrics() *DebeziumMetrics {
	return &DebeziumMetrics{
		ConnectorsTotal: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "debezium_connectors_total",
			Help: "Total number of Debezium connectors",
		}),
		ConnectorsRunning: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "debezium_connectors_running",
			Help: "Number of running Debezium connectors",
		}),
		ConnectorsFailed: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "debezium_connectors_failed",
			Help: "Number of failed Debezium connectors",
		}),
		ConnectorRestarts: promauto.NewCounter(prometheus.CounterOpts{
			Name: "debezium_connector_restarts_total",
			Help: "Total number of connector restarts",
		}),
		SourceRecordsPolled: promauto.NewCounter(prometheus.CounterOpts{
			Name: "debezium_source_records_polled_total",
			Help: "Total number of source records polled",
		}),
		SourceRecordsFiltered: promauto.NewCounter(prometheus.CounterOpts{
			Name: "debezium_source_records_filtered_total",
			Help: "Total number of source records filtered",
		}),
		TasksTotal: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "debezium_tasks_total",
			Help: "Total number of Debezium tasks",
		}),
		TasksRunning: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "debezium_tasks_running",
			Help: "Number of running Debezium tasks",
		}),
		TasksFailed: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "debezium_tasks_failed",
			Help: "Number of failed Debezium tasks",
		}),
		OffsetCommits: promauto.NewCounter(prometheus.CounterOpts{
			Name: "debezium_offset_commits_total",
			Help: "Total number of offset commits",
		}),
		OffsetCommitLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "debezium_offset_commit_latency_seconds",
			Help:    "Histogram of offset commit latencies",
			Buckets: prometheus.DefBuckets,
		}),
		HealthCheckDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "debezium_health_check_duration_seconds",
			Help:    "Duration of Debezium health checks",
			Buckets: prometheus.DefBuckets,
		}),
		APIResponseTime: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "debezium_api_response_time_seconds",
			Help:    "Response time for Debezium Connect API calls",
			Buckets: prometheus.DefBuckets,
		}),
	}
}
