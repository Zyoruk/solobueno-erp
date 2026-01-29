# Contracts: CI/CD Pipeline

**Feature**: 003-ci-pipeline

## Overview

This feature is a DevOps/Infrastructure feature that does not expose APIs. Therefore, there are no API contracts to define.

## CI Contracts

Instead of API contracts, this feature defines **CI contracts** - the expected behavior of CI workflows:

### CI Workflow Contract

| Trigger           | Expected Behavior        |
| ----------------- | ------------------------ |
| PR opened/updated | Lint → Build → Test runs |
| Push to main      | Lint → Build → Test runs |
| Weekly schedule   | CodeQL full scan         |

### Job Status Contract

| Job    | Success Criteria                          |
| ------ | ----------------------------------------- |
| Lint   | Zero ESLint errors, zero go vet errors    |
| Build  | All packages compile without errors       |
| Test   | All tests pass, coverage ≥80%             |
| CodeQL | Analysis completes (findings don't block) |

### Cache Contract

| Cache      | Expected Behavior                |
| ---------- | -------------------------------- |
| pnpm store | Restored on pnpm-lock.yaml match |
| Go modules | Restored on go.sum match         |
| Turborepo  | Restored on branch/SHA match     |

### Timing Contract

| Scenario     | Max Duration |
| ------------ | ------------ |
| Cached build | 5 minutes    |
| Clean build  | 10 minutes   |

### Branch Protection Contract

| Action              | Expected Result            |
| ------------------- | -------------------------- |
| PR with failing CI  | Cannot merge               |
| PR without approval | Cannot merge               |
| Direct push to main | Rejected                   |
| Admin override      | Allowed with documentation |
