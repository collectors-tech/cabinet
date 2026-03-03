# Auto Wave 70 Summary

- Issue: #276
- Scope: enforce UI build-before-Go for local + installer packaging entrypoints
- Status: done

## What changed
- Removed `-SkipUIBuild` bypass from `scripts/build-cabinet.ps1`.
- Kept unconditional `build-ui-static.ps1` execution before local `go build`.
- Added script contract tests in `tests/build_pipeline_contract_test.go`:
  - local build script cannot expose UI skip path
  - installer packaging script must run UI build before cross-platform loop and fail fast
  - README canonical entrypoint must not document skip bypass
- Updated README build notes to reflect enforced canonical flow.

## Commands run
1. `go test ./tests -run Build -count=1`
2. `pwsh -NoLogo -NoProfile -File .\\cypress.ps1 -Spec cypress/e2e/general/api-docs/spec.cy.ts -Browser chrome`
3. `go test ./internal/app -count=1`
4. `go test ./tests -count=1`
5. `openspec validate --all`

## Results
- Build-focused tests: pass.
- Cypress API docs suite: 2 passing, 0 failing.
- `go test ./internal/app`: pass.
- `go test ./tests`: pass.
- `openspec validate --all`: pass.

## Blockers
- none
