# Auto Wave 93 Summary

- Issue: #287
- Scope: enforce project-local runtime executable preference for managed Cypress lifecycle.
- Requirement IDs: `RUNTIME-CORE-006`

## Fail-first
- Added script policy tests:
  - `TestCypressScriptPrefersProjectBinRuntimePath`
  - `TestCypressScriptRejectsEphemeralTempRuntimePathByDefault`
- Initial run failed because `cypress.ps1` defaulted to `go run ./cmd/cabinet` and had no temp-path guard.

## Implementation
- Updated `cypress.ps1` to:
  - prefer `bin/cabinet.exe` by default when present
  - support explicit `-RuntimeExecutablePath` override
  - reject temp/template executable paths unless `-AllowTempRuntimePath` is provided
  - log `Runtime executable resolved: ...` before launch for auditable checkpoint evidence
  - retain `go run ./cmd/cabinet` fallback only when project-local executable is absent

## Validation
- `go test ./tests -run TestCypressScript -count=1` ✅
- `go build -o bin/cabinet.exe ./cmd/cabinet` ✅
- `pwsh -NoLogo -NoProfile -File .\\cypress.ps1 -Spec cypress/e2e/integrations/ui-screen-market-watch/spec.cy.ts -Browser chrome -RequireE2EHooks` ✅
  - output includes `Runtime executable resolved: D:\projects\collectors-tech\cabinet\bin\cabinet.exe`
- `go test ./internal/app -count=1` ✅
- `go test ./tests -count=1` ✅
- `openspec validate --all` ✅

## Traceability
- Added `RUNTIME-CORE-006` mapping in `openspec/traceability.md` with executable test evidence.