// Package telemetry provides comprehensive observability implementation
// This file implements error tracking for the Event Bus Service
// following enterprise best practices for error monitoring and alerting.
package telemetry

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/Mir00r/X-Form-Backend/services/event-bus-service/internal/config"
)

// ErrorLevel represents error severity levels
type ErrorLevel string

const (
	LevelDebug   ErrorLevel = "debug"
	LevelInfo    ErrorLevel = "info"
	LevelWarning ErrorLevel = "warning"
	LevelError   ErrorLevel = "error"
	LevelFatal   ErrorLevel = "fatal"
)

// ErrorProvider manages error tracking and reporting
type ErrorProvider struct {
	config  *config.Config
	logger  *zap.Logger
	enabled bool
}

// ErrorContext provides additional context for error reporting
type ErrorContext struct {
	UserID      string                 `json:"user_id,omitempty"`
	SessionID   string                 `json:"session_id,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
	TraceID     string                 `json:"trace_id,omitempty"`
	SpanID      string                 `json:"span_id,omitempty"`
	Component   string                 `json:"component,omitempty"`
	Operation   string                 `json:"operation,omitempty"`
	EventType   string                 `json:"event_type,omitempty"`
	Topic       string                 `json:"topic,omitempty"`
	Extra       map[string]interface{} `json:"extra,omitempty"`
	Tags        map[string]string      `json:"tags,omitempty"`
	Level       ErrorLevel             `json:"level,omitempty"`
	Fingerprint []string               `json:"fingerprint,omitempty"`
}

// ErrorEvent represents a captured error event
type ErrorEvent struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	Level       ErrorLevel             `json:"level"`
	Message     string                 `json:"message"`
	Error       string                 `json:"error,omitempty"`
	Context     *ErrorContext          `json:"context,omitempty"`
	StackTrace  string                 `json:"stack_trace,omitempty"`
	Environment string                 `json:"environment"`
	Service     string                 `json:"service"`
	Version     string                 `json:"version"`
	ServerName  string                 `json:"server_name"`
	Runtime     map[string]interface{} `json:"runtime"`
}

// NewErrorProvider creates a new error tracking provider
func NewErrorProvider(cfg *config.Config, logger *zap.Logger) (*ErrorProvider, error) {
	ep := &ErrorProvider{
		config:  cfg,
		logger:  logger,
		enabled: true, // Can be configured based on environment
	}

	ep.logger.Info("Error tracking provider initialized",
		zap.String("environment", cfg.Environment),
		zap.Bool("enabled", ep.enabled))

	return ep, nil
}

// CaptureError captures an error with optional context
func (ep *ErrorProvider) CaptureError(err error, ctx ...*ErrorContext) string {
	if !ep.enabled || err == nil {
		return ""
	}

	event := ep.createErrorEvent(err.Error(), LevelError, err, ctx...)
	ep.reportError(event)

	return event.ID
}

// CaptureMessage captures a message with optional context
func (ep *ErrorProvider) CaptureMessage(message string, level ErrorLevel, ctx ...*ErrorContext) string {
	if !ep.enabled {
		return ""
	}

	event := ep.createErrorEvent(message, level, nil, ctx...)
	ep.reportError(event)

	return event.ID
}

// CaptureHTTPError captures HTTP-related errors
func (ep *ErrorProvider) CaptureHTTPError(r *http.Request, err error, statusCode int) string {
	ctx := &ErrorContext{
		Component: "http",
		Operation: fmt.Sprintf("%s %s", r.Method, r.URL.Path),
		Extra: map[string]interface{}{
			"status_code": statusCode,
			"method":      r.Method,
			"url":         r.URL.String(),
			"user_agent":  r.UserAgent(),
			"remote_addr": r.RemoteAddr,
		},
		Tags: map[string]string{
			"http.method":      r.Method,
			"http.status_code": fmt.Sprintf("%d", statusCode),
		},
	}

	// Extract trace information from headers
	if traceID := r.Header.Get("X-Trace-ID"); traceID != "" {
		ctx.TraceID = traceID
	}
	if requestID := r.Header.Get("X-Request-ID"); requestID != "" {
		ctx.RequestID = requestID
	}

	// Set error level based on status code
	if statusCode >= 500 {
		ctx.Level = LevelError
	} else if statusCode >= 400 {
		ctx.Level = LevelWarning
	}

	return ep.CaptureError(err, ctx)
}

// CaptureKafkaError captures Kafka-related errors
func (ep *ErrorProvider) CaptureKafkaError(operation, topic string, partition int32, err error) string {
	ctx := &ErrorContext{
		Component: "kafka",
		Operation: operation,
		Topic:     topic,
		Extra: map[string]interface{}{
			"kafka.topic":     topic,
			"kafka.partition": partition,
			"kafka.operation": operation,
		},
		Tags: map[string]string{
			"kafka.topic":     topic,
			"kafka.operation": operation,
		},
		Level: LevelError,
	}

	return ep.CaptureError(err, ctx)
}

// CaptureEventProcessingError captures event processing errors
func (ep *ErrorProvider) CaptureEventProcessingError(eventType, source string, err error, eventData interface{}) string {
	ctx := &ErrorContext{
		Component: "event_processor",
		Operation: "process_event",
		EventType: eventType,
		Extra: map[string]interface{}{
			"event.type":   eventType,
			"event.source": source,
			"event.data":   eventData,
		},
		Tags: map[string]string{
			"event.type":   eventType,
			"event.source": source,
		},
		Level: LevelError,
	}

	return ep.CaptureError(err, ctx)
}

// CaptureCDCError captures CDC/Debezium related errors
func (ep *ErrorProvider) CaptureCDCError(connector, table, operation string, err error) string {
	ctx := &ErrorContext{
		Component: "cdc",
		Operation: operation,
		Extra: map[string]interface{}{
			"cdc.connector": connector,
			"cdc.table":     table,
			"cdc.operation": operation,
		},
		Tags: map[string]string{
			"cdc.connector": connector,
			"cdc.table":     table,
			"cdc.operation": operation,
		},
		Level: LevelError,
	}

	return ep.CaptureError(err, ctx)
}

// CapturePanic captures panic with stack trace
func (ep *ErrorProvider) CapturePanic(panicValue interface{}) string {
	ctx := &ErrorContext{
		Component: "goroutine",
		Operation: "panic_recovery",
		Level:     LevelFatal,
		Extra: map[string]interface{}{
			"panic_value": panicValue,
			"stack_trace": string(debug.Stack()),
		},
	}

	var err error
	switch v := panicValue.(type) {
	case error:
		err = v
	case string:
		err = fmt.Errorf("panic: %s", v)
	default:
		err = fmt.Errorf("panic: %v", v)
	}

	return ep.CaptureError(err, ctx)
}

// HTTPMiddleware provides error tracking HTTP middleware
func (ep *ErrorProvider) HTTPMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					ep.CapturePanic(recovered)
					// Re-panic to maintain original behavior
					panic(recovered)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// Recovery provides a recovery function for goroutines
func (ep *ErrorProvider) Recovery() func() {
	return func() {
		if recovered := recover(); recovered != nil {
			ep.CapturePanic(recovered)
			// Re-panic to maintain original behavior
			panic(recovered)
		}
	}
}

// Shutdown gracefully shuts down error tracking
func (ep *ErrorProvider) Shutdown() {
	ep.logger.Info("Shutting down error tracking provider")
}

// Helper methods

// createErrorEvent creates an error event with full context
func (ep *ErrorProvider) createErrorEvent(message string, level ErrorLevel, err error, ctx ...*ErrorContext) *ErrorEvent {
	event := &ErrorEvent{
		ID:          ep.generateEventID(),
		Timestamp:   time.Now(),
		Level:       level,
		Message:     message,
		Environment: ep.config.Environment,
		Service:     "event-bus-service",
		Version:     ep.config.Version,
		ServerName:  ep.getServerName(),
		Runtime: map[string]interface{}{
			"go_version": runtime.Version(),
			"goroutines": runtime.NumGoroutine(),
			"go_os":      runtime.GOOS,
			"go_arch":    runtime.GOARCH,
		},
	}

	if err != nil {
		event.Error = err.Error()
		event.StackTrace = string(debug.Stack())
	}

	if len(ctx) > 0 && ctx[0] != nil {
		event.Context = ctx[0]
	}

	return event
}

// reportError reports the error event (can be extended to send to external services)
func (ep *ErrorProvider) reportError(event *ErrorEvent) {
	// Log the error with structured logging
	logFields := []zap.Field{
		zap.String("event_id", event.ID),
		zap.String("level", string(event.Level)),
		zap.String("message", event.Message),
		zap.Time("timestamp", event.Timestamp),
	}

	if event.Error != "" {
		logFields = append(logFields, zap.String("error", event.Error))
	}

	if event.Context != nil {
		if event.Context.Component != "" {
			logFields = append(logFields, zap.String("component", event.Context.Component))
		}
		if event.Context.Operation != "" {
			logFields = append(logFields, zap.String("operation", event.Context.Operation))
		}
		if event.Context.TraceID != "" {
			logFields = append(logFields, zap.String("trace_id", event.Context.TraceID))
		}
		if event.Context.RequestID != "" {
			logFields = append(logFields, zap.String("request_id", event.Context.RequestID))
		}
	}

	// Log based on level
	switch event.Level {
	case LevelDebug:
		ep.logger.Debug("Error captured", logFields...)
	case LevelInfo:
		ep.logger.Info("Error captured", logFields...)
	case LevelWarning:
		ep.logger.Warn("Error captured", logFields...)
	case LevelError:
		ep.logger.Error("Error captured", logFields...)
	case LevelFatal:
		ep.logger.Fatal("Error captured", logFields...)
	default:
		ep.logger.Error("Error captured", logFields...)
	}

	// Here you can add integration with external error tracking services
	// For example: Sentry, Bugsnag, Rollbar, etc.
	// ep.sendToExternalService(event)
}

// generateEventID generates a unique event ID
func (ep *ErrorProvider) generateEventID() string {
	return fmt.Sprintf("err_%d_%d", time.Now().UnixNano(), os.Getpid())
}

// getServerName gets the server name for error context
func (ep *ErrorProvider) getServerName() string {
	if hostname, err := os.Hostname(); err == nil {
		return hostname
	}
	return "unknown"
}

// Configuration methods for future Sentry integration

// IsSentryEnabled checks if Sentry error tracking should be enabled
func (ep *ErrorProvider) IsSentryEnabled() bool {
	return os.Getenv("SENTRY_DSN") != ""
}

// GetSentryDSN returns the Sentry DSN from environment
func (ep *ErrorProvider) GetSentryDSN() string {
	return os.Getenv("SENTRY_DSN")
}

// GetSentryEnvironment returns the Sentry environment
func (ep *ErrorProvider) GetSentryEnvironment() string {
	if env := os.Getenv("SENTRY_ENVIRONMENT"); env != "" {
		return env
	}
	return ep.config.Environment
}

// GetSentryRelease returns the Sentry release version
func (ep *ErrorProvider) GetSentryRelease() string {
	if release := os.Getenv("SENTRY_RELEASE"); release != "" {
		return release
	}
	return fmt.Sprintf("event-bus-service@%s", ep.config.Version)
}

// GetSentrySampleRate returns the Sentry sample rate
func (ep *ErrorProvider) GetSentrySampleRate() float64 {
	if rate := os.Getenv("SENTRY_SAMPLE_RATE"); rate != "" {
		if parsed, err := strconv.ParseFloat(rate, 64); err == nil {
			return parsed
		}
	}
	return 1.0 // 100% by default for errors
}
