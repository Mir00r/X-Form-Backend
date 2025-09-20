# Form Service API - Comprehensive Swagger Documentation

This repository contains a production-ready Form Service with comprehensive Swagger/OpenAPI documentation following industry best practices.

## 🚀 Features

- **Complete Form Management**: Create, update, delete, publish, and manage forms
- **Dynamic Questions**: Support for multiple question types (text, email, select, file uploads, etc.)
- **Form Publishing**: Publish forms with custom settings and access controls
- **Response Collection**: Collect and manage form responses
- **Statistics & Analytics**: Get detailed form statistics and insights
- **Clean Architecture**: Following SOLID principles and microservices patterns
- **Comprehensive API Documentation**: Full Swagger/OpenAPI 3.0 specification
- **Security**: JWT authentication, rate limiting, CORS protection
- **Health Monitoring**: Health checks and system metrics
- **Database Integration**: PostgreSQL with GORM ORM
- **Caching**: Redis integration for performance
- **Graceful Shutdown**: Proper server lifecycle management

## 📖 API Documentation

### Available Endpoints

The service provides comprehensive Swagger documentation accessible through multiple endpoints:

- **Swagger UI**: `http://localhost:8080/swagger/index.html`
- **OpenAPI JSON**: `http://localhost:8080/swagger/doc.json`
- **OpenAPI YAML**: `http://localhost:8080/docs/swagger.yaml`
- **Health Check**: `http://localhost:8080/health`
- **Service Info**: `http://localhost:8080/`

### Form Management APIs

#### 📝 Forms
- `POST /api/v1/forms` - Create a new form
- `GET /api/v1/forms` - List forms with filtering and pagination
- `GET /api/v1/forms/{id}` - Get a specific form
- `PUT /api/v1/forms/{id}` - Update a form
- `DELETE /api/v1/forms/{id}` - Delete a form
- `POST /api/v1/forms/{id}/publish` - Publish a form
- `POST /api/v1/forms/{id}/close` - Close a form
- `POST /api/v1/forms/{id}/archive` - Archive a form
- `GET /api/v1/forms/{id}/statistics` - Get form statistics
- `POST /api/v1/forms/{id}/duplicate` - Duplicate a form

#### 🌍 Public Access
- `GET /api/v1/public/forms/{id}` - Get public form (no authentication)

#### 🏥 Health & Monitoring
- `GET /api/v1/health` - API health check
- `GET /api/v1/health/ready` - Readiness probe
- `GET /api/v1/health/live` - Liveness probe
- `GET /api/v1/metrics` - System metrics

## 🛠 Installation & Setup

### Prerequisites

- Go 1.23 or higher
- PostgreSQL 12+ (optional for demo)
- Redis 6+ (optional for demo)

### Quick Start

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd services/form-service
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

4. **Generate Swagger documentation**
   ```bash
   go install github.com/swaggo/swag/cmd/swag@latest
   swag init -g cmd/demo-swagger-server/main.go -o docs
   ```

5. **Run the demo server (no database required)**
   ```bash
   go run cmd/demo-swagger-server/main.go
   ```

6. **Run the full server (requires database)**
   ```bash
   go run cmd/full-swagger-server/main.go
   ```

### Environment Configuration

Create a `.env` file with the following variables:

```env
# Server Configuration
PORT=8080
NODE_ENV=development

# Database Configuration (optional for demo)
DATABASE_URL=postgresql://username:password@localhost:5432/database_name?sslmode=disable

# Redis Configuration (optional for demo)
REDIS_URL=redis://localhost:6379

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-for-development-only
```

## 🏃‍♂️ Running the Application

### Demo Server (Recommended for Testing)

The demo server runs without external dependencies and provides full Swagger documentation:

```bash
# Build and run demo server
go build -o bin/demo-swagger-server cmd/demo-swagger-server/main.go
./bin/demo-swagger-server
```

**Access Points:**
- Swagger UI: http://localhost:8080/swagger/index.html
- Health Check: http://localhost:8080/health
- Service Info: http://localhost:8080/

### Full Production Server

For production use with database and Redis:

```bash
# Ensure PostgreSQL and Redis are running
# Update .env with your database credentials

# Build and run full server
go build -o bin/full-swagger-server cmd/full-swagger-server/main.go
./bin/full-swagger-server
```

### Using Docker (Optional)

```bash
# Build Docker image
docker build -t form-service .

# Run with environment variables
docker run -p 8080:8080 \
  -e DATABASE_URL="your-database-url" \
  -e REDIS_URL="your-redis-url" \
  form-service
```

## 📚 API Usage Examples

### Create a Form

```bash
curl -X POST "http://localhost:8080/api/v1/forms" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-jwt-token" \
  -d '{
    "title": "Customer Feedback Form",
    "description": "Please provide your feedback",
    "questions": [
      {
        "type": "text",
        "label": "Your Name",
        "required": true,
        "order": 1
      },
      {
        "type": "email",
        "label": "Email Address",
        "required": true,
        "order": 2
      }
    ]
  }'
```

### List Forms

```bash
curl -X GET "http://localhost:8080/api/v1/forms?page=1&pageSize=20" \
  -H "Authorization: Bearer your-jwt-token"
```

