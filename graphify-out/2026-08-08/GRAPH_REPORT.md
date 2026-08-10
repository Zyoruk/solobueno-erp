# Graph Report - solobueno-erp  (2026-08-08)

## Corpus Check
- 208 files · ~136,741 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 1864 nodes · 3198 edges · 151 communities (93 shown, 58 thin omitted)
- Extraction: 88% EXTRACTED · 12% INFERRED · 0% AMBIGUOUS · INFERRED: 371 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `40ceab7f`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- ci.yml GitHub Actions Workflow
- Solobueno ERP README
- validate.py
- tasks
- caveman-compress (skill overview)
- NewPasswordService
- compilerOptions
- devDependencies
- backoffice/package.json
- mobile/package.json
- Feature Specification: Docker Local Development Environment
- Role
- admin/package.json
- kitchen-display/package.json
- graphql-client/package.json
- update-agent-context.sh
- Order
- analytics/package.json
- i18n/package.json
- types/package.json
- ui/package.json
- types/src/index.ts
- setupTestDB
- analytics/src/index.ts
- admin/tsconfig.json
- backoffice/tsconfig.json
- kitchen-display/tsconfig.json
- mobile/tsconfig.json
- ui/tsconfig.json
- Orders Module
- analytics/tsconfig.json
- graphql-client/tsconfig.json
- i18n/tsconfig.json
- types/tsconfig.json
- cavecrew (SKILL instructions)
- version_test.go
- create-new-feature.sh
- speckit.constitution command
- i18n/src/index.ts
- speckit.checklist command
- graphql-client/src/index.ts
- admin/src/index.ts
- backoffice/src/index.ts
- kitchen-display/src/index.ts
- mobile/src/index.ts
- speckit.specify Command
- speckit.clarify command
- speckit.tasks Command
- ui/src/index.ts
- Observability & Monitoring
- __init__.py
- NewConnection
- health-check.sh
- NotificationTemplate
- vitest.config.ts
- github.com/solobueno/erp
- User-Centric Interface Design
- Role
- Session
- Tenant
- User
- BrandingSettings
- BusinessSettings
- FeatureFlag
- GlobalConfig
- TenantConfig
- Analytics Tracker
- GraphQL Client
- Locale Files
- Shared Packages
- Translation Function
- Type Definitions
- UI Components
- Category
- MenuItemImage
- Modifier
- ModifierGroup
- Reservation
- Section
- ServerAssignment
- TableSession
- Check
- KitchenTicket
- OrderModification
- IngredientCategory
- Supplier
- CashDrawer
- PaymentMethod
- PaymentPlugin
- PaymentTerminal
- Refund
- BillingPlugin
- CreditNote
- InvoiceLine
- TaxRate
- DeliveryAttempt
- NotificationChannel
- NotificationPreference
- Dashboard
- MetricDefinition
- Report (Analytics)
- ComplaintCategory
- FeedbackResponse
- ReportExport
- ReportPermission
- ScheduledReport
- setupAuthMiddleware
- UserService
- KeyManager
- TokenService
- Tasks: Authentication Module
- dto.go
- setupUserHandler
- NewMemoryRateLimiter
- speckit-analyze/SKILL.md
- Quickstart: Authentication Module
- AuthEvent
- Domain Entities (Go with GORM)
- Tenant
- Context
- PasswordResetToken
- Endpoints
- UUID
- Execution Steps
- Research Tasks
- GormSessionRepository
- GormUserRepository
- Session
- GormAuthEventRepository
- User
- GormPasswordResetRepository
- GormUserTenantRoleRepository
- speckit-plan/SKILL.md
- speckit-specify/SKILL.md
- speckit-tasks/SKILL.md
- Core Principles
- user_test.go
- Implementation Plan: Authentication Module
- session_test.go
- speckit-checklist/SKILL.md
- password_reset_token_test.go
- speckit-clarify/SKILL.md
- speckit-implement/SKILL.md
- Seed
- AuthError
- speckit-constitution/SKILL.md
- speckit-taskstoissues/SKILL.md
- tenant_test.go

