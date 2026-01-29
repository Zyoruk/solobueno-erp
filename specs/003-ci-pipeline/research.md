# Research: CI/CD Pipeline

**Feature**: 003-ci-pipeline  
**Date**: 2025-01-29

## Research Tasks

### 1. GitHub Actions Workflow Structure for Monorepo

**Decision**: Single main CI workflow with parallel jobs for Go and TypeScript

**Rationale**:

- Single workflow file is easier to maintain
- Parallel jobs maximize speed
- Turborepo handles incremental builds within TypeScript packages

**Workflow Structure**:

```yaml
ci.yml:
  - Job: lint (parallel)
    - pnpm lint (TypeScript/ESLint)
    - go vet + staticcheck (Go)
  - Job: build (parallel, depends on lint)
    - pnpm build (Turborepo)
    - go build ./... (Go)
  - Job: test (parallel, depends on build)
    - pnpm test --coverage (vitest)
    - go test -coverprofile (Go)
  - Job: coverage-report (depends on test)
    - Combine coverage, check threshold
```

**Alternatives Considered**:

- Separate workflows per language: More files to maintain, harder to coordinate
- Single sequential job: Slower, no parallelization

### 2. Caching Strategy for Fast Builds

**Decision**: Multi-layer caching with pnpm store, Go modules, and Turborepo cache

**Rationale**:

- pnpm store caching avoids re-downloading packages
- Go modules cache speeds up dependency resolution
- Turborepo remote cache enables incremental builds across PRs

**Cache Configuration**:

| Cache      | Key Pattern                                  | Restore Keys                         | Expected Hit Rate |
| ---------- | -------------------------------------------- | ------------------------------------ | ----------------- |
| pnpm store | `pnpm-${{ hashFiles('**/pnpm-lock.yaml') }}` | `pnpm-`                              | >90%              |
| Go modules | `go-${{ hashFiles('**/go.sum') }}`           | `go-`                                | >90%              |
| Turborepo  | `turbo-${{ github.ref }}-${{ github.sha }}`  | `turbo-${{ github.ref }}-`, `turbo-` | >80%              |

**Alternatives Considered**:

- No caching: Builds would take 15+ minutes
- Docker layer caching: Overkill for this use case

### 3. Code Coverage Tooling

**Decision**: Use native coverage tools (vitest + go test) with combined threshold check

**Rationale**:

- vitest has built-in coverage via c8/istanbul
- go test has native -coverprofile
- Simple threshold check in workflow avoids external services

**Coverage Strategy**:

| Component  | Tool                  | Output Format | Threshold |
| ---------- | --------------------- | ------------- | --------- |
| TypeScript | vitest --coverage     | lcov          | 80%       |
| Go         | go test -coverprofile | text          | 80%       |

**Threshold Enforcement**:

```bash
# TypeScript: vitest exits non-zero if below threshold
vitest --coverage --coverage.thresholds.lines=80

# Go: parse coverage percentage from go tool cover
go tool cover -func=coverage.out | grep total | awk '{print $3}' | check >= 80
```

**Alternatives Considered**:

- Codecov/Coveralls: Adds external dependency, free tier limitations
- SonarQube: Overkill for initial setup

### 4. CodeQL Configuration

**Decision**: Use default CodeQL setup for JavaScript/TypeScript and Go

**Rationale**:

- GitHub's CodeQL is free for public repos and included in GitHub Advanced Security
- Default queries catch common security issues
- Can be extended with custom queries later

**Languages to Scan**:

| Language              | Query Suite       | Trigger          |
| --------------------- | ----------------- | ---------------- |
| javascript-typescript | security-extended | PR, push to main |
| go                    | security-extended | PR, push to main |

**Alternatives Considered**:

- Snyk: Commercial, adds cost
- Semgrep: Good alternative but CodeQL is already integrated

### 5. Dependabot Configuration

**Decision**: Weekly updates for npm and Go modules, grouped PRs

**Rationale**:

- Weekly cadence balances security with PR noise
- Grouped updates reduce number of PRs
- Auto-merge for patch versions with passing CI

**Configuration**:

| Ecosystem      | Directory | Schedule        | Groups                  |
| -------------- | --------- | --------------- | ----------------------- |
| npm            | /         | weekly (Monday) | production, development |
| gomod          | /backend  | weekly (Monday) | all                     |
| github-actions | /         | weekly (Monday) | N/A                     |

**Alternatives Considered**:

- Daily updates: Too noisy
- Renovate: More features but Dependabot is built-in

### 6. Branch Protection Rules

**Decision**: Require CI passage + 1 approval for main branch

**Rationale**:

- Prevents broken code from being merged
- Single approval keeps velocity high for small team
- Admin bypass allows emergency fixes

**Rules**:

| Rule                        | Value                  |
| --------------------------- | ---------------------- |
| Require PR before merge     | Yes                    |
| Required approvals          | 1                      |
| Dismiss stale approvals     | Yes                    |
| Require status checks       | ci (lint, build, test) |
| Require branches up to date | Yes                    |
| Restrict pushes to main     | Yes (admins exempt)    |
| Allow force pushes          | No                     |
| Allow deletions             | No                     |

**Alternatives Considered**:

- 2 required approvals: Too slow for small team
- No approval required: Risky for code quality

### 7. Concurrency and Cancellation

**Decision**: Cancel in-progress runs on same PR, allow parallel PR runs

**Rationale**:

- Cancelling stale runs saves CI minutes
- Different PRs should run independently
- Main branch runs should never be cancelled

**Concurrency Groups**:

```yaml
concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: ${{ github.ref != 'refs/heads/main' }}
```

**Alternatives Considered**:

- No concurrency control: Wastes CI minutes on outdated commits
- Global concurrency: Would queue PRs unnecessarily

## Summary of Decisions

| Area               | Decision                            |
| ------------------ | ----------------------------------- |
| Workflow Structure | Single ci.yml with parallel jobs    |
| Caching            | pnpm store + Go modules + Turborepo |
| Coverage           | vitest + go test, 80% threshold     |
| Security Scanning  | CodeQL (JS/TS + Go)                 |
| Dependencies       | Dependabot weekly, grouped PRs      |
| Branch Protection  | CI required + 1 approval            |
| Concurrency        | Cancel in-progress on same PR       |

## Open Questions Resolved

All technical decisions made based on GitHub Actions best practices and spec clarifications. No outstanding questions.
