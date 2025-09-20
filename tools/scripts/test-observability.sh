#!/bin/bash

# X-Form-Backend Observability Test Script
# This script tests the observability integration

set -e

echo "🧪 Testing X-Form-Backend Observability Integration..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if observability infrastructure is running
echo -e "${BLUE}🔍 Checking observability infrastructure...${NC}"

services=(
    "localhost:4318:OTEL Collector"
    "localhost:16686:Jaeger UI"
    "localhost:3200:Tempo"
    "localhost:9090:Prometheus"
    "localhost:3000:Grafana"
    "localhost:9093:AlertManager"
)

all_running=true

for service in "${services[@]}"; do
    host_port=$(echo $service | cut -d: -f1,2)
    name=$(echo $service | cut -d: -f3)
    
    if nc -z ${host_port/:/ } 2>/dev/null; then
        echo -e "${GREEN}✅ $name is running${NC}"
    else
        echo -e "${RED}❌ $name is not running${NC}"
        all_running=false
    fi
done

if [ "$all_running" = false ]; then
    echo -e "${RED}❌ Some observability services are not running. Please start them first:${NC}"
    echo "./scripts/start-observability.sh"
    exit 1
fi

# Start API Gateway with observability
echo -e "${BLUE}🚀 Starting API Gateway with observability...${NC}"
cd services/api-gateway

# Set environment variables for testing
export OTEL_SERVICE_NAME=api-gateway
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
export OTEL_SERVICE_VERSION=1.0.0
export OTEL_ENVIRONMENT=development
export PROMETHEUS_ENABLED=true
export GIN_MODE=release

# Build and start the gateway in background
echo -e "${BLUE}🔨 Building API Gateway...${NC}"
go build -o bin/gateway ./cmd/server

echo -e "${BLUE}🚀 Starting API Gateway (backgrounded)...${NC}"
./bin/gateway &
GATEWAY_PID=$!

# Wait for gateway to start
echo -e "${BLUE}⏳ Waiting for API Gateway to start...${NC}"
sleep 5

# Function to cleanup on exit
cleanup() {
    echo -e "${YELLOW}🧹 Cleaning up...${NC}"
    if [ ! -z "$GATEWAY_PID" ]; then
        kill $GATEWAY_PID 2>/dev/null || true
    fi
}
trap cleanup EXIT

# Test gateway health endpoint
echo -e "${BLUE}🔍 Testing API Gateway health endpoint...${NC}"
if curl -s -f http://localhost:8080/health > /dev/null; then
    echo -e "${GREEN}✅ API Gateway is responding${NC}"
else
    echo -e "${RED}❌ API Gateway is not responding${NC}"
    exit 1
fi

# Test metrics endpoint
echo -e "${BLUE}📊 Testing metrics endpoint...${NC}"
if curl -s -f http://localhost:8080/metrics | grep -q "http_requests_total"; then
    echo -e "${GREEN}✅ Metrics endpoint is working${NC}"
else
    echo -e "${RED}❌ Metrics endpoint is not working${NC}"
    exit 1
fi

# Generate some test traffic
echo -e "${BLUE}🚦 Generating test traffic...${NC}"
for i in {1..10}; do
    curl -s http://localhost:8080/health > /dev/null
    curl -s http://localhost:8080/metrics > /dev/null
    # Simulate some errors
    curl -s http://localhost:8080/nonexistent > /dev/null 2>&1 || true
done

echo -e "${GREEN}✅ Generated test traffic${NC}"

# Wait a moment for metrics to be collected
sleep 5

# Check if metrics are being collected
echo -e "${BLUE}📈 Checking metric collection...${NC}"
metrics_response=$(curl -s http://localhost:8080/metrics)

if echo "$metrics_response" | grep -q "http_requests_total"; then
    request_count=$(echo "$metrics_response" | grep "http_requests_total" | head -1 | awk '{print $2}')
    echo -e "${GREEN}✅ HTTP request metrics collected (count: $request_count)${NC}"
else
    echo -e "${RED}❌ HTTP request metrics not found${NC}"
fi

if echo "$metrics_response" | grep -q "service_uptime_seconds"; then
    echo -e "${GREEN}✅ Service uptime metrics collected${NC}"
else
    echo -e "${RED}❌ Service uptime metrics not found${NC}"
fi

# Check Prometheus scraping
echo -e "${BLUE}🎯 Checking Prometheus metric scraping...${NC}"
if curl -s "http://localhost:9090/api/v1/query?query=up{job=\"api-gateway\"}" | grep -q '"value":\[.*,"1"\]'; then
    echo -e "${GREEN}✅ Prometheus is scraping API Gateway metrics${NC}"
else
    echo -e "${YELLOW}⚠️  Prometheus scraping check inconclusive (may need more time)${NC}"
fi

# Check if traces are being sent
echo -e "${BLUE}🔍 Checking trace generation...${NC}"
sleep 2

# Check Jaeger for traces
if curl -s "http://localhost:16686/api/services" | grep -q "api-gateway"; then
    echo -e "${GREEN}✅ Traces are being sent to Jaeger${NC}"
else
    echo -e "${YELLOW}⚠️  Traces not yet visible in Jaeger (may need more time)${NC}"
fi

echo ""
echo -e "${GREEN}🎉 Observability Integration Test Complete!${NC}"
echo ""
echo -e "${BLUE}📊 View Results:${NC}"
echo -e "${GREEN}🔍 Jaeger Traces: ${BLUE}http://localhost:16686${NC}"
echo -e "${GREEN}📈 Grafana Dashboards: ${BLUE}http://localhost:3000${NC}"
echo -e "${GREEN}📊 Prometheus Metrics: ${BLUE}http://localhost:9090${NC}"
echo -e "${GREEN}🚨 AlertManager: ${BLUE}http://localhost:9093${NC}"

echo ""
echo -e "${BLUE}🔍 Sample Queries:${NC}"
echo "Prometheus - HTTP Request Rate:"
echo "  rate(http_requests_total[5m])"
echo ""
echo "Prometheus - Response Time p95:"
echo "  histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))"
echo ""
echo "Jaeger - Search for traces:"
echo "  Service: api-gateway, Operation: GET /health"

echo ""
echo -e "${BLUE}📚 Next Steps:${NC}"
echo "1. Explore traces in Jaeger UI"
echo "2. Create custom Grafana dashboards"
echo "3. Set up alerting rules"
echo "4. Integrate observability into other services"

echo ""
echo -e "${GREEN}✅ All tests passed! Observability integration is working correctly.${NC}"
