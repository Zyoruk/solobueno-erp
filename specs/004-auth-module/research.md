# Research: Authentication Module

**Feature**: 004-auth-module  
**Date**: 2026-01-29

## Research Tasks

### 1. Password Hashing: Argon2id Configuration

**Decision**: Use Argon2id with recommended OWASP parameters

**Rationale**:

- Argon2id is the winner of the Password Hashing Competition
- Resistant to both GPU and side-channel attacks
- Native Go implementation available (alexedwards/argon2id)

**Configuration**:

| Parameter   | Value    | Rationale                          |
| ----------- | -------- | ---------------------------------- |
| Memory      | 64 MB    | OWASP recommendation for 2023+     |
| Iterations  | 3        | Balance between security and speed |
| Parallelism | 4        | Typical server CPU core count      |
| Salt Length | 16 bytes | Standard recommendation            |
| Key Length  | 32 bytes | 256-bit derived key                |

**Library**: `github.com/alexedwards/argon2id`

**Alternatives Considered**:

- bcrypt: Older, no memory-hardness, limited to 72 bytes input
- scrypt: Good but Argon2id supersedes it

### 2. JWT Implementation: RS256 with Key Rotation

**Decision**: RS256 (RSA-SHA256) asymmetric signing with key rotation support

**Rationale**:

- Public key can be distributed for verification without exposing signing secret
- Supports future microservice extraction where services only need public key
- Industry standard for OAuth2/OIDC systems

**Token Structure**:

```json
{
  "header": {
    "alg": "RS256",
    "typ": "JWT",
    "kid": "key-2026-01"
  },
  "payload": {
    "sub": "user-uuid",
    "iss": "solobueno-erp",
    "aud": ["solobueno-api"],
    "exp": 1234567890,
    "iat": 1234567890,
    "jti": "unique-token-id",
    "tenant_id": "tenant-uuid",
    "role": "manager",
    "email": "user@example.com"
  }
}
```

**Key Management**:

- RSA 2048-bit keys minimum (4096 recommended for long-lived keys)
- Keys stored in environment/secrets manager (AWS SSM Parameter Store)
- Key ID (`kid`) in header for rotation support
- Refresh tokens are opaque (UUID), not JWT

**Library**: `github.com/golang-jwt/jwt/v5`

**Alternatives Considered**:

- HS256: Simpler but secret must be shared with all verifiers
- EdDSA: Modern but less library support

### 3. Session Storage: Database-Backed Refresh Tokens

**Decision**: Store refresh tokens in PostgreSQL for immediate revocation capability

**Rationale**:

- FR-014 requires invalidating all sessions on password change
- Logout must immediately invalidate tokens
- Enables "logged in devices" feature in future
- Access tokens remain stateless (short-lived, not stored)

**Session Table Design**:

| Column        | Type        | Description                    |
| ------------- | ----------- | ------------------------------ |
| id            | UUID        | Primary key                    |
| user_id       | UUID        | FK to users                    |
| tenant_id     | UUID        | FK to tenants                  |
| refresh_token | VARCHAR     | Hashed token (not plain!)      |
| device_info   | VARCHAR     | User agent / device identifier |
| ip_address    | INET        | Last known IP                  |
| created_at    | TIMESTAMPTZ | Session creation time          |
| expires_at    | TIMESTAMPTZ | 30 days from creation          |
| revoked_at    | TIMESTAMPTZ | NULL if active, set on logout  |

**Alternatives Considered**:

- Redis: Fast but adds infrastructure; PostgreSQL sufficient for initial scale
- Stateless only: Cannot support immediate revocation

### 4. Rate Limiting: In-Memory with Redis Upgrade Path

**Decision**: Start with in-memory rate limiter, abstract for Redis upgrade

**Rationale**:

- FR-011 requires 5 requests/minute/IP for login
- Single server initially, in-memory sufficient
- Abstract interface allows Redis swap when scaling

**Implementation**:

- Sliding window algorithm
- Key: `login:ip:{ip_address}`
- Window: 60 seconds
- Limit: 5 attempts

**Library**: `golang.org/x/time/rate` (stdlib compatible) or custom sliding window

**Alternatives Considered**:

- Redis from start: Overkill for single server
- No abstraction: Harder to scale later

### 5. Multi-Tenant User Model

**Decision**: Users have globally unique emails, UserTenantRole junction table for roles

**Rationale**:

- One person may work at multiple restaurants
- Same email, different roles per tenant
- Simplifies password reset (single account)

**Data Model**:

```
users (globally unique email)
    └── user_tenant_roles (user_id, tenant_id, role)
            └── defines role per tenant
```

**Token Behavior**:

- Login requires tenant selection if user has multiple tenants
- Token contains single tenant_id for current session
- Switching tenants requires new token

### 6. Password Reset Flow

**Decision**: Time-limited secure token via email

**Rationale**:

- FR-013 requires password reset
- Token-based is stateless and secure
- Email delivery via existing AWS SES

**Flow**:

1. User requests reset with email
2. System generates cryptographically secure token (32 bytes, base64url encoded)
3. Token stored hashed in DB with 1-hour expiry
4. Email sent with reset link containing token
5. User submits new password with token
6. Token validated, password updated, all sessions invalidated (FR-014)

**Security**:

- Token hashed before storage (prevent DB leak exploitation)
- One-time use (deleted after use)
- Rate limited (1 request per email per 5 minutes)

### 7. Role-Based Access Control (RBAC)

**Decision**: Simple role-based permissions, middleware enforcement

**Rationale**:

- FR-007 defines 7 roles with clear hierarchy
- Keep simple for MVP, can evolve to fine-grained permissions later

**Role Hierarchy**:

| Role    | Level | Can Manage         | Typical Actions                    |
| ------- | ----- | ------------------ | ---------------------------------- |
| owner   | 100   | Everything         | Billing, tenant settings           |
| admin   | 90    | All except billing | User management, all operations    |
| manager | 70    | Staff, menu, hours | Create users, update menu, reports |
| cashier | 50    | Payments           | Process payments, view orders      |
| waiter  | 40    | Orders, tables     | Take orders, manage tables         |
| kitchen | 30    | Kitchen display    | View/update order status           |
| viewer  | 10    | Nothing            | Read-only access                   |

**Middleware**:

- Extract token from Authorization header
- Validate signature and expiry
- Check role against endpoint requirements
- Inject user context into request

## Summary of Decisions

| Area             | Decision                                             |
| ---------------- | ---------------------------------------------------- |
| Password Hashing | Argon2id (64MB, 3 iterations, 4 parallelism)         |
| JWT Signing      | RS256 with key rotation support                      |
| Session Storage  | PostgreSQL (refresh tokens stored, access stateless) |
| Rate Limiting    | In-memory sliding window, Redis-ready interface      |
| User Model       | Global email, UserTenantRole junction                |
| Password Reset   | Hashed time-limited token via email                  |
| RBAC             | 7-role hierarchy with middleware enforcement         |

## Open Questions Resolved

All technical decisions made. No outstanding questions.
