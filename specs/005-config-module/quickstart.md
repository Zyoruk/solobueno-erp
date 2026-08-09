# Quickstart: Configuration Module

**Feature**: 005-config-module
**Audience**: Developers integrating with the config module

## Overview

The config module manages per-tenant branding, business settings, and feature flags, backed
by PostgreSQL with a Redis read-through cache (60s TTL). Requires a logged-in session from
004-auth-module — every endpoint needs `Authorization: Bearer <access_token>`.

## Prerequisites

- `make docker-up` running (PostgreSQL + Redis containers)
- A logged-in access token for an owner-role user in some tenant (see
  `specs/004-auth-module/quickstart.md` for login)

## Quick Start

### 1. Read current config

```bash
curl -X GET http://localhost:8080/api/v1/config \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

First call for a tenant with no `tenant_configs` row returns all-default values (FR-006) —
no 404, config always resolves to _something_.

### 2. Update branding (owner only)

```bash
curl -X PATCH http://localhost:8080/api/v1/config/branding \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "Acme Diner", "primary_color": "#ff6600"}'
```

### 3. Toggle a feature flag (manager+)

```bash
curl -X PATCH http://localhost:8080/api/v1/config/features/reservations \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"enabled": true}'
```

### 4. Verify propagation window (SC-001)

Re-`GET /config` immediately after step 2 or 3 — the Redis cache may still serve the
pre-change value for up to 60 seconds (FR-005's own stated bound), not because it's stale
data being returned incorrectly. Wait up to 60s and re-fetch to see the new value if testing
this by hand.

## Validation Scenarios (maps to spec.md Acceptance Scenarios)

| Scenario                         | Command                                                                       | Expected                                                      |
| -------------------------------- | ----------------------------------------------------------------------------- | ------------------------------------------------------------- |
| US1/AS1: rename tenant           | `PATCH /config/branding {"name": "..."}` then `GET /config` (after cache TTL) | New name in response                                          |
| US1/AS3: brand colors            | `PATCH /config/branding {"primary_color": "#..."}`                            | 200, echoes updated color; invalid hex → 400 `invalid_config` |
| US2/AS1: disabled feature hidden | `GET /config/features` with a flag never toggled on                           | `false` in response (default)                                 |
| US2/AS2: manager enables feature | `PATCH /config/features/{key} {"enabled": true}` as manager                   | 200; `GET /config/features` reflects it after TTL             |
| US3/AS1: currency formatting     | `PATCH /config/business-settings {"currency": "CRC"}`                         | 200; invalid code → 400                                       |
| US3/AS3: tax rate                | `PATCH /config/business-settings {"tax_rate": 13.0}`                          | 200; out-of-range (e.g. `150`) → 400                          |
| Edge case: global override       | Global flag disabled, tenant `PATCH .../features/{key} {"enabled": true}`     | 409 `globally_disabled`                                       |

## Out of Scope

US4 (system-admin global settings management) has no write endpoint in this pass — see
`research.md` for why. `global_configs` rows currently require direct DB access to seed.
