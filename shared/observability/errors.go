package observability

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"
)

// ErrorProvider handles error tracking and reporting
type ErrorProvider struct {
	sentryEnabled bool
	serviceName   string
	environment   string
	logger        *zap.Logger
}

// ErrorConfig holds configuration for error tracking
type ErrorConfig struct {
	ServiceName   string
	Environment   string
	SentryDSN     string
	SentryRelease string
	SampleRate    float64
	EnableSentry  bool
}

// ErrorContext provides additional context for errors
type ErrorContext struct {
	UserID     string                 `json:"user_id,omitempty"`
	RequestID  string                 `json:"request_id,omitempty"`
	TraceID    string                 `json:"trace_id,omitempty"`
	SpanID     string                 `json:"span_id,omitempty"`
	Endpoint   string                 `json:"endpoint,omitempty"`
	Method     string                 `json:"method,omitempty"`
	StatusCode int                    `json:"status_code,omitempty"`
	Component  string                 `json:"component,omitempty"`
	Operation  string                 `json:"operation,omitempty"`
	Extra      map[string]interface{} `json:"extra,omitempty"`
}

// NewErrorProvider creates a new error provider
func NewErrorProvider(config ErrorConfig, logger *zap.Logger) (*ErrorProvider, error) {
	ep := &ErrorProvider{
		serviceName:   config.ServiceName,
		environment:   config.Environment,
		sentryEnabled: config.EnableSentry,
		logger:        logger,
	}

	// Initialize Sentry if enabled
	if config.EnableSentry && config.SentryDSN != "" {
		err := sentry.Init(sentry.ClientOptions{
			Dsn:              config.SentryDSN,
			Environment:      config.Environment,
			Release:          config.SentryRelease,
			TracesSampleRate: config.SampleRate,
			AttachStacktrace: true,
			BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
				// Add service context
				event.Tags["service"] = config.ServiceName
				event.Tags["environment"] = config.Environment
				return event
			},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to initialize Sentry: %w", err)
		}

		logger.Info("Sentry error tracking initialized",
			zap.String("service", config.ServiceName),
			zap.String("environment", config.Environment),
			zap.Float64("sample_rate", config.SampleRate),
		)
	} else {
		logger.Info("Error tracking initialized without Sentry",
			zap.String("service", config.ServiceName),
		)
	}

	return ep, nil
}

// CaptureError captures an error with context
func (ep *ErrorProvider) CaptureError(err error, context ErrorContext) string {
	if err == nil {
		return ""
	}

	// Log error locally
	ep.logError(err, context)

	// Send to Sentry if enabled
	if ep.sentryEnabled {
		return ep.sendToSentry(err, context)
	}

	return ""
}

// CaptureException captures an exception with message
func (ep *ErrorProvider) CaptureException(message string, context ErrorContext) string {
	// Log exception locally
	ep.logException(message, context)

	// Send to Sentry if enabled
	if ep.sentryEnabled {
		return ep.sendExceptionToSentry(message, context)
	}

	return ""
}

// CapturePanic captures a panic with recovery
func (ep *ErrorProvider) CapturePanic(context ErrorContext) {
	if r := recover(); r != nil {
		stackTrace := string(debug.Stack())

		// Log panic locally
		ep.logger.Error("Panic recovered",
			zap.Any("panic", r),
			zap.String("stack_trace", stackTrace),
			zap.String("service", ep.serviceName),
			zap.Any("context", context),
		)

		// Send to Sentry if enabled
		if ep.sentryEnabled {
			sentry.WithScope(func(scope *sentry.Scope) {
				ep.setSentryScope(scope, context)
				sentry.CaptureException(fmt.Errorf("panic: %v", r))
			})
		}

		// Re-panic to maintain normal panic behavior
		panic(r)
	}
}

// CaptureMessage captures a message with context
func (ep *ErrorProvider) CaptureMessage(message string, level string, context ErrorContext) string {
	// Log message locally
	switch level {
	case "debug":
		ep.logger.Debug(message, zap.Any("context", context))
	case "info":
		ep.logger.Info(message, zap.Any("context", context))
	case "warn":
		ep.logger.Warn(message, zap.Any("context", context))
	case "error":
		ep.logger.Error(message, zap.Any("context", context))
	default:
		ep.logger.Info(message, zap.Any("context", context))
	}

	// Send to Sentry if enabled
	if ep.sentryEnabled {
		return ep.sendMessageToSentry(message, level, context)
	}

	return ""
}

// SetUserContext sets user context for subsequent error captures
func (ep *ErrorProvider) SetUserContext(userID, email, username string) {
	if ep.sentryEnabled {
		sentry.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetUser(sentry.User{
				ID:       userID,
				Email:    email,
				Username: username,
			})
		})
	}
}

// AddBreadcrumb adds a breadcrumb for debugging
func (ep *ErrorProvider) AddBreadcrumb(message, category string, data map[string]interface{}) {
	if ep.sentryEnabled {
		sentry.AddBreadcrumb(&sentry.Breadcrumb{
			Message:   message,
			Category:  category,
			Data:      data,
			Level:     sentry.LevelInfo,
			Timestamp: time.Now(),
		})
	}

	ep.logger.Debug("Breadcrumb added",
		zap.String("message", message),
		zap.String("category", category),
		zap.Any("data", data),
	)
}

