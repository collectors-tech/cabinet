## Purpose
Define folder browser as a real hierarchical tree control for inventory navigation.

## Requirements
### Requirement UI-SCREEN-INVENTORY-FOLDER-TREE-001: Inventory folder browser SHALL be a true tree control
Inventory folder navigation MUST use a hierarchical tree control (expand/collapse, parent-child nesting), not a flat list imitation.

#### Scenario: Expand/collapse hierarchy
- **GIVEN** inventory has nested folder structure
- **WHEN** user expands or collapses nodes
- **THEN** tree MUST render deterministic parent/child relationships and preserve current selection context

### Requirement UI-SCREEN-INVENTORY-FOLDER-TREE-002: Tree control SHALL support keyboard and accessibility semantics
Tree MUST expose standard treeview semantics and keyboard interactions.

#### Scenario: Keyboard tree navigation
- **GIVEN** folder tree has focus
- **WHEN** user uses arrow keys and Enter/Space
- **THEN** tree MUST support navigate/expand/collapse/select with accessible roles (`tree`, `treeitem`, `group`) and focus visibility

### Requirement UI-SCREEN-INVENTORY-FOLDER-TREE-003: Tree selection SHALL drive inventory filtering deterministically
Selecting a folder node MUST update inventory content view to that folder scope.

#### Scenario: Select folder node
- **GIVEN** user selects a folder node in tree
- **WHEN** selection is applied
- **THEN** inventory list/cards MUST update to selected folder scope and breadcrumb/context labels MUST match selected node

### Requirement UI-SCREEN-INVENTORY-FOLDER-TREE-004: Tree control SHALL support scalable rendering for deep/large hierarchies
Tree rendering MUST remain responsive for large folder sets.

#### Scenario: Large tree performance baseline
- **GIVEN** folder tree contains deep hierarchy and high node count
- **WHEN** user expands/collapses and selects nodes
- **THEN** interactions MUST remain responsive and avoid blocking UI thread

## Implementation recommendation
Preferred component strategy:
- `@react-aria/tree` + `@react-stately/tree` for accessible tree semantics
- optional virtualization for large trees (`@tanstack/react-virtual`)
- optional drag/reorder later via `dnd-kit` (out of scope for initial tree contract)
