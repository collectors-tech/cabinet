# Auto Wave 20 Summary

- Issue: #151
- Scope: Windows Cypress startup remediation for npm-run scripts without WMIC dependency.
- Date: 2026-03-02

## Changes delivered
- Updated `ui.web/package.json` scripts to run Cypress through managed PowerShell runner `../cypress.ps1`.
- Removed dependency on `start-server-and-test` process-tree behavior that triggers `spawn wmic.exe ENOENT` on this host.
- Verified both acceptance scripts execute successfully end-to-end.

## Commands run
1. `npm run e2e:run-smoke` (workdir: `ui.web`)
2. `npm run e2e:run-inventory` (workdir: `ui.web`)
3. `go test ./internal/app -count=1`
4. `go test ./tests -count=1`
5. `openspec validate --all`

## Results
- `npm run e2e:run-smoke`: **pass** (`3 passing, 0 failing`)
- `npm run e2e:run-inventory`: **pass** (`5 passing, 0 failing`)
- `go test ./internal/app -count=1`: **pass**
- `go test ./tests -count=1`: **pass**
- `openspec validate --all`: **pass**

## Blocker resolved
- Previous first actionable error: `Error: spawn wmic.exe ENOENT`
- Current status: removed from npm E2E script execution path.
