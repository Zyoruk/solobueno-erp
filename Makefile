.PHONY: help install dev build test lint clean docker-up docker-down docker-logs docker-reset

# Colors
BLUE=\033[0;34m
NC=\033[0m # No Color

help: ## Show this help message
	@echo "Solobueno ERP - Available Commands"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(BLUE)%-20s$(NC) %s\n", $$1, $$2}'

# =============================================================================
# Development
# =============================================================================

install: ## Install all dependencies
	pnpm install

dev: ## Start development servers
	pnpm dev

build: ## Build all packages and apps
	pnpm build

test: ## Run all tests
	pnpm test

lint: ## Run linting
	pnpm lint

format: ## Format code
	pnpm format

clean: ## Clean all build artifacts
	pnpm clean
	cd backend && go clean -cache

# =============================================================================
# Docker Services
# =============================================================================

docker-up: ## Start Docker services (PostgreSQL, Redis, MinIO)
	docker compose -f infrastructure/docker/docker-compose.yml up -d
	@echo "Waiting for services to be healthy..."
	@sleep 5
	@docker compose -f infrastructure/docker/docker-compose.yml ps

docker-down: ## Stop Docker services
	docker compose -f infrastructure/docker/docker-compose.yml down

docker-logs: ## View Docker service logs
	docker compose -f infrastructure/docker/docker-compose.yml logs -f

docker-reset: ## Reset Docker services (removes all data!)
	docker compose -f infrastructure/docker/docker-compose.yml down -v
	docker compose -f infrastructure/docker/docker-compose.yml up -d

# =============================================================================
# Backend
# =============================================================================

backend-build: ## Build Go backend
	cd backend && go build -v ./...

backend-test: ## Test Go backend
	cd backend && go test -v ./...

backend-run: ## Run Go backend server
	cd backend && go run ./cmd/server

# =============================================================================
# Database
# =============================================================================

migrate-up: ## Run database migrations
	cd backend && go run ./cmd/migrate up

migrate-down: ## Rollback last migration
	cd backend && go run ./cmd/migrate down

migrate-status: ## Show migration status
	cd backend && go run ./cmd/migrate status

# =============================================================================
# Setup
# =============================================================================

setup: install docker-up ## Full development setup
	@echo ""
	@echo "✅ Setup complete!"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Copy infrastructure/config/dev.env.example to .env"
	@echo "  2. Run 'make dev' to start development servers"
	@echo ""
