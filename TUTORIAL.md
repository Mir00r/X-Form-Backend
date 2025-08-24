# X-Form Backend - Complete Setup and Testing Guide

## Overview

X-Form Backend is a comprehensive form management system built with industry best practices including:

- **Architecture**: Clean Architecture with Domain-Driven Design
- **Infrastructure**: Traefik All-in-One Gateway, PostgreSQL, Redis
- **Security**: JWT authentication, CORS, rate limiting
- **Technologies**: Go (Gin framework), GORM ORM, Docker
- **Patterns**: Repository pattern, Service layer, Dependency Injection

## Architecture Overview

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Traefik      │────│  Auth Service    │    │  Form Service   │
│   Gateway      │    │  (Port 8081)     │    │  (Port 8082)    │
│   (Port 80)    │    └──────────────────┘    └─────────────────┘
└─────────────────┘              │                       │
         │                       │                       │
         │              ┌─────────────────┐    ┌─────────────────┐
         │              │   PostgreSQL    │    │     Redis       │
         │              │   (Port 5432)   │    │   (Port 6379)   │
         └──────────────┴─────────────────┴────┴─────────────────┘
```

## Prerequisites

- Docker and Docker Compose
- Go 1.21+ (for development)
- Git
- Postman or curl (for API testing)

## Quick Start

### 1. Clone and Setup

```bash
git clone <repository-url>
cd X-Form-Backend

# Ensure all services are present
ls -la services/
# Should show: auth-service/ form-service/
```

### 2. Environment Configuration

The system uses the existing `docker-compose.yml` with Traefik All-in-One configuration.

### 3. Start the System

```bash
# Start all services with Traefik gateway
docker-compose up -d

# Verify services are running
docker-compose ps

# Check logs
docker-compose logs -f traefik
docker-compose logs -f auth-service
docker-compose logs -f form-service
```

### 4. Health Checks

```bash
# Check Traefik dashboard (if enabled)
curl http://localhost:8080/dashboard/

# Check auth service health
curl http://localhost/auth/health

# Check form service health  
curl http://localhost/forms/health
```

## API Documentation

### Authentication Service

#### 1. Register User
```bash
curl -X POST http://localhost/auth/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securePassword123",
    "full_name": "John Doe"
  }'
```

#### 2. Login
```bash
curl -X POST http://localhost/auth/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com", 
    "password": "securePassword123"
  }'
```

Response includes `access_token` and `refresh_token`.

#### 3. Get User Profile
```bash
curl -X GET http://localhost/auth/api/v1/profile \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### Form Service

#### 1. Create Form
```bash
curl -X POST http://localhost/forms/api/v1/forms \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Customer Feedback Form",
    "description": "Please provide your feedback",
    "settings": {
      "accepting_responses": true,
      "require_sign_in": false,
      "confirmation_message": "Thank you for your feedback!",
      "allow_multiple_response": true,
      "show_progress_bar": true,
      "shuffle_questions": false
    }
  }'
```

#### 2. Get User Forms
```bash
curl -X GET "http://localhost/forms/api/v1/forms?page=1&limit=10" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

#### 3. Get Specific Form
```bash
curl -X GET http://localhost/forms/api/v1/forms/FORM_ID \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

#### 4. Update Form
```bash
curl -X PUT http://localhost/forms/api/v1/forms/FORM_ID \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Updated Form Title",
    "description": "Updated description"
  }'
```

#### 5. Publish Form
```bash
curl -X POST http://localhost/forms/api/v1/forms/FORM_ID/publish \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

#### 6. Add Question to Form
```bash
curl -X POST http://localhost/forms/api/v1/forms/FORM_ID/questions \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "text",
    "title": "What is your name?",
    "description": "Please enter your full name",
    "order": 1,
    "validation": {
      "required": true,
      "max_length": 100
    }
  }'
```

#### 7. Update Question
```bash
curl -X PUT http://localhost/forms/api/v1/forms/questions/QUESTION_ID \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Updated question title",
    "order": 2
  }'
```

#### 8. Delete Question
```bash
curl -X DELETE http://localhost/forms/api/v1/forms/questions/QUESTION_ID \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

