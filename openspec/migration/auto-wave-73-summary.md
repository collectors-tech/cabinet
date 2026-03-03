# Auto Wave 73 Summary

- Issue: #229
- Scope: passkey/WebAuthn passwordless sign-in requirement
- Status: done

## What changed
- Added `Sign in with Passkey` interaction in auth form.
- Added deterministic fallback error for unavailable passkey environments.
- Added E2E passkey success + fallback coverage in onboarding-auth suite.
- Updated OpenSpec UC mapping (UC-ONB-08/09) and traceability for `UI-SCREEN-ONBOARDING-AUTH-008`.

## Commands run
1. `pwsh -NoLogo -NoProfile -File .\\scripts\\build-ui-static.ps1`
2. `pwsh -NoLogo -NoProfile -File .\\cypress.ps1 -Spec cypress/e2e/general/ui-screen-onboarding-auth/spec.cy.ts -Browser chrome`
3. `go test ./internal/app -count=1`
4. `go test ./tests -count=1`
5. `openspec validate --all`

## Results
- onboarding-auth Cypress suite: 9 passing, 0 failing
- `go test ./internal/app`: pass
- `go test ./tests`: pass
- `openspec validate --all`: pass

## IDs moved to implemented
- UI-SCREEN-ONBOARDING-AUTH-008

## Blockers
- none
