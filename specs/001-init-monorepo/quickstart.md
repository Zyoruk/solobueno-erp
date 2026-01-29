# Quickstart: Solobueno ERP Development

**Feature**: 001-init-monorepo  
**Date**: 2025-01-29

## Prerequisites

Before you begin, ensure you have the following installed:

| Tool | Required Version | Check Command |
|------|------------------|---------------|
| Node.js | 20.x or higher | `node --version` |
| pnpm | 8.x or higher | `pnpm --version` |
| Go | 1.22 or higher | `go version` |
| Docker | Latest | `docker --version` |
| Docker Compose | Latest (bundled with Docker Desktop) | `docker compose version` |
| Git | Latest | `git --version` |

### Installing Prerequisites

**Node.js** (via nvm recommended):
```bash
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash
nvm install 20
nvm use 20
```

**pnpm**:
```bash
npm install -g pnpm@8
```

**Go**:
- macOS: `brew install go`
- Linux: Download from https://go.dev/dl/
- Windows: Download installer from https://go.dev/dl/

**Docker Desktop**:
- Download from https://www.docker.com/products/docker-desktop

## Quick Setup (5 minutes)

### 1. Clone the Repository

```bash
git clone https://github.com/solobueno/erp.git
cd solobueno-erp
```

### 2. Install Dependencies

```bash
# Install all JavaScript/TypeScript dependencies
pnpm install

# Install Go dependencies
cd backend && go mod download && cd ..
```

### 3. Start Local Services

```bash
# Start PostgreSQL, Redis, and MinIO
docker compose -f infrastructure/docker/docker-compose.yml up -d

# Verify services are running
docker compose -f infrastructure/docker/docker-compose.yml ps
```

### 4. Configure Environment

```bash
# Copy environment template
cp infrastructure/config/dev.env.example .env

# Edit if needed (defaults work for local development)
```

### 5. Verify Setup

```bash
# Build all packages
pnpm build

# Run linting
pnpm lint

# Run tests
pnpm test
```

## Development Commands

### Root Commands (from repository root)

| Command | Description |
|---------|-------------|
| `pnpm dev` | Start all apps in development mode |
| `pnpm build` | Build all packages and apps |
| `pnpm test` | Run all tests |
| `pnpm lint` | Lint all code |
| `pnpm format` | Format all code with Prettier |
| `pnpm clean` | Remove all build artifacts and node_modules |

### Filtered Commands (specific packages)

```bash
# Run dev for specific app
pnpm --filter @solobueno/mobile dev
pnpm --filter @solobueno/backoffice dev

# Build specific package
pnpm --filter @solobueno/ui build

# Test specific package
pnpm --filter @solobueno/types test
```

### Backend Commands

```bash
cd backend

# Run backend server
go run cmd/server/main.go

# Run tests
go test ./...

# Run with live reload (requires air)
air
```

### Docker Commands

```bash
# Start services
docker compose -f infrastructure/docker/docker-compose.yml up -d

# Stop services
docker compose -f infrastructure/docker/docker-compose.yml down

# View logs
docker compose -f infrastructure/docker/docker-compose.yml logs -f

# Reset data (removes volumes)
docker compose -f infrastructure/docker/docker-compose.yml down -v
```

## Project Structure Overview

```
solobueno-erp/
├── apps/                   # Application packages
│   ├── mobile/             # React Native mobile app
│   ├── kitchen-display/    # Kitchen display tablet app
│   ├── backoffice/         # Back office web app
│   └── admin/              # Admin portal web app
│
├── backend/                # Go backend
│   ├── cmd/                # Application entrypoints
│   ├── internal/           # Private packages (domain modules)
│   ├── api/                # API definitions (GraphQL, REST)
│   ├── migrations/         # Database migrations
│   └── plugins/            # Plugin implementations
│
├── packages/               # Shared packages
│   ├── ui/                 # Shared UI components
│   ├── i18n/               # Internationalization
│   ├── types/              # Shared TypeScript types
│   ├── graphql-client/     # Generated GraphQL client
│   └── analytics/          # Analytics helpers
│
├── infrastructure/         # Infrastructure configuration
│   ├── docker/             # Docker configurations
│   └── config/             # Environment configurations
│
└── docs/                   # Documentation
```

## Common Tasks

### Adding a New Package

1. Create the package directory:
   ```bash
   mkdir -p packages/my-package/src
   ```

2. Create `package.json`:
   ```json
   {
     "name": "@solobueno/my-package",
     "version": "0.0.1",
     "private": true,
     "main": "./dist/index.js",
     "scripts": {
       "build": "tsup src/index.ts --format cjs,esm --dts",
       "dev": "tsup src/index.ts --format cjs,esm --dts --watch"
     }
   }
   ```

3. Create `src/index.ts`:
   ```typescript
   export const hello = () => 'Hello from my-package';
   ```

4. Install dependencies and build:
   ```bash
   pnpm install
   pnpm --filter @solobueno/my-package build
   ```

### Adding a Backend Module

1. Create the module directory:
   ```bash
   mkdir -p backend/internal/my-module/{domain,ports,adapters,application}
   ```

2. Create the module entry point `backend/internal/my-module/module.go`:
   ```go
   package mymodule

   type Module struct {
       // Module dependencies
   }

   func NewModule() *Module {
       return &Module{}
   }
   ```

### Running Database Migrations

```bash
cd backend

# Create a new migration
go run cmd/migrate/main.go create my_migration

# Run migrations
go run cmd/migrate/main.go up

# Rollback last migration
go run cmd/migrate/main.go down 1
```

## Troubleshooting

### Common Issues

**pnpm install fails with peer dependency errors**
```bash
pnpm install --shamefully-hoist
```

**Docker containers won't start**
```bash
# Check if ports are already in use
lsof -i :5432  # PostgreSQL
lsof -i :6379  # Redis
lsof -i :9000  # MinIO

# Kill conflicting processes or change ports in docker-compose.yml
```

**Go module issues**
```bash
cd backend
go mod tidy
go mod download
```

**Turborepo cache issues**
```bash
pnpm clean
rm -rf .turbo
pnpm install
pnpm build
```

## Getting Help

- Check the `docs/` directory for detailed documentation
- Review Architecture Decision Records in `docs/adr/`
- Consult the project constitution in `.specify/memory/constitution.md`
