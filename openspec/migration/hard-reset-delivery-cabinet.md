# Hard Reset Delivery - Cabinet

Date: 2026-03-01

## Scope
1. Login/onboarding flow
- `ui.web/cypress/e2e/general/ui-screen-onboarding-auth/spec.cy.ts`

2. Inventory flow
- `ui.web/cypress/e2e/inventory/ui-screen-inventory-items/spec.cy.ts`

## Exact Commands Run
1. `cd D:\projects\collectors-tech\cabinet\ui.web`
2. `pwsh ../cypress.ps1 -Browser chrome -Spec cypress/e2e/general/ui-screen-onboarding-auth/spec.cy.ts`
3. `pwsh ../cypress.ps1 -Browser chrome -Spec cypress/e2e/inventory/ui-screen-inventory-items/spec.cy.ts`
4. `cd D:\projects\collectors-tech\cabinet`
5. `go test ./internal/app -count=1`
6. `go test ./tests -count=1`
7. `openspec validate --all`

## Cypress Pass/Fail Counts
- `ui-screen-onboarding-auth/spec.cy.ts`: `tests=2`, `failures=0`
- `ui-screen-inventory-items/spec.cy.ts`: `tests=1`, `failures=0`

## IDs Moved To Implemented
- `UI-SCREEN-INVENTORY-ITEMS-001`

## IDs Updated With New E2E Evidence (still partial)
- `UI-SCREEN-INVENTORY-ITEMS-002`
- `UI-SCREEN-INVENTORY-ITEMS-003`
- `UI-SCREEN-ONBOARDING-AUTH-001` (evidence augmented)

## Blockers Encountered and Resolved
1. Command/runtime blocker:
- first actionable error: `managed wrapper dependency failure`
- required fix: remove managed wrapper usage and execute Cypress directly via `cypress.ps1`

2. Command/runtime blocker:
- first actionable error: `runtime failed: listen tcp 127.0.0.1:17880: bind: Only one usage...`
- required fix: pre-clear process bound to `:17880` before managed run

3. Test assertion blocker:
- first actionable error: inventory route opened `404` when test used `/_authenticated/inventory/`
- required code/test fix: use routed path `/inventory/` in `ui.web/cypress/e2e/inventory/ui-screen-inventory-items/spec.cy.ts`

## Final Validation Results
- `go test ./internal/app -count=1` -> pass
- `go test ./tests -count=1` -> pass
- `openspec validate --all` -> pass (`5 passed, 0 failed`)

