# Auto Wave 27 Summary

- Issue: #172
- Scope: close collection-context propagation contract (`UI-FOUNDATION-SHELL-NAVIGATION-003`).
- Date: 2026-03-02

## Requirement IDs moved to implemented
- `UI-FOUNDATION-SHELL-NAVIGATION-003`

## Changes delivered
- Added active collection context state to collection workspace:
  - context label in body header
  - selectable folder context buttons
  - active context summary field
- Extended shell-navigation E2E with context selection assertions.
- Updated traceability mapping to executable Cypress evidence.
- Rebuilt embedded frontend bundle for runtime parity.

## Commands run
1. `./cypress.ps1 -Spec "cypress/e2e/general/ui-foundation-shell-navigation/spec.cy.ts" -Browser chrome`
2. `go test ./internal/app -count=1`
3. `go test ./tests -count=1`
4. `openspec validate --all`

## Results
- Cypress (`ui.web/cypress/e2e/general/ui-foundation-shell-navigation/spec.cy.ts`): **4 passing, 0 failing**
- `go test ./internal/app -count=1`: **pass**
- `go test ./tests -count=1`: **pass**
- `openspec validate --all`: **pass**
