# 🚀 X-Form Backend - Local Development Guide

A comprehensive guide for setting up and developing the X-Form Backend microservices platform locally.

## 📋 Prerequisites

### Required Tools

#### 1. **Runtime Environments**
- **Node.js**: v20.10.0+ (use .nvmrc file)
  ```bash
  # Install using nvm
  nvm install 20.10.0
  nvm use 20.10.0
  ```
- **Go**: v1.21+
  ```bash
  # macOS
  brew install go
  
  # Linux
  wget https://go.dev/dl/go1.21.6.linux-amd64.tar.gz
  sudo tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz
  ```
- **Python**: v3.11+
  ```bash
  # macOS
  brew install python@3.11
  
  # Linux
  sudo apt-get install python3.11 python3.11-pip
  ```

#### 2. **Containerization**
- **Docker**: v24.0+
  ```bash
  # macOS
  brew install --cask docker
  
  # Linux
  curl -fsSL https://get.docker.com -o get-docker.sh
  sh get-docker.sh
  ```
- **Docker Compose**: v2.20+
  ```bash
  # Usually comes with Docker Desktop
  docker compose version
  ```

#### 3. **Database Tools**
- **PostgreSQL Client**: v15+
  ```bash
  # macOS
  brew install postgresql@15
  
  # Linux
  sudo apt-get install postgresql-client-15
  ```

#### 4. **Development Tools**
- **Git**: v2.40+
- **Make**: v4.0+
- **curl**: For API testing
- **jq**: For JSON processing
  ```bash
  # macOS
  brew install git make curl jq
  
  # Linux
  sudo apt-get install git make curl jq
  ```

#### 5. **Optional but Recommended**
- **k6**: For load testing
  ```bash
  # macOS
  brew install k6
  
  # Linux
  sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
  echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
  sudo apt-get update
  sudo apt-get install k6
  ```
- **Postman**: For API testing
- **VS Code**: With recommended extensions

### System Requirements
- **RAM**: 8GB minimum, 16GB recommended
- **CPU**: 4 cores minimum
- **Storage**: 10GB free space
- **OS**: macOS 12+, Ubuntu 20.04+, Windows 10+ (with WSL2)

## 🏗️ Project Structure Overview

```
X-Form-Backend/
├── 📁 apps/                    # Microservices applications
├── 📁 packages/                # Shared libraries
├── 📁 tools/                   # Development tools
├── 📁 infrastructure/          # Infrastructure configs
├── 📁 configs/                 # Environment configurations
├── 📁 docs/                    # Documentation
├── 📁 tests/                   # Cross-service tests
└── 📁 migrations/              # Database migrations
```

## 🚀 Quick Start (5 minutes)

### 1. Clone and Setup
```bash
# Clone the repository
git clone https://github.com/your-org/X-Form-Backend.git
cd X-Form-Backend

# Setup development environment
make setup
```

### 2. Start All Services
```bash
# Start all services with Traefik
make start

# Or start development mode (faster startup)
make dev
```

### 3. Verify Installation
```bash
# Check service health
make health

# Run quick tests
make test-quick
```

### 4. Access Services
- **Main API**: http://api.localhost
- **Traefik Dashboard**: http://traefik.localhost:8080
- **Swagger UI**: http://api.localhost/docs
- **Grafana**: http://grafana.localhost:3000
- **Prometheus**: http://prometheus.localhost:9091

## 📚 Detailed Setup Guide

### Step 1: Environment Configuration

#### Clone Repository
```bash
git clone https://github.com/your-org/X-Form-Backend.git
cd X-Form-Backend
```

#### Configure Environment Variables
```bash
# Copy environment template
cp configs/environments/.env.example .env

# Edit environment variables
nano .env
```

**Key Environment Variables:**
```bash
# Database
DATABASE_URL=postgresql://user:password@localhost:5432/xform_dev
REDIS_URL=redis://localhost:6379

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-here
JWT_EXPIRE=24h

# External Services
AWS_ACCESS_KEY_ID=your-aws-key
AWS_SECRET_ACCESS_KEY=your-aws-secret
AWS_REGION=us-east-1
AWS_S3_BUCKET=xform-files-dev

# Service Ports
AUTH_SERVICE_PORT=3001
FORM_SERVICE_PORT=8001
RESPONSE_SERVICE_PORT=3002
REALTIME_SERVICE_PORT=8002
ANALYTICS_SERVICE_PORT=5001
```

