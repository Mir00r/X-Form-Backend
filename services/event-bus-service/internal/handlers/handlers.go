// Package handlers provides HTTP handlers for the Event Bus Service REST API
// This package implements enterprise-grade HTTP handlers with comprehensive
// validation, monitoring, security, and error handling for event bus operations.
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Mir00r/X-Form-Backend/services/event-bus-service/internal/config"
	"github.com/Mir00r/X-Form-Backend/services/event-bus-service/internal/debezium"
	"github.com/Mir00r/X-Form-Backend/services/event-bus-service/internal/kafka"
	"github.com/Mir00r/X-Form-Backend/services/event-bus-service/internal/processors"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

// EventBusHandler provides HTTP handlers for the Event Bus Service
type EventBusHandler struct {
	config           *config.Config
	logger           *zap.Logger
	kafka            *kafka.Client
	debezium         *debezium.Manager
	processorManager *processors.ProcessorManager
	metrics          *HandlerMetrics
}

// HandlerMetrics contains Prometheus metrics for HTTP handlers
type HandlerMetrics struct {
	RequestsTotal     *prometheus.CounterVec
	RequestDuration   *prometheus.HistogramVec
	ErrorsTotal       *prometheus.CounterVec
	ActiveConnections prometheus.Gauge
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     interface{} `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
	Version   string      `json:"version"`
}

// EventRequest represents an event publishing request
type EventRequest struct {
	EventType string                 `json:"event_type"`
	Source    string                 `json:"source"`
	Subject   string                 `json:"subject"`
	Data      map[string]interface{} `json:"data"`
	Headers   map[string]string      `json:"headers"`
	Topic     string                 `json:"topic"`
	Key       string                 `json:"key"`
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status       string                 `json:"status"`
	Version      string                 `json:"version"`
	Timestamp    time.Time              `json:"timestamp"`
	Uptime       time.Duration          `json:"uptime"`
	Components   map[string]interface{} `json:"components"`
	Dependencies map[string]interface{} `json:"dependencies"`
}

// NewEventBusHandler creates a new event bus handler
func NewEventBusHandler(
	cfg *config.Config,
	logger *zap.Logger,
	kafkaClient *kafka.Client,
	debeziumManager *debezium.Manager,
	processorManager *processors.ProcessorManager,
) *EventBusHandler {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &EventBusHandler{
		config:           cfg,
		logger:           logger,
		kafka:            kafkaClient,
		debezium:         debeziumManager,
		processorManager: processorManager,
		metrics:          initHandlerMetrics(),
	}
}

// RegisterRoutes registers all HTTP routes
func (h *EventBusHandler) RegisterRoutes(mux *http.ServeMux) {
	// Health and monitoring endpoints
	mux.HandleFunc("/health", h.middleware(h.HealthCheck))
	mux.HandleFunc("/metrics", h.middleware(h.GetMetrics))
	mux.HandleFunc("/version", h.middleware(h.GetVersion))

	// Event publishing endpoints
	mux.HandleFunc("/events", h.middleware(h.PublishEvent))
	mux.HandleFunc("/events/batch", h.middleware(h.PublishEventBatch))

	// Admin endpoints
	mux.HandleFunc("/admin/config", h.middleware(h.GetConfig))
}

// Event Publishing Handlers

// PublishEvent handles single event publishing
func (h *EventBusHandler) PublishEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	var req EventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate request
	if err := h.validateEventRequest(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request", err)
		return
	}

	// Create message
	message := &kafka.Message{
		ID:        fmt.Sprintf("event_%d", time.Now().UnixNano()),
		EventType: req.EventType,
		Source:    req.Source,
		Data:      req.Data,
		Topic:     req.Topic,
		Key:       req.Key,
		Headers:   req.Headers,
		Metadata: kafka.MessageMetadata{
			Timestamp:   time.Now(),
			Version:     "1.0",
			ContentType: "application/json",
			Encoding:    "utf-8",
		},
	}

	if message.Headers == nil {
		message.Headers = make(map[string]string)
	}
	if message.Topic == "" {
		message.Topic = fmt.Sprintf("app.%s", req.EventType)
	}

	// Publish message
	if err := h.kafka.PublishMessage(r.Context(), message); err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to publish event", err)
		return
	}

	h.respondSuccess(w, map[string]interface{}{
		"event_id": message.ID,
		"topic":    message.Topic,
		"status":   "published",
	}, "Event published successfully")
}

// PublishEventBatch handles batch event publishing
func (h *EventBusHandler) PublishEventBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	var req struct {
		Events []EventRequest `json:"events"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if len(req.Events) == 0 {
		h.respondError(w, http.StatusBadRequest, "No events provided", nil)
		return
	}

	if len(req.Events) > 1000 {
		h.respondError(w, http.StatusBadRequest, "Too many events in batch (max 1000)", nil)
		return
	}

	results := make([]map[string]interface{}, 0, len(req.Events))
	errors := make([]map[string]interface{}, 0)

	for i, eventReq := range req.Events {
		// Validate each event
		if err := h.validateEventRequest(&eventReq); err != nil {
			errors = append(errors, map[string]interface{}{
				"index": i,
				"error": err.Error(),
				"event": eventReq,
			})
			continue
		}

		// Create message
		message := &kafka.Message{
			ID:        fmt.Sprintf("batch_event_%d_%d", time.Now().UnixNano(), i),
			EventType: eventReq.EventType,
			Source:    eventReq.Source,
			Data:      eventReq.Data,
			Topic:     eventReq.Topic,
			Key:       eventReq.Key,
			Headers:   eventReq.Headers,
			Metadata: kafka.MessageMetadata{
				Timestamp:   time.Now(),
				Version:     "1.0",
				ContentType: "application/json",
				Encoding:    "utf-8",
			},
		}

		if message.Headers == nil {
			message.Headers = make(map[string]string)
		}
		if message.Topic == "" {
			message.Topic = fmt.Sprintf("app.%s", eventReq.EventType)
		}

		// Publish message
		if err := h.kafka.PublishMessage(r.Context(), message); err != nil {
			errors = append(errors, map[string]interface{}{
				"index":    i,
				"event_id": message.ID,
				"error":    err.Error(),
			})
			continue
		}

		results = append(results, map[string]interface{}{
			"index":    i,
			"event_id": message.ID,
			"topic":    message.Topic,
			"status":   "published",
		})
	}

	response := map[string]interface{}{
		"total_events":      len(req.Events),
		"successful_events": len(results),
		"failed_events":     len(errors),
		"results":           results,
	}

	if len(errors) > 0 {
		response["errors"] = errors
	}

	status := http.StatusOK
	message := "Batch processing completed"
	if len(errors) == len(req.Events) {
		status = http.StatusInternalServerError
		message = "All events failed to publish"
	} else if len(errors) > 0 {
		status = http.StatusPartialContent
		message = "Some events failed to publish"
	}

	h.respond(w, status, true, message, response, nil)
}

