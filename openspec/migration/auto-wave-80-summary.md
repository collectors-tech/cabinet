# Auto Wave 80 Summary

- Issue: #249
- Scope: Setup Wizard identity step (instance/profile + config-path preview + inline validation)
- Requirement IDs: SETUP-WIZ-012

## What Changed

1. Runtime identity behavior
- Made `profile_key` optional in setup-complete request handling.
- Added deterministic profile-key derivation from instance name when profile key is blank.
- Added API test proving derived profile key persistence.

2. Setup wizard identity UI
- Added config-path preview on identity step (`setup-config-path-preview`).
- Added inline validation on identity step to block next transition when instance name is blank.
- Preserved entered state and kept step navigation deterministic.

3. E2E and OpenSpec updates
- Added Cypress coverage for identity path preview, optional profile derivation, and inline validation.
- Added OpenSpec requirement `SETUP-WIZ-012` and moved UC-SW-14/15/16 to implemented.
- Updated traceability to implemented for `SETUP-WIZ-012` with Cypress + API evidence.

## Commands Run

- `pwsh -NoLogo -NoProfile -File .\\cypress.ps1 -Spec cypress/e2e/general/setup-wizard-first-run/spec.cy.ts -Browser chrome` (fail-first baseline)
- `go test ./internal/app -run "TestRuntimeSetupCompleteDerivesProfileKeyWhenBlank|TestRuntimeSetupStatusAndCompleteContract|TestRuntimeSetupImportExistingConfigContract" -count=1`
- `pwsh -NoLogo -NoProfile -File .\\scripts\\build-ui-static.ps1`
- `pwsh -NoLogo -NoProfile -File .\\cypress.ps1 -Spec cypress/e2e/general/setup-wizard-first-run/spec.cy.ts -Browser chrome`
- `go test ./internal/app -count=1`
- `go test ./tests -count=1`
- `openspec validate --all`

## Results

- Target Cypress setup-wizard spec: 12 passing, 0 failing.
- `go test ./internal/app -count=1`: pass.
- `go test ./tests -count=1`: pass.
- `openspec validate --all`: pass.
