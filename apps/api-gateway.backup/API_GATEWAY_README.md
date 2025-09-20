# X-Form API Gateway

A comprehensive, enterprise-grade API Gateway for the X-Form microservices architecture with advanced security, monitoring, and integration capabilities.

## 🚀 Features

### Core Features
- **🔐 Authentication & Authorization**: JWT-based authentication with JWKS support
- **🌐 Service Discovery**: Automatic service registration and health monitoring
- **🔄 Request Routing**: Intelligent routing with circuit breaker patterns
- **📊 Observability**: Comprehensive metrics, tracing, and logging
- **🛡️ Security**: CORS, rate limiting, security headers, and mTLS support

### Integrations
- **🌊 Traefik Integration**: Advanced ingress controller with middleware chains
- **🔧 Tyk Integration**: Complete API management platform integration
- **📈 Prometheus Metrics**: Production-ready observability
- **📝 Swagger Documentation**: Auto-generated API documentation

## 📋 Prerequisites

- Go 1.21 or higher
- Docker (optional, for Traefik/Tyk)
- Redis (optional, for events)

## 🛠️ Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/Mir00r/X-Form-Backend.git
   cd X-Form-Backend/services/api-gateway
   ```

2. **Install dependencies**
   ```bash
   go mod download
   go mod tidy
   ```

3. **Configure environment**
   ```bash
   cp config.env.example .env
   # Edit .env with your configuration
   ```

4. **Build the application**
   ```bash
   go build -o bin/api-gateway cmd/server/main.go
   ```

## ⚙️ Configuration

The API Gateway uses environment variables for configuration. See `config.env.example` for all available options.

### Essential Configuration

```bash
# Server
SERVER_HOST=localhost
SERVER_PORT=8080
SERVER_ENVIRONMENT=development

# JWT Security
SECURITY_JWT_SECRET=your-super-secret-jwt-key

# Service URLs
SERVICES_AUTH_SERVICE_URL=http://localhost:3001
SERVICES_FORM_SERVICE_URL=http://localhost:3002
SERVICES_RESPONSE_SERVICE_URL=http://localhost:3003
SERVICES_ANALYTICS_SERVICE_URL=http://localhost:3004
SERVICES_COLLABORATION_SERVICE_URL=http://localhost:3005
SERVICES_REALTIME_SERVICE_URL=http://localhost:3006
```

### Advanced Configuration

#### Traefik Integration
```bash
TRAEFIK_ENABLED=true
TRAEFIK_API_URL=http://localhost:8081
TRAEFIK_API_USERNAME=admin
TRAEFIK_API_PASSWORD=admin
```

#### Tyk Integration
```bash
TYK_ENABLED=true
TYK_GATEWAY_URL=http://localhost:8080
TYK_DASHBOARD_URL=http://localhost:3000
TYK_API_KEY=your-tyk-api-key
```

#### JWKS (JSON Web Key Set)
```bash
SECURITY_JWKS_ENDPOINT=https://your-auth-server.com/.well-known/jwks.json
SECURITY_JWKS_CACHE_TIMEOUT=1h
```

## 🏃‍♂️ Running the Application

### Development Mode
```bash
# Set environment variables
export SERVER_PORT=8080
export SECURITY_JWT_SECRET=your-secret-key

# Run the application
go run cmd/server/main.go
```

### Production Mode
```bash
# Build the binary
go build -o bin/api-gateway cmd/server/main.go

# Run the binary
SERVER_ENVIRONMENT=production ./bin/api-gateway
```

### Using Docker
```bash
# Build Docker image
docker build -t x-form-api-gateway .

# Run container
docker run -p 8080:8080 \
  -e SERVER_PORT=8080 \
  -e SECURITY_JWT_SECRET=your-secret-key \
  x-form-api-gateway
```

## 📚 API Documentation

Once the gateway is running, access the comprehensive API documentation at:

- **Swagger UI**: http://localhost:8080/swagger/index.html
- **JSON Spec**: http://localhost:8080/swagger/doc.json

## 🔍 Monitoring & Health Checks

### Health Endpoints
- **Health Check**: `GET /health` - Basic health status
- **Readiness**: `GET /ready` - Service readiness with dependencies
- **Liveness**: `GET /live` - Application liveness probe

### Metrics
- **Prometheus Metrics**: `GET /metrics` - Application and system metrics
- **Service Metrics**: `GET /api/gateway/services/metrics` - Microservice metrics

### Example Health Check Response
```json
{
  "status": "healthy",
  "timestamp": "2024-01-20T10:30:00Z",
  "version": "2.0.0",
  "environment": "development",
  "services": {
    "auth-service": "healthy",
    "form-service": "healthy",
    "response-service": "healthy"
  }
}
```

## 🔐 Authentication

The API Gateway supports multiple authentication methods:

### JWT Authentication
```bash
curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
     http://localhost:8080/api/v1/forms
