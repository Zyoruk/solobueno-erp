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

## Knowledge Graph (graphify)

This repo can be explored as a knowledge graph instead of raw grep/browsing.
[graphify](https://github.com/safishamsi/graphify) extracts code (AST) and
docs/specs (semantic) into a graph with community detection, "god node"
analysis, and cross-file relationship queries. The built graph is checked
into [`graphify-out/`](graphify-out/) (`graph.json`, `graph.html`,
`GRAPH_REPORT.md`) so it's queryable without a build step; `graphify-out/cache/`
and `cost.json` are local bookkeeping and stay gitignored.

### Install

```bash
pip install graphifyy
# or, with uv:
uv tool install graphifyy
```

If you use [Claude Code](https://claude.com/claude-code), the
[graphify skill](https://github.com/safishamsi/graphify) is available as a
slash command — no separate install step needed beyond the skill itself.

### Build and query

```bash
# Build/refresh the graph for the whole repo (writes to graphify-out/, committed)
graphify .

# Ask a question - returns a scoped subgraph instead of a full-text dump
graphify query "How does the auth module handle token refresh?"

# Trace a relationship between two concepts
graphify path "AuthModule" "OrdersModule"

# Explain a single concept
graphify explain "Saga Pattern"
```

In Claude Code, use `/graphify` (build) and `/graphify query "<question>"`
(ask) directly - see [CLAUDE.md](CLAUDE.md) for the project's rules on when
to query the graph vs. read source directly. After changing code, run
`graphify update .` to refresh the graph incrementally (AST-only, no LLM
cost) and commit the updated `graphify-out/` files with your change.

## License

Proprietary - All rights reserved
