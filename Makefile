.PHONY: help check-versions install dev build test lint clean
.PHONY: docker-up docker-down docker-restart docker-reset docker-status docker-health
.PHONY: docker-logs docker-logs-postgres docker-logs-redis docker-logs-minio docker-shell-postgres

# Colors
BLUE=\033[0;34m
NC=\033[0m # No Color

help: ## Show this help message
	@echo "Solobueno ERP - Available Commands"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(BLUE)%-20s$(NC) %s\n", $$1, $$2}'

# =============================================================================
# Version Checks
# =============================================================================

check-versions: ## Verify required tool versions
	@echo "Checking required versions..."
	@command -v node >/dev/null 2>&1 || (echo "Error: Node.js not found. Install Node.js 20+" && exit 1)
	@command -v pnpm >/dev/null 2>&1 || (echo "Error: pnpm not found. Install pnpm 8+" && exit 1)
	@command -v go >/dev/null 2>&1 || (echo "Warning: Go not found. Install Go 1.22+ for backend development")
	@command -v docker >/dev/null 2>&1 || (echo "Warning: Docker not found. Install Docker for local services")
	@echo "✓ Version check complete"

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

DOCKER_COMPOSE := docker compose -f infrastructure/docker/docker-compose.yml

docker-up: ## Start Docker services (PostgreSQL, Redis, MinIO)
	@command -v docker >/dev/null 2>&1 || (echo "Error: Docker not found. Install Docker Desktop or Docker Engine." && exit 1)
	@docker info >/dev/null 2>&1 || (echo "Error: Docker is not running. Start Docker Desktop or the Docker daemon." && exit 1)
	@lsof -i :5432 >/dev/null 2>&1 && (echo "Error: Port 5432 already in use. Run: lsof -i :5432" && exit 1) || true
	@lsof -i :6379 >/dev/null 2>&1 && (echo "Error: Port 6379 already in use. Run: lsof -i :6379" && exit 1) || true
	@lsof -i :9000 >/dev/null 2>&1 && (echo "Error: Port 9000 already in use. Run: lsof -i :9000" && exit 1) || true
	@lsof -i :9001 >/dev/null 2>&1 && (echo "Error: Port 9001 already in use. Run: lsof -i :9001" && exit 1) || true
	$(DOCKER_COMPOSE) up -d
	@echo "Waiting for services to be healthy..."
	@sleep 5
	@$(MAKE) docker-status

docker-down: ## Stop Docker services (preserve data)
	$(DOCKER_COMPOSE) down

docker-restart: ## Restart Docker services
	$(DOCKER_COMPOSE) restart

docker-reset: ## Reset Docker services (removes all data!)
	@echo "⚠️  This will delete all local data. Press Ctrl+C to cancel..."
	@sleep 3
	$(DOCKER_COMPOSE) down -v
	$(DOCKER_COMPOSE) up -d
	@echo "✓ Services reset with fresh data."

docker-status: ## Show service status and health
	@echo ""
	@echo "Service Status:"
	@$(DOCKER_COMPOSE) ps
	@echo ""
	@echo "Health Check:"
	@bash infrastructure/scripts/health-check.sh || true

docker-health: ## Run health check script
	@bash infrastructure/scripts/health-check.sh

docker-logs: ## Tail logs from all services
	$(DOCKER_COMPOSE) logs -f

docker-logs-postgres: ## Tail PostgreSQL logs
	$(DOCKER_COMPOSE) logs -f postgres

docker-logs-redis: ## Tail Redis logs
	$(DOCKER_COMPOSE) logs -f redis

docker-logs-minio: ## Tail MinIO logs
	$(DOCKER_COMPOSE) logs -f minio

docker-shell-postgres: ## Open PostgreSQL shell
	docker exec -it solobueno-postgres psql -U solobueno -d solobueno_dev

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

setup: check-versions install docker-up ## Full development setup
	@echo ""
	@echo "✅ Setup complete!"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Copy infrastructure/config/dev.env.example to .env"
	@echo "  2. Run 'make dev' to start development servers"
	@echo ""