## God Nodes (most connected - your core abstractions)
1. `setupTestDB()` - 42 edges
2. `NewPasswordService()` - 37 edges
3. `Role` - 35 edges
4. `setupUserHandler()` - 35 edges
5. `setupUserService()` - 35 edges
6. `setupAuthService()` - 32 edges
7. `setupWiredAuthHandler()` - 31 edges
8. `UserService` - 29 edges
9. `AuthService` - 25 edges
10. `NewAuthHandler()` - 23 edges

## Surprising Connections (you probably didn't know these)
- `Graphify Project Instructions (CLAUDE.md)` --semantically_similar_to--> `Agent Context File Template`  [INFERRED] [semantically similar]
  CLAUDE.md → .specify/templates/agent-file-template.md
- `CI GitHub Actions Workflow` --implements--> `Test-Driven Development`  [INFERRED]
  .github/workflows/ci.yml → .specify/memory/constitution.md
- `Solobueno ERP README` --references--> `Mobile-First for Operations`  [INFERRED]
  README.md → .specify/memory/constitution.md
- `Solobueno ERP README` --references--> `Offline-First Operations`  [INFERRED]
  README.md → .specify/memory/constitution.md
- `Solobueno ERP README` --references--> `Plugin-Driven Architecture`  [INFERRED]
  README.md → .specify/memory/constitution.md

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Caveman toolkit skill family** — _agents_skills_caveman_readme_caveman, _agents_skills_caveman_commit_readme_caveman_commit, _agents_skills_caveman_compress_readme_caveman_compress, _agents_skills_caveman_help_readme_caveman_help, _agents_skills_caveman_review_readme_caveman_review, _agents_skills_caveman_stats_readme_caveman_stats, _agents_skills_cavecrew_readme_cavecrew [INFERRED 0.85]
- **cavecrew locate-fix-verify delegation chain** — _agents_skills_cavecrew_readme_cavecrew_investigator, _agents_skills_cavecrew_readme_cavecrew_builder, _agents_skills_cavecrew_readme_cavecrew_reviewer [EXTRACTED 1.00]
- **speckit spec-driven development pipeline** — _cursor_commands_speckit_constitution_speckit_constitution, _cursor_commands_speckit_clarify_speckit_clarify, _cursor_commands_speckit_checklist_speckit_checklist, _cursor_commands_speckit_analyze_speckit_analyze, _cursor_commands_speckit_implement_speckit_implement [INFERRED 0.85]
- **Frontline Operations UX Pattern** — specify_memory_constitution_mobilefirstforoperations, specify_memory_constitution_offlinefirstoperations, specify_memory_constitution_usercentricinterfacedesign [INFERRED 0.85]
- **Multi-Country Expansion Strategy** — specify_memory_constitution_plugindrivenarchitecture, specify_memory_constitution_i18npluggablecompliance, specify_memory_constitution_whitelabelmultitenant [INFERRED 0.85]
- **001-init-monorepo Feature Artifact Bundle** — specs_001_init_monorepo_checklists_requirements_speckitchecklist, specs_001_init_monorepo_contracts_readme_contractsoverview, specs_001_init_monorepo_data_model_datamodeldoc [EXTRACTED 1.00]
- **CI Pipeline Job Flow (lint to build to test)** — ci_workflow_ci_yml, lint_job, build_job, test_job [EXTRACTED 1.00]
- **Local Development Services Stack** — postgresql_container, redis_container, minio_container, docker_network_solobueno, docker_volumes_named [EXTRACTED 1.00]
- **Security Scanning and Merge Gating** — codeql_workflow_codeql_yml, dependabot_config, specs_003_ci_pipeline_branch_protection_rules_main [EXTRACTED 1.00]
- **Pluggable Integration Pattern (Payments, Billing, Notifications)** — specs_011_payments_module_spec_paymentplugin, specs_012_billing_module_spec_billingplugin, specs_013_notifications_module_spec_notificationchannel [INFERRED 0.75]
- **Order as Cross-Module Reference Hub** — specs_009_orders_module_spec_order, specs_011_payments_module_spec_payment, specs_012_billing_module_spec_invoice, specs_015_feedback_module_spec_rating [INFERRED 0.85]
- **Reporting Module Aggregates All Business Modules** — specs_016_reporting_module_spec_reporting_module, specs_009_orders_module_spec_orders_module, specs_010_inventory_module_spec_inventory_module, specs_011_payments_module_spec_payments_module, specs_012_billing_module_spec_billing_module, specs_014_analytics_module_spec_analytics_module, specs_015_feedback_module_spec_feedback_module [EXTRACTED 1.00]

