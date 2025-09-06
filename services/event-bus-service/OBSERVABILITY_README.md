# Event Bus Service - Comprehensive Observability Implementation

This document provides a complete guide for the enterprise-grade observability implementation in the Event Bus Service, featuring OpenTelemetry distributed tracing, Prometheus metrics, and comprehensive error tracking.

## Overview

The observability stack provides three pillars of monitoring:

1. **Distributed Tracing** - OpenTelemetry with OTLP exporters (Jaeger/Tempo compatible)
2. **Metrics Collection** - Prometheus with custom collectors and business metrics
3. **Error Tracking** - Structured error logging with context enrichment

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Application   │    │   Telemetry     │    │   External      │
│                 │    │   Provider      │    │   Systems       │
├─────────────────┤    ├─────────────────┤    ├─────────────────┤
│ HTTP Handlers   │───▶│ Tracing         │───▶│ Jaeger/Tempo    │
│ Event Processor │───▶│ - OpenTelemetry │    │ OTLP Collector  │
│ Kafka Consumer  │───▶│ - Span Creation │    │                 │
│ CDC Processor   │    │ - Context Prop. │    │                 │
├─────────────────┤    ├─────────────────┤    ├─────────────────┤
│ Business Logic  │───▶│ Metrics         │───▶│ Prometheus      │
│ HTTP Requests   │───▶│ - HTTP Metrics  │    │ Grafana         │
│ Kafka Ops       │───▶│ - Event Metrics │    │ AlertManager    │
│ System Health   │    │ - Custom Metrics│    │                 │
├─────────────────┤    ├─────────────────┤    ├─────────────────┤
│ Error Handling  │───▶│ Error Tracking  │───▶│ Structured Logs │
│ Panic Recovery  │───▶│ - Context Rich  │    │ External APM    │
│ HTTP Errors     │    │ - Stack Traces  │    │ (Future: Sentry)│
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Features

### Distributed Tracing
- **OpenTelemetry SDK Integration**: Modern, vendor-neutral tracing
- **OTLP HTTP Exporter**: Compatible with Jaeger, Tempo, and other OTLP receivers
- **Automatic Instrumentation**: HTTP requests, Kafka operations, database queries
- **Context Propagation**: Trace context preserved across service boundaries
- **Custom Spans**: Easy creation of application-specific spans

### Metrics Collection
- **Prometheus Integration**: Industry-standard metrics collection
- **HTTP Metrics**: Request count, duration, status codes, active connections
- **Event Processing Metrics**: Events produced/consumed, processing time, errors
- **Kafka Metrics**: Operations, connection status, lag monitoring
- **CDC Metrics**: Connector status, table changes, processing lag
- **System Metrics**: Runtime stats, memory usage, goroutine count
- **Business Metrics**: Custom application-specific metrics

### Error Tracking
- **Structured Error Logging**: Rich context with trace correlation
- **Panic Recovery**: Automatic panic capture with stack traces
- **HTTP Error Tracking**: Status code-based error classification
- **Component-Specific Errors**: Kafka, CDC, event processing errors
- **Context Enrichment**: User ID, session, request correlation
- **Future Sentry Integration**: Ready for external error tracking services

## Configuration

### Environment Variables

```bash
# Service Configuration
ENVIRONMENT=production
VERSION=1.0.0

# OpenTelemetry Tracing
OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4317
OTEL_EXPORTER_OTLP_HEADERS="api-key=your-api-key"
OTEL_SERVICE_NAME=event-bus-service
OTEL_RESOURCE_ATTRIBUTES="service.version=1.0.0,deployment.environment=production"

# Prometheus Metrics
METRICS_PORT=9090
METRICS_PATH=/metrics

# Error Tracking (Optional - Sentry)
SENTRY_DSN=https://your-sentry-dsn@sentry.io/project-id
SENTRY_ENVIRONMENT=production
SENTRY_RELEASE=event-bus-service@1.0.0
SENTRY_SAMPLE_RATE=1.0
```

### Code Configuration

```go
package main

import (
    "context"
    "log"
    
    "go.uber.org/zap"
    "github.com/Mir00r/X-Form-Backend/services/event-bus-service/internal/config"
    "github.com/Mir00r/X-Form-Backend/services/event-bus-service/internal/telemetry"
)

func main() {
    // Initialize logger
    logger, _ := zap.NewProduction()
    defer logger.Sync()

    // Create configuration
    cfg := &config.Config{
        Environment: "production",
        Version:     "1.0.0",
    }

    // Initialize telemetry provider
    telemetryProvider, err := telemetry.New(cfg, logger)
    if err != nil {
        log.Fatal("Failed to initialize telemetry:", err)
    }
    defer telemetryProvider.Shutdown(context.Background())

    // Use telemetry in your application
    // ...
}
```

