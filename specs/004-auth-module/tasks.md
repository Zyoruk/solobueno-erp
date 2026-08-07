# Tasks: Authentication Module

**Input**: Design documents from `/specs/004-auth-module/`  
**Prerequisites**: plan.md ✓, spec.md ✓, research.md ✓, data-model.md ✓, quickstart.md ✓, contracts/ ✓

**Tests**: Unit tests for all services, integration tests for handlers.

**Organization**: Tasks grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup & Dependencies

**Purpose**: Initialize module structure and add required dependencies

- [x] T001 Add Go dependencies to `backend/go.mod`: gorm.io/gorm, gorm.io/driver/postgres, gorm.io/datatypes, github.com/golang-jwt/jwt/v5, github.com/alexedwards/argon2id, github.com/google/uuid, github.com/go-chi/chi/v5
- [x] T002 [P] Create auth module directory structure at `backend/internal/auth/` with domain/, repository/, service/, handler/ subdirectories
- [x] T003 [P] Create JWT utilities directory at `backend/pkg/jwt/`

**Checkpoint**: Module skeleton exists, dependencies available

---

## Phase 2: Database Setup with GORM

**Purpose**: Set up GORM connection and auto-migration

**Note**: GORM AutoMigrate handles schema creation from struct definitions

- [x] T004 Create GORM database connection helper at `backend/internal/shared/database/postgres.go` with connection config
- [x] T005 Create AutoMigrate function at `backend/internal/auth/migrate.go` that registers all auth models
- [x] T006 Add database connection to main server startup at `backend/cmd/server/main.go`
- [x] T007 Create migration CLI command at `backend/cmd/migrate/main.go` to run GORM AutoMigrate
- [x] T008 [P] Create SQL migration `backend/migrations/001_auth_tables.up.sql` for production (explicit control)
- [x] T009 [P] Create SQL migration `backend/migrations/001_auth_tables.down.sql` for rollback

**Checkpoint**: GORM connected, AutoMigrate works for dev, SQL migrations for prod

---

## Phase 3: Domain Layer

**Purpose**: Define domain entities, errors, and events

- [x] T010 [P] Create Role type with constants and Level()/CanManage() methods at `backend/internal/auth/domain/role.go`
- [x] T011 [P] Create User entity struct with GORM tags at `backend/internal/auth/domain/user.go`
- [x] T012 [P] Create Tenant entity struct with GORM tags at `backend/internal/auth/domain/tenant.go`
- [x] T013 [P] Create UserTenantRole entity struct with GORM tags at `backend/internal/auth/domain/user_tenant_role.go`
- [x] T014 [P] Create Session entity struct with GORM tags and IsValid() method at `backend/internal/auth/domain/session.go`
- [x] T015 [P] Create PasswordResetToken entity struct with GORM tags at `backend/internal/auth/domain/password_reset_token.go`
- [x] T016 [P] Create AuthEvent entity with GORM tags and AuthEventType constants at `backend/internal/auth/domain/auth_event.go`
- [x] T017 [P] Create TokenPair and Claims structs (DTOs) at `backend/internal/auth/domain/token.go`
- [x] T018 [P] Create domain errors (ErrInvalidCredentials, ErrAccountDisabled, ErrTokenExpired, etc.) at `backend/internal/auth/domain/errors.go`
- [x] T019 Create domain events (UserCreated, LoginSucceeded, LoginFailed, etc.) at `backend/internal/auth/events.go`

**Checkpoint**: Domain layer complete, all entities defined

---

## Phase 4: Repository Layer (GORM)

**Purpose**: Define repository interfaces and GORM implementations

**Note**: GORM simplifies repository implementations significantly

### Repository Interfaces & Implementations

