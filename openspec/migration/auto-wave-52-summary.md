# Auto Wave 52 Summary

## Issue
- #224 `[Spec Backlog] Remove Collections widget above primary nav (DB switcher only at top)`

## Requirement IDs
- UI-FOUNDATION-SHELL-NAVIGATION-005
- UI-FOUNDATION-SHELL-NAVIGATION-006

## Commands Run
1. `pwsh -NoLogo -NoProfile -File .\\cypress.ps1 -Spec cypress/e2e/general/ui-foundation-shell-navigation/spec.cy.ts -Browser chrome -RequireE2EHooks` (fail-first: top widget exists; collections section missing)
2. `npx @tanstack/router-cli@latest generate` (route map update for new collections route)
3. `npm run build` (ui.web)
4. `pwsh -NoLogo -NoProfile -File .\\cypress.ps1 -Spec cypress/e2e/general/ui-foundation-shell-navigation/spec.cy.ts -Browser chrome -RequireE2EHooks` (green)
5. `go test ./internal/app -count=1` (pass)
6. `go test ./tests -count=1` (pass)
7. `openspec validate --all` (pass)

## Key Results
- Removed sidebar-top Collections management widget above primary nav; sidebar top area now DB/profile switcher only.
- Added dedicated Collections section route and UI (`/collections`) for list/create management.
- Added inline collection picker quick-create in inventory collection browser with auto-select behavior.
- Updated shell-navigation E2E assertions and traceability for IDs 005/006 to match new contract language.

## Gate Results
- Managed Cypress touched scope: pass (`ui.web/cypress/e2e/general/ui-foundation-shell-navigation/spec.cy.ts`)
- `go test ./internal/app -count=1`: pass
- `go test ./tests -count=1`: pass
- `openspec validate --all`: pass

## Status
- Ready for commit/push-proof/close.
