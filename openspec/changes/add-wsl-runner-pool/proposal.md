## Why

Cabinet does not have a repository-owned way to provision dedicated WSL 2
self-hosted GitHub Actions runners. Sharing an unbounded WSL or Docker host
mixes runner identity, Docker storage, workspaces, and logs across workloads,
which makes capacity failures and VHDX growth difficult to diagnose or contain.

## What Changes

- Add a PowerShell installer for one or more isolated Cabinet WSL runners.
- Give each runner its own bounded VHDX, Linux account, Docker daemon, GitHub
  registration, cleanup timer, and Windows logon autostart task.
- Install the Cabinet Linux CI toolchain: Node.js 22, Go from `go.mod`, Docker,
  PowerShell, and GitHub Actions runner dependencies.
- Add a contract test and an operator runbook covering safe provisioning,
  scale-out, inspection, cleanup, and VHD reclamation.

## Capabilities

### New Capabilities

- `wsl-runner-pool`: bounded, repository-scoped WSL runner provisioning and
  lifecycle management.

### Modified Capabilities

- None.

## Impact

- Affected code: `scripts/github-actions-runner/install-wsl-runner.ps1`.
- Affected tests: `scripts/github-actions-wsl-runner.issue-1949.test.mjs`.
- Affected docs: `docs/ci/self-hosted-runner.md`.
- Related issue: `#1949`.
