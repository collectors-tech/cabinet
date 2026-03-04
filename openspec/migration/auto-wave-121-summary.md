# Auto Wave 121 Summary (#236)

## Issue
- #236 `[Spec Backlog] Configure AU webshop provider allowlist (Bonza + hobby stores)`

## Spec IDs
- INTEGRATION-011
- PROVIDER-AU-WEBSHOPS-004

## What changed
- Extended AU webshop allowlist contract to include all approved providers:
  - `bonzaslotcars.com.au`
  - `frontlinehobbies.com.au`
  - `hobbytechtoys.com.au`
  - `andrewshobbies.com.au`
  - `voglers.com.au`
  - `acercmodels.com`
  - `mrtoys.com.au`
  - `hobbyco.com.au`
  - `metrohobbies.com.au`
- Added explicit executable requirement for `PROVIDER-AU-WEBSHOPS-004` to enforce deterministic registry allowlist output.
- Updated traceability mapping to mark requirement as implemented with executable proof.

## Commands run
1. `go test ./internal/app -count=1`
2. `go test ./tests -count=1`
3. `openspec validate --all`
4. `pwsh -File .\\scripts\\build-cabinet.ps1`

## Results
- All commands passed.

## Notes
- Existing runtime provider registry test (`TestWave4ProvidersRegistryContract`) provides deterministic proof for `PROVIDER-AU-WEBSHOPS-004`.
