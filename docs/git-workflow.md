# Git Workflow

This project uses atomic changes and atomic PRs.

## Rules

- One focused change per commit and per PR.
- Commit message prefix is required.
- PR title should match the same prefix and scope.
- Avoid mixing refactors, features, and bug fixes in one PR.

## Prefix Convention

Use one of these prefixes:

- `[feat]` new user-facing capability
- `[bug]` bug fix or behavior correction
- `[docs]` documentation-only change
- `[chore]` maintenance or tooling updates
- `[refactor]` internal restructuring without behavior change
- `[test]` tests only

## Commit Message Format

```text
[feat] add request resolver for named specs
[bug] fix jsonpath assertion on numeric values
[docs] clarify env precedence and secret storage
```

## PR Guidelines

- PR should contain one atomic intent.
- Keep PR description explicit:
  - what changed
  - why
  - any migration impact
- Include sample command output or screenshots for UX changes.
