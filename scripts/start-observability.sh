#!/bin/bash

# X-Form-Backend Observability Quick Start Script
# This script sets up the complete observability infrastructure

set -e

echo "🚀 Starting X-Form-Backend Observability Infrastructure Setup..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}❌ Docker is not running. Please start Docker and try again.${NC}"
    exit 1
fi

# Check if Docker Compose is available
if ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}❌ Docker Compose is not installed. Please install Docker Compose and try again.${NC}"
    exit 1
fi

echo -e "${BLUE}📋 Checking prerequisites...${NC}"
echo -e "${GREEN}✅ Docker is running${NC}"
echo -e "${GREEN}✅ Docker Compose is available${NC}"

# Create necessary directories
echo -e "${BLUE}📁 Creating necessary directories...${NC}"
mkdir -p data/prometheus
mkdir -p data/grafana
mkdir -p data/jaeger
mkdir -p data/tempo
mkdir -p logs

# Set proper permissions
chmod -R 777 data/
chmod -R 777 logs/

echo -e "${GREEN}✅ Directories created${NC}"

# Stop any existing observability services
echo -e "${BLUE}🛑 Stopping any existing observability services...${NC}"
docker-compose -f infrastructure/observability-infrastructure.yml down 2>/dev/null || true

# Start the observability infrastructure
echo -e "${BLUE}🚀 Starting observability infrastructure...${NC}"
docker-compose -f infrastructure/observability-infrastructure.yml up -d

# Wait for services to be ready
echo -e "${BLUE}⏳ Waiting for services to be ready...${NC}"
sleep 30

# Check service health
echo -e "${BLUE}🔍 Checking service health...${NC}"

services=(
    "http://localhost:4318/v1/traces:OTEL Collector"
    "http://localhost:16686:Jaeger UI"
    "http://localhost:3200/ready:Tempo"
    "http://localhost:9090/-/ready:Prometheus"
    "http://localhost:3000/api/health:Grafana"
    "http://localhost:9093/-/ready:AlertManager"
)

all_healthy=true

for service in "${services[@]}"; do
    url=$(echo $service | cut -d: -f1,2)
    name=$(echo $service | cut -d: -f3)
    
    if curl -s -f "$url" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ $name is healthy${NC}"
    else
        echo -e "${RED}❌ $name is not responding${NC}"
        all_healthy=false
    fi
done

echo ""
echo -e "${BLUE}🎯 Observability Infrastructure Status:${NC}"

if [ "$all_healthy" = true ]; then
    echo -e "${GREEN}✅ All services are running successfully!${NC}"
else
    echo -e "${YELLOW}⚠️  Some services may still be starting up. Check the logs if issues persist.${NC}"
fi

echo ""
echo -e "${BLUE}📊 Access Your Monitoring Dashboards:${NC}"
echo -e "${GREEN}🔍 Jaeger UI (Distributed Tracing): ${BLUE}http://localhost:16686${NC}"
echo -e "${GREEN}📈 Grafana (Metrics & Dashboards): ${BLUE}http://localhost:3000${NC} (admin/admin)"
echo -e "${GREEN}📊 Prometheus (Raw Metrics): ${BLUE}http://localhost:9090${NC}"
echo -e "${GREEN}🚨 AlertManager (Alerts): ${BLUE}http://localhost:9093${NC}"
echo -e "${GREEN}🔌 OTEL Collector (Traces): ${BLUE}http://localhost:4318${NC}"
echo -e "${GREEN}🏷️  Tempo (Trace Storage): ${BLUE}http://localhost:3200${NC}"

echo ""
echo -e "${BLUE}🛠️  API Gateway Configuration:${NC}"
echo "Add these environment variables to your API Gateway service:"
echo ""
echo "OTEL_SERVICE_NAME=api-gateway"
echo "OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318"
echo "OTEL_SERVICE_VERSION=1.0.0"
echo "OTEL_ENVIRONMENT=development"
echo "PROMETHEUS_ENABLED=true"

echo ""
echo -e "${BLUE}🚀 Quick Test Commands:${NC}"
echo "# Test API Gateway with observability:"
echo "cd services/api-gateway && go run ./cmd/server"
echo ""
echo "# Send test requests:"
echo "curl http://localhost:8080/health"
echo "curl http://localhost:8080/api/v1/forms"
echo ""
echo "# View traces in Jaeger:"
echo "open http://localhost:16686"

echo ""
echo -e "${BLUE}📚 Documentation:${NC}"
echo "📖 Full implementation guide: OBSERVABILITY_IMPLEMENTATION_GUIDE.md"
echo "🔧 Service integration: See guide for remaining services"

echo ""
echo -e "${BLUE}🔧 Management Commands:${NC}"
echo "# Stop observability infrastructure:"
echo "docker-compose -f infrastructure/observability-infrastructure.yml down"
echo ""
echo "# View logs:"
echo "docker-compose -f infrastructure/observability-infrastructure.yml logs -f"
echo ""
echo "# Restart services:"
echo "docker-compose -f infrastructure/observability-infrastructure.yml restart"

echo ""
echo -e "${GREEN}🎉 Observability infrastructure is ready!${NC}"
echo -e "${YELLOW}💡 Next steps: Integrate observability into your remaining microservices${NC}"
echo -e "${BLUE}📋 Follow the OBSERVABILITY_IMPLEMENTATION_GUIDE.md for detailed instructions${NC}"
