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
