# API Gateway - Swagger Documentation Setup

## Overview

This API Gateway provides a unified entry point for all microservices in the X-Form Backend system. It includes comprehensive Swagger/OpenAPI 3.0 documentation for all endpoints.

## Features

- **Comprehensive API Documentation**: Complete OpenAPI 3.0 specification with Swagger UI
- **JWT Authentication**: Bearer token-based authentication with role-based access control
- **Health Monitoring**: Health check endpoints and Prometheus metrics
- **Request/Response Validation**: Input validation and standardized response formats
- **Middleware Stack**: Logging, CORS, rate limiting, and authentication middleware
- **Service Routing**: Routes to auth, form, response, and analytics services

## API Endpoints

### System Endpoints
- `GET /health` - Health check
- `GET /metrics` - Prometheus metrics
- `GET /swagger/*` - Swagger UI documentation

### Authentication Service (`/api/v1/auth`)
- `POST /register` - User registration
- `POST /login` - User authentication
- `POST /logout` - User logout
- `POST /refresh` - Token refresh
- `GET /profile` - Get user profile
- `PUT /profile` - Update user profile
- `DELETE /profile` - Delete user account

### Form Service (`/api/v1/forms`)
- `GET /forms` - List forms (with pagination)
- `POST /forms` - Create new form
- `GET /forms/{id}` - Get form details
- `PUT /forms/{id}` - Update form
- `DELETE /forms/{id}` - Delete form
- `POST /forms/{id}/publish` - Publish form
- `POST /forms/{id}/unpublish` - Unpublish form

### Response Service (`/api/v1/responses`)
- `GET /responses` - List responses (with filtering)
- `POST /responses/{formId}/submit` - Submit form response
- `GET /responses/{id}` - Get response details
- `PUT /responses/{id}` - Update response
- `DELETE /responses/{id}` - Delete response

### Analytics Service (`/api/v1/analytics`)
- `GET /analytics/forms/{formId}` - Get form analytics
- `GET /analytics/responses/{responseId}` - Get response analytics
- `GET /analytics/dashboard` - Get dashboard analytics

## Data Models

### Standard Response Format
```json
{
  "success": true,
  "message": "Operation completed successfully",
  "data": {},
  "timestamp": "2025-09-06T12:00:00Z"
}
```

### Error Response Format
```json
{
  "success": false,
  "error": "Error description",
  "code": "ERROR_CODE",
  "timestamp": "2025-09-06T12:00:00Z"
}
```

## Authentication

The API uses JWT Bearer tokens for authentication. Include the token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

## Rate Limiting

- **Default**: 100 requests per minute per IP
- **Authenticated**: 1000 requests per minute per user
- **Headers**: Rate limit information included in response headers

## CORS Configuration

The API supports cross-origin requests with the following configuration:
- **Allowed Origins**: Configurable (default: all origins in development)
- **Allowed Methods**: GET, POST, PUT, DELETE, OPTIONS
- **Allowed Headers**: Content-Type, Authorization, X-Requested-With

## Error Handling

The API returns standardized error responses with appropriate HTTP status codes:

- `400 Bad Request` - Invalid request format or parameters
- `401 Unauthorized` - Missing or invalid authentication
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource conflict (e.g., duplicate email)
- `422 Unprocessable Entity` - Validation errors
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Server error

## Security Features

- **JWT Authentication**: Secure token-based authentication
- **Request Validation**: Input validation and sanitization
- **Rate Limiting**: Protection against abuse
- **CORS Protection**: Controlled cross-origin access
- **Security Headers**: Standard security headers included
- **Request Logging**: Comprehensive request/response logging

## Monitoring and Observability

- **Health Checks**: `/health` endpoint for service monitoring
- **Metrics**: Prometheus metrics at `/metrics` endpoint
- **Structured Logging**: JSON-formatted logs with correlation IDs
- **Request Tracing**: Request ID tracking across services

## Environment Configuration

The gateway supports the following environment variables:

```bash
# Server Configuration
PORT=8080
GIN_MODE=release  # Use 'debug' for development

# Authentication
JWT_SECRET=your-secret-key
JWT_EXPIRY=24h

# Rate Limiting
RATE_LIMIT_PER_MINUTE=100
RATE_LIMIT_BURST=10

# CORS
ALLOWED_ORIGINS=*
ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
ALLOWED_HEADERS=Content-Type,Authorization,X-Requested-With

# Service URLs (for future proxy implementation)
AUTH_SERVICE_URL=http://auth-service:3001
FORM_SERVICE_URL=http://form-service:3002
RESPONSE_SERVICE_URL=http://response-service:3003
ANALYTICS_SERVICE_URL=http://analytics-service:3004
```

## Swagger Documentation Features

### Interactive API Testing
- **Try It Out**: Test endpoints directly from the Swagger UI
- **Authentication**: JWT token input for authenticated endpoints
- **Request/Response Examples**: Complete examples for all endpoints
- **Schema Validation**: Real-time validation of request payloads

### Code Generation
- **Client SDKs**: Generate client libraries in multiple languages
- **API Contracts**: Export OpenAPI specification for contract testing
- **Mock Servers**: Use specification for mock server generation

### Documentation Quality
- **Complete Coverage**: All endpoints documented with examples
- **Rich Descriptions**: Detailed descriptions for all parameters and responses
- **Type Safety**: Strongly typed request/response models
- **Status Codes**: Complete HTTP status code documentation

## Development Workflow

1. **Add New Endpoint**: Create handler function with Swagger annotations
2. **Update Documentation**: Run `swag init` to regenerate docs
3. **Test Integration**: Use Swagger UI for endpoint testing
4. **Validate Changes**: Ensure all tests pass and documentation is complete

## Production Considerations

- **Security**: Update JWT secrets and configure CORS appropriately
- **Performance**: Configure rate limits based on expected load
- **Monitoring**: Set up proper logging and metrics collection
- **Documentation**: Keep Swagger documentation in sync with implementation
- **Versioning**: Use API versioning for backward compatibility
