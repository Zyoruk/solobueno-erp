#!/usr/bin/env bash
set -e

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
