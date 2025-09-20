# GitHub Copilot Instructions for X-Form Backend

## Architecture Overview

This is a **Traefik All-in-One microservices** architecture that replaces traditional API Gateway + Load Balancer setups. Traefik handles ingress, API gateway, and management layers in a single component.

### Key Architecture Pattern
```
Internet → Traefik (Ingress + Gateway + Management) → Microservices → Data Layer
```

## Service Stack & Technologies

- **Auth Service**: Node.js + TypeScript + Express (port 3001)
- **Form Service**: Go + Gin + GORM (port 8001) 
- **Response Service**: Node.js + TypeScript + Express (port 3002)
- **Real-time Service**: Go + WebSockets + Redis (port 8002)
- **Analytics Service**: Python + FastAPI (port 5001)
- **File Service**: AWS Lambda + S3 (serverless)

### Database Strategy
- **PostgreSQL**: Users, forms, structured data
- **Firestore**: Form responses, document storage  
- **Redis**: Real-time features, caching, pub/sub
- **BigQuery**: Analytics, reporting
- **S3**: File storage

## Development Workflow

### Primary Commands (use Makefile)
```bash
make setup          # Initial environment setup
make start           # Start full Traefik stack
make traefik-only    # Start only Traefik for development
make health          # Check service health
make api-test        # Test all endpoints
make load-test       # Performance testing
```

### Service Development Pattern
Each service follows **Clean Architecture** with:
- `cmd/server/main.go` (Go) or `src/app.ts` (Node.js) - entry points
- Dependency injection containers (`createAuthServiceContainer()`)
- Repository pattern for data access
- Service layer for business logic
- HTTP handlers/controllers for API

### Docker Compose Files
- `docker-compose-traefik.yml` - **Primary**: Full production-like stack
- `docker-compose.yml` - Legacy/alternative setup
- `enhanced-architecture/docker-compose-complete.yml` - Enhanced version

## Code Patterns & Conventions

### Clean Architecture Compliance
All services implement:
- **Domain Layer**: Business entities and rules
- **Application Layer**: Use cases and application services  
- **Infrastructure Layer**: External concerns (DB, HTTP, etc.)
- **Interface Layer**: Controllers, adapters, presentation

### Common Patterns
- **Dependency Injection**: Used in all services via containers
- **Repository Pattern**: Data access abstraction (`repository/` directories)
- **Middleware Chains**: Authentication, CORS, rate limiting, logging
- **Error Handling**: Structured error responses with codes
- **Health Checks**: `/health` endpoints for all services

### Node.js Services (auth-service, response-service)
- TypeScript with strict configuration
- Express.js with middleware pattern
- Joi/express-validator for validation
- Winston for logging
- Jest for testing
- Swagger documentation at `/docs`

### Go Services (form-service, realtime-service)  
- Gin web framework
- GORM for database operations
- Structured logging with logrus
- Swagger generation with swaggo
- Clean error handling patterns

## Integration Points

### Authentication Flow
1. Client → Traefik → Auth Service (JWT generation)
2. Subsequent requests include JWT in Authorization header
3. Traefik validates JWT via middleware
4. Services receive validated user context

### Inter-Service Communication
- **Synchronous**: Direct HTTP calls between services
- **Asynchronous**: Redis pub/sub for real-time features
- **Event Bus**: Planned for complex workflows

### Traefik Configuration
- **Static config**: `infrastructure/traefik/traefik.yml`
- **Dynamic config**: `infrastructure/traefik/dynamic/` directory
- **Service discovery**: Docker labels auto-configure routing

## Critical Files to Understand

### Architecture Documentation
- `ARCHITECTURE_V2.md` - Current architecture specification
- `enhanced-architecture/README.md` - Production implementation details
- `IMPLEMENTATION_GUIDE.md` - Detailed implementation notes

### Configuration
- `.env.example` - Environment variables template
- `infrastructure/traefik/` - Traefik configuration
- Each service has its own `Dockerfile` and config

### Scripts & Automation
- `scripts/setup.sh` - Environment initialization
- `Makefile` - Primary development interface
- `.github/workflows/` - CI/CD pipelines

## Development Guidelines

### When Adding New Services
1. Follow the established service structure pattern
2. Implement health checks at `/health`
3. Add Docker labels for Traefik routing
4. Update `docker-compose-traefik.yml`
5. Add service to Makefile commands

### When Modifying Traefik Config
- Test with `make traefik-config` (validates configuration)
- Use `make traefik-logs` for debugging
- Access Traefik dashboard at `http://traefik.localhost:8080`

### Database Changes
- Go services: Use GORM AutoMigrate in `database.Migrate()`
- Node.js services: Manual migration scripts in `scripts/`
- Always test migrations with `make setup`

## Debugging & Monitoring

### Local Development Access Points
- **Main API**: `http://api.localhost`
- **WebSocket**: `ws://ws.localhost`
- **Traefik Dashboard**: `http://traefik.localhost:8080`
- **Grafana**: `http://grafana.localhost:3000`
- **Prometheus**: `http://prometheus.localhost:9091`

### Log Investigation
```bash
make logs                    # All service logs
make traefik-logs           # Traefik-specific logs
docker-compose logs [service] # Individual service logs
```

### Common Issues
- **Service not routing**: Check Traefik labels in docker-compose
- **JWT errors**: Verify JWT_SECRET environment variable consistency
- **Database connection**: Ensure services wait for DB health checks
- **CORS issues**: Check Traefik CORS middleware configuration

## Performance Considerations

This architecture achieves:
- **60% lower latency** vs traditional multi-proxy setups
- **100% higher throughput** with Traefik's single-component design
- **Built-in load balancing** and circuit breakers
- **Observability** with Prometheus metrics and Jaeger tracing