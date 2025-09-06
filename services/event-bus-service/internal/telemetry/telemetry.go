// Package telemetry provides comprehensive observability implementation
// This file implements the main telemetry provider that integrates
// tracing, metrics, and error tracking for enterprise observability.
package telemetry

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"

	"github.com/Mir00r/X-Form-Backend/services/event-bus-service/internal/config"
)

// Provider represents the comprehensive telemetry provider
type Provider struct {
	config      *config.Config
	logger      *zap.Logger
	tracing     *TracerProvider
	metrics     *MetricsProvider
	errors      *ErrorProvider
	initialized bool
}

// ProviderConfig defines telemetry provider configuration
type ProviderConfig struct {
	ServiceName    string `json:"service_name"`
	ServiceVersion string `json:"service_version"`
	Environment    string `json:"environment"`
	EnableTracing  bool   `json:"enable_tracing"`
	EnableMetrics  bool   `json:"enable_metrics"`
	EnableErrors   bool   `json:"enable_errors"`
}

// New creates a new comprehensive telemetry provider
func New(cfg *config.Config, logger *zap.Logger) (*Provider, error) {
	provider := &Provider{
		config: cfg,
		logger: logger,
	}

	if err := provider.initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize telemetry provider: %w", err)
	}

	return provider, nil
}

// initialize sets up all telemetry components
func (p *Provider) initialize() error {
	p.logger.Info("Initializing telemetry provider",
		zap.String("service", "event-bus-service"),
		zap.String("environment", p.config.Environment))

	// Initialize tracing provider
	if err := p.initializeTracing(); err != nil {
		return fmt.Errorf("failed to initialize tracing: %w", err)
	}

	// Initialize metrics provider
	if err := p.initializeMetrics(); err != nil {
		return fmt.Errorf("failed to initialize metrics: %w", err)
	}

	// Initialize error tracking provider
	if err := p.initializeErrors(); err != nil {
		return fmt.Errorf("failed to initialize error tracking: %w", err)
	}

	p.initialized = true
	p.logger.Info("Telemetry provider initialized successfully")

	return nil
}

// initializeTracing sets up distributed tracing
func (p *Provider) initializeTracing() error {
	tracingProvider, err := NewTracerProvider(p.config, p.logger)
	if err != nil {
		return fmt.Errorf("failed to create tracing provider: %w", err)
	}

	p.tracing = tracingProvider
	p.logger.Info("Tracing provider initialized")
	return nil
}

// initializeMetrics sets up metrics collection
func (p *Provider) initializeMetrics() error {
	metricsProvider, err := NewMetricsProvider(p.config, p.logger)
	if err != nil {
		return fmt.Errorf("failed to create metrics provider: %w", err)
	}

	p.metrics = metricsProvider
	p.logger.Info("Metrics provider initialized")
	return nil
}

// initializeErrors sets up error tracking
func (p *Provider) initializeErrors() error {
	errorProvider, err := NewErrorProvider(p.config, p.logger)
	if err != nil {
		return fmt.Errorf("failed to create error provider: %w", err)
	}

	p.errors = errorProvider
	p.logger.Info("Error tracking provider initialized")
	return nil
}

// Tracing returns the tracing provider
func (p *Provider) Tracing() *TracerProvider {
	return p.tracing
}

// Metrics returns the metrics provider
func (p *Provider) Metrics() *MetricsProvider {
	return p.metrics
}

// Errors returns the error tracking provider
func (p *Provider) Errors() *ErrorProvider {
	return p.errors
}

// IsInitialized returns whether the provider is fully initialized
func (p *Provider) IsInitialized() bool {
	return p.initialized
}

// HTTPMiddleware returns comprehensive HTTP observability middleware
func (p *Provider) HTTPMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create tracing span
			ctx, span := p.tracing.StartHTTPSpan(r.Context(), r)
			defer span.End()

			// Wrap response writer for metrics
			wrapped := &httpResponseWriter{
				ResponseWriter: w,
				statusCode:     200,
			}

			// Error recovery
			defer func() {
				if recovered := recover(); recovered != nil {
					p.errors.CapturePanic(recovered)
					// Re-panic to maintain original behavior
					panic(recovered)
				}
			}()

			// Process request
			next.ServeHTTP(wrapped, r.WithContext(ctx))

			// Record metrics
			duration := time.Since(start)
			p.metrics.RecordHTTPRequest(r.Method, r.URL.Path, wrapped.statusCode, duration)

			// Record error if status code indicates error
			if wrapped.statusCode >= 400 {
				err := fmt.Errorf("HTTP %d: %s %s", wrapped.statusCode, r.Method, r.URL.Path)
				p.errors.CaptureHTTPError(r, err, wrapped.statusCode)
			}

			// Add span attributes
			span.SetAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.route", r.URL.Path),
				attribute.Int("http.status_code", wrapped.statusCode),
				attribute.Int64("http.response_time_ms", duration.Milliseconds()),
			)

			if wrapped.statusCode >= 400 {
				span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", wrapped.statusCode))
			}
		})
	}
}

