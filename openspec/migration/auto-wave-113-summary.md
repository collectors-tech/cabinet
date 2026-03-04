# Auto Wave 113 Summary

- Issue: #256
- Scope: Stabilize setup/auth helper path used by accessibility and setup-wizard journeys.
- Requirement IDs: UI-FOUNDATION-ACCESSIBILITY-001, UI-FOUNDATION-ACCESSIBILITY-002, UI-FOUNDATION-ACCESSIBILITY-003

## Changes
- Updated `ui.web/cypress/support/commands.ts` (`useBootstrappedProfile`) to:
  - set active profile via `PUT /api/profiles/active`
  - use deterministic sign-in route with redirect to target
  - perform credential sign-in before target assertions
  - retain safe fallback for profile-selection button variants
- Updated `ui.web/cypress/e2e/general/ui-foundation-accessibility/spec.cy.ts` to use deterministic E2E reset/bootstrap/profile helper flow.

## Commands Run
- `pwsh -File .\\cypress.ps1 -Spec "cypress/e2e/general/ui-foundation-accessibility/spec.cy.ts" -Browser chrome` (pass)
- `pwsh -File .\\cypress.ps1 -Spec "cypress/e2e/general/setup-wizard-first-run/spec.cy.ts" -Browser chrome` (pass)
- `go test ./internal/app -count=1` (pass)
- `go test ./tests -count=1` (pass)
- `openspec validate --all` (pass)
- `go build -o bin/cabinet.exe ./cmd/cabinet` (pass)

## Results
- Accessibility spec: 4 passing, 0 failing
- Setup wizard spec: 33 passing, 0 failing
- Go tests: all passing in touched suites
- OpenSpec validation: all passing

## Blockers
- None
