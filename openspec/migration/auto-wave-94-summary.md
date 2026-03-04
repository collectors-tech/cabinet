# Auto Wave 94 Summary

- Issue: #292
- Scope: Frontline Algolia config discovery + drift-safe fallback from site assets.
- Requirement IDs: `INTEGRATION-015`, `UC-AU-04`, `UC-AU-05`

## Fail-first
- Added API tests in `internal/app/provider_frontline_discovery_api_test.go`.
- Initial run failed with `404` because `/api/providers/frontline/discovery` endpoint did not exist.

## Implementation
- Added `POST /api/providers/frontline/discovery` in `internal/app/app.go`.
- Implemented Frontline discovery workflow:
  - fetch discovery asset(s)
  - parse `Glgoliasearch('<appId>','<searchKey>')`
  - parse index name entries
  - generate deterministic config hash + discovered timestamp
- Implemented drift-safe fallback:
  - persist last-known-good config in `app_state` key `frontline_algolia_last_known_good`
  - on discovery/parse failure, return cached config with `fallback_used=true` and warning
- Added helper functions for provider asset fetch, parser, cache persist/load.

## Validation
- `go test ./internal/app -run TestFrontlineDiscovery -count=1` ✅
- `pwsh -NoLogo -NoProfile -File .\\cypress.ps1 -Spec cypress/e2e/integrations/ui-screen-integrations/spec.cy.ts -Browser chrome -RequireE2EHooks` ✅
- `go test ./internal/app -count=1` ✅
- `go test ./tests -count=1` ✅
- `openspec validate --all` ✅
- `go build -o bin/cabinet.exe ./cmd/cabinet` ✅

## Traceability
- Updated `INTEGRATION-015` mapping to `openspec/specs/integrations/provider-au-webshops/spec.md`
- Evidence now points to:
  - `TestFrontlineDiscoveryParsesAssetAndCachesLastKnownGood`
  - `TestFrontlineDiscoveryFallsBackToCachedConfigOnDriftFailure`