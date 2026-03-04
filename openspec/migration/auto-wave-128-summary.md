# Auto Wave 128 Summary

- Issue: #309
- Requirement IDs: POKEMON-COMP-006
- Scope: Pokemon dynamic list template catalog + apply contract

## Commands Run
1. `pwsh -NoLogo -NoProfile -File .\\cypress.ps1 -Spec cypress/e2e/integrations/pokemon-competitive-gap-parity/dynamic-list-templates.cy.ts -Browser chrome` (fail-first: missing spec)
2. `pwsh -NoLogo -NoProfile -File .\\cypress.ps1 -Spec cypress/e2e/integrations/pokemon-competitive-gap-parity/dynamic-list-templates.cy.ts -Browser chrome` (fail with 404 before latest-runtime override)
3. `go test ./internal/app -run TestPokemonDynamicListTemplates -count=1` (pass)
4. `pwsh -NoLogo -NoProfile -File .\\cypress.ps1 -Spec cypress/e2e/integrations/pokemon-competitive-gap-parity/dynamic-list-templates.cy.ts -Browser chrome -RuntimeExecutablePath .tmp/cabinet-latest.exe -AllowTempRuntimePath` (pass)
5. `go test ./internal/app -count=1` (pass)
6. `go test ./tests -count=1` (pass)
7. `openspec validate --all` (pass)
8. `pwsh -NoLogo -NoProfile -File .\\scripts\\build-cabinet.ps1` (pass)

## Key Results
- Expanded `POKEMON-COMP-006` to executable API contracts in OpenSpec.
- Added runtime endpoints:
  - `GET /api/integrations/pokemon/list-templates`
  - `POST /api/integrations/pokemon/list-templates/apply`
- Implemented deterministic invalid-template rejection with `400 {"error":"invalid_template_id"}`.
- Added API tests:
  - `internal/app/pokemon_dynamic_list_templates_api_test.go`
- Added Cypress proof at mapped path:
  - `ui.web/cypress/e2e/integrations/pokemon-competitive-gap-parity/dynamic-list-templates.cy.ts`
- Updated traceability:
  - `POKEMON-COMP-006` => implemented.

## Blockers
- None.
