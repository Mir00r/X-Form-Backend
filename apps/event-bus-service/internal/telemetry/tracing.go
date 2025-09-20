// Package telemetry provides comprehensive observability implementation
// This package implements OpenTelemetry tracing, Prometheus metrics, and Sentry error tracking
// following enterprise best practices for distributed systems monitoring.
package telemetry

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/Mir00r/X-Form-Backend/services/event-bus-service/internal/config"
)

// TracerProvider manages OpenTelemetry tracing configuration
type TracerProvider struct {
	config   *config.Config
	logger   *zap.Logger
	provider *sdktrace.TracerProvider
	tracer   trace.Tracer
}

// TracingConfig defines tracing configuration options
type TracingConfig struct {
	ServiceName    string            `json:"service_name"`
	ServiceVersion string            `json:"service_version"`
	Environment    string            `json:"environment"`
	SampleRate     float64           `json:"sample_rate"`
	JaegerEndpoint string            `json:"jaeger_endpoint"`
	OTLPEndpoint   string            `json:"otlp_endpoint"`
	Headers        map[string]string `json:"headers"`
	Attributes     map[string]string `json:"attributes"`
	BatchTimeout   time.Duration     `json:"batch_timeout"`
	MaxBatchSize   int               `json:"max_batch_size"`
	MaxQueueSize   int               `json:"max_queue_size"`
}

// SpanContext provides context information for spans
type SpanContext struct {
	TraceID    string            `json:"trace_id"`
	SpanID     string            `json:"span_id"`
	Operation  string            `json:"operation"`
	Component  string            `json:"component"`
	Tags       map[string]string `json:"tags"`
	StartTime  time.Time         `json:"start_time"`
	Duration   time.Duration     `json:"duration"`
	StatusCode string            `json:"status_code"`
	Error      error             `json:"error,omitempty"`
}

// NewTracerProvider creates a new OpenTelemetry tracer provider
func NewTracerProvider(cfg *config.Config, logger *zap.Logger) (*TracerProvider, error) {
	tp := &TracerProvider{
		config: cfg,
		logger: logger,
	}

	if err := tp.initializeTracing(); err != nil {
		return nil, fmt.Errorf("failed to initialize tracing: %w", err)
	}

	return tp, nil
}

// initializeTracing sets up OpenTelemetry tracing with multiple exporters
func (tp *TracerProvider) initializeTracing() error {
	// Create resource with service information
	res, err := tp.createResource()
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}

	// Create trace exporters
	exporters, err := tp.createExporters()
	if err != nil {
		return fmt.Errorf("failed to create exporters: %w", err)
	}

	// Create batch span processor options
	batchOptions := []sdktrace.BatchSpanProcessorOption{
		sdktrace.WithBatchTimeout(5 * time.Second),
		sdktrace.WithMaxExportBatchSize(512),
		sdktrace.WithMaxQueueSize(2048),
	}

	// Create trace provider with exporters
	var spanProcessors []sdktrace.SpanProcessor
	for _, exporter := range exporters {
		processor := sdktrace.NewBatchSpanProcessor(exporter, batchOptions...)
		spanProcessors = append(spanProcessors, processor)
	}

	// Configure sampling
	sampler := tp.createSampler()

	// Create and configure trace provider
	tp.provider = sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	// Add span processors
	for _, processor := range spanProcessors {
		tp.provider.RegisterSpanProcessor(processor)
	}

	// Set global provider and propagator
	otel.SetTracerProvider(tp.provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Create tracer instance
	tp.tracer = tp.provider.Tracer(
		"event-bus-service",
		trace.WithInstrumentationVersion("v1.0.0"),
		trace.WithSchemaURL(semconv.SchemaURL),
	)

	tp.logger.Info("OpenTelemetry tracing initialized successfully",
		zap.String("service_name", tp.config.Observability.Tracing.ServiceName),
		zap.Float64("sample_rate", tp.config.Observability.Tracing.SampleRate))

	return nil
}

