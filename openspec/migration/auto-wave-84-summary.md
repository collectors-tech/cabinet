# Auto Wave 84 Summary

- Issue: #253
- Scope: Setup Wizard integrations baseline step and feature toggle persistence
- Requirement IDs: SETUP-WIZ-016

## What Changed

1. Setup integrations baseline step (UI)
- Expanded setup wizard to 6 steps:
  - Identity -> Storage -> Runtime -> Auth -> Integrations -> Review
- Added integrations baseline controls in `ui.web/src/features/auth/sign-in/index.tsx`:
  - `setup-feature-scanner`
  - `setup-feature-chat`
  - `setup-feature-providers`
- Added explicit guidance copy:
  - integrations are optional and editable later in Settings
- Review step now summarizes enabled feature set.

2. Runtime setup feature toggle contract (backend)
- Extended setup request contract (`internal/app/app.go`):
  - `feature_chat`
  - `feature_providers`
  - `feature_scanner`
- Added default-enabled semantics when request fields are omitted.
- Persisted selected feature toggle values into `cabinet.json` `features.*` payload.

3. Tests + OpenSpec/traceability
- Added backend persistence contract test:
  - `TestRuntimeSetupCompletePersistsFeatureToggles`
- Added and greened E2E coverage:
  - `UC-SW-26` integrations defaults + guidance
  - `UC-SW-27` integrations persistence
  - `UC-SW-28` integrations optional progression
- Updated existing setup wizard E2E flow assertions for 6-step progression.
- Updated OpenSpec + traceability:
  - `SETUP-WIZ-016` implemented with Cypress + Go evidence.

## Commands Run

- `go test ./internal/app -run "TestRuntimeSetupCompletePersists(ClerkAuthConfiguration|FeatureToggles)" -count=1`
- `npm run build` (in `ui.web`)
- `pwsh -NoLogo -NoProfile -File .\\cypress.ps1 -RequireE2EHooks -Spec cypress/e2e/general/setup-wizard-first-run/spec.cy.ts -Browser chrome`
- `go test ./internal/app -count=1`
- `go test ./tests -count=1`
- `openspec validate --all`

## Results

- Target Cypress setup-wizard spec: 24 passing, 0 failing.
- `go test ./internal/app -count=1`: pass.
- `go test ./tests -count=1`: pass.
- `openspec validate --all`: pass.
