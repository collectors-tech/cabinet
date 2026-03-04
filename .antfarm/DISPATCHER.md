# Antfarm Dispatchers (Cabinet)

This repo now runs with two lanes:

1) **Build lane** (issue-driven)
- Script: `.antfarm/scripts/issue-dispatcher.ps1`
- Starts workflow: `cabinet`
- Picks next eligible issue (excluding `blocked`) with labels: `ready` OR `priority:p*` OR `high-priority`, then oldest update within highest priority.
- Goal: keep coding/build work moving issue-by-issue.

2) **Validator lane** (artifact-driven)
- Script: `.antfarm/scripts/validator-dispatcher.ps1`
- Starts workflow: `cabinet-validator`
- Picks newest manifest from `.antfarm/artifacts/release-*.json`
- Validates only latest artifact; older artifacts may be skipped intentionally.

## Artifact publishing
- Script: `.antfarm/scripts/publish-artifact.ps1`
- Writes manifest JSON into `.antfarm/artifacts/` with `release-<name>-<gitsha>-<timestamp>.json`.

## Why this split
Build can continue while validator is running. Validator always catches up to the newest artifact instead of blocking developer throughput.
