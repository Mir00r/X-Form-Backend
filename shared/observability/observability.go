package observability

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

// Provider is the main observability provider that combines tracing, metrics, and error tracking
type Provider struct {
	tracing *TracingProvider
	metrics *MetricsProvider
	errors  *ErrorProvider
	logger  *zap.Logger
	config  Config
}

// Config holds configuration for the observability provider
type Config struct {
	ServiceName string
	Environment string
	Version     string
	Tracing     TracingConfig
	Metrics     MetricsConfig
	Errors      ErrorConfig
}

// New creates a new observability provider
func New(config Config, logger *zap.Logger) (*Provider, error) {
	// Initialize tracing
	tracing, err := NewTracingProvider(config.Tracing, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize tracing: %w", err)
	}

	// Initialize metrics
	metrics := NewMetricsProvider(config.Metrics, logger)

	// Initialize error tracking
	errors, err := NewErrorProvider(config.Errors, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize error tracking: %w", err)
	}

	provider := &Provider{
		tracing: tracing,
		metrics: metrics,
		errors:  errors,
		logger:  logger,
		config:  config,
	}

	logger.Info("Observability provider initialized",
		zap.String("service", config.ServiceName),
		zap.String("environment", config.Environment),
		zap.String("version", config.Version),
	)

	return provider, nil
}

// Tracing returns the tracing provider
func (p *Provider) Tracing() *TracingProvider {
	return p.tracing
}

// Metrics returns the metrics provider
func (p *Provider) Metrics() *MetricsProvider {
	return p.metrics
}

// Errors returns the error provider
func (p *Provider) Errors() *ErrorProvider {
	return p.errors
}

// HTTPMiddleware returns middleware for HTTP requests that automatically instruments requests
func (p *Provider) HTTPMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Increment active connections
			p.metrics.IncrementActiveConnections()
			defer p.metrics.DecrementActiveConnections()

			// Start tracing span
			ctx, span := p.tracing.StartSpan(r.Context(), fmt.Sprintf("HTTP %s %s", r.Method, r.URL.Path))
			defer span.End()

			// Add request attributes to span
			p.tracing.AddSpanAttributes(span,
				p.requestAttributes(r)...,
			)

			// Create response writer wrapper to capture status and size
			wrapper := &responseWriter{
				ResponseWriter: w,
				statusCode:     200,
			}

			// Add breadcrumb
			p.errors.AddBreadcrumb(
				fmt.Sprintf("HTTP %s %s", r.Method, r.URL.Path),
				"http.request",
				map[string]interface{}{
					"method": r.Method,
					"url":    r.URL.String(),
				},
			)

			// Execute request with context
			next.ServeHTTP(wrapper, r.WithContext(ctx))

			// Record metrics
			duration := time.Since(start)
			userID := p.extractUserID(r)
			endpoint := r.URL.Path
			statusCode := fmt.Sprintf("%d", wrapper.statusCode)

			p.metrics.RecordHTTPRequest(
				r.Method,
				endpoint,
				statusCode,
				userID,
				duration,
				r.ContentLength,
				int64(wrapper.size),
			)

			// Add final span attributes
			p.tracing.AddSpanAttributes(span,
				attribute.Int("http.status_code", wrapper.statusCode),
				attribute.Int64("http.response.size", int64(wrapper.size)),
				attribute.Float64("http.duration_ms", float64(duration.Nanoseconds())/1e6),
			)

			// Capture errors for 4xx/5xx responses
			if wrapper.statusCode >= 400 {
				errorContext := ErrorContext{
					UserID:     userID,
					RequestID:  p.extractRequestID(r),
					TraceID:    p.tracing.ExtractTraceID(ctx),
					SpanID:     p.tracing.ExtractSpanID(ctx),
					Endpoint:   endpoint,
					Method:     r.Method,
					StatusCode: wrapper.statusCode,
					Component:  "http",
					Operation:  "request",
				}

				if wrapper.statusCode >= 500 {
					p.errors.CaptureError(
						fmt.Errorf("HTTP %d error: %s %s", wrapper.statusCode, r.Method, r.URL.Path),
						errorContext,
					)
					p.metrics.RecordServiceError("http_error", "http", "error")
				} else {
					p.metrics.RecordServiceError("http_client_error", "http", "warning")
				}
			}
		})
	}
}

// DatabaseObservability provides observability for database operations
func (p *Provider) DatabaseObservability(ctx context.Context, operation, table string) (context.Context, func(error)) {
	start := time.Now()

	// Start span
	newCtx, span := p.tracing.StartSpan(ctx, fmt.Sprintf("DB %s %s", operation, table))

	// Add span attributes
	p.tracing.AddSpanAttributes(span,
		attribute.String("db.operation", operation),
		attribute.String("db.table", table),
		attribute.String("component", "database"),
	)

	// Return cleanup function
	return newCtx, func(err error) {
		defer span.End()

		duration := time.Since(start)
		status := "success"

		if err != nil {
			status = "error"
			p.tracing.RecordError(span, err)

			// Capture error
			errorContext := ErrorContext{
				TraceID:   p.tracing.ExtractTraceID(newCtx),
				SpanID:    p.tracing.ExtractSpanID(newCtx),
				Component: "database",
				Operation: operation,
				Extra: map[string]interface{}{
					"table": table,
				},
			}
			p.errors.CaptureError(err, errorContext)
		}

		// Record metrics
		p.metrics.RecordDBQuery(operation, table, status, duration)
	}
}

