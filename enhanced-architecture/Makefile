# Enhanced Architecture Makefile
# Provides development, testing, and deployment commands

.PHONY: help build test lint dev-start dev-stop clean docker-build docker-up docker-down setup dev-local test-local

# Default target
help: ## Show this help message
	@echo "Enhanced X-Form Backend Architecture"
	@echo ""
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Local Development (No Docker Required)
setup: ## Setup development environment
	@echo "Setting up development environment..."
	@chmod +x setup-dev-env.sh
	@./setup-dev-env.sh

dev-local: ## Start development server locally (no Docker)
	@echo "Starting local development server..."
	@chmod +x run-dev.sh
	@./run-dev.sh

test-local: ## Run tests locally (no Docker)
	@echo "Running tests locally..."
	@chmod +x test-all.sh
	@./test-all.sh

# Build commands
build: ## Build API Gateway
	@echo "Building API Gateway..."
	@cd api-gateway && go build -o bin/gateway ./cmd/server

test: ## Run all tests
	@echo "Running tests..."
	@cd api-gateway && go test -v -race ./...

lint: ## Run linter
	@echo "Running linter..."
	@cd api-gateway && \
	if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	elif test -f $(HOME)/go/bin/golangci-lint; then \
		$(HOME)/go/bin/golangci-lint run; \
	else \
		echo "⚠️  golangci-lint not found. Installing..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		$(HOME)/go/bin/golangci-lint run; \
	fi

# Docker Development commands
dev-start: ## Start development environment with Docker
	@echo "Starting enhanced architecture development environment..."
	@if command -v docker-compose >/dev/null 2>&1; then \
		docker-compose -f docker-compose.dev.yml up -d; \
		echo "Services starting up..."; \
		echo "API Gateway: http://localhost:8080"; \
		echo "Traefik Dashboard: http://localhost:8081"; \
		echo "Prometheus: http://localhost:9090"; \
	else \
		echo "❌ Docker not available. Use 'make dev-local' instead"; \
		exit 1; \
	fi

dev-stop: ## Stop development environment
	@echo "Stopping development environment..."
	@if command -v docker-compose >/dev/null 2>&1; then \
		docker-compose -f docker-compose.dev.yml down; \
	else \
		echo "❌ Docker not available"; \
	fi

# Code quality
fmt: ## Format code
	@echo "Formatting code..."
	@cd api-gateway && go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	@cd api-gateway && go vet ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@cd api-gateway && go test -v -race -coverprofile=coverage.out ./...
	@cd api-gateway && go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: api-gateway/coverage.html"

# Monitoring
health: ## Check health status
	@echo "Checking health status..."
	@curl -f http://localhost:8080/health || echo "❌ Health check failed"

metrics: ## Show current metrics
	@echo "Showing current metrics..."
	@curl -s http://localhost:8080/metrics | head -20

# Quick workflows
dev: build test lint ## Full development workflow
	@echo "✅ Development workflow completed"

ci: test lint ## CI workflow (test + lint)
	@echo "✅ CI workflow completed"

quick-start: setup build ## Complete setup for new developers
	@echo "✅ Quick start completed"
	@echo "📋 Next: Run 'make dev-local' to start the server"

# Cleanup
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf api-gateway/bin/
	@rm -rf api-gateway/logs/*.log
	@rm -rf api-gateway/coverage.out
	@rm -rf api-gateway/coverage.html
	@echo "✅ Cleanup completed"
