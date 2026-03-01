# Auto Wave 36 Summary

- Issue: `#172`
- Scope: close auth menu + shortcut partial IDs with executable Cypress proof

## Requirement IDs moved to implemented

- `UI-FOUNDATION-AUTH-MENUS-SHORTCUTS-001`
- `UI-FOUNDATION-AUTH-MENUS-SHORTCUTS-002`
- `UI-FOUNDATION-AUTH-MENUS-SHORTCUTS-003`

## Commands Run

1. `pwsh -NoLogo -NoProfile -File .\cypress.ps1 -Spec cypress/e2e/general/ui-foundation-auth-menus-shortcuts/spec.cy.ts -Browser chrome`
2. `go test ./internal/app -count=1`
3. `go test ./tests -count=1`
4. `openspec validate --all`

## Key Results

- Added spec-aligned E2E suite:
  - `ui.web/cypress/e2e/general/ui-foundation-auth-menus-shortcuts/spec.cy.ts`
- Proved:
  - platform-aware shortcut notation in active account menu
  - unsupported template actions (`New Team`, `Billing`) are absent
  - sidebar upsell row (`Upgrade to Pro`) is absent
- Validation gates passed:
  - Cypress target spec: `3 passing`
  - `go test ./internal/app -count=1`: pass
  - `go test ./tests -count=1`: pass
  - `openspec validate --all`: pass

## Blockers

- None.
