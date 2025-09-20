# X-Form Backend - Google Forms-like Survey Builder

A microservices-based backend system for building and managing surveys with real-time collaboration features, powered by **Traefik All-in-One Architecture**.

## 🏗️ **Architecture Overview**

```
                     ┌─────────────────────────────┐
                     │     Frontend (React)        │
                     │   • REST API calls          │
                     │   • WebSocket connections   │
                     │   • Real-time updates       │
                     └──────────────┬──────────────┘
                                    │
                                    ▼
                     ┌─────────────────────────────┐
                     │    TRAEFIK ALL-IN-ONE       │
                     │                             │
                     │  🔒 Ingress Controller      │
                     │    • TLS termination        │
                     │    • Service discovery      │
                     │    • Load balancing         │
                     │                             │
                     │  🚀 API Gateway             │
                     │    • JWT authentication     │
                     │    • CORS handling          │
                     │    • Request routing        │
                     │    • API versioning         │
                     │                             │
                     │  📊 API Management          │
                     │    • Rate limiting          │
                     │    • Analytics & monitoring │
                     │    • Circuit breaker        │
                     │    • Request/response logs  │
                     └──────────────┬──────────────┘
                                    │
                    ┌───────────────┼───────────────┐
                    │               │               │
                    ▼               ▼               ▼
        ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
        │  Auth Service   │ │  Form Service   │ │Response Service │
        │   (Node.js)     │ │     (Go)        │ │   (Node.js)     │
        │ • JWT tokens    │ │ • Form CRUD     │ │ • Submissions   │
        │ • User mgmt     │ │ • PostgreSQL    │ │ • Firestore     │
        │ • OAuth        │ │ • Validation    │ │ • Public forms  │
        └─────────────────┘ └─────────────────┘ └─────────────────┘
                    │               │               │
                    ▼               ▼               ▼
        ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
        │Real-time Service│ │Analytics Service│ │  File Service   │
        │     (Go)        │ │   (Python)      │ │    (NGINX)      │
        │ • WebSockets    │ │ • BigQuery      │ │ • File uploads  │
        │ • Redis Pub/Sub │ │ • Reporting     │ │ • S3 storage    │
        │ • Collaboration │ │ • AI insights   │ │ • CDN delivery  │
        └─────────────────┘ └─────────────────┘ └─────────────────┘
                    │               │               │
                    └───────────────┼───────────────┘
                                    │
                            ┌───────▼───────┐
                            │   DATA LAYER  │
                            │               │
                            │ • PostgreSQL  │
                            │ • Redis       │
                            │ • Firestore   │
                            │ • BigQuery    │
                            │ • S3          │
                            └───────────────┘
```

## 🚀 **Key Features**

### **Traefik All-in-One Benefits**
- 🔥 **Single Component**: Replaces complex multi-service proxy setup
- ⚡ **High Performance**: 60% lower latency, 100% higher throughput
- 🔒 **Enterprise Security**: Multi-layer security with JWT, CORS, rate limiting
- 📊 **Full Observability**: Metrics, tracing, and analytics built-in
- 🛠️ **Easy Operations**: Hot reloading, health checks, auto-scaling

### **Microservices Features**
- 🔐 **Authentication**: JWT-based auth with Google OAuth support
- 📋 **Form Management**: Dynamic form builder with validation
- 📝 **Response Collection**: Real-time form submissions and storage
- 🔄 **Real-time Collaboration**: WebSocket-based live updates
- 📈 **Analytics**: Advanced reporting and AI-powered insights
- 📎 **File Handling**: Secure file uploads and CDN delivery
## ⚡ **Quick Start**

### **🔧 For Developers - Complete Setup Guide**
> **📖 [Complete Local Development Guide](docs/development/LOCAL_DEVELOPMENT_COMPLETE_GUIDE.md)** - Everything you need to know for local development, testing, and contributing

> **⚡ [Developer Quick Reference](docs/development/DEVELOPER_QUICK_REFERENCE.md)** - Essential commands and endpoints for daily development

### **Prerequisites**
- Docker and Docker Compose
- Node.js 18+, Go 1.21+, Python 3.8+
- `hey` load testing tool: `go install github.com/rakyll/hey@latest`

### **1. Clone and Setup**
```bash
git clone <repository-url>
cd X-Form-Backend
make setup
```

### **2. Start the Stack**
```bash
# Start Traefik + all microservices
make start

# OR for development with hot reload
make dev

# Check system health
make health

# Test API endpoints
make api-test
```

### **3. Access Points**
- 🌐 **Main API**: http://api.localhost (or http://localhost:8080)
- 📚 **API Documentation**: http://localhost:8080/swagger/
- 🔌 **WebSocket**: ws://ws.localhost
- 📊 **Traefik Dashboard**: http://traefik.localhost:8080
- 📈 **Grafana**: http://grafana.localhost:3000
- 🔍 **Prometheus**: http://prometheus.localhost:9091