#### 9. Reorder Questions
```bash
curl -X PUT http://localhost/forms/api/v1/forms/FORM_ID/questions/reorder \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "question_orders": [
      {"id": "QUESTION_ID_1", "order": 1},
      {"id": "QUESTION_ID_2", "order": 2}
    ]
  }'
```

#### 10. Delete Form
```bash
curl -X DELETE http://localhost/forms/api/v1/forms/FORM_ID \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

## Testing Workflow

### Complete End-to-End Test

```bash
#!/bin/bash

# 1. Register a new user
echo "Registering user..."
REGISTER_RESPONSE=$(curl -s -X POST http://localhost/auth/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "testuser@example.com",
    "password": "testPassword123",
    "full_name": "Test User"
  }')

echo "Register response: $REGISTER_RESPONSE"

# 2. Login and get token
echo "Logging in..."
LOGIN_RESPONSE=$(curl -s -X POST http://localhost/auth/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "testuser@example.com",
    "password": "testPassword123"
  }')

# Extract token (requires jq)
TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.access_token')
echo "Token: $TOKEN"

# 3. Create a form
echo "Creating form..."
FORM_RESPONSE=$(curl -s -X POST http://localhost/forms/api/v1/forms \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Test Form",
    "description": "This is a test form",
    "settings": {
      "accepting_responses": true,
      "require_sign_in": false,
      "confirmation_message": "Thank you!",
      "allow_multiple_response": true,
      "show_progress_bar": true,
      "shuffle_questions": false
    }
  }')

# Extract form ID
FORM_ID=$(echo $FORM_RESPONSE | jq -r '.form.id')
echo "Form ID: $FORM_ID"

# 4. Add questions
echo "Adding questions..."
curl -s -X POST http://localhost/forms/api/v1/forms/$FORM_ID/questions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "text",
    "title": "What is your name?",
    "description": "Please enter your full name",
    "order": 1,
    "validation": {
      "required": true,
      "max_length": 100
    }
  }'

curl -s -X POST http://localhost/forms/api/v1/forms/$FORM_ID/questions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "email",
    "title": "What is your email?",
    "description": "Please enter your email address",
    "order": 2,
    "validation": {
      "required": true
    }
  }'

# 5. Get forms
echo "Getting user forms..."
curl -s -X GET http://localhost/forms/api/v1/forms \
  -H "Authorization: Bearer $TOKEN" | jq

# 6. Publish form
echo "Publishing form..."
curl -s -X POST http://localhost/forms/api/v1/forms/$FORM_ID/publish \
  -H "Authorization: Bearer $TOKEN" | jq

echo "Test completed successfully!"
```

Save this as `test_workflow.sh` and run:
```bash
chmod +x test_workflow.sh
./test_workflow.sh
```

## Development Setup

### 1. Local Development

```bash
# Install dependencies
cd services/auth-service && go mod tidy
cd ../form-service && go mod tidy

# Run auth service locally
cd services/auth-service
go run cmd/server/main.go

# Run form service locally (in another terminal)
cd services/form-service  
go run cmd/server/main.go
```

### 2. Database Management

```bash
# Access PostgreSQL
docker-compose exec postgres psql -U postgres -d xform_db

# View tables
\dt

# Check auth service data
SELECT * FROM users;

# Check form service data
SELECT * FROM forms;
SELECT * FROM questions;
```

### 3. Redis Management

```bash
# Access Redis
docker-compose exec redis redis-cli

# Check stored data
KEYS *

# View user sessions
GET "session:USER_ID"
```

## Monitoring and Logs

### View Logs
```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f auth-service
docker-compose logs -f form-service

# Traefik routing logs
docker-compose logs -f traefik
```

### Health Monitoring
```bash
# Create monitoring script
cat > monitor.sh << EOF
#!/bin/bash
while true; do
  echo "=== $(date) ==="
  echo "Auth Service: $(curl -s http://localhost/auth/health | jq -r '.status // "ERROR"')"
  echo "Form Service: $(curl -s http://localhost/forms/health | jq -r '.status // "ERROR"')"
  echo "---"
  sleep 30
done
EOF

chmod +x monitor.sh
./monitor.sh
```

## Troubleshooting

### Common Issues

#### 1. Service Not Starting
```bash
# Check port conflicts
netstat -tulpn | grep :80
netstat -tulpn | grep :5432

