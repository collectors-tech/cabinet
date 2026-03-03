# Auto Wave 81 Summary

- Issue: #250
- Scope: Setup Wizard storage step (exe-local default, custom path validation, portable mode)
- Requirement IDs: SETUP-WIZ-013

## What Changed

1. Runtime setup storage contract
- Added setup status fields exposing default storage path:
  - `default_storage_data_dir`
  - `default_storage_media_dir`
- Added runtime storage validation API:
  - `POST /api/runtime/setup-storage-validate`
  - deterministic payload includes `writable`, `free_space_ok`, `free_space_status`, and message
- Extended setup completion payload contract:
  - accepts `storage_mode`, `storage_data_dir`, `portable_mode`
  - persists selected storage paths to `cabinet.json`

2. Setup wizard storage UI
- Expanded wizard to 4 steps (Identity -> Storage -> Auth -> Review).
- Added storage mode selector (`exe_local` / `custom`).
- Added storage path preview and custom directory input.
- Added portable mode toggle.
- Added inline validation for blank custom path.
- Added runtime writable-path check before advancing from storage step.

3. Tests + OpenSpec/traceability
- Added Cypress coverage:
  - UC-SW-17 storage defaults
  - UC-SW-18 custom path validation
  - UC-SW-19 storage persistence
- Added/updated API tests:
  - `TestRuntimeSetupStorageValidateContract`
  - `TestRuntimeSetupCompletePersistsSelectedStoragePath`
- Added OpenSpec requirement `SETUP-WIZ-013`.
- Updated traceability to implemented for `SETUP-WIZ-013` with executable proof.

## Commands Run

- `pwsh -NoLogo -NoProfile -File .\\cypress.ps1 -Spec cypress/e2e/general/setup-wizard-first-run/spec.cy.ts -Browser chrome` (fail-first baseline)
- `go test ./internal/app -run "TestRuntimeSetupStorageValidateContract|TestRuntimeSetupCompletePersistsSelectedStoragePath|TestRuntimeSetupCompleteDerivesProfileKeyWhenBlank" -count=1`
- `pwsh -NoLogo -NoProfile -File .\\scripts\\build-ui-static.ps1`
- `pwsh -NoLogo -NoProfile -File .\\cypress.ps1 -Spec cypress/e2e/general/setup-wizard-first-run/spec.cy.ts -Browser chrome`
- `go test ./internal/app -count=1`
- `go test ./tests -count=1`
- `openspec validate --all`

## Results

- Target Cypress setup-wizard spec: 15 passing, 0 failing.
- `go test ./internal/app -count=1`: pass.
- `go test ./tests -count=1`: pass.
- `openspec validate --all`: pass.
