#!/bin/bash

# Enhanced Architecture Environment Setup Script
# This script sets up the development environment without requiring Docker

set -e

echo "🚀 Setting up Enhanced Architecture Development Environment..."

# Create necessary directories
echo "📁 Creating directory structure..."
mkdir -p api-gateway/logs
mkdir -p api-gateway/certs
mkdir -p api-gateway/config

# Create development configuration
echo "⚙️  Creating development configuration..."
cat > api-gateway/config/dev.yaml << 'EOF'
# Development Configuration for Enhanced API Gateway

environment: development
log:
  level: debug
  format: json

server:
  port: 8080
  timeout: 30s
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 60s

auth:
  jwt:
    secret: "dev-secret-key-for-development-only"
    expiration: 24h
    issuer: "x-form-api-gateway"
    audience: "x-form-backend"

cors:
  enabled: true
  allowed_origins:
    - "http://localhost:3000"
    - "http://localhost:8080"
  allowed_methods:
    - "GET"
    - "POST"
    - "PUT"
    - "DELETE"
    - "OPTIONS"
  allowed_headers:
    - "Content-Type"
    - "Authorization"
    - "X-Request-ID"

validation:
  enabled: true
  max_body_size: 10485760  # 10MB
  timeout: 5s

whitelist:
  enabled: false  # Disabled for development
  ips: []

rate_limit:
  enabled: true
  requests_per_second: 100
  burst: 200
  cleanup_interval: 60s

services:
  auth-service:
    url: "http://localhost:8081"
    timeout: 30s
  form-service:
    url: "http://localhost:8082"
    timeout: 30s
  response-service:
    url: "http://localhost:8083"
    timeout: 30s
EOF

# Create environment file
echo "🔧 Creating environment file..."
cat > .env << 'EOF'
# Environment Configuration
ENV=development
LOG_LEVEL=debug
JWT_SECRET=dev-secret-key-for-development-only
SERVER_PORT=8080

# Database (when needed)
DATABASE_URL=postgres://xform:password@localhost:5432/xform_dev

# Redis (when needed)
REDIS_URL=redis://localhost:6379

# Monitoring
PROMETHEUS_ENABLED=true
METRICS_PORT=9090
EOF

# Create self-signed certificates for development
echo "🔐 Creating development certificates..."
if command -v openssl &> /dev/null; then
    if [ ! -f api-gateway/certs/server.crt ]; then
        openssl req -x509 -newkey rsa:4096 -keyout api-gateway/certs/server.key -out api-gateway/certs/server.crt -days 365 -nodes -subj "/C=US/ST=Dev/L=Dev/O=X-Form/CN=localhost"
        echo "✅ Development certificates created"
    else
        echo "✅ Development certificates already exist"
    fi
else
    echo "⚠️  OpenSSL not found, skipping certificate generation"
fi

# Initialize go modules if not done
echo "📦 Setting up Go modules..."
cd api-gateway
if [ ! -f go.sum ]; then
    go mod tidy
    echo "✅ Go modules initialized"
else
    echo "✅ Go modules already set up"
fi

# Create a simple test to verify setup
echo "🧪 Creating verification test..."
cat > cmd/server/main_test.go << 'EOF'
package main

import (
	"testing"
	"os"
)

func TestMain(m *testing.M) {
	// Setup test environment
	os.Setenv("ENV", "test")
	os.Setenv("LOG_LEVEL", "error")
	
	// Run tests
	code := m.Run()
	
	// Cleanup
	os.Exit(code)
}

func TestApplicationCreation(t *testing.T) {
	// Set required environment variables for test
	os.Setenv("ENV", "test")
	os.Setenv("JWT_SECRET", "test-secret")
	
	app, err := NewApplication()
	if err != nil {
		t.Fatalf("Failed to create application: %v", err)
	}
	
	if app == nil {
		t.Fatal("Application should not be nil")
	}
	
	if app.config == nil {
		t.Fatal("Application config should not be nil")
	}
}
EOF

cd ..

# Create a simple run script
echo "🏃 Creating run scripts..."
cat > run-dev.sh << 'EOF'
#!/bin/bash

echo "🚀 Starting Enhanced API Gateway in Development Mode..."

# Set environment variables
export ENV=development
export LOG_LEVEL=debug
export CONFIG_PATH=./api-gateway/config/dev.yaml

# Build and run
cd api-gateway
go build -o bin/gateway ./cmd/server
echo "✅ Build completed"

echo "🌐 Starting server on http://localhost:8080"
echo "📊 Health check: http://localhost:8080/health"
echo "📈 Metrics: http://localhost:8080/metrics"
echo ""
echo "Press Ctrl+C to stop..."

./bin/gateway
EOF

chmod +x run-dev.sh

# Create test script
cat > test-all.sh << 'EOF'
#!/bin/bash

echo "🧪 Running Enhanced API Gateway Tests..."

cd api-gateway

# Run tests with coverage
echo "📊 Running tests with coverage..."
go test -v -race -coverprofile=coverage.out ./...

# Generate coverage report
if [ -f coverage.out ]; then
    echo "📈 Generating coverage report..."
    go tool cover -html=coverage.out -o coverage.html
    echo "✅ Coverage report generated: coverage.html"
fi

# Run linting (if available)
if command -v golangci-lint &> /dev/null; then
    echo "🔍 Running linter..."
    golangci-lint run
else
    echo "⚠️  golangci-lint not found, skipping linting"
fi

echo "✅ All tests completed"
EOF

chmod +x test-all.sh

echo ""
echo "🎉 Enhanced Architecture Development Environment Setup Complete!"
echo ""
echo "📋 Next Steps:"
echo "   1. Start development server:   ./run-dev.sh"
echo "   2. Run tests:                  ./test-all.sh"
echo "   3. Check health:               curl http://localhost:8080/health"
echo "   4. View metrics:               curl http://localhost:8080/metrics"
echo ""
echo "📁 Files created:"
echo "   - api-gateway/config/dev.yaml  (Development configuration)"
echo "   - .env                         (Environment variables)"
echo "   - api-gateway/certs/           (Development certificates)"
echo "   - run-dev.sh                   (Development server script)"
echo "   - test-all.sh                  (Test execution script)"
echo ""
echo "🔧 To customize configuration, edit:"
echo "   - api-gateway/config/dev.yaml"
echo "   - .env"
echo ""
echo "Happy coding! 🚀"
EOF

chmod +x setup-dev-env.sh
