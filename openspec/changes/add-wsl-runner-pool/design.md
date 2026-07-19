## Context

Cabinet CI currently runs on GitHub-hosted workers. The repository needs an
optional, repeatable path for dedicated local capacity without returning to a
shared, ever-growing WSL/Docker environment.

## Goals / Non-Goals

**Goals:**

- Provision isolated WSL distributions with a configurable 60 GB default VHD
  ceiling and aggregate Windows free-space guard.
- Make pool scale-out idempotent and give every member a unique identity,
  registration token, and runner label.
- Bound Docker logs and reap unused Docker, workspace, journal, package-cache,
  and filesystem space only while the runner is idle.
- Keep runners online after Windows logon through hidden per-distro tasks.

**Non-Goals:**

- Retarget existing workflows to self-hosted labels.
- Create, recreate, unregister, or prune a live distribution during tests.
- Share Docker daemons or workspaces between runner pool members.

## Decisions

1. Use one WSL distribution per runner.
   - This creates a clear storage and failure boundary and makes a damaged or
     oversized member disposable without affecting other repositories.

2. Treat `RunnerCount` as desired state.
   - The unnumbered distro is member 1. Members 2 onward use `-02`, `-03`, and
     are skipped when an Actions runner systemd unit is already configured.

3. Acquire registration tokens through authenticated GitHub CLI by default.
   - GitHub registration tokens are single-use. The parent requests a fresh
     token for every missing pool member rather than duplicating a supplied
     token.

4. Use `overlay2`, bounded `json-file` logs, and idle-aware cleanup.
   - Cleanup locks against job hooks and checks `Runner.Worker` before deleting
     workspaces or unused Docker state.

## Risks / Trade-offs

- WSL `--vhd-size` is a virtual ceiling, not preallocated disk usage. The
  installer therefore checks aggregate worst-case headroom for missing members.
- Aggressive Docker volume pruning can remove unused caches. It only runs after
  a job or while the runner is demonstrably idle.
- A current-user logon task starts after interactive sign-in, not before it.

## Migration Plan

1. Merge the script, test, and runbook without changing workflow selectors.
2. Run `-WhatIf` to review names, locations, and capacity.
3. Provision the requested pool with GitHub CLI authentication.
4. Retarget selected jobs only after runners are visible and idle in GitHub.

Rollback:

- Retarget jobs to GitHub-hosted workers.
- Remove the runner in GitHub, unregister its dedicated distro, and remove its
  matching Windows autostart task.
