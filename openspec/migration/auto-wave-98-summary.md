# Auto Wave 98 Summary

- Issue: #297
- Scope: BigCommerce provider family support (`PROVIDER-FAMILY-006`, `UC-PF-07`, `UC-PF-08`)

## What Changed
- Added fail-first BigCommerce API contract tests:
  - `internal/app/provider_bigcommerce_api_test.go`
  - `TestBigCommerceRegistryExposesFamilyAndActiveMode`
  - `TestBigCommerceRunStorefrontModeDeclaresCapabilityLimits`
  - `TestBigCommerceRunTokenModeUnlocksRicherDepth`
- Added BigCommerce provider runtime endpoint:
  - `POST /api/providers/bigcommerce/run`
- Added BigCommerce storefront/token execution paths:
  - storefront-access mode when token absent
  - token-enabled GraphQL mode when token present
  - deterministic run response with `auth_mode`, `data_depth_source`, `capability_limits`, and normalized candidates
- Updated registry payload classification:
  - voglers domain mapped to `api_family=bigcommerce`
  - runtime `active_mode` reflects token state (`storefront_public` / `token_enabled`)
- Updated traceability:
  - `PROVIDER-FAMILY-006` -> implemented with executable test evidence

## Commands Run
1. `go test ./internal/app -run "TestBigCommerce" -count=1` (red, then green)
2. `pwsh -NoLogo -NoProfile -ExecutionPolicy Bypass -File .\\cypress.ps1 -Spec cypress/e2e/integrations/ui-screen-integrations/spec.cy.ts -Browser chrome -RequireE2EHooks`
3. `go test ./internal/app -count=1`
4. `go test ./tests -count=1`
5. `openspec validate --all`
6. `go build -o bin/cabinet.exe ./cmd/cabinet`

## Gate Results
- Targeted Cypress: pass (5 passing, 0 failing)
- `go test ./internal/app -count=1`: pass
- `go test ./tests -count=1`: pass
- `openspec validate --all`: pass
- `go build -o bin/cabinet.exe ./cmd/cabinet`: pass
