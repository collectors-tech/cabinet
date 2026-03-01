# Auto Wave 9 Summary

## Scope
- Issue: #189
- Requirement IDs: `UI-LOGIN-SESSION-003`
- Spec binding: `openspec/specs/general/ui-login-session/spec.md`
- E2E binding: `ui.web/cypress/e2e/general/ui-login-session/spec.cy.ts`

## Commands Run
1. `./cypress.ps1 -Spec "cypress/e2e/general/ui-login-session/spec.cy.ts" -Browser chrome -RequireE2EHooks` (failed: E2E hook unavailable on active runtime)
2. `npx cypress cache clear`
3. `npx cypress install`
4. `npx cypress verify` (failed initially due `ELECTRON_RUN_AS_NODE=1`)
5. `Remove-Item Env:ELECTRON_RUN_AS_NODE -ErrorAction SilentlyContinue; npx cypress run --browser chrome --config-file .\\cypress.config.ts --spec .\\cypress\\e2e\\general\\ui-login-session\\spec.cy.ts`
6. `go test ./internal/app -count=1`
7. `go test ./tests -count=1`
8. `openspec validate --all`

## Key Results
- Cypress spec passed: 3 passing, 0 failing.
- Implemented active-profile switch E2E path with scoped API follow-up validation.
- `go test ./internal/app -count=1` passed after i18n shell-token contract fix.
- `go test ./tests -count=1` passed.
- `openspec validate --all` passed.

## Requirement Status Changes
- `UI-LOGIN-SESSION-003`: `partial` -> `implemented`.

## Notes
- Host environment sets `ELECTRON_RUN_AS_NODE=1`; Cypress must run with that env var cleared.
- Existing runtime on port `17880` may not expose `/api/test/reset`; this wave used direct UI E2E proof against running runtime.
