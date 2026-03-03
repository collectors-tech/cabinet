# Auto Wave 75 Summary

## Issue
- #231 `[Spec Backlog] Folder tree UX: node + add-child, root folder creation, and hierarchy connector lines`

## Implemented IDs
- UI-SCREEN-INVENTORY-FOLDER-TREE-007
- UI-SCREEN-INVENTORY-FOLDER-TREE-008
- UI-SCREEN-INVENTORY-FOLDER-TREE-009

## What Changed
- Added node-level add-child affordance on each folder node.
- Added root-level folder creation action in folder pane.
- Added hierarchy connector line visuals for nested nodes.
- Added/updated Cypress scenarios proving 007/008/009.
- Updated traceability statuses to implemented with executable proof.

## Validation Evidence
- `pwsh -NoLogo -NoProfile -File .\\cypress.ps1 -Spec cypress/e2e/inventory/folder-tree-control/spec.cy.ts -Browser chrome -RequireE2EHooks` => pass (9/9)
- `go test ./internal/app -count=1` => pass
- `go test ./tests -count=1` => pass
- `openspec validate --all` => pass

## Blockers
- None
