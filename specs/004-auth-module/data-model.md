# Data Model: Authentication Module

**Feature**: 004-auth-module  
**Date**: 2026-01-29

## Overview

This document defines the database schema and domain entities for the authentication module.

## Database Schema

### Table: users

Core user entity with globally unique email.

```sql
CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email           VARCHAR(255) NOT NULL UNIQUE,
    password_hash   VARCHAR(255) NOT NULL,
    first_name      VARCHAR(100) NOT NULL,
    last_name       VARCHAR(100) NOT NULL,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    must_reset_pwd  BOOLEAN NOT NULL DEFAULT false,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_active ON users(is_active) WHERE is_active = true;
```

### Table: tenants

Restaurant/business entity (minimal for auth context).

```sql
CREATE TABLE tenants (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(255) NOT NULL,
    slug            VARCHAR(100) NOT NULL UNIQUE,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_tenants_slug ON tenants(slug);
```

### Table: user_tenant_roles

Junction table mapping users to tenants with roles.

```sql
CREATE TYPE user_role AS ENUM (
    'owner', 'admin', 'manager', 'cashier', 'waiter', 'kitchen', 'viewer'
);

CREATE TABLE user_tenant_roles (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    role            user_role NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(user_id, tenant_id)
);

CREATE INDEX idx_utr_user ON user_tenant_roles(user_id);
CREATE INDEX idx_utr_tenant ON user_tenant_roles(tenant_id);
CREATE INDEX idx_utr_role ON user_tenant_roles(role);
```

### Table: sessions

Active refresh tokens for revocation support.

```sql
CREATE TABLE sessions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    refresh_token   VARCHAR(255) NOT NULL UNIQUE,  -- Hashed token
    device_info     VARCHAR(500),
    ip_address      INET,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ NOT NULL,
    revoked_at      TIMESTAMPTZ,

    CONSTRAINT valid_expiry CHECK (expires_at > created_at)
);

CREATE INDEX idx_sessions_user ON sessions(user_id);
CREATE INDEX idx_sessions_tenant ON sessions(tenant_id);
CREATE INDEX idx_sessions_token ON sessions(refresh_token);
CREATE INDEX idx_sessions_active ON sessions(user_id, tenant_id)
    WHERE revoked_at IS NULL AND expires_at > NOW();
```

### Table: password_reset_tokens

Time-limited tokens for password reset flow.

```sql
CREATE TABLE password_reset_tokens (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash      VARCHAR(255) NOT NULL UNIQUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ NOT NULL,
    used_at         TIMESTAMPTZ,

    CONSTRAINT valid_expiry CHECK (expires_at > created_at)
);

CREATE INDEX idx_prt_user ON password_reset_tokens(user_id);
CREATE INDEX idx_prt_token ON password_reset_tokens(token_hash);
```

### Table: auth_events

Audit log for all authentication actions.

```sql
CREATE TYPE auth_event_type AS ENUM (
    'login_success', 'login_failed', 'logout', 'token_refresh',
    'password_changed', 'password_reset_requested', 'password_reset_completed',
    'account_created', 'account_disabled', 'account_enabled',
    'role_changed', 'session_revoked'
);

CREATE TABLE auth_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID REFERENCES users(id) ON DELETE SET NULL,
    tenant_id       UUID REFERENCES tenants(id) ON DELETE SET NULL,
    event_type      auth_event_type NOT NULL,
    ip_address      INET,
    user_agent      VARCHAR(500),
    metadata        JSONB,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_auth_events_user ON auth_events(user_id);
CREATE INDEX idx_auth_events_tenant ON auth_events(tenant_id);
CREATE INDEX idx_auth_events_type ON auth_events(event_type);
CREATE INDEX idx_auth_events_created ON auth_events(created_at DESC);
```

## Domain Entities (Go)

### User

```go
type User struct {
    ID           uuid.UUID  `json:"id"`
    Email        string     `json:"email"`
    PasswordHash string     `json:"-"` // Never serialize
    FirstName    string     `json:"first_name"`
    LastName     string     `json:"last_name"`
    IsActive     bool       `json:"is_active"`
    MustResetPwd bool       `json:"must_reset_password"`
    CreatedAt    time.Time  `json:"created_at"`
    UpdatedAt    time.Time  `json:"updated_at"`
}
```

### Role

```go
type Role string

const (
    RoleOwner   Role = "owner"
    RoleAdmin   Role = "admin"
    RoleManager Role = "manager"
    RoleCashier Role = "cashier"
    RoleWaiter  Role = "waiter"
    RoleKitchen Role = "kitchen"
    RoleViewer  Role = "viewer"
)

func (r Role) Level() int {
    levels := map[Role]int{
        RoleOwner:   100,
        RoleAdmin:   90,
        RoleManager: 70,
        RoleCashier: 50,
        RoleWaiter:  40,
        RoleKitchen: 30,
        RoleViewer:  10,
    }
    return levels[r]
}

func (r Role) CanManage(other Role) bool {
    return r.Level() > other.Level()
}
```