// Flush ensures all events are sent before shutdown
func (ep *ErrorProvider) Flush(timeout time.Duration) bool {
	if ep.sentryEnabled {
		return sentry.Flush(timeout)
	}
	return true
}

// logError logs error locally with structured logging
func (ep *ErrorProvider) logError(err error, context ErrorContext) {
	fields := []zap.Field{
		zap.Error(err),
		zap.String("service", ep.serviceName),
		zap.String("environment", ep.environment),
	}

	if context.UserID != "" {
		fields = append(fields, zap.String("user_id", context.UserID))
	}
	if context.RequestID != "" {
		fields = append(fields, zap.String("request_id", context.RequestID))
	}
	if context.TraceID != "" {
		fields = append(fields, zap.String("trace_id", context.TraceID))
	}
	if context.SpanID != "" {
		fields = append(fields, zap.String("span_id", context.SpanID))
	}
	if context.Endpoint != "" {
		fields = append(fields, zap.String("endpoint", context.Endpoint))
	}
	if context.Method != "" {
		fields = append(fields, zap.String("method", context.Method))
	}
	if context.StatusCode != 0 {
		fields = append(fields, zap.Int("status_code", context.StatusCode))
	}
	if context.Component != "" {
		fields = append(fields, zap.String("component", context.Component))
	}
	if context.Operation != "" {
		fields = append(fields, zap.String("operation", context.Operation))
	}
	if context.Extra != nil {
		fields = append(fields, zap.Any("extra", context.Extra))
	}

	ep.logger.Error("Error captured", fields...)
}

// logException logs exception locally
func (ep *ErrorProvider) logException(message string, context ErrorContext) {
	ep.logger.Error("Exception captured",
		zap.String("message", message),
		zap.String("service", ep.serviceName),
		zap.Any("context", context),
	)
}

// sendToSentry sends error to Sentry
func (ep *ErrorProvider) sendToSentry(err error, context ErrorContext) string {
	return string(sentry.WithScope(func(scope *sentry.Scope) {
		ep.setSentryScope(scope, context)
		sentry.CaptureException(err)
	}))
}

// sendExceptionToSentry sends exception to Sentry
func (ep *ErrorProvider) sendExceptionToSentry(message string, context ErrorContext) string {
	return string(sentry.WithScope(func(scope *sentry.Scope) {
		ep.setSentryScope(scope, context)
		sentry.CaptureException(fmt.Errorf(message))
	}))
}

// sendMessageToSentry sends message to Sentry
func (ep *ErrorProvider) sendMessageToSentry(message, level string, context ErrorContext) string {
	return string(sentry.WithScope(func(scope *sentry.Scope) {
		ep.setSentryScope(scope, context)

		var sentryLevel sentry.Level
		switch level {
		case "debug":
			sentryLevel = sentry.LevelDebug
		case "info":
			sentryLevel = sentry.LevelInfo
		case "warn":
			sentryLevel = sentry.LevelWarning
		case "error":
			sentryLevel = sentry.LevelError
		default:
			sentryLevel = sentry.LevelInfo
		}

		sentry.CaptureMessage(message)
		scope.SetLevel(sentryLevel)
	}))
}

// setSentryScope sets Sentry scope with context
func (ep *ErrorProvider) setSentryScope(scope *sentry.Scope, context ErrorContext) {
	if context.UserID != "" {
		scope.SetTag("user_id", context.UserID)
	}
	if context.RequestID != "" {
		scope.SetTag("request_id", context.RequestID)
	}
	if context.TraceID != "" {
		scope.SetTag("trace_id", context.TraceID)
	}
	if context.SpanID != "" {
		scope.SetTag("span_id", context.SpanID)
	}
	if context.Endpoint != "" {
		scope.SetTag("endpoint", context.Endpoint)
	}
	if context.Method != "" {
		scope.SetTag("method", context.Method)
	}
	if context.StatusCode != 0 {
		scope.SetTag("status_code", fmt.Sprintf("%d", context.StatusCode))
	}
	if context.Component != "" {
		scope.SetTag("component", context.Component)
	}
	if context.Operation != "" {
		scope.SetTag("operation", context.Operation)
	}
	if context.Extra != nil {
		for key, value := range context.Extra {
			scope.SetExtra(key, value)
		}
	}
}

// DefaultErrorConfig returns default error configuration
func DefaultErrorConfig(serviceName string) ErrorConfig {
	return ErrorConfig{
		ServiceName:   serviceName,
		Environment:   getEnv("ENVIRONMENT", "development"),
		SentryDSN:     getEnv("SENTRY_DSN", ""),
		SentryRelease: getEnv("SENTRY_RELEASE", serviceName+"@"+getEnv("SERVICE_VERSION", "1.0.0")),
		SampleRate:    1.0, // Sample all errors in development
		EnableSentry:  getEnv("SENTRY_DSN", "") != "",
	}
}
