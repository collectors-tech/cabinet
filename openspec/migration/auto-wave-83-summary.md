# Auto Wave 83 Summary

- Issue: #252
- Scope: Setup Wizard auth step Local/Clerk readiness + deterministic transition validation
- Requirement IDs: SETUP-WIZ-015

## What Changed

1. Auth step UI readiness contract
- Added deterministic auth readiness indicator in setup auth step:
  - `Ready: Local auth`
  - `Missing Clerk key`
  - `Configured`
- Added deterministic mode-switch behavior:
  - switching to `local` clears inactive Clerk key value from form state.
- Added auth-step transition validation:
  - when `auth_mode=clerk` and key is blank, `Next` is blocked with actionable inline error.

2. Runtime setup auth persistence proof
- Added backend API contract test:
  - `TestRuntimeSetupCompletePersistsClerkAuthConfiguration`
- Test verifies persisted config contract:
  - `auth.mode=clerk`
  - `auth.clerk.enabled=true`
  - `auth.clerk.publishableKey` matches submitted key

3. E2E + OpenSpec + traceability
- Added and greened auth-step E2E coverage:
  - `UC-SW-23` auth mode switch contract
  - `UC-SW-24` clerk readiness missing blocks next
  - `UC-SW-25` clerk readiness configured persists auth payload
- Updated setup wizard spec mapping for `UC-SW-23..25` to implemented.
- Updated traceability with `SETUP-WIZ-015` as implemented with Cypress + Go evidence.

## Commands Run

- `pwsh -NoLogo -NoProfile -File .\\cypress.ps1 -RequireE2EHooks -Spec cypress/e2e/general/setup-wizard-first-run/spec.cy.ts -Browser chrome` (fail-first baseline)
- `npm run build` (in `ui.web`)
- `pwsh -NoLogo -NoProfile -File .\\cypress.ps1 -RequireE2EHooks -Spec cypress/e2e/general/setup-wizard-first-run/spec.cy.ts -Browser chrome`
- `go test ./internal/app -count=1`
- `go test ./tests -count=1`
- `openspec validate --all`

## Results

- Target Cypress setup-wizard spec: 21 passing, 0 failing.
- `go test ./internal/app -count=1`: pass.
- `go test ./tests -count=1`: pass.
- `openspec validate --all`: pass.
