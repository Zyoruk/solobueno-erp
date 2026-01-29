# Tasks: Initialize Monorepo Structure

**Input**: Design documents from `/specs/001-init-monorepo/`  
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅

**Tests**: Not requested for this infrastructure feature.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Root configs**: `./` (repository root)
- **Apps**: `apps/{app-name}/`
- **Packages**: `packages/{package-name}/`
- **Backend**: `backend/`
- **Infrastructure**: `infrastructure/`

---

## Phase 1: Setup (Root Configuration)

**Purpose**: Create root-level configuration files for monorepo management

- [ ] T001 [P] Create root package.json with Turborepo scripts and devDependencies in `./package.json`
- [ ] T002 [P] Create Turborepo configuration with build/test/lint pipelines in `./turbo.json`
- [ ] T003 [P] Create pnpm workspace configuration in `./pnpm-workspace.yaml`
- [ ] T004 [P] Create shared TypeScript base configuration in `./tsconfig.base.json`
- [ ] T005 [P] Create ESLint configuration for TypeScript in `./.eslintrc.js`
- [ ] T006 [P] Create Prettier configuration in `./.prettierrc`
- [ ] T007 [P] Create .gitignore with Node, Go, IDE, and build patterns in `./.gitignore`
- [ ] T008 [P] Create EditorConfig for consistent formatting in `./.editorconfig`
- [ ] T009 [P] Create Node version file specifying Node 20 in `./.nvmrc`

**Checkpoint**: Root configuration complete - can run `pnpm install`

---

## Phase 2: Foundational (Backend & Infrastructure)

**Purpose**: Core infrastructure that MUST be complete before apps/packages can be scaffolded

**⚠️ CRITICAL**: No app or package work can begin until this phase is complete

- [ ] T010 Create backend directory structure per plan.md in `backend/`
- [ ] T011 [P] Create Go module configuration in `backend/go.mod`
- [ ] T012 [P] Create placeholder server entrypoint in `backend/cmd/server/main.go`
- [ ] T013 [P] Create placeholder migration CLI in `backend/cmd/migrate/main.go`
- [ ] T014 Create all 16 internal domain module directories with .gitkeep in `backend/internal/`
- [ ] T015 [P] Create shared infrastructure directories in `backend/internal/shared/`
- [ ] T016 [P] Create API layer directories in `backend/api/graphql/` and `backend/api/rest/`
- [ ] T017 [P] Create plugin directories for billing, payments, analytics in `backend/plugins/`
- [ ] T018 Create migrations directory in `backend/migrations/`
- [ ] T019 [P] Create Docker Compose configuration with PostgreSQL, Redis, MinIO in `infrastructure/docker/docker-compose.yml`
- [ ] T020 [P] Create backend Dockerfile in `infrastructure/docker/Dockerfile.backend`
- [ ] T021 [P] Create dev environment config template in `infrastructure/config/dev.env.example`
- [ ] T022 [P] Create test environment config template in `infrastructure/config/test.env.example`
- [ ] T023 [P] Create staging environment config template in `infrastructure/config/staging.env.example`
- [ ] T024 [P] Create prod environment config template in `infrastructure/config/prod.env.example`
- [ ] T025 Create Kubernetes directory structure in `infrastructure/k8s/`
- [ ] T026 [P] Create tools directory structure in `tools/codegen/` and `tools/scripts/`
- [ ] T027 [P] Create docs directory structure in `docs/adr/`, `docs/events/`, `docs/api/`
- [ ] T028 [P] Create GitHub Actions CI workflow in `.github/workflows/ci.yml`

**Checkpoint**: Foundation ready - app and package scaffolding can begin

---

## Phase 3: User Story 1 - Developer Clones and Runs Project (Priority: P1) 🎯 MVP

**Goal**: Enable a developer to clone the repo and have a working environment within 10 minutes

**Independent Test**: Clone repo on fresh machine, run `make setup`, verify all services start

### Implementation for User Story 1

- [ ] T029 [US1] Create comprehensive README.md with quick start instructions in `./README.md`
- [ ] T030 [US1] Create Makefile with setup, dev, build, test, docker commands in `./Makefile`
- [ ] T031 [US1] Verify Docker Compose services start correctly with health checks
- [ ] T032 [US1] Create developer quickstart guide in `specs/001-init-monorepo/quickstart.md`

**Checkpoint**: Developer can clone, run `make setup`, and have working environment

---

## Phase 4: User Story 2 - Developer Builds Individual Packages (Priority: P2)

**Goal**: Enable independent build/test of individual packages with caching

**Independent Test**: Run `pnpm build --filter=@solobueno/types`, verify only types package builds

### Apps Scaffolding

