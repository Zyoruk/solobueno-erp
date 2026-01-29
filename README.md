# Solobueno ERP

A modern restaurant ERP system built with Go, React Native, and TypeScript.

## Features

- **Mobile-First**: Order taking and table management optimized for tablets and phones
- **Multi-Tenant**: White-label solution supporting multiple restaurants
- **Offline-First**: Works without internet connection, syncs when online
- **Plugin-Driven**: Extensible billing and payment integrations
- **Internationalized**: Built-in support for Spanish and English

## Quick Start

### Prerequisites

- Node.js 20+
- pnpm 8+
- Go 1.22+
- Docker & Docker Compose

### Setup

```bash
# Clone the repository
git clone https://github.com/solobueno/erp.git
cd solobueno-erp

# Install dependencies and start services
make setup

# Copy environment file
cp infrastructure/config/dev.env.example .env

# Start development
make dev
```

### Available Commands

```bash
make help          # Show all available commands
make install       # Install dependencies
make dev           # Start development servers
make build         # Build all packages
make test          # Run tests
make docker-up     # Start Docker services
make docker-down   # Stop Docker services
```

## Project Structure

```
solobueno-erp/
├── apps/                    # Applications
│   ├── mobile/              # React Native mobile app
│   ├── kitchen-display/     # Kitchen display tablet app
│   ├── backoffice/          # Manager web app
│   └── admin/               # Platform admin web app
├── backend/                 # Go backend
│   ├── cmd/                 # Entry points
│   ├── internal/            # Domain modules
│   ├── api/                 # GraphQL & REST APIs
│   └── plugins/             # Billing & payment plugins
├── packages/                # Shared packages
│   ├── ui/                  # UI components
│   ├── i18n/                # Translations
│   ├── types/               # TypeScript types
│   ├── graphql-client/      # GraphQL client
│   └── analytics/           # Analytics tracking
├── infrastructure/          # Docker, K8s, configs
├── docs/                    # Documentation
└── specs/                   # Feature specifications
```

## Architecture

- **Backend**: Go with Chi router, gqlgen for GraphQL
- **Frontend**: React Native (mobile), React (web)
- **Database**: PostgreSQL 16
- **Cache**: Redis 7
- **Storage**: S3-compatible (MinIO for dev)
- **Monorepo**: Turborepo with pnpm workspaces

## Documentation

- [Project Constitution](.specify/memory/constitution.md)
- [Feature Specifications](specs/)
- [API Documentation](docs/api/)

## License

Proprietary - All rights reserved
