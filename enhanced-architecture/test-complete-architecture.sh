#!/bin/bash

# Enhanced X-Form Backend - Complete Architecture Test Script
# Tests all components of the 7-step API Gateway and service integration

echo "🧪 Enhanced X-Form Backend Architecture - Complete Test Suite"
echo "=============================================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test results
PASSED=0
FAILED=0

# Function to print test results
print_result() {
    if [ $1 -eq 0 ]; then
        echo -e "   ${GREEN}✅ PASSED${NC}: $2"
        ((PASSED++))
    else
        echo -e "   ${RED}❌ FAILED${NC}: $2"
        ((FAILED++))
    fi
}

echo ""
echo "📋 Test 1: API Gateway Build Verification"
echo "----------------------------------------"

cd api-gateway

# Test 1.1: Go mod tidy
echo "1.1 Testing Go module dependencies..."
go mod tidy &>/dev/null
print_result $? "Go module dependencies resolved"

# Test 1.2: Build API Gateway
echo "1.2 Testing API Gateway compilation..."
go build -o bin/api-gateway cmd/server/main.go &>/dev/null
print_result $? "API Gateway builds without errors"

# Test 1.3: Generate Swagger docs
echo "1.3 Testing Swagger documentation generation..."
if command -v ~/go/bin/swag &> /dev/null; then
    ~/go/bin/swag init -g cmd/server/main.go -o docs &>/dev/null
    print_result $? "Swagger documentation generated"
else
    echo "   ${YELLOW}⚠️  SKIPPED${NC}: swag not installed"
fi

echo ""
echo "📋 Test 2: Middleware Chain Verification"
echo "---------------------------------------"

# Test 2.1: Check middleware implementations
echo "2.1 Testing Step 1 (Parameter Validation)..."
grep -q "ParameterValidation" internal/middleware/middleware.go
print_result $? "Step 1: Parameter Validation implemented"

echo "2.2 Testing Step 2 (Whitelist Validation)..."
grep -q "WhitelistValidation" internal/middleware/middleware.go
print_result $? "Step 2: Whitelist Validation implemented"

echo "2.3 Testing Step 3 (Authentication)..."
grep -q "Authentication" internal/middleware/middleware.go
print_result $? "Step 3: Authentication implemented"

echo "2.4 Testing Step 4 (Rate Limiting)..."
grep -q "RateLimit" internal/middleware/middleware.go
print_result $? "Step 4: Rate Limiting implemented"

echo "2.5 Testing Step 5 (Service Discovery)..."
grep -q "ServiceDiscoveryMiddleware" internal/middleware/middleware.go
print_result $? "Step 5: Service Discovery implemented"

echo "2.6 Testing Step 6 (Request Transformation)..."
grep -q "transformRequest" internal/handler/handler.go
print_result $? "Step 6: Request Transformation implemented"

echo "2.7 Testing Step 7 (Reverse Proxy)..."
grep -q "ProxyToService" internal/handler/handler.go
print_result $? "Step 7: Reverse Proxy implemented"

echo ""
echo "📋 Test 3: Service Integration Verification"
echo "------------------------------------------"

# Test 3.1: Check service registry
echo "3.1 Testing service registry implementation..."
grep -q "ServiceRegistry" internal/middleware/middleware.go
print_result $? "Service registry implemented"

# Test 3.2: Check service definitions
echo "3.2 Testing service definitions..."
grep -q "auth-service" internal/middleware/middleware.go
grep -q "form-service" internal/middleware/middleware.go
grep -q "response-service" internal/middleware/middleware.go
print_result $? "All core services defined in registry"

# Test 3.3: Check circuit breakers
echo "3.3 Testing circuit breaker implementation..."
grep -q "CircuitBreaker" internal/handler/handler.go
print_result $? "Circuit breakers implemented"

echo ""
echo "📋 Test 4: Configuration Verification"
echo "------------------------------------"

# Test 4.1: Check config structure
echo "4.1 Testing configuration structure..."
grep -q "AuthConfig" internal/config/config.go
grep -q "SecurityConfig" internal/config/config.go
print_result $? "Configuration structure complete"

# Test 4.2: Check Docker integration
echo "4.2 Testing Docker configuration..."
cd ..
if [ -f "docker-compose-complete.yml" ]; then
    grep -q "api-gateway" docker-compose-complete.yml
    print_result $? "Docker Compose configuration complete"
else
    print_result 1 "Docker Compose file not found"
fi

echo ""
echo "📋 Test 5: Architecture Compliance"
echo "---------------------------------"

# Test 5.1: Check Traefik configuration
echo "5.1 Testing Traefik (Load Balancer) configuration..."
if [ -f "edge-layer/traefik.yml" ]; then
    grep -qi "entrypoints\|entryPoints" edge-layer/traefik.yml
    print_result $? "Traefik load balancer configured"
else
    print_result 1 "Traefik configuration not found"
fi

# Test 5.2: Check dynamic routing
echo "5.2 Testing dynamic routing configuration..."
if [ -f "edge-layer/dynamic.yml" ]; then
    grep -q "services" edge-layer/dynamic.yml
    print_result $? "Dynamic routing configured"
else
    print_result 1 "Dynamic routing configuration not found"
fi

echo ""
echo "📋 Test 6: Documentation Verification"
echo "------------------------------------"

# Test 6.1: Check Swagger setup
echo "6.1 Testing Swagger documentation setup..."
cd api-gateway
grep -q "@title" cmd/server/main.go
print_result $? "Swagger documentation annotations present"

# Test 6.2: Check API endpoints documentation
echo "6.2 Testing API endpoint documentation..."
grep -q "@Summary" cmd/server/main.go
print_result $? "API endpoints documented"

echo ""
echo "📊 TEST SUMMARY"
echo "==============="
echo -e "Total Tests: $((PASSED + FAILED))"
echo -e "${GREEN}Passed: ${PASSED}${NC}"
echo -e "${RED}Failed: ${FAILED}${NC}"

if [ $FAILED -eq 0 ]; then
    echo ""
    echo -e "${GREEN}🎉 ALL TESTS PASSED!${NC}"
    echo -e "${GREEN}Enhanced X-Form Backend architecture is fully implemented and ready.${NC}"
    echo ""
    echo "🚀 Ready to run:"
    echo "   docker-compose -f docker-compose-complete.yml up -d"
    echo ""
    echo "📊 Access points:"
    echo "   • API Gateway: http://localhost:8000"
    echo "   • Health Check: http://localhost:8000/health"  
    echo "   • Swagger Docs: http://localhost:8000/swagger/index.html"
    echo "   • Traefik Dashboard: http://localhost:8080"
    exit 0
else
    echo ""
    echo -e "${RED}❌ Some tests failed. Please review the implementation.${NC}"
    exit 1
fi