## 🔧 **Development Commands**

```bash
# Architecture management
make traefik-only      # Start only Traefik
make traefik-config    # Validate configuration
make traefik-logs      # Show Traefik logs
make arch-info         # Show architecture info

# Development workflow
make dev               # Start all services with hot reload
make install-deps      # Install all dependencies
make db-setup          # Setup databases
make health            # Check all service health

# Testing & monitoring
make test              # Run all tests
make load-test         # Performance testing
make monitor           # Open dashboards
make api-test          # Test endpoints

# Individual services
make auth-dev          # Auth service development
make form-dev          # Form service development
make analytics-dev     # Analytics service development
```

## 🧪 **Quick API Testing**

### **Authentication Flow**
```bash
# Register user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"dev@example.com","username":"dev","password":"SecurePass123!","firstName":"Dev","lastName":"User"}'

# Login and get token
TOKEN=$(curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"dev@example.com","password":"SecurePass123!"}' | jq -r '.token')

# Get user profile
curl -X GET http://localhost:8080/api/v1/auth/profile \
  -H "Authorization: Bearer $TOKEN"
```

### **Form Management**
```bash
# Create form
curl -X POST http://localhost:8080/api/v1/forms \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Test Form","description":"A test form","questions":[{"type":"text","title":"Your name?","required":true}]}'

# List forms
curl -X GET http://localhost:8080/api/v1/forms \
  -H "Authorization: Bearer $TOKEN"
```

> 📚 **For complete API documentation**: Visit http://localhost:8080/swagger/ after starting the services

          │
          ▼
┌──────────────────────────────────────────────┐
│              File Service (Lambda)            │
│ - /upload → S3                                │
│ - /files/:id → metadata                       │
│ - Handles thumbnails, PDF parsing             │
└──────────────────┬───────────────────────────┘
                   │
              ┌────▼────┐
              │   S3    │
              │ storage │
              └─────────┘
