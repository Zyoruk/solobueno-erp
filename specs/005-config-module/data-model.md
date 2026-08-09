# Data Model: Configuration Module

**Feature**: 005-config-module
**Date**: 2026-08-09

## Overview

This document defines the database schema and domain entities for the configuration module.
Storage-shape rationale is in [research.md](./research.md).

## Database Schema

### Table: tenant_configs

One row per tenant. Branding and business settings are JSONB so new fields don't need a
migration (FR-003, FR-004, FR-006, FR-007).

```sql
CREATE TABLE tenant_configs (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           UUID NOT NULL UNIQUE REFERENCES tenants(id),
    branding            JSONB NOT NULL DEFAULT '{}',
    business_settings   JSONB NOT NULL DEFAULT '{}',
    updated_by          UUID NOT NULL REFERENCES users(id),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_tenant_configs_tenant ON tenant_configs(tenant_id);
```

`branding` shape (FR-003): `{"name": string, "logo_url": string, "primary_color": "#rrggbb",
"secondary_color": "#rrggbb"}`. Any key absent falls back to the default in application code
(FR-006) — never written as a stored default.

`business_settings` shape (FR-004): `{"currency": "ISO 4217", "timezone": "IANA tz name",
"tax_rate": number 0-100, "service_charge": number 0-100}`.

### Table: feature_flags

Many rows per tenant, plus global rows (`tenant_id IS NULL`) for FR-009. Resolution reads
the tenant-scoped row first and falls back to the global row (FR-010).

```sql
CREATE TABLE feature_flags (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID REFERENCES tenants(id),  -- NULL = global-scope row
    key         VARCHAR(100) NOT NULL,
    enabled     BOOLEAN NOT NULL DEFAULT false,
    updated_by  UUID NOT NULL REFERENCES users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Partial unique indexes: one row per (tenant, key), one row per (NULL, key)
CREATE UNIQUE INDEX idx_feature_flags_tenant_key ON feature_flags(tenant_id, key) WHERE tenant_id IS NOT NULL;
CREATE UNIQUE INDEX idx_feature_flags_global_key ON feature_flags(key) WHERE tenant_id IS NULL;
```

Resolution rule (FR-010, "Given a global feature is disabled, When any tenant tries to
enable it, Then it remains unavailable" — Edge Case, US4 acceptance scenario 2): effective
value = tenant row's `enabled` AND global row's `enabled` (both default `false`/absent =
disabled). A tenant can never enable a globally-disabled feature; a tenant can disable a
globally-enabled one for itself.

### Table: global_configs

Platform-wide key/value settings outside the feature-flag model (e.g. maintenance window).
Write path is out of scope for this pass (see research.md, US4 descoped) — table and reads
exist because feature-flag resolution needs the read side regardless.

```sql
CREATE TABLE global_configs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key         VARCHAR(100) NOT NULL UNIQUE,
    value       JSONB NOT NULL,
    updated_by  UUID NOT NULL REFERENCES users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### Table: config_changes

Append-only audit log (FR-008, SC-006). Never updated or deleted.

```sql
CREATE TABLE config_changes (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    UUID REFERENCES tenants(id),  -- NULL for global_configs changes
    entity_type  VARCHAR(50) NOT NULL,          -- 'branding' | 'business_settings' | 'feature_flag' | 'global_config'
    entity_key   VARCHAR(100),                  -- feature/global-config key, NULL for branding/business_settings
    old_value    JSONB,
    new_value    JSONB NOT NULL,
    changed_by   UUID NOT NULL REFERENCES users(id),
    changed_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_config_changes_tenant ON config_changes(tenant_id, changed_at DESC);
```

## Domain Entities (Go)

- `TenantConfig` — mirrors `tenant_configs`; `Branding`/`BusinessSettings` as typed structs
  (not raw `map[string]any`) that (de)serialize to/from the JSONB columns via GORM, so
  validation (FR-007) operates on typed fields.
- `BrandingSettings` — `Name`, `LogoURL`, `PrimaryColor`, `SecondaryColor` (all optional;
  zero value = "use default").
- `BusinessSettings` — `Currency`, `Timezone`, `TaxRate`, `ServiceCharge`.
- `FeatureFlag` — mirrors `feature_flags`; `TenantID *uuid.UUID` (nil = global).
- `GlobalConfig` — mirrors `global_configs`, key/value only, read path only in this pass.
- `ConfigChange` — mirrors `config_changes`, audit entity, write-only from the service layer
  (never mutated after insert, matching auth's `AuthEvent` precedent).

## Cache Keys (Redis, research.md)

- `config:tenant:{tenant_id}` → serialized `TenantConfig`, 60s TTL.
- `config:features:{tenant_id}` → serialized effective feature-flag map (tenant ∩ global),
  60s TTL.

## Relationships

- `tenant_configs.tenant_id` → `tenants.id` (1:1, same `tenants` table auth already owns).
- `feature_flags.tenant_id` → `tenants.id` (N:1, nullable).
- `config_changes.tenant_id` → `tenants.id` (N:1, nullable).
- All `updated_by`/`changed_by` → `users.id` (auth's existing table).

No new tenant/user tables — config module is a pure consumer of auth's `tenants`/`users`,
per Constitution II's "modules communicate via Go interfaces... not HTTP/RPC" (this module
reads `tenants.id` via a foreign key at the DB level, and would call auth's repository
interfaces, not HTTP, for anything requiring live tenant/user data).
