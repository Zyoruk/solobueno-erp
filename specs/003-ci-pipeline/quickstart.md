# Quickstart: CI/CD Pipeline

**Feature**: 003-ci-pipeline  
**Audience**: Developers working with the CI pipeline

## Overview

The CI pipeline automatically runs on every pull request and push to main. It includes:

- **Lint**: ESLint + Prettier (TypeScript), go vet + staticcheck (Go)
- **Build**: Turborepo build (TypeScript), go build (Go)
- **Test**: vitest (TypeScript), go test (Go) with 80% coverage threshold
- **Security**: CodeQL scanning, Dependabot updates

## Quick Start

### 1. Create a Pull Request

```bash
# Create feature branch
git checkout -b feature/my-feature

# Make changes
# ...

# Commit and push
git add .
git commit -m "feat: add my feature"
git push -u origin feature/my-feature
```

Open a PR on GitHub. CI will automatically trigger.

### 2. Check CI Status

Go to the PR page on GitHub. You'll see:

| Check      | Description                 | Required |
| ---------- | --------------------------- | -------- |
| **Lint**   | Code formatting and linting | Yes      |
| **Build**  | Compilation of all packages | Yes      |
| **Test**   | Unit tests with coverage    | Yes      |
| **CodeQL** | Security scanning           | Yes      |

All checks must pass before merging.

### 3. Fix CI Failures

**Lint failures**:

```bash
# Auto-fix formatting
pnpm format

# Check linting issues
pnpm lint
```

**Build failures**:

```bash
# Build all packages
pnpm build

# Build Go backend
cd backend && go build ./...
```

**Test failures**:

```bash
# Run tests locally
pnpm test

# Run Go tests
cd backend && go test ./...
```

**Coverage below 80%**:

```bash
# Check coverage report
pnpm test --coverage

# For Go
cd backend && go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out  # Opens browser
```

## Running CI Locally

### Lint

```bash
# TypeScript
pnpm lint

# Go
cd backend
go vet ./...
staticcheck ./...  # Install: go install honnef.co/go/tools/cmd/staticcheck@latest
```

### Build

```bash
# TypeScript (with Turborepo caching)
pnpm build

# Go
cd backend && go build -v ./...
```

### Test with Coverage

```bash
# TypeScript
pnpm test --coverage

# Go
cd backend
go test -v -race -coverprofile=coverage.out ./...
go tool cover -func=coverage.out  # Shows coverage %
```

## Dependabot Updates

Dependabot creates PRs weekly (Monday) for:

- npm packages (grouped by production/development)
- Go modules
- GitHub Actions versions

### Handling Dependabot PRs

1. Review the changelog/release notes
2. Ensure CI passes
3. Merge if safe

For security updates, Dependabot may create immediate PRs.

## CodeQL Security Alerts

CodeQL scans code for security vulnerabilities. Alerts appear in:

- **Security tab** on GitHub
- **PR checks** (for new code)

### Reviewing Alerts

1. Go to Security → Code scanning alerts
2. Review each alert
3. Fix or dismiss with reason

## Branch Protection

The `main` branch has protections:

| Rule               | Setting |
| ------------------ | ------- |
| Require PR         | Yes     |
| Required approvals | 1       |
| Require CI to pass | Yes     |
| Up-to-date branch  | Yes     |
| Direct push        | Blocked |
| Force push         | Blocked |

### Merging a PR

1. All CI checks pass (green)
2. At least 1 approval
3. Branch is up to date with main
4. Click "Merge pull request"

### Emergency Override

Admins can bypass protections for emergencies:

1. Go to PR → Merge button dropdown
2. Select "Merge without waiting for requirements"
3. Document reason in PR comment

## Troubleshooting

### CI is slow (>5 minutes)

- Check if cache is being used (look for "Cache hit" in logs)
- Turborepo cache may be cold after main branch changes
- First run after pnpm-lock.yaml change will be slower

### "Coverage below 80%" error

Add more tests! Check which files lack coverage:

```bash
# TypeScript
pnpm test --coverage
# Look at coverage/lcov-report/index.html

# Go
go tool cover -html=coverage.out
```

### CodeQL taking too long

CodeQL runs separately from main CI. It doesn't block PR merge unless findings are critical.

### Dependabot PR conflicts

```bash
# Rebase Dependabot PR
git fetch origin
git checkout dependabot/npm_and_yarn/...
git rebase origin/main
git push --force-with-lease
```

## File Locations

| File                                         | Purpose                            |
| -------------------------------------------- | ---------------------------------- |
| `.github/workflows/ci.yml`                   | Main CI workflow (lint/build/test) |
| `.github/workflows/codeql.yml`               | Security scanning (JS/TS + Go)     |
| `.github/dependabot.yml`                     | Dependency updates (weekly)        |
| `.github/CODEOWNERS`                         | Code review assignments            |
| `specs/003-ci-pipeline/BRANCH_PROTECTION.md` | Branch protection setup guide      |