### UserTenantRole

```go
type UserTenantRole struct {
    ID        uuid.UUID `json:"id"`
    UserID    uuid.UUID `json:"user_id"`
    TenantID  uuid.UUID `json:"tenant_id"`
    Role      Role      `json:"role"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

### Session

```go
type Session struct {
    ID           uuid.UUID  `json:"id"`
    UserID       uuid.UUID  `json:"user_id"`
    TenantID     uuid.UUID  `json:"tenant_id"`
    RefreshToken string     `json:"-"` // Hashed, never expose
    DeviceInfo   string     `json:"device_info,omitempty"`
    IPAddress    string     `json:"ip_address,omitempty"`
    CreatedAt    time.Time  `json:"created_at"`
    ExpiresAt    time.Time  `json:"expires_at"`
    RevokedAt    *time.Time `json:"revoked_at,omitempty"`
}

func (s *Session) IsValid() bool {
    return s.RevokedAt == nil && time.Now().Before(s.ExpiresAt)
}
```

### AuthEvent

```go
type AuthEventType string

const (
    EventLoginSuccess           AuthEventType = "login_success"
    EventLoginFailed            AuthEventType = "login_failed"
    EventLogout                 AuthEventType = "logout"
    EventTokenRefresh           AuthEventType = "token_refresh"
    EventPasswordChanged        AuthEventType = "password_changed"
    EventPasswordResetRequested AuthEventType = "password_reset_requested"
    EventPasswordResetCompleted AuthEventType = "password_reset_completed"
    EventAccountCreated         AuthEventType = "account_created"
    EventAccountDisabled        AuthEventType = "account_disabled"
    EventAccountEnabled         AuthEventType = "account_enabled"
    EventRoleChanged            AuthEventType = "role_changed"
    EventSessionRevoked         AuthEventType = "session_revoked"
)

type AuthEvent struct {
    ID        uuid.UUID              `json:"id"`
    UserID    *uuid.UUID             `json:"user_id,omitempty"`
    TenantID  *uuid.UUID             `json:"tenant_id,omitempty"`
    EventType AuthEventType          `json:"event_type"`
    IPAddress string                 `json:"ip_address,omitempty"`
    UserAgent string                 `json:"user_agent,omitempty"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
    CreatedAt time.Time              `json:"created_at"`
}
```

### TokenPair

```go
type TokenPair struct {
    AccessToken  string    `json:"access_token"`
    RefreshToken string    `json:"refresh_token"`
    TokenType    string    `json:"token_type"` // "Bearer"
    ExpiresIn    int       `json:"expires_in"` // Seconds until access token expires
    ExpiresAt    time.Time `json:"expires_at"`
}
```

### Claims (JWT Payload)

```go
type Claims struct {
    jwt.RegisteredClaims
    TenantID uuid.UUID `json:"tenant_id"`
    Role     Role      `json:"role"`
    Email    string    `json:"email"`
}
```

## Entity Relationships

```
┌─────────┐       ┌──────────────────┐       ┌─────────┐
│  users  │──1:N──│ user_tenant_roles│──N:1──│ tenants │
└─────────┘       └──────────────────┘       └─────────┘
     │                                            │
     │ 1:N                                        │ 1:N
     ▼                                            ▼
┌──────────┐                                ┌──────────┐
│ sessions │                                │ sessions │
└──────────┘                                └──────────┘
     │
     │ 1:N
     ▼
┌─────────────┐
│ auth_events │
└─────────────┘
```

## Indexes Summary

| Table                 | Index                    | Purpose                       |
| --------------------- | ------------------------ | ----------------------------- |
| users                 | email (unique)           | Login lookup                  |
| users                 | is_active (partial)      | Active user queries           |
| tenants               | slug (unique)            | Tenant lookup by slug         |
| user_tenant_roles     | user_id                  | User's tenants lookup         |
| user_tenant_roles     | tenant_id                | Tenant's users lookup         |
| user_tenant_roles     | (user_id, tenant_id) UK  | Prevent duplicate assignments |
| sessions              | refresh_token (unique)   | Token validation              |
| sessions              | user_id, tenant_id       | Active sessions lookup        |
| password_reset_tokens | token_hash (unique)      | Token validation              |
| auth_events           | user_id, tenant_id, type | Audit queries                 |
| auth_events           | created_at DESC          | Recent events                 |
