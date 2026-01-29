# Tasks: Docker Local Development Environment

**Input**: Design documents from `/specs/002-docker-local-dev/`  
**Prerequisites**: plan.md ✓, spec.md ✓, research.md ✓, data-model.md ✓, quickstart.md ✓

**Tests**: Not explicitly requested in feature specification. Manual verification via connectivity tests and health check scripts.

**Organization**: Tasks grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup

**Purpose**: Create required directory structure (minimal - most exists from 001-init-monorepo)

- [x] T001 Create scripts directory at `infrastructure/scripts/`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T002 Enhance docker-compose.yml with named network `solobueno-network` in `infrastructure/docker/docker-compose.yml`
- [x] T003 Update docker-compose.yml volumes to use explicit naming (`solobueno_postgres_data`, `solobueno_redis_data`, `solobueno_minio_data`) in `infrastructure/docker/docker-compose.yml`
- [x] T004 Add `start_period` to all health checks in `infrastructure/docker/docker-compose.yml`
- [x] T005 Update PostgreSQL health check to include database name (`pg_isready -U solobueno -d solobueno_dev`) in `infrastructure/docker/docker-compose.yml`
- [x] T006 Create health-check.sh script with PostgreSQL, Redis, and MinIO checks at `infrastructure/scripts/health-check.sh`
- [x] T007 Make health-check.sh executable with `chmod +x infrastructure/scripts/health-check.sh`
- [x] T008 Define DOCKER_COMPOSE variable in Makefile at `Makefile`

**Checkpoint**: Foundation ready - user story implementation can now begin

---

## Phase 3: User Story 1 - Developer Starts Local Services (Priority: P1) 🎯 MVP

**Goal**: Start all required backend services (PostgreSQL, Redis, MinIO) with a single command within 60 seconds

**Independent Test**: Run `make docker-up` and verify all services are accessible on their ports (5432, 6379, 9000, 9001)

### Implementation for User Story 1

- [x] T009 [US1] Enhance docker-up target with Docker installation check in `Makefile`
- [x] T010 [US1] Enhance docker-up target with Docker daemon running check in `Makefile`
- [x] T011 [US1] Add port availability check (5432, 6379, 9000, 9001) with helpful error messages in `Makefile`
- [x] T012 [US1] Add docker-status target that shows `docker compose ps` output in `Makefile`
- [x] T013 [US1] Add docker-health target that runs health-check.sh in `Makefile`
- [x] T014 [US1] Update docker-up target to call docker-status after starting services in `Makefile`

**Checkpoint**: User Story 1 complete - developers can start all services with `make docker-up`

---

## Phase 4: User Story 2 - Developer Manages Service Lifecycle (Priority: P2)

**Goal**: Stop, restart, and reset services easily while managing data persistence appropriately

**Independent Test**: Run lifecycle commands (`make docker-down`, `make docker-restart`, `make docker-reset`) and verify service states and data persistence

### Implementation for User Story 2

- [x] T015 [US2] Add docker-restart target that restarts all services in `Makefile`
- [x] T016 [US2] Enhance docker-reset target with warning message and 3-second delay in `Makefile`
- [x] T017 [US2] Update docker-reset to output confirmation message after reset in `Makefile`

**Checkpoint**: User Story 2 complete - developers can manage service lifecycle with `make docker-down/restart/reset`

---

## Phase 5: User Story 3 - Developer Views Service Logs (Priority: P3)

**Goal**: View logs from any service for debugging and diagnostics

**Independent Test**: Generate activity in services and verify logs appear with `make docker-logs` and service-specific log commands

### Implementation for User Story 3

- [x] T018 [P] [US3] Add docker-logs-postgres target for PostgreSQL logs in `Makefile`
- [x] T019 [P] [US3] Add docker-logs-redis target for Redis logs in `Makefile`
- [x] T020 [P] [US3] Add docker-logs-minio target for MinIO logs in `Makefile`

**Checkpoint**: User Story 3 complete - developers can view logs with `make docker-logs[-service]`

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Additional convenience features and documentation

- [x] T021 [P] Add docker-shell-postgres target to open psql shell in `Makefile`
- [x] T022 [P] Update .PHONY declaration with all new targets in `Makefile`
- [x] T023 Update infrastructure/config/dev.env.example with connection strings for all services
- [x] T024 Run quickstart.md validation (verify all documented commands work) - All targets present in `make help`
- [ ] T025 Test data persistence across 10 stop/start cycles per SC-003 - Requires Docker running

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-5)**: All depend on Foundational phase completion
  - US1, US2, US3 can proceed sequentially (P1 → P2 → P3)
  - Some tasks within US3 can run in parallel
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - Uses docker-status from US1 but can be implemented independently
- **User Story 3 (P3)**: Can start after Foundational (Phase 2) - Independent of US1/US2

### Within Each Phase

- Foundational tasks T002-T005 all modify same file (docker-compose.yml) - do sequentially
- T006-T007 (health-check.sh) can run parallel to docker-compose.yml updates
- US3 tasks T018-T020 are all parallelizable (different targets, no dependencies)
- Polish tasks T021-T022 are parallelizable (different targets/sections)

### Parallel Opportunities

```text
Phase 2 (Foundational):
  Sequential: T002 → T003 → T004 → T005 (same file)
  Parallel:   T006, T007 (can run with above)
              T008 (separate file, can run parallel)

Phase 5 (US3):
  Parallel: T018, T019, T020 (all different Makefile targets)

Phase 6 (Polish):
  Parallel: T021, T022 (different targets/sections)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001)
2. Complete Phase 2: Foundational (T002-T008)
3. Complete Phase 3: User Story 1 (T009-T014)
4. **STOP and VALIDATE**: Run `make docker-up` and verify:
   - All containers start within 60 seconds
   - PostgreSQL accessible on port 5432
   - Redis accessible on port 6379
   - MinIO accessible on ports 9000/9001
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → MVP complete!
3. Add User Story 2 → Test lifecycle commands
4. Add User Story 3 → Test log viewing
5. Add Polish → Full feature complete

### Estimated Task Counts

| Phase        | Tasks  | Description                                                         |
| ------------ | ------ | ------------------------------------------------------------------- |
| Setup        | 1      | Directory creation                                                  |
| Foundational | 7      | docker-compose.yml enhancements, health-check.sh, Makefile variable |
| US1 (P1)     | 6      | Start services with pre-flight checks, port availability            |
| US2 (P2)     | 3      | Lifecycle management (restart, reset)                               |
| US3 (P3)     | 3      | Log viewing commands                                                |
| Polish       | 5      | Shell access, documentation, validation                             |
| **Total**    | **25** |                                                                     |

---

## Notes

- Most infrastructure exists from 001-init-monorepo - this feature enhances it
- [P] tasks = different files or different Makefile targets, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Edge cases (Docker not running, ports in use) are handled by pre-flight checks in US1
