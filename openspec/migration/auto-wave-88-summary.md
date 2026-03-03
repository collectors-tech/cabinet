# Auto Wave 88 Summary

- Issue: #270
- Scope: rename Scanner user-facing workflow naming to Market Watch while retaining `/scanner` route compatibility
- Requirement IDs: `UI-SCREEN-SCANNER-005`

## What Changed

1. OpenSpec terminology + executable requirement
- Updated `openspec/specs/integrations/ui-screen-scanner/spec.md` to add `UI-SCREEN-SCANNER-005`:
  - user-facing naming uses `Market Watch`
  - `/scanner` route remains backward-compatible
- Updated scanner screen spec wording from Scanner-centric copy to Market Watch user terminology while keeping stable requirement IDs.

2. Fail-first + implementation
- Added fail-first Cypress test:
  - `UI-SCREEN-SCANNER-005 uses Market Watch naming with scanner route compatibility`
  - path: `ui.web/cypress/e2e/integrations/ui-screen-scanner/spec.cy.ts`
- Implemented naming updates across user-facing provider query workflow surfaces:
  - sidebar nav label: `Scanner` -> `Market Watch`
  - scanner screen heading/subcopy/loading/empty/error text
  - integrations sync guidance copy and settings operations copy
  - setup toggle label: `Enable Market Watch`

3. Related nav test alignment
- Updated nav-id expectations from `scanner` to `market-watch` in:
  - `ui.web/cypress/e2e/general/ui-foundation-shell-navigation/spec.cy.ts`

4. Traceability
- Updated `openspec/traceability.md`:
  - `UI-SCREEN-SCANNER-005` -> implemented with Cypress evidence.

## Commands Run

- `pwsh -NoLogo -NoProfile -File ./cypress.ps1 -Spec cypress/e2e/integrations/ui-screen-scanner/spec.cy.ts -Browser chrome -RequireE2EHooks` (fail-first)
- `./scripts/build-ui-static.ps1` (refresh embedded UI bundle)
- `pwsh -NoLogo -NoProfile -File ./cypress.ps1 -Spec cypress/e2e/integrations/ui-screen-scanner/spec.cy.ts -Browser chrome -RequireE2EHooks`
- `pwsh -NoLogo -NoProfile -File ./cypress.ps1 -Spec cypress/e2e/general/ui-foundation-shell-navigation/spec.cy.ts -Browser chrome -RequireE2EHooks`
- `go test ./internal/app -count=1`
- `go test ./tests -count=1`
- `openspec validate --all`
- `go build -o bin/cabinet.exe ./cmd/cabinet`

## Results

- integrations scanner Cypress suite: pass (`5 passing, 0 failing`).
- shell navigation Cypress suite: pass (`10 passing, 0 failing`).
- `go test ./internal/app -count=1`: pass.
- `go test ./tests -count=1`: pass.
- `openspec validate --all`: pass.
- `bin/cabinet.exe` rebuilt successfully.
