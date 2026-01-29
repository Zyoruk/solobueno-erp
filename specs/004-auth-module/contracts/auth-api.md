# Auth API Contracts

**Module**: 004-auth-module  
**Base Path**: `/api/v1/auth`

## Endpoints

### POST /login

Authenticate user and return token pair.

**Request:**

```json
{
  "email": "string (required)",
  "password": "string (required)",
  "tenant_id": "uuid (optional, required if user has multiple tenants)"
}
```

**Response 200:**

```json
{
  "access_token": "string (JWT)",
  "refresh_token": "string (opaque)",
  "token_type": "Bearer",
  "expires_in": 3600,
  "expires_at": "2026-01-29T16:00:00Z",
  "user": {
    "id": "uuid",
    "email": "string",
    "first_name": "string",
    "last_name": "string",
    "role": "string (enum: owner|admin|manager|cashier|waiter|kitchen|viewer)",
    "tenant_id": "uuid"
  }
}
```

**Response 400 (Tenant Required):**

```json
{
  "error": {
    "code": "tenant_required",
    "message": "User belongs to multiple tenants. Please specify tenant_id.",
    "tenants": [{ "id": "uuid", "name": "string", "slug": "string" }]
  }
}
```

**Response 401:**

```json
{
  "error": {
    "code": "invalid_credentials | account_disabled",
    "message": "string"
  }
}
```

**Response 429:**

```json
{
  "error": {
    "code": "rate_limit_exceeded",
    "message": "Too many login attempts. Try again in X seconds.",
    "retry_after": 60
  }
}
```

---

### POST /refresh

Refresh access token using refresh token.

**Request:**

```json
{
  "refresh_token": "string (required)"
}
```

**Response 200:**

```json
{
  "access_token": "string (JWT)",
  "refresh_token": "string (opaque, new token)",
  "token_type": "Bearer",
  "expires_in": 3600,
  "expires_at": "2026-01-29T16:00:00Z"
}
```

**Response 401:**

```json
{
  "error": {
    "code": "token_invalid | token_expired | session_revoked",
    "message": "string"
  }
}
```

---

### POST /logout

Invalidate current session.

**Headers:**

```
Authorization: Bearer <access_token>
```

**Response 204:** No content

**Response 401:**

```json
{
  "error": {
    "code": "token_invalid",
    "message": "string"
  }
}
```

---

### POST /password-reset/request

Request password reset email.

**Request:**

```json
{
  "email": "string (required)"
}
```

**Response 202:**

```json
{
  "message": "If the email exists, a reset link has been sent."
}
```

Note: Always returns 202 to prevent email enumeration.

**Response 429:**

```json
{
  "error": {
    "code": "rate_limit_exceeded",
    "message": "Please wait before requesting another reset.",
    "retry_after": 300
  }
}
```

---

### POST /password-reset/complete

Complete password reset with token.

**Request:**

```json
{
  "token": "string (required, from email)",
  "new_password": "string (required, min 8 chars)"
}
```

**Response 200:**

```json
{
  "message": "Password has been reset successfully."
}
```

**Response 400:**

```json
{
  "error": {
    "code": "token_invalid | token_expired | password_weak",
    "message": "string"
  }
}
```

---

### GET /me

Get current user info.

**Headers:**

```
Authorization: Bearer <access_token>
```

**Response 200:**

```json
{
  "id": "uuid",
  "email": "string",
  "first_name": "string",
  "last_name": "string",
  "role": "string",
  "tenant_id": "uuid",
  "tenant_name": "string",
  "must_reset_password": false,
  "tenants": [{ "id": "uuid", "name": "string", "role": "string" }]
}
```

---

### POST /change-password

Change current user's password.

**Headers:**

```
Authorization: Bearer <access_token>
```

**Request:**

```json
{
  "current_password": "string (required)",
  "new_password": "string (required, min 8 chars)"
}
```

**Response 200:**

```json
{
  "message": "Password changed successfully. All other sessions have been invalidated."
}
```

**Response 400:**

```json
{
  "error": {
    "code": "current_password_incorrect | password_weak",
    "message": "string"
  }
}
```

---

## User Management Endpoints

**Base Path**: `/api/v1/users`

### POST /users

Create new user (Manager+).

**Headers:**

```
Authorization: Bearer <access_token>
```

**Request:**

```json
{
  "email": "string (required)",
  "first_name": "string (required)",
  "last_name": "string (required)",
  "role": "string (required, enum: manager|cashier|waiter|kitchen|viewer)"
}
```

Note: Cannot create users with role higher than your own.

**Response 201:**

```json
{
  "id": "uuid",
  "email": "string",
  "first_name": "string",
  "last_name": "string",
  "role": "string",
  "temporary_password": "string (display once)",
  "must_reset_password": true,
  "created_at": "2026-01-29T12:00:00Z"
}
```

**Response 400:**

```json
{
  "error": {
    "code": "email_exists | invalid_role",
    "message": "string"
  }
}
```

**Response 403:**

```json
{
  "error": {
    "code": "insufficient_role",
    "message": "Cannot create users with role equal or higher than your own."
  }
}
```

---

### GET /users

List users in current tenant (Manager+).

**Headers:**

```
Authorization: Bearer <access_token>
```

**Query Parameters:**

- `role`: Filter by role
- `active`: Filter by active status (true/false)
- `page`: Page number (default: 1)
- `limit`: Items per page (default: 20, max: 100)

**Response 200:**

```json
{
  "data": [
    {
      "id": "uuid",
      "email": "string",
      "first_name": "string",
      "last_name": "string",
      "role": "string",
      "is_active": true,
      "created_at": "2026-01-29T12:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 45,
    "total_pages": 3
  }
}
```

---

### GET /users/{id}

Get user details (Manager+).

**Response 200:**

```json
{
  "id": "uuid",
  "email": "string",
  "first_name": "string",
  "last_name": "string",
  "role": "string",
  "is_active": true,
  "must_reset_password": false,
  "created_at": "2026-01-29T12:00:00Z",
  "updated_at": "2026-01-29T12:00:00Z"
}
```

---

### PATCH /users/{id}

Update user details (Manager+).

**Request:**

```json
{
  "first_name": "string (optional)",
  "last_name": "string (optional)",
  "is_active": "boolean (optional)"
}
```

**Response 200:** Updated user object

---

### PATCH /users/{id}/role

Change user role (Manager+, cannot promote to equal or higher role).

**Request:**

```json
{
  "role": "string (required)"
}
```

**Response 200:** Updated user object

**Response 403:**

```json
{
  "error": {
    "code": "insufficient_role",
    "message": "Cannot assign role equal or higher than your own."
  }
}
```

---

## Common Error Codes

| Code                       | HTTP | Description                                  |
| -------------------------- | ---- | -------------------------------------------- |
| invalid_credentials        | 401  | Email or password incorrect                  |
| account_disabled           | 401  | User account is disabled                     |
| token_invalid              | 401  | JWT signature invalid or malformed           |
| token_expired              | 401  | JWT has expired                              |
| session_revoked            | 401  | Refresh token was revoked                    |
| tenant_required            | 400  | Must specify tenant_id for multi-tenant user |
| insufficient_role          | 403  | User role cannot perform this action         |
| rate_limit_exceeded        | 429  | Too many requests                            |
| email_exists               | 400  | Email already registered                     |
| password_weak              | 400  | Password doesn't meet requirements           |
| current_password_incorrect | 400  | Current password verification failed         |

## JWT Claims

```json
{
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
```
