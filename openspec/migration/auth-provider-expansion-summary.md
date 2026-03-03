# Auth Provider Expansion Summary

- Issue: #228
- OpenSpec IDs: UI-SCREEN-ONBOARDING-AUTH-006, UI-SCREEN-ONBOARDING-AUTH-007

## Delivered
- Added runtime API contract `GET /api/auth/provider-options` with deterministic identity mode + provider states.
- Added E2E test override hook `POST /api/test/auth/provider-options` for deterministic Cypress scenarios.
- Updated sign-in UI to:
  - fetch runtime provider options,
  - show explicit identity mode indicator,
  - render Google/Apple/Microsoft buttons with deterministic enabled/disabled state.
- Added API tests for default/env-driven provider options and E2E override behavior.
- Added/updated Cypress proofs for 006/007 and aligned 009 flow to setup completion state contract.

## Validation
- Managed Cypress: `ui.web/cypress/e2e/general/ui-screen-onboarding-auth/spec.cy.ts` (7 passing)
- `go test ./internal/app -count=1` (pass)
- `go test ./tests -count=1` (pass)
- `openspec validate --all` (pass)

## Traceability
- `UI-SCREEN-ONBOARDING-AUTH-006` -> implemented
- `UI-SCREEN-ONBOARDING-AUTH-007` -> implemented