## Communities (151 total, 58 thin omitted)

### Community 0 - "ci.yml GitHub Actions Workflow"
Cohesion: 0.12
Nodes (40): Build Job (ci.yml), Multi-layer CI Caching Strategy, ci.yml GitHub Actions Workflow, GitHub Actions Initial CI Setup, 80% Code Coverage Threshold, Prettier + ESLint + golangci-lint, CODEOWNERS File, CodeQL Analyze Job (+32 more)

### Community 1 - "Solobueno ERP README"
Cohesion: 0.07
Nodes (40): Graphify Project Instructions (CLAUDE.md), Phase 0: Outline & Research, Phase 1: Design & Contracts, speckit.plan Command, Dependabot Configuration, CI GitHub Actions Workflow, Go Coverage Threshold (interim 10%), CodeQL Security Analysis Workflow (+32 more)

### Community 2 - "validate.py"
Cohesion: 0.06
Nodes (56): benchmark_pair(), count_tokens(), main(), print_table(), Path, main(), print_usage(), backup_dir_for() (+48 more)

### Community 3 - "tasks"
Cohesion: 0.07
Nodes (31): .env, **/.env.*local, .next/**, NODE_ENV, *.tsbuildinfo, dependsOn, env, outputs (+23 more)

### Community 4 - "caveman-compress (skill overview)"
Cohesion: 0.08
Nodes (29): caveman-commit (skill overview), Conventional Commits format, Auto-Clarity rule (commit), caveman-commit (SKILL instructions), out-of-tree backup dir (README), caveman-compress (skill overview), caveman-compress (SECURITY doc), Snyk High Risk rating (false positive) (+21 more)

### Community 5 - "NewPasswordService"
Cohesion: 0.05
Nodes (102): T, TestE2E_LoginUseRefreshLogout(), NewAuthHandler(), T, setupWiredAuthHandler(), TestAuthHandler_ChangePassword_IncorrectCurrent(), TestAuthHandler_ChangePassword_MissingFields(), TestAuthHandler_ChangePassword_Success() (+94 more)

### Community 6 - "compilerOptions"
Cohesion: 0.08
Nodes (25): ES2022, compilerOptions, declaration, declarationMap, esModuleInterop, forceConsistentCasingInFileNames, isolatedModules, lib (+17 more)

### Community 7 - "devDependencies"
Cohesion: 0.04
Nodes (47): eslint, eslint-config-prettier, husky, lint-staged, devDependencies, eslint, eslint-config-prettier, husky (+39 more)

### Community 8 - "backoffice/package.json"
Cohesion: 0.08
Nodes (23): dependencies, @solobueno/analytics, @solobueno/graphql, @solobueno/i18n, @solobueno/types, @solobueno/ui, devDependencies, typescript (+15 more)

### Community 9 - "mobile/package.json"
Cohesion: 0.08
Nodes (23): dependencies, @solobueno/analytics, @solobueno/graphql, @solobueno/i18n, @solobueno/types, @solobueno/ui, devDependencies, typescript (+15 more)

### Community 10 - "Feature Specification: Docker Local Development Environment"
Cohesion: 0.17
Nodes (24): Named Volume Data Persistence Strategy, Simple Development Credentials, Enhanced docker-compose.yml (002), solobueno-network Docker Network, Named Docker Volumes (postgres/redis/minio data), health-check.sh Script, Native Docker Health Check Strategy, Makefile docker-* Targets (+16 more)

### Community 11 - "Role"
Cohesion: 0.06
Nodes (45): BaseEvent, DomainEvent, LoginFailedEvent, LoginSucceededEvent, LogoutEvent, PasswordChangedEvent, RoleChangedEvent, SessionRevokedEvent (+37 more)

### Community 12 - "admin/package.json"
Cohesion: 0.09
Nodes (21): dependencies, @solobueno/graphql, @solobueno/i18n, @solobueno/types, @solobueno/ui, devDependencies, typescript, @solobueno/graphql (+13 more)

### Community 13 - "kitchen-display/package.json"
Cohesion: 0.09
Nodes (21): dependencies, @solobueno/graphql, @solobueno/i18n, @solobueno/types, @solobueno/ui, devDependencies, typescript, @solobueno/graphql (+13 more)

### Community 14 - "graphql-client/package.json"
Cohesion: 0.09
Nodes (21): dependencies, @solobueno/types, devDependencies, tsup, typescript, exports, @solobueno/types, tsup (+13 more)

### Community 15 - "update-agent-context.sh"
Cohesion: 0.10
Nodes (23): check-prerequisites.sh script, check_dir(), check_feature_branch(), check_file(), get_feature_paths(), has_git(), common.sh script, setup-plan.sh script (+15 more)

### Community 16 - "Order"
Cohesion: 0.12
Nodes (20): AuthEvent, ConfigChange, MenuItem, Table, Order, OrderItem, Ingredient, Recipe (+12 more)

### Community 17 - "analytics/package.json"
Cohesion: 0.11
Nodes (17): devDependencies, tsup, typescript, exports, tsup, typescript, main, module (+9 more)

### Community 18 - "i18n/package.json"
Cohesion: 0.11
Nodes (17): devDependencies, tsup, typescript, exports, tsup, typescript, main, module (+9 more)

### Community 19 - "types/package.json"
Cohesion: 0.11
Nodes (17): devDependencies, tsup, typescript, exports, tsup, typescript, main, module (+9 more)

### Community 20 - "ui/package.json"
Cohesion: 0.11
Nodes (17): devDependencies, tsup, typescript, exports, tsup, typescript, main, module (+9 more)

### Community 21 - "types/src/index.ts"
Cohesion: 0.25
Nodes (14): BaseEntity, MenuCategory, MenuItem, Order, OrderItem, OrderStatus, Payment, PaymentMethod (+6 more)

### Community 22 - "setupTestDB"
Cohesion: 0.11
Nodes (60): ModuleConfig, NewAuthEvent(), T, TestAuthEvent_TableName(), TestAuthEvent_WithMetadata(), TestAuthEventType_GormDataType(), TestAuthEventType_String(), TestMetadata_GormDataType() (+52 more)

### Community 23 - "analytics/src/index.ts"
Cohesion: 0.27
Nodes (11): AnalyticsEvent, eventQueue, generateId(), getQueuedEvents(), initAnalytics(), QueuedEvent, TODO: Process queued events, TODO: Send to analytics service (+3 more)

### Community 24 - "admin/tsconfig.json"
Cohesion: 0.17
Nodes (11): compilerOptions, jsx, outDir, rootDir, exclude, extends, include, dist (+3 more)

### Community 25 - "backoffice/tsconfig.json"
Cohesion: 0.17
Nodes (11): compilerOptions, jsx, outDir, rootDir, exclude, extends, include, dist (+3 more)

### Community 26 - "kitchen-display/tsconfig.json"
Cohesion: 0.17
Nodes (11): compilerOptions, jsx, outDir, rootDir, exclude, extends, include, dist (+3 more)

### Community 27 - "mobile/tsconfig.json"
Cohesion: 0.17
Nodes (11): compilerOptions, jsx, outDir, rootDir, exclude, extends, include, dist (+3 more)

### Community 28 - "ui/tsconfig.json"
Cohesion: 0.17
Nodes (11): compilerOptions, jsx, outDir, rootDir, exclude, extends, include, dist (+3 more)

### Community 29 - "Orders Module"
Cohesion: 0.30
Nodes (12): Authentication Module, Configuration Module, Menu Module, Tables Module, Orders Module, Inventory Module, Payments Module, Billing Module (+4 more)

### Community 30 - "analytics/tsconfig.json"
Cohesion: 0.18
Nodes (10): compilerOptions, outDir, rootDir, exclude, extends, include, dist, node_modules (+2 more)

### Community 31 - "graphql-client/tsconfig.json"
Cohesion: 0.18
Nodes (10): compilerOptions, outDir, rootDir, exclude, extends, include, dist, node_modules (+2 more)

### Community 32 - "i18n/tsconfig.json"
Cohesion: 0.18
Nodes (10): compilerOptions, outDir, rootDir, exclude, extends, include, dist, node_modules (+2 more)

### Community 33 - "types/tsconfig.json"
Cohesion: 0.18
Nodes (10): compilerOptions, outDir, rootDir, exclude, extends, include, dist, node_modules (+2 more)

### Community 34 - "cavecrew (SKILL instructions)"
Cohesion: 0.31
Nodes (9): cavecrew (skill overview), cavecrew-builder subagent, cavecrew-investigator subagent, cavecrew-reviewer subagent, Caveman toolkit root README, cavecrew (SKILL instructions), Code Reviewer (vanilla agent), Explore (vanilla agent) (+1 more)

### Community 35 - "version_test.go"
Cohesion: 0.33
Nodes (7): Info(), IsPreRelease(), T, TestAppName(), TestInfo(), TestIsPreRelease(), TestVersion()

### Community 37 - "speckit.constitution command"
Cohesion: 0.29
Nodes (7): .specify/scripts/bash/check-prerequisites.sh, .specify/memory/constitution.md, speckit.analyze command, .specify/memory/constitution.md (target), speckit.constitution command, handoff to speckit.specify, Sync Impact Report

### Community 38 - "i18n/src/index.ts"
Cohesion: 0.43
Nodes (5): getLocale(), Locale, locales, setLocale(), t()

### Community 39 - "speckit.checklist command"
Cohesion: 0.33
Nodes (6): .specify/templates/checklist-template.md, speckit.checklist command, "Unit Tests for English" principle, FEATURE_DIR/checklists/ gate, speckit.implement command, speckit.tasks command (referenced)

### Community 40 - "graphql-client/src/index.ts"
Cohesion: 0.33
Nodes (4): GRAPHQL_VERSION, GraphQLClientConfig, TODO: Generated GraphQL operations will be exported here, TODO: Implement GraphQL client

### Community 41 - "admin/src/index.ts"
Cohesion: 0.50
Nodes (3): APP_NAME, APP_VERSION, TODO: Implement React admin portal

### Community 42 - "backoffice/src/index.ts"
Cohesion: 0.50
Nodes (3): APP_NAME, APP_VERSION, TODO: Implement React web app

### Community 43 - "kitchen-display/src/index.ts"
Cohesion: 0.50
Nodes (3): APP_NAME, APP_VERSION, TODO: Implement React Native kitchen display

### Community 44 - "mobile/src/index.ts"
Cohesion: 0.50
Nodes (3): APP_NAME, APP_VERSION, TODO: Implement React Native app entry point

### Community 45 - "speckit.specify Command"
Cohesion: 0.67
Nodes (4): speckit.specify Command, Checklist Template, Feature Specification Template, Init Monorepo Spec Quality Checklist

### Community 46 - "speckit.clarify command"
Cohesion: 0.67
Nodes (3): ## Clarifications spec section, speckit.clarify command, handoff to speckit.plan

### Community 47 - "speckit.tasks Command"
Cohesion: 0.67
Nodes (3): speckit.tasks Command, speckit.taskstoissues Command, Tasks Template

### Community 49 - "Observability & Monitoring"
Cohesion: 0.67
Nodes (3): Clickstream & Analytics Events, Feedback Module, Observability & Monitoring

### Community 51 - "NewConnection"
Cohesion: 0.25
Nodes (11): main(), AutoMigrate(), DropAll(), DB, DefaultConfig(), getEnv(), DB, Duration (+3 more)

### Community 108 - "setupAuthMiddleware"
Cohesion: 0.10
Nodes (39): Request, ResponseWriter, writeError(), writeJSON(), User, ToUserResponse(), extractBearerToken(), GetClaims() (+31 more)

### Community 109 - "UserService"
Cohesion: 0.08
Nodes (30): Module, NewAuthMiddleware(), Router(), UserRouter(), Claims, Context, User, UUID (+22 more)

### Community 110 - "KeyManager"
Cohesion: 0.09
Nodes (27): DefaultTokenGeneratorConfig(), GetUserIDFromClaims(), Duration, RegisteredClaims, Time, UUID, NewTokenGenerator(), NewTokenValidator() (+19 more)

### Community 111 - "TokenService"
Cohesion: 0.07
Nodes (20): RegisteredClaims, Time, UUID, NewClaims(), NewTokenPair(), T, TestClaims_GetUserID(), TestClaims_IsExpired() (+12 more)

### Community 112 - "Tasks: Authentication Module"
Cohesion: 0.05
Nodes (38): Dependencies & Execution Order, Endpoint Protection, Estimated Task Counts, Format: `[ID] [P?] [Story] Description`, Implementation, Implementation, Implementation, Implementation (+30 more)

### Community 113 - "dto.go"
Cohesion: 0.10
Nodes (32): Time, UUID, T, TestToLoginResponse(), TestToTenantOptions(), TestToTenantOptions_Empty(), TestToTokenResponse(), TestToUserResponse() (+24 more)

### Community 114 - "setupUserHandler"
Cohesion: 0.23
Nodes (30): authedContext(), Context, Request, T, UUID, setupUserHandler(), TestUserHandler_Create_CannotAssignRole(), TestUserHandler_Create_EmailExists() (+22 more)

### Community 115 - "NewMemoryRateLimiter"
Cohesion: 0.13
Nodes (20): DefaultLoginRateLimiterConfig(), DefaultPasswordResetRateLimiterConfig(), Duration, Context, RWMutex, Time, NewMemoryRateLimiter(), T (+12 more)

### Community 116 - "speckit-analyze/SKILL.md"
Cohesion: 0.08
Nodes (25): 1. Initialize Analysis Context, 2. Load Artifacts (Progressive Disclosure), 3. Build Semantic Models, 4. Detection Passes (Token-Efficient Analysis), 5. Severity Assignment, 6. Produce Compact Analysis Report, 7. Provide Next Actions, 8. Offer Remediation (+17 more)

### Community 117 - "Quickstart: Authentication Module"
Cohesion: 0.08
Nodes (24): 1. Login, 2. Access Protected Endpoint, 3. Refresh Token, 4. Logout, Authentication Flows, Checking Permissions, Complete Reset, Create User (+16 more)

### Community 118 - "AuthEvent"
Cohesion: 0.14
Nodes (8): Time, UUID, Value, Time, AuthEvent, AuthEventType, Metadata, MockAuthEventRepository

### Community 119 - "Domain Entities (Go with GORM)"
Cohesion: 0.09
Nodes (22): AuthEvent, Claims (JWT Payload, not persisted), Data Model: Authentication Module, Database Schema, Domain Entities (Go with GORM), Entity Relationships, GORM AutoMigrate, Indexes Summary (+14 more)

### Community 120 - "Tenant"
Cohesion: 0.14
Nodes (8): Time, UUID, Context, DB, UUID, Tenant, MockTenantRepository, GormTenantRepository

### Community 121 - "Context"
Cohesion: 0.22
Nodes (4): Context, User, MockSessionRepository, MockUserRepository

### Community 122 - "PasswordResetToken"
Cohesion: 0.12
Nodes (6): Time, User, UUID, RWMutex, PasswordResetToken, MockPasswordResetRepository

### Community 123 - "Endpoints"
Cohesion: 0.11
Nodes (17): Auth API Contracts, Common Error Codes, Endpoints, GET /me, GET /users, GET /users/{id}, JWT Claims, PATCH /users/{id} (+9 more)

### Community 124 - "UUID"
Cohesion: 0.22
Nodes (6): Time, User, UUID, UUID, UserTenantRole, MockUserTenantRoleRepository

### Community 125 - "Execution Steps"
Cohesion: 0.12
Nodes (15): 1. Initialize Convergence Context, 2. Load Artifacts (Progressive Disclosure), 3. Build the Intent Inventory, 4. Assess the Codebase and Classify Findings, 5. Assign Severity, 6. Present the In-Session Findings Summary, 7. Append Convergence Tasks (or report converged), 8. Provide Next Actions (Handoff) (+7 more)

### Community 126 - "Research Tasks"
Cohesion: 0.13
Nodes (14): 1. Password Hashing: Argon2id Configuration, 2. JWT Implementation: RS256 with Key Rotation, 3. Session Storage: Database-Backed Refresh Tokens, 4. Rate Limiting: In-Memory with Redis Upgrade Path, 5. Multi-Tenant User Model, 6. Password Reset Flow, 7. Role-Based Access Control (RBAC), 8. ORM Selection (+6 more)

### Community 127 - "GormSessionRepository"
Cohesion: 0.31
Nodes (4): Context, DB, UUID, GormSessionRepository

### Community 128 - "GormUserRepository"
Cohesion: 0.35
Nodes (5): Context, DB, User, UUID, GormUserRepository

### Community 129 - "Session"
Cohesion: 0.17
Nodes (5): Duration, Time, User, UUID, Session

### Community 130 - "GormAuthEventRepository"
Cohesion: 0.32
Nodes (5): Context, DB, Time, UUID, GormAuthEventRepository

### Community 131 - "User"
Cohesion: 0.22
Nodes (3): Time, UUID, User

### Community 132 - "GormPasswordResetRepository"
Cohesion: 0.31
Nodes (5): Context, DB, Time, UUID, GormPasswordResetRepository

### Community 133 - "GormUserTenantRoleRepository"
Cohesion: 0.38
Nodes (4): Context, DB, UUID, GormUserTenantRoleRepository

### Community 134 - "speckit-plan/SKILL.md"
Cohesion: 0.18
Nodes (10): Completion Report, Done When, Key rules, Mandatory Post-Execution Hooks, Outline, Phase 0: Outline & Research, Phase 1: Design & Contracts, Phases (+2 more)

### Community 135 - "speckit-specify/SKILL.md"
Cohesion: 0.18
Nodes (10): Completion Report, Done When, For AI Generation, Mandatory Post-Execution Hooks, Outline, Pre-Execution Checks, Quick Guidelines, Section Requirements (+2 more)

### Community 136 - "speckit-tasks/SKILL.md"
Cohesion: 0.18
Nodes (10): Checklist Format (REQUIRED), Completion Report, Done When, Mandatory Post-Execution Hooks, Outline, Phase Structure, Pre-Execution Checks, Task Generation Rules (+2 more)

### Community 137 - "Core Principles"
Cohesion: 0.18
Nodes (10): Core Principles, Governance, [PRINCIPLE_1_NAME], [PRINCIPLE_2_NAME], [PRINCIPLE_3_NAME], [PRINCIPLE_4_NAME], [PRINCIPLE_5_NAME], [PROJECT_NAME] Constitution (+2 more)

### Community 138 - "user_test.go"
Cohesion: 0.39
Nodes (8): T, TestUser_CanLogin(), TestUser_FullName(), TestUser_GetRoleForTenant(), TestUser_HasTenant(), TestUser_NeedsPasswordReset(), TestUser_TableName(), TestUser_TenantCount()

### Community 139 - "Implementation Plan: Authentication Module"
Cohesion: 0.22
Nodes (8): Complexity Tracking, Constitution Check, Documentation (this feature), Implementation Plan: Authentication Module, Project Structure, Source Code (repository root), Summary, Technical Context

### Community 140 - "session_test.go"
Cohesion: 0.43
Nodes (7): T, TestSession_IsExpired(), TestSession_IsRevoked(), TestSession_IsValid(), TestSession_Revoke(), TestSession_TableName(), TestSession_TimeUntilExpiry()

### Community 141 - "speckit-checklist/SKILL.md"
Cohesion: 0.25
Nodes (7): Anti-Examples: What NOT To Do, Checklist Purpose: "Unit Tests for English", Example Checklist Types & Sample Items, Execution Steps, Post-Execution Checks, Pre-Execution Checks, User Input

### Community 142 - "password_reset_token_test.go"
Cohesion: 0.48
Nodes (6): T, TestPasswordResetToken_IsExpired(), TestPasswordResetToken_IsUsed(), TestPasswordResetToken_IsValid(), TestPasswordResetToken_MarkUsed(), TestPasswordResetToken_TableName()

### Community 143 - "speckit-clarify/SKILL.md"
Cohesion: 0.29
Nodes (6): Completion Report, Done When, Mandatory Post-Execution Hooks, Outline, Pre-Execution Checks, User Input

### Community 144 - "speckit-implement/SKILL.md"
Cohesion: 0.29
Nodes (6): Completion Report, Done When, Mandatory Post-Execution Hooks, Outline, Pre-Execution Checks, User Input

### Community 145 - "Seed"
Cohesion: 0.40
Nodes (5): SeedData, Context, DB, User, Seed()

### Community 147 - "speckit-constitution/SKILL.md"
Cohesion: 0.33
Nodes (5): Outline, Post-Execution Checks, Pre-Execution Checks, Scope Guard, User Input

### Community 148 - "speckit-taskstoissues/SKILL.md"
Cohesion: 0.40
Nodes (4): Outline, Post-Execution Checks, Pre-Execution Checks, User Input

### Community 149 - "tenant_test.go"
Cohesion: 0.67
Nodes (3): T, TestTenant_IsOperational(), TestTenant_TableName()

## Ambiguous Edges - Review These
- `Prettier + ESLint + golangci-lint` → `Lint Job (ci.yml)`  [AMBIGUOUS]
  specs/003-ci-pipeline/data-model.md · relation: implements

## Knowledge Gaps
- **561 isolated node(s):** `common.sh script`, `create-new-feature.sh script`, `SPECIFY_FEATURE`, `name`, `version` (+556 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **58 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **What is the exact relationship between `Prettier + ESLint + golangci-lint` and `Lint Job (ci.yml)`?**
  _Edge tagged AMBIGUOUS (relation: implements) - confidence is low._
- **Why does `Role` connect `Role` to `User`, `setupAuthMiddleware`, `UserService`, `TokenService`, `dto.go`, `setupUserHandler`, `UUID`?**
  _High betweenness centrality (0.045) - this node is a cross-community bridge._
- **Why does `NewModule()` connect `setupTestDB` to `NewPasswordService`, `NewMemoryRateLimiter`, `UserService`?**
  _High betweenness centrality (0.027) - this node is a cross-community bridge._
- **Why does `UserTenantRole` connect `UUID` to `Tenant`, `Role`, `User`, `GormUserTenantRoleRepository`?**
  _High betweenness centrality (0.018) - this node is a cross-community bridge._
- **Are the 35 inferred relationships involving `NewPasswordService()` (e.g. with `TestE2E_LoginUseRefreshLogout()` and `TestAuthHandler_ChangePassword_IncorrectCurrent()`) actually correct?**
  _`NewPasswordService()` has 35 INFERRED edges - model-reasoned connections that need verification._
- **What connects `common.sh script`, `create-new-feature.sh script`, `SPECIFY_FEATURE` to the rest of the system?**
  _561 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `ci.yml GitHub Actions Workflow` be split into smaller, more focused modules?**
  _Cohesion score 0.11538461538461539 - nodes in this community are weakly interconnected._