# Auto Wave 28 Summary

- Issue: #172
- Scope: close onboarding resume contract (`UI-SCREEN-ONBOARDING-AUTH-002`).
- Date: 2026-03-02

## Requirement IDs moved to implemented
- `UI-SCREEN-ONBOARDING-AUTH-002`

## Changes delivered
- Added profile-scoped onboarding progress recovery with deterministic fallback migration from default scope.
- Added/kept explicit onboarding test IDs for step label and next-step control in dashboard onboarding widget.
- Extended onboarding/auth Cypress suite with persisted step resume assertion after reload.
- Updated traceability mapping from partial to implemented with executable Cypress proof.
- Rebuilt embedded frontend bundle for runtime parity.

## Commands run
1. `./cypress.ps1 -Spec "cypress/e2e/general/ui-screen-onboarding-auth/spec.cy.ts" -Browser chrome`
2. `go test ./internal/app -count=1`
3. `go test ./tests -count=1`
4. `openspec validate --all`

## Results
- Cypress (`ui.web/cypress/e2e/general/ui-screen-onboarding-auth/spec.cy.ts`): **3 passing, 0 failing**
- `go test ./internal/app -count=1`: **pass**
- `go test ./tests -count=1`: **pass**
- `openspec validate --all`: **pass**