// createResource creates OpenTelemetry resource with service metadata
func (tp *TracerProvider) createResource() (*resource.Resource, error) {
	hostname, _ := os.Hostname()

	attributes := []attribute.KeyValue{
		semconv.ServiceName(tp.config.Observability.Tracing.ServiceName),
		semconv.ServiceVersion(tp.config.Version),
		semconv.ServiceInstanceID(fmt.Sprintf("%s-%d", hostname, os.Getpid())),
		semconv.DeploymentEnvironment(tp.config.Environment),
		attribute.String("service.type", "event-bus"),
		attribute.String("service.framework", "go"),
		attribute.String("service.runtime", runtime.Version()),
		attribute.String("host.name", hostname),
		attribute.Int("process.pid", os.Getpid()),
	}

	// Add custom attributes from configuration
	for key, value := range tp.getCustomAttributes() {
		attributes = append(attributes, attribute.String(key, value))
	}

	return resource.NewWithAttributes(
		semconv.SchemaURL,
		attributes...,
	), nil
}

// createExporters creates trace exporters for different backends
func (tp *TracerProvider) createExporters() ([]sdktrace.SpanExporter, error) {
	var exporters []sdktrace.SpanExporter

	tracingConfig := tp.config.Observability.Tracing

	// Jaeger exporter
	if tracingConfig.Endpoint != "" && tp.isJaegerEndpoint(tracingConfig.Endpoint) {
		jaegerExp, err := tp.createJaegerExporter()
		if err != nil {
			tp.logger.Error("Failed to create Jaeger exporter", zap.Error(err))
		} else {
			exporters = append(exporters, jaegerExp)
			tp.logger.Info("Jaeger exporter created successfully")
		}
	}

	// OTLP exporter (for Tempo, OTEL Collector, etc.)
	if tp.hasOTLPConfig() {
		otlpExp, err := tp.createOTLPExporter()
		if err != nil {
			tp.logger.Error("Failed to create OTLP exporter", zap.Error(err))
		} else {
			exporters = append(exporters, otlpExp)
			tp.logger.Info("OTLP exporter created successfully")
		}
	}

	if len(exporters) == 0 {
		return nil, fmt.Errorf("no trace exporters configured")
	}

	return exporters, nil
}

// createJaegerExporter creates Jaeger trace exporter
func (tp *TracerProvider) createJaegerExporter() (sdktrace.SpanExporter, error) {
	// Use OTLP HTTP exporter as Jaeger supports OTLP protocol
	// This maintains compatibility while using modern standards
	return tp.createOTLPExporter()
}

// createOTLPExporter creates OTLP trace exporter for Tempo/OTEL Collector
func (tp *TracerProvider) createOTLPExporter() (sdktrace.SpanExporter, error) {
	ctx := context.Background()

	// Configure OTLP HTTP client options
	options := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(tp.getOTLPEndpoint()),
	}

	// Add headers if configured
	if headers := tp.getOTLPHeaders(); len(headers) > 0 {
		options = append(options, otlptracehttp.WithHeaders(headers))
	}

	// Configure authentication if needed
	if tp.isOTLPSecure() {
		// TLS configuration handled by the HTTP client
	} else {
		options = append(options, otlptracehttp.WithInsecure())
	}

	// Create OTLP HTTP client
	client := otlptracehttp.NewClient(options...)

	// Create and return OTLP exporter
	return otlptrace.New(ctx, client)
}

// createSampler creates trace sampler based on configuration
func (tp *TracerProvider) createSampler() sdktrace.Sampler {
	sampleRate := tp.config.Observability.Tracing.SampleRate

	if sampleRate <= 0 {
		return sdktrace.NeverSample()
	} else if sampleRate >= 1.0 {
		return sdktrace.AlwaysSample()
	}

	return sdktrace.TraceIDRatioBased(sampleRate)
}

// StartSpan starts a new trace span with comprehensive metadata
func (tp *TracerProvider) StartSpan(ctx context.Context, operationName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	// Add default attributes
	defaultOpts := []trace.SpanStartOption{
		trace.WithAttributes(
			attribute.String("service.name", tp.config.Observability.Tracing.ServiceName),
			attribute.String("service.version", tp.config.Version),
			attribute.String("operation.name", operationName),
		),
		trace.WithSpanKind(trace.SpanKindInternal),
	}

	// Merge with provided options
	allOpts := append(defaultOpts, opts...)

	return tp.tracer.Start(ctx, operationName, allOpts...)
}

