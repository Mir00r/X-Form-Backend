# X-Form-Backend Observability Implementation Guide

## Overview

This document provides a comprehensive guide to the observability implementation across the X-Form-Backend microservices architecture. The implementation includes distributed tracing, metrics collection, error tracking, and comprehensive monitoring.

## Architecture Components

### 1. Shared Observability Package

Location: `shared/observability/`

The shared package provides a unified observability solution across all microservices:

- **Tracing (`tracing.go`)**: OpenTelemetry distributed tracing with OTLP HTTP exporters
- **Metrics (`metrics.go`)**: Prometheus metrics collection with 15+ metric types
- **Error Tracking (`errors.go`)**: Sentry integration with structured logging
- **Main Provider (`observability.go`)**: Unified interface combining all observability features
- **Gin Middleware (`gin.go`)**: HTTP-specific middleware for automatic instrumentation

### 2. Infrastructure Components

Location: `infrastructure/`

- **observability-infrastructure.yml**: Complete Docker Compose stack
- **otel-collector-config.yaml**: OpenTelemetry Collector configuration
- **prometheus.yml**: Prometheus scraping configuration
- **alert_rules.yml**: Comprehensive alerting rules

## Implementation Details

### Distributed Tracing

- **Provider**: OpenTelemetry v1.31.0
- **Exporters**: OTLP HTTP to Jaeger and Tempo
- **Features**:
  - Automatic span creation for HTTP requests
  - Context propagation across service boundaries
  - Custom span attributes and events
  - Sampling configuration
  - Resource identification

### Metrics Collection

- **Provider**: Prometheus
- **Metric Types**:
  - HTTP request metrics (count, duration, size)
  - Service-level metrics (uptime, health)
  - Database metrics (connections, query duration)
  - External service metrics (response times, errors)
  - Business metrics (form submissions, user actions)

### Error Tracking

- **Provider**: Sentry integration ready
- **Features**:
  - Structured error logging
  - Context enrichment
  - Panic recovery
  - User context tracking
  - Breadcrumb trails

## Service Integration Status

### ✅ API Gateway (Completed)

Location: `services/api-gateway/`

**Integration Points**:
- Main application observability initialization
- Gin middleware for HTTP observability
- Proxy request observability with trace propagation
- External service call instrumentation

**Key Features**:
- Automatic trace context propagation to downstream services
- Request/response metrics collection
- Error tracking and logging
- Service proxy observability

**Usage**:
```go
// In main.go
obsProvider := observability.NewProvider(config.ServiceName)
defer obsProvider.Shutdown()

router.Use(obsProvider.GinMiddleware())

// In gateway.go
gateway := NewGateway(httpClient, obsProvider)
```

### 🔄 Remaining Services (Implementation Needed)

The following services need observability integration:

1. **Auth Service** (`services/auth-service/`)
2. **Form Service** (`services/form-service/`)
3. **Response Service** (`services/response-service/`)
4. **Collaboration Service** (`services/collaboration-service/`)
5. **Realtime Service** (`services/realtime-service/`)
6. **Analytics Service** (`services/analytics-service/`)

## Deployment Guide

### 1. Start Observability Infrastructure

```bash
# From project root
docker-compose -f infrastructure/observability-infrastructure.yml up -d
```

This starts:
- OpenTelemetry Collector (port 4318)
- Jaeger UI (port 16686)
- Tempo (port 3200)
- Prometheus (port 9090)
- Grafana (port 3000)
- AlertManager (port 9093)

### 2. Configure Services

Each service needs the following environment variables:

```bash
# Observability Configuration
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
OTEL_SERVICE_NAME=<service-name>
OTEL_SERVICE_VERSION=1.0.0
OTEL_ENVIRONMENT=development

# Sentry Configuration (optional)
SENTRY_DSN=<your-sentry-dsn>

# Prometheus Configuration
PROMETHEUS_ENABLED=true
PROMETHEUS_PORT=9090
```

### 3. Access Monitoring Dashboards

- **Jaeger UI**: http://localhost:16686 (Distributed tracing)
- **Grafana**: http://localhost:3000 (Metrics dashboards)
- **Prometheus**: http://localhost:9090 (Raw metrics)
- **AlertManager**: http://localhost:9093 (Alert management)

## Integration Guide for Remaining Services

### For Go Services (Form Service, Collaboration Service)

1. **Update go.mod**:
```bash
cd services/<service-name>
go mod edit -require=github.com/kamkaiz/x-form-backend/shared@v0.0.0
go mod edit -replace=github.com/kamkaiz/x-form-backend/shared=../../shared
go mod tidy
```