```

## Tech Stack

- **Languages:** Go (high-performance APIs), Node.js (auth + integrations), Python (analytics)
- **Databases:** PostgreSQL (users/forms), Firestore (responses), Redis (real-time), BigQuery (analytics)
- **Cloud & Infra:** AWS (API Gateway, Lambda, S3, RDS, EventBridge)
- **Containerization:** Docker + Kubernetes

## 📁 Project Structure

```
X-Form-Backend/
├── 📁 apps/                    # Microservices and applications
│   ├── auth-service/           # User authentication service (Node.js)
│   ├── form-service/           # Form management service (Go)
│   ├── response-service/       # Response collection service (Node.js)
│   ├── realtime-service/       # Real-time collaboration service (Go)
│   ├── analytics-service/      # Analytics and reporting service (Python)
│   ├── api-gateway/           # API gateway service
│   ├── file-upload-service/   # File handling service
│   ├── collaboration-service/ # Collaboration features
│   └── event-bus-service/     # Event messaging service
├── 📁 infrastructure/         # Infrastructure and deployment configs
│   ├── containers/            # Docker Compose files
│   ├── kubernetes/            # Kubernetes manifests
│   ├── terraform/             # Infrastructure as Code
│   ├── monitoring/            # Observability configs
│   └── reverse-proxy/         # Load balancer configs
├── 📁 tools/                  # Development and build tools
│   ├── automation/            # Build automation (Makefile)
│   └── scripts/               # Setup and deployment scripts
├── 📁 docs/                   # Documentation
│   ├── architecture/          # Architecture documentation
│   ├── development/           # Development guides
│   ├── deployment/            # Deployment guides
│   └── operations/            # Operations guides
├── 📁 configs/                # Configuration files
│   └── environments/          # Environment-specific configs
├── 📁 packages/               # Shared libraries and packages
└── 📁 tests/                  # Integration and E2E tests
```

This structure follows industry standards with clear separation of concerns:
- **Apps**: Business logic and microservices
- **Infrastructure**: Deployment and infrastructure code
- **Tools**: Development automation and scripts
- **Docs**: Comprehensive documentation
- **Configs**: Environment configurations

## 📚 **Documentation and Guides**

### **🚀 Getting Started**
- **[Tools Installation Guide](docs/development/TOOLS_INSTALLATION_GUIDE.md)** - Install all required development tools
- **[Local Development Complete Guide](docs/development/LOCAL_DEVELOPMENT_COMPLETE_GUIDE.md)** - Complete setup, development, and testing guide
- **[Developer Quick Reference](docs/development/DEVELOPER_QUICK_REFERENCE.md)** - Essential commands and endpoints
- **[Contributing Guide](CONTRIBUTING.md)** - How to contribute to the project

### **🏗️ Architecture and Design**
- **[Architecture V2](docs/architecture/ARCHITECTURE_V2.md)** - Current system architecture
- **[Implementation Guide](docs/development/IMPLEMENTATION_GUIDE.md)** - Detailed implementation notes
- **[Enhanced Architecture](docs/architecture/enhanced/ARCHITECTURE_IMPLEMENTATION_COMPLETE.md)** - Production-ready architecture

### **🔧 Development Workflows**
- **[Local Development Tutorial](docs/development/TUTORIAL.md)** - Step-by-step development tutorial
- **[API Testing Guide](docs/development/LOCAL_DEVELOPMENT_COMPLETE_GUIDE.md#-testing-and-api-usage)** - Complete API testing examples
- **[Service Development](docs/development/LOCAL_DEVELOPMENT_COMPLETE_GUIDE.md#-development-workflows)** - Individual service development

### **🚀 Deployment and Operations**
- **[Deployment Guide](docs/deployment/DEPLOYMENT_GUIDE.md)** - Production deployment
- **[CI/CD Implementation](docs/deployment/CI_CD_INFRASTRUCTURE_IMPLEMENTATION_GUIDE.md)** - Continuous integration setup
- **[Observability Guide](docs/operations/OBSERVABILITY_COMPLETE.md)** - Monitoring and observability

### **📖 Service-Specific Documentation**
- **[Auth Service](apps/auth-service/README_CLEAN_ARCHITECTURE.md)** - Authentication and user management
- **[Form Service](apps/form-service/README.md)** - Form creation and management
- **[Response Service](apps/response-service/README.md)** - Form response handling
- **[Analytics Service](apps/analytics-service/QUICK_START.md)** - Analytics and reporting
- **[Realtime Service](apps/realtime-service/README_COMPLETE.md)** - Real-time collaboration

### **🧪 Testing Resources**
- **[API Documentation](http://localhost:8080/swagger/)** - Interactive API documentation (after starting services)
- **[Testing Strategies](docs/development/LOCAL_DEVELOPMENT_COMPLETE_GUIDE.md#-testing-guidelines)** - Unit, integration, and E2E testing
- **[Load Testing](docs/development/LOCAL_DEVELOPMENT_COMPLETE_GUIDE.md#performance-testing)** - Performance testing guides

## Services

### Core Services

1. **Auth Service** (Node.js + JWT/OAuth)
   - User authentication and authorization
   - JWT token management
   - Google OAuth integration

2. **Form Service** (Go + PostgreSQL)
   - CRUD operations for forms
   - Form schema management
   - User permissions

3. **Real-Time Service** (Go + WebSockets + Redis)
   - Live collaboration features
   - Real-time form editing
   - User presence tracking

4. **Response Service** (Node.js + Firestore)
   - Response collection and storage
   - Data validation
   - Response querying

5. **Analytics Service** (Python + BigQuery)
   - Response analytics and reporting
   - Data export (CSV, Google Sheets)
   - Statistical analysis

6. **File Service** (AWS Lambda + S3)
   - File upload handling
   - Image and document processing
   - Metadata management

## Getting Started

### Prerequisites

- Go 1.21+
- Node.js 18+
- Python 3.9+
- Docker & Docker Compose
- PostgreSQL
- Redis

### Local Development

1. Clone the repository:
```bash
git clone https://github.com/Mir00r/X-Form-Backend.git
cd X-Form-Backend
```

2. Start infrastructure services:
```bash
docker-compose up -d postgres redis
```

3. Set up environment variables:
```bash
cp .env.example .env
# Edit .env with your configuration
```

4. Start services:
```bash
# Auth Service
cd services/auth-service
npm install
npm run dev

# Form Service
cd services/form-service
go mod tidy
go run cmd/server/main.go

# Real-Time Service
cd services/realtime-service
go mod tidy
go run cmd/server/main.go

# Response Service
cd services/response-service
npm install
npm run dev

# Analytics Service
cd services/analytics-service
pip install -r requirements.txt
python main.py
```

## MVP Features

✅ **Implemented:**
- User authentication (email/password, Google OAuth)
- Form CRUD operations
- Real-time collaboration
- Response collection
- Basic analytics
- File uploads

🚧 **Planned:**
- Advanced analytics dashboards
- Team collaboration features
- Third-party integrations
- AI-powered features

## API Documentation

API documentation is available at `/docs` endpoint for each service when running in development mode.

## Deployment

The system is designed to run on Kubernetes with the following components:
- API Gateway (AWS API Gateway or NGINX Ingress)
- Microservices (EKS/GKE pods)
- Databases (AWS RDS, Firestore, ElastiCache)
- File Storage (S3)

See `deployment/` directory for Kubernetes manifests and Terraform configurations.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License - see LICENSE file for details.
