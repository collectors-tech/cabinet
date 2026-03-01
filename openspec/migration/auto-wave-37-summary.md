# Auto Wave 37 Summary

- Issue: `#172`
- Scope: close shippable theme + locale shell contracts in UI foundation

## Requirement IDs moved to implemented

- `UI-FOUNDATION-THEME-RTL-I18N-001`
- `UI-FOUNDATION-THEME-RTL-I18N-004`

## Commands Run

1. `pwsh -NoLogo -NoProfile -File .\cypress.ps1 -Spec cypress/e2e/general/ui-foundation-theme-rtl-i18n/spec.cy.ts -Browser chrome`
2. `npm run build` (from `ui.web`)
3. `pwsh -NoLogo -NoProfile -File .\cypress.ps1 -Spec cypress/e2e/general/ui-foundation-theme-rtl-i18n/spec.cy.ts -Browser chrome`
4. `go test ./internal/app -count=1`
5. `go test ./tests -count=1`
6. `openspec validate --all`

## Key Results

- Added spec-aligned E2E suite:
  - `ui.web/cypress/e2e/general/ui-foundation-theme-rtl-i18n/spec.cy.ts`
- Added deterministic header locale switch selectors in:
  - `ui.web/src/components/language-switch.tsx`
- Proved:
  - theme selection persists across reload
  - header locale switch remains available and fallback to supported locale stays stable
- Validation gates passed:
  - Cypress target spec: `2 passing`
  - `go test ./internal/app -count=1`: pass
  - `go test ./tests -count=1`: pass
  - `openspec validate --all`: pass

## Remaining Partial IDs in this spec

- `UI-FOUNDATION-THEME-RTL-I18N-002`
- `UI-FOUNDATION-THEME-RTL-I18N-003`

## Blockers

- None for this wave.
