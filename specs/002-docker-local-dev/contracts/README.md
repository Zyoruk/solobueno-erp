# Contracts: Docker Local Development

**Feature**: 002-docker-local-dev

## Overview

This feature is an infrastructure/DevOps feature that does not expose APIs. Therefore, there are no API contracts to define.

## Service Contracts

Instead of API contracts, this feature defines **service contracts** - the expected behavior of Docker services:

### PostgreSQL Service Contract

| Property     | Value                                      |
| ------------ | ------------------------------------------ |
| Image        | `postgres:16-alpine`                       |
| Port         | `5432`                                     |
| Database     | `solobueno_dev`                            |
| Username     | `solobueno`                                |
| Password     | `solobueno_dev`                            |
| Health Check | `pg_isready -U solobueno -d solobueno_dev` |
| Persistence  | Named volume `solobueno_postgres_data`     |

### Redis Service Contract

| Property       | Value                               |
| -------------- | ----------------------------------- |
| Image          | `redis:7-alpine`                    |
| Port           | `6379`                              |
| Authentication | None (local dev only)               |
| Persistence    | AOF (append-only file)              |
| Health Check   | `redis-cli ping`                    |
| Persistence    | Named volume `solobueno_redis_data` |

### MinIO Service Contract

| Property     | Value                                          |
| ------------ | ---------------------------------------------- |
| Image        | `minio/minio:latest`                           |
| API Port     | `9000`                                         |
| Console Port | `9001`                                         |
| Access Key   | `solobueno`                                    |
| Secret Key   | `solobueno_dev`                                |
| Health Check | `curl http://localhost:9000/minio/health/live` |
| Persistence  | Named volume `solobueno_minio_data`            |

## Makefile Target Contract

| Target           | Behavior                            |
| ---------------- | ----------------------------------- |
| `docker-up`      | Start all services, wait for health |
| `docker-down`    | Stop services, preserve data        |
| `docker-restart` | Restart services                    |
| `docker-reset`   | Delete data and restart             |
| `docker-logs`    | Tail all service logs               |
| `docker-status`  | Show status and health              |
| `docker-health`  | Run health check script             |