// StartHTTPSpan starts a span for HTTP requests with automatic instrumentation
func (tp *TracerProvider) StartHTTPSpan(ctx context.Context, r *http.Request) (context.Context, trace.Span) {
	operationName := fmt.Sprintf("%s %s", r.Method, r.URL.Path)

	spanCtx, span := tp.tracer.Start(ctx, operationName,
		trace.WithAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.target", r.URL.Path),
			attribute.String("http.url", r.URL.String()),
			attribute.String("http.scheme", r.URL.Scheme),
			attribute.String("http.host", r.Host),
			attribute.String("http.user_agent", r.UserAgent()),
			attribute.Int64("http.request_content_length", r.ContentLength),
		),
		trace.WithSpanKind(trace.SpanKindServer),
	)

	return spanCtx, span
}

// StartKafkaProducerSpan starts a span for Kafka producer operations
func (tp *TracerProvider) StartKafkaProducerSpan(ctx context.Context, topic string, partition int32) (context.Context, trace.Span) {
	operationName := fmt.Sprintf("kafka.produce %s", topic)

	spanCtx, span := tp.tracer.Start(ctx, operationName,
		trace.WithAttributes(
			attribute.String("messaging.system", "kafka"),
			attribute.String("messaging.destination.name", topic),
			attribute.String("messaging.operation", "publish"),
			attribute.Int("messaging.kafka.partition", int(partition)),
		),
		trace.WithSpanKind(trace.SpanKindProducer),
	)

	return spanCtx, span
}

// StartKafkaConsumerSpan starts a span for Kafka consumer operations
func (tp *TracerProvider) StartKafkaConsumerSpan(ctx context.Context, topic string, partition int32, offset int64) (context.Context, trace.Span) {
	operationName := fmt.Sprintf("kafka.consume %s", topic)

	spanCtx, span := tp.tracer.Start(ctx, operationName,
		trace.WithAttributes(
			attribute.String("messaging.system", "kafka"),
			attribute.String("messaging.destination.name", topic),
			attribute.String("messaging.operation", "receive"),
			attribute.Int("messaging.kafka.partition", int(partition)),
			attribute.Int64("messaging.kafka.offset", offset),
		),
		trace.WithSpanKind(trace.SpanKindConsumer),
	)

	return spanCtx, span
}

// StartDatabaseSpan starts a span for database operations
func (tp *TracerProvider) StartDatabaseSpan(ctx context.Context, operation, table string) (context.Context, trace.Span) {
	operationName := fmt.Sprintf("db.%s %s", operation, table)

	spanCtx, span := tp.tracer.Start(ctx, operationName,
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", operation),
			attribute.String("db.name", table),
		),
		trace.WithSpanKind(trace.SpanKindClient),
	)

	return spanCtx, span
}

// RecordError records an error in the current span
func (tp *TracerProvider) RecordError(span trace.Span, err error, options ...trace.EventOption) {
	if err == nil {
		return
	}

	span.RecordError(err, options...)
	span.SetStatus(codes.Error, err.Error())
}

// AddEvent adds an event to the current span
func (tp *TracerProvider) AddEvent(span trace.Span, name string, attributes ...attribute.KeyValue) {
	span.AddEvent(name, trace.WithAttributes(attributes...))
}

// SetAttributes sets attributes on the current span
func (tp *TracerProvider) SetAttributes(span trace.Span, attributes ...attribute.KeyValue) {
	span.SetAttributes(attributes...)
}

// GetTraceID extracts trace ID from context
func (tp *TracerProvider) GetTraceID(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.IsValid() {
		return spanCtx.TraceID().String()
	}
	return ""
}

// GetSpanID extracts span ID from context
func (tp *TracerProvider) GetSpanID(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.IsValid() {
		return spanCtx.SpanID().String()
	}
	return ""
}

