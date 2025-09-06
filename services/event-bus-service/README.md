# Event Bus Service

A comprehensive Enterprise-grade Event Bus and Change Data Capture (CDC) service built with Go, Apache Kafka, and Debezium. This service provides real-time event streaming, database change capture, and event processing capabilities for the X-Form microservices ecosystem.

## 🚀 Features

### Core Capabilities
- **Event Streaming**: High-throughput event publishing and consumption using Apache Kafka
- **Change Data Capture**: Real-time database change capture using Debezium for PostgreSQL WAL
- **Event Processing**: Multi-processor pipeline with transformations, filtering, and routing
- **Microservices Integration**: Seamless integration with existing X-Form services
- **Enterprise Security**: JWT authentication, SASL/TLS support, event signing
- **Observability**: Comprehensive metrics, logging, and distributed tracing
- **High Availability**: Circuit breakers, health checks, graceful shutdown

### Event Types Supported
- **CDC Events**: Database change events from PostgreSQL WAL
- **Application Events**: Form lifecycle, response handling, analytics
- **System Events**: Service health, metrics, audit logs

### Technical Stack
- **Language**: Go 1.21+
- **Event Streaming**: Apache Kafka with Sarama client
- **CDC**: Debezium PostgreSQL connector
- **Configuration**: Viper with YAML/Environment variables
- **Observability**: Prometheus metrics, Zap logging
- **Security**: JWT, SASL/TLS, AES encryption
- **Caching**: Redis for metadata and rate limiting

## 📋 Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- Apache Kafka 3.5+
- PostgreSQL 15+ with logical replication
- Redis 7+
- Debezium Connect 2.4+

## 🛠️ Installation

### Local Development with Docker Compose

1. **Clone the repository**:
```bash
git clone <repository-url>
cd services/event-bus-service
```

2. **Start all services**:
```bash
docker-compose up -d
```

This will start:
- Event Bus Service (port 8080)
- PostgreSQL (port 5432)
- Redis (port 6379)
- Apache Kafka (port 9092)
- Debezium Connect (port 8083)
- Kafka UI (port 8081)
- Prometheus (port 9091)
- Grafana (port 3000)

3. **Verify installation**:
```bash
# Check service health
curl http://localhost:8080/health

# Check Kafka UI
open http://localhost:8081

# Check Grafana
open http://localhost:3000 (admin/admin)
```

### Manual Installation

1. **Install dependencies**:
```bash
go mod download
```

2. **Configure the service**:
```bash
cp config/config.yaml config/local.yaml
# Edit config/local.yaml with your settings
```

3. **Start external dependencies**:
```bash
# Start PostgreSQL, Kafka, Redis, Debezium separately
```

4. **Run the service**:
```bash
go run cmd/server/main.go
```

## ⚙️ Configuration

### Environment Variables

The service supports configuration via environment variables that override YAML settings:

```bash
# Server Configuration
export SERVER_HOST=0.0.0.0
export SERVER_PORT=8080

# Kafka Configuration
export KAFKA_BROKERS=localhost:9092
export KAFKA_CLIENT_ID=event-bus-service
export KAFKA_GROUP_ID=event-bus-group

# Database Configuration
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=eventbus
export DB_USER=eventbus
export DB_PASSWORD=eventbus_password

# Redis Configuration
export REDIS_HOST=localhost
export REDIS_PORT=6379

# Debezium Configuration
export DEBEZIUM_CONNECT_URL=http://localhost:8083

# Observability
export METRICS_ENABLED=true
export METRICS_PORT=9090
export LOG_LEVEL=info
export LOG_FORMAT=json
```

### Configuration File

The service uses a YAML configuration file located at `config/config.yaml`. See the file for detailed configuration options.

## 🔌 API Endpoints

### Health and Monitoring

- `GET /health` - Service health check
- `GET /version` - Service version information
- `GET /metrics` - Prometheus metrics (port 9090)

### Event Publishing

- `POST /events` - Publish application events

**Example Request**:
```bash
curl -X POST http://localhost:8080/events \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "form.created",
    "source": "form-service",
    "data": {
      "form_id": "f123",
      "title": "User Survey",
      "created_by": "user123"
    },
    "topic": "app.form.created",
    "key": "f123"
  }'
```

### Administration

- `GET /admin/config` - Get sanitized configuration

## 📊 Event Processing

### CDC Events

The service automatically captures database changes from PostgreSQL tables:

```sql
-- Tables monitored for CDC
public.forms
public.responses  
public.analytics
```

CDC events are published to topics with the pattern: `cdc.{table_name}`

### Application Events

Application events follow a structured format:

```json
{
  "id": "event_1234567890",
  "event_type": "form.created",
  "source": "form-service",
  "data": {
    "form_id": "f123",
    "title": "User Survey"
  },
  "metadata": {
    "timestamp": "2024-01-01T12:00:00Z",
    "version": "1.0",
    "content_type": "application/json"
  }
}
```

### Event Processors

The service includes specialized processors:

- **CDCEventProcessor**: Processes database change events
- **FormEventProcessor**: Handles form lifecycle events
- **ResponseEventProcessor**: Manages response events
- **AnalyticsEventProcessor**: Processes analytics data

## 🔍 Monitoring and Observability

### Metrics

Prometheus metrics are available at `http://localhost:9090/metrics`:

- `event_bus_events_total` - Total events processed
- `event_bus_processing_duration_seconds` - Event processing latency
- `event_bus_kafka_operations_total` - Kafka operation metrics
- `event_bus_debezium_connector_status` - Debezium connector health

### Logging

Structured logging with configurable levels and formats:

```json
{
  "level": "info",
  "timestamp": "2024-01-01T12:00:00Z",
  "message": "Event processed successfully",
  "event_id": "event_123",
  "processor": "form_processor",
  "duration_ms": 150
}
```

### Health Checks

Comprehensive health checks for all components:

```bash
curl http://localhost:8080/health
```

Returns:
```json
{
  "success": true,
  "status": "healthy",
  "components": {
    "kafka": {"status": "healthy"},
    "debezium": {"status": "healthy"},
    "database": {"status": "healthy"},
    "redis": {"status": "healthy"}
  }
}
```

## 🔒 Security

### Authentication and Authorization

- JWT-based authentication for API endpoints
- Service-to-service authentication via tokens
- Role-based access control for admin endpoints

### Encryption

- TLS encryption for all external communications
- SASL authentication for Kafka connections
- Event payload encryption for sensitive data

### Rate Limiting

- Configurable rate limiting per endpoint
- Redis-backed rate limit store
- Circuit breakers for external dependencies

## 🚀 Deployment

### Docker

```bash
# Build image
docker build -t event-bus-service:latest .

# Run container
docker run -p 8080:8080 -p 9090:9090 event-bus-service:latest
```

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: event-bus-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: event-bus-service
  template:
    metadata:
      labels:
        app: event-bus-service
    spec:
      containers:
      - name: event-bus-service
        image: event-bus-service:latest
        ports:
        - containerPort: 8080
        - containerPort: 9090
        env:
        - name: KAFKA_BROKERS
          value: "kafka:9092"
        # Add other environment variables
```

## 🧪 Testing

### Unit Tests

```bash
go test ./...
```

### Integration Tests

```bash
# Start test environment
docker-compose -f docker-compose.test.yml up -d

# Run integration tests
go test -tags=integration ./...
```

### Load Testing

```bash
# Install artillery
npm install -g artillery

# Run load tests
artillery run test/load-test.yml
```

## 📚 Development

### Project Structure

```
event-bus-service/
├── cmd/
│   └── server/           # Main application
├── internal/
│   ├── config/          # Configuration management
│   ├── kafka/           # Kafka client implementation
│   ├── debezium/        # Debezium integration
│   ├── events/          # Event structures and types
│   ├── processors/      # Event processing pipeline
│   └── handlers/        # HTTP request handlers
├── config/              # Configuration files
├── scripts/            # Setup and utility scripts
├── test/               # Test files and fixtures
├── monitoring/         # Prometheus and Grafana configs
└── docs/               # Additional documentation
```

### Building from Source

```bash
# Download dependencies
go mod download

# Build binary
go build -o event-bus-service cmd/server/main.go

# Run service
./event-bus-service
```

### Code Quality

```bash
# Format code
go fmt ./...

# Lint code
golangci-lint run

# Run security checks
gosec ./...
```

## 🤝 Integration with X-Form Services

### Form Service Integration

The Event Bus Service integrates with the Form Service to capture form lifecycle events:

```go
// Publishing form events
eventData := map[string]interface{}{
    "form_id": "f123",
    "title": "User Survey", 
    "status": "active",
}

err := eventBus.PublishEvent("form.created", "form-service", eventData)
```

### Response Service Integration

Captures response submission and processing events:

```go
// Publishing response events
responseData := map[string]interface{}{
    "response_id": "r456",
    "form_id": "f123",
    "submitted_by": "user789",
}

err := eventBus.PublishEvent("response.submitted", "response-service", responseData)
```

### Analytics Service Integration

Processes analytics events and aggregations:

```go
// Publishing analytics events
analyticsData := map[string]interface{}{
    "event_type": "form_view",
    "form_id": "f123",
    "user_id": "user789",
    "timestamp": time.Now(),
}

err := eventBus.PublishEvent("analytics.event", "analytics-service", analyticsData)
```

## 🔧 Troubleshooting

### Common Issues

**Kafka Connection Issues**:
```bash
# Check Kafka connectivity
curl http://localhost:8083/connectors

# Verify Kafka topics
docker exec -it event-bus-kafka kafka-topics --bootstrap-server localhost:9092 --list
```

**Debezium Connector Issues**:
```bash
# Check connector status
curl http://localhost:8083/connectors/postgres-connector/status

# Restart connector
curl -X POST http://localhost:8083/connectors/postgres-connector/restart
```

**Database Connection Issues**:
```bash
# Test database connectivity
psql -h localhost -p 5432 -U eventbus -d eventbus -c "SELECT 1"
```

### Debugging

Enable debug logging:
```bash
export LOG_LEVEL=debug
```

Check service logs:
```bash
docker logs event-bus-service -f
```

### Performance Tuning

- Adjust Kafka producer/consumer settings
- Configure event processing workers
- Tune database connection pools
- Optimize Redis cache settings

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📞 Support

For support and questions:

- Create an issue in the repository
- Check the documentation in the `docs/` directory
- Review the troubleshooting section above

## 🔄 Roadmap

- [ ] Support for AWS EventBridge
- [ ] Schema registry integration
- [ ] Event replay capabilities
- [ ] Advanced event filtering
- [ ] Multi-tenant support
- [ ] GraphQL subscription support
- [ ] Stream processing with Apache Flink