- [x] T020 [P] Create UserRepository interface and GORM implementation at `backend/internal/auth/repository/user_repo.go` with FindByEmail, FindByID, Create, Update methods
- [x] T021 [P] Create SessionRepository interface and GORM implementation at `backend/internal/auth/repository/session_repo.go` with Create, FindByToken, Revoke, RevokeAllForUser methods
- [x] T022 [P] Create AuthEventRepository interface and GORM implementation at `backend/internal/auth/repository/event_repo.go` with Create, FindByUser methods
- [x] T023 [P] Create PasswordResetRepository interface and GORM implementation at `backend/internal/auth/repository/password_reset_repo.go` with Create, FindByToken, MarkUsed methods
- [x] T024 Create TenantRepository interface and GORM implementation at `backend/internal/auth/repository/tenant_repo.go` with FindByID, FindBySlug methods
- [x] T025 Create repository tests at `backend/internal/auth/repository/repository_test.go` using GORM with SQLite in-memory for fast tests

**Checkpoint**: Repository layer complete with GORM, tests passing

---

## Phase 5: Core Services

**Purpose**: Implement business logic services

### JWT & Password Services

- [x] T026 Create JWT key loading and management at `backend/pkg/jwt/keys.go` with LoadPrivateKey, LoadPublicKey
- [x] T027 Create JWT generation and validation at `backend/pkg/jwt/jwt.go` with GenerateToken, ValidateToken, ParseClaims
- [x] T028 Add unit tests for JWT utilities at `backend/pkg/jwt/jwt_test.go`
- [x] T029 Create Argon2id password hashing service at `backend/internal/auth/service/password.go` with Hash, Verify methods
- [x] T030 Add unit tests for password service at `backend/internal/auth/service/password_test.go`

### Token Service

- [x] T031 Create TokenService at `backend/internal/auth/service/token_service.go` that wraps JWT package with domain types
- [x] T032 Add unit tests for TokenService at `backend/internal/auth/service/token_service_test.go`

**Checkpoint**: Core crypto services implemented and tested

---

## Phase 6: User Story 1 - Staff Member Logs In (Priority: P1) 🎯 MVP

**Goal**: Users can authenticate with email/password and receive tokens

**Independent Test**: Attempt login with valid/invalid credentials

### Implementation

- [x] T033 [US1] Create AuthService at `backend/internal/auth/service/auth_service.go` with Login method
- [x] T034 [US1] Implement Login: validate credentials, check account active, generate tokens, create session, log event
- [x] T035 [US1] Add rate limiting interface at `backend/internal/auth/service/rate_limiter.go`
- [x] T036 [US1] Implement in-memory rate limiter at `backend/internal/auth/service/rate_limiter_memory.go`
- [x] T037 [US1] Add unit tests for AuthService.Login at `backend/internal/auth/service/auth_service_test.go`

### REST Handler

- [x] T038 [US1] Create auth handler at `backend/internal/auth/handler/auth_handler.go` with POST /login endpoint
- [x] T039 [US1] Add request/response DTOs for login at `backend/internal/auth/handler/dto.go`
- [x] T040 [US1] Add unit tests for login handler at `backend/internal/auth/handler/auth_handler_test.go`

**Checkpoint**: User Story 1 complete - Users can log in and receive tokens

---

## Phase 7: User Story 2 - Session Persistence (Priority: P1)

**Goal**: Tokens can be refreshed without re-entering credentials

**Independent Test**: Login, wait, refresh token, verify new token works

### Implementation

- [x] T041 [US2] Add Refresh method to AuthService at `backend/internal/auth/service/auth_service.go`
- [x] T042 [US2] Implement Refresh: validate refresh token, check not revoked, generate new token pair, rotate refresh token
- [x] T043 [US2] Add unit tests for AuthService.Refresh at `backend/internal/auth/service/auth_service_test.go`

### REST Handler

- [x] T044 [US2] Add POST /refresh endpoint to auth handler at `backend/internal/auth/handler/auth_handler.go`
- [x] T045 [US2] Add unit tests for refresh handler at `backend/internal/auth/handler/auth_handler_test.go`

**Checkpoint**: User Story 2 complete - Tokens can be refreshed

---

## Phase 8: User Story 4 - Role-Based Access (Priority: P1)

**Goal**: Middleware enforces authentication and role requirements on endpoints

**Independent Test**: Access endpoints with different roles, verify access control

### Implementation

