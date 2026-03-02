# Inventory Folder Tree Control Summary

- Issue: #217
- Spec: `openspec/specs/inventory/folder-tree-control/spec.md`
- Requirement IDs:
  - `UI-SCREEN-INVENTORY-FOLDER-TREE-001`
  - `UI-SCREEN-INVENTORY-FOLDER-TREE-002`
  - `UI-SCREEN-INVENTORY-FOLDER-TREE-003`
  - `UI-SCREEN-INVENTORY-FOLDER-TREE-004`

## Implementation
- Replaced flat folder button list with a hierarchical tree control in `ui.web/src/features/collection/index.tsx`.
- Added deterministic expand/collapse node state with explicit toggles for parent nodes.
- Added keyboard interactions for treeitems: `ArrowRight`, `ArrowLeft`, `ArrowDown`, `ArrowUp`, `Enter`, `Space`.
- Added ARIA semantics: `role=tree`, `role=treeitem`, `role=group`, `aria-level`, `aria-selected`, `aria-expanded`.
- Preserved context binding so folder selection updates `collection-context-label` and active context summary.
- Expanded tree dataset to provide scalable baseline (>20 visible nodes + nested warehouse hierarchy).

## E2E Proof
- Added `ui.web/cypress/e2e/inventory/folder-tree-control/spec.cy.ts` covering all four requirement IDs.
- Test results: 4 passing, 0 failing.

## Mandatory Gates
1. `pwsh -NoLogo -NoProfile -File .\\cypress.ps1 -Spec cypress/e2e/inventory/folder-tree-control/spec.cy.ts -Browser chrome -RequireE2EHooks` ✅
2. `go test ./internal/app -count=1` ✅
3. `go test ./tests -count=1` ✅
4. `openspec validate --all` ✅

## Traceability
- Updated `openspec/traceability.md`:
  - `UI-SCREEN-INVENTORY-FOLDER-TREE-001` -> implemented
  - `UI-SCREEN-INVENTORY-FOLDER-TREE-002` -> implemented
  - `UI-SCREEN-INVENTORY-FOLDER-TREE-003` -> implemented
  - `UI-SCREEN-INVENTORY-FOLDER-TREE-004` -> implemented

## Commit
- Commit: pending
- Push-proof: pending