// ExternalServiceObservability provides observability for external service calls
func (p *Provider) ExternalServiceObservability(ctx context.Context, service, method, endpoint string) (context.Context, func(int, error)) {
	start := time.Now()

	// Start span
	newCtx, span := p.tracing.StartSpan(ctx, fmt.Sprintf("HTTP %s %s", method, endpoint))

	// Add span attributes
	p.tracing.AddSpanAttributes(span,
		attribute.String("http.method", method),
		attribute.String("http.url", endpoint),
		attribute.String("external.service", service),
		attribute.String("component", "external_service"),
	)

	// Return cleanup function
	return newCtx, func(statusCode int, err error) {
		defer span.End()

		duration := time.Since(start)
		statusCodeStr := fmt.Sprintf("%d", statusCode)

		// Add response attributes
		p.tracing.AddSpanAttributes(span,
			attribute.Int("http.status_code", statusCode),
		)

		if err != nil {
			p.tracing.RecordError(span, err)

			// Capture error
			errorContext := ErrorContext{
				TraceID:    p.tracing.ExtractTraceID(newCtx),
				SpanID:     p.tracing.ExtractSpanID(newCtx),
				Component:  "external_service",
				Operation:  method,
				StatusCode: statusCode,
				Extra: map[string]interface{}{
					"service":  service,
					"endpoint": endpoint,
				},
			}
			p.errors.CaptureError(err, errorContext)
		}

		// Record metrics
		p.metrics.RecordExternalServiceCall(service, method, endpoint, statusCodeStr, duration)
	}
}

// BusinessObservability provides observability for business operations
func (p *Provider) BusinessObservability(ctx context.Context, operation, category string) (context.Context, func(error)) {
	// Start span
	newCtx, span := p.tracing.StartSpan(ctx, fmt.Sprintf("Business %s", operation))

	// Add span attributes
	p.tracing.AddSpanAttributes(span,
		attribute.String("business.operation", operation),
		attribute.String("business.category", category),
		attribute.String("component", "business"),
	)

	// Return cleanup function
	return newCtx, func(err error) {
		defer span.End()

		status := "success"

		if err != nil {
			status = "error"
			p.tracing.RecordError(span, err)

			// Capture error
			errorContext := ErrorContext{
				TraceID:   p.tracing.ExtractTraceID(newCtx),
				SpanID:    p.tracing.ExtractSpanID(newCtx),
				Component: "business",
				Operation: operation,
				Extra: map[string]interface{}{
					"category": category,
				},
			}
			p.errors.CaptureError(err, errorContext)
		}

		// Record business metrics
		p.metrics.RecordBusinessEvent(operation, category, status)
		p.metrics.RecordServiceOperation(operation, status, "business")
	}
}

// Shutdown gracefully shuts down all observability providers
func (p *Provider) Shutdown(ctx context.Context) error {
	p.logger.Info("Shutting down observability provider")

	// Flush error tracking
	p.errors.Flush(5 * time.Second)

	// Shutdown tracing
	if err := p.tracing.Shutdown(ctx); err != nil {
		p.logger.Error("Failed to shutdown tracing", zap.Error(err))
		return err
	}

	return nil
}

// requestAttributes extracts attributes from HTTP request
func (p *Provider) requestAttributes(r *http.Request) []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		attribute.String("http.method", r.Method),
		attribute.String("http.url", r.URL.String()),
		attribute.String("http.scheme", r.URL.Scheme),
		attribute.String("http.host", r.Host),
		attribute.String("http.user_agent", r.UserAgent()),
	}

	if r.ContentLength > 0 {
		attrs = append(attrs, attribute.Int64("http.request.size", r.ContentLength))
	}

	return attrs
}

// extractUserID extracts user ID from request
func (p *Provider) extractUserID(r *http.Request) string {
	// Try different methods to extract user ID
	if userID := r.Header.Get("X-User-ID"); userID != "" {
		return userID
	}

	// Extract from JWT or other auth mechanisms
	// This would be implementation specific

	return ""
}

// extractRequestID extracts request ID from request
func (p *Provider) extractRequestID(r *http.Request) string {
	if requestID := r.Header.Get("X-Request-ID"); requestID != "" {
		return requestID
	}
	if requestID := r.Header.Get("Request-ID"); requestID != "" {
		return requestID
	}
	return ""
}

// responseWriter wraps http.ResponseWriter to capture status code and response size
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(data)
	rw.size += size
	return size, err
}

// DefaultConfig returns default observability configuration
func DefaultConfig(serviceName string) Config {
	return Config{
		ServiceName: serviceName,
		Environment: getEnv("ENVIRONMENT", "development"),
		Version:     getEnv("SERVICE_VERSION", "1.0.0"),
		Tracing:     DefaultTracingConfig(serviceName),
		Metrics:     DefaultMetricsConfig(serviceName),
		Errors:      DefaultErrorConfig(serviceName),
	}
}
