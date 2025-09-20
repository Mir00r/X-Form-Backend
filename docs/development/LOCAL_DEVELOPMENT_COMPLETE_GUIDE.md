# X-Form Backend - Complete Local Development Guide

> **🎯 Complete Guide for Local Development, Testing, and Contributing**  
> Everything you need to get started with X-Form Backend development in under 15 minutes!

## 📋 Table of Contents

1. [🛠️ Prerequisites and Tools](#️-prerequisites-and-tools)
2. [⚡ Quick Start (5 Minutes)](#-quick-start-5-minutes)
3. [🔧 Development Setup](#-development-setup)
4. [🚀 Running Services](#-running-services)
5. [🧪 Testing and API Usage](#-testing-and-api-usage)
6. [🔐 Authentication Flow](#-authentication-flow)
7. [📊 Monitoring and Observability](#-monitoring-and-observability)
8. [🐛 Development Workflows](#-development-workflows)
9. [🚨 Troubleshooting](#-troubleshooting)
10. [📚 Additional Resources](#-additional-resources)

---

## 🛠️ Prerequisites and Tools

### Required Tools Installation

#### 1. **Core Development Tools**
```bash
# Install Node.js (v18+ recommended)
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.3/install.sh | bash
nvm install 18
nvm use 18

# Install Go (v1.21+ recommended)
# macOS with Homebrew
brew install go

# Verify Go installation
go version  # Should show 1.21+

# Install Python (v3.8+ recommended)
# macOS with Homebrew
brew install python3

# Verify Python installation
python3 --version  # Should show 3.8+
```

#### 2. **Container and Infrastructure Tools**
```bash
# Install Docker Desktop
# Download from: https://www.docker.com/products/docker-desktop

# Verify Docker installation
docker --version
docker-compose --version

# Install Make (usually pre-installed on macOS/Linux)
make --version
```

#### 3. **Optional Development Tools**
```bash
# Install useful development tools
npm install -g nodemon          # Auto-restart Node.js apps
go install github.com/cosmtrek/air@latest  # Auto-restart Go apps
pip3 install virtualenv         # Python virtual environments

# Install API testing tools
brew install curl               # Already installed on most systems
npm install -g @apidevtools/swagger-cli  # Swagger validation
brew install httpie            # Modern HTTP client
```

#### 4. **Code Editor Setup**
**Recommended: VS Code with Extensions**
```bash
# Install VS Code extensions
code --install-extension ms-vscode.vscode-typescript-next
code --install-extension golang.go
code --install-extension ms-python.python
code --install-extension ms-vscode.docker
code --install-extension humao.rest-client
```

---

## ⚡ Quick Start (5 Minutes)

### 1. **Clone and Setup**
```bash
# Clone the repository
git clone https://github.com/your-org/X-Form-Backend.git
cd X-Form-Backend

# Run automated setup
make setup
```

### 2. **Start Development Environment**
```bash
# Start all services in development mode
make dev

# OR start specific mode
make enhanced-dev  # For enhanced architecture
```

### 3. **Verify Installation**
```bash
# Check service health
make health

# Access key endpoints
open http://localhost:8080              # API Gateway
open http://localhost:8080/swagger/     # API Documentation
open http://traefik.localhost:8080      # Traefik Dashboard
```

---

## 🔧 Development Setup

### 1. **Environment Configuration**

#### Create and Configure `.env` File
```bash
# Copy environment template
cp configs/environments/.env.example .env

# Edit with your preferred editor
nano .env  # or code .env
```

#### **Essential Environment Variables**
```bash
# Database Configuration
DATABASE_URL=postgresql://xform_user:xform_password@localhost:5432/xform_db
REDIS_URL=redis://localhost:6379

# JWT Configuration (IMPORTANT: Change in production!)
JWT_SECRET=your-super-secret-jwt-key-for-development-change-in-production
JWT_EXPIRE=24h

# Service Ports
AUTH_SERVICE_PORT=3001
FORM_SERVICE_PORT=8001
RESPONSE_SERVICE_PORT=3002
REALTIME_SERVICE_PORT=8002
ANALYTICS_SERVICE_PORT=5001
API_GATEWAY_PORT=8080

# Development Settings
NODE_ENV=development
LOG_LEVEL=debug

# External Services (Optional for local development)
AWS_ACCESS_KEY_ID=your-aws-key  
AWS_SECRET_ACCESS_KEY=your-aws-secret
AWS_REGION=us-east-1
GOOGLE_CLIENT_ID=your-google-client-id
FIREBASE_PROJECT_ID=your-firebase-project-id
```

### 2. **Dependency Installation**

#### Install All Dependencies at Once
```bash
# Install all service dependencies
make install-deps

# This runs:
# - npm install (Node.js services)
# - go mod download (Go services) 
# - pip install -r requirements.txt (Python services)
```

#### Manual Installation (Alternative)
```bash
# Root dependencies
npm install

# Node.js services
cd apps/auth-service && npm install && cd ../..
cd apps/response-service && npm install && cd ../..

# Go services
cd apps/form-service && go mod download && cd ../..
cd apps/api-gateway && go mod download && cd ../..
cd apps/realtime-service && go mod download && cd ../..

# Python services
cd apps/analytics-service
python3 -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate
pip install -r requirements.txt
cd ../..
```

### 3. **Database Setup**

#### Start Database Services
```bash
# Start PostgreSQL and Redis using Docker
docker-compose -f infrastructure/containers/docker-compose.yml up -d postgres redis

# Verify database connectivity
docker-compose logs postgres  # Check PostgreSQL logs
docker-compose logs redis     # Check Redis logs
```

#### Initialize Database
```bash
# Run database migrations and seed data
make db-setup

# Manual setup (if needed)
psql $DATABASE_URL -c "CREATE DATABASE xform_db;"
psql $DATABASE_URL -f migrations/postgres/001_initial_schema.sql
```

---

## 🚀 Running Services

### Development Modes

#### 1. **Full Stack Development (Recommended)**
```bash
# Start all services with hot reload
make dev

# This starts:
# - PostgreSQL & Redis (Docker)
# - All microservices with hot reload
# - Traefik reverse proxy
# - Monitoring stack (optional)
```

#### 2. **Enhanced Architecture Mode**
```bash
# Start with enhanced API Gateway
make enhanced-dev

# Features:
# - Production-ready API Gateway
# - Advanced middleware stack
# - Comprehensive observability
# - TLS termination
```

#### 3. **Individual Service Development**
```bash
# Start infrastructure only
make infra-start

# Then start services individually:

# Auth Service (Node.js + TypeScript)
cd apps/auth-service
npm run dev  # Hot reload with nodemon

# Form Service (Go)
cd apps/form-service
air  # Hot reload with air
# OR: go run cmd/server/main.go

# Response Service (Node.js)
cd apps/response-service
npm start

# Analytics Service (Python + FastAPI)
cd apps/analytics-service
source venv/bin/activate
uvicorn main:app --reload --port 5001

# Realtime Service (Go + WebSockets)
cd apps/realtime-service
air  # Hot reload
# OR: go run cmd/server/main.go
```

#### 4. **Production Mode (Local)**
```bash
# Start all services in production mode
make start

# Use Docker containers for all services
docker-compose -f infrastructure/containers/docker-compose-traefik.yml up -d
```

### Service Access Points

| Service | URL | Description |
|---------|-----|-------------|
| **API Gateway** | `http://localhost:8080` | Main API entry point |
| **Swagger Docs** | `http://localhost:8080/swagger/` | Interactive API documentation |
| **Auth Service** | `http://localhost:3001` | Authentication & user management |
| **Form Service** | `http://localhost:8001` | Form creation & management |
| **Response Service** | `http://localhost:3002` | Form response handling |
| **Realtime Service** | `http://localhost:8002` | WebSocket & real-time features |
| **Analytics Service** | `http://localhost:5001` | Analytics & reporting |
| **Traefik Dashboard** | `http://traefik.localhost:8080` | Reverse proxy dashboard |
| **Grafana** | `http://grafana.localhost:3000` | Metrics dashboard (admin/admin) |
| **Prometheus** | `http://prometheus.localhost:9091` | Metrics collection |

---

## 🧪 Testing and API Usage

### Health Checks

#### Check All Services
```bash
# Automated health check
make health

# Manual checks
curl http://localhost:8080/health              # API Gateway
curl http://localhost:3001/health              # Auth Service
curl http://localhost:8001/health              # Form Service
curl http://localhost:3002/health              # Response Service
curl http://localhost:8002/health              # Realtime Service
curl http://localhost:5001/health              # Analytics Service
```

### API Testing with cURL

#### 1. **Authentication Endpoints**

**Register a New User**
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "developer@example.com",
    "username": "developer",
    "password": "SecurePass123!",
    "firstName": "John",
    "lastName": "Developer"
  }'
```

**Login User**
```bash
# Login and save token
RESPONSE=$(curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "developer@example.com",
    "password": "SecurePass123!"
  }')

# Extract token (requires jq)
TOKEN=$(echo $RESPONSE | jq -r '.token')
echo "Token: $TOKEN"

# Save token for subsequent requests
export AUTH_TOKEN=$TOKEN
```

**Get User Profile**
```bash
curl -X GET http://localhost:8080/api/v1/auth/profile \
  -H "Authorization: Bearer $AUTH_TOKEN"
```

#### 2. **Form Management Endpoints**

**Create a New Form**
```bash
curl -X POST http://localhost:8080/api/v1/forms \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Customer Feedback Form",
    "description": "Please provide your feedback about our service",
    "questions": [
      {
        "type": "text",
        "title": "What is your name?",
        "required": true
      },
      {
        "type": "rating",
        "title": "How would you rate our service?",
        "required": true,
        "options": [1, 2, 3, 4, 5]
      }
    ]
  }'
```

**List User Forms**
```bash
curl -X GET http://localhost:8080/api/v1/forms \
  -H "Authorization: Bearer $AUTH_TOKEN"
```

**Get Specific Form**
```bash
# Replace {form_id} with actual form ID
curl -X GET http://localhost:8080/api/v1/forms/{form_id} \
  -H "Authorization: Bearer $AUTH_TOKEN"
```

#### 3. **Response Submission**

**Submit Form Response (Public)**
```bash
curl -X POST http://localhost:8080/api/v1/responses/{form_id}/submit \
  -H "Content-Type: application/json" \
  -d '{
    "responses": {
      "name": "Jane Smith",
      "rating": 5
    }
  }'
```

#### 4. **Analytics Endpoints**

**Get Form Analytics**
```bash
curl -X GET http://localhost:8080/api/v1/analytics/forms/{form_id} \
  -H "Authorization: Bearer $AUTH_TOKEN"
```

### Interactive API Testing

#### 1. **Swagger UI (Recommended)**
```bash
# Open Swagger documentation
open http://localhost:8080/swagger/

# Features:
# - Try out APIs directly
# - Authentication with JWT tokens
# - Request/response examples
# - Schema validation
```

#### 2. **HTTPie (Modern CLI tool)**
```bash
# Install HTTPie
brew install httpie

# Register user
http POST localhost:8080/api/v1/auth/register \
  email=dev@example.com \
  username=devuser \
  password=SecurePass123!

# Login and get token
http POST localhost:8080/api/v1/auth/login \
  email=dev@example.com \
  password=SecurePass123!

# Use token in subsequent requests
http GET localhost:8080/api/v1/auth/profile \
  Authorization:"Bearer YOUR_TOKEN_HERE"
```

#### 3. **Postman Collection**
```bash
# Generate Postman collection from Swagger
swagger-codegen generate -i http://localhost:8080/swagger/swagger.json \
  -l postman2 -o postman-collection/
```

---

## 🔐 Authentication Flow

### Development Authentication

#### 1. **Standard JWT Flow**
```bash
# 1. Register user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123","username":"testuser"}'

# 2. Login to get tokens
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'

# Response:
# {
#   "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
#   "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
#   "expires_in": 86400
# }

# 3. Use access token in requests
curl -X GET http://localhost:8080/api/v1/auth/profile \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"

# 4. Refresh token when expired
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"YOUR_REFRESH_TOKEN"}'
```

#### 2. **Development Mock Authentication**
```bash
# For testing purposes, you can use mock tokens
# Analytics service accepts "dev-token" for development
curl -X GET http://localhost:5001/analytics/form123/summary \
  -H "Authorization: Bearer dev-token"
```

#### 3. **Service-to-Service Authentication**
```bash
# API Key for service communication
curl -X GET http://localhost:3002/api/v1/responses \
  -H "X-API-Key: your-api-key-for-service-to-service-communication"
```

---

## 📊 Monitoring and Observability

### Local Monitoring Stack

#### 1. **Start Monitoring Services**
```bash
# Start full observability stack
make monitoring

# Manual start
docker-compose -f infrastructure/monitoring/docker-compose.yml up -d
```

#### 2. **Access Monitoring Dashboards**

| Tool | URL | Credentials | Purpose |
|------|-----|-------------|---------|
| **Grafana** | `http://grafana.localhost:3000` | admin/admin | Metrics visualization |
| **Prometheus** | `http://prometheus.localhost:9091` | - | Metrics collection |
| **Jaeger** | `http://localhost:16686` | - | Distributed tracing |
| **AlertManager** | `http://localhost:9093` | - | Alert management |

#### 3. **Metrics and Health Monitoring**
```bash
# Check all service metrics
curl http://localhost:8080/metrics         # API Gateway
curl http://localhost:3001/metrics         # Auth Service
curl http://localhost:8001/metrics         # Form Service

# View Prometheus targets
curl http://prometheus.localhost:9091/api/v1/targets

# Check Grafana datasources
curl -u admin:admin http://grafana.localhost:3000/api/datasources
```

#### 4. **Log Aggregation**
```bash
# View logs for all services
make logs

# View specific service logs
docker-compose logs -f auth-service
docker-compose logs -f form-service

# Follow logs in real-time
make logs-follow
```

### Performance Testing

#### 1. **Load Testing with Apache Bench**
```bash
# Install Apache Bench
brew install apache2

# Test health endpoint
ab -n 1000 -c 50 http://localhost:8080/health

# Test auth endpoint
ab -n 100 -c 10 -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/auth/profile
```

#### 2. **Load Testing with Artillery**
```bash
# Install Artillery
npm install -g artillery

# Create load test configuration
cat > load-test.yml << EOF
config:
  target: http://localhost:8080
  phases:
    - duration: 60
      arrivalRate: 10
scenarios:
  - name: "Health check load test"
    requests:
      - get:
          url: "/health"
EOF

# Run load test
artillery run load-test.yml
```

---

## 🐛 Development Workflows

### Code Development Workflow

#### 1. **Setting Up Development Environment**
```bash
# 1. Start infrastructure
make infra-start

# 2. Start service in development mode
cd apps/auth-service
npm run dev:watch  # Auto-restart on changes

# 3. In another terminal, run tests
npm run test:watch  # Auto-run tests on changes
```

#### 2. **Making Changes to Services**

**For Node.js Services (auth-service, response-service):**
```bash
cd apps/auth-service

# Install dependencies
npm install

# Start development server with hot reload
npm run dev

# Run tests
npm test
npm run test:coverage

# Lint and format code
npm run lint
npm run lint:fix
```

**For Go Services (form-service, api-gateway, realtime-service):**
```bash
cd apps/form-service

# Install dependencies
go mod download
go mod tidy

# Start with hot reload
air

# Build binary
go build -o bin/form-service cmd/server/main.go

# Run tests
go test ./...
go test -v ./internal/...

# Run with race detection
go test -race ./...
```

**For Python Services (analytics-service):**
```bash
cd apps/analytics-service

# Activate virtual environment
source venv/bin/activate

# Install dependencies
pip install -r requirements.txt

# Start development server
uvicorn main:app --reload --port 5001

# Run tests
pytest

# Check code quality
black .
flake8 .
```

#### 3. **Database Migrations**
```bash
# Run database migrations
make db-migrate

# Reset database (WARNING: Destructive)
make db-reset

# Seed test data
make db-seed
```

#### 4. **Testing Individual Services**
```bash
# Test specific service
make test-auth-service
make test-form-service
make test-response-service

# Integration tests
make test-integration

# End-to-end tests
make test-e2e
```

### Git Workflow

#### 1. **Feature Development**
```bash
# Create feature branch
git checkout -b feature/user-authentication-improvements

# Make changes
# ... code changes ...

# Run quality checks
make lint
make test
make security-scan

# Commit changes
git add .
git commit -m "feat: improve user authentication flow"

# Push and create PR
git push origin feature/user-authentication-improvements
```

#### 2. **Code Quality Checks**
```bash
# Run all quality checks
make quality-check

# Individual checks
make lint          # Code linting
make test          # All tests
make security-scan # Security scanning
make docs          # Generate documentation
```

---

## 🚨 Troubleshooting

### Common Issues and Solutions

#### 1. **Port Already in Use**
```bash
# Find and kill process using port
lsof -ti:8080 | xargs kill -9

# Or use different ports
PORT=8081 make dev
```

#### 2. **Docker Issues**
```bash
# Reset Docker environment
make clean
docker system prune -a

# Restart Docker services
make stop
make start
```

#### 3. **Database Connection Issues**
```bash
# Check if PostgreSQL is running
docker-compose ps postgres

# Check PostgreSQL logs
docker-compose logs postgres

# Test database connection
psql $DATABASE_URL -c "SELECT version();"

# Reset database
make db-reset
```

#### 4. **Service Not Starting**
```bash
# Check service logs
make logs-auth-service
make logs-form-service

# Check service health
curl http://localhost:3001/health

# Debug service startup
cd apps/auth-service
DEBUG=* npm run dev
```

#### 5. **JWT Token Issues**
```bash
# Verify JWT secret is consistent across services
grep JWT_SECRET .env

# Test token generation
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password"}'

# Decode JWT token (for debugging)
echo "YOUR_JWT_TOKEN" | cut -d. -f2 | base64 -d | jq
```

#### 6. **CORS Issues**
```bash
# Check CORS configuration
curl -I -H "Origin: http://localhost:3000" \
  http://localhost:8080/api/v1/auth/health

# Update CORS settings in .env
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001
```

#### 7. **WebSocket Connection Issues**
```bash
# Test WebSocket connection
wscat -c ws://localhost:8002/forms/123/updates

# Check WebSocket service logs
docker-compose logs realtime-service
```

### Debug Mode

#### 1. **Enable Debug Logging**
```bash
# Set debug environment
export LOG_LEVEL=debug
export NODE_ENV=development

# For Go services
export GIN_MODE=debug

# For Python services
export UVICORN_LOG_LEVEL=debug
```

#### 2. **Debug Individual Services**
```bash
# Node.js debugging
cd apps/auth-service
node --inspect=0.0.0.0:9229 dist/app.js

# Go debugging with Delve
cd apps/form-service
dlv debug cmd/server/main.go --headless --listen=:2345
```

---

## 📚 Additional Resources

### Documentation Links

- [**Architecture Guide**](../architecture/ARCHITECTURE_V2.md) - System architecture overview
- [**API Documentation**](http://localhost:8080/swagger/) - Interactive API docs
- [**Deployment Guide**](../deployment/DEPLOYMENT_GUIDE.md) - Production deployment
- [**Observability Guide**](../operations/OBSERVABILITY_COMPLETE.md) - Monitoring setup

### Service-Specific Guides

- [**Auth Service Guide**](../../apps/auth-service/README_CLEAN_ARCHITECTURE.md)
- [**Form Service Guide**](../../apps/form-service/README.md)
- [**Response Service Guide**](../../apps/response-service/README.md)
- [**Analytics Service Guide**](../../apps/analytics-service/QUICK_START.md)
- [**Realtime Service Guide**](../../apps/realtime-service/README_COMPLETE.md)

### Development Tools

#### 1. **Code Generation**
```bash
# Generate API client
swagger-codegen generate -i http://localhost:8080/swagger/swagger.json \
  -l javascript -o client-sdk/

# Generate mock server
swagger-codegen generate -i http://localhost:8080/swagger/swagger.json \
  -l nodejs-server -o mock-server/
```

#### 2. **Database Tools**
```bash
# Database GUI tools
# TablePlus: https://tableplus.com/
# pgAdmin: https://www.pgadmin.org/

# Command line tools
psql $DATABASE_URL
redis-cli -u $REDIS_URL
```

#### 3. **API Testing Tools**
```bash
# Postman: https://www.postman.com/
# Insomnia: https://insomnia.rest/
# Thunder Client (VS Code): code --install-extension rangav.vscode-thunder-client
```

### Community and Support

- **GitHub Issues**: [Report bugs or request features](https://github.com/your-org/X-Form-Backend/issues)
- **Discussions**: [Community discussions](https://github.com/your-org/X-Form-Backend/discussions)
- **Wiki**: [Additional documentation](https://github.com/your-org/X-Form-Backend/wiki)

---

## 🎯 Quick Reference Commands

### Essential Commands
```bash
make setup              # Initial setup
make dev                # Start development environment
make health             # Check all services
make stop               # Stop all services
make clean              # Clean up containers and volumes
make logs               # View all service logs
make test               # Run all tests
make help               # Show all available commands
```

### Service Management
```bash
make start-auth-service     # Start auth service
make start-form-service     # Start form service
make start-response-service # Start response service
make restart-auth-service   # Restart auth service
make logs-auth-service      # View auth service logs
```

### Testing Commands
```bash
make test-unit              # Unit tests
make test-integration       # Integration tests
make test-e2e              # End-to-end tests
make test-coverage         # Coverage report
make load-test             # Performance testing
```

---

**🎉 You're ready to start developing with X-Form Backend!**

This guide covers everything you need to get started. For specific questions or issues, refer to the troubleshooting section or check the service-specific documentation.

Happy coding! 🚀