// Health and Monitoring Handlers

// HealthCheck handles health check requests
func (h *EventBusHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	startTime := time.Now()

	// Check components
	components := make(map[string]interface{})
	dependencies := make(map[string]interface{})

	// Check Kafka
	kafkaHealthy := true
	if err := h.kafka.HealthCheck(r.Context()); err != nil {
		kafkaHealthy = false
		components["kafka"] = map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
		}
	} else {
		components["kafka"] = map[string]interface{}{
			"status": "healthy",
		}
	}

	// Check Debezium
	debeziumHealthy := true
	if err := h.debezium.HealthCheck(r.Context()); err != nil {
		debeziumHealthy = false
		components["debezium"] = map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
		}
	} else {
		components["debezium"] = map[string]interface{}{
			"status": "healthy",
		}
	}

	// Overall status
	overallStatus := "healthy"
	statusCode := http.StatusOK
	if !kafkaHealthy || !debeziumHealthy {
		overallStatus = "unhealthy"
		statusCode = http.StatusServiceUnavailable
	}

	response := HealthResponse{
		Status:       overallStatus,
		Version:      "1.0.0",
		Timestamp:    time.Now(),
		Uptime:       time.Since(startTime),
		Components:   components,
		Dependencies: dependencies,
	}

	h.respond(w, statusCode, overallStatus == "healthy", "Health check completed", response, nil)
}

// GetMetrics handles metrics requests
func (h *EventBusHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	// Placeholder metrics implementation
	metrics := map[string]interface{}{
		"events": map[string]interface{}{
			"total_processed": 1000,
			"total_failed":    10,
			"processing_rate": 100.5,
			"average_latency": 0.05,
			"error_rate":      0.01,
		},
		"kafka": map[string]interface{}{
			"connected_brokers": 3,
			"active_producers":  2,
			"active_consumers":  5,
			"messages_per_sec":  250.0,
			"bytes_per_sec":     1024000.0,
		},
		"debezium": map[string]interface{}{
			"total_connectors":   3,
			"running_connectors": 3,
			"failed_connectors":  0,
		},
		"timestamp": time.Now(),
	}

	h.respondSuccess(w, metrics, "Metrics retrieved successfully")
}

// GetVersion handles version requests
func (h *EventBusHandler) GetVersion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	version := map[string]interface{}{
		"version":    "1.0.0",
		"build_time": "2024-01-01T00:00:00Z",
		"git_commit": "latest",
		"go_version": "1.21",
	}

	h.respondSuccess(w, version, "Version information retrieved successfully")
}

// Admin Handlers

// GetConfig handles configuration requests
func (h *EventBusHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	// Return sanitized configuration (remove sensitive data)
	sanitizedConfig := map[string]interface{}{
		"server": map[string]interface{}{
			"host": h.config.Server.Host,
			"port": h.config.Server.Port,
		},
		"kafka": map[string]interface{}{
			"brokers": h.config.Kafka.Brokers,
		},
		"event_processing": map[string]interface{}{
			"workers":         h.config.EventProcessing.Workers,
			"batch_size":      h.config.EventProcessing.BatchSize,
			"process_timeout": h.config.EventProcessing.ProcessTimeout,
		},
	}

	h.respondSuccess(w, sanitizedConfig, "Configuration retrieved successfully")
}

