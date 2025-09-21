# X-Form Backend - Modern Microservices Platform
# Makefile for development, testing, and deployment automation

# Color codes for output
RED    := \033[31m
GREEN  := \033[32m
YELLOW := \033[33m
BLUE   := \033[34m
PURPLE := \033[35m
CYAN   := \033[36m
WHITE  := \033[37m
RESET  := \033[0m

# Project configuration
PROJECT_NAME := x-form-backend
DOCKER_COMPOSE_DEV := infrastructure/containers/docker-compose.yml
DOCKER_COMPOSE_PROD := infrastructure/containers/docker-compose-traefik.yml
DOCKER_COMPOSE_TEST := infrastructure/containers/docker-compose-v2.yml
DOCKER_COMPOSE_ENHANCED := infrastructure/containers/docker-compose.enhanced.yml
DOCKER_COMPOSE_ENHANCED_DEV := infrastructure/containers/docker-compose.enhanced.dev.yml

# Service directories
SERVICES := auth-service form-service response-service realtime-service analytics-service
NODE_SERVICES := auth-service response-service
GO_SERVICES := form-service realtime-service
PYTHON_SERVICES := analytics-service

# Default target
.DEFAULT_GOAL := help

# Help target
.PHONY: help
help: ## Display this help message
	@echo "$(CYAN)╔══════════════════════════════════════════════════════════════════════════════╗$(RESET)"
	@echo "$(CYAN)║                          X-Form Backend - Makefile                          ║$(RESET)"
	@echo "$(CYAN)╚══════════════════════════════════════════════════════════════════════════════╝$(RESET)"
	@echo ""
	@echo "$(GREEN)📚 Available commands:$(RESET)"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf ""} /^[a-zA-Z_-]+:.*?##/ { printf "  $(CYAN)%-20s$(RESET) %s\n", $$1, $$2 } /^##@/ { printf "\n$(YELLOW)%s$(RESET)\n", substr($$0, 5) }' $(MAKEFILE_LIST)
	@echo ""

##@ 🚀 Quick Start Commands
.PHONY: setup start dev stop clean

setup: ## Initial project setup (run this first)
	@echo "$(GREEN)🔧 Setting up X-Form Backend development environment...$(RESET)"
	@chmod +x tools/scripts/setup.sh
	@./tools/scripts/setup.sh
	@$(MAKE) install-deps
	@$(MAKE) setup-env
	@echo "$(GREEN)✅ Setup complete! Run 'make dev' to start development.$(RESET)"

verify: ## Verify development environment setup
	@echo "$(BLUE)🔍 Verifying development environment...$(RESET)"
	@chmod +x tools/scripts/verify-dev-environment.sh
	@./tools/scripts/verify-dev-environment.sh

start: ## Start all services in production mode
	@echo "$(GREEN)🚀 Starting all services in production mode...$(RESET)"
	@[ -f .env ] || (echo "$(RED)❌ .env file not found. Copy .env.example to .env first.$(RESET)" && exit 1)
	@docker compose --env-file .env -f $(DOCKER_COMPOSE_PROD) up -d
	@$(MAKE) wait-for-services
	@$(MAKE) health
	@echo "$(GREEN)✅ All services started successfully!$(RESET)"

dev: ## Start development environment with hot reload
	@echo "$(GREEN)🔥 Starting development environment...$(RESET)"
	@docker compose -f $(DOCKER_COMPOSE_DEV) up -d postgres redis
	@$(MAKE) dev-services
	@echo "$(GREEN)✅ Development environment ready!$(RESET)"

stop: ## Stop all services
	@echo "$(YELLOW)🛑 Stopping all services...$(RESET)"
	@docker compose -f $(DOCKER_COMPOSE_DEV) down
	@docker compose -f $(DOCKER_COMPOSE_PROD) down
	@echo "$(GREEN)✅ All services stopped.$(RESET)"

clean: ## Clean up containers, volumes, and cache
	@echo "$(YELLOW)🧹 Cleaning up containers and volumes...$(RESET)"
	@docker compose -f $(DOCKER_COMPOSE_DEV) down -v
	@docker compose -f $(DOCKER_COMPOSE_PROD) down -v
	@docker system prune -f
	@echo "$(GREEN)✅ Cleanup complete.$(RESET)"

##@ 📦 Dependencies and Installation
.PHONY: install-deps install-node-deps install-go-deps install-python-deps

install-deps: install-node-deps install-go-deps install-python-deps ## Install all dependencies

install-node-deps: ## Install Node.js dependencies
	@echo "$(BLUE)📦 Installing Node.js dependencies...$(RESET)"
	@npm install
	@for service in $(NODE_SERVICES); do \
		echo "$(BLUE)Installing deps for $$service...$(RESET)"; \
		cd apps/$$service && npm install && cd ../..; \
	done

install-go-deps: ## Install Go dependencies
	@echo "$(BLUE)📦 Installing Go dependencies...$(RESET)"
	@for service in $(GO_SERVICES); do \
		echo "$(BLUE)Installing deps for $$service...$(RESET)"; \
		cd apps/$$service && go mod download && go mod tidy && cd ../..; \
	done

install-python-deps: ## Install Python dependencies
	@echo "$(BLUE)📦 Installing Python dependencies...$(RESET)"
	@for service in $(PYTHON_SERVICES); do \
		echo "$(BLUE)Installing deps for $$service...$(RESET)"; \
		cd apps/$$service && pip install -r requirements.txt && cd ../..; \
	done

##@ 🏗️ Development Commands
.PHONY: dev-services build lint format test

dev-services: ## Start individual services in development mode
	@echo "$(GREEN)🔥 Starting development services...$(RESET)"
	@$(MAKE) -j $(words $(SERVICES)) $(addprefix dev-, $(SERVICES))

build: ## Build all services
	@echo "$(BLUE)🔨 Building all services...$(RESET)"
	@$(MAKE) -j $(words $(SERVICES)) $(addprefix build-, $(SERVICES))

lint: ## Run linting on all services
	@echo "$(BLUE)🔍 Running linting...$(RESET)"
	@$(MAKE) -j $(words $(SERVICES)) $(addprefix lint-, $(SERVICES))

format: ## Format code in all services
	@echo "$(BLUE)🎨 Formatting code...$(RESET)"
	@$(MAKE) -j $(words $(SERVICES)) $(addprefix format-, $(SERVICES))

test: ## Run tests for all services
	@echo "$(BLUE)🧪 Running tests...$(RESET)"
	@$(MAKE) test-unit
	@$(MAKE) test-integration

##@ 🧪 Testing Commands
.PHONY: test-unit test-integration test-e2e test-api test-load

test-unit: ## Run unit tests
	@echo "$(BLUE)🧪 Running unit tests...$(RESET)"
	@$(MAKE) -j $(words $(SERVICES)) $(addprefix test-unit-, $(SERVICES))

test-integration: ## Run integration tests
	@echo "$(BLUE)🔗 Running integration tests...$(RESET)"
	@docker compose -f $(DOCKER_COMPOSE_TEST) up -d
	@sleep 10
	@$(MAKE) -j $(words $(SERVICES)) $(addprefix test-integration-, $(SERVICES))
	@docker compose -f $(DOCKER_COMPOSE_TEST) down

test-e2e: ## Run end-to-end tests
	@echo "$(BLUE)🎭 Running E2E tests...$(RESET)"
	@cd tests/e2e && npm test

test-api: ## Test API endpoints
	@echo "$(BLUE)📡 Testing API endpoints...$(RESET)"
	@./tools/scripts/test-api.sh

test-load: ## Run load tests
	@echo "$(BLUE)⚡ Running load tests...$(RESET)"
	@k6 run tests/performance/load-test.js

##@ 🗄️ Database Commands
.PHONY: db-setup db-migrate db-seed db-reset db-backup db-restore

db-setup: ## Setup databases
	@echo "$(BLUE)🗄️ Setting up databases...$(RESET)"
	@docker compose -f $(DOCKER_COMPOSE_DEV) up -d postgres redis
	@sleep 5
	@$(MAKE) db-migrate
	@$(MAKE) db-seed

db-migrate: ## Run database migrations
	@echo "$(BLUE)📊 Running database migrations...$(RESET)"
	@cd migrations && ./migrate.sh up

db-seed: ## Seed database with test data
	@echo "$(BLUE)🌱 Seeding database...$(RESET)"
	@cd migrations && ./seed.sh

db-reset: ## Reset databases
	@echo "$(YELLOW)🔄 Resetting databases...$(RESET)"
	@docker compose -f $(DOCKER_COMPOSE_DEV) down -v postgres redis
	@docker compose -f $(DOCKER_COMPOSE_DEV) up -d postgres redis
	@sleep 5
	@$(MAKE) db-migrate
	@$(MAKE) db-seed

db-backup: ## Backup database
	@echo "$(BLUE)💾 Creating database backup...$(RESET)"
	@./tools/scripts/backup-db.sh

db-restore: ## Restore database from backup
	@echo "$(BLUE)📥 Restoring database...$(RESET)"
	@./tools/scripts/restore-db.sh $(BACKUP_FILE)

##@ 📊 Monitoring and Health
.PHONY: health logs monitoring metrics

health: ## Check service health
	@echo "$(BLUE)🏥 Checking service health...$(RESET)"
	@./tools/scripts/health-check.sh

logs: ## View all service logs
	@echo "$(BLUE)📋 Showing service logs...$(RESET)"
	@docker compose -f $(DOCKER_COMPOSE_DEV) logs -f

monitoring: ## Start monitoring stack
	@echo "$(BLUE)📊 Starting monitoring stack...$(RESET)"
	@docker compose -f infrastructure/monitoring/docker-compose.monitoring.yml up -d
	@echo "$(GREEN)✅ Monitoring available at:$(RESET)"
	@echo "  Grafana: http://grafana.localhost:3000"
	@echo "  Prometheus: http://prometheus.localhost:9091"

metrics: ## View metrics
	@echo "$(BLUE)📈 Fetching metrics...$(RESET)"
	@curl -s http://localhost:9090/metrics | head -20

##@ 🔧 Utility Commands
.PHONY: setup-env generate-secrets security-scan docs

setup-env: ## Setup environment variables
	@echo "$(BLUE)⚙️ Setting up environment variables...$(RESET)"
	@cp configs/environments/.env.example .env
	@echo "$(YELLOW)⚠️ Please update .env file with your configuration$(RESET)"

generate-secrets: ## Generate JWT and other secrets
	@echo "$(BLUE)🔐 Generating secrets...$(RESET)"
	@./tools/scripts/generate-secrets.sh

security-scan: ## Run security scans
	@echo "$(BLUE)🔒 Running security scans...$(RESET)"
	@npm audit
	@for service in $(GO_SERVICES); do \
		cd apps/$$service && gosec ./... && cd ../..; \
	done

docs: ## Generate documentation
	@echo "$(BLUE)📚 Generating documentation...$(RESET)"
	@./tools/scripts/generate-docs.sh
	@echo "$(GREEN)📖 Documentation available at docs/$(RESET)"

##@ 🚀 Deployment Commands
.PHONY: deploy-dev deploy-staging deploy-prod

deploy-dev: ## Deploy to development environment
	@echo "$(GREEN)🚀 Deploying to development...$(RESET)"
	@./tools/scripts/deploy.sh dev

deploy-staging: ## Deploy to staging environment
	@echo "$(GREEN)🚀 Deploying to staging...$(RESET)"
	@./tools/scripts/deploy.sh staging

deploy-prod: ## Deploy to production environment
	@echo "$(GREEN)🚀 Deploying to production...$(RESET)"
	@./tools/scripts/deploy.sh prod

##@ 📋 Information Commands
.PHONY: info status ports

info: ## Show project information
	@echo "$(CYAN)╔══════════════════════════════════════════════════════════════════════════════╗$(RESET)"
	@echo "$(CYAN)║                          X-Form Backend Project                             ║$(RESET)"
	@echo "$(CYAN)╚══════════════════════════════════════════════════════════════════════════════╝$(RESET)"
	@echo ""
	@echo "$(GREEN)📊 Project Status:$(RESET)"
	@echo "  Name: $(PROJECT_NAME)"
	@echo "  Services: $(words $(SERVICES))"
	@echo "  Node.js Services: $(NODE_SERVICES)"
	@echo "  Go Services: $(GO_SERVICES)"
	@echo "  Python Services: $(PYTHON_SERVICES)"
	@echo ""
	@echo "$(GREEN)🌐 Service URLs (when running):$(RESET)"
	@echo "  Main API: http://api.localhost"
	@echo "  Traefik Dashboard: http://traefik.localhost:8080"
	@echo "  Swagger UI: http://api.localhost/docs"
	@echo "  Grafana: http://grafana.localhost:3000"
	@echo "  Prometheus: http://prometheus.localhost:9091"

status: ## Show running services status
	@echo "$(BLUE)📊 Service Status:$(RESET)"
	@docker compose -f $(DOCKER_COMPOSE_DEV) ps 2>/dev/null || echo "No services running"

ports: ## Show port usage
	@echo "$(BLUE)🔌 Port Usage:$(RESET)"
	@echo "  3001 - Auth Service"
	@echo "  3002 - Response Service"
	@echo "  8001 - Form Service"
	@echo "  8002 - Realtime Service"
	@echo "  5001 - Analytics Service"
	@echo "  5432 - PostgreSQL"
	@echo "  6379 - Redis"
	@echo "  8080 - Traefik Dashboard"
	@echo "  3000 - Grafana"
	@echo "  9091 - Prometheus"

##@ 🏗️ Service-specific Build Targets
.PHONY: $(addprefix build-, $(SERVICES)) $(addprefix dev-, $(SERVICES)) $(addprefix lint-, $(SERVICES))

# Build targets for each service
build-auth-service:
	@echo "$(GREEN)🔨 Building auth-service...$(RESET)"
	@cd apps/auth-service && npm run build

build-response-service:
	@echo "$(GREEN)🔨 Building response-service...$(RESET)"
	@cd apps/response-service && npm run build

build-form-service:
	@echo "$(GREEN)🔨 Building form-service...$(RESET)"
	@cd apps/form-service && go build -o bin/form-service ./cmd/server

build-realtime-service:
	@echo "$(GREEN)🔨 Building realtime-service...$(RESET)"
	@cd apps/realtime-service && go build -o bin/realtime-service ./cmd/server

build-analytics-service:
	@echo "$(GREEN)🔨 Building analytics-service...$(RESET)"
	@cd apps/analytics-service && echo "Python service ready"

# Development targets for each service
dev-auth-service:
	@echo "$(YELLOW)🔥 Starting auth-service in development mode...$(RESET)"
	@cd apps/auth-service && npm run dev

dev-response-service:
	@echo "$(YELLOW)🔥 Starting response-service in development mode...$(RESET)"
	@cd apps/response-service && npm run dev

dev-form-service:
	@echo "$(YELLOW)🔥 Starting form-service in development mode...$(RESET)"
	@cd apps/form-service && go run cmd/server/main.go

dev-realtime-service:
	@echo "$(YELLOW)🔥 Starting realtime-service in development mode...$(RESET)"
	@cd apps/realtime-service && go run cmd/server/main.go

dev-analytics-service:
	@echo "$(YELLOW)🔥 Starting analytics-service in development mode...$(RESET)"
	@cd apps/analytics-service && python main.py

# Lint targets for each service
lint-auth-service:
	@echo "$(BLUE)🔍 Linting auth-service...$(RESET)"
	@cd apps/auth-service && npm run lint

lint-response-service:
	@echo "$(BLUE)🔍 Linting response-service...$(RESET)"
	@cd apps/response-service && npm run lint

lint-form-service:
	@echo "$(BLUE)🔍 Linting form-service...$(RESET)"
	@cd apps/form-service && golangci-lint run

lint-realtime-service:
	@echo "$(BLUE)🔍 Linting realtime-service...$(RESET)"
	@cd apps/realtime-service && golangci-lint run

lint-analytics-service:
	@echo "$(BLUE)🔍 Linting analytics-service...$(RESET)"
	@cd apps/analytics-service && flake8 .

##@ 🚀 Enhanced Architecture Commands
.PHONY: enhanced-start enhanced-dev enhanced-stop enhanced-logs enhanced-status

enhanced-start: ## Start enhanced architecture with production-ready API Gateway
	@echo "$(GREEN)🚀 Starting enhanced architecture in production mode...$(RESET)"
	@docker compose -f $(DOCKER_COMPOSE_ENHANCED) up -d
	@$(MAKE) wait-for-services
	@echo "$(GREEN)✅ Enhanced architecture started successfully!$(RESET)"
	@echo "$(CYAN)🌐 API Gateway: http://api.localhost$(RESET)"
	@echo "$(CYAN)📊 Traefik Dashboard: http://traefik.localhost:8080$(RESET)"

enhanced-dev: ## Start enhanced architecture in development mode
	@echo "$(YELLOW)🔥 Starting enhanced architecture in development mode...$(RESET)"
	@docker compose -f $(DOCKER_COMPOSE_ENHANCED_DEV) up -d
	@$(MAKE) wait-for-services
	@echo "$(GREEN)✅ Enhanced development environment ready!$(RESET)"

enhanced-stop: ## Stop enhanced architecture
	@echo "$(RED)🛑 Stopping enhanced architecture...$(RESET)"
	@docker compose -f $(DOCKER_COMPOSE_ENHANCED) down
	@docker compose -f $(DOCKER_COMPOSE_ENHANCED_DEV) down

enhanced-logs: ## View enhanced architecture logs
	@echo "$(BLUE)📋 Enhanced architecture logs:$(RESET)"
	@docker compose -f $(DOCKER_COMPOSE_ENHANCED) logs -f

enhanced-status: ## Show enhanced architecture status
	@echo "$(BLUE)📊 Enhanced Architecture Status:$(RESET)"
	@docker compose -f $(DOCKER_COMPOSE_ENHANCED) ps 2>/dev/null || echo "Enhanced services not running"
