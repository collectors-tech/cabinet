# Integrations Spec-Implementation Alignment Summary

Date: 2026-03-01 (Australia/Sydney)

## Scope
Three-phase alignment for integrations:
1. Spec alignment
2. Implementation alignment
3. Tests + traceability

## Phase 1: Spec alignment
Updated:
- `openspec/specs/integrations/spec.md`
- `openspec/specs/integrations/provider-registry/spec.md`
- `openspec/specs/integrations/ui-screen-integrations/spec.md`

Key outcomes:
- Canonical provider-focused model clarified.
- Card-first contract explicitly represented.
- Added/expanded requirements for:
  - `/api/providers/registry` as list source-of-truth
  - provider `health` + `last_run` display contract
  - deterministic bootstrap load/error states
  - write-only credential + replace-token flow
- New IDs added append-only: `INTEGRATION-022`, `INTEGRATION-023`, `UI-SCREEN-INTEGRATIONS-006`, `UI-SCREEN-INTEGRATIONS-007`.

## Phase 2: Implementation alignment
Updated:
- `ui.web/src/features/apps/index.tsx`
- `internal/app/app.go`

Key outcomes:
- Replaced static integration card seed dependency with runtime registry fetch (`GET /api/providers/registry`).
- Connect state now runtime-derived per provider (`has_token`, auth mode, settings flags), not hardcoded by provider list.
- Provider detail panel now includes:
  - setup instructions
  - health status + last-run status
  - actions: Validate / Sync / Save
- Credential UX hardened:
  - token is write-only in UI
  - no clear token rehydration into inputs
  - explicit replace-token flow
- Added deterministic bootstrap error/retry state for active-profile/registry/settings load failures.
- Added URL-backed cards/rows view state (`view=rows|cards`, default cards).
- Backend provider registry now returns:
  - `has_token`
  - `setup_instructions`
  - `health` object
  - `last_run` object

## Phase 3: Tests + traceability
Updated tests:
- `ui.web/cypress/e2e/integrations/ui-screen-integrations/spec.cy.ts` (new)
- `internal/app/traceability_wave4_provider_scanner_test.go` (registry contract fields)

Traceability updates:
- `openspec/traceability.md` updated for `INTEGRATION-020/021/022/023` and `UI-SCREEN-INTEGRATIONS-001..007`.
- Implemented status applied only where executable proof exists.
- UI/E2E IDs remain partial with explicit blocker due Cypress startup failure.

## Commands run
1. `go test ./internal/app -count=1`
2. `go test ./tests -count=1`
3. `npm run build` (cwd: `ui.web`)
4. `go test ./internal/app -run "TestWave4ProvidersRegistryContract" -count=1`
5. `npx cypress run --browser chrome --config-file .\cypress.config.ts --spec .\cypress\e2e\specs\integrations\ui-screen-integrations\spec.cy.ts` (cwd: `ui.web`)
6. `npx cypress cache clear`
7. `npx cypress install`
8. `npx cypress verify`
9. `openspec validate --all`

## Results
- `go test ./tests -count=1`: PASS
- `go test ./internal/app -count=1`: FAIL (pre-existing migration-audit path assertion unrelated to integrations scope)
- `go test ./internal/app -run "TestWave4ProvidersRegistryContract" -count=1`: PASS
- `npm run build` (ui.web): PASS
- Cypress run/verify: BLOCKED by host binary startup error
- `openspec validate --all`: PASS

## Blockers
1. Command: `npx cypress run --browser chrome --config-file .\cypress.config.ts --spec .\cypress\e2e\specs\integrations\ui-screen-integrations\spec.cy.ts`
   - First actionable error: `Cypress.exe: bad option: --smoke-test`
   - Required fix: repair host Cypress executable/runtime mismatch in `C:\Users\maxbarrass\AppData\Local\Cypress\Cache\13.15.2\Cypress\Cypress.exe` or pin a known-good Cypress binary/channel for this machine.
2. Command: `go test ./internal/app -count=1`
   - First actionable error: `migration audit target path does not exist: openspec/specs/wishlist/wishlist-pricing-dashboard/spec.md`
   - Required fix: update `openspec/migrations/legacy-docs-file-audit.yaml` target path mapping to current spec structure.

## Commit
- commitHash: pending
