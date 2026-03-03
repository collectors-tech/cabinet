# Inventory Tree Child/Root/Connector Alignment Summary (#231)

## Scope
- Issue: #231
- Spec IDs: UI-SCREEN-INVENTORY-FOLDER-TREE-007, UI-SCREEN-INVENTORY-FOLDER-TREE-008, UI-SCREEN-INVENTORY-FOLDER-TREE-009
- Spec path: `openspec/specs/inventory/folder-tree-control/spec.md`

## Changes
- Added node-level `+` child folder action on every tree row.
- Added explicit `Add Root Folder` action in folder pane.
- Implemented folder-create dialog with deterministic IDs and immediate tree insertion.
- Added connector line/indent guide visuals on tree nodes.
- Updated traceability for 007/008/009 from `partial` to `implemented`.

## Commands Run
- `pwsh -NoLogo -NoProfile -File .\\scripts\\build-ui-static.ps1`
- `pwsh -NoLogo -NoProfile -File .\\cypress.ps1 -Spec cypress/e2e/inventory/folder-tree-control/spec.cy.ts -Browser chrome -RequireE2EHooks`
- `go test ./internal/app -count=1`
- `go test ./tests -count=1`
- `openspec validate --all`

## Results
- Cypress target suite: **9 passing / 0 failing**
- `go test ./internal/app`: **pass**
- `go test ./tests`: **pass**
- `openspec validate --all`: **pass**

## Outcome
- Folder tree now supports quick child/root creation and hierarchy guides while preserving existing expand/collapse/select behavior.
