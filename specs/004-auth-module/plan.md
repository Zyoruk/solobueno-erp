# Implementation Plan: Authentication Module

**Branch**: `004-auth-module` | **Date**: 2026-01-29 | **Spec**: [spec.md](./spec.md)  
**Input**: Feature specification from `/specs/004-auth-module/spec.md`

## Summary

Implement a complete authentication and authorization module for Solobueno ERP using JWT-based authentication with RS256 signing, Argon2id password hashing, database-backed sessions for revocation, and role-based access control (RBAC). Users have globally unique emails and can belong to multiple tenants with different roles.

## Technical Context

**Language/Version**: Go 1.22+  
**Primary Dependencies**: Chi (REST), GORM (ORM), golang-jwt/jwt, alexedwards/argon2id  
**Storage**: PostgreSQL 16 (users, sessions, auth_events tables)  
**Testing**: go test with table-driven tests  
**Target Platform**: Linux server (Docker/ubuntu-latest)  
**Project Type**: Backend module (modular monolith)  
**Performance Goals**: Login <500ms, token refresh <200ms  
**Constraints**: Argon2id hashing, RS256 JWT, 60min access token, 30-day refresh token  
**Scale/Scope**: Multi-tenant, users can have roles in multiple tenants

## Constitution Check

_GATE: Must pass before Phase 0 research. Re-check after Phase 1 design._

| Principle                      | Status  | Notes                                                                                                                                                                                                                                                                             |
| ------------------------------ | ------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| I. Mobile-First                | N/A     | Backend module, no UI                                                                                                                                                                                                                                                             |
| II. Domain-Driven Design       | ✅ Pass | Auth is a bounded context per constitution                                                                                                                                                                                                                                        |
| III. API-First                 | ✅ Pass | REST endpoints defined in contracts                                                                                                                                                                                                                                               |
| IV. Offline-First              | ✅ Pass | JWT tokens enable offline validation                                                                                                                                                                                                                                              |
| V. Plugin-Driven               | N/A     | No plugins for auth                                                                                                                                                                                                                                                               |
| VI. White-Label & Multi-Tenant | ✅ Pass | Tenant context in tokens, multi-tenant users                                                                                                                                                                                                                                      |
| VII. Type Safety               | ✅ Pass | Go strongly typed                                                                                                                                                                                                                                                                 |
| VIII. Test-Driven              | ✅ Pass | Tests for all auth flows                                                                                                                                                                                                                                                          |
| IX. Internationalization       | N/A     | Error codes, not user-facing messages                                                                                                                                                                                                                                             |
| X. User-Centric                | ✅ Pass | Fast login, session persistence                                                                                                                                                                                                                                                   |
| XI. Observability              | ✅ Pass | Auth events logging (FR-010); operational error logging via `internal/shared/observability` (zerolog), `request_id` correlation, unrecoverable-error logging at every handler boundary (FR-017); per-request access logging for every outcome via `accessLog` middleware (FR-018) |
| XII. Security                  | ✅ Pass | Argon2id, RS256, rate limiting, secure tokens                                                                                                                                                                                                                                     |

**Gate Status**: ✅ PASSED - All applicable principles satisfied

## Project Structure

### Documentation (this feature)

```text
specs/004-auth-module/
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
│   └── auth/
│       ├── domain/
│       │   ├── user.go           # User entity
│       │   ├── session.go        # Session entity (refresh tokens)
│       │   ├── role.go           # Role enum and permissions
│       │   ├── auth_event.go     # Audit log entity
│       │   └── errors.go         # Domain errors
│       ├── repository/
│       │   ├── user_repo.go           # User repository (interface + GORM impl)
│       │   ├── session_repo.go        # Session repository (interface + GORM impl)
│       │   ├── event_repo.go          # Auth event repository (interface + GORM impl)
│       │   ├── tenant_repo.go         # Tenant repository (interface + GORM impl)
│       │   ├── password_reset_repo.go # Password reset repository (interface + GORM impl)
│       │   └── repository_test.go     # Repository tests (SQLite in-memory)
│       ├── service/
│       │   ├── auth_service.go   # Login, logout, refresh
│       │   ├── user_service.go   # User CRUD, password reset
│       │   ├── token_service.go  # JWT generation/validation
│       │   └── password.go       # Argon2id hashing
│       ├── handler/
│       │   ├── auth_handler.go   # REST endpoints
│       │   └── middleware.go     # Auth middleware
│       ├── events.go             # Domain events (UserCreated, LoginSucceeded, etc.)
│       └── service.go            # Public module interface
├── migrations/
│   ├── 001_auth_tables.up.sql
│   └── 001_auth_tables.down.sql
└── pkg/
    └── jwt/
        ├── jwt.go                # JWT utilities (RS256)
        └── keys.go               # Key management
```

**Structure Decision**: Following constitution's modular monolith pattern with `internal/auth` as the bounded context. JWT utilities in `pkg/` for potential reuse by other modules. GORM implementations co-located with interfaces (no separate `postgres/` directory) for simplicity.

## Design: Request/Response Access Logging (FR-018)

Closes the gap where expected 4xx auth rejections (revoked token, wrong password, locked
account, insufficient role) produced zero structured log output - only chi's bare unstructured
access line, no `user_id`/`tenant_id`/reason.

- **Where**: root chi router in `cmd/server/main.go`, replacing `middleware.Logger` - so it
  covers every route mounted on the router, not just auth, as future modules mount onto the
  same router.
- **Status capture**: `chi/middleware.NewWrapResponseWriter` (already an indirect dependency via
  chi) wraps the `ResponseWriter` to expose `.Status()` after the handler returns - no changes to
  any handler needed.
- **User/tenant capture**: `RequireAuth` runs _inside_ `next.ServeHTTP`, after this middleware's
  own `r` variable is fixed - a later `r.Context()` read in the outer middleware would miss
  values `RequireAuth` attaches via `r.WithContext`. Fix: inject a single mutable
  `*accessLogFields` pointer into the context _before_ calling `next`; `RequireAuth` (and any
  future auth-aware middleware) sets `UserID`/`TenantID` fields on that same pointer instead of
  creating a new context value. The outer middleware reads the pointer after `next.ServeHTTP`
  returns and sees whatever was set.
- **One line per request**, level `info` for 2xx/4xx and `error` for 5xx (5xx already gets its
  own detailed entry from `writeInternalError`/FR-017 - this line is the summary, not a
  duplicate): `request_id`, `method`, `path`, `status`, `duration_ms`, `user_id`/`tenant_id` when
  set. Never email, password, or token values (Sensitive Data Handling, Principle XII).

## Complexity Tracking

No violations - straightforward module following established patterns.
