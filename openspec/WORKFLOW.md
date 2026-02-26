# OpenSpec Workflow

## Purpose
Define how this repository executes spec-first delivery using OpenSpec and GitHub issues.

## Source of Truth
- OpenSpec specs under `openspec/specs/` are normative requirements.
- GitHub issues are normative execution tracking.

## Required Flow
1. Create/reuse a GitHub issue.
2. Create/update OpenSpec change artifacts.
3. Validate OpenSpec:
   - `openspec validate --changes --strict --no-interactive`
4. Implement with test-first workflow.
5. Re-run validation/tests.
6. Commit/push with issue-prefixed commit message.
7. Close issue only when acceptance criteria and test evidence are complete.

## Commands
```powershell
openspec list
openspec validate --all --strict --no-interactive
openspec new change <change-name>
openspec archive <change-name>
```

## Local Hook Setup
```powershell
./scripts/install-githooks.ps1
```

Installed hooks enforce:
- commit issue prefix
- OpenSpec validation on commit/push
- fast API smoke validation on push