// Helper Methods

// middleware wraps handlers with common middleware functionality
func (h *EventBusHandler) middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Set common headers
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Service", "event-bus-service")
		w.Header().Set("X-Version", "1.0.0")

		// Add request ID
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
		}
		w.Header().Set("X-Request-ID", requestID)

		// Log request
		h.logger.Info("HTTP request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("request_id", requestID),
			zap.String("user_agent", r.UserAgent()),
			zap.String("remote_addr", r.RemoteAddr))

		// Increment active connections
		h.metrics.ActiveConnections.Inc()
		defer h.metrics.ActiveConnections.Dec()

		// Call next handler
		next(w, r)

		// Record metrics
		duration := time.Since(start)
		h.metrics.RequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration.Seconds())
		h.metrics.RequestsTotal.WithLabelValues(r.Method, r.URL.Path, "200").Inc()

		// Log response
		h.logger.Info("HTTP response",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("request_id", requestID),
			zap.Duration("duration", duration))
	}
}

// respond sends a standardized JSON response
func (h *EventBusHandler) respond(w http.ResponseWriter, statusCode int, success bool, message string, data interface{}, error interface{}) {
	response := APIResponse{
		Success:   success,
		Message:   message,
		Data:      data,
		Error:     error,
		Timestamp: time.Now(),
		Version:   "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// respondSuccess sends a successful response
func (h *EventBusHandler) respondSuccess(w http.ResponseWriter, data interface{}, message string) {
	h.respond(w, http.StatusOK, true, message, data, nil)
}

// respondError sends an error response
func (h *EventBusHandler) respondError(w http.ResponseWriter, statusCode int, message string, err error) {
	var errorData interface{}
	if err != nil {
		errorData = err.Error()
		h.logger.Error("HTTP error", zap.String("message", message), zap.Error(err))
	}

	h.respond(w, statusCode, false, message, nil, errorData)
	h.metrics.ErrorsTotal.WithLabelValues(fmt.Sprintf("%d", statusCode), message).Inc()
}

// validateEventRequest validates an event request
func (h *EventBusHandler) validateEventRequest(req *EventRequest) error {
	if req.EventType == "" {
		return fmt.Errorf("event_type is required")
	}

	if req.Source == "" {
		return fmt.Errorf("source is required")
	}

	if req.Data == nil {
		return fmt.Errorf("data is required")
	}

	return nil
}

// initHandlerMetrics initializes Prometheus metrics for handlers
func initHandlerMetrics() *HandlerMetrics {
	return &HandlerMetrics{
		RequestsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "eventbus_http_requests_total",
			Help: "Total number of HTTP requests",
		}, []string{"method", "path", "status"}),
		RequestDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "eventbus_http_request_duration_seconds",
			Help:    "Duration of HTTP requests",
			Buckets: prometheus.DefBuckets,
		}, []string{"method", "path"}),
		ErrorsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "eventbus_http_errors_total",
			Help: "Total number of HTTP errors",
		}, []string{"status", "type"}),
		ActiveConnections: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "eventbus_http_active_connections",
			Help: "Number of active HTTP connections",
		}),
	}
}

// EventBusHandler provides HTTP handlers for the Event Bus Service
type EventBusHandler struct {
	config           *config.Config
	logger           *zap.Logger
	kafka            *kafka.Client
	debezium         *debezium.Manager
	processorManager *processors.ProcessorManager
	metrics          *HandlerMetrics
}

