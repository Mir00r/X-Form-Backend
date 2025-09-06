# X-Form API Gateway

A high-performance, feature-rich API Gateway for the X-Form microservices architecture built with Go and Gin framework.

## Features

- **🚀 High Performance**: Built with Go and Gin for maximum throughput
- **🔐 Authentication & Authorization**: JWT-based authentication with role-based access control
- **📊 Monitoring & Metrics**: Prometheus metrics, structured logging, and health checks
- **🛡️ Security**: CORS, security headers, rate limiting, and request validation
- **📖 API Documentation**: Comprehensive Swagger/OpenAPI 3.0 documentation
- **🔄 Load Balancing**: Service discovery and load balancing across microservices
- **📈 Observability**: Request tracing, logging, and performance monitoring
- **🐳 Container Ready**: Docker support with multi-stage builds
- **🎯 Microservice Routing**: Intelligent routing to backend services

## Architecture

The API Gateway serves as the single entry point for all client requests, routing them to appropriate microservices:

```
┌─────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Client    │───▶│   API Gateway   │───▶│  Microservices  │
│ (Web/Mobile)│    │                 │    │                 │
└─────────────┘    └─────────────────┘    └─────────────────┘
                           │
                           ├─ Auth Service
                           ├─ Form Service  
                           ├─ Response Service
                           ├─ Analytics Service
                           ├─ File Service
                           └─ Realtime Service
```

## Quick Start

### Prerequisites

- Go 1.21 or higher
- Docker (optional)
- Access to microservices

### Local Development

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd services/api-gateway
   ```

2. **Install dependencies**
   ```bash
   go mod tidy
   ```

3. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Run the gateway**
   ```bash
   go run cmd/server/main.go
   ```

5. **Access the API**
   - Gateway: http://localhost:8080
   - Health Check: http://localhost:8080/health
   - Swagger Docs: http://localhost:8080/swagger/index.html
   - Metrics: http://localhost:9090/metrics

### Docker

1. **Build the image**
   ```bash
   docker build -t x-form-gateway .
   ```

2. **Run the container**
   ```bash
   docker run -p 8080:8080 -p 9090:9090 --env-file .env x-form-gateway
   ```

### Using Docker Compose

```bash
# From the root directory
docker-compose up api-gateway
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Gateway port | `8080` |
| `METRICS_PORT` | Metrics port | `9090` |
| `ENVIRONMENT` | Environment (development/production) | `development` |
| `JWT_SECRET` | JWT signing secret | `your-jwt-secret-key` |
| `AUTH_SERVICE_URL` | Auth service URL | `http://auth-service:3001` |
| `FORM_SERVICE_URL` | Form service URL | `http://form-service:8001` |
| `RESPONSE_SERVICE_URL` | Response service URL | `http://response-service:3002` |
| `ANALYTICS_SERVICE_URL` | Analytics service URL | `http://analytics-service:5001` |
| `FILE_SERVICE_URL` | File service URL | `http://file-service:3003` |
| `REALTIME_SERVICE_URL` | Realtime service URL | `http://realtime-service:8002` |

### Example .env file

```env
PORT=8080
METRICS_PORT=9090
ENVIRONMENT=development
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production

# Service URLs
AUTH_SERVICE_URL=http://localhost:3001
FORM_SERVICE_URL=http://localhost:8001
RESPONSE_SERVICE_URL=http://localhost:3002
ANALYTICS_SERVICE_URL=http://localhost:5001
FILE_SERVICE_URL=http://localhost:3003
REALTIME_SERVICE_URL=http://localhost:8002

# Optional: Redis for caching and rate limiting
REDIS_URL=redis://localhost:6379
```

## API Documentation

### Swagger/OpenAPI

The gateway provides comprehensive API documentation via Swagger UI:

- **URL**: http://localhost:8080/swagger/index.html
- **Redirect**: http://localhost:8080/docs
- **JSON Spec**: http://localhost:8080/swagger/doc.json

### Key Endpoints

#### Authentication
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/refresh` - Refresh JWT token
- `POST /api/v1/auth/logout` - User logout

#### Forms
- `GET /api/v1/forms` - List forms (authenticated)
- `POST /api/v1/forms` - Create form (authenticated)
- `GET /api/v1/forms/{id}` - Get form details
- `PUT /api/v1/forms/{id}` - Update form (authenticated)
- `DELETE /api/v1/forms/{id}` - Delete form (authenticated)

#### Responses
- `GET /api/v1/responses` - List responses (authenticated)
- `POST /api/v1/responses/{formId}/submit` - Submit form response (public)
- `GET /api/v1/responses/{id}` - Get response details (authenticated)

#### Analytics
- `GET /api/v1/analytics` - Get analytics dashboard (authenticated)
- `GET /api/v1/analytics/{formId}` - Get form analytics (authenticated)

#### Files
- `POST /api/v1/files/upload` - Upload file (authenticated)
- `GET /api/v1/files/{id}` - Get file details (authenticated)

#### Public Access
- `GET /forms/{formId}` - Public form access
- `POST /forms/{formId}/submit` - Public form submission

### Authentication

The gateway uses JWT Bearer tokens for authentication:

```bash
# Include in request headers
Authorization: Bearer <your-jwt-token>

