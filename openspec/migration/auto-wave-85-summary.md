# Auto Wave 85 Summary

- Issue: #254
- Scope: Setup Wizard review/create action contract hardening
- Requirement IDs: SETUP-WIZ-017

## What Changed

1. Review action contract
- Updated review action label in `ui.web/src/features/auth/sign-in/index.tsx`:
  - idle: `Create Config & Launch`
  - in-flight: `Creating Config...`
- Kept in-flight disabled semantics on create action during setup completion request.

2. Review/create E2E coverage
- Added and greened setup wizard review-step E2E:
  - `UC-SW-29` review summary completeness + create label
  - `UC-SW-30` deterministic create action and completion metadata visibility
  - `UC-SW-31` in-flight create action disabled/progress state

3. OpenSpec + traceability
- Added requirement `SETUP-WIZ-017` in `openspec/specs/general/setup-wizard-first-run/spec.md`.
- Updated UC matrix statuses:
  - `UC-SW-29..31` -> implemented.
- Updated `openspec/traceability.md`:
  - `SETUP-WIZ-017` -> implemented with Cypress + Go evidence.

## Commands Run

- `pwsh -NoLogo -NoProfile -File .\\cypress.ps1 -RequireE2EHooks -Spec cypress/e2e/general/setup-wizard-first-run/spec.cy.ts -Browser chrome` (fail-first baseline)
- `npm run build` (in `ui.web`)
- `pwsh -NoLogo -NoProfile -File .\\cypress.ps1 -RequireE2EHooks -Spec cypress/e2e/general/setup-wizard-first-run/spec.cy.ts -Browser chrome`
- `go test ./internal/app -count=1`
- `go test ./tests -count=1`
- `openspec validate --all`

## Results

- Target Cypress setup-wizard spec: 27 passing, 0 failing.
- `go test ./internal/app -count=1`: pass.
- `go test ./tests -count=1`: pass.
- `openspec validate --all`: pass.
