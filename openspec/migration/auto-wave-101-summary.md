# Auto Wave 101 Summary

- Issue: #300
- Section: general
- Spec IDs: SETUP-WIZ-019
- Status: done

## What changed
- Added deterministic Cypress setup helpers:
  - `cy.e2eSetSetupState("missing" | "present")`
  - `cy.e2eCompleteSetupHelper(...)` (calls `/api/runtime/setup-complete` with deterministic defaults)
- Added global Cypress setup-gate policy for non-setup specs:
  - one-time `before()` in `cypress/support/e2e.ts` seeds setup-complete state
  - setup-flow specs remain excluded from automatic bypass
- Added explicit setup-helper coverage in setup wizard spec:
  - `UC-SW-35` bypass path proof
  - `UC-SW-36` completion-helper path proof
- Updated OpenSpec setup wizard UC matrix and traceability with executable evidence for `SETUP-WIZ-019`.

## Commands run
1. `pwsh -File d:\projects\collectors-tech\cabinet\cypress.ps1 -Spec cypress/e2e/general/setup-wizard-first-run/spec.cy.ts -Browser chrome -RequireE2EHooks`
2. `pwsh -File d:\projects\collectors-tech\cabinet\cypress.ps1 -Spec cypress/e2e/general/ui-foundation-shell-navigation/spec.cy.ts -Browser chrome -RequireE2EHooks`
3. `go test ./internal/app -count=1`
4. `go test ./tests -count=1`
5. `openspec validate --all`
6. `go build -o bin/cabinet.exe ./cmd/cabinet`

## Gate results
- Managed Cypress setup-wizard: PASS (`32 passing`)
- Managed Cypress shell-navigation: PASS (`10 passing`)
- `go test ./internal/app -count=1`: PASS
- `go test ./tests -count=1`: PASS
- `openspec validate --all`: PASS (`5 passed, 0 failed`)

## Commit
- Commit: 0d4e4d11286e77dacb553b8a3bb762d83a860063
- Branch: main
