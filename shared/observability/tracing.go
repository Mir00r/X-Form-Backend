package observability

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	oteltrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// TracingProvider handles OpenTelemetry tracing configuration
type TracingProvider struct {
	tracer   oteltrace.Tracer
	provider *trace.TracerProvider
	logger   *zap.Logger
}

// TracingConfig holds configuration for tracing
type TracingConfig struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	OTLPEndpoint   string
	OTLPHeaders    map[string]string
	SamplingRatio  float64
	EnableConsole  bool
}

// NewTracingProvider creates a new tracing provider
func NewTracingProvider(config TracingConfig, logger *zap.Logger) (*TracingProvider, error) {
	// Create resource with service information
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion(config.ServiceVersion),
			semconv.DeploymentEnvironment(config.Environment),
			attribute.String("telemetry.sdk.language", "go"),
			attribute.String("telemetry.sdk.name", "opentelemetry"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create OTLP HTTP exporter
	otlpExporter, err := otlptracehttp.New(
		context.Background(),
		otlptracehttp.WithEndpoint(config.OTLPEndpoint),
		otlptracehttp.WithHeaders(config.OTLPHeaders),
		otlptracehttp.WithInsecure(), // Use HTTPS in production
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create trace provider with batch processor
	tp := trace.NewTracerProvider(
		trace.WithResource(res),
		trace.WithBatcher(otlpExporter,
			trace.WithBatchTimeout(time.Second*5),
			trace.WithMaxExportBatchSize(512),
		),
		trace.WithSampler(trace.TraceIDRatioBased(config.SamplingRatio)),
	)

	// Set global trace provider
	otel.SetTracerProvider(tp)

	// Set global propagator for trace context propagation
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Create tracer
	tracer := tp.Tracer(config.ServiceName)

	logger.Info("Tracing provider initialized",
		zap.String("service", config.ServiceName),
		zap.String("environment", config.Environment),
		zap.String("otlp_endpoint", config.OTLPEndpoint),
		zap.Float64("sampling_ratio", config.SamplingRatio),
	)

	return &TracingProvider{
		tracer:   tracer,
		provider: tp,
		logger:   logger,
	}, nil
}

// StartSpan creates a new span with the given name and options
func (t *TracingProvider) StartSpan(ctx context.Context, spanName string, opts ...oteltrace.SpanStartOption) (context.Context, oteltrace.Span) {
	return t.tracer.Start(ctx, spanName, opts...)
}

// AddSpanAttributes adds attributes to the current span
func (t *TracingProvider) AddSpanAttributes(span oteltrace.Span, attrs ...attribute.KeyValue) {
	span.SetAttributes(attrs...)
}

// AddSpanEvent adds an event to the current span
func (t *TracingProvider) AddSpanEvent(span oteltrace.Span, name string, attrs ...attribute.KeyValue) {
	span.AddEvent(name, oteltrace.WithAttributes(attrs...))
}

// RecordError records an error in the current span
func (t *TracingProvider) RecordError(span oteltrace.Span, err error, attrs ...attribute.KeyValue) {
	span.RecordError(err, oteltrace.WithAttributes(attrs...))
	span.SetStatus(oteltrace.StatusCodeError, err.Error())
}

// GetTracer returns the tracer instance
func (t *TracingProvider) GetTracer() oteltrace.Tracer {
	return t.tracer
}

// ExtractTraceID extracts trace ID from context
func (t *TracingProvider) ExtractTraceID(ctx context.Context) string {
	span := oteltrace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

// ExtractSpanID extracts span ID from context
func (t *TracingProvider) ExtractSpanID(ctx context.Context) string {
	span := oteltrace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return span.SpanContext().SpanID().String()
	}
	return ""
}

// Shutdown gracefully shuts down the tracing provider
func (t *TracingProvider) Shutdown(ctx context.Context) error {
	t.logger.Info("Shutting down tracing provider")
	return t.provider.Shutdown(ctx)
}

// DefaultTracingConfig returns default tracing configuration
func DefaultTracingConfig(serviceName string) TracingConfig {
	return TracingConfig{
		ServiceName:    serviceName,
		ServiceVersion: getEnv("SERVICE_VERSION", "1.0.0"),
		Environment:    getEnv("ENVIRONMENT", "development"),
		OTLPEndpoint:   getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318"),
		OTLPHeaders: map[string]string{
			"Content-Type": "application/json",
		},
		SamplingRatio: 1.0, // Sample all traces in development
		EnableConsole: getEnv("ENVIRONMENT", "development") == "development",
	}
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