### Get Form Statistics

```bash
curl -X GET "http://localhost:8080/api/v1/forms/{form-id}/statistics" \
  -H "Authorization: Bearer your-jwt-token"
```

## 🔧 Development

### Regenerating Swagger Documentation

After making changes to API annotations:

```bash
swag init -g cmd/demo-swagger-server/main.go -o docs
```

### Adding New Endpoints

1. Add Swagger annotations to your handler functions:
   ```go
   // @Summary Create a new form
   // @Description Create a new form with questions and settings
   // @Tags Forms
   // @Accept json
   // @Produce json
   // @Security BearerAuth
   // @Param request body dto.CreateFormRequestDTO true "Form creation request"
   // @Success 201 {object} dto.SuccessResponse{data=dto.FormResponseDTO}
   // @Failure 400 {object} dto.ErrorResponse
   // @Router /forms [post]
   func (h *FormHandler) CreateForm(c *gin.Context) {
       // Implementation
   }
   ```

2. Regenerate documentation:
   ```bash
   swag init -g cmd/demo-swagger-server/main.go -o docs
   ```

3. Restart the server to see changes

### Testing the API

The Swagger UI provides an interactive interface for testing all endpoints:

1. Open http://localhost:8080/swagger/index.html
2. Expand any endpoint section
3. Click "Try it out"
4. Fill in parameters and request body
5. Click "Execute"

## 🏗 Architecture

### Clean Architecture Layers

```
cmd/                    # Application entry points
├── demo-swagger-server/   # Demo server (no external dependencies)
└── full-swagger-server/   # Production server (with database)

internal/
├── application/        # Application services (use cases)
├── domain/            # Domain entities and business logic
├── infrastructure/    # External concerns (database, redis)
├── handlers/          # HTTP handlers (controllers)
├── dto/              # Data Transfer Objects
├── middleware/       # HTTP middleware
├── routes/           # Route definitions
├── validation/       # Input validation
├── repository/       # Data access layer
└── config/           # Configuration management

docs/                 # Generated Swagger documentation
├── docs.go          # Go bindings
├── swagger.json     # OpenAPI JSON
└── swagger.yaml     # OpenAPI YAML
```

### Dependencies

- **Web Framework**: Gin (high-performance HTTP router)
- **Database**: PostgreSQL with GORM ORM
- **Cache**: Redis
- **Documentation**: Swaggo for Swagger generation
- **Logging**: Logrus for structured logging
- **Configuration**: Godotenv for environment management

## 🔒 Security

- **Authentication**: JWT Bearer token authentication
- **Rate Limiting**: Request rate limiting by IP
- **CORS**: Cross-Origin Resource Sharing protection
- **Security Headers**: Standard security headers
- **Input Validation**: Comprehensive request validation
- **SQL Injection Protection**: GORM with prepared statements

## 📊 Monitoring

### Health Checks

- **Basic Health**: `GET /health`
- **Readiness Probe**: `GET /api/v1/health/ready`
- **Liveness Probe**: `GET /api/v1/health/live`

### Metrics

- **System Metrics**: `GET /api/v1/metrics`
- **Request Metrics**: Automatic request/response logging
- **Performance Metrics**: Response time tracking

## 🐛 Troubleshooting

### Common Issues

1. **Swagger documentation not loading**
   - Ensure docs are generated: `swag init -g cmd/demo-swagger-server/main.go -o docs`
   - Check if docs directory exists and contains files

2. **Database connection errors**
   - Use demo server for testing without database
   - Check DATABASE_URL format and credentials

3. **Port already in use**
   - Change PORT in .env file
   - Kill existing processes: `lsof -ti:8080 | xargs kill`

### Debug Mode

Run with debug logging:
```bash
GIN_MODE=debug go run cmd/demo-swagger-server/main.go
```

## 📞 Support

- **Documentation**: Check Swagger UI at `/swagger/index.html`
- **Health Status**: Monitor `/health` endpoint
- **Service Info**: View available endpoints at `/`

## 🎯 Best Practices Implemented

- ✅ **RESTful API Design**: Proper HTTP methods and status codes
- ✅ **Clean Architecture**: Separation of concerns and SOLID principles  
- ✅ **Comprehensive Documentation**: Full OpenAPI 3.0 specification
- ✅ **Error Handling**: Structured error responses with details
- ✅ **Input Validation**: Request validation with detailed error messages
- ✅ **Security**: Authentication, rate limiting, and security headers
- ✅ **Monitoring**: Health checks and metrics endpoints
- ✅ **Graceful Shutdown**: Proper server lifecycle management
- ✅ **Structured Logging**: Consistent logging with context
- ✅ **Configuration Management**: Environment-based configuration
- ✅ **Database Migrations**: Automatic schema management
- ✅ **API Versioning**: Clear API versioning strategy

---

## 🚀 Quick Links

- **Swagger UI**: http://localhost:8080/swagger/index.html
- **Health Check**: http://localhost:8080/health
- **Service Info**: http://localhost:8080/
- **API Base**: http://localhost:8080/api/v1

Start the server and explore the comprehensive API documentation!
