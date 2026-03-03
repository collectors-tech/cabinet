# Auto Wave 82 Summary

- Issue: #251
- Scope: Setup Wizard runtime/network step (auto/fixed port mode, resolved URL preview, fixed-port validation)
- Requirement IDs: SETUP-WIZ-014

## What Changed

1. Runtime setup API defaults + persistence
- Extended `GET /api/runtime/setup-status` payload with deterministic runtime defaults:
  - `default_runtime_host`
  - `default_runtime_port`
  - `default_runtime_port_mode`
  - `default_runtime_url`
- Existing setup completion contract now receives runtime step values from UI:
  - `runtime_port_mode`
  - `runtime_fixed_port`
- Added backend persistence test for fixed runtime mode/port:
  - `TestRuntimeSetupCompletePersistsFixedPortRuntime`

2. Setup wizard runtime step UI
- Expanded setup wizard from 4 to 5 steps:
  - Identity -> Storage -> Runtime -> Auth -> Review
- Added runtime step controls:
  - runtime port mode select (`auto` / `fixed`)
  - fixed port numeric input (fixed mode)
  - resolved URL preview
  - fallback message for auto-mode multi-instance port usage
- Added inline runtime validation:
  - blocks next when fixed mode selected and fixed port is not valid (> 0)
- Setup completion request now persists selected runtime mode/port.

3. E2E + OpenSpec + traceability
- Added and greened runtime E2E coverage:
  - `UC-SW-20` runtime defaults
  - `UC-SW-21` fixed-port validation
  - `UC-SW-22` runtime persistence
- Updated setup wizard spec requirement mapping:
  - `UC-SW-20..22` -> implemented
- Updated traceability:
  - `SETUP-WIZ-014` -> implemented with Cypress + Go proof

## Commands Run

- `pwsh -NoLogo -NoProfile -File .\\cypress.ps1 -Spec cypress/e2e/general/setup-wizard-first-run/spec.cy.ts -Browser chrome` (fail-first baseline)
- `go test ./internal/app -run "TestRuntimeSetup(StatusAndCompleteContract|StorageValidateContract|CompletePersistsSelectedStoragePath|CompletePersistsFixedPortRuntime)" -count=1`
- `npm run build` (in `ui.web`)
- `pwsh -NoLogo -NoProfile -File .\\cypress.ps1 -RequireE2EHooks -Spec cypress/e2e/general/setup-wizard-first-run/spec.cy.ts -Browser chrome`
- `go test ./internal/app -count=1`
- `go test ./tests -count=1`
- `openspec validate --all`

## Results

- Target Cypress setup-wizard spec: 18 passing, 0 failing.
- `go test ./internal/app -count=1`: pass.
- `go test ./tests -count=1`: pass.
- `openspec validate --all`: pass.