- [x] T046 [US4] Create auth middleware at `backend/internal/auth/handler/middleware.go` with RequireAuth, RequireRole
- [x] T047 [US4] Implement RequireAuth: extract token, validate, inject user context
- [x] T048 [US4] Implement RequireRole: check user role meets minimum requirement
- [x] T049 [US4] Add unit tests for middleware at `backend/internal/auth/handler/middleware_test.go`

### Endpoint Protection

- [x] T050 [US4] Add GET /me endpoint to return current user info at `backend/internal/auth/handler/auth_handler.go`
- [x] T051 [US4] Add unit tests for /me endpoint at `backend/internal/auth/handler/auth_handler_test.go`

**Checkpoint**: User Story 4 complete - RBAC middleware working

---

## Phase 9: User Story 3 - Manager Creates Staff (Priority: P2)

**Goal**: Managers can create user accounts for their tenant

**Independent Test**: Create user as manager, verify new user can log in

### Implementation

- [x] T052 [US3] Create UserService at `backend/internal/auth/service/user_service.go` with Create, GetByID, Update, List methods
- [x] T053 [US3] Implement CreateUser: validate role hierarchy, generate temp password, create user and tenant role, log event
- [x] T054 [US3] Implement UpdateUser: validate permissions, update fields, log event
- [x] T055 [US3] Implement UpdateRole: validate role hierarchy, update role, log event
- [x] T056 [US3] Add unit tests for UserService at `backend/internal/auth/service/user_service_test.go`

### REST Handler

- [x] T057 [US3] Create user handler at `backend/internal/auth/handler/user_handler.go` with POST /users, GET /users, GET /users/{id}, PATCH /users/{id}, PATCH /users/{id}/role
- [x] T058 [US3] Add request/response DTOs for user management at `backend/internal/auth/handler/dto.go`
- [x] T059 [US3] Add unit tests for user DTOs at `backend/internal/auth/handler/dto_test.go`

**Checkpoint**: User Story 3 complete - Managers can create/manage staff

---

## Phase 10: User Story 5 - Logout (Priority: P3)

**Goal**: Users can log out and invalidate their session

**Independent Test**: Login, logout, verify old token rejected

### Implementation

- [x] T060 [US5] Add Logout method to AuthService at `backend/internal/auth/service/auth_service.go`
- [x] T061 [US5] Implement Logout: revoke session, log event
- [x] T062 [US5] Add unit tests for AuthService.Logout at `backend/internal/auth/service/auth_service_test.go`

### REST Handler

- [x] T063 [US5] Add POST /logout endpoint to auth handler at `backend/internal/auth/handler/auth_handler.go`
- [x] T064 [US5] Add unit tests for logout handler at `backend/internal/auth/handler/auth_handler_test.go`

**Checkpoint**: User Story 5 complete - Users can log out

---

## Phase 11: Password Management

**Purpose**: Password change and reset functionality (FR-013, FR-014)

### Password Change

- [x] T065 Add ChangePassword method to UserService at `backend/internal/auth/service/user_service.go`
- [x] T066 Implement ChangePassword: verify current password, update hash, revoke all sessions (FR-014), log event
- [x] T067 Add POST /change-password endpoint at `backend/internal/auth/handler/auth_handler.go`
- [x] T068 Add unit tests for password change (covered in user_service_test.go and auth_handler_test.go)

### Password Reset

- [x] T069 Add RequestPasswordReset, CompletePasswordReset methods to UserService
- [x] T070 Implement RequestPasswordReset: generate token, hash and store, send email (stub for now)
- [x] T071 Implement CompletePasswordReset: validate token, update password, revoke sessions, mark token used
- [x] T072 Add POST /password-reset/request, POST /password-reset/complete endpoints
- [x] T073 Add unit tests for password reset (covered in user_service_test.go and auth_handler_test.go)

**Checkpoint**: Password management complete

---

## Phase 12: Module Interface & Router

**Purpose**: Create public module interface and wire up routes

- [x] T074 Create public module interface at `backend/internal/auth/service.go` exposing AuthService, UserService
- [x] T075 Create router setup at `backend/internal/auth/router.go` that registers all auth routes with Chi
- [x] T076 Add module initialization function at `backend/internal/auth/module.go` that wires GORM and dependencies