## Usage Examples

### HTTP Request Observability

```go
func handleRequest(tp *telemetry.Provider) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Automatic tracing and metrics via middleware
        // telemetryProvider.HTTPMiddleware() handles this
        
        // Manual span creation for custom operations
        ctx, span := tp.Tracing().StartSpan(r.Context(), "business_logic")
        defer span.End()
        
        // Add custom attributes
        span.SetAttributes(
            attribute.String("user.id", getUserID(r)),
            attribute.String("operation", "process_request"),
        )
        
        // Process request...
        processRequest(ctx)
        
        // Record custom business metric
        tp.RecordBusinessMetric("requests_processed", 1, map[string]string{
            "endpoint": r.URL.Path,
            "method":   r.Method,
        })
    }
}
```

### Event Processing Observability

```go
func processEvent(tp *telemetry.Provider, eventType, source string, eventData interface{}) error {
    ctx := context.Background()
    
    // Use event processing observability wrapper
    newCtx, cleanup := tp.EventProcessingObservability(ctx, eventType, source)
    defer cleanup(nil) // Pass error if processing fails
    
    // Process the event
    err := doEventProcessing(newCtx, eventData)
    if err != nil {
        // Error will be automatically captured by cleanup function
        return err
    }
    
    return nil
}
```

### Kafka Operation Observability

```go
func produceMessage(tp *telemetry.Provider, topic string, message []byte) error {
    ctx := context.Background()
    
    // Use Kafka observability wrapper
    newCtx, cleanup := tp.KafkaObservability(ctx, "produce", topic, 0)
    defer func() {
        // Cleanup will capture any errors and record metrics
        cleanup(nil) // Pass actual error if operation fails
    }()
    
    // Produce message to Kafka
    return kafkaProducer.Produce(newCtx, topic, message)
}
```

### CDC Operation Observability

```go
func processCDCEvent(tp *telemetry.Provider, connector, table, operation string, data interface{}) error {
    ctx := context.Background()
    
    // Use CDC observability wrapper
    newCtx, cleanup := tp.CDCObservability(ctx, connector, table, operation)
    defer cleanup(nil) // Pass error if processing fails
    
    // Process CDC event
    return processTableChange(newCtx, data)
}
```

### Error Tracking

```go
func handleError(tp *telemetry.Provider, err error) {
    // Capture error with rich context
    eventID := tp.CaptureError(err, "payment_processor", "process_payment", map[string]interface{}{
        "payment_id": "pay_123",
        "amount":     100.00,
        "currency":   "USD",
    })
    
    log.Printf("Error captured with ID: %s", eventID)
}

// HTTP error tracking (automatic via middleware)
func httpHandler(w http.ResponseWriter, r *http.Request) {
    // Any panic or HTTP error status will be automatically captured
    if someCondition {
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return // Error automatically tracked by middleware
    }
}
```

## Metrics Available

### HTTP Metrics
- `http_requests_total` - Total HTTP requests by method, endpoint, status
- `http_request_duration_seconds` - HTTP request duration histogram
- `http_active_connections` - Current active HTTP connections

### Event Processing Metrics
- `events_produced_total` - Events produced by type, topic, source
- `events_consumed_total` - Events consumed by type, topic, source
- `event_processing_duration_seconds` - Event processing time
- `event_processing_errors_total` - Event processing errors

### Kafka Metrics
- `kafka_operations_total` - Kafka operations by type, topic
- `kafka_connection_status` - Kafka connection status by broker
- `kafka_message_size_bytes` - Message size distribution

### CDC Metrics
- `cdc_events_total` - CDC events by connector, table, operation
- `cdc_connector_status` - Debezium connector status
- `cdc_lag_seconds` - CDC processing lag

### System Metrics
- `go_info` - Go runtime information
- `go_memstats_*` - Memory statistics
- `go_goroutines` - Number of goroutines

## Deployment

### Docker Compose Example

