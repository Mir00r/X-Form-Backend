# X-Form Backend Makefile
# Traefik All-in-One Architecture Commands

.PHONY: help setup build start stop test clean deploy

# Default target
help:
	@echo "🚀 X-Form Backend - Traefik All-in-One Architecture"
	@echo ""
	@echo "Quick Start:"
	@echo "  setup          - Initial project setup"
	@echo "  start          - Start Traefik + all services"
	@echo "  health         - Check system health"
	@echo "  stop           - Stop all services"
	@echo ""
	@echo "Architecture Management:"
	@echo "  traefik-only   - Start only Traefik (for development)"
	@echo "  traefik-config - Validate Traefik configuration"
	@echo "  traefik-logs   - Show Traefik logs"
	@echo "  traefik-dash   - Open Traefik dashboard"
	@echo ""
	@echo "Development:"
	@echo "  build          - Build all services"
	@echo "  restart        - Restart all services"
	@echo "  logs           - Show logs from all services"
	@echo ""
	@echo "Testing & Monitoring:"
	@echo "  test           - Run all tests"
	@echo "  load-test      - Run load tests against Traefik"
	@echo "  monitor        - Open monitoring dashboards"
	@echo "  api-test       - Test API endpoints through Traefik"
	@echo ""
	@echo "Utilities:"
	@echo "  clean          - Clean up containers and volumes"
	@echo "  arch-info      - Show architecture information"

# Setup
setup:
	@echo "🚀 Setting up X-Form Backend with Traefik..."
	@chmod +x scripts/setup.sh
	@./scripts/setup.sh

# Build all services
build:
	@echo "🔨 Building all services..."
	@docker-compose -f docker-compose-traefik.yml build

# Start Traefik All-in-One stack
start:
	@echo "🚀 Starting X-Form Backend with Traefik All-in-One..."
	@docker-compose -f docker-compose-traefik.yml up -d
	@echo ""
	@echo "✅ Traefik All-in-One stack started!"
	@echo ""
	@echo "🌐 Access Points:"
	@echo "   📡 Main API:           http://api.localhost"
	@echo "   🔌 WebSocket:          ws://ws.localhost"  
	@echo "   📊 Traefik Dashboard:  http://traefik.localhost:8080"
	@echo "   📈 Grafana:            http://grafana.localhost:3000"
	@echo "   🔍 Prometheus:         http://prometheus.localhost:9091"
	@echo "   🔎 Jaeger:             http://jaeger.localhost:16686"
	@echo ""
	@echo "💡 Next steps:"
	@echo "   • Run 'make health' to check service health"
	@echo "   • Run 'make api-test' to test API endpoints"
	@echo "   • Run 'make monitor' to open all dashboards"

# Stop all services
stop:
	@echo "⏹️  Stopping Traefik stack..."
	@docker-compose -f docker-compose-traefik.yml down

# Restart services
restart: stop start

# Show logs
logs:
	@docker-compose -f docker-compose-traefik.yml logs -f

# Start only Traefik (for development)
traefik-only:
	@echo "🚀 Starting Traefik only..."
	@docker-compose -f docker-compose-traefik.yml up -d traefik
	@echo "✅ Traefik started: http://traefik.localhost:8080"

# Show Traefik logs
traefik-logs:
	@echo "📋 Traefik logs:"
	@docker-compose -f docker-compose-traefik.yml logs -f traefik

# Validate Traefik configuration
traefik-config:
	@echo "🔍 Validating Traefik configuration..."
	@docker run --rm -v $(PWD)/infrastructure/traefik:/config traefik:v3.0 traefik --configfile=/config/traefik.yml --dry-run

# Open Traefik dashboard
traefik-dash:
	@echo "🌐 Opening Traefik dashboard..."
	@open http://traefik.localhost:8080 || echo "Please open http://traefik.localhost:8080 in your browser"

# Health check
health:
	@echo "🏥 Checking system health..."
	@echo ""
	@echo "Traefik Health:"
	@curl -s http://traefik.localhost:8080/ping && echo "✅ Traefik: OK" || echo "❌ Traefik: FAILED"
	@echo ""
	@echo "Individual Services:"
	@docker-compose -f docker-compose-traefik.yml ps

# API endpoint testing
api-test:
	@echo "🧪 Testing API endpoints through Traefik..."
	@echo ""
	@echo "Testing Auth endpoints:"
	@curl -s -o /dev/null -w "Status: %{http_code}\n" http://api.localhost/api/v1/auth/health || echo "❌ Auth service unreachable"

# Load testing
load-test:
	@echo "⚡ Running load tests..."
	@command -v hey >/dev/null 2>&1 || { echo >&2 "❌ 'hey' required but not installed. Install with: go install github.com/rakyll/hey@latest"; exit 1; }
	@echo "Testing API performance:"
	@hey -n 1000 -c 50 -t 30 http://api.localhost/health

# Open monitoring dashboards
monitor:
	@echo "📊 Opening monitoring dashboards..."
	@open http://traefik.localhost:8080 || echo "Traefik Dashboard: http://traefik.localhost:8080"
	@open http://grafana.localhost:3000 || echo "Grafana: http://grafana.localhost:3000 (admin/admin)"
	@open http://prometheus.localhost:9091 || echo "Prometheus: http://prometheus.localhost:9091"
	@open http://jaeger.localhost:16686 || echo "Jaeger: http://jaeger.localhost:16686"

# Architecture information
arch-info:
	@echo "🏗️  X-Form Backend Architecture Information"
	@echo ""
	@echo "📋 Current Architecture: Traefik All-in-One"
	@echo "   ├── Ingress Controller: Traefik (ports 80, 443, 8080)"
	@echo "   ├── API Gateway: Traefik Middlewares (JWT, CORS, Routing)"
	@echo "   ├── API Management: Traefik Plugins (Rate Limiting, Analytics)"
	@echo "   └── Load Balancer: Traefik LoadBalancer"
	@echo ""
	@echo "🔗 Traffic Flow:"
	@echo "   Internet → Traefik (Ingress) → Traefik (Gateway) → Traefik (Management) → Microservices"
	@echo ""
	@echo "🚀 Services:"
	@echo "   ├── Auth Service (Node.js): JWT, User management"
	@echo "   ├── Form Service (Go): Form CRUD operations"
	@echo "   ├── Response Service (Node.js): Form submissions"
	@echo "   ├── Real-time Service (Go): WebSocket connections"
	@echo "   ├── Analytics Service (Python): Data analytics"
	@echo "   └── File Service (NGINX): File uploads/downloads"

# Testing
test:
	@echo "🧪 Running all tests..."
	@echo "Tests will be implemented as services are completed"

# Utilities
clean:
	@echo "🧹 Cleaning up..."
	@docker-compose -f docker-compose-traefik.yml down -v
	@docker system prune -f
	@docker volume prune -f
	@echo "✅ Cleanup completed"
