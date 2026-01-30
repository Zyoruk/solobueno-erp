# Quickstart: Authentication Module

**Feature**: 004-auth-module  
**Audience**: Developers integrating with the auth module

## Overview

The authentication module provides JWT-based authentication with RS256 signing, Argon2id password hashing, and role-based access control for multi-tenant users.

## Quick Start

### 1. Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "secret123",
    "tenant_id": "550e8400-e29b-41d4-a716-446655440000"
  }'
```

**Response:**

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIs...",
  "refresh_token": "dGhpcyBpcyBhIHJlZnJlc2ggdG9rZW4...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "expires_at": "2026-01-29T16:00:00Z"
}
```

### 2. Access Protected Endpoint

```bash
curl -X GET http://localhost:8080/api/v1/orders \
  -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIs..."
```

### 3. Refresh Token

```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "dGhpcyBpcyBhIHJlZnJlc2ggdG9rZW4..."
  }'
```

### 4. Logout

```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIs..."
```

## Authentication Flows

### Login Flow

```
┌──────┐      ┌──────────┐      ┌────────┐      ┌──────────┐
│Client│      │Auth API  │      │Database│      │TokenSvc  │
└──┬───┘      └────┬─────┘      └───┬────┘      └────┬─────┘
   │   POST /login │                │                │
   │──────────────>│                │                │
   │               │ Find user      │                │
   │               │───────────────>│                │
   │               │   User data    │                │
   │               │<───────────────│                │
   │               │ Verify password│                │
   │               │ (Argon2id)     │                │
   │               │                │                │
   │               │ Generate tokens│                │
   │               │───────────────────────────────>│
   │               │ Access + Refresh               │
   │               │<───────────────────────────────│
   │               │ Store session  │                │
   │               │───────────────>│                │
   │               │                │                │
   │  TokenPair    │                │                │
   │<──────────────│                │                │
```

### Token Refresh Flow

```
┌──────┐      ┌──────────┐      ┌────────┐      ┌──────────┐
│Client│      │Auth API  │      │Database│      │TokenSvc  │
└──┬───┘      └────┬─────┘      └───┬────┘      └────┬─────┘
   │  POST /refresh│                │                │
   │──────────────>│                │                │
   │               │ Find session   │                │
   │               │───────────────>│                │
   │               │ Session data   │                │
   │               │<───────────────│                │
   │               │ Validate token │                │
   │               │ Check not revoked               │
   │               │                │                │
   │               │ Generate new tokens             │
   │               │───────────────────────────────>│
   │               │<───────────────────────────────│
   │               │ Update session │                │
   │               │───────────────>│                │
   │  New TokenPair│                │                │
   │<──────────────│                │                │
```

## Role-Based Access

### Role Hierarchy

| Role    | Level | Capabilities                     |
| ------- | ----- | -------------------------------- |
| owner   | 100   | Everything including billing     |
| admin   | 90    | All except billing settings      |
| manager | 70    | Staff, menu, reports, operations |
| cashier | 50    | Process payments, view orders    |
| waiter  | 40    | Create/manage orders, tables     |
| kitchen | 30    | View/update order status         |
| viewer  | 10    | Read-only access                 |

### Checking Permissions

The middleware automatically:

1. Extracts JWT from `Authorization: Bearer <token>` header
2. Validates signature and expiry
3. Injects user context into request
4. Checks role against endpoint requirements

```go
// Example: Endpoint requires manager or higher
router.With(auth.RequireRole(auth.RoleManager)).
    Post("/staff", handler.CreateStaff)
```

## Error Responses

| Code | Error               | Description                      |
| ---- | ------------------- | -------------------------------- |
| 401  | invalid_credentials | Email/password incorrect         |
| 401  | account_disabled    | User account is disabled         |
| 401  | token_expired       | Access token has expired         |
| 401  | token_invalid       | Token signature invalid          |
| 401  | session_revoked     | Refresh token was revoked        |
| 403  | insufficient_role   | User role cannot access resource |
| 429  | rate_limit_exceeded | Too many login attempts          |

**Error Response Format:**

```json
{
  "error": {
    "code": "invalid_credentials",
    "message": "Invalid email or password"
  }
}
```

## Multi-Tenant Users

Users can belong to multiple tenants with different roles:

```bash
# Login requires tenant_id
POST /api/v1/auth/login
{
  "email": "user@example.com",
  "password": "secret",
  "tenant_id": "tenant-uuid-here"
}

# If user has multiple tenants and none specified:
{
  "error": {
    "code": "tenant_required",
    "message": "User belongs to multiple tenants",
    "tenants": [
      {"id": "uuid-1", "name": "Restaurant A"},
      {"id": "uuid-2", "name": "Restaurant B"}
    ]
  }
}
```

## Password Reset

### Request Reset

```bash
POST /api/v1/auth/password-reset/request
{
  "email": "user@example.com"
}
```

### Complete Reset

```bash
POST /api/v1/auth/password-reset/complete
{
  "token": "reset-token-from-email",
  "new_password": "newSecurePassword123"
}
```

## User Management (Manager+)

### Create User

```bash
POST /api/v1/users
Authorization: Bearer <manager-token>
{
  "email": "newstaff@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "role": "waiter"
}
```

Response includes temporary password that user must change on first login.

### Update User Role

```bash
PATCH /api/v1/users/{id}/role
Authorization: Bearer <manager-token>
{
  "role": "cashier"
}
```

## Testing Locally

```bash
# Run migrations
make migrate-up

# Start server
make run

# Create test user (via seed or direct SQL)
psql -d solobueno_dev -c "
INSERT INTO users (email, password_hash, first_name, last_name)
VALUES ('test@example.com', '\$argon2id\$...', 'Test', 'User');
"

# Test login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"test123"}'
```

## File Locations

| File                                        | Purpose                               |
| ------------------------------------------- | ------------------------------------- |
| `backend/internal/auth/`                    | Auth module code                      |
| `backend/internal/auth/domain/`             | Domain entities and errors            |
| `backend/internal/auth/repository/`         | Repository interfaces + GORM impl     |
| `backend/internal/auth/service/`            | Business logic services               |
| `backend/internal/auth/handler/`            | HTTP handlers and middleware          |
| `backend/migrations/001_auth_tables.up.sql` | Database schema (users, tenants, etc) |
| `backend/pkg/jwt/`                          | JWT utilities (RS256)                 |

## Running Tests

```bash
# Run all auth tests
cd backend && go test ./internal/auth/... -v

# Run with coverage
cd backend && go test ./internal/auth/... -cover

# Current coverage:
# - domain: 88.8%
# - service: 85.3%
# - handler: 26.9%
# - repository: 14.0%
```
