# Tasks: CI/CD Pipeline

**Input**: Design documents from `/specs/003-ci-pipeline/`  
**Prerequisites**: plan.md ✓, spec.md ✓, research.md ✓, data-model.md ✓, quickstart.md ✓

**Tests**: Not explicitly requested. Manual verification by creating test PRs.

**Organization**: Tasks grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup

**Purpose**: Create required directory structure for GitHub Actions

- [x] T001 Create .github/workflows directory at `.github/workflows/`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core workflow structure that all user stories depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T002 Create ci.yml workflow file with name, triggers (push/PR to main), and environment variables at `.github/workflows/ci.yml`
- [x] T003 Add concurrency configuration to cancel in-progress runs on same PR in `.github/workflows/ci.yml`

**Checkpoint**: Foundation ready - CI workflow skeleton exists

---

## Phase 3: User Story 1 - Developer Gets Automated Feedback on PR (Priority: P1) 🎯 MVP

**Goal**: Automated lint, build, and test checks run on every PR and report status

**Independent Test**: Open a PR with intentional lint/test errors and verify checks fail with clear messages

### Implementation for User Story 1

- [x] T004 [US1] Add lint job with checkout, pnpm setup, Node.js setup in `.github/workflows/ci.yml`
- [x] T005 [US1] Add Go setup to lint job with cache configuration in `.github/workflows/ci.yml`
- [x] T006 [US1] Add pnpm install and lint commands to lint job in `.github/workflows/ci.yml`
- [x] T007 [US1] Add Go vet and staticcheck commands to lint job in `.github/workflows/ci.yml`
- [x] T008 [US1] Add build job with needs: lint dependency in `.github/workflows/ci.yml`
- [x] T009 [US1] Add pnpm build and go build commands to build job in `.github/workflows/ci.yml`
- [x] T010 [US1] Add test job with needs: build dependency in `.github/workflows/ci.yml`
- [x] T011 [US1] Add TypeScript test with coverage command to test job in `.github/workflows/ci.yml`
- [x] T012 [US1] Add Go test with coverage and threshold check to test job in `.github/workflows/ci.yml`
- [x] T013 [US1] Add coverage artifact upload step to test job in `.github/workflows/ci.yml`

**Checkpoint**: User Story 1 complete - PRs trigger lint→build→test pipeline

---

## Phase 4: User Story 2 - Developer Sees Build Results Quickly (Priority: P2)

**Goal**: Pipeline completes within 5 minutes (cached) / 10 minutes (clean) through caching

**Independent Test**: Run CI twice on same branch and measure time improvement

### Implementation for User Story 2

- [x] T014 [US2] Add Turborepo cache configuration with restore-keys to build job in `.github/workflows/ci.yml`
- [x] T015 [US2] Verify pnpm cache is configured in setup-node action in `.github/workflows/ci.yml`
- [x] T016 [US2] Verify Go module cache is configured in setup-go action in `.github/workflows/ci.yml`

**Checkpoint**: User Story 2 complete - Subsequent runs use cached dependencies

---

## Phase 5: User Story 3 - Main Branch Stays Deployable (Priority: P1)

**Goal**: Branch protection prevents merging without passing CI + approval

**Independent Test**: Attempt to merge failing PR and verify it's blocked

### Implementation for User Story 3

- [x] T017 [P] [US3] Create dependabot.yml with npm, gomod, and github-actions ecosystems at `.github/dependabot.yml`
- [x] T018 [P] [US3] Create codeql.yml workflow with JS/TS and Go language matrix at `.github/workflows/codeql.yml`
- [x] T019 [P] [US3] Create CODEOWNERS file for code review assignments at `.github/CODEOWNERS`
- [x] T020 [US3] Document branch protection rules configuration in `specs/003-ci-pipeline/BRANCH_PROTECTION.md`
- [x] T021 [US3] Add required status checks list to branch protection documentation

**Checkpoint**: User Story 3 complete - Security scanning configured, protection rules documented

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Finalize configuration and verify end-to-end functionality

- [x] T022 [P] Verify ci.yml syntax with actionlint or GitHub Actions validator
- [x] T023 [P] Verify codeql.yml syntax with actionlint or GitHub Actions validator
- [x] T024 [P] Verify dependabot.yml syntax via GitHub documentation
- [ ] T025 Test full CI pipeline by creating a test PR (manual - requires push to GitHub)
- [ ] T026 Verify coverage threshold enforcement by intentionally lowering coverage (manual - requires GitHub)
- [ ] T027 Verify cache hit rate >80% on second CI run per SC-004 (manual - requires GitHub)
- [x] T028 Update quickstart.md with actual workflow file locations

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup - BLOCKS all user stories
- **User Story 1 (Phase 3)**: Depends on Foundational - Creates core CI workflow
- **User Story 2 (Phase 4)**: Depends on US1 - Adds caching to existing workflow
- **User Story 3 (Phase 5)**: Depends on Foundational - Can run parallel to US1/US2 for separate files
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Core pipeline - no dependencies on other stories
- **User Story 2 (P2)**: Depends on US1 (adds caching to existing jobs)
- **User Story 3 (P1)**: Independent of US1/US2 (separate files: dependabot.yml, codeql.yml)

### Within Each Phase

- US1 tasks must be sequential (building up ci.yml incrementally)
- US3 tasks T017, T018, T019 can run in parallel (different files)
- Polish tasks T022-T024 can run in parallel (different validations)

### Parallel Opportunities

```text
Phase 5 (US3):
  Parallel: T017 (dependabot.yml), T018 (codeql.yml), T019 (CODEOWNERS)
  Sequential: T020 → T021 (documentation)

Phase 6 (Polish):
  Parallel: T022, T023, T024 (syntax validation)
  Sequential: T025 → T026 → T027 → T028 (integration testing)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001)
2. Complete Phase 2: Foundational (T002-T003)
3. Complete Phase 3: User Story 1 (T004-T013)
4. **STOP and VALIDATE**: Create a test PR and verify:
   - Lint job runs and reports status
   - Build job runs after lint passes
   - Test job runs with coverage reporting
   - Failed checks block PR merge
5. Deploy if ready

### Incremental Delivery

1. Complete Setup + Foundational → CI workflow skeleton
2. Add User Story 1 → Core lint/build/test pipeline (MVP!)
3. Add User Story 2 → Caching for faster builds
4. Add User Story 3 → Security scanning + branch protection
5. Add Polish → Validation and documentation

### Estimated Task Counts

| Phase        | Tasks  | Description                          |
| ------------ | ------ | ------------------------------------ |
| Setup        | 1      | Directory creation                   |
| Foundational | 2      | Workflow triggers, concurrency       |
| US1 (P1)     | 10     | Lint, build, test jobs               |
| US2 (P2)     | 3      | Caching configuration                |
| US3 (P1)     | 5      | Dependabot, CodeQL, CODEOWNERS, docs |
| Polish       | 7      | Validation, testing, cache verify    |
| **Total**    | **28** |                                      |

---

## Notes

- All CI configuration lives in `.github/` directory
- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Branch protection must be configured manually via GitHub UI/API after CI is working
- CodeQL requires repository to have GitHub Advanced Security enabled (free for public repos)
- Test the pipeline by creating actual PRs - no automated tests for CI itself
