# X-Form Backend Makefile
# Provides convenient commands for development and deployment

.PHONY: help setup build start stop test clean deploy

# Default target
help:
	@echo "X-Form Backend - Available Commands:"
	@echo ""
	@echo "Development:"
	@echo "  setup          - Initial project setup"
	@echo "  build          - Build all services"
	@echo "  start          - Start all services with Docker Compose"
	@echo "  stop           - Stop all services"
	@echo "  restart        - Restart all services"
	@echo "  logs           - Show logs from all services"
	@echo ""
	@echo "Individual Services:"
	@echo "  auth-dev       - Start auth service in development mode"
	@echo "  form-dev       - Start form service in development mode"
	@echo "  response-dev   - Start response service in development mode"
	@echo "  analytics-dev  - Start analytics service in development mode"
	@echo ""
	@echo "Testing:"
	@echo "  test           - Run all tests"
	@echo "  test-auth      - Run auth service tests"
	@echo "  test-form      - Run form service tests"
	@echo "  test-response  - Run response service tests"
	@echo ""
	@echo "Database:"
	@echo "  db-setup       - Initialize database"
	@echo "  db-migrate     - Run database migrations"
	@echo "  db-reset       - Reset database (WARNING: destroys data)"
	@echo ""
	@echo "Deployment:"
	@echo "  k8s-deploy     - Deploy to Kubernetes"
	@echo "  k8s-delete     - Delete Kubernetes deployment"
	@echo ""
	@echo "Utilities:"
	@echo "  clean          - Clean up containers and volumes"
	@echo "  format         - Format code in all services"
	@echo "  lint           - Lint code in all services"

# Setup
setup:
	@echo "🚀 Setting up X-Form Backend..."
	@chmod +x scripts/setup.sh
	@./scripts/setup.sh

# Build all services
build:
	@echo "🔨 Building all services..."
	docker-compose build

# Start all services
start:
	@echo "▶️  Starting all services..."
	docker-compose up -d
	@echo "✅ All services started!"
	@echo "🌐 API Gateway: http://localhost:8080"
	@echo "🔐 Auth Service: http://localhost:3001"
	@echo "📋 Form Service: http://localhost:8001"
	@echo "📝 Response Service: http://localhost:3002"
	@echo "📊 Analytics Service: http://localhost:5001"

# Stop all services
stop:
	@echo "⏹️  Stopping all services..."
	docker-compose down

# Restart services
restart: stop start

# Show logs
logs:
	docker-compose logs -f

# Individual service development
auth-dev:
	@echo "🔐 Starting Auth Service in development mode..."
	cd services/auth-service && npm run dev

form-dev:
	@echo "📋 Starting Form Service in development mode..."
	cd services/form-service && go run cmd/server/main.go

response-dev:
	@echo "📝 Starting Response Service in development mode..."
	cd services/response-service && npm run dev

analytics-dev:
	@echo "📊 Starting Analytics Service in development mode..."
	cd services/analytics-service && python main.py

# Testing
test:
	@echo "🧪 Running all tests..."
	@$(MAKE) test-auth
	@$(MAKE) test-form
	@$(MAKE) test-response

test-auth:
	@echo "🧪 Testing Auth Service..."
	cd services/auth-service && npm test

test-form:
	@echo "🧪 Testing Form Service..."
	cd services/form-service && go test ./...

test-response:
	@echo "🧪 Testing Response Service..."
	cd services/response-service && npm test

# Database operations
db-setup:
	@echo "🗄️  Setting up database..."
	docker-compose up -d postgres
	@sleep 5
	@echo "Database initialized!"

db-migrate:
	@echo "🗄️  Running database migrations..."
	cd services/form-service && go run cmd/migrate/main.go

db-reset:
	@echo "⚠️  Resetting database (THIS WILL DESTROY ALL DATA)..."
	@read -p "Are you sure? Type 'yes' to continue: " confirm && [ "$$confirm" = "yes" ]
	docker-compose down -v
	docker volume rm xform-backend_postgres_data || true
	docker-compose up -d postgres

# Kubernetes deployment
k8s-deploy:
	@echo "🚀 Deploying to Kubernetes..."
	kubectl apply -f deployment/k8s/infrastructure.yaml
	kubectl apply -f deployment/k8s/services.yaml
	@echo "✅ Deployed to Kubernetes!"
	@echo "Check status with: kubectl get pods -n xform"

k8s-delete:
	@echo "🗑️  Deleting Kubernetes deployment..."
	kubectl delete -f deployment/k8s/services.yaml
	kubectl delete -f deployment/k8s/infrastructure.yaml

# Utilities
clean:
	@echo "🧹 Cleaning up..."
	docker-compose down -v
	docker system prune -f
	docker volume prune -f

format:
	@echo "💅 Formatting code..."
	cd services/auth-service && npm run lint:fix
	cd services/response-service && npm run lint:fix
	cd services/form-service && go fmt ./...
	cd services/analytics-service && black . && isort .

lint:
	@echo "🔍 Linting code..."
	cd services/auth-service && npm run lint
	cd services/response-service && npm run lint
	cd services/form-service && golangci-lint run
	cd services/analytics-service && flake8 . && mypy .

# Development dependencies
install-deps:
	@echo "📦 Installing development dependencies..."
	# Auth Service
	cd services/auth-service && npm install
	# Response Service
	cd services/response-service && npm install
	# Form Service
	cd services/form-service && go mod tidy
	# Analytics Service
	cd services/analytics-service && pip install -r requirements.txt

# Environment setup
env-setup:
	@if [ ! -f .env ]; then \
		echo "📝 Creating .env file..."; \
		cp .env.example .env; \
		echo "⚠️  Please update .env with your configuration"; \
	else \
		echo "✅ .env file already exists"; \
	fi

# Generate API documentation
docs:
	@echo "📚 Generating API documentation..."
	# This would generate OpenAPI/Swagger docs for each service
	@echo "API documentation available at service endpoints /docs"

# Health check
health:
	@echo "🏥 Checking service health..."
	@curl -s http://localhost:8080/health | jq . || echo "API Gateway not responding"
	@curl -s http://localhost:3001/health | jq . || echo "Auth Service not responding"
	@curl -s http://localhost:8001/health | jq . || echo "Form Service not responding"
	@curl -s http://localhost:3002/health | jq . || echo "Response Service not responding"
	@curl -s http://localhost:5001/health | jq . || echo "Analytics Service not responding"
