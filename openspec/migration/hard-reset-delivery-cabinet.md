# Hard Reset Delivery - Cabinet

Date: 2026-03-01

## Scope
1. Login/onboarding flow
- `ui.web/cypress/e2e/profile-onboarding.cy.ts`

2. Inventory flow
- `ui.web/cypress/e2e/inventory-management.cy.ts`

## Exact Commands Run
1. `cd D:\projects\collectors-tech\cabinet\ui.web`
2. `Remove-Item Env:ELECTRON_RUN_AS_NODE -ErrorAction SilentlyContinue`
3. `npx start-server-and-test "go run ../cmd/cabinet" http://127.0.0.1:17880 "npx cypress run --browser chrome --config-file D:/projects/collectors-tech/cabinet/ui.web/cypress.config.ts --spec D:/projects/collectors-tech/cabinet/ui.web/cypress/e2e/profile-onboarding.cy.ts --reporter junit --reporter-options mochaFile=D:/projects/collectors-tech/cabinet/ui.web/cypress/artifacts/results/profile-onboarding.xml,toConsole=true"`
4. `npx start-server-and-test "go run ../cmd/cabinet" http://127.0.0.1:17880 "npx cypress run --browser chrome --config-file D:/projects/collectors-tech/cabinet/ui.web/cypress.config.ts --spec D:/projects/collectors-tech/cabinet/ui.web/cypress/e2e/inventory-management.cy.ts --reporter junit --reporter-options mochaFile=D:/projects/collectors-tech/cabinet/ui.web/cypress/artifacts/results/inventory-management.xml,toConsole=true"`
5. `cd D:\projects\collectors-tech\cabinet`
6. `go test ./internal/app -count=1`
7. `go test ./tests -count=1`
8. `openspec validate --all`

## Cypress Pass/Fail Counts
- `profile-onboarding.cy.ts`: `tests=1`, `failures=0` (from `ui.web/cypress/artifacts/results/profile-onboarding.xml`)
- `inventory-management.cy.ts`: `tests=1`, `failures=0` (from `ui.web/cypress/artifacts/results/inventory-management.xml`)

## IDs Moved To Implemented
- `UI-SCREEN-INVENTORY-ITEMS-001`

## IDs Updated With New E2E Evidence (still partial)
- `UI-SCREEN-INVENTORY-ITEMS-002`
- `UI-SCREEN-INVENTORY-ITEMS-003`
- `UI-SCREEN-ONBOARDING-AUTH-001` (evidence augmented)

## Blockers Encountered and Resolved
1. Command/runtime blocker:
- first actionable error: `Error: spawn wmic.exe ENOENT`
- required fix: prepend WMIC shim path for managed runs: `D:\projects\collectors-tech\cabinet\.agentbus\bin`

2. Command/runtime blocker:
- first actionable error: `runtime failed: listen tcp 127.0.0.1:17880: bind: Only one usage...`
- required fix: pre-clear process bound to `:17880` before managed run

3. Test assertion blocker:
- first actionable error: inventory route opened `404` when test used `/_authenticated/inventory/`
- required code/test fix: use routed path `/inventory/` in `inventory-management.cy.ts`

## Final Validation Results
- `go test ./internal/app -count=1` -> pass
- `go test ./tests -count=1` -> pass
- `openspec validate --all` -> pass (`5 passed, 0 failed`)
