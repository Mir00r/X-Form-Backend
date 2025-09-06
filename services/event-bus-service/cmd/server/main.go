// Event Bus Service - Main Application Entry Point
// This is the main entry point for the Event Bus Service, implementing enterprise-grade
// Change Data Capture (CDC) with Kafka and Debezium for microservices architecture.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Mir00r/X-Form-Backend/services/event-bus-service/internal/config"
	"github.com/Mir00r/X-Form-Backend/services/event-bus-service/internal/debezium"
	"github.com/Mir00r/X-Form-Backend/services/event-bus-service/internal/kafka"
	"github.com/Mir00r/X-Form-Backend/services/event-bus-service/internal/processors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Application represents the main application
type Application struct {
	config           *config.Config
	logger           *zap.Logger
	kafka            *kafka.Client
	debezium         *debezium.Manager
	processorManager *processors.ProcessorManager
	httpServer       *http.Server
	metricsServer    *http.Server
	stopCh           chan struct{}
}

// EventBusHandler provides basic HTTP handlers for the Event Bus Service
type EventBusHandler struct {
	config           *config.Config
	logger           *zap.Logger
	kafka            *kafka.Client
	debezium         *debezium.Manager
	processorManager *processors.ProcessorManager
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     interface{} `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	Version   string      `json:"version"`
}

// EventRequest represents an event publishing request
type EventRequest struct {
	EventType string                 `json:"event_type"`
	Source    string                 `json:"source"`
	Data      map[string]interface{} `json:"data"`
	Topic     string                 `json:"topic"`
	Key       string                 `json:"key"`
	Headers   map[string]string      `json:"headers"`
}

// main is the application entry point
func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	logger, err := initLogger(cfg)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting Event Bus Service",
		zap.String("version", "1.0.0"),
		zap.String("environment", "development"))

	// Create application
	app, err := NewApplication(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to create application", zap.Error(err))
	}

	// Start application
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := app.Start(ctx); err != nil {
		logger.Fatal("Failed to start application", zap.Error(err))
	}

	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	logger.Info("Event Bus Service started successfully")
	<-sigCh

	logger.Info("Shutdown signal received")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := app.Stop(shutdownCtx); err != nil {
		logger.Error("Error during shutdown", zap.Error(err))
	}

	logger.Info("Event Bus Service stopped")
}

// NewApplication creates a new application instance
func NewApplication(cfg *config.Config, logger *zap.Logger) (*Application, error) {
	app := &Application{
		config: cfg,
		logger: logger,
		stopCh: make(chan struct{}),
	}

	// Initialize Kafka client
	kafkaClient, err := kafka.NewClient(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka client: %w", err)
	}
	app.kafka = kafkaClient

	// Initialize Debezium manager
	debeziumManager, err := debezium.NewManager(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create Debezium manager: %w", err)
	}
	app.debezium = debeziumManager

	// Initialize processor manager
	processorManager, err := processors.NewProcessorManager(cfg, logger, kafkaClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create processor manager: %w", err)
	}
	app.processorManager = processorManager

	// Setup HTTP servers
	if err := app.setupHTTPServers(); err != nil {
		return nil, fmt.Errorf("failed to setup HTTP servers: %w", err)
	}

	return app, nil
}

// Start starts the application and all its components
func (app *Application) Start(ctx context.Context) error {
	app.logger.Info("Starting application components")

	// Start Debezium manager
	if err := app.debezium.Start(ctx); err != nil {
		return fmt.Errorf("failed to start Debezium manager: %w", err)
	}

	// Start processor manager
	if err := app.processorManager.Start(ctx); err != nil {
		return fmt.Errorf("failed to start processor manager: %w", err)
	}

	// Start HTTP servers
	if err := app.startHTTPServers(); err != nil {
		return fmt.Errorf("failed to start HTTP servers: %w", err)
	}

	return nil
}

// Stop stops the application and all its components
func (app *Application) Stop(ctx context.Context) error {
	app.logger.Info("Stopping application components")

	// Stop HTTP servers
	if err := app.stopHTTPServers(ctx); err != nil {
		app.logger.Error("Error stopping HTTP servers", zap.Error(err))
	}

	// Stop processor manager
	if err := app.processorManager.Stop(); err != nil {
		app.logger.Error("Error stopping processor manager", zap.Error(err))
	}

	// Stop Debezium manager
	if err := app.debezium.Stop(); err != nil {
		app.logger.Error("Error stopping Debezium manager", zap.Error(err))
	}

	// Close Kafka client
	if err := app.kafka.Close(); err != nil {
		app.logger.Error("Error closing Kafka client", zap.Error(err))
	}

	close(app.stopCh)
	return nil
}

// setupHTTPServers sets up the HTTP servers for API and metrics
func (app *Application) setupHTTPServers() error {
	// Setup main API server
	mux := http.NewServeMux()

	// Create handler
	handler := &EventBusHandler{
		config:           app.config,
		logger:           app.logger,
		kafka:            app.kafka,
		debezium:         app.debezium,
		processorManager: app.processorManager,
	}

	// Register routes
	handler.RegisterRoutes(mux)

	app.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", app.config.Server.Host, app.config.Server.Port),
		Handler:      mux,
		ReadTimeout:  app.config.Server.ReadTimeout,
		WriteTimeout: app.config.Server.WriteTimeout,
		IdleTimeout:  app.config.Server.IdleTimeout,
	}

	// Setup metrics server if enabled
	if app.config.Observability.Metrics.Enabled {
		metricsMux := http.NewServeMux()
		metricsMux.Handle("/metrics", promhttp.Handler())

		app.metricsServer = &http.Server{
			Addr:         fmt.Sprintf(":%d", app.config.Observability.Metrics.Port),
			Handler:      metricsMux,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		}
	}

	return nil
}

// startHTTPServers starts the HTTP servers
func (app *Application) startHTTPServers() error {
	// Start main API server
	go func() {
		app.logger.Info("Starting HTTP server",
			zap.String("address", app.httpServer.Addr))

		if err := app.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			app.logger.Error("HTTP server error", zap.Error(err))
		}
	}()

	// Start metrics server if configured
	if app.metricsServer != nil {
		go func() {
			app.logger.Info("Starting metrics server",
				zap.String("address", app.metricsServer.Addr))

			if err := app.metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				app.logger.Error("Metrics server error", zap.Error(err))
			}
		}()
	}

	return nil
}

// stopHTTPServers stops the HTTP servers
func (app *Application) stopHTTPServers(ctx context.Context) error {
	var lastErr error

	// Stop main API server
	if app.httpServer != nil {
		app.logger.Info("Stopping HTTP server")
		if err := app.httpServer.Shutdown(ctx); err != nil {
			app.logger.Error("Error stopping HTTP server", zap.Error(err))
			lastErr = err
		}
	}

	// Stop metrics server
	if app.metricsServer != nil {
		app.logger.Info("Stopping metrics server")
		if err := app.metricsServer.Shutdown(ctx); err != nil {
			app.logger.Error("Error stopping metrics server", zap.Error(err))
			lastErr = err
		}
	}

	return lastErr
}

// HTTP Handler Methods

// RegisterRoutes registers all HTTP routes
func (h *EventBusHandler) RegisterRoutes(mux *http.ServeMux) {
	// Health and monitoring endpoints
	mux.HandleFunc("/health", h.middleware(h.HealthCheck))
	mux.HandleFunc("/version", h.middleware(h.GetVersion))

	// Event publishing endpoints
	mux.HandleFunc("/events", h.middleware(h.PublishEvent))

	// Admin endpoints
	mux.HandleFunc("/admin/config", h.middleware(h.GetConfig))
}

// HealthCheck handles health check requests
func (h *EventBusHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	// Check components
	components := make(map[string]interface{})

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

	response := map[string]interface{}{
		"status":     overallStatus,
		"version":    "1.0.0",
		"timestamp":  time.Now(),
		"components": components,
	}

	h.respond(w, statusCode, overallStatus == "healthy", "Health check completed", response, nil)
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

// PublishEvent handles event publishing
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
	if req.EventType == "" {
		h.respondError(w, http.StatusBadRequest, "event_type is required", nil)
		return
	}
	if req.Source == "" {
		h.respondError(w, http.StatusBadRequest, "source is required", nil)
		return
	}
	if req.Data == nil {
		h.respondError(w, http.StatusBadRequest, "data is required", nil)
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
			"workers":    h.config.EventProcessing.Workers,
			"batch_size": h.config.EventProcessing.BatchSize,
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
			zap.String("request_id", requestID))

		// Call next handler
		next(w, r)

		// Log response
		duration := time.Since(start)
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
}

// Utility Functions

// initLogger initializes the logger based on configuration
func initLogger(cfg *config.Config) (*zap.Logger, error) {
	// Configure logger based on environment
	var zapConfig zap.Config

	if cfg.Observability.Logging.Level == "debug" {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	} else {
		zapConfig = zap.NewProductionConfig()
		switch cfg.Observability.Logging.Level {
		case "info":
			zapConfig.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
		case "warn":
			zapConfig.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
		case "error":
			zapConfig.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
		default:
			zapConfig.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
		}
	}

	// Configure output format
	if cfg.Observability.Logging.Format == "json" {
		zapConfig.Encoding = "json"
	} else {
		zapConfig.Encoding = "console"
	}

	logger, err := zapConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return logger, nil
}
