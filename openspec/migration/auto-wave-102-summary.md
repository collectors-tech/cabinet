# Auto Wave 102 Summary

- Issue: #299
- Section: general
- Spec IDs: SETUP-WIZ-008
- Status: done

## What changed
- Added explicit `Use Defaults` action to Setup Wizard welcome screen (`setup-use-defaults`).
- Implemented deterministic defaults-complete submission path in sign-in setup flow:
  - instance name `Cabinet Local`
  - profile key `default`
  - local storage mode + default storage dir
  - runtime auto mode
  - local auth mode
  - feature toggles enabled
- Completion state now shows explicit defaults-applied feedback when defaults path is used.
- Added Cypress proof for defaults path (`UC-SW-37`) and welcome action visibility update.
- Resolved duplicate requirement ID collision in setup spec by assigning runtime URL metadata sync to append-only `SETUP-WIZ-020`.
- Updated traceability for `SETUP-WIZ-008` (Use Defaults) and `SETUP-WIZ-020` (runtime URL sync).

## Commands run
1. `pwsh -File d:\projects\collectors-tech\cabinet\cypress.ps1 -Spec cypress/e2e/general/setup-wizard-first-run/spec.cy.ts -Browser chrome -RequireE2EHooks` (fail-first)
2. `pwsh -File .\scripts\build-ui-static.ps1`
3. `go build -o bin/cabinet.exe ./cmd/cabinet`
4. `pwsh -File d:\projects\collectors-tech\cabinet\cypress.ps1 -Spec cypress/e2e/general/setup-wizard-first-run/spec.cy.ts -Browser chrome -RequireE2EHooks` (green)
5. `go test ./internal/app -count=1`
6. `go test ./tests -count=1`
7. `openspec validate --all`

## Gate results
- Managed Cypress setup-wizard: PASS (`33 passing`)
- `go test ./internal/app -count=1`: PASS
- `go test ./tests -count=1`: PASS
- `openspec validate --all`: PASS (`5 passed, 0 failed`)

## Commit
- Commit: 8e49df8af8ae3a6287c2a294ad9a37c1bf54c741
- Branch: main