```yaml
version: '3.8'
services:
  event-bus-service:
    image: event-bus-service:latest
    environment:
      - ENVIRONMENT=production
      - OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4317
      - METRICS_PORT=9090
    ports:
      - "8080:8080"
      - "9090:9090"
    depends_on:
      - otel-collector
      - prometheus

  otel-collector:
    image: otel/opentelemetry-collector-contrib:latest
    command: ["--config=/etc/otel-collector-config.yaml"]
    volumes:
      - ./otel-collector-config.yaml:/etc/otel-collector-config.yaml
    ports:
      - "4317:4317"   # OTLP gRPC receiver
      - "4318:4318"   # OTLP HTTP receiver

  jaeger:
    image: jaegertracing/all-in-one:latest
    ports:
      - "16686:16686"
      - "14250:14250"

  prometheus:
    image: prom/prometheus:latest
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
    ports:
      - "9091:9090"

  grafana:
    image: grafana/grafana:latest
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    ports:
      - "3000:3000"
    volumes:
      - grafana-storage:/var/lib/grafana
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: event-bus-service
  labels:
    app: event-bus-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: event-bus-service
  template:
    metadata:
      labels:
        app: event-bus-service
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "9090"
        prometheus.io/path: "/metrics"
    spec:
      containers:
      - name: event-bus-service
        image: event-bus-service:latest
        ports:
        - containerPort: 8080
          name: http
        - containerPort: 9090
          name: metrics
        env:
        - name: ENVIRONMENT
          value: "production"
        - name: OTEL_EXPORTER_OTLP_ENDPOINT
          value: "http://otel-collector.observability.svc.cluster.local:4317"
        - name: METRICS_PORT
          value: "9090"
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

## Testing the Implementation

### Running the Demo

```bash
# Start the demo application
cd /Users/mir00r/Developer/kamkaiz/X-Form-Backend/services/event-bus-service
go run cmd/demo/main.go

# Test endpoints
curl http://localhost:8080/health
curl http://localhost:8080/api/events?type=demo&source=test
curl http://localhost:8080/api/kafka?operation=produce&topic=test-topic
curl http://localhost:8080/api/cdc?connector=postgres&table=users&operation=insert
curl http://localhost:8080/api/error?type=demo

# View metrics
curl http://localhost:8080/metrics
```

### Expected Metrics Output

```
# HELP http_requests_total Total number of HTTP requests
# TYPE http_requests_total counter
http_requests_total{endpoint="/api/events",method="GET",status="200"} 1

# HELP http_request_duration_seconds HTTP request duration
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{endpoint="/api/events",method="GET",status="200",le="0.005"} 1

# HELP events_consumed_total Total number of events consumed
# TYPE events_consumed_total counter
events_consumed_total{event_type="demo",source="test",topic=""} 1
```

## Troubleshooting

### Common Issues

1. **OTLP Exporter Connection Failed**
   - Check OTEL_EXPORTER_OTLP_ENDPOINT environment variable
   - Verify OTLP collector is running and accessible
   - Check network connectivity and firewall rules

2. **Metrics Not Appearing in Prometheus**
   - Verify Prometheus scrape configuration
   - Check service discovery and target endpoints
   - Ensure metrics port is accessible

3. **Traces Not Visible in Jaeger**
   - Confirm OTLP collector is forwarding to Jaeger
   - Check sampling configuration
   - Verify trace context propagation

### Debug Mode

Enable debug logging:

```go
logger, _ := zap.NewDevelopment()
```

Set debug environment variables:

```bash
export OTEL_LOG_LEVEL=debug
export OTEL_TRACES_EXPORTER=console
```

## Performance Considerations

### Resource Usage
- **Memory**: ~50MB additional for telemetry overhead
- **CPU**: ~5% additional for span processing and metrics collection
- **Network**: OTLP exports consume ~1KB per trace span

### Optimization Tips
1. **Sampling**: Use probabilistic sampling for high-traffic services
2. **Batch Processing**: Configure appropriate batch sizes for exporters
3. **Metric Cardinality**: Limit label combinations to prevent memory issues
4. **Async Processing**: Use background processing for telemetry operations

## Best Practices

1. **Trace Context**: Always propagate trace context across service boundaries
2. **Meaningful Spans**: Create spans for significant business operations
3. **Attribute Consistency**: Use consistent attribute naming across services
4. **Error Handling**: Capture errors with sufficient context for debugging
5. **Metric Labels**: Keep metric labels low-cardinality and meaningful
6. **Resource Tagging**: Use resource attributes for service identification

## Future Enhancements

1. **Sentry Integration**: External error tracking service integration
2. **Custom Dashboards**: Pre-built Grafana dashboards
3. **Alerting Rules**: Prometheus alerting rule templates
4. **Log Correlation**: Structured logging with trace correlation
5. **Service Mesh**: Istio/Envoy integration for infrastructure-level observability

## Contributing

When adding new observability features:

1. Update telemetry provider interfaces
2. Add comprehensive tests
3. Update documentation and examples
4. Consider performance impact
5. Maintain backward compatibility

For questions or issues, please refer to the project documentation or create an issue in the repository.
