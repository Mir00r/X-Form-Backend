# API Gateway - Complete Setup and Run Guide

## Quick Start

### 1. Prerequisites
```bash
# Ensure Go is installed (version 1.19 or higher)
go version

# Ensure you're in the api-gateway directory
cd /path/to/X-Form-Backend/services/api-gateway
```

### 2. Install Dependencies
```bash
# Download and install all Go dependencies
go mod download
go mod tidy
```

### 3. Install Swagger Tool (if not already installed)
```bash
# Install the swag command line tool
go install github.com/swaggo/swag/cmd/swag@latest

# Verify installation
which swag || echo "Add $GOPATH/bin to your PATH"
```

### 4. Generate Swagger Documentation
```bash
# Generate Swagger docs from annotations
/Users/mir00r/go/bin/swag init -g cmd/server/main.go -o docs/

# Alternative if swag is in PATH
swag init -g cmd/server/main.go -o docs/
```

### 5. Build the Application
```bash
# Build the binary
go build -o bin/api-gateway cmd/server/main.go

# Or build and run directly
go run cmd/server/main.go
```

### 6. Run the Application
```bash
# Method 1: Run directly
go run cmd/server/main.go

# Method 2: Run built binary
./bin/api-gateway

# The server will start on port 8080
```

### 7. Access Swagger Documentation
Open your browser and navigate to:
```
http://localhost:8080/swagger/index.html
```

## Detailed Setup Instructions

### Environment Setup

1. **Create environment file** (optional):
```bash
# Create .env file for local development
cat > .env << EOF
PORT=8080
GIN_MODE=debug
JWT_SECRET=your-super-secret-key-here
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001
EOF
```

2. **Load environment variables**:
```bash
# If using .env file
export $(grep -v '^#' .env | xargs)
```

### Development Workflow

1. **Make changes** to handlers or add new endpoints
2. **Add Swagger annotations** to new functions
3. **Regenerate documentation**:
```bash
swag init -g cmd/server/main.go -o docs/
```
4. **Test the changes**:
```bash
go run cmd/server/main.go
```

### Build for Production

```bash
# Set production mode
export GIN_MODE=release

# Build optimized binary
go build -ldflags="-w -s" -o bin/api-gateway cmd/server/main.go

# Run production build
./bin/api-gateway
```

## Verification Steps

### 1. Health Check
```bash
# Test health endpoint
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "healthy",
  "service": "api-gateway",
  "version": "1.0.0",
  "timestamp": "2025-09-06T12:00:00Z"
}
```

### 2. Swagger Documentation
- Navigate to `http://localhost:8080/swagger/index.html`
- Verify all endpoints are documented
- Test the "Try it out" functionality

### 3. Metrics Endpoint
```bash
# Check Prometheus metrics
curl http://localhost:8080/metrics
```

### 4. Test Authentication Endpoints
```bash
# Test registration endpoint
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Test User","email":"test@example.com","password":"password123"}'

# Test login endpoint
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'
```

## Troubleshooting

### Common Issues

1. **"swag command not found"**
```bash
# Add Go bin to PATH
export PATH=$PATH:$(go env GOPATH)/bin

# Or use full path
/Users/mir00r/go/bin/swag init -g cmd/server/main.go -o docs/
```

2. **"Port already in use"**
```bash
# Find process using port 8080
lsof -i :8080

# Kill the process
kill -9 <PID>

# Or use different port
PORT=8081 go run cmd/server/main.go
```

3. **"Module not found" errors**
```bash
# Clean and reinstall dependencies
go clean -modcache
go mod download
go mod tidy
```

4. **CORS errors in browser**
```bash
# Set allowed origins
export CORS_ALLOWED_ORIGINS=http://localhost:3000
```

### Development Tips

1. **Auto-reload during development**:
```bash
# Install air for hot reloading
go install github.com/cosmtrek/air@latest

# Create .air.toml config and run
air
```

2. **View detailed logs**:
```bash
# Run with debug mode
GIN_MODE=debug go run cmd/server/main.go
```

3. **Test with different HTTP methods**:
```bash
# Using httpie (install with: brew install httpie)
http GET localhost:8080/health
http POST localhost:8080/api/v1/auth/register name="Test" email="test@example.com" password="test123"
```

## File Structure

```
api-gateway/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── handlers/
│   │   └── handlers.go          # HTTP handlers with Swagger annotations
│   ├── middleware/              # Middleware functions
│   ├── config/                  # Configuration management
│   └── gateway/                 # Gateway-specific logic
├── docs/                        # Generated Swagger documentation
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── bin/                         # Built binaries
├── go.mod                       # Go module definition
├── go.sum                       # Go module checksums
└── README_SWAGGER_COMPLETE.md   # This documentation
```

## Next Steps

1. **Service Integration**: Implement actual service proxy logic
2. **Authentication**: Connect to real auth service
3. **Database**: Add database connectivity if needed
4. **Testing**: Add comprehensive test suite
5. **Docker**: Create Docker images for deployment
6. **CI/CD**: Set up continuous integration and deployment

## Support

For issues or questions:
1. Check the Swagger documentation at `/swagger/index.html`
2. Review the application logs
3. Verify environment configuration
4. Test endpoints using the provided curl examples

## Best Practices

1. **Always regenerate docs** after adding new endpoints
2. **Use proper HTTP status codes** in responses
3. **Include comprehensive examples** in Swagger annotations
4. **Validate input data** in all handlers
5. **Use structured logging** for better debugging
6. **Keep documentation up to date** with implementation
