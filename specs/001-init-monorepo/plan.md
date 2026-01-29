# Implementation Plan: Initialize Monorepo Structure

**Branch**: `001-init-monorepo` | **Date**: 2025-01-29 | **Spec**: [spec.md](./spec.md)  
**Input**: Feature specification from `/specs/001-init-monorepo/spec.md`

## Summary

Initialize the Solobueno ERP monorepo with Turborepo, establishing the complete folder structure defined in the project constitution. This includes scaffolding for all applications (mobile, web, backend), shared packages, infrastructure configurations, and development tooling. The goal is to enable developers to clone and start working within 10 minutes.

## Technical Context

**Language/Version**: Go 1.22+ (backend), TypeScript 5.x (frontend), Node.js 20+  
**Primary Dependencies**: Turborepo, pnpm (workspace management), Docker Compose  
**Storage**: N/A (structure only, no database setup in this feature)  
**Testing**: Vitest (frontend), go test (backend) - scaffolding only  
**Target Platform**: macOS, Linux, Windows (via WSL)  
**Project Type**: Monorepo with mobile + web + backend  
**Performance Goals**: Full build < 2 minutes, cached build < 30 seconds  
**Constraints**: Must work offline after initial setup, cross-platform compatible  
**Scale/Scope**: 4 apps, 5 shared packages, 1 backend, infrastructure configs

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Mobile-First | ✅ Pass | Structure includes `apps/mobile/` as primary app |
| II. Domain-Driven Design | ✅ Pass | Backend structure follows module pattern from constitution |
| III. API-First | ✅ Pass | Includes `backend/api/graphql/` and `backend/api/rest/` |
| IV. Offline-First | ✅ Pass | Dev environment works offline after setup |
| V. Plugin-Driven | ✅ Pass | Includes `backend/plugins/` directory |
| VI. White-Label & Multi-Tenant | ✅ Pass | Structure supports; implementation in future features |
| VII. Type Safety | ✅ Pass | TypeScript strict mode, Go native typing |
| VIII. Test-Driven | ✅ Pass | Test directories included in structure |
| IX. Internationalization | ✅ Pass | Includes `packages/i18n/` |
| X. User-Centric | N/A | UX not applicable to project structure |
| XI. Observability | ✅ Pass | Includes `shared/observability/` |
| XII. Security | N/A | Security implementation in future features |

**Gate Status**: ✅ PASSED - All applicable principles satisfied

## Project Structure

### Documentation (this feature)

```text
specs/001-init-monorepo/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output (minimal - config files)
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output (minimal)
└── checklists/
    └── requirements.md  # Spec validation checklist
```

### Source Code (repository root)

```text
solobueno-erp/
├── apps/
│   ├── mobile/                 # React Native app (placeholder)
│   │   ├── package.json
│   │   └── src/
│   ├── kitchen-display/        # React Native tablet app (placeholder)
│   │   ├── package.json
│   │   └── src/
│   ├── backoffice/             # React web app (placeholder)
│   │   ├── package.json
│   │   └── src/
│   └── admin/                  # React web app (placeholder)
│       ├── package.json
│       └── src/
│
├── backend/
│   ├── go.mod
│   ├── go.sum
│   ├── cmd/
│   │   ├── server/
│   │   │   └── main.go         # Placeholder entrypoint
│   │   └── migrate/
│   │       └── main.go         # Placeholder migration CLI
│   ├── internal/
│   │   ├── auth/
│   │   ├── menu/
│   │   ├── orders/
│   │   ├── tables/
│   │   ├── inventory/
│   │   ├── billing/
│   │   ├── payments/
│   │   ├── reporting/
│   │   ├── config/
│   │   ├── feedback/
│   │   ├── analytics/
│   │   ├── notifications/
│   │   ├── audit/
│   │   ├── jobs/
│   │   ├── media/
│   │   ├── search/
│   │   └── shared/
│   │       ├── events/
│   │       ├── types/
│   │       ├── saga/
│   │       ├── observability/
│   │       ├── cache/
│   │       ├── errors/
│   │       └── resilience/
│   ├── api/
│   │   ├── graphql/
│   │   └── rest/
│   ├── migrations/
│   └── plugins/
│       ├── billing/
│       │   ├── costarica/
│       │   └── generic/
│       ├── payments/
│       │   ├── stripe/
│       │   └── manual/
│       └── analytics/
│           └── mixpanel/
│
├── packages/
│   ├── ui/                     # @solobueno/ui
│   │   ├── package.json
│   │   └── src/
│   ├── i18n/                   # @solobueno/i18n
│   │   ├── package.json
│   │   └── src/
│   │       └── locales/
│   │           ├── es-419.json
│   │           └── en.json
│   ├── types/                  # @solobueno/types
│   │   ├── package.json
│   │   └── src/
│   ├── graphql-client/         # @solobueno/graphql
│   │   ├── package.json
│   │   └── src/
│   └── analytics/              # @solobueno/analytics
│       ├── package.json
│       └── src/
│
├── tools/
│   ├── codegen/
│   └── scripts/
│
├── docs/
│   ├── adr/
│   ├── events/
│   └── api/
│
├── infrastructure/
│   ├── docker/
│   │   ├── Dockerfile.backend
│   │   └── docker-compose.yml
│   ├── k8s/
│   │   ├── base/
│   │   └── overlays/
│   │       ├── dev/
│   │       ├── test/
│   │       ├── staging/
│   │       └── prod/
│   └── config/
│       ├── dev.env.example
│       ├── test.env.example
│       ├── staging.env.example
│       └── prod.env.example
│
├── .github/
│   └── workflows/
│       └── ci.yml
│
├── turbo.json                  # Turborepo configuration
├── pnpm-workspace.yaml         # pnpm workspace configuration
├── package.json                # Root package.json
├── .gitignore
├── .nvmrc                      # Node version
├── .editorconfig               # Editor configuration
├── .prettierrc                 # Prettier configuration
├── .eslintrc.js                # ESLint configuration
├── Makefile                    # Common commands
└── README.md                   # Project documentation
```

**Structure Decision**: Monorepo with Turborepo following the exact layout specified in the constitution. Apps are React/React Native with TypeScript, backend is Go with domain modules.

## Complexity Tracking

No violations - structure matches constitution exactly.
