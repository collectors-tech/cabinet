# Auto Wave 86 Summary

- Issue: #255
- Scope: Setup Wizard Step 8 launch-confirmation completion actions
- Requirement IDs: `SETUP-WIZ-018`

## What Changed

1. Setup completion launch actions
- Updated completion state in `ui.web/src/features/auth/sign-in/index.tsx` to render deterministic actions:
  - `Open Cabinet` (`setup-open-cabinet`)
  - `Open Config Folder` (`setup-open-config-folder`)
  - `Finish` (`setup-finish`)
- Added completion metadata rendering:
  - instance identity (`setup-complete-instance-name`)
  - profile key (`setup-complete-profile-key`)
  - deterministic feedback region (`setup-complete-feedback`)

2. Runtime setup-complete/import contract metadata
- Updated `internal/app/app.go` setup responses to return `instance_name` and `profile_key` on:
  - `POST /api/runtime/setup-complete`
  - `POST /api/runtime/setup-import`

3. E2E proof for completion-step contract
- Added and greened:
  - `UC-SW-32 setup-wizard-completion-summary shows launch confirmation actions`
  - `UC-SW-33 setup-wizard-open-cabinet exits completion state`
  - `UC-SW-34 setup-wizard-open-config-folder shows feedback`
- Updated legacy completion label expectation to `Open Cabinet`.

4. OpenSpec + traceability updates
- `openspec/specs/general/setup-wizard-first-run/spec.md`:
  - `UC-SW-32..34` moved to `implemented`.
- `openspec/traceability.md`:
  - `SETUP-WIZ-018` moved to `implemented` with Cypress proof.

## Commands Run

- `pwsh -NoLogo -NoProfile -File ./cypress.ps1 -Spec cypress/e2e/general/setup-wizard-first-run/spec.cy.ts -Browser chrome -RequireE2EHooks`
- `go test ./internal/app -count=1`
- `go test ./tests -count=1`
- `openspec validate --all`

## Results

- Managed Cypress setup-wizard spec: `30 passing, 0 failing`.
- `go test ./internal/app -count=1`: pass.
- `go test ./tests -count=1`: pass.
- `openspec validate --all`: pass.
