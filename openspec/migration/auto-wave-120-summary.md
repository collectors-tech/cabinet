# Auto Wave 120 Summary

- Issue: #235
- Scope: Saved searches for Bonza/Amazon/eBay with provider filters, run-now + scheduled refresh, and discoveries/wishlist handoff.
- Spec IDs: DEFAULT-SITE-SEARCH-004, DEFAULT-SITE-SEARCH-005, DEFAULT-SITE-SEARCH-006

## Changes
- `openspec/specs/integrations/default-site-search/spec.md`
- `ui.web/src/features/scanner/index.tsx`
- `ui.web/cypress/e2e/integrations/default-site-search/spec.cy.ts`
- `openspec/traceability.md`
- `openspec/migration/saved-searches-provider-filters-summary.md`
- `openspec/migration/saved-searches-provider-filters-changed-files.txt`

## Gate Evidence
- Managed Cypress: `pwsh -File .\\cypress.ps1 -Spec cypress/e2e/integrations/default-site-search/spec.cy.ts -Browser chrome` -> PASS (3/3)
- `go test ./internal/app -count=1` -> PASS
- `go test ./tests -count=1` -> PASS
- `openspec validate --all` -> PASS
- `pwsh -File .\\scripts\\build-cabinet.ps1` -> PASS

## Notes
- Fail-first repro was captured via missing Cypress spec path before implementation.
- Traceability moved IDs 004-006 to implemented with executable E2E proof.
