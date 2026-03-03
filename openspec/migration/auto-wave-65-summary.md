# Auto Wave 65 Summary

- Issue: #285
- Scope: Setup Wizard deterministic `cabinet.json` schema contract + Clerk required-field validation.
- Requirement IDs: `SETUP-WIZ-007` (UC-SW-08, UC-SW-09)

## What Changed
1. Added fail-first Cypress proof for setup schema + Clerk validation:
   - `ui.web/cypress/e2e/general/setup-wizard-first-run/spec.cy.ts`
2. Implemented setup payload contract on runtime API:
   - `/api/runtime/setup-complete` now accepts JSON payload, validates required fields, and returns deterministic validation envelope on failure.
3. Implemented deterministic setup config schema write:
   - Required sections written: `instance`, `storage`, `runtime`, `auth`, `bootstrap`, `features`, `meta`.
4. Enforced Clerk required-field validation:
   - `auth_mode=clerk` without `clerk_publishable_key` returns `400` with `error_code=SETUP_CLERK_PUBLISHABLE_KEY_REQUIRED`.
5. Added test hook for setup-config inspection:
   - `GET /api/test/runtime/setup-config`.
6. Added/updated API unit tests for setup contract and validation.
7. Updated OpenSpec mapping and traceability for implemented proof.

## Commands Run
- `pwsh -NoLogo -NoProfile -File .\\cypress.ps1 -Spec cypress/e2e/general/setup-wizard-first-run/spec.cy.ts -Browser chrome`
- `pwsh -NoLogo -NoProfile -File .\\scripts\\build-ui-static.ps1`
- `go test ./internal/app -count=1`
- `go test ./tests -count=1`
- `openspec validate --all`

## Results
- Cypress targeted setup-wizard spec: PASS (2 passed, 0 failed)
- `go test ./internal/app -count=1`: PASS
- `go test ./tests -count=1`: PASS
- `openspec validate --all`: PASS

## Notes
- UI static was rebuilt so embedded runtime serves updated setup wizard behavior.
