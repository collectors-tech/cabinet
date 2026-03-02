# Auto Wave 43 Summary

- Issue: #222
- Status: done
- Requirement IDs:
  - UI-FOUNDATION-SHELL-NAVIGATION-007
- Spec paths:
  - openspec/specs/general/ui-foundation-shell-navigation/spec.md

## Changes
- `ui.web/src/components/layout/app-sidebar.tsx`
  - Added explicit nav edit draft order state (`navEditOrder`) for live reorder rendering.
  - Reorder actions now mutate the draft order directly.
  - Closing nav edit mode persists the exact shown order into stored nav preferences.
- `ui.web/cypress/e2e/general/ui-foundation-shell-navigation/spec.cy.ts`
  - Scoped reorder actions/assertions to visible nav edit panel.
  - Maintained requirement-named test coverage for UI-FOUNDATION-SHELL-NAVIGATION-007.
- `openspec/traceability.md`
  - Marked UI-FOUNDATION-SHELL-NAVIGATION-007 as implemented with executable Cypress evidence.
- `internal/ui/static/*`
  - Rebuilt embedded UI assets.

## Commands run
1. `pwsh -NoLogo -NoProfile -File .\\cypress.ps1 -Spec cypress/e2e/general/ui-foundation-shell-navigation/spec.cy.ts -Browser chrome -RequireE2EHooks` (fail, reproducible baseline)
2. `npm run build` (ui.web)
3. `pwsh -NoLogo -NoProfile -File .\\cypress.ps1 -Spec cypress/e2e/general/ui-foundation-shell-navigation/spec.cy.ts -Browser chrome -RequireE2EHooks` (fail against reused server)
4. forced recycle on 17880 listener
5. `pwsh -NoLogo -NoProfile -File .\\cypress.ps1 -Spec cypress/e2e/general/ui-foundation-shell-navigation/spec.cy.ts -Browser chrome -RequireE2EHooks` (pass)
6. `go test ./internal/app -count=1` (pass)
7. `go test ./tests -count=1` (pass)
8. `openspec validate --all` (pass)

## Key results
- Managed Cypress target suite: 5 passing, 0 failing.
- Go suites passed in touched mandatory gates.
- OpenSpec validation passed.
- Requirement UI-FOUNDATION-SHELL-NAVIGATION-007 moved to implemented with proof.