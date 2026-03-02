# Accessibility Action Labels Summary

- Issue: #218
- Spec: `openspec/specs/general/ui-foundation-accessibility/spec.md`
- Requirement ID: `UI-FOUNDATION-ACCESSIBILITY-003`

## Implementation
- Added explicit accessible name to password visibility toggle button:
  - `ui.web/src/components/password-input.tsx`
  - `aria-label` now switches between `Show password` and `Hide password`.
- Added explicit accessible name to mobile top navigation menu trigger:
  - `ui.web/src/components/layout/top-nav.tsx`
  - `aria-label='Open navigation menu'`.

## E2E Proof
- Added Cypress suite:
  - `ui.web/cypress/e2e/general/ui-foundation-accessibility/spec.cy.ts`
- Coverage validates icon-action controls expose non-empty accessible names on:
  - sign-in screen
  - mobile inventory header flow

## Gates
1. `pwsh -NoLogo -NoProfile -File .\\cypress.ps1 -Spec cypress/e2e/general/ui-foundation-accessibility/spec.cy.ts -Browser chrome -RequireE2EHooks` ✅
2. `go test ./internal/app -count=1` ✅
3. `go test ./tests -count=1` ✅
4. `openspec validate --all` ✅

## Traceability
- Updated `openspec/traceability.md`:
  - `UI-FOUNDATION-ACCESSIBILITY-003` -> implemented

## Commit
- Commit: pending
- Push-proof: pending