```

### API Key Authentication
```bash
curl -H "X-API-Key: your-api-key" \
     http://localhost:8080/api/v1/forms
```

## 🛣️ API Routes

### Gateway Management
```
GET    /health                              # Health check
GET    /ready                               # Readiness probe
GET    /live                                # Liveness probe
GET    /metrics                             # Prometheus metrics
GET    /swagger/*                           # API documentation
```

### Service Discovery
```
GET    /api/gateway/services                # List registered services
GET    /api/gateway/services/{service}/health # Service health
GET    /api/gateway/services/metrics        # Service metrics
```

### JWT Management
```
POST   /api/gateway/jwt/validate            # Validate JWT token
GET    /api/gateway/jwt/jwks               # Get JWKS public keys
```

### Microservice Proxying
All `/api/v1/*` routes are automatically proxied to the appropriate microservice:

```
/api/v1/auth/*           → auth-service
/api/v1/forms/*          → form-service
/api/v1/responses/*      → response-service
/api/v1/analytics/*      → analytics-service
/api/v1/collaboration/*  → collaboration-service
/api/v1/realtime/*       → realtime-service
```

## 🔧 Development

### Project Structure
```
.
├── cmd/server/          # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── middleware/      # HTTP middleware
│   ├── handlers/        # HTTP handlers
│   ├── jwt/            # JWT authentication service
│   ├── discovery/      # Service discovery
│   ├── traefik/        # Traefik integration
│   └── tyk/            # Tyk integration
├── docs/               # Swagger documentation
└── README.md
```

### Adding a New Service

1. **Update configuration** in `internal/config/config.go`:
   ```go
   type ServicesConfig struct {
       // ... existing services
       NewService ServiceConfig `mapstructure:"new_service"`
   }
   ```

2. **Add environment variables** in `config.env.example`:
   ```bash
   SERVICES_NEW_SERVICE_URL=http://localhost:3007
   SERVICES_NEW_SERVICE_HEALTH_PATH=/health
   ```

3. **Add routing** in `cmd/server/main.go`:
   ```go
   newService := v1.Group("/new-service")
   {
       newService.Any("/*path", serviceDiscovery.ProxyRequest("new-service"))
   }
   ```

### Running Tests
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/jwt/
```

### Generating Swagger Docs
```bash
# Install swag
go install github.com/swaggo/swag/cmd/swag@latest

# Generate documentation
swag init -g cmd/server/main.go -o docs/
```

## 🐳 Docker Integration

### Building the Image
```bash
docker build -t x-form-api-gateway .
```

### Running with Docker Compose
```yaml
version: '3.8'
services:
  api-gateway:
    build: .
    ports:
      - "8080:8080"
    environment:
      - SERVER_PORT=8080
      - SECURITY_JWT_SECRET=your-secret-key
      - SERVICES_AUTH_SERVICE_URL=http://auth-service:3001
    depends_on:
      - auth-service
      - form-service
```

## 🚀 Deployment

### Environment Variables for Production
```bash
# Required
SERVER_ENVIRONMENT=production
SECURITY_JWT_SECRET=your-production-secret-key

# Optional but recommended
SERVER_TLS_ENABLED=true
SERVER_TLS_CERT_FILE=/path/to/cert.pem
SERVER_TLS_KEY_FILE=/path/to/key.pem
OBSERVABILITY_METRICS_ENABLED=true
```

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
      - name: api-gateway
        image: x-form-api-gateway:latest
        ports:
        - containerPort: 8080
        env:
        - name: SERVER_ENVIRONMENT
          value: "production"
        - name: SECURITY_JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: api-gateway-secrets
              key: jwt-secret
```

## 🔍 Troubleshooting

### Common Issues

1. **JWT Token Invalid**
   ```
   Error: "invalid or expired token"
   Solution: Check JWT secret configuration and token expiry
   ```

2. **Service Discovery Failed**
   ```
   Error: "service not available"
   Solution: Verify service URLs and health endpoints
   ```

3. **Traefik Integration Issues**
   ```
   Error: "failed to connect to Traefik API"
   Solution: Check Traefik API URL and credentials
   ```

### Debug Mode
```bash
# Enable debug logging
export OBSERVABILITY_LOGGING_LEVEL=debug
go run cmd/server/main.go
```

### Health Check Debug
```bash
# Check gateway health
curl http://localhost:8080/health

# Check service health
curl http://localhost:8080/api/gateway/services/auth-service/health
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📞 Support

For questions and support:

- 📧 Email: api-support@x-form.com
- 🐛 Issues: [GitHub Issues](https://github.com/Mir00r/X-Form-Backend/issues)
- 📖 Documentation: [Wiki](https://github.com/Mir00r/X-Form-Backend/wiki)

---

**X-Form API Gateway** - Enterprise-grade microservices gateway with advanced security and monitoring 🚀
