# Auto Wave 78 Summary

## Issue
- #247 `[Spec Backlog] Startup console output must show resolved runtime URL`

## Implemented IDs
- RUNTIME-CORE-004

## What Changed
- Runtime now binds listener before serve to obtain actual resolved endpoint.
- Startup prints parseable context line including resolved URL + requested/resolved ports + instance/profile/data_dir.
- Added startup console output proof test using `App.Run` with `127.0.0.1:0` bind.
- Updated runtime-core spec and traceability for `RUNTIME-CORE-004`.

## Validation Evidence
- `go test ./internal/app -run "RuntimeStartupConsole" -count=1` => pass
- `pwsh -NoLogo -NoProfile -File .\\cypress.ps1 -Spec cypress/e2e/general/setup-wizard-first-run/spec.cy.ts -Browser chrome -RequireE2EHooks` => pass (7/7)
- `go test ./internal/app -count=1` => pass
- `go test ./tests -count=1` => pass
- `openspec validate --all` => pass

## Blockers
- None
