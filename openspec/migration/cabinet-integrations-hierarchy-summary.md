# Cabinet Integrations Hierarchy + Contract Alignment Summary

Date: 2026-03-01

## Scope Executed
- `openspec/specs/integrations/**`
- `ui.web/cypress/e2e/**` (integrations-related)
- `openspec/traceability.md`
- integrations UI/backend validation through executable tests

## Hierarchy Alignment
- Verified 1:1 placement for integrations screen spec and E2E test path:
  - Spec: `openspec/specs/integrations/ui-screen-integrations/spec.md`
  - Test: `ui.web/cypress/e2e/integrations/ui-screen-integrations/spec.cy.ts`
- Verified no stale flat integrations test remains in `ui.web/cypress/e2e` root.

## Contract Alignment and Remediation
- Kept integrations runtime provider-focused and registry-backed (`/api/providers/registry`) with card-first UX.
- Fixed deterministic Cypress runtime startup on this host by hardening managed runner script:
  - `cypress.ps1` now removes `ELECTRON_RUN_AS_NODE` for Cypress child process.
- Stabilized integrations journey assertions:
  - sign-in redirect lands directly on `/integrations/` route.
  - route assertion accepts normalized trailing-slash variants.

## Commands Run
1. `pwsh ./cypress.ps1 -Browser chrome -Spec cypress/e2e/integrations/ui-screen-integrations/spec.cy.ts`
2. `go test ./internal/app -run "Wave4|Scanner|Provider|provider" -count=1`
3. `go test ./tests -run "Provider|provider|Amazon|Ebay|shop" -count=1`
4. `go test ./internal/app -count=1`
5. `go test ./tests -count=1`
6. `openspec validate --all`
7. `pwsh ./cypress.ps1 -Browser chrome -Spec cypress/e2e/integrations/ui-screen-integrations/spec.cy.ts` (post-fix)
8. `pwsh ./cypress.ps1 -Browser chrome -Spec cypress/e2e/integrations/ui-screen-integrations/spec.cy.ts` (final proof)
9. `go test ./internal/app -count=1` (final)
10. `go test ./tests -count=1` (final)
11. `openspec validate --all` (final)

## Results
- Integrations Cypress (managed lifecycle): **PASS**
  - `ui.web/cypress/e2e/integrations/ui-screen-integrations/spec.cy.ts`
  - 4 passing, 0 failing
- `go test ./internal/app -count=1`: **PASS**
- `go test ./tests -count=1`: **PASS**
- `openspec validate --all`: **PASS**

## Traceability Status Updates (with executable E2E proof)
Moved to `implemented`:
- `INTEGRATION-020`
- `INTEGRATION-022`
- `UI-SCREEN-INTEGRATIONS-001`
- `UI-SCREEN-INTEGRATIONS-002`
- `UI-SCREEN-INTEGRATIONS-003`
- `UI-SCREEN-INTEGRATIONS-004`
- `UI-SCREEN-INTEGRATIONS-005`
- `UI-SCREEN-INTEGRATIONS-006`
- `UI-SCREEN-INTEGRATIONS-007`

Evidence path for all above:
- `ui.web/cypress/e2e/integrations/ui-screen-integrations/spec.cy.ts`

## Blockers
- None remaining for this task.
