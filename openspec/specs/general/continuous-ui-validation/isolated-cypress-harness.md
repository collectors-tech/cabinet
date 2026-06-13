# Isolated Cypress Harness

Cabinet's isolated Cypress harness runs browser specs against fresh runtime lanes instead of a shared desktop server. Use it when validating issue branches, PR merge gates, or repeatable QA slices that need known runtime, profile, data, and commit evidence.

## Local Prerequisites

- PowerShell 7 available as `pwsh`.
- Go, Node.js, npm, and Cypress dependencies installed for the repo checkout.
- Docker Desktop running when using container-backed lanes.
- A repo-local Cabinet image built from the exact checkout under test:

```powershell
docker build -t cabinet:e2e .
```

- Free local ports for the selected lane range. The runner maps lane one to `-BasePort`, lane two to `-BasePort + 1`, and so on.
- A log path under `.work-agent\logs\...` for durable evidence.

## Single-Spec Container Lane

Run one browser spec against one fresh container lane:

```powershell
pwsh -NoLogo -NoProfile -File .\scripts\run-cypress-matrix.ps1 `
  -SpecGlob 'ui.web/cypress/e2e/general/ui-login-session/spec.cy.ts' `
  -LaneCount 1 `
  -MaxWorkers 1 `
  -BasePort 18012 `
  -UseContainerImage `
  -ContainerImage 'cabinet:e2e' `
  -ApiContractSmoke `
  -RequireE2EHooks `
  -RunId '<run-id>'
```

The runner starts a named container, mounts a lane-specific data volume, waits for `/healthz`, runs the API contract smoke preflight, invokes `cypress.ps1` with `-ReuseServer`, and removes the lane container and volume unless `-KeepContainers` is set.

## Bounded Matrix Command

Run a bounded group of specs across isolated lanes:

```powershell
pwsh -NoLogo -NoProfile -File .\scripts\run-cypress-matrix.ps1 `
  -SpecGlob 'ui.web/cypress/e2e/general/ui-*/spec.cy.ts' `
  -LaneCount 2 `
  -MaxWorkers 2 `
  -BasePort 18020 `
  -UseContainerImage `
  -ContainerImage 'cabinet:e2e' `
  -ApiContractSmoke `
  -RequireE2EHooks `
  -RunId '<run-id>'
```

Use `-PlanOnly` before live execution to inspect lane assignment without starting containers or Cypress.

## Evidence And Summaries

Each run writes its machine-readable matrix summary to:

```text
.work-agent\logs\cypress-matrix\<run-id>\matrix.summary.json
```

The summary records the source commit, spec count, worker limit, lane ports, data directories, profiles, instance names, container names, volumes, result paths, aggregate passed/failed lane and spec counts, and run/lane/per-spec timing metadata. Per-spec Cypress logs and summaries are linked from each result entry.

## Failure Stages

Use the `failure_stage` value in `matrix.summary.json` to triage failures:

- `container_image`: the configured image was unavailable before lane fanout. Build the image from the repository root and rerun.
- `port_preflight`: an active lane host port was already accepting TCP connections before lane fanout. Stop the stale listener or choose a different `-BasePort`.
- `container_start`: Docker accepted the image but the lane container could not start.
- `runtime_health`: the lane container started but did not pass `/healthz` before the timeout.
- `cypress`: the runtime was ready and Cypress or the API contract smoke preflight failed.

## Fallback When Docker Is Unavailable

Do not reuse stale shared desktop runtimes as isolated-lane proof.

When Docker is unavailable, use `cypress.ps1` directly with a unique `-BaseUrl`, `-RuntimeExecutablePath`, data directory, profile, and instance name for the focused spec. Record the fallback explicitly in the issue or PR comment, include the command log and Cypress summary path, and classify it as non-container fallback evidence rather than isolated container-lane proof.