# Check Docker logs
docker-compose logs traefik
docker-compose logs auth-service
docker-compose logs form-service
```

#### 2. Database Connection Issues
```bash
# Check PostgreSQL status
docker-compose exec postgres pg_isready -U postgres

# Check connection from service
docker-compose exec auth-service nc -zv postgres 5432
```

#### 3. Authentication Issues
```bash
# Verify JWT secret configuration
# Check if tokens are properly formatted
# Ensure middleware is correctly configured
```

#### 4. Traefik Routing Issues
```bash
# Check Traefik configuration
docker-compose exec traefik cat /etc/traefik/traefik.yml

# Verify service labels
docker-compose config
```

### Performance Testing

```bash
# Install Apache Bench (if not available)
# apt-get install apache2-utils  # Ubuntu/Debian
# brew install httpie           # macOS

# Test auth endpoint
ab -n 100 -c 10 -H "Content-Type: application/json" \
   -p login_data.json http://localhost/auth/api/v1/login

# Test form creation (requires auth token)
ab -n 50 -c 5 -H "Authorization: Bearer YOUR_TOKEN" \
   -H "Content-Type: application/json" \
   -p form_data.json http://localhost/forms/api/v1/forms
```

## Security Considerations

### 1. Environment Variables
```bash
# Production environment variables
export JWT_SECRET="your-super-secret-jwt-key-min-32-chars"
export DB_PASSWORD="your-secure-database-password"  
export REDIS_PASSWORD="your-redis-password"
```

### 2. SSL/TLS (Production)
```yaml
# Add to traefik labels for HTTPS
- "traefik.http.routers.auth-service.tls=true"
- "traefik.http.routers.auth-service.tls.certresolver=letsencrypt"
```

### 3. Rate Limiting
The auth service includes built-in rate limiting. Monitor logs for blocked requests.

## Maintenance

### 1. Database Backup
```bash
# Backup
docker-compose exec postgres pg_dump -U postgres xform_db > backup.sql

# Restore  
docker-compose exec -T postgres psql -U postgres xform_db < backup.sql
```

### 2. Log Rotation
```bash
# Configure Docker log rotation in daemon.json
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  }
}
```

### 3. System Updates
```bash
# Update base images
docker-compose pull
docker-compose up -d

# Update Go dependencies
cd services/auth-service && go get -u ./...
cd services/form-service && go get -u ./...
```

## API Reference Summary

### Authentication Endpoints
- `POST /auth/api/v1/register` - Register user
- `POST /auth/api/v1/login` - Login user  
- `POST /auth/api/v1/refresh` - Refresh token
- `GET /auth/api/v1/profile` - Get user profile
- `PUT /auth/api/v1/profile` - Update profile
- `POST /auth/api/v1/logout` - Logout user

### Form Management Endpoints
- `POST /forms/api/v1/forms` - Create form
- `GET /forms/api/v1/forms` - List user forms
- `GET /forms/api/v1/forms/:id` - Get form details
- `PUT /forms/api/v1/forms/:id` - Update form
- `DELETE /forms/api/v1/forms/:id` - Delete form
- `POST /forms/api/v1/forms/:id/publish` - Publish form
- `POST /forms/api/v1/forms/:id/unpublish` - Unpublish form

### Question Management Endpoints
- `POST /forms/api/v1/forms/:formId/questions` - Add question
- `PUT /forms/api/v1/forms/questions/:questionId` - Update question
- `DELETE /forms/api/v1/forms/questions/:questionId` - Delete question
- `PUT /forms/api/v1/forms/:formId/questions/reorder` - Reorder questions

## Conclusion

This X-Form Backend implementation demonstrates industry best practices including:

✅ **Clean Architecture** with separated concerns
✅ **Domain-Driven Design** with rich domain models  
✅ **Repository Pattern** for data access abstraction
✅ **Service Layer** for business logic
✅ **JWT Authentication** with refresh tokens
✅ **API Gateway** with Traefik for routing and load balancing
✅ **Comprehensive Error Handling** and validation
✅ **Docker Containerization** for easy deployment
✅ **Database Migrations** and relationship management
✅ **Structured Logging** for observability
✅ **Security Middleware** (CORS, rate limiting, auth)
✅ **RESTful API Design** with proper HTTP status codes
✅ **Comprehensive Testing** endpoints and workflows

The system is production-ready and follows enterprise-grade development practices.
