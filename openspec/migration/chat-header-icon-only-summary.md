# Chat Header Icon-Only Summary

- Issue: #219
- Spec: `openspec/specs/chats/ui-screen-chat-copilot/spec.md`
- Requirement ID: `UI-SCREEN-CHAT-COPILOT-006`

## Implementation
- Updated shell header chat trigger to icon-only control in `ui.web/src/components/layout/header.tsx`.
- Removed visible inline text label from trigger.
- Preserved accessibility via `aria-label` and `title` on the icon trigger.
- Preserved interaction behavior: trigger still opens/closes chat rail.

## E2E Proof
- Added dedicated requirement test:
  - `ui.web/cypress/e2e/chats/ui-screen-chat-copilot/header-trigger-icon-only.cy.ts`
- Scenario validates:
  - icon-only trigger rendering
  - accessible name attributes
  - chat rail open/close behavior

## Mandatory Gates
1. `pwsh -NoLogo -NoProfile -File .\\cypress.ps1 -Spec cypress/e2e/chats/ui-screen-chat-copilot/header-trigger-icon-only.cy.ts -Browser chrome -RequireE2EHooks` ✅
2. `go test ./internal/app -count=1` ✅
3. `go test ./tests -count=1` ✅
4. `openspec validate --all` ✅

## Traceability
- Updated `openspec/traceability.md`:
  - `UI-SCREEN-CHAT-COPILOT-006` -> implemented

## Commit
- Commit: pending
- Push-proof: pending