2. **Update main.go**:
```go
import "github.com/kamkaiz/x-form-backend/shared/observability"

func main() {
    // Initialize observability
    obsProvider := observability.NewProvider("service-name")
    defer obsProvider.Shutdown()
    
    // Add middleware
    router.Use(obsProvider.GinMiddleware())
    
    // Start server
    router.Run(":8080")
}
```

### For Node.js Services (Auth Service, Response Service, Realtime Service)

1. **Install dependencies**:
```bash
npm install @opentelemetry/api @opentelemetry/sdk-node @opentelemetry/auto-instrumentations-node
npm install prom-client @sentry/node
```

2. **Create observability module** similar to the Go implementation
3. **Integrate with Express/Fastify middleware**

### For Python Services (Analytics Service)

1. **Install dependencies**:
```bash
pip install opentelemetry-api opentelemetry-sdk opentelemetry-exporter-otlp
pip install prometheus-client sentry-sdk
```

2. **Create observability module** similar to the Go implementation
3. **Integrate with FastAPI/Flask middleware**

## Monitoring and Alerting

### Key Metrics to Monitor

1. **Service Health**:
   - Service uptime
   - Response times (p50, p95, p99)
   - Error rates
   - Request throughput

2. **Infrastructure**:
   - CPU and memory usage
   - Network I/O
   - Database connections
   - Cache hit rates

3. **Business Metrics**:
   - Form submission rates
   - User registration rates
   - API usage patterns
   - Feature adoption

### Alerting Rules

The implementation includes comprehensive alerting rules for:
- High error rates (>5% for 5 minutes)
- High response times (>1s p95 for 5 minutes)
- Service downtime
- Infrastructure resource exhaustion

## Trace Context Propagation

The implementation ensures trace context propagation across:

1. **HTTP Headers**: Automatic propagation via OpenTelemetry
2. **Custom Headers**: X-Trace-ID and X-Span-ID for debugging
3. **Database Queries**: Trace context in database spans
4. **External API Calls**: Context propagation to external services
5. **Message Queues**: Context propagation through async operations

## Best Practices

### 1. Span Naming
- Use descriptive, consistent span names
- Include operation type (e.g., "http.request", "db.query")
- Use service and operation prefixes

### 2. Metric Labels
- Keep cardinality low to avoid Prometheus performance issues
- Use consistent label names across services
- Include service, method, and status code labels

### 3. Error Handling
- Always capture errors in spans
- Include error context and stack traces
- Use structured logging with trace correlation

### 4. Performance
- Use sampling for high-traffic services
- Batch metric updates where possible
- Monitor observability overhead

## Troubleshooting

### Common Issues

1. **Missing Traces**: Check OTLP endpoint configuration
2. **High Cardinality Metrics**: Review metric labels
3. **Performance Impact**: Adjust sampling rates
4. **Service Discovery**: Ensure Prometheus can reach service endpoints

### Debug Commands

```bash
# Check service metrics endpoint
curl http://localhost:8080/metrics

# Verify OTLP collector
curl http://localhost:4318/v1/traces

# Check service health
curl http://localhost:8080/health
```

## Next Steps

1. **Complete Service Integration**: Implement observability in remaining services
2. **Custom Dashboards**: Create service-specific Grafana dashboards
3. **Advanced Alerting**: Implement more sophisticated alerting rules
4. **Log Aggregation**: Add centralized logging with ELK or similar
5. **SLA/SLO Monitoring**: Implement service level objectives
6. **Cost Optimization**: Optimize observability data retention and sampling

## Configuration Reference

### Environment Variables

```bash
# Required
OTEL_SERVICE_NAME=service-name
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318

# Optional
OTEL_SERVICE_VERSION=1.0.0
OTEL_ENVIRONMENT=development
OTEL_SAMPLING_RATIO=1.0
PROMETHEUS_ENABLED=true
PROMETHEUS_PORT=9090
SENTRY_DSN=your-sentry-dsn
LOG_LEVEL=info
```

### Docker Compose Integration

```yaml
version: '3.8'
services:
  api-gateway:
    build: .
    environment:
      - OTEL_SERVICE_NAME=api-gateway
      - OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4318
    depends_on:
      - otel-collector
```

This observability implementation provides comprehensive monitoring, tracing, and error tracking across the entire X-Form-Backend microservices architecture, enabling better debugging, performance optimization, and operational insights.
