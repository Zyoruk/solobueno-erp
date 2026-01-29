# Data Model: Docker Local Development Environment

**Feature**: 002-docker-local-dev  
**Date**: 2025-01-29

## Overview

This feature does not introduce database entities. Instead, it defines the Docker Compose configuration and supporting scripts for local development services.

## Docker Compose Configuration

### Enhanced docker-compose.yml

**File**: `infrastructure/docker/docker-compose.yml`

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:16-alpine
    container_name: solobueno-postgres
    restart: unless-stopped
    ports:
      - '5432:5432'
    environment:
      POSTGRES_USER: solobueno
      POSTGRES_PASSWORD: solobueno_dev
      POSTGRES_DB: solobueno_dev
    volumes:
      - solobueno_postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ['CMD-SHELL', 'pg_isready -U solobueno -d solobueno_dev']
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 10s
    networks:
      - solobueno-network

  redis:
    image: redis:7-alpine
    container_name: solobueno-redis
    restart: unless-stopped
    ports:
      - '6379:6379'
    volumes:
      - solobueno_redis_data:/data
    command: redis-server --appendonly yes
    healthcheck:
      test: ['CMD', 'redis-cli', 'ping']
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 5s
    networks:
      - solobueno-network

  minio:
    image: minio/minio:latest
    container_name: solobueno-minio
    restart: unless-stopped
    ports:
      - '9000:9000'
      - '9001:9001'
    environment:
      MINIO_ROOT_USER: solobueno
      MINIO_ROOT_PASSWORD: solobueno_dev
    volumes:
      - solobueno_minio_data:/data
    command: server /data --console-address ":9001"
    healthcheck:
      test: ['CMD', 'curl', '-f', 'http://localhost:9000/minio/health/live']
      interval: 30s
      timeout: 20s
      retries: 3
      start_period: 10s
    networks:
      - solobueno-network

networks:
  solobueno-network:
    driver: bridge
    name: solobueno-network

volumes:
  solobueno_postgres_data:
    name: solobueno_postgres_data
  solobueno_redis_data:
    name: solobueno_redis_data
  solobueno_minio_data:
    name: solobueno_minio_data
```

### Key Enhancements from 001-init-monorepo

| Enhancement                     | Purpose                               |
| ------------------------------- | ------------------------------------- |
| Named network                   | Service discovery by container name   |
| Named volumes                   | Explicit volume naming for management |
| `start_period` in health checks | Grace period for slow containers      |
| Network isolation               | Services communicate internally       |

## Health Check Script

**File**: `infrastructure/scripts/health-check.sh`

```bash
#!/usr/bin/env bash
set -e

COMPOSE_FILE="infrastructure/docker/docker-compose.yml"

echo "Checking service health..."

# Check PostgreSQL
echo -n "PostgreSQL: "
if docker exec solobueno-postgres pg_isready -U solobueno -d solobueno_dev > /dev/null 2>&1; then
    echo "✓ healthy"
else
    echo "✗ unhealthy"
    exit 1
fi

# Check Redis
echo -n "Redis: "
if docker exec solobueno-redis redis-cli ping | grep -q PONG; then
    echo "✓ healthy"
else
    echo "✗ unhealthy"
    exit 1
fi

# Check MinIO
echo -n "MinIO: "
if curl -sf http://localhost:9000/minio/health/live > /dev/null 2>&1; then
    echo "✓ healthy"
else
    echo "✗ unhealthy"
    exit 1
fi

echo ""
echo "All services healthy!"
```

## Makefile Targets

**File**: `Makefile` (additions)

```makefile
# =============================================================================
# Docker Services (Enhanced)
# =============================================================================

.PHONY: docker-up docker-down docker-restart docker-reset docker-status
.PHONY: docker-logs docker-logs-postgres docker-logs-redis docker-logs-minio
.PHONY: docker-shell-postgres docker-health

DOCKER_COMPOSE := docker compose -f infrastructure/docker/docker-compose.yml

docker-up: ## Start Docker services (PostgreSQL, Redis, MinIO)
	@command -v docker >/dev/null 2>&1 || (echo "Error: Docker not found" && exit 1)
	@docker info >/dev/null 2>&1 || (echo "Error: Docker is not running" && exit 1)
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
	@echo "Services reset with fresh data."

docker-status: ## Show service status and health
	@echo ""
	@echo "Service Status:"
	@$(DOCKER_COMPOSE) ps
	@echo ""
	@echo "Health Check:"
	@bash infrastructure/scripts/health-check.sh || true

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

docker-health: ## Run health check script
	@bash infrastructure/scripts/health-check.sh
```

## Service Connectivity Information

### Connection Strings (for development)

| Service       | Connection String / URL                                                           |
| ------------- | --------------------------------------------------------------------------------- |
| PostgreSQL    | `postgres://solobueno:solobueno_dev@localhost:5432/solobueno_dev?sslmode=disable` |
| Redis         | `redis://localhost:6379`                                                          |
| MinIO API     | `http://localhost:9000`                                                           |
| MinIO Console | `http://localhost:9001`                                                           |

### Internal Network (container-to-container)

| Service    | Internal Hostname                            |
| ---------- | -------------------------------------------- |
| PostgreSQL | `postgres:5432` or `solobueno-postgres:5432` |
| Redis      | `redis:6379` or `solobueno-redis:6379`       |
| MinIO      | `minio:9000` or `solobueno-minio:9000`       |

## Volume Management

### Inspect Volumes

```bash
# List all Solobueno volumes
docker volume ls | grep solobueno

# Inspect specific volume
docker volume inspect solobueno_postgres_data
```

### Backup/Restore (manual)

```bash
# Backup PostgreSQL
docker exec solobueno-postgres pg_dump -U solobueno solobueno_dev > backup.sql

# Restore PostgreSQL
cat backup.sql | docker exec -i solobueno-postgres psql -U solobueno -d solobueno_dev
```

## Directory Structure

```text
infrastructure/
├── docker/
│   └── docker-compose.yml    # Main compose file (enhanced)
├── scripts/
│   └── health-check.sh       # Health verification script (new)
└── config/
    ├── dev.env.example       # Dev environment (exists)
    ├── test.env.example      # Test environment (exists)
    ├── staging.env.example   # Staging environment (exists)
    └── prod.env.example      # Prod environment (exists)
```
