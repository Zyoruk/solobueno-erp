# Research: Configuration Module

**Feature**: 005-config-module
**Date**: 2026-08-09

## Decision: API surface stays REST, not GraphQL

**Decision**: Config module exposes REST endpoints under `/api/v1/config`, mirroring
004-auth-module's precedent, not the constitution's GraphQL-primary mandate.

**Rationale**: No GraphQL server exists in the codebase yet (`backend/api/graphql/` is an
empty placeholder, no `gqlgen` dependency). Standing one up is a project-wide undertaking
(schema, codegen, resolver wiring, auth-context plumbing) that every module after this one
would then depend on — far larger than "add config CRUD," and not something to smuggle into
a single module's plan. User confirmed via a scoping question: keep REST consistent with
auth for now; GraphQL migration is deferred to its own future tracked feature. Recorded in
Complexity Tracking below as a Principle III deviation, same as the (undocumented) one auth
already made.

**Alternatives considered**: Standing up gqlgen now — rejected, disproportionate scope for
this module; the constitution's own module list treats `config` as one bounded context among
sixteen, not the one that should also own bootstrapping the primary API layer.

## Decision: Storage shape — one row per tenant + separate flags/audit tables

**Decision**: `tenant_configs` (1:1 with tenant, JSONB `branding`/`business_settings`
columns), `feature_flags` (many rows, `tenant_id` nullable = global-scope row),
`global_configs` (key/value JSONB, platform-wide non-flag settings), `config_changes`
(append-only audit log).

**Rationale**: FR-003/FR-004 enumerate a small, stable field set per tenant (name, logo,
colors, currency, timezone, tax rate, service charge) — a single JSONB blob per concern
avoids a wide flat table or premature per-field columns, while still being one row to
fetch/cache per tenant (matches SC-005's <100ms cached-read target). Feature flags are
naturally many-rows-per-tenant and need a nullable `tenant_id` to represent FR-009's global
scope with FR-010's tenant-overrides-global resolution — one flag table serves both, keyed
by `(tenant_id, key)` with `tenant_id IS NULL` meaning global. `config_changes` is separate
because it's append-only and unbounded, unlike the other three.

**Alternatives considered**: A single wide `tenant_configs` table with individual typed
columns per setting — rejected, every new business setting would need a migration; JSONB
with app-level validation (FR-007) gets the same safety without the schema churn, consistent
with the constitution's own stated PostgreSQL rationale ("JSONB flexibility").

## Decision: Caching — Redis, TTL-only, no invalidation pub/sub

**Decision**: New `shared/cache.Cache` interface (`Get`/`Set`/`Delete`, matching the
adapter-interface pattern `shared/observability` already established) + a Redis adapter via
`github.com/redis/go-redis/v9` (not yet a dependency). Config reads go through a
cache-aside path with a **60-second TTL**, no explicit invalidation on write.

**Rationale**: FR-005 requires refresh within 1 minute of a change — a flat 60s TTL satisfies
that exactly by construction, without needing pub/sub invalidation, cache-tag bookkeeping, or
cross-instance coordination. Redis is already provisioned in `docker-compose.yml`
(`solobueno-redis`) and named for this purpose in the constitution's stack table, but nothing
in the codebase uses it yet — this module is the first consumer.

**Alternatives considered**: Explicit invalidation (`DEL` on write) — would make changes feel
instant instead of eventually-consistent within 60s, but adds a second failure mode (cache
and DB now need to agree on write success) for a requirement the TTL alone already meets.
Deferred; can be added later without changing the interface if 60s proves too slow for some
UI flow.

## Decision: US4 (platform-wide global settings management) descoped from this pass

**Decision**: This plan covers US1 (branding), US2 (feature flags), US3 (business settings) —
all P1/P1/P2. US4 (a system administrator managing global settings across all tenants) is
**not** included; `global_configs` exists as a table and feature-flag resolution already
reads it (FR-010's override rule needs the read side regardless), but no write endpoint or
authorization path ships in this pass.

**Rationale**: `domain.Role` (`internal/auth/domain/role.go`) has no role above per-tenant
`RoleOwner` — there is no existing "platform administrator" concept to gate a
cross-tenant-write endpoint behind, and inventing one is its own RBAC design decision, not a
config-module concern. US4 is already the spec's lowest priority (P3). Building the write
path now would mean either bypassing authorization entirely (unacceptable per Principle XII)
or quietly extending the Role enum with platform-admin semantics inside an unrelated module's
plan.

**Alternatives considered**: Gate it behind an env-var allowlist of user IDs — rejected as a
throwaway hack that would need to be redone once real platform-admin auth exists; better to
leave it undone and visible than half-done and invisible.

## Decision: Domain events emitted, no event bus wiring required

**Decision**: `internal/config/events.go` defines `ConfigChanged`/`FeatureFlagToggled` event
structs, matching auth's `internal/auth/events.go` precedent. No NATS/event-bus wiring is
in scope — same boundary auth already drew (events are defined for future consumers; nothing
in the codebase yet subscribes to any domain event bus, dev uses in-memory per the
constitution's stack table and nothing has stood that up either).

**Rationale**: Consistency with the existing module; other modules (notifications, analytics)
will eventually want to react to config changes, so the event shape should exist even before
a real subscriber does.

## Decision: Validation rules (FR-007)

**Decision**: Currency — ISO 4217 3-letter code, checked against a fixed allowlist (not a
full external library) since business scope is currently CRC/USD-class. Timezone — validated
via Go stdlib `time.LoadLocation` (IANA tz db, already in the Go binary, zero new
dependency). Tax rate / service charge — numeric, `0 <= x <= 100`. Brand colors — `^#[0-9a-fA-F]{6}$`.

**Rationale**: Stdlib (`time.LoadLocation`) and a small allowlist cover this without pulling
in a currency-data package for a handful of supported currencies.
