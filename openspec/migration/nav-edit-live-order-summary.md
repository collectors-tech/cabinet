# Nav Edit Live Order Summary

- Issue: #222
- Requirement ID: UI-FOUNDATION-SHELL-NAVIGATION-007
- Spec: openspec/specs/general/ui-foundation-shell-navigation/spec.md

## What changed
- Updated sidebar nav edit behavior to use an explicit edit-order draft list so move up/down reorders immediately in the edit dialog.
- Persisted the exact draft order to nav preferences when exiting edit mode, ensuring saved nav order matches what was shown pre-save.
- Tightened Cypress shell-navigation test selectors to scope reorder interactions/assertions to the visible edit panel instance.
- Rebuilt embedded UI assets into internal/ui/static.
- Updated traceability for UI-FOUNDATION-SHELL-NAVIGATION-007 to implemented with executable Cypress evidence.

## Commands run
1. `pwsh -NoLogo -NoProfile -File .\\cypress.ps1 -Spec cypress/e2e/general/ui-foundation-shell-navigation/spec.cy.ts -Browser chrome -RequireE2EHooks` (initial fail reproduced)
2. `npm run build` (ui.web)
3. `pwsh -NoLogo -NoProfile -File .\\cypress.ps1 -Spec cypress/e2e/general/ui-foundation-shell-navigation/spec.cy.ts -Browser chrome -RequireE2EHooks` (still failing while stale reused server)
4. `Get-NetTCPConnection ... Stop-Process` (force recycle port 17880 listener)
5. `pwsh -NoLogo -NoProfile -File .\\cypress.ps1 -Spec cypress/e2e/general/ui-foundation-shell-navigation/spec.cy.ts -Browser chrome -RequireE2EHooks` (pass, 5/5)
6. `go test ./internal/app -count=1` (pass)
7. `go test ./tests -count=1` (pass)
8. `openspec validate --all` (pass)

## Result
- UI-FOUNDATION-SHELL-NAVIGATION-007: implemented with passing managed Cypress proof.
- No mandatory gate failures remain for this issue.