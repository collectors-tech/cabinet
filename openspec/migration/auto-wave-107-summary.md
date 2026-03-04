# Auto Wave 107 Summary

- **Issue**: #301
- **Title**: [Execution] Add CLI flag to disable browser auto-open on startup (automation/Cypress mode)
- **Spec IDs**: `RUNTIME-CORE-007`
- **Status**: done

## What changed

- Added CLI runtime launch resolution for browser behavior:
  - `--no-open-browser` disables browser launch deterministically.
  - Existing `CABINET_OPEN_BROWSER` env override remains supported.
  - Startup now logs explicit note when browser auto-open is disabled.
- Updated managed Cypress launcher script to pass `--no-open-browser` when starting runtime.
- Added/updated tests that prove CLI suppression behavior and managed-runtime policy.

## Commands run

1. `go test ./tests -run TestCypressScriptDisablesBrowserAutoOpenForManagedRuns -count=1` (fail-first: red baseline)
2. `go test ./cmd/cabinet -count=1`
3. `pwsh -File ./cypress.ps1 -Spec "cypress/e2e/general/ui-foundation-shell-navigation/spec.cy.ts" -Browser chrome`
4. `go test ./internal/app -count=1`
5. `go test ./tests -count=1`
6. `openspec validate --all`
7. `go build -o bin/cabinet.exe ./cmd/cabinet`

## Results

- `go test ./cmd/cabinet -count=1`: pass
- Managed Cypress spec (`ui-foundation-shell-navigation`): pass (10/10)
- `go test ./internal/app -count=1`: pass
- `go test ./tests -count=1`: pass
- `openspec validate --all`: pass (5/5 items)
- `go build -o bin/cabinet.exe ./cmd/cabinet`: pass

## Traceability update

- `RUNTIME-CORE-007` moved to **implemented** with proof:
  - `cmd/cabinet/main_cli_test.go` (`TestResolveBrowserLaunch`)
  - `cmd/cabinet/main_test.go` (`TestOpenBrowserEnabled`)
  - `tests/runtime_script_policy_test.go` (`TestCypressScriptDisablesBrowserAutoOpenForManagedRuns`)
  - Managed Cypress run evidence via `cypress.ps1`.
