# ChatGPT Copilot Inventory/Wishlist Summary

- Issue: #233
- Requirement IDs:
  - UI-SCREEN-CHAT-COPILOT-007
  - UI-SCREEN-CHAT-COPILOT-008
  - PROVIDER-OPENAI-004

## Implemented
- Added chat action mode contract in Chats UI for:
  - `create_inventory_item`
  - `create_wishlist_entry`
  - `update_inventory_item`
- Added explicit confirm-before-apply dialog and summary before mutation.
- Added mobile image-attachment flow coverage in chat copilot E2E.
- Added backend chat action support for inventory create/update and wishlist create.
- Added OpenAI adapter structured operation proposal method and tests.
- Migrated chat copilot Cypress path to hierarchy-aligned location:
  - `ui.web/cypress/e2e/chats/ui-screen-chat-copilot/spec.cy.ts`

## Evidence
- Cypress:
  - `pwsh -File .\\cypress.ps1 -Spec "cypress/e2e/chats/ui-screen-chat-copilot/spec.cy.ts" -Browser chrome` (3 passing)
- Go tests:
  - `go test ./internal/ai -count=1` (pass)
  - `go test ./internal/chat -count=1` (pass)
  - `go test ./internal/app -count=1` (pass)
  - `go test ./tests -count=1` (pass)
- Spec validation:
  - `openspec validate --all` (pass)
- Build:
  - `pwsh -File .\\scripts\\build-cabinet.ps1` (pass)

## Traceability
- `UI-SCREEN-CHAT-COPILOT-007` -> implemented
- `UI-SCREEN-CHAT-COPILOT-008` -> implemented
- `PROVIDER-OPENAI-004` -> implemented
- `CHAT-COPILOT-001` path updated to hierarchy-aligned E2E file
