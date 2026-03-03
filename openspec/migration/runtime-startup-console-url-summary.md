# Runtime Startup Console URL Summary (#247)

## Scope
- Issue: #247
- Spec ID: RUNTIME-CORE-004
- Spec path: `openspec/specs/general/runtime-core/spec.md`

## Changes
- Added startup listener bind flow in `App.Run` to resolve actual endpoint before serving.
- Added machine-parseable startup console line with runtime context:
  - `url=<resolved-url>`
  - `requested_addr`, `resolved_addr`
  - `instance`, `profile`, `data_dir`
  - `requested_port`, `resolved_port`
- Added startup notice test hook on `App` and integration test proving output contract.
- Added runtime URL metadata sync using resolved listener URL during startup.
- Updated OpenSpec + traceability with explicit startup console requirement and executable proof mapping.

## Commands Run
- `go test ./internal/app -run "RuntimeStartupConsole" -count=1`
- `pwsh -NoLogo -NoProfile -File .\\cypress.ps1 -Spec cypress/e2e/general/setup-wizard-first-run/spec.cy.ts -Browser chrome -RequireE2EHooks`
- `go test ./internal/app -count=1`
- `go test ./tests -count=1`
- `openspec validate --all`

## Results
- startup console test: **pass**
- setup-wizard Cypress suite: **7 passing / 0 failing**
- `go test ./internal/app`: **pass**
- `go test ./tests`: **pass**
- `openspec validate --all`: **pass**

## Outcome
- Runtime startup now consistently emits resolved URL and execution context in a parseable console line after bind success.
