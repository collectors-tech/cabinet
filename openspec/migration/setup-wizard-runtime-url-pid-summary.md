# Setup Wizard Startup Runtime URL + PID Contract Summary (#246)

## Scope
- Issue: #246
- Spec IDs: SETUP-WIZ-007, SETUP-WIZ-008, SETUP-WIZ-009
- Spec path: `openspec/specs/general/setup-wizard-first-run/spec.md`

## Changes
- Added `meta.currentUrl` to setup config schema and completion payload write path.
- Added startup config sync to reconcile `meta.currentUrl` with resolved runtime URL when `cabinet.json` already exists.
- Added PID lifecycle helpers and runtime integration:
  - PID file path: `<dataDir>/cabinet.pid`
  - content contract: numeric PID only
  - cleanup on runtime exit.
- Added/extended runtime setup tests for URL metadata sync and PID-only file semantics.
- Extended OpenSpec with `SETUP-WIZ-008` and `SETUP-WIZ-009` and mapped traceability to executable tests.

## Commands Run
- `go test ./internal/app -run "RuntimeSetup|RuntimePID" -count=1`
- `pwsh -NoLogo -NoProfile -File .\\cypress.ps1 -Spec cypress/e2e/general/setup-wizard-first-run/spec.cy.ts -Browser chrome -RequireE2EHooks`
- `go test ./internal/app -count=1`
- `go test ./tests -count=1`
- `openspec validate --all`

## Results
- runtime setup/PID targeted tests: **pass**
- setup-wizard Cypress suite: **7 passing / 0 failing**
- `go test ./internal/app`: **pass**
- `go test ./tests`: **pass**
- `openspec validate --all`: **pass**

## Outcome
- First-run setup remains deterministic and now preserves startup runtime URL metadata.
- Runtime lock lifecycle now uses an explicit PID-only file separate from `cabinet.json`.
