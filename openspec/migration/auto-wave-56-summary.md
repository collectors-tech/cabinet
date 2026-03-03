# Auto Wave 56 Summary

- Issue: #263
- Status: done
- Requirement IDs implemented:
  - UI-SCREEN-ONBOARDING-AUTH-010

## Why selected
- High-priority open bug for first-time user onboarding discoverability.

## Changes
- Added explicit Create account CTA on sign-in footer to route users to `/sign-up`.
- Added/updated E2E scenario for onboarding auth create-account visibility.
- Updated OpenSpec requirement and traceability mapping.
- Rebuilt embedded static UI bundle.

## Commands Run
- `pwsh -NoLogo -NoProfile -File .\\cypress.ps1 -Spec cypress/e2e/general/ui-screen-onboarding-auth/spec.cy.ts -Browser chrome`
- `go test ./internal/app -count=1`
- `go test ./tests -count=1`
- `openspec validate --all`

## Results
- Cypress onboarding-auth spec: 5 passing, 0 failing.
- `go test ./internal/app -count=1`: pass.
- `go test ./tests -count=1`: pass.
- `openspec validate --all`: pass.

## Evidence Paths
- `ui.web/cypress/e2e/general/ui-screen-onboarding-auth/spec.cy.ts`
- `openspec/specs/general/ui-screen-onboarding-auth/spec.md`
- `openspec/traceability.md`
- `ui.web/src/features/auth/sign-in/index.tsx`

## Blockers
- None.
