# Feature Specification: Authentication Module

**Feature Branch**: `004-auth-module`  
**Created**: 2025-01-29  
**Status**: Draft  
**Dependencies**: 002-docker-local-dev

## Clarifications

### Session 2026-01-29

- Q: Password hashing algorithm choice (bcrypt vs argon2)? → A: Argon2id (most secure, modern)
- Q: Email uniqueness scope (global vs per-tenant)? → A: Globally unique (one account per email, can belong to multiple tenants)
- Q: JWT signing algorithm? → A: RS256 (asymmetric, public/private key pair)
- Q: MFA scope? → A: Deferred to future feature (out of scope for 004)
- Q: Session storage strategy? → A: Database-backed (refresh tokens stored in DB, revocable)

### Session 2026-08-08

- Q: User can hold roles in multiple tenants, but a token carries one tenant_id — how does the system pick which tenant's context goes into the token at login? → A: Prompt user to pick tenant. If user has roles in >1 tenant, login response lists tenants; user selects one before token issued. Single-tenant users skip this step.
- Q: FR-011 rate-limits login attempts by IP only — should the system also lock out a specific account after repeated failed attempts regardless of source IP? → A: Yes, add per-account lockout after 5 failed attempts within 15 minutes, independent of the IP rate limit.
- Q: FR-013 says password reset uses "a secure token" but doesn't state its expiration — how long should the token stay valid? → A: 1 hour.
- Q: US3 says a manager creates a staff account and "a temporary password is generated" — how does the new staff member receive it? → A: Emailed to the user's account email address.
- Q: If a manager creates a staff account using an email that already has a global account in a different tenant, what happens? → A: System links a new role for this tenant to the existing user and notifies them by email; no new account or password is created.

## User Scenarios & Testing _(mandatory)_

### User Story 1 - Staff Member Logs In (Priority: P1)

As a restaurant staff member, I want to log in with my credentials, so that I can access the system features assigned to my role.

**Why this priority**: No system functionality is accessible without authentication. This is the gateway to all features.

**Independent Test**: Can be fully tested by attempting login with valid and invalid credentials.

**Acceptance Scenarios**:

1. **Given** a staff member has valid credentials, **When** they submit their email and password, **Then** they receive an access token and are logged in successfully.

2. **Given** a staff member enters an incorrect password, **When** they submit the login form, **Then** they see a generic error message (not revealing which field was wrong).

3. **Given** a staff member's account is disabled, **When** they attempt to login, **Then** they see an "account disabled" message.

4. **Given** a staff member is logged in, **When** their session is active, **Then** they can access features allowed by their role.

5. **Given** a staff member has roles in more than one tenant, **When** they submit valid credentials, **Then** they are shown a list of their tenants and must select one before a token is issued.

6. **Given** a staff member has a role in only one tenant, **When** they submit valid credentials, **Then** a token for that tenant is issued immediately without a selection step.

---

### User Story 2 - Staff Member Session Persists (Priority: P1)

As a staff member using the mobile app during a busy shift, I want my session to persist, so that I don't have to log in repeatedly.

**Why this priority**: Frequent re-authentication during service would severely impact restaurant operations.

**Independent Test**: Can be fully tested by logging in and verifying session persists across app restarts.

**Acceptance Scenarios**:

1. **Given** a staff member is logged in, **When** they close and reopen the app within the session period, **Then** they remain logged in.

2. **Given** an access token is about to expire, **When** the app makes a request, **Then** the token is automatically refreshed without user intervention.

3. **Given** a refresh token has expired, **When** the app attempts to refresh, **Then** the user is prompted to log in again.

---

### User Story 3 - Manager Creates Staff Account (Priority: P2)

As a restaurant manager, I want to create accounts for my staff, so that they can access the system with appropriate permissions.

**Why this priority**: Staff management is essential but secondary to core authentication.

**Independent Test**: Can be fully tested by creating a user and verifying they can log in.

**Acceptance Scenarios**:

1. **Given** a manager is logged in, **When** they create a new staff account with email and role, **Then** the account is created, a temporary password is generated, and it is emailed to the new account's email address.

2. **Given** a new account is created, **When** the staff member first logs in, **Then** they are prompted to set their own password.

3. **Given** a manager creates an account, **When** they assign a role (waiter, kitchen, cashier), **Then** that user has only the permissions for that role.

4. **Given** a manager creates a staff account using an email that already has a global account in a different tenant, **When** the account is created, **Then** the system links a new role for this tenant to the existing user, notifies them by email, and does not create a duplicate account or temporary password.

---

