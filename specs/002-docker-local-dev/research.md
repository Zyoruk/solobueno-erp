# Research: Docker Local Development Environment

**Feature**: 002-docker-local-dev  
**Date**: 2025-01-29

## Research Tasks

### 1. Docker Compose Health Check Best Practices

**Decision**: Use built-in health check commands for each service

**Rationale**:

- Native health checks integrate with Docker Compose `depends_on` with `condition: service_healthy`
- No external dependencies required
- Standard approach across all container orchestration platforms

**Service-Specific Checks**:

| Service    | Health Check Command                              | Interval | Retries |
| ---------- | ------------------------------------------------- | -------- | ------- |
| PostgreSQL | `pg_isready -U solobueno`                         | 10s      | 5       |
| Redis      | `redis-cli ping`                                  | 10s      | 5       |
| MinIO      | `curl -f http://localhost:9000/minio/health/live` | 30s      | 3       |

**Alternatives Considered**:

- External health check service: Adds complexity, not needed for local dev
- TCP port checks: Less reliable than service-specific checks

### 2. Data Persistence Strategy

**Decision**: Named Docker volumes with explicit volume management

**Rationale**:

- Named volumes survive container removal
- Easy to inspect and backup
- Clear separation between data and containers

**Volume Mapping**:

| Service    | Volume Name               | Mount Path                 |
| ---------- | ------------------------- | -------------------------- |
| PostgreSQL | `solobueno_postgres_data` | `/var/lib/postgresql/data` |
| Redis      | `solobueno_redis_data`    | `/data`                    |
| MinIO      | `solobueno_minio_data`    | `/data`                    |

**Alternatives Considered**:

- Bind mounts: Platform-specific paths, permission issues
- Anonymous volumes: Lost on container removal

### 3. Makefile Target Design

**Decision**: Prefix all Docker targets with `docker-` for clarity

**Rationale**:

- Clear namespace prevents conflicts with other targets
- Consistent with 001-init-monorepo patterns
- Self-documenting when running `make help`

**Target List**:

| Target                  | Purpose                           |
| ----------------------- | --------------------------------- |
| `docker-up`             | Start all services                |
| `docker-down`           | Stop all services (preserve data) |
| `docker-restart`        | Stop and start services           |
| `docker-reset`          | Stop, remove volumes, restart     |
| `docker-logs`           | Tail logs from all services       |
| `docker-logs-postgres`  | Tail PostgreSQL logs              |
| `docker-logs-redis`     | Tail Redis logs                   |
| `docker-logs-minio`     | Tail MinIO logs                   |
| `docker-status`         | Show service status and health    |
| `docker-shell-postgres` | Open psql shell                   |

**Alternatives Considered**:

- Individual scripts: Harder to discover, more files to maintain
- Docker Compose profiles: Overkill for simple start/stop

### 4. Service Port Configuration

**Decision**: Use standard ports with clear documentation

**Rationale**:

- Standard ports are expected by tools and libraries
- Easy to remember and type
- Matches production conventions

**Port Mapping**:

| Service       | Host Port | Container Port | Purpose                 |
| ------------- | --------- | -------------- | ----------------------- |
| PostgreSQL    | 5432      | 5432           | Database connections    |
| Redis         | 6379      | 6379           | Cache/queue connections |
| MinIO API     | 9000      | 9000           | S3-compatible API       |
| MinIO Console | 9001      | 9001           | Web management UI       |

**Alternatives Considered**:

- Non-standard ports: Avoids conflicts but adds confusion
- Dynamic ports: Requires lookup, breaks simple configs

### 5. Error Handling for Common Issues

**Decision**: Add pre-flight checks in Makefile targets

**Rationale**:

- Better developer experience with clear error messages
- Prevents confusing Docker errors
- Documents prerequisites

**Checks Implemented**:

| Check              | Error Message                                  | Resolution                  |
| ------------------ | ---------------------------------------------- | --------------------------- |
| Docker not running | "Docker is not running. Start Docker Desktop." | Start Docker                |
| Port in use        | "Port 5432 already in use by [process]"        | Kill process or change port |
| Low disk space     | "Warning: Less than 5GB free disk space"       | Clean up disk               |

**Alternatives Considered**:

- Let Docker fail naturally: Cryptic error messages
- Separate check script: Extra step for developers

### 6. Development Credentials

**Decision**: Use simple, memorable credentials for local dev only

**Rationale**:

- Easy to type and remember
- Clearly marked as development-only
- Consistent across all services

**Credentials**:

| Service    | Username    | Password        |
| ---------- | ----------- | --------------- |
| PostgreSQL | `solobueno` | `solobueno_dev` |
| MinIO      | `solobueno` | `solobueno_dev` |
| Redis      | (no auth)   | N/A             |

**Security Note**: These credentials are for local development only. Production uses AWS SSM Parameter Store for secrets.

## Summary of Decisions

| Area             | Decision                                |
| ---------------- | --------------------------------------- |
| Health Checks    | Native Docker health check commands     |
| Data Persistence | Named Docker volumes                    |
| Makefile Design  | `docker-*` prefixed targets             |
| Ports            | Standard ports (5432, 6379, 9000, 9001) |
| Error Handling   | Pre-flight checks in Makefile           |
| Credentials      | Simple dev credentials                  |

## Open Questions Resolved

All technical decisions made based on Docker best practices and constitution requirements. No outstanding clarifications needed.
