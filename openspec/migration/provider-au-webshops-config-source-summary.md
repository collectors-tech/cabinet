# Provider AU Webshops Config Source Summary (#237)

## Issue
- #237 `[Spec Backlog] Move AU webshop domains out of hardcoded code into content/config source`

## Spec IDs
- PROVIDER-AU-WEBSHOPS-005
- PROVIDER-AU-WEBSHOPS-004

## Changes
- Added OpenSpec executable requirement `PROVIDER-AU-WEBSHOPS-005` for config-driven domain source and deterministic fallback.
- Added fail-first contract test proving registry domains can come from profile setting `integration.au_webshops.domains`.
- Implemented runtime domain resolver:
  - Reads `integration.au_webshops.domains` (comma-separated)
  - Normalizes domains
  - Deduplicates values preserving order
  - Falls back deterministically to default allowlist when missing/invalid
- Updated traceability for `PROVIDER-AU-WEBSHOPS-005` to implemented with executable proof.

## Commands run
1. `go test ./internal/app -run TestWave4AUWebshopDomainsConfigSourceContract -count=1` (fail-first then pass)
2. `go test ./internal/app -run TestWave4ProvidersRegistryContract -count=1`
3. `go test ./internal/app -count=1`
4. `go test ./tests -count=1`
5. `openspec validate --all`
6. `pwsh -File .\\scripts\\build-cabinet.ps1`

## Results
- All mandatory gates passed.
- Config-driven domain contract proven via runtime test.
