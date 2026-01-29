# Branch Protection Rules

**Feature**: 003-ci-pipeline  
**Target Branch**: `main`

## Overview

Branch protection rules prevent broken code from being merged to main and ensure all changes go through proper review and CI validation.

## Required Configuration

Configure these rules via GitHub UI: **Settings → Branches → Add branch protection rule**

### Rule: `main`

| Setting                                   | Value  | Reason                           |
| ----------------------------------------- | ------ | -------------------------------- |
| **Require a pull request before merging** | ✅ Yes | Enforces code review             |
| Required approving reviews                | 1      | Balance between speed and safety |
| Dismiss stale pull request approvals      | ✅ Yes | Re-review after changes          |
| Require review from Code Owners           | ❌ No  | Optional for small teams         |
| **Require status checks to pass**         | ✅ Yes | Ensures CI passes before merge   |
| Require branches to be up to date         | ✅ Yes | Prevents merge conflicts         |
| **Require conversation resolution**       | ❌ No  | Optional                         |
| **Require signed commits**                | ❌ No  | Optional                         |
| **Require linear history**                | ❌ No  | Allows merge commits             |
| **Include administrators**                | ❌ No  | Allows emergency fixes           |
| **Restrict who can push**                 | ✅ Yes | Prevents direct pushes           |
| **Allow force pushes**                    | ❌ No  | Protects commit history          |
| **Allow deletions**                       | ❌ No  | Protects branch                  |

## Required Status Checks

The following status checks MUST pass before a PR can be merged:

### CI Workflow (`ci.yml`)

| Check Name | Job     | Description                            |
| ---------- | ------- | -------------------------------------- |
| **Lint**   | `lint`  | TypeScript + Go linting                |
| **Build**  | `build` | TypeScript + Go compilation            |
| **Test**   | `test`  | Unit tests with 80% coverage threshold |

### CodeQL Workflow (`codeql.yml`)

| Check Name                          | Job       | Description             |
| ----------------------------------- | --------- | ----------------------- |
| **Analyze (javascript-typescript)** | `analyze` | Security scan for JS/TS |
| **Analyze (go)**                    | `analyze` | Security scan for Go    |

## How to Configure

### Via GitHub UI

1. Go to repository **Settings**
2. Click **Branches** in the left sidebar
3. Click **Add branch protection rule**
4. Enter `main` as the branch name pattern
5. Configure settings as described above
6. Under "Require status checks to pass before merging":
   - Search and select: `Lint`, `Build`, `Test`
   - Search and select: `Analyze (javascript-typescript)`, `Analyze (go)`
7. Click **Create** or **Save changes**

### Via GitHub API

```bash
curl -X PUT \
  -H "Authorization: token $GITHUB_TOKEN" \
  -H "Accept: application/vnd.github.v3+json" \
  https://api.github.com/repos/OWNER/REPO/branches/main/protection \
  -d '{
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
  }'
```

## Emergency Bypass

In case of critical issues when CI is down:

1. Repository admins can merge PRs with failing checks
2. Use the **"Merge without waiting for requirements"** option
3. Document the bypass in the PR description
4. Fix any issues immediately after merge

## Verification

To verify branch protection is working:

1. Create a PR with a failing test
2. Attempt to merge → Should be blocked
3. Fix the test and push
4. Wait for CI to pass → Merge should be allowed

## Related Files

- `.github/workflows/ci.yml` - Main CI workflow
- `.github/workflows/codeql.yml` - Security scanning
- `.github/dependabot.yml` - Dependency updates
- `.github/CODEOWNERS` - Code ownership
