#!/bin/bash

# Health Check Script for X-Form Backend Services
# Checks the health of all running services

set -e

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Service endpoints
SERVICES=(
    "auth-service:http://localhost:3001/health"
    "form-service:http://localhost:8001/health"
    "response-service:http://localhost:3002/health"
    "realtime-service:http://localhost:8002/health"
    "analytics-service:http://localhost:5001/health"
)

# Infrastructure services
INFRASTRUCTURE=(
    "PostgreSQL:postgresql://postgres:password@localhost:5432/xform_dev"
    "Redis:redis://localhost:6379"
)

echo -e "${BLUE}🏥 X-Form Backend Health Check${NC}"
echo -e "${BLUE}═══════════════════════════════${NC}"
echo ""

# Check application services
echo -e "${YELLOW}📊 Application Services:${NC}"
healthy_services=0
total_services=${#SERVICES[@]}

for service_info in "${SERVICES[@]}"; do
    IFS=':' read -r service_name service_url <<< "$service_info"
    
    if curl -s --max-time 5 "$service_url" >/dev/null 2>&1; then
        echo -e "  ${GREEN}✓${NC} $service_name"
        ((healthy_services++))
    else
        echo -e "  ${RED}✗${NC} $service_name (not responding)"
    fi
done

echo ""

# Check infrastructure services
echo -e "${YELLOW}🗄️ Infrastructure Services:${NC}"
healthy_infra=0
total_infra=${#INFRASTRUCTURE[@]}

# Check PostgreSQL
if pg_isready -h localhost -p 5432 -U postgres >/dev/null 2>&1; then
    echo -e "  ${GREEN}✓${NC} PostgreSQL"
    ((healthy_infra++))
else
    echo -e "  ${RED}✗${NC} PostgreSQL (not responding)"
fi

# Check Redis
if redis-cli -h localhost -p 6379 ping >/dev/null 2>&1; then
    echo -e "  ${GREEN}✓${NC} Redis"
    ((healthy_infra++))
else
    echo -e "  ${RED}✗${NC} Redis (not responding)"
fi

echo ""

# Summary
echo -e "${YELLOW}📈 Health Summary:${NC}"
echo -e "  Application Services: ${healthy_services}/${total_services} healthy"
echo -e "  Infrastructure: ${healthy_infra}/${total_infra} healthy"

total_healthy=$((healthy_services + healthy_infra))
total_services=$((total_services + total_infra))

if [ $total_healthy -eq $total_services ]; then
    echo -e "  ${GREEN}✅ All services healthy!${NC}"
    exit 0
elif [ $total_healthy -gt 0 ]; then
    echo -e "  ${YELLOW}⚠️ Some services unhealthy${NC}"
    exit 1
else
    echo -e "  ${RED}❌ All services down${NC}"
    exit 2
fi