// CreateSpanContext creates span context for logging and monitoring
func (tp *TracerProvider) CreateSpanContext(ctx context.Context, operation string) *SpanContext {
	span := trace.SpanFromContext(ctx)
	spanCtx := span.SpanContext()

	return &SpanContext{
		TraceID:   spanCtx.TraceID().String(),
		SpanID:    spanCtx.SpanID().String(),
		Operation: operation,
		StartTime: time.Now(),
	}
}

// HTTPMiddleware provides HTTP tracing middleware
func (tp *TracerProvider) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tp.StartHTTPSpan(r.Context(), r)
		defer span.End()

		// Create response writer wrapper to capture status code
		wrappedWriter := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Add trace ID to response headers
		traceID := tp.GetTraceID(ctx)
		if traceID != "" {
			w.Header().Set("X-Trace-ID", traceID)
		}

		// Execute request with tracing context
		next.ServeHTTP(wrappedWriter, r.WithContext(ctx))

		// Set final span attributes
		span.SetAttributes(
			semconv.HTTPStatusCode(wrappedWriter.statusCode),
			semconv.HTTPResponseContentLength(int(wrappedWriter.bytesWritten)),
		)

		// Set span status based on HTTP status code
		if wrappedWriter.statusCode >= 400 {
			span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", wrappedWriter.statusCode))
		}
	})
}

// Shutdown gracefully shuts down the tracer provider
func (tp *TracerProvider) Shutdown(ctx context.Context) error {
	if tp.provider == nil {
		return nil
	}

	tp.logger.Info("Shutting down tracing provider")
	return tp.provider.Shutdown(ctx)
}

// Helper methods

func (tp *TracerProvider) isJaegerEndpoint(endpoint string) bool {
	return strings.Contains(endpoint, "jaeger") || strings.Contains(endpoint, ":14268")
}

func (tp *TracerProvider) hasOTLPConfig() bool {
	return tp.getOTLPEndpoint() != ""
}

func (tp *TracerProvider) getOTLPEndpoint() string {
	// Check for OTLP-specific endpoint in environment or config
	if endpoint := os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT"); endpoint != "" {
		return endpoint
	}
	if endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"); endpoint != "" {
		return endpoint + "/v1/traces"
	}

	// Default OTLP endpoint configuration
	tracingConfig := tp.config.Observability.Tracing
	if !tp.isJaegerEndpoint(tracingConfig.Endpoint) {
		return tracingConfig.Endpoint
	}

	return ""
}

func (tp *TracerProvider) getOTLPHeaders() map[string]string {
	headers := make(map[string]string)

	// Add headers from environment
	if auth := os.Getenv("OTEL_EXPORTER_OTLP_HEADERS"); auth != "" {
		// Parse header string (format: "key1=value1,key2=value2")
		for _, header := range strings.Split(auth, ",") {
			parts := strings.SplitN(header, "=", 2)
			if len(parts) == 2 {
				headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}

	return headers
}

func (tp *TracerProvider) isOTLPSecure() bool {
	endpoint := tp.getOTLPEndpoint()
	return strings.HasPrefix(endpoint, "https://")
}

func (tp *TracerProvider) getCustomAttributes() map[string]string {
	attributes := make(map[string]string)

	// Add Kubernetes attributes if available
	if namespace := os.Getenv("KUBERNETES_NAMESPACE"); namespace != "" {
		attributes["k8s.namespace.name"] = namespace
	}
	if podName := os.Getenv("KUBERNETES_POD_NAME"); podName != "" {
		attributes["k8s.pod.name"] = podName
	}
	if nodeName := os.Getenv("KUBERNETES_NODE_NAME"); nodeName != "" {
		attributes["k8s.node.name"] = nodeName
	}

	// Add Docker attributes if available
	if containerID := os.Getenv("HOSTNAME"); containerID != "" {
		attributes["container.id"] = containerID
	}

	return attributes
}

// responseWriter wraps http.ResponseWriter to capture response details
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += int64(n)
	return n, err
}