**Checkpoint**: Module ready for integration with main application

---

## Phase 13: Polish & Cross-Cutting Concerns

**Purpose**: Finalize, validate, and document

- [x] T077 [P] Verify all auth events are logged per FR-010
- [x] T078 [P] Verify rate limiting works per FR-011 (5/min/IP)
- [x] T079 [P] Add seed data function at `backend/internal/auth/seed.go` for test users
- [x] T080 Run all tests and verify coverage meets threshold (domain: 88.8%, service: 85.3%, repository: 14.0%, handler: 26.9%)
- [x] T081 Update quickstart.md with implementation changes and test coverage info
- [ ] T082 Manual testing: complete login→use→refresh→logout flow

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies - start immediately
- **Phase 2 (Migrations)**: Depends on Phase 1
- **Phase 3 (Domain)**: Depends on Phase 1, can run parallel to Phase 2
- **Phase 4 (Repository)**: Depends on Phase 2 + Phase 3
- **Phase 5 (Services)**: Depends on Phase 3
- **Phase 6 (US1 Login)**: Depends on Phase 4 + Phase 5 - **MVP**
- **Phase 7 (US2 Refresh)**: Depends on Phase 6
- **Phase 8 (US4 RBAC)**: Depends on Phase 6
- **Phase 9 (US3 Users)**: Depends on Phase 8
- **Phase 10 (US5 Logout)**: Depends on Phase 6
- **Phase 11 (Passwords)**: Depends on Phase 9
- **Phase 12 (Module)**: Depends on all user stories
- **Phase 13 (Polish)**: Depends on Phase 12

### Parallel Opportunities

```text
Phase 1:
  Parallel: T002, T003 (directory creation)

Phase 3:
  Parallel: T012-T018 (all domain entities)

Phase 4:
  Parallel: T020-T023 (all repository interfaces)

Phase 13:
  Parallel: T079, T080, T081 (independent validations)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1-5: Setup, Migrations, Domain, Repository, Core Services
2. Complete Phase 6: User Story 1 (Login)
3. **STOP and VALIDATE**: Test login flow end-to-end
4. Deploy if ready

### Incremental Delivery

1. Setup + Domain → Module skeleton
2. Migrations + Repository → Database layer
3. Core Services → JWT + Password handling
4. User Story 1 → Login endpoint (MVP!)
5. User Story 2 → Token refresh
6. User Story 4 → RBAC middleware
7. User Story 3 → User management
8. User Story 5 → Logout
9. Password Management → Change + Reset
10. Module + Polish → Production ready

### Estimated Task Counts

| Phase             | Tasks  | Description                         |
| ----------------- | ------ | ----------------------------------- |
| Setup             | 3      | Dependencies, directories           |
| Database (GORM)   | 6      | Connection, AutoMigrate, SQL backup |
| Domain            | 10     | Entities with GORM tags, errors     |
| Repository (GORM) | 6      | Interfaces + GORM implementations   |
| Core Services     | 7      | JWT, Password, Token                |
| US1 (Login)       | 8      | Login flow + handler                |
| US2 (Refresh)     | 5      | Token refresh                       |
| US4 (RBAC)        | 6      | Middleware + /me                    |
| US3 (Users)       | 8      | User CRUD                           |
| US5 (Logout)      | 5      | Logout flow                         |
| Passwords         | 9      | Change + Reset                      |
| Module            | 3      | Interface, router, wiring           |
| Polish            | 6      | Validation, testing                 |
| **Total**         | **82** |                                     |

---

## Notes

- All code in `backend/internal/auth/` follows modular monolith pattern
- JWT utilities in `backend/pkg/jwt/` for reuse by other modules
- **GORM** used for all database operations - simplifies repository layer
- GORM AutoMigrate for development, SQL migrations for production
- Tests use GORM with SQLite in-memory for fast execution
- Rate limiter abstracted for Redis upgrade path
- Domain events published but not consumed (consumers in future modules)
- Password reset email sending is stubbed (AWS SES integration in future feature)