### Step 2: Install Dependencies

#### Install Node.js Dependencies
```bash
# Use correct Node.js version
nvm use

# Install root dependencies
npm install

# Install service dependencies
cd apps/auth-service && npm install && cd ../..
cd apps/response-service && npm install && cd ../..
```

#### Install Go Dependencies
```bash
# Form service
cd apps/form-service
go mod download
cd ../..

# Realtime service
cd apps/realtime-service
go mod download
cd ../..
```

#### Install Python Dependencies
```bash
# Analytics service
cd apps/analytics-service
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate
pip install -r requirements.txt
cd ../..
```

### Step 3: Database Setup

#### Start Database Services
```bash
# Start PostgreSQL and Redis
docker compose -f infrastructure/docker/environments/docker-compose.dev.yml up -d postgres redis
```

#### Run Database Migrations
```bash
# PostgreSQL migrations
make migrate-postgres

# Initialize test data
make seed-data
```

#### Verify Database Connection
```bash
# Test PostgreSQL
psql $DATABASE_URL -c "SELECT version();"

# Test Redis
redis-cli ping
```

### Step 4: Service Development

#### Start Individual Services

**Auth Service (Node.js/TypeScript)**
```bash
cd apps/auth-service

# Development mode with hot reload
npm run dev

# Or production mode
npm start
```

**Form Service (Go)**
```bash
cd apps/form-service

# Development mode with hot reload
air  # requires 'go install github.com/cosmtrek/air@latest'

# Or normal run
go run cmd/server/main.go
```

**Response Service (Node.js/TypeScript)**
```bash
cd apps/response-service
npm run dev
```

**Realtime Service (Go)**
```bash
cd apps/realtime-service
go run cmd/server/main.go
```

**Analytics Service (Python)**
```bash
cd apps/analytics-service
source venv/bin/activate
uvicorn main:app --reload --port 5001
```

#### Start All Services with Traefik
```bash
# Production-like environment
make start

# Development environment (faster)
make dev

# Only infrastructure services
make infrastructure
```

## 🧪 Testing Guide

### Unit Tests
```bash
# Run all unit tests
make test

# Test specific service
make test-auth
make test-form
make test-response
make test-realtime
make test-analytics
```

### Integration Tests
```bash
# Run integration tests
make test-integration

# Test API endpoints
make test-api
```

### End-to-End Tests
```bash
# Run E2E tests
make test-e2e

# Test with real data
make test-e2e-full
```

### Load Testing
```bash
# Performance tests
make load-test

# Stress tests
make stress-test
```

## 🔧 Development Workflow

### 1. **Feature Development**
```bash
# Create feature branch
git checkout -b feature/your-feature-name

# Start development environment
make dev

# Run tests continuously
make test-watch

# Code and test your changes
# ...

# Run full test suite before commit
make test-all
```

### 2. **Service-Specific Development**

#### Adding New API Endpoint
```bash
# 1. Define route in service
# 2. Implement handler
# 3. Add validation
# 4. Write tests
# 5. Update Swagger documentation
# 6. Test integration with other services
```

#### Database Schema Changes
```bash
# 1. Create migration file
cd migrations/postgres
create-migration.sh add_new_table

# 2. Apply migration
make migrate-postgres

# 3. Update models/entities
# 4. Test migration rollback
make rollback-postgres
```

### 3. **Debugging**

#### View Service Logs
```bash
# All services
make logs

# Specific service
make logs-auth
make logs-form
docker compose logs -f auth-service
```

#### Debug Database Issues
```bash
# Connect to PostgreSQL
make db-connect

# View Redis data
make redis-cli

# Reset databases
make db-reset
```

#### Debug Traefik Routing
```bash
# Check Traefik configuration
make traefik-config

# View Traefik logs
make traefik-logs

# Access Traefik dashboard
open http://traefik.localhost:8080
```

## 📊 Monitoring and Observability

### Local Monitoring Stack
```bash
# Start monitoring services
make monitoring

# Access dashboards
open http://grafana.localhost:3000      # Grafana (admin/admin)
open http://prometheus.localhost:9091   # Prometheus
```

### Health Checks
```bash
# Check all service health
make health

# Individual service health
curl http://api.localhost/auth/health
curl http://api.localhost/forms/health
curl http://api.localhost/responses/health
curl http://api.localhost/realtime/health
curl http://api.localhost/analytics/health
```

