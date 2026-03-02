# Auto Wave 38 Summary

- Issue: `#172`
- Scope: close remaining i18n + RTL partial contracts in UI foundation

## Requirement IDs moved to implemented

- `UI-FOUNDATION-THEME-RTL-I18N-002`
- `UI-FOUNDATION-THEME-RTL-I18N-003`

## Commands Run

1. `pwsh -NoLogo -NoProfile -File .\cypress.ps1 -Spec cypress/e2e/general/ui-foundation-theme-rtl-i18n/spec.cy.ts -Browser chrome`
2. `npm run build` (from `ui.web`)
3. `pwsh -NoLogo -NoProfile -File .\cypress.ps1 -Spec cypress/e2e/general/ui-foundation-theme-rtl-i18n/spec.cy.ts -Browser chrome`
4. `go test ./internal/app -count=1`
5. `go test ./tests -count=1`
6. `openspec validate --all`

## Key Results

- Expanded E2E suite:
  - `ui.web/cypress/e2e/general/ui-foundation-theme-rtl-i18n/spec.cy.ts`
  - added explicit proof cases for `002` and `003`
- Implemented locale + RTL runtime behavior:
  - added Arabic locale option to language switch
  - configured i18n with `supportedLngs: ['en', 'ar']`
  - added document `dir`/`lang` updates on language init + change
  - added Arabic common locale resource with English fallback for missing namespaces
- Validation gates passed:
  - Cypress target spec: `4 passing`
  - `go test ./internal/app -count=1`: pass
  - `go test ./tests -count=1`: pass
  - `openspec validate --all`: pass

## Blockers

- None.