# Or as query parameter
?token=<your-jwt-token>
```

## Monitoring & Observability

### Health Checks

```bash
# Gateway health
curl http://localhost:8080/health

# Response
{
  "status": "healthy",
  "service": "api-gateway", 
  "timestamp": "2024-01-01T00:00:00Z",
  "version": "1.0.0"
}
```

### Metrics

Prometheus metrics available at http://localhost:9090/metrics:

- `api_gateway_http_requests_total` - Total HTTP requests
- `api_gateway_http_duration_seconds` - Request duration
- `api_gateway_active_connections` - Active connections
- `api_gateway_errors_total` - Error count by type

### Logging

Structured JSON logging with configurable levels:

```json
{
  "level": "info",
  "msg": "Proxied request",
  "method": "GET",
  "target_url": "http://form-service:8001/api/forms",
  "duration_ms": 45,
  "request_id": "req-123",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

## Security Features

### CORS Configuration

```go
// Configurable allowed origins
allowedOrigins := []string{
  "http://localhost:3000",
  "https://app.xform.dev",
  "https://api.xform.dev"
}
```

### Security Headers

- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 1; mode=block`
- `Referrer-Policy: strict-origin-when-cross-origin`
- `Content-Security-Policy: default-src 'self'`

### Rate Limiting

- Configurable per-service rate limits
- Redis-backed distributed rate limiting
- Burst protection

## Development

### Project Structure

```
services/api-gateway/
├── cmd/
│   └── server/
│       └── main.go           # Application entry point
├── internal/
│   ├── config/               # Configuration management
│   ├── gateway/              # Gateway core logic
│   ├── handlers/             # HTTP handlers
│   ├── middleware/           # Middleware functions
│   ├── models/               # Data models
│   └── proxy/                # Proxy management
├── docs/                     # Swagger documentation
├── bin/                      # Compiled binaries
├── Dockerfile               # Container definition
├── go.mod                   # Go module definition
└── README.md               # This file
```

### Building from Source

```bash
# Build for current platform
go build -o bin/gateway cmd/server/main.go

# Build for Linux (production)
GOOS=linux GOARCH=amd64 go build -o bin/gateway-linux cmd/server/main.go

# Build with version info
go build -ldflags="-X main.Version=v1.0.0" -o bin/gateway cmd/server/main.go
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...
```

### Code Generation

```bash
# Generate Swagger documentation
swag init -g cmd/server/main.go -o docs/

# Format code
go fmt ./...

# Run linter
golangci-lint run
```

## Deployment

### Production Considerations

1. **Environment Variables**: Set proper values for production
2. **SSL/TLS**: Configure HTTPS termination (usually handled by load balancer)
3. **Resource Limits**: Set appropriate CPU and memory limits
4. **Health Checks**: Configure load balancer health checks
5. **Monitoring**: Set up Prometheus scraping and alerting
6. **Logging**: Configure log aggregation (ELK, Splunk, etc.)

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-gateway
spec:
  replicas: 3
  selector:
    matchLabels:
      app: api-gateway
  template:
    metadata:
      labels:
        app: api-gateway
    spec:
      containers:
      - name: gateway
        image: x-form-gateway:latest
        ports:
        - containerPort: 8080
        - containerPort: 9090
        env:
        - name: ENVIRONMENT
          value: "production"
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: gateway-secrets
              key: jwt-secret
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "200m"
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

### Docker Compose

```yaml
version: '3.8'
services:
  api-gateway:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
      - "9090:9090"
    environment:
      - ENVIRONMENT=production
      - JWT_SECRET=${JWT_SECRET}
      - AUTH_SERVICE_URL=http://auth-service:3001
      - FORM_SERVICE_URL=http://form-service:8001
      - RESPONSE_SERVICE_URL=http://response-service:3002
    depends_on:
      - auth-service
      - form-service
      - response-service
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
```

## Troubleshooting

### Common Issues

1. **Service Connection Errors**
   ```bash
   # Check service connectivity
   curl http://auth-service:3001/health
   
   # Check gateway logs
   docker logs api-gateway
   ```

2. **Authentication Issues**
   ```bash
   # Verify JWT secret matches across services
   echo $JWT_SECRET
   
   # Test token validation
   curl -H "Authorization: Bearer <token>" http://localhost:8080/api/v1/forms
   ```

3. **CORS Issues**
   ```bash
   # Check allowed origins in configuration
   # Verify request headers
   ```

### Debug Mode

Enable debug logging:

```bash
export ENVIRONMENT=development
export LOG_LEVEL=debug
```

### Performance Tuning

1. **Connection Pooling**: Configure HTTP client connection pools
2. **Timeouts**: Adjust service timeouts based on requirements
3. **Caching**: Implement Redis caching for frequently accessed data
4. **Load Balancing**: Use multiple gateway instances behind a load balancer

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Run linting and tests
6. Submit a pull request

### Code Style

- Follow Go conventions
- Use meaningful variable names
- Add comments for complex logic
- Write tests for new features

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For support and questions:

- **Documentation**: https://docs.xform.dev
- **Issues**: Create an issue in the repository
- **Email**: support@xform.dev
- **Community**: Join our Discord/Slack channel

## Changelog

### v1.0.0 (2024-01-01)
- Initial release
- JWT authentication
- Service routing
- Swagger documentation
- Prometheus metrics
- Docker support
- Comprehensive middleware stack
