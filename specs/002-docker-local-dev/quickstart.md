# Quickstart: Docker Local Development

**Feature**: 002-docker-local-dev  
**Audience**: Developers setting up local development environment

## Prerequisites

- Docker Desktop 4.0+ (macOS/Windows) or Docker Engine 20+ (Linux)
- Docker Compose V2 (included with Docker Desktop)
- GNU Make (pre-installed on macOS/Linux; install via Git Bash on Windows)

## Quick Start

### 1. Start All Services

```bash
make docker-up
```

This starts:

- **PostgreSQL 16** on port 5432
- **Redis 7** on port 6379
- **MinIO** on ports 9000 (API) and 9001 (Console)

### 2. Verify Services Are Healthy

```bash
make docker-health
```

Expected output:

```
Checking service health...
PostgreSQL: ✓ healthy
Redis: ✓ healthy
MinIO: ✓ healthy

All services healthy!
```

### 3. Connect to Services

**PostgreSQL**:

```bash
# Via psql
make docker-shell-postgres

# Or with connection string
psql "postgres://solobueno:solobueno_dev@localhost:5432/solobueno_dev"
```

**Redis**:

```bash
redis-cli -h localhost -p 6379
```

**MinIO Console**:
Open http://localhost:9001 in your browser

- Username: `solobueno`
- Password: `solobueno_dev`

## Common Operations

### View Logs

```bash
# All services
make docker-logs

# Specific service
make docker-logs-postgres
make docker-logs-redis
make docker-logs-minio
```

### Stop Services

```bash
# Stop but keep data
make docker-down

# Restart
make docker-restart
```

### Reset Everything

```bash
# ⚠️ Deletes all data!
make docker-reset
```

### Check Status

```bash
make docker-status
```

## Troubleshooting

### Docker Not Running

**Error**: `Error: Docker is not running`

**Solution**: Start Docker Desktop or the Docker daemon:

```bash
# macOS
open -a Docker

# Linux
sudo systemctl start docker
```

### Port Already in Use

**Error**: `Bind for 0.0.0.0:5432 failed: port is already allocated`

**Solution**: Find and stop the conflicting process:

```bash
# Find what's using the port
lsof -i :5432

# Kill the process (replace PID)
kill -9 <PID>
```

Or change the port in `docker-compose.yml`.

### Health Check Fails

**Symptom**: Services start but health check fails

**Solution**: Wait a few more seconds and retry:

```bash
sleep 10 && make docker-health
```

If still failing, check logs:

```bash
make docker-logs-postgres  # or redis/minio
```

### Out of Disk Space

**Symptom**: Container won't start, logs show disk errors

**Solution**: Clean up Docker:

```bash
# Remove unused containers, networks, images
docker system prune -a

# Remove unused volumes (⚠️ deletes data!)
docker volume prune
```

### Permission Denied (Linux)

**Symptom**: Docker commands fail with permission denied

**Solution**: Add your user to the docker group:

```bash
sudo usermod -aG docker $USER
# Log out and back in
```

## Connection Details

| Service       | Host      | Port | Username  | Password      |
| ------------- | --------- | ---- | --------- | ------------- |
| PostgreSQL    | localhost | 5432 | solobueno | solobueno_dev |
| Redis         | localhost | 6379 | (none)    | (none)        |
| MinIO API     | localhost | 9000 | solobueno | solobueno_dev |
| MinIO Console | localhost | 9001 | solobueno | solobueno_dev |

## Environment Variables

Copy the example environment file:

```bash
cp infrastructure/config/dev.env.example .env
```

Key variables:

```bash
DATABASE_URL=postgres://solobueno:solobueno_dev@localhost:5432/solobueno_dev
REDIS_URL=redis://localhost:6379
MINIO_ENDPOINT=http://localhost:9000
MINIO_ACCESS_KEY=solobueno
MINIO_SECRET_KEY=solobueno_dev
```

## Data Persistence

Data is stored in named Docker volumes:

- `solobueno_postgres_data` - PostgreSQL database files
- `solobueno_redis_data` - Redis append-only file
- `solobueno_minio_data` - MinIO object storage

Data persists across:

- Container restarts
- `make docker-down` / `make docker-up`
- System reboots

Data is **deleted** by:

- `make docker-reset`
- `docker volume rm solobueno_*`
- `docker volume prune`

## Next Steps

1. Run database migrations: `make migrate-up`
2. Start the backend server: `make backend-run`
3. Start frontend development: `make dev`
