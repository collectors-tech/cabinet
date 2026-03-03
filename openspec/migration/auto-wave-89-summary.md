# Auto Wave 89 Summary

- Issue: #271
- Scope: Market Watch provider selector + provider-scoped query create/run contracts.
- Requirement IDs: `UI-SCREEN-MARKET-WATCH-001`, `UI-SCREEN-MARKET-WATCH-002`, `UI-SCREEN-MARKET-WATCH-003`

## What changed
- Added Market Watch provider selector controls (single/multi) with deterministic validation when provider scope is empty.
- Added provider scope into create query-set payload.
- Added provider scope into run payload and rendered provider attribution in query rows and run candidates.
- Added Cypress E2E spec mapped to OpenSpec hierarchy:
  - `ui.web/cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts`
- Updated traceability mappings for the three IDs to `implemented` with executable Cypress proof.

## Commands run
1. `pwsh -NoLogo -NoProfile -File .\\cypress.ps1 -Spec cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts -Browser chrome -RequireE2EHooks`
2. `go test ./internal/app -count=1`
3. `go test ./tests -count=1`
4. `openspec validate --all`
5. `go build -o bin/cabinet.exe ./cmd/cabinet`

## Results
- Cypress targeted spec: 3 passing, 0 failing.
- `go test ./internal/app`: pass.
- `go test ./tests`: pass.
- `openspec validate --all`: pass.
- `bin/cabinet.exe` rebuilt successfully.

## Blockers
- None.
