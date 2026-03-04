# Auto Wave 95 Summary

- Issue: #293
- Scope: Hobbytech Shopify/Boost provider run support with session drift recovery.
- Requirement IDs: `INTEGRATION-016`, `UC-AU-06`, `UC-AU-07`

## Fail-first
- Added API tests in `internal/app/provider_hobbytech_run_api_test.go`.
- Initial run failed with `404` because `/api/providers/hobbytech/run` did not exist.

## Implementation
- Added `POST /api/providers/hobbytech/run` in `internal/app/app.go`.
- Implemented Hobbytech runtime workflow:
  - discovery parsing from asset scripts (`shop`, `sid`, `template`, `widget`)
  - configurable `search_url` execution with query/page/limit/session inputs
  - deterministic candidate extraction + pagination traversal
  - cache last-known-good Hobbytech config (`hobbytech_boost_last_known_good` in `app_state`)
  - bounded drift recovery retry using fallback discovery assets on first failed run attempt
  - drift warning + `drift_recovered` outcome in response payload

## Validation
- `go test ./internal/app -run TestHobbytechRun -count=1` ✅
- `pwsh -NoLogo -NoProfile -File .\\cypress.ps1 -Spec cypress/e2e/integrations/ui-screen-integrations/spec.cy.ts -Browser chrome -RequireE2EHooks` ✅
- `go test ./internal/app -count=1` ✅
- `go test ./tests -count=1` ✅
- `openspec validate --all` ✅
- `go build -o bin/cabinet.exe ./cmd/cabinet` ✅

## Traceability
- Updated `INTEGRATION-016` row to point to `openspec/specs/integrations/provider-au-webshops/spec.md`
- Evidence updated to new Hobbytech runtime tests.