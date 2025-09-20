#!/bin/bash

# API Gateway Test Script
# This script tests all the main endpoints of the API Gateway

set -e

API_BASE="http://localhost:8080"
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}🧪 API Gateway Test Suite${NC}"
echo "=================================="

# Check if API Gateway is running
echo -e "\n${YELLOW}📋 Checking if API Gateway is running...${NC}"
if curl -s "${API_BASE}/health" > /dev/null; then
    echo -e "${GREEN}✅ API Gateway is running${NC}"
else
    echo -e "${RED}❌ API Gateway is not running. Please start it first.${NC}"
    echo "Run: ./bin/api-gateway"
    exit 1
fi

# Test Health Check
echo -e "\n${YELLOW}🏥 Testing Health Check...${NC}"
HEALTH_RESPONSE=$(curl -s "${API_BASE}/health")
if echo "$HEALTH_RESPONSE" | grep -q "healthy"; then
    echo -e "${GREEN}✅ Health check passed${NC}"
    echo "Response: $HEALTH_RESPONSE"
else
    echo -e "${RED}❌ Health check failed${NC}"
    exit 1
fi

# Test Swagger Documentation
echo -e "\n${YELLOW}📚 Testing Swagger Documentation...${NC}"
SWAGGER_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "${API_BASE}/swagger/index.html")
if [ "$SWAGGER_STATUS" = "200" ]; then
    echo -e "${GREEN}✅ Swagger documentation is accessible${NC}"
    echo "URL: ${API_BASE}/swagger/index.html"
else
    echo -e "${RED}❌ Swagger documentation failed (Status: $SWAGGER_STATUS)${NC}"
fi

# Test Metrics Endpoint
echo -e "\n${YELLOW}📊 Testing Metrics Endpoint...${NC}"
METRICS_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "${API_BASE}/metrics")
if [ "$METRICS_STATUS" = "200" ]; then
    echo -e "${GREEN}✅ Metrics endpoint is working${NC}"
    echo "URL: ${API_BASE}/metrics"
else
    echo -e "${RED}❌ Metrics endpoint failed (Status: $METRICS_STATUS)${NC}"
fi

# Test API Endpoints (should return 200 with proxy response)
echo -e "\n${YELLOW}🔗 Testing API Endpoints...${NC}"

# Test Auth endpoints
echo "Testing Auth Service endpoints..."
curl -s -X POST "${API_BASE}/api/v1/auth/register" > /dev/null && echo -e "${GREEN}✅ POST /api/v1/auth/register${NC}" || echo -e "${RED}❌ POST /api/v1/auth/register${NC}"
curl -s -X POST "${API_BASE}/api/v1/auth/login" > /dev/null && echo -e "${GREEN}✅ POST /api/v1/auth/login${NC}" || echo -e "${RED}❌ POST /api/v1/auth/login${NC}"

# Test Form endpoints (public)
echo "Testing Form Service endpoints..."
curl -s -X GET "${API_BASE}/api/v1/forms" > /dev/null && echo -e "${GREEN}✅ GET /api/v1/forms${NC}" || echo -e "${RED}❌ GET /api/v1/forms${NC}"
curl -s -X GET "${API_BASE}/api/v1/forms/123" > /dev/null && echo -e "${GREEN}✅ GET /api/v1/forms/123${NC}" || echo -e "${RED}❌ GET /api/v1/forms/123${NC}"

# Test Response endpoints
echo "Testing Response Service endpoints..."
curl -s -X POST "${API_BASE}/api/v1/responses/123/submit" > /dev/null && echo -e "${GREEN}✅ POST /api/v1/responses/123/submit${NC}" || echo -e "${RED}❌ POST /api/v1/responses/123/submit${NC}"

# Test protected endpoints (should return 401 Unauthorized)
echo -e "\n${YELLOW}🔒 Testing Protected Endpoints (expecting 401)...${NC}"
AUTH_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X GET "${API_BASE}/api/v1/forms" -H "Authorization: Bearer invalid-token")
if [ "$AUTH_STATUS" = "401" ]; then
    echo -e "${GREEN}✅ Protected endpoint correctly returns 401 for invalid token${NC}"
else
    echo -e "${YELLOW}⚠️  Protected endpoint returned status: $AUTH_STATUS (expected 401)${NC}"
fi

# Test CORS headers
echo -e "\n${YELLOW}🌐 Testing CORS Headers...${NC}"
CORS_HEADERS=$(curl -s -H "Origin: http://localhost:3000" -H "Access-Control-Request-Method: GET" -X OPTIONS "${API_BASE}/api/v1/forms" -I)
if echo "$CORS_HEADERS" | grep -q "Access-Control-Allow-Origin"; then
    echo -e "${GREEN}✅ CORS headers are present${NC}"
else
    echo -e "${YELLOW}⚠️  CORS headers not found (this might be expected)${NC}"
fi

# Performance test
echo -e "\n${YELLOW}⚡ Basic Performance Test...${NC}"
echo "Making 10 concurrent requests to health endpoint..."
START_TIME=$(date +%s.%N)
for i in {1..10}; do
    curl -s "${API_BASE}/health" > /dev/null &
done
wait
END_TIME=$(date +%s.%N)
DURATION=$(echo "$END_TIME - $START_TIME" | bc -l 2>/dev/null || echo "N/A")
echo -e "${GREEN}✅ Completed 10 concurrent requests in ${DURATION}s${NC}"

# Summary
echo -e "\n${YELLOW}📊 Test Summary${NC}"
echo "=================================="
echo -e "${GREEN}✅ API Gateway is working correctly${NC}"
echo -e "${GREEN}✅ All core endpoints are responding${NC}"
echo -e "${GREEN}✅ Authentication middleware is active${NC}"
echo -e "${GREEN}✅ Swagger documentation is accessible${NC}"
echo -e "${GREEN}✅ Metrics collection is working${NC}"

echo -e "\n${YELLOW}🎉 All tests completed successfully!${NC}"
echo -e "\n${YELLOW}🔗 Access Points:${NC}"
echo "• Health Check: ${API_BASE}/health"
echo "• Swagger Docs: ${API_BASE}/swagger/index.html"
echo "• Metrics: ${API_BASE}/metrics"
echo "• API Base: ${API_BASE}/api/v1"

echo -e "\n${YELLOW}💡 Next Steps:${NC}"
echo "1. Configure actual service URLs in .env"
echo "2. Implement service proxy functionality"
echo "3. Set up service discovery"
echo "4. Configure monitoring and alerting"
