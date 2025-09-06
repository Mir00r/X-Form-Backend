#!/bin/bash

# X-Form API Gateway Startup Script
# This script starts the API Gateway with basic configuration

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default configuration
SERVER_HOST=${SERVER_HOST:-"localhost"}
SERVER_PORT=${SERVER_PORT:-"8080"}
SERVER_ENVIRONMENT=${SERVER_ENVIRONMENT:-"development"}
SECURITY_JWT_SECRET=${SECURITY_JWT_SECRET:-"your-super-secret-jwt-key-change-this-in-production"}

# Service URLs (update these to match your microservices)
SERVICES_AUTH_SERVICE_URL=${SERVICES_AUTH_SERVICE_URL:-"http://localhost:3001"}
SERVICES_FORM_SERVICE_URL=${SERVICES_FORM_SERVICE_URL:-"http://localhost:3002"}
SERVICES_RESPONSE_SERVICE_URL=${SERVICES_RESPONSE_SERVICE_URL:-"http://localhost:3003"}
SERVICES_ANALYTICS_SERVICE_URL=${SERVICES_ANALYTICS_SERVICE_URL:-"http://localhost:3004"}
SERVICES_COLLABORATION_SERVICE_URL=${SERVICES_COLLABORATION_SERVICE_URL:-"http://localhost:3005"}
SERVICES_REALTIME_SERVICE_URL=${SERVICES_REALTIME_SERVICE_URL:-"http://localhost:3006"}

# Optional integrations
TRAEFIK_ENABLED=${TRAEFIK_ENABLED:-"false"}
TYK_ENABLED=${TYK_ENABLED:-"false"}

echo -e "${BLUE}🚀 Starting X-Form API Gateway...${NC}"
echo -e "${YELLOW}📋 Configuration:${NC}"
echo -e "  Host: ${SERVER_HOST}"
echo -e "  Port: ${SERVER_PORT}"
echo -e "  Environment: ${SERVER_ENVIRONMENT}"
echo -e "  Traefik: ${TRAEFIK_ENABLED}"
echo -e "  Tyk: ${TYK_ENABLED}"
echo

# Build the application if binary doesn't exist
if [ ! -f "./bin/api-gateway" ]; then
    echo -e "${YELLOW}🔨 Building API Gateway...${NC}"
    go build -o bin/api-gateway cmd/server/main.go
    echo -e "${GREEN}✅ Build completed${NC}"
fi

# Export environment variables
export SERVER_HOST
export SERVER_PORT
export SERVER_ENVIRONMENT
export SECURITY_JWT_SECRET
export SERVICES_AUTH_SERVICE_URL
export SERVICES_FORM_SERVICE_URL
export SERVICES_RESPONSE_SERVICE_URL
export SERVICES_ANALYTICS_SERVICE_URL
export SERVICES_COLLABORATION_SERVICE_URL
export SERVICES_REALTIME_SERVICE_URL
export TRAEFIK_ENABLED
export TYK_ENABLED

# Enable metrics and health checks
export OBSERVABILITY_METRICS_ENABLED=true
export OBSERVABILITY_METRICS_PATH=/metrics

echo -e "${GREEN}🌐 Starting server...${NC}"
echo -e "${BLUE}📚 Swagger documentation will be available at: http://${SERVER_HOST}:${SERVER_PORT}/swagger/index.html${NC}"
echo -e "${BLUE}💚 Health check: http://${SERVER_HOST}:${SERVER_PORT}/health${NC}"
echo -e "${BLUE}📊 Metrics: http://${SERVER_HOST}:${SERVER_PORT}/metrics${NC}"
echo

# Start the API Gateway
./bin/api-gateway
