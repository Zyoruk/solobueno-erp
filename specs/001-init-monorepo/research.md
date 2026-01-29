# Research: Initialize Monorepo Structure

**Feature**: 001-init-monorepo  
**Date**: 2025-01-29

## Research Tasks

### 1. Turborepo Configuration Best Practices

**Decision**: Use Turborepo with pnpm workspaces

**Rationale**:
- Turborepo provides excellent build caching and incremental builds
- pnpm offers fast installation and efficient disk usage via hard links
- Both are well-maintained and widely adopted in the industry
- Constitution already specifies Turborepo as the monorepo tool

**Alternatives Considered**:
- Nx: More features but higher complexity, steeper learning curve
- Lerna: Less active maintenance, Turborepo is the modern successor
- Rush: Enterprise-focused, overkill for this project size

**Key Configuration Points**:
```json
// turbo.json
{
  "$schema": "https://turbo.build/schema.json",
  "globalDependencies": ["**/.env.*local"],
  "pipeline": {
    "build": {
      "dependsOn": ["^build"],
      "outputs": ["dist/**", ".next/**", "build/**"]
    },
    "test": {
      "dependsOn": ["build"],
      "outputs": []
    },
    "lint": {
      "outputs": []
    },
    "dev": {
      "cache": false,
      "persistent": true
    }
  }
}
```

### 2. Package Manager Selection

**Decision**: pnpm 8.x with workspaces

**Rationale**:
- Strict dependency resolution prevents phantom dependencies
- Efficient disk usage through content-addressable storage
- Native workspace support works seamlessly with Turborepo
- Faster installation than npm/yarn

**Alternatives Considered**:
- npm workspaces: Slower, less efficient disk usage
- Yarn (classic): Adequate but pnpm is faster
- Yarn (berry/PnP): Compatibility issues with some tools

**Workspace Configuration**:
```yaml
# pnpm-workspace.yaml
packages:
  - 'apps/*'
  - 'packages/*'
```

### 3. Go Module Structure for Monorepo

**Decision**: Single Go module at `backend/` with internal packages

**Rationale**:
- Simpler dependency management with single go.mod
- Internal packages enforce encapsulation
- Standard Go project layout conventions
- Easy to extract modules later if needed

**Alternatives Considered**:
- Multiple Go modules: Adds complexity, harder dependency management
- Go workspace (go.work): Newer feature, less tooling support

**Module Structure**:
```
backend/
├── go.mod              # github.com/solobueno/erp
├── go.sum
├── cmd/                # Executables
│   └── server/
└── internal/           # Private packages (cannot be imported externally)
    └── [modules]/
```

### 4. TypeScript Configuration Strategy

**Decision**: Shared base tsconfig with per-package extensions

**Rationale**:
- Consistent TypeScript settings across all packages
- Per-package overrides for specific needs (React Native vs Web)
- Enables project references for faster incremental builds

**Configuration Hierarchy**:
```
tsconfig.base.json          # Shared strict settings
├── apps/mobile/tsconfig.json
├── apps/backoffice/tsconfig.json
├── packages/ui/tsconfig.json
└── packages/types/tsconfig.json
```

### 5. Docker Compose for Local Development

**Decision**: Single docker-compose.yml for all local services

**Rationale**:
- Simple one-command startup for all dependencies
- Consistent environment across developer machines
- Easy to add/remove services as needed

**Services**:
- PostgreSQL 16 (port 5432)
- Redis 7 (port 6379)
- MinIO for S3-compatible storage (port 9000, console 9001)

### 6. Pre-commit Hooks Strategy

**Decision**: Husky + lint-staged for Git hooks

**Rationale**:
- Industry standard for JavaScript/TypeScript projects
- Works with pnpm workspaces
- Can run different checks for different file types

**Alternatives Considered**:
- pre-commit (Python): Would require Python installation
- lefthook: Good but less ecosystem support
- Simple git hooks: Harder to maintain and share

### 7. Code Formatting and Linting

**Decision**: Prettier + ESLint for TypeScript, golangci-lint for Go

**Rationale**:
- Prettier for consistent formatting (no debates)
- ESLint for code quality rules
- golangci-lint bundles multiple Go linters

**Configuration Files**:
- `.prettierrc` - Formatting rules
- `.eslintrc.js` - Linting rules
- `.golangci.yml` - Go linting rules

### 8. CI/CD Initial Setup

**Decision**: GitHub Actions with basic CI workflow

**Rationale**:
- Constitution specifies GitHub Actions
- Free tier sufficient for initial development
- Easy integration with GitHub repository

**Initial Workflow**:
- Lint all code
- Build all packages
- Run tests
- Cache pnpm and Turborepo artifacts

## Summary of Decisions

| Area | Decision |
|------|----------|
| Monorepo Tool | Turborepo |
| Package Manager | pnpm 8.x |
| Go Structure | Single module with internal packages |
| TypeScript | Shared base tsconfig |
| Local Services | Docker Compose (PostgreSQL, Redis, MinIO) |
| Git Hooks | Husky + lint-staged |
| Formatting | Prettier + ESLint + golangci-lint |
| CI/CD | GitHub Actions |

## Open Questions Resolved

All technical decisions have been made based on constitution requirements and industry best practices. No outstanding clarifications needed.