### Metrics and Logs
```bash
# View metrics
curl http://api.localhost/metrics

# Check log aggregation
make logs-aggregated

# Export logs for analysis
make logs-export
```

## 🔐 Authentication & Authorization

### JWT Token Management
```bash
# Get authentication token
TOKEN=$(curl -X POST http://api.localhost/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"dev@example.com","password":"password"}' \
  | jq -r '.token')

# Use token in requests
curl -H "Authorization: Bearer $TOKEN" \
  http://api.localhost/forms
```

### User Management
```bash
# Create test users
make create-test-users

# Reset user data
make reset-users
```

## 🚀 Production-Like Testing

### Environment Simulation
```bash
# Start production-like environment
make prod-simulation

# Test with production configurations
make test-prod-config

# Load test with realistic data
make load-test-realistic
```

### Security Testing
```bash
# Security scans
make security-scan

# Dependency vulnerability check
make audit

# SSL/TLS testing
make test-ssl
```

## 🛠️ Common Development Tasks

### Adding New Service
```bash
# 1. Create service directory
mkdir apps/new-service

# 2. Use service template
make create-service name=new-service lang=go

# 3. Implement service following clean architecture
# 4. Add to docker-compose configuration
# 5. Update Traefik routing
# 6. Add monitoring and health checks
# 7. Write tests
```

### Updating Dependencies
```bash
# Update Node.js dependencies
npm update

# Update Go dependencies
cd apps/form-service
go get -u ./...
go mod tidy

# Update Python dependencies
cd apps/analytics-service
pip-upgrade-all  # or manually update requirements.txt
```

### Database Operations
```bash
# Backup database
make db-backup

# Restore database
make db-restore backup-file.sql

# Generate test data
make generate-test-data

# Clean test data
make clean-test-data
```

## 🔧 Troubleshooting

### Common Issues

#### Port Conflicts
```bash
# Check port usage
lsof -i :3001  # Auth service port
lsof -i :8001  # Form service port

# Kill processes using ports
make kill-ports
```

#### Docker Issues
```bash
# Clean Docker system
docker system prune -a

# Rebuild containers
make rebuild

# Reset Docker volumes
make reset-volumes
```

#### Service Communication Issues
```bash
# Test service connectivity
make test-connectivity

# Check Traefik routing
curl -H "Host: api.localhost" http://localhost/auth/health

# Verify DNS resolution
ping api.localhost
```

#### Database Connection Issues
```bash
# Test database connectivity
make test-db-connection

# Reset database containers
make reset-db

# Check database logs
docker compose logs postgres
```

### Performance Issues
```bash
# Profile application
make profile

# Check resource usage
make resource-usage

# Analyze slow queries
make analyze-queries
```

## 📖 Additional Resources

### Documentation
- [Architecture Guide](./architecture/overview.md)
- [API Documentation](./api/openapi.yml)
- [Deployment Guide](./deployment/local.md)
- [Contributing Guidelines](./development/contributing.md)

### External Tools
- [Postman Collection](./api/postman/X-Form-Backend.postman_collection.json)
- [VS Code Extensions](./.vscode/extensions.json)
- [Docker Compose Files](../infrastructure/docker/environments/)

### Support
- **Issues**: Create GitHub issues for bugs
- **Discussions**: Use GitHub discussions for questions
- **Documentation**: Update docs when adding features

## ⚡ Quick Commands Reference

```bash
# Setup and Start
make setup              # Initial setup
make start              # Start all services
make dev                # Development mode
make stop               # Stop all services

# Testing
make test               # Run all tests
make test-api          # API integration tests
make load-test         # Performance tests

# Monitoring
make health            # Service health checks
make logs              # View all logs
make monitoring        # Start monitoring stack

# Database
make migrate-postgres  # Run migrations
make db-reset         # Reset databases
make seed-data        # Add test data

# Development
make lint             # Code linting
make format           # Code formatting
make security-scan    # Security checks
make docs             # Generate documentation
```

---

## 🎯 Next Steps

1. **Complete Setup**: Follow the quick start guide
2. **Explore Services**: Check individual service documentation
3. **Run Tests**: Verify everything works correctly
4. **Start Development**: Create your first feature
5. **Read Architecture**: Understand the system design
6. **Configure IDE**: Set up your development environment

Happy coding! 🚀