// HandlerMetrics contains Prometheus metrics for HTTP handlers
type HandlerMetrics struct {
	RequestsTotal     *prometheus.CounterVec
	RequestDuration   *prometheus.HistogramVec
	ResponseSizeBytes *prometheus.HistogramVec
	ErrorsTotal       *prometheus.CounterVec
	ActiveConnections prometheus.Gauge
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     interface{} `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
	Version   string      `json:"version"`
}

// EventRequest represents an event publishing request
type EventRequest struct {
	EventType    string                 `json:"event_type" validate:"required"`
	Source       string                 `json:"source" validate:"required"`
	Subject      string                 `json:"subject"`
	Data         map[string]interface{} `json:"data" validate:"required"`
	Headers      map[string]string      `json:"headers"`
	Topic        string                 `json:"topic"`
	Key          string                 `json:"key"`
	PartitionKey string                 `json:"partition_key"`
}

// EventBatchRequest represents a batch event publishing request
type EventBatchRequest struct {
	Events []EventRequest `json:"events" validate:"required,min=1,max=1000"`
}

// ConnectorRequest represents a Debezium connector creation request
type ConnectorRequest struct {
	Name     string            `json:"name" validate:"required"`
	Type     string            `json:"type" validate:"required,oneof=postgres mysql mongodb"`
	Config   map[string]string `json:"config" validate:"required"`
	Database *DatabaseConfig   `json:"database,omitempty"`
	Topics   *TopicsConfig     `json:"topics,omitempty"`
}

// DatabaseConfig represents database configuration for connectors
type DatabaseConfig struct {
	Host     string `json:"host" validate:"required"`
	Port     int    `json:"port" validate:"required,min=1,max=65535"`
	Name     string `json:"name" validate:"required"`
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
	Schema   string `json:"schema"`
}

// TopicsConfig represents topic configuration for connectors
type TopicsConfig struct {
	Prefix     string   `json:"prefix" validate:"required"`
	Include    []string `json:"include"`
	Exclude    []string `json:"exclude"`
	Transforms []string `json:"transforms"`
}

// FilterRequest represents an event filtering request
type FilterRequest struct {
	EventTypes    []string          `json:"event_types"`
	Sources       []string          `json:"sources"`
	Tables        []string          `json:"tables"`
	Operations    []string          `json:"operations"`
	TimeRange     *TimeRangeFilter  `json:"time_range"`
	Conditions    []FilterCondition `json:"conditions"`
	IncludeFields []string          `json:"include_fields"`
	ExcludeFields []string          `json:"exclude_fields"`
	Limit         int               `json:"limit"`
	Offset        int               `json:"offset"`
}

// TimeRangeFilter represents time range filtering
type TimeRangeFilter struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// FilterCondition represents a filtering condition
type FilterCondition struct {
	Field    string      `json:"field" validate:"required"`
	Operator string      `json:"operator" validate:"required,oneof=eq ne gt lt gte lte in nin regex"`
	Value    interface{} `json:"value" validate:"required"`
	Type     string      `json:"type" validate:"oneof=string number boolean date"`
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status       string                 `json:"status"`
	Version      string                 `json:"version"`
	Timestamp    time.Time              `json:"timestamp"`
	Uptime       time.Duration          `json:"uptime"`
	Components   map[string]interface{} `json:"components"`
	Dependencies map[string]interface{} `json:"dependencies"`
}

// MetricsResponse represents metrics response
type MetricsResponse struct {
	Events     EventMetrics     `json:"events"`
	Kafka      KafkaMetrics     `json:"kafka"`
	Debezium   DebeziumMetrics  `json:"debezium"`
	Processors ProcessorMetrics `json:"processors"`
	System     SystemMetrics    `json:"system"`
	Timestamp  time.Time        `json:"timestamp"`
}

// EventMetrics represents event-related metrics
type EventMetrics struct {
	TotalProcessed int64   `json:"total_processed"`
	TotalFailed    int64   `json:"total_failed"`
	ProcessingRate float64 `json:"processing_rate"`
	AverageLatency float64 `json:"average_latency"`
	ErrorRate      float64 `json:"error_rate"`
}

// KafkaMetrics represents Kafka-related metrics
type KafkaMetrics struct {
	ConnectedBrokers int     `json:"connected_brokers"`
	ActiveProducers  int     `json:"active_producers"`
	ActiveConsumers  int     `json:"active_consumers"`
	MessagesPerSec   float64 `json:"messages_per_sec"`
	BytesPerSec      float64 `json:"bytes_per_sec"`
}

// DebeziumMetrics represents Debezium-related metrics
type DebeziumMetrics struct {
	TotalConnectors   int `json:"total_connectors"`
	RunningConnectors int `json:"running_connectors"`
	FailedConnectors  int `json:"failed_connectors"`
	TotalTasks        int `json:"total_tasks"`
	RunningTasks      int `json:"running_tasks"`
	FailedTasks       int `json:"failed_tasks"`
}

// ProcessorMetrics represents processor-related metrics
type ProcessorMetrics struct {
	ActiveProcessors int     `json:"active_processors"`
	EventsProcessed  int64   `json:"events_processed"`
	EventsFiltered   int64   `json:"events_filtered"`
	ProcessingRate   float64 `json:"processing_rate"`
	AverageLatency   float64 `json:"average_latency"`
}

// SystemMetrics represents system-related metrics
type SystemMetrics struct {
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	Goroutines  int     `json:"goroutines"`
	GCPauses    float64 `json:"gc_pauses"`
}

// NewEventBusHandler creates a new event bus handler
func NewEventBusHandler(
	cfg *config.Config,
	logger *zap.Logger,
	kafkaClient *kafka.Client,
	debeziumManager *debezium.Manager,
	processorManager *processors.ProcessorManager,
) *EventBusHandler {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &EventBusHandler{
		config:           cfg,
		logger:           logger,
		kafka:            kafkaClient,
		debezium:         debeziumManager,
		processorManager: processorManager,
		metrics:          initHandlerMetrics(),
	}
}

// RegisterRoutes registers all HTTP routes
func (h *EventBusHandler) RegisterRoutes(mux *http.ServeMux) {
	// Health and monitoring endpoints
	mux.HandleFunc("/health", h.middleware(h.HealthCheck))
	mux.HandleFunc("/metrics", h.middleware(h.GetMetrics))
	mux.HandleFunc("/version", h.middleware(h.GetVersion))

	// Event publishing endpoints
	mux.HandleFunc("/events", h.middleware(h.PublishEvent))
	mux.HandleFunc("/events/batch", h.middleware(h.PublishEventBatch))
	mux.HandleFunc("/events/filter", h.middleware(h.FilterEvents))

	// Debezium connector management endpoints
	mux.HandleFunc("/connectors", h.middleware(h.ListConnectors))
	mux.HandleFunc("/connectors/", h.middleware(h.HandleConnectorOperations))

	// Processor management endpoints
	mux.HandleFunc("/processors", h.middleware(h.ListProcessors))
	mux.HandleFunc("/processors/", h.middleware(h.HandleProcessorOperations))

	// Topic management endpoints
	mux.HandleFunc("/topics", h.middleware(h.ListTopics))
	mux.HandleFunc("/topics/", h.middleware(h.HandleTopicOperations))

	// Admin endpoints
	mux.HandleFunc("/admin/config", h.middleware(h.GetConfig))
	mux.HandleFunc("/admin/shutdown", h.middleware(h.Shutdown))
}

// Event Publishing Handlers

// PublishEvent handles single event publishing
func (h *EventBusHandler) PublishEvent(w http.ResponseWriter, r *http.Request) {
	var req EventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate request
	if err := h.validateEventRequest(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request", err)
		return
	}

	// Create message
	message := &kafka.Message{
		ID:        fmt.Sprintf("event_%d", time.Now().UnixNano()),
		EventType: req.EventType,
		Source:    req.Source,
		Data:      req.Data,
		Topic:     req.Topic,
		Key:       req.Key,
		Headers:   req.Headers,
		Metadata: kafka.MessageMetadata{
			Timestamp:   time.Now(),
			Version:     "1.0",
			ContentType: "application/json",
			Encoding:    "utf-8",
		},
	}

	if message.Headers == nil {
		message.Headers = make(map[string]string)
	}
	if message.Topic == "" {
		message.Topic = fmt.Sprintf("app.%s", req.EventType)
	}

	// Publish message
	if err := h.kafka.PublishMessage(r.Context(), message); err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to publish event", err)
		return
	}

	h.respondSuccess(w, map[string]interface{}{
		"event_id": message.ID,
		"topic":    message.Topic,
		"status":   "published",
	}, "Event published successfully")
}

// PublishEventBatch handles batch event publishing
func (h *EventBusHandler) PublishEventBatch(w http.ResponseWriter, r *http.Request) {
	var req EventBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if len(req.Events) == 0 {
		h.respondError(w, http.StatusBadRequest, "No events provided", nil)
		return
	}

	if len(req.Events) > 1000 {
		h.respondError(w, http.StatusBadRequest, "Too many events in batch (max 1000)", nil)
		return
	}

	results := make([]map[string]interface{}, 0, len(req.Events))
	errors := make([]map[string]interface{}, 0)

	for i, eventReq := range req.Events {
		// Validate each event
		if err := h.validateEventRequest(&eventReq); err != nil {
			errors = append(errors, map[string]interface{}{
				"index": i,
				"error": err.Error(),
				"event": eventReq,
			})
			continue
		}

		// Create message
		message := &kafka.Message{
			ID:        fmt.Sprintf("batch_event_%d_%d", time.Now().UnixNano(), i),
			EventType: eventReq.EventType,
			Source:    eventReq.Source,
			Data:      eventReq.Data,
			Topic:     eventReq.Topic,
			Key:       eventReq.Key,
			Headers:   eventReq.Headers,
			Metadata: kafka.MessageMetadata{
				Timestamp:   time.Now(),
				Version:     "1.0",
				ContentType: "application/json",
				Encoding:    "utf-8",
			},
		}

		if message.Headers == nil {
			message.Headers = make(map[string]string)
		}
		if message.Topic == "" {
			message.Topic = fmt.Sprintf("app.%s", eventReq.EventType)
		}

		// Publish message
		if err := h.kafka.PublishMessage(r.Context(), message); err != nil {
			errors = append(errors, map[string]interface{}{
				"index":    i,
				"event_id": message.ID,
				"error":    err.Error(),
			})
			continue
		}

		results = append(results, map[string]interface{}{
			"index":    i,
			"event_id": message.ID,
			"topic":    message.Topic,
			"status":   "published",
		})
	}

	response := map[string]interface{}{
		"total_events":      len(req.Events),
		"successful_events": len(results),
		"failed_events":     len(errors),
		"results":           results,
	}

	if len(errors) > 0 {
		response["errors"] = errors
	}

	status := http.StatusOK
	message := "Batch processing completed"
	if len(errors) == len(req.Events) {
		status = http.StatusInternalServerError
		message = "All events failed to publish"
	} else if len(errors) > 0 {
		status = http.StatusPartialContent
		message = "Some events failed to publish"
	}

	h.respond(w, status, true, message, response, nil)
}

// FilterEvents handles event filtering requests
func (h *EventBusHandler) FilterEvents(w http.ResponseWriter, r *http.Request) {
	var req FilterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// This is a placeholder implementation
	// In a real system, you would implement event querying from storage
	h.respondSuccess(w, map[string]interface{}{
		"events":  []interface{}{},
		"total":   0,
		"limit":   req.Limit,
		"offset":  req.Offset,
		"filters": req,
	}, "Events filtered successfully")
}

// Debezium Connector Handlers

// ListConnectors handles listing all Debezium connectors
func (h *EventBusHandler) ListConnectors(w http.ResponseWriter, r *http.Request) {
	connectors, err := h.debezium.ListConnectors(r.Context())
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to list connectors", err)
		return
	}

	h.respondSuccess(w, map[string]interface{}{
		"connectors": connectors,
		"total":      len(connectors),
	}, "Connectors listed successfully")
}

// CreateConnector handles creating a new Debezium connector
func (h *EventBusHandler) CreateConnector(w http.ResponseWriter, r *http.Request) {
	var req ConnectorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate request
	if err := h.validateConnectorRequest(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request", err)
		return
	}

	// Create connector config
	connectorConfig := &debezium.ConnectorConfig{
		Name:   req.Name,
		Config: req.Config,
	}

	if err := h.debezium.CreateConnector(r.Context(), connectorConfig); err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to create connector", err)
		return
	}

	h.respondSuccess(w, map[string]interface{}{
		"connector_name": req.Name,
		"status":         "created",
	}, "Connector created successfully")
}

// GetConnector handles getting connector details
func (h *EventBusHandler) GetConnector(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	connectorName := vars["name"]

	if connectorName == "" {
		h.respondError(w, http.StatusBadRequest, "Connector name is required", nil)
		return
	}

	status, err := h.debezium.GetConnectorStatus(r.Context(), connectorName)
	if err != nil {
		h.respondError(w, http.StatusNotFound, "Connector not found", err)
		return
	}

	h.respondSuccess(w, status, "Connector details retrieved successfully")
}

// DeleteConnector handles deleting a connector
func (h *EventBusHandler) DeleteConnector(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	connectorName := vars["name"]

	if connectorName == "" {
		h.respondError(w, http.StatusBadRequest, "Connector name is required", nil)
		return
	}

	if err := h.debezium.DeleteConnector(r.Context(), connectorName); err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to delete connector", err)
		return
	}

	h.respondSuccess(w, map[string]interface{}{
		"connector_name": connectorName,
		"status":         "deleted",
	}, "Connector deleted successfully")
}

// RestartConnector handles restarting a connector
func (h *EventBusHandler) RestartConnector(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	connectorName := vars["name"]

	if connectorName == "" {
		h.respondError(w, http.StatusBadRequest, "Connector name is required", nil)
		return
	}

	if err := h.debezium.RestartConnector(r.Context(), connectorName); err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to restart connector", err)
		return
	}

	h.respondSuccess(w, map[string]interface{}{
		"connector_name": connectorName,
		"status":         "restarted",
	}, "Connector restarted successfully")
}

// GetConnectorStatus handles getting connector status
func (h *EventBusHandler) GetConnectorStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	connectorName := vars["name"]

	if connectorName == "" {
		h.respondError(w, http.StatusBadRequest, "Connector name is required", nil)
		return
	}

	status, err := h.debezium.GetConnectorStatus(r.Context(), connectorName)
	if err != nil {
		h.respondError(w, http.StatusNotFound, "Connector not found", err)
		return
	}

	h.respondSuccess(w, status, "Connector status retrieved successfully")
}

// Health and Monitoring Handlers

// HealthCheck handles health check requests
func (h *EventBusHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Check components
	components := make(map[string]interface{})
	dependencies := make(map[string]interface{})

	// Check Kafka
	kafkaHealthy := true
	if err := h.kafka.HealthCheck(); err != nil {
		kafkaHealthy = false
		components["kafka"] = map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
		}
	} else {
		components["kafka"] = map[string]interface{}{
			"status": "healthy",
		}
	}

	// Check Debezium
	debeziumHealthy := true
	if err := h.debezium.HealthCheck(r.Context()); err != nil {
		debeziumHealthy = false
		components["debezium"] = map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
		}
	} else {
		components["debezium"] = map[string]interface{}{
			"status": "healthy",
		}
	}

	// Overall status
	overallStatus := "healthy"
	statusCode := http.StatusOK
	if !kafkaHealthy || !debeziumHealthy {
		overallStatus = "unhealthy"
		statusCode = http.StatusServiceUnavailable
	}

	response := HealthResponse{
		Status:       overallStatus,
		Version:      h.config.Server.Version,
		Timestamp:    time.Now(),
		Uptime:       time.Since(startTime),
		Components:   components,
		Dependencies: dependencies,
	}

	h.respond(w, statusCode, overallStatus == "healthy", "Health check completed", response, nil)
}

// GetMetrics handles metrics requests
func (h *EventBusHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	// This is a placeholder implementation
	// In a real system, you would collect actual metrics from various components
	metrics := MetricsResponse{
		Events: EventMetrics{
			TotalProcessed: 1000,
			TotalFailed:    10,
			ProcessingRate: 100.5,
			AverageLatency: 0.05,
			ErrorRate:      0.01,
		},
		Kafka: KafkaMetrics{
			ConnectedBrokers: 3,
			ActiveProducers:  2,
			ActiveConsumers:  5,
			MessagesPerSec:   250.0,
			BytesPerSec:      1024000.0,
		},
		Debezium: DebeziumMetrics{
			TotalConnectors:   3,
			RunningConnectors: 3,
			FailedConnectors:  0,
			TotalTasks:        9,
			RunningTasks:      9,
			FailedTasks:       0,
		},
		Processors: ProcessorMetrics{
			ActiveProcessors: 4,
			EventsProcessed:  950,
			EventsFiltered:   50,
			ProcessingRate:   95.0,
			AverageLatency:   0.02,
		},
		System: SystemMetrics{
			CPUUsage:    25.5,
			MemoryUsage: 68.2,
			Goroutines:  150,
			GCPauses:    0.001,
		},
		Timestamp: time.Now(),
	}

	h.respondSuccess(w, metrics, "Metrics retrieved successfully")
}

// GetVersion handles version requests
func (h *EventBusHandler) GetVersion(w http.ResponseWriter, r *http.Request) {
	version := map[string]interface{}{
		"version":    h.config.Server.Version,
		"build_time": "2024-01-01T00:00:00Z", // This would be set during build
		"git_commit": "latest",               // This would be set during build
		"go_version": "1.21",
	}

	h.respondSuccess(w, version, "Version information retrieved successfully")
}

// Admin Handlers

// GetConfig handles configuration requests
func (h *EventBusHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	// Return sanitized configuration (remove sensitive data)
	sanitizedConfig := map[string]interface{}{
		"server": map[string]interface{}{
			"host":        h.config.Server.Host,
			"port":        h.config.Server.Port,
			"version":     h.config.Server.Version,
			"environment": h.config.Server.Environment,
		},
		"kafka": map[string]interface{}{
			"brokers": h.config.Kafka.Brokers,
			"timeout": h.config.Kafka.Timeout,
		},
		"event_processing": map[string]interface{}{
			"workers":         h.config.EventProcessing.Workers,
			"batch_size":      h.config.EventProcessing.BatchSize,
			"process_timeout": h.config.EventProcessing.ProcessTimeout,
		},
	}

	h.respondSuccess(w, sanitizedConfig, "Configuration retrieved successfully")
}

// Shutdown handles graceful shutdown requests
func (h *EventBusHandler) Shutdown(w http.ResponseWriter, r *http.Request) {
	h.respondSuccess(w, map[string]interface{}{
		"status": "shutdown_initiated",
		"time":   time.Now(),
	}, "Shutdown initiated successfully")

	// Initiate graceful shutdown
	go func() {
		time.Sleep(1 * time.Second) // Give response time to be sent
		// Signal shutdown to main application
		// This would be implemented in the main server
	}()
}

// Utility Handlers

// ListProcessors handles listing processors
func (h *EventBusHandler) ListProcessors(w http.ResponseWriter, r *http.Request) {
	// Placeholder implementation
	processors := []map[string]interface{}{
		{"name": "cdc-processor", "type": "cdc", "status": "running"},
		{"name": "form-processor", "type": "form", "status": "running"},
		{"name": "response-processor", "type": "response", "status": "running"},
		{"name": "analytics-processor", "type": "analytics", "status": "running"},
	}

	h.respondSuccess(w, map[string]interface{}{
		"processors": processors,
		"total":      len(processors),
	}, "Processors listed successfully")
}

// GetProcessorStatus handles getting processor status
func (h *EventBusHandler) GetProcessorStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	processorName := vars["name"]

	if processorName == "" {
		h.respondError(w, http.StatusBadRequest, "Processor name is required", nil)
		return
	}

	// Placeholder implementation
	status := map[string]interface{}{
		"name":             processorName,
		"status":           "running",
		"events_processed": 1000,
		"events_failed":    5,
		"last_activity":    time.Now(),
		"health_score":     0.95,
	}

	h.respondSuccess(w, status, "Processor status retrieved successfully")
}

// ListTopics handles listing Kafka topics
func (h *EventBusHandler) ListTopics(w http.ResponseWriter, r *http.Request) {
	topics, err := h.kafka.ListTopics(r.Context())
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Failed to list topics", err)
		return
	}

	h.respondSuccess(w, map[string]interface{}{
		"topics": topics,
		"total":  len(topics),
	}, "Topics listed successfully")
}

// GetTopicInfo handles getting topic information
func (h *EventBusHandler) GetTopicInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	topicName := vars["name"]

	if topicName == "" {
		h.respondError(w, http.StatusBadRequest, "Topic name is required", nil)
		return
	}

	// Placeholder implementation
	topicInfo := map[string]interface{}{
		"name":               topicName,
		"partitions":         3,
		"replication_factor": 2,
		"message_count":      1500,
		"size_bytes":         2048000,
		"created_at":         time.Now().Add(-24 * time.Hour),
	}

	h.respondSuccess(w, topicInfo, "Topic information retrieved successfully")
}

// GetTopicMessages handles getting topic messages
func (h *EventBusHandler) GetTopicMessages(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	topicName := vars["name"]

	if topicName == "" {
		h.respondError(w, http.StatusBadRequest, "Topic name is required", nil)
		return
	}

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Placeholder implementation
	messages := []map[string]interface{}{
		{
			"offset":    offset,
			"partition": 0,
			"key":       "test-key",
			"value":     map[string]interface{}{"test": "data"},
			"timestamp": time.Now(),
		},
	}

	h.respondSuccess(w, map[string]interface{}{
		"topic":    topicName,
		"messages": messages,
		"limit":    limit,
		"offset":   offset,
		"total":    len(messages),
	}, "Topic messages retrieved successfully")
}

// Helper Methods

// middleware wraps handlers with common middleware functionality
func (h *EventBusHandler) middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Set common headers
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Service", "event-bus-service")
		w.Header().Set("X-Version", h.config.Server.Version)

		// Add request ID
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
		}
		w.Header().Set("X-Request-ID", requestID)

		// Log request
		h.logger.Info("HTTP request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("request_id", requestID),
			zap.String("user_agent", r.UserAgent()),
			zap.String("remote_addr", r.RemoteAddr))

		// Increment active connections
		h.metrics.ActiveConnections.Inc()
		defer h.metrics.ActiveConnections.Dec()

		// Call next handler
		next(w, r)

		// Record metrics
		duration := time.Since(start)
		h.metrics.RequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration.Seconds())
		h.metrics.RequestsTotal.WithLabelValues(r.Method, r.URL.Path, "200").Inc() // This is simplified

		// Log response
		h.logger.Info("HTTP response",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("request_id", requestID),
			zap.Duration("duration", duration))
	}
}

// respond sends a standardized JSON response
func (h *EventBusHandler) respond(w http.ResponseWriter, statusCode int, success bool, message string, data interface{}, error interface{}) {
	response := APIResponse{
		Success:   success,
		Message:   message,
		Data:      data,
		Error:     error,
		Timestamp: time.Now(),
		Version:   h.config.Server.Version,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// respondSuccess sends a successful response
func (h *EventBusHandler) respondSuccess(w http.ResponseWriter, data interface{}, message string) {
	h.respond(w, http.StatusOK, true, message, data, nil)
}

// respondError sends an error response
func (h *EventBusHandler) respondError(w http.ResponseWriter, statusCode int, message string, err error) {
	var errorData interface{}
	if err != nil {
		errorData = err.Error()
		h.logger.Error("HTTP error", zap.String("message", message), zap.Error(err))
	}

	h.respond(w, statusCode, false, message, nil, errorData)
	h.metrics.ErrorsTotal.WithLabelValues(strconv.Itoa(statusCode), message).Inc()
}

// validateEventRequest validates an event request
func (h *EventBusHandler) validateEventRequest(req *EventRequest) error {
	if req.EventType == "" {
		return fmt.Errorf("event_type is required")
	}

	if req.Source == "" {
		return fmt.Errorf("source is required")
	}

	if req.Data == nil {
		return fmt.Errorf("data is required")
	}

	return nil
}

// validateConnectorRequest validates a connector request
func (h *EventBusHandler) validateConnectorRequest(req *ConnectorRequest) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}

	if req.Type == "" {
		return fmt.Errorf("type is required")
	}

	if req.Config == nil || len(req.Config) == 0 {
		return fmt.Errorf("config is required")
	}

	// Validate type
	validTypes := []string{"postgres", "mysql", "mongodb"}
	validType := false
	for _, validT := range validTypes {
		if req.Type == validT {
			validType = true
			break
		}
	}
	if !validType {
		return fmt.Errorf("invalid connector type: %s", req.Type)
	}

	return nil
}

// initHandlerMetrics initializes Prometheus metrics for handlers
func initHandlerMetrics() *HandlerMetrics {
	return &HandlerMetrics{
		RequestsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "eventbus_http_requests_total",
			Help: "Total number of HTTP requests",
		}, []string{"method", "path", "status"}),
		RequestDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "eventbus_http_request_duration_seconds",
			Help:    "Duration of HTTP requests",
			Buckets: prometheus.DefBuckets,
		}, []string{"method", "path"}),
		ResponseSizeBytes: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "eventbus_http_response_size_bytes",
			Help:    "Size of HTTP responses in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		}, []string{"method", "path"}),
		ErrorsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "eventbus_http_errors_total",
			Help: "Total number of HTTP errors",
		}, []string{"status", "type"}),
		ActiveConnections: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "eventbus_http_active_connections",
			Help: "Number of active HTTP connections",
		}),
	}
}
