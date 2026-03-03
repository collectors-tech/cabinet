# Auto Wave 77 Summary

## Issue
- #246 `[Spec Backlog] First-run startup wizard when cabinet.json is missing + runtime URL persistence`

## Implemented IDs
- SETUP-WIZ-007
- SETUP-WIZ-008
- SETUP-WIZ-009

## What Changed
- Setup config metadata now persists `meta.currentUrl`.
- Startup reconciles runtime URL into existing `cabinet.json` metadata.
- Runtime now writes PID-only lifecycle file (`cabinet.pid`) and removes it on exit.
- Added runtime setup tests for URL metadata sync and PID-only file contract.
- Updated OpenSpec + traceability with explicit startup metadata/PID requirements and proof.

## Validation Evidence
- `go test ./internal/app -run "RuntimeSetup|RuntimePID" -count=1` => pass
- `pwsh -NoLogo -NoProfile -File .\\cypress.ps1 -Spec cypress/e2e/general/setup-wizard-first-run/spec.cy.ts -Browser chrome -RequireE2EHooks` => pass (7/7)
- `go test ./internal/app -count=1` => pass
- `go test ./tests -count=1` => pass
- `openspec validate --all` => pass

## Blockers
- None
