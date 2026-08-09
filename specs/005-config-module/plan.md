# Implementation Plan: Configuration Module

**Branch**: `005-config-module` | **Date**: 2026-08-09 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/005-config-module/spec.md`

## Summary

Per-tenant configuration (branding, business settings, feature flags) plus platform-wide
global settings, backed by PostgreSQL with a 60-second-TTL Redis read-through cache so reads
stay under 100ms (SC-005) while changes still propagate within the spec's own 1-minute bound
(FR-005) without needing invalidation plumbing. REST API under `/api/v1/config`, gated by
004-auth-module's existing `AuthMiddleware`/role hierarchy — no new auth concepts. US4
(cross-tenant global-settings management) is descoped from this pass; see research.md.

## Technical Context

**Language/Version**: Go 1.22+
**Primary Dependencies**: Chi (REST, reused from auth), GORM (ORM, reused from auth),
`github.com/redis/go-redis/v9` (new)
**Storage**: PostgreSQL 16 (`tenant_configs`, `feature_flags`, `global_configs`,
`config_changes`), Redis 7 (cache, first consumer in this codebase)
**Testing**: go test, table-driven, SQLite in-memory for repository tests (matches auth's
`repository_test.go` pattern); `miniredis` or a real Redis container for cache-layer tests
**Target Platform**: Linux server (Docker), same deployment as auth
**Project Type**: Backend module (modular monolith), depends on 004-auth-module
**Performance Goals**: Cached config reads <100ms (SC-005); config propagation <1 minute
(SC-001, met by TTL alone)
**Constraints**: REST-only (Complexity Tracking deviation from Principle III, see below);
JSONB storage for branding/business-settings to avoid per-field migrations
**Scale/Scope**: Multi-tenant; 1 `tenant_configs` row + N `feature_flags` rows per tenant;
global scope via nullable `tenant_id`

## Constitution Check

_GATE: Must pass before Phase 0 research. Re-check after Phase 1 design._

| Principle                      | Status       | Notes                                                                                                                                                       |
| ------------------------------ | ------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------- |
| I. Mobile-First                | N/A          | Backend module, no UI                                                                                                                                       |
| II. Domain-Driven Design       | ✅ Pass      | `config` is a bounded context per constitution's module table; events defined for future cross-module reactions                                             |
| III. API-First                 | ⚠️ Deviation | REST, not GraphQL — no GraphQL server exists yet; same precedent as auth. Documented in Complexity Tracking                                                 |
| IV. Offline-First              | ✅ Pass      | Mobile app pre-caches config per constitution's own IV; server-side TTL cache doesn't block that                                                            |
| V. Plugin-Driven               | N/A          | No plugins for config                                                                                                                                       |
| VI. White-Label & Multi-Tenant | ✅ Pass      | This module _is_ the white-label mechanism (branding, per-tenant business settings)                                                                         |
| VII. Type Safety               | ✅ Pass      | Go strongly typed; JSONB columns map to typed structs, not `map[string]any`, at the domain layer                                                            |
| VIII. Test-Driven              | ✅ Pass      | Repository/service/handler tests planned per auth's established pattern                                                                                     |
| IX. Internationalization       | ✅ Pass      | Currency/timezone are exactly what this module configures per-tenant                                                                                        |
| X. User-Centric                | ✅ Pass      | Sub-100ms cached reads (SC-005), no restart required for changes (FR-012)                                                                                   |
| XI. Observability              | ✅ Pass      | `config_changes` audit table (FR-008/SC-006); reuses `internal/shared/observability` logger for error paths                                                 |
| XII. Security                  | ✅ Pass      | Role-gated writes via existing `AuthMiddleware`; global-settings writes deliberately not implemented rather than under-authorized (see Complexity Tracking) |

**Gate Status**: ✅ PASSED — one documented deviation (Principle III), justified below; no
unjustified violations.

## Project Structure

### Documentation (this feature)

```text
specs/005-config-module/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output (REST API contracts)
└── tasks.md             # Phase 2 output
```

### Source Code (repository root)

```text
backend/
├── internal/
│   ├── config/
│   │   ├── domain/
│   │   │   ├── tenant_config.go   # TenantConfig, BrandingSettings, BusinessSettings
│   │   │   ├── feature_flag.go    # FeatureFlag entity + resolution rule
│   │   │   ├── global_config.go   # GlobalConfig entity (read path only)
│   │   │   ├── config_change.go   # ConfigChange audit entity
│   │   │   └── errors.go          # Domain errors
│   │   ├── repository/
│   │   │   ├── tenant_config_repo.go
│   │   │   ├── feature_flag_repo.go
│   │   │   ├── global_config_repo.go
│   │   │   ├── config_change_repo.go
│   │   │   └── repository_test.go
│   │   ├── service/
│   │   │   ├── config_service.go  # Branding/business-settings read+update, validation
│   │   │   └── feature_service.go # Flag read+toggle, tenant/global resolution
│   │   ├── handler/
│   │   │   ├── config_handler.go  # REST endpoints
│   │   │   └── dto.go
│   │   ├── events.go               # ConfigChanged, FeatureFlagToggled (defined, unwired)
│   │   └── service.go              # Public module interface (mirrors auth.Module)
│   └── shared/
│       └── cache/
│           ├── cache.go            # Cache interface (Get/Set/Delete) - new, was .gitkeep
│           └── redis_cache.go      # Redis adapter (go-redis/v9)
├── migrations/
│   ├── 002_config_tables.up.sql
│   └── 002_config_tables.down.sql
└── cmd/server/main.go              # Wire config.NewModule + mount routes (edit, not new)
```

**Structure Decision**: Mirrors `internal/auth`'s established layout (domain/repository/
service/handler + module-level `service.go`) for consistency and so future
`/speckit-converge`-style gap analysis reuses the same mental model across modules. New
`shared/cache` package fills the placeholder the constitution's own directory layout already
reserved for it.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation                                        | Why Needed                                                                                                                                                                   | Simpler Alternative Rejected Because                                                                                                                                                                                                      |
| ------------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| REST instead of GraphQL (Principle III)          | No GraphQL server exists yet in this codebase (`backend/api/graphql/` is an empty placeholder, no `gqlgen` dependency); auth module already established this same precedent. | Standing up gqlgen as part of this module would make "add config CRUD" also mean "build the project's primary API layer" — a separate, much larger decision the user explicitly deferred to a future dedicated feature (see research.md). |
| US4 (global settings write path) not implemented | `domain.Role` has no role above per-tenant `RoleOwner`; no platform-admin authorization concept exists to gate a cross-tenant write endpoint.                                | An env-var user-ID allowlist would bypass real authorization (violates Principle XII) and would need to be redone once platform-admin auth exists — worse than leaving it undone and tracked.                                             |
