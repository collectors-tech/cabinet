# Auto Wave 90 Summary

- Issue: #291
- Requirement IDs: `OPS-002`, `UC-AU-03`
- Spec path: `openspec/specs/integrations/provider-au-webshops/spec.md`

## What changed
- Added scanner query-set `items_per_page` persistence (`scanner_query_sets.items_per_page`) with migration-safe default.
- Added runtime safe-cap policy for items-per-page:
  - conservative default when unset: `24`
  - Bonza scope cap: `36`
  - general cap: `50`
- Added run summary contract fields:
  - `items_per_page_requested`
  - `items_per_page_effective`
  - `items_per_page_warning` (when clamped)
- Added provider-level settings wiring:
  - new key: `integration.<provider>.items_per_page`
  - scanner run now applies provider setting override before execution.
- Updated Integrations UI provider settings dialog to include `Items per page` value and persist it.

## Fail-first + proof
- Added `TestScannerRunItemsPerPageSummaryAppliesSafeCap` in `internal/app/scanner_api_test.go`.
- The test initially failed before implementation and now passes with deterministic assertions for requested/effective/warning fields.

## Commands run
1. `go test ./internal/app -run TestScannerRunItemsPerPageSummaryAppliesSafeCap -count=1`
2. `pwsh -NoLogo -NoProfile -File .\\cypress.ps1 -Spec cypress/e2e/integrations/ui-screen-integrations/spec.cy.ts -Browser chrome -RequireE2EHooks`
3. `go test ./internal/app -count=1`
4. `go test ./tests -count=1`
5. `openspec validate --all`
6. `./scripts/build-ui-static.ps1`
7. `go build -o bin/cabinet.exe ./cmd/cabinet`

## Results
- Targeted API test: pass.
- Integrations Cypress spec: 5 passing.
- `go test ./internal/app`: pass.
- `go test ./tests`: pass.
- `openspec validate --all`: pass.
- Static UI build + `bin/cabinet.exe` build: pass.

## Blockers
- None.
