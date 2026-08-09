# Config API Contracts

**Module**: 005-config-module
**Base Path**: `/api/v1/config`
**Auth**: All endpoints require `Authorization: Bearer <access_token>` (reuses
`internal/auth/handler.AuthMiddleware`). Role gates per endpoint below use the same
`domain.Role` hierarchy auth already defines (owner &gt; admin &gt; manager &gt;
cashier/waiter/kitchen &gt; viewer).

## Endpoints

### GET /config

Returns the authenticated user's tenant's effective configuration (branding + business
settings + resolved feature flags). Cached read, up to 60s stale (FR-005/SC-005).

**Response 200:**

```json
{
  "branding": {
    "name": "string",
    "logo_url": "string",
    "primary_color": "#rrggbb",
    "secondary_color": "#rrggbb"
  },
  "business_settings": {
    "currency": "CRC",
    "timezone": "America/Costa_Rica",
    "tax_rate": 13.0,
    "service_charge": 10.0
  },
  "features": {
    "reservations": true,
    "kitchen_display": false
  },
  "updated_at": "2026-08-09T12:00:00Z"
}
```

Any field omitted from storage returns the system default (FR-006) — the response is always
fully populated, never partial.

### PATCH /config/branding

Update branding. Requires `RoleOwner` (US1).

**Request** (partial update, only included fields change):

```json
{
  "name": "string (optional)",
  "logo_url": "string (optional, URL)",
  "primary_color": "string (optional, #rrggbb)",
  "secondary_color": "string (optional, #rrggbb)"
}
```

**Response 200:** updated `branding` object (same shape as `GET /config`'s `branding`).

**Response 400 (Validation):**

```json
{ "error": { "code": "invalid_config", "message": "primary_color must be a #rrggbb hex color" } }
```

**Response 403:** `{ "error": { "code": "insufficient_role", "message": "..." } }`

### PATCH /config/business-settings

Update business settings. Requires `RoleOwner` (US3).

**Request:**

```json
{
  "currency": "string (optional, ISO 4217)",
  "timezone": "string (optional, IANA tz name)",
  "tax_rate": "number (optional, 0-100)",
  "service_charge": "number (optional, 0-100)"
}
```

**Response 200:** updated `business_settings` object.

**Response 400:** unknown currency code, invalid timezone (fails `time.LoadLocation`), or
tax_rate/service_charge outside `[0, 100]`.

### GET /config/features

List effective feature flags for the caller's tenant (tenant override resolved against
global, per FR-010).

**Response 200:**

```json
{ "features": { "reservations": true, "kitchen_display": false } }
```

### PATCH /config/features/{key}

Toggle a single feature flag for the tenant. Requires `RoleManager` or higher (US2).

**Request:**

```json
{ "enabled": true }
```

**Response 200:**

```json
{ "key": "reservations", "enabled": true, "effective": true }
```

`effective` reflects the resolved value after applying the global override rule — e.g.
`enabled: true` with a globally-disabled flag still resolves `effective: false`.

**Response 409 (Global override):**

```json
{
  "error": {
    "code": "globally_disabled",
    "message": "This feature is disabled platform-wide and cannot be enabled per-tenant"
  }
}
```

## Out of scope (this pass)

No `/admin/global-config` write endpoint ships — see research.md, US4 descoped pending a
platform-admin authorization concept that doesn't exist yet in `domain.Role`.