- [ ] T033 [P] [US2] Create mobile app scaffold with package.json, tsconfig.json, src/index.ts in `apps/mobile/`
- [ ] T034 [P] [US2] Create kitchen-display app scaffold in `apps/kitchen-display/`
- [ ] T035 [P] [US2] Create backoffice app scaffold in `apps/backoffice/`
- [ ] T036 [P] [US2] Create admin app scaffold in `apps/admin/`

### Packages Scaffolding

- [ ] T037 [P] [US2] Create @solobueno/ui package with placeholder exports in `packages/ui/`
- [ ] T038 [P] [US2] Create @solobueno/i18n package with es-419.json and en.json locales in `packages/i18n/`
- [ ] T039 [P] [US2] Create @solobueno/types package with domain type definitions in `packages/types/`
- [ ] T040 [P] [US2] Create @solobueno/graphql package placeholder in `packages/graphql-client/`
- [ ] T041 [P] [US2] Create @solobueno/analytics package with event tracking utilities in `packages/analytics/`

### Build System Verification

- [ ] T042 [US2] Verify Turborepo caching works (run build twice, second should be cached)
- [ ] T043 [US2] Verify package filtering works with --filter flag
- [ ] T044 [US2] Verify incremental builds only rebuild changed packages

**Checkpoint**: Each app and package can be built independently

---

## Phase 5: User Story 3 - Developer Adds a New Package (Priority: P3)

**Goal**: Enable developers to add new packages following consistent patterns

**Independent Test**: Follow documented process to create a test package, verify it integrates with build

### Documentation for Adding Packages

- [ ] T045 [US3] Document package creation process in README.md (or separate CONTRIBUTING.md)
- [ ] T046 [US3] Ensure package template in data-model.md is accurate and complete
- [ ] T047 [US3] Verify new package integrates with workspace by testing package creation

**Checkpoint**: Developer can add new packages following documented process

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final quality improvements and validation

- [ ] T048 [P] Configure Husky pre-commit hooks in `.husky/pre-commit`
- [ ] T049 [P] Verify lint-staged configuration in package.json works correctly
- [ ] T050 Run `pnpm lint` and fix any issues
- [ ] T051 Run `pnpm format:check` and fix any formatting issues
- [ ] T052 Validate all success criteria from spec.md:
  - SC-001: Setup within 10 minutes
  - SC-002: Full build < 2 minutes
  - SC-003: Cached build < 30 seconds
  - SC-004: Independent package builds
  - SC-005: Folder structure matches constitution
  - SC-006: Lint/format pass with zero config
- [ ] T053 Final README review and documentation polish
- [ ] T054 Document edge cases in README.md:
  - Version mismatch detection (Node/Go version checks)
  - Network failure recovery (retry instructions for pnpm install)
  - Cross-OS compatibility (WSL2 requirements for Windows)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on T001-T003 (package.json, turbo.json, pnpm-workspace)
- **User Story 1 (Phase 3)**: Depends on Phase 2 completion
- **User Story 2 (Phase 4)**: Depends on Phase 2 completion (can parallel with US1)
- **User Story 3 (Phase 5)**: Depends on Phase 4 completion (needs packages to exist)
- **Polish (Phase 6)**: Depends on all user stories

### User Story Dependencies

- **User Story 1 (P1)**: Independent after Foundational phase
- **User Story 2 (P2)**: Independent after Foundational phase (can parallel with US1)
- **User Story 3 (P3)**: Depends on US2 (needs packages to document)

### Within Each Phase

- Tasks marked [P] can run in parallel
- Sequential tasks should complete in order listed

### Parallel Opportunities

**Phase 1 (all parallel):**

```
T001, T002, T003, T004, T005, T006, T007, T008, T009
```

**Phase 2 (parallel groups):**

```
Group A: T011, T012, T013 (Go files)
Group B: T019, T020 (Docker files)
Group C: T021, T022, T023, T024 (env configs)
Group D: T026, T027, T028 (tools, docs, CI)
```

**Phase 4 (apps parallel, packages parallel):**

```
Apps: T033, T034, T035, T036
Packages: T037, T038, T039, T040, T041
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (9 tasks)
2. Complete Phase 2: Foundational (19 tasks)
3. Complete Phase 3: User Story 1 (4 tasks)
4. **STOP and VALIDATE**: Can a developer clone and run the project?
5. This delivers a working monorepo structure

### Incremental Delivery

1. Setup + Foundational → Basic structure ready
2. Add US1 → Clone-and-run works → **MVP Complete**
3. Add US2 → Individual builds work → Enhanced developer experience
4. Add US3 → Self-documenting process → Full feature complete
5. Polish → Production-ready quality

### Parallel Team Strategy

With 2+ developers after Foundational phase:

- Developer A: User Story 1 (README, Makefile, validation)
- Developer B: User Story 2 (Apps + Packages scaffolding)

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Commit after each task or logical group
- Verify builds work after each checkpoint
- Total: 54 tasks across 6 phases