### User Story 4 - System Enforces Role-Based Access (Priority: P1)

As a system administrator, I want users to only access features permitted by their role, so that sensitive operations are protected.

**Why this priority**: Authorization is as critical as authentication for system security.

**Independent Test**: Can be fully tested by attempting operations with different role accounts.

**Acceptance Scenarios**:

1. **Given** a waiter is logged in, **When** they attempt to access admin settings, **Then** access is denied with a "forbidden" response.

2. **Given** a manager is logged in, **When** they access staff management, **Then** they can view and modify staff in their restaurant only.

3. **Given** an admin is logged in, **When** they access any feature, **Then** they have full access across all tenants they manage.

---

### User Story 5 - Staff Member Logs Out (Priority: P3)

As a staff member ending my shift, I want to log out securely, so that the next person cannot access my account.

**Why this priority**: Important for security but not blocking for core functionality.

**Independent Test**: Can be fully tested by logging out and verifying session is terminated.

**Acceptance Scenarios**:

1. **Given** a staff member is logged in, **When** they tap the logout button, **Then** their session is terminated and tokens are invalidated.

2. **Given** a staff member has logged out, **When** someone attempts to use their old token, **Then** the request is rejected as unauthorized.

---

### Edge Cases

- What happens when the same user logs in from multiple devices? All sessions should remain valid independently.
- What happens during a password reset? Current sessions should remain valid until password is changed.
- What happens if a user's role changes while they're logged in? Changes take effect on next token refresh.
- What happens after 5 consecutive failed login attempts on the same account within 15 minutes (regardless of source IP)? The account is locked and further login attempts are rejected until the lockout expires or an authorized user unlocks it.

## Requirements _(mandatory)_

### Functional Requirements

- **FR-001**: System MUST authenticate users with email and password.

- **FR-002**: System MUST issue JWT access tokens signed with RS256 (asymmetric) upon successful authentication.

- **FR-003**: Access tokens MUST expire within 60 minutes.

- **FR-004**: System MUST issue refresh tokens with 30-day expiration.

- **FR-005**: System MUST support token refresh without re-entering credentials.

- **FR-006**: System MUST include tenant context (tenant_id) in all tokens. If a user holds roles in more than one tenant, the system MUST prompt them to select a tenant after credential verification and before issuing a token; users with a role in only one tenant skip this step.

- **FR-007**: System MUST support these roles: owner, admin, manager, cashier, waiter, kitchen, viewer.

- **FR-008**: System MUST enforce role-based access control on all API endpoints.

- **FR-009**: System MUST hash passwords using Argon2id with secure parameters.

- **FR-010**: System MUST log all authentication events (login, logout, failed attempts).

- **FR-011**: System MUST rate-limit login attempts to 5 per minute per IP.

- **FR-011a**: System MUST lock a user account after 5 consecutive failed login attempts within 15 minutes, independent of source IP, and reject further login attempts on that account until the lockout expires or an authorized user unlocks it.

- **FR-012**: System MUST support account creation by managers for their tenant. Upon creation, the system MUST email the generated temporary password to the new account's email address. If the email already has a global account in a different tenant, the system MUST link a new role for this tenant to the existing user and notify them by email, instead of creating a duplicate account or temporary password.

- **FR-013**: System MUST support password reset via a secure token that expires 1 hour after issuance.

- **FR-014**: System MUST invalidate all sessions when password is changed.

### Key Entities

- **User**: Person with access to the system; has globally unique email, hashed password, and can have roles in multiple tenants.
- **Role**: Permission set (owner, admin, manager, cashier, waiter, kitchen, viewer).
- **Session**: Active authentication represented by access and refresh tokens; refresh tokens stored in database for revocation support.
- **Tenant**: Restaurant business; users can have roles in multiple tenants.
- **AuthEvent**: Audit log of authentication actions (login, logout, failed attempt).

## Success Criteria _(mandatory)_

### Measurable Outcomes

- **SC-001**: Users can log in and receive tokens within 500ms under normal load.

- **SC-002**: Token refresh completes within 200ms without user-perceived delay.

- **SC-003**: 100% of API endpoints enforce authentication (no unprotected routes except login, refresh, and password-reset request/complete).

- **SC-004**: 100% of protected endpoints enforce role-based authorization.

- **SC-005**: Failed login attempts are rate-limited, blocking after 5 failures within 1 minute.

- **SC-006**: All authentication events are logged with timestamp, user, IP, and result.

- **SC-007**: Password hashing uses Argon2id with 64MB memory, 3 iterations, and 4-way parallelism (OWASP-recommended parameters).
