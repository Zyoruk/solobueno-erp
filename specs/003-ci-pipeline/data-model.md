# Data Model: CI/CD Pipeline

**Feature**: 003-ci-pipeline  
**Date**: 2025-01-29

## Overview

This feature does not introduce database entities. Instead, it defines GitHub Actions workflow configurations and Dependabot settings.

## Workflow: ci.yml

**File**: `.github/workflows/ci.yml`

```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: ${{ github.ref != 'refs/heads/main' }}

env:
  NODE_VERSION: '20'
  GO_VERSION: '1.22'
  PNPM_VERSION: '8'

jobs:
  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup pnpm
        uses: pnpm/action-setup@v2
        with:
          version: ${{ env.PNPM_VERSION }}

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}
          cache: 'pnpm'

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true
          cache-dependency-path: backend/go.sum

      - name: Install dependencies
        run: pnpm install --frozen-lockfile

      - name: Lint TypeScript
        run: pnpm lint

      - name: Lint Go
        working-directory: backend
        run: |
          go vet ./...
          go install honnef.co/go/tools/cmd/staticcheck@latest
          staticcheck ./...

  build:
    name: Build
    runs-on: ubuntu-latest
    needs: lint
    steps:
      - uses: actions/checkout@v4

      - name: Setup pnpm
        uses: pnpm/action-setup@v2
        with:
          version: ${{ env.PNPM_VERSION }}

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}
          cache: 'pnpm'

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true
          cache-dependency-path: backend/go.sum

      - name: Setup Turborepo cache
        uses: actions/cache@v4
        with:
          path: .turbo
          key: turbo-${{ github.ref }}-${{ github.sha }}
          restore-keys: |
            turbo-${{ github.ref }}-
            turbo-

      - name: Install dependencies
        run: pnpm install --frozen-lockfile

      - name: Build TypeScript packages
        run: pnpm build

      - name: Build Go backend
        working-directory: backend
        run: go build -v ./...

  test:
    name: Test
    runs-on: ubuntu-latest
    needs: build
    steps:
      - uses: actions/checkout@v4

      - name: Setup pnpm
        uses: pnpm/action-setup@v2
        with:
          version: ${{ env.PNPM_VERSION }}

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}
          cache: 'pnpm'

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
          cache: true
          cache-dependency-path: backend/go.sum

      - name: Install dependencies
        run: pnpm install --frozen-lockfile

      - name: Test TypeScript with coverage
        run: pnpm test --coverage

      - name: Test Go with coverage
        working-directory: backend
        run: |
          go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
          go tool cover -func=coverage.out

      - name: Check Go coverage threshold
        working-directory: backend
        run: |
          COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print substr($3, 1, length($3)-1)}')
          echo "Go coverage: ${COVERAGE}%"
          if (( $(echo "$COVERAGE < 80" | bc -l) )); then
            echo "Coverage ${COVERAGE}% is below 80% threshold"
            exit 1
          fi

      - name: Upload coverage artifacts
        uses: actions/upload-artifact@v4
        with:
          name: coverage-reports
          path: |
            coverage/
            backend/coverage.out
```

## Workflow: codeql.yml

**File**: `.github/workflows/codeql.yml`

```yaml
name: CodeQL

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]
  schedule:
    - cron: '0 6 * * 1' # Weekly on Monday at 6am UTC

jobs:
  analyze:
    name: Analyze
    runs-on: ubuntu-latest
    permissions:
      actions: read
      contents: read
      security-events: write

    strategy:
      fail-fast: false
      matrix:
        language: ['javascript-typescript', 'go']

    steps:
      - name: Checkout repository
        uses: actions/checkout@v4

      - name: Initialize CodeQL
        uses: github/codeql-action/init@v3
        with:
          languages: ${{ matrix.language }}
          queries: security-extended

      - name: Autobuild
        uses: github/codeql-action/autobuild@v3

      - name: Perform CodeQL Analysis
        uses: github/codeql-action/analyze@v3
        with:
          category: '/language:${{ matrix.language }}'
```

## Dependabot Configuration

**File**: `.github/dependabot.yml`

```yaml
version: 2
updates:
  # npm packages
  - package-ecosystem: 'npm'
    directory: '/'
    schedule:
      interval: 'weekly'
      day: 'monday'
    groups:
      production:
        patterns:
          - '*'
        exclude-patterns:
          - '@types/*'
          - 'eslint*'
          - 'prettier*'
          - 'typescript'
          - 'vitest'
      development:
        patterns:
          - '@types/*'
          - 'eslint*'
          - 'prettier*'
          - 'typescript'
          - 'vitest'
    open-pull-requests-limit: 10

  # Go modules
  - package-ecosystem: 'gomod'
    directory: '/backend'
    schedule:
      interval: 'weekly'
      day: 'monday'
    open-pull-requests-limit: 5

  # GitHub Actions
  - package-ecosystem: 'github-actions'
    directory: '/'
    schedule:
      interval: 'weekly'
      day: 'monday'
    open-pull-requests-limit: 5
```

## Branch Protection Rules

**Configured via GitHub UI or API**

```json
{
  "required_status_checks": {
    "strict": true,
    "contexts": ["Lint", "Build", "Test", "Analyze (javascript-typescript)", "Analyze (go)"]
  },
  "enforce_admins": false,
  "required_pull_request_reviews": {
    "dismiss_stale_reviews": true,
    "require_code_owner_reviews": false,
    "required_approving_review_count": 1
  },
  "restrictions": null,
  "required_linear_history": false,
  "allow_force_pushes": false,
  "allow_deletions": false
}
```

## Directory Structure

```text
.github/
├── workflows/
│   ├── ci.yml              # Main CI workflow
│   └── codeql.yml          # Security scanning
└── dependabot.yml          # Dependency updates
```

## Job Dependencies

```
ci.yml:
  lint ─────────► build ─────────► test
                    │
                    └─────────────► (artifacts uploaded)

codeql.yml:
  analyze (javascript-typescript) ──► (parallel)
  analyze (go) ─────────────────────► (parallel)
```

## Cache Keys

| Cache      | Key                                                                 | TTL    |
| ---------- | ------------------------------------------------------------------- | ------ |
| pnpm store | `pnpm-store-${{ runner.os }}-${{ hashFiles('**/pnpm-lock.yaml') }}` | 7 days |
| Go modules | Built into actions/setup-go                                         | 7 days |
| Turborepo  | `turbo-${{ github.ref }}-${{ github.sha }}`                         | 7 days |