// EventProcessingObservability provides observability for event processing
func (p *Provider) EventProcessingObservability(ctx context.Context, eventType, source string) (context.Context, func(error)) {
	// Start tracing
	newCtx, span := p.tracing.StartSpan(ctx, "event_processing")
	span.SetAttributes(
		attribute.String("event.type", eventType),
		attribute.String("event.source", source),
		attribute.String("component", "event_processor"),
	)

	// Return cleanup function
	cleanup := func(err error) {
		// Record metrics
		p.metrics.RecordEventConsumed(eventType, "", source)

		// Handle errors
		if err != nil {
			p.errors.CaptureEventProcessingError(eventType, source, err, nil)
			span.SetStatus(codes.Error, err.Error())
		}

		span.End()
	}

	return newCtx, cleanup
}

// KafkaObservability provides observability for Kafka operations
func (p *Provider) KafkaObservability(ctx context.Context, operation, topic string, partition int32) (context.Context, func(error)) {
	// Start tracing
	newCtx, span := p.tracing.StartSpan(ctx, operation)
	span.SetAttributes(
		attribute.String("kafka.operation", operation),
		attribute.String("kafka.topic", topic),
		attribute.Int("kafka.partition", int(partition)),
	)

	// Return cleanup function
	cleanup := func(err error) {
		// Record metrics
		p.metrics.RecordKafkaOperation(operation, topic, partition)

		// Handle errors
		if err != nil {
			p.errors.CaptureKafkaError(operation, topic, partition, err)
			span.SetStatus(codes.Error, err.Error())
		}

		span.End()
	}

	return newCtx, cleanup
}

// CDCObservability provides observability for CDC operations
func (p *Provider) CDCObservability(ctx context.Context, connector, table, operation string) (context.Context, func(error)) {
	// Start tracing
	newCtx, span := p.tracing.StartDatabaseSpan(ctx, operation, table)
	span.SetAttributes(
		attribute.String("cdc.connector", connector),
		attribute.String("cdc.table", table),
		attribute.String("cdc.operation", operation),
	)

	// Return cleanup function
	cleanup := func(err error) {
		// Record metrics
		p.metrics.RecordCDCEvent(connector, table, operation)

		// Handle errors
		if err != nil {
			p.errors.CaptureCDCError(connector, table, operation, err)
			span.SetStatus(codes.Error, err.Error())
		}

		span.End()
	}

	return newCtx, cleanup
}

// Health returns the health status of all telemetry components
func (p *Provider) Health() map[string]interface{} {
	health := map[string]interface{}{
		"initialized": p.initialized,
		"tracing": map[string]interface{}{
			"enabled": p.tracing != nil,
		},
		"metrics": map[string]interface{}{
			"enabled": p.metrics != nil,
		},
		"errors": map[string]interface{}{
			"enabled": p.errors != nil,
		},
	}

	return health
}

// Shutdown gracefully shuts down all telemetry components
func (p *Provider) Shutdown(ctx context.Context) error {
	p.logger.Info("Shutting down telemetry provider")

	// Shutdown in reverse order of initialization
	if p.errors != nil {
		p.errors.Shutdown()
	}

	if p.metrics != nil {
		p.metrics.Shutdown(ctx)
	}

	if p.tracing != nil {
		if err := p.tracing.Shutdown(ctx); err != nil {
			p.logger.Error("Failed to shutdown tracing provider", zap.Error(err))
			return err
		}
	}

	p.initialized = false
	p.logger.Info("Telemetry provider shutdown completed")

	return nil
}

// httpResponseWriter wraps http.ResponseWriter to capture status code
type httpResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *httpResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *httpResponseWriter) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}

// Configuration helper functions

// GetDefaultProviderConfig returns default telemetry provider configuration
func GetDefaultProviderConfig(cfg *config.Config) *ProviderConfig {
	return &ProviderConfig{
		ServiceName:    "event-bus-service",
		ServiceVersion: cfg.Version,
		Environment:    cfg.Environment,
		EnableTracing:  true,
		EnableMetrics:  true,
		EnableErrors:   true,
	}
}

// ValidateProviderConfig validates telemetry provider configuration
func ValidateProviderConfig(cfg *ProviderConfig) error {
	if cfg.ServiceName == "" {
		return fmt.Errorf("service name is required")
	}

	if cfg.ServiceVersion == "" {
		return fmt.Errorf("service version is required")
	}

	if cfg.Environment == "" {
		return fmt.Errorf("environment is required")
	}

	return nil
}

// Observability utility functions

// WithTracing adds tracing context to an operation
func (p *Provider) WithTracing(ctx context.Context, operationName string, attributes map[string]interface{}) (context.Context, func()) {
	newCtx, span := p.tracing.StartSpan(ctx, operationName)

	for key, value := range attributes {
		span.SetAttributes(attribute.String(key, fmt.Sprintf("%v", value)))
	}

	return newCtx, func() {
		span.End()
	}
}

// WithMetrics records metrics for an operation
func (p *Provider) WithMetrics(operation string, labels map[string]string) func(error, time.Duration) {
	return func(err error, duration time.Duration) {
		// Record error if present
		if err != nil {
			p.metrics.RecordError("operation_error", "telemetry", "error")
		}
	}
}

// RecordBusinessMetric records a business-specific metric
func (p *Provider) RecordBusinessMetric(name string, value float64, labels map[string]string) {
	// Use analytics event for business metrics
	p.metrics.RecordAnalyticsEvent(name, "business")
}

// CaptureError provides a simple interface to capture errors
func (p *Provider) CaptureError(err error, component, operation string, extra map[string]interface{}) string {
	ctx := &ErrorContext{
		Component: component,
		Operation: operation,
		Extra:     extra,
		Level:     LevelError,
	}
	return p.errors.CaptureError(err, ctx)
}
