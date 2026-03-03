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

### Requirement UI-SCREEN-INVENTORY-FOLDER-TREE-005: Tree pane SHALL be independently scrollable and MUST NOT expand whole page layout
Expanding tree nodes MUST not force global page growth; tree pane shall handle its own overflow.

#### Scenario: Vertical overflow in tree pane
- **GIVEN** tree has more nodes than visible pane height
- **WHEN** user expands branches
- **THEN** tree pane MUST provide internal vertical scrolling and page layout height MUST remain stable

### Requirement UI-SCREEN-INVENTORY-FOLDER-TREE-006: Tree pane SHALL support horizontal overflow for deep indentation
Deeply nested nodes MUST remain accessible via internal horizontal scrolling or equivalent overflow handling.

#### Scenario: Horizontal overflow in deep tree
- **GIVEN** nested hierarchy exceeds available tree pane width
- **WHEN** user navigates deep nodes
- **THEN** tree pane MUST allow horizontal access without clipping labels or expanding overall page width

### Requirement UI-SCREEN-INVENTORY-FOLDER-TREE-007: Tree nodes SHALL provide subtle add-child affordance
Each folder node SHALL provide a subtle `+` affordance to create a child folder directly under that node.

#### Scenario: Add child from node control
- **GIVEN** folder node is visible in tree
- **WHEN** user clicks node-level `+` affordance and submits valid name
- **THEN** new folder MUST be created as child of selected node and rendered in expanded tree context

### Requirement UI-SCREEN-INVENTORY-FOLDER-TREE-008: Tree SHALL provide explicit root node for top-level folder creation
Tree control SHALL expose a root context that allows creating top-level folders.

#### Scenario: Create top-level folder at root
- **GIVEN** user is in folder tree view
- **WHEN** user chooses `Add Root Folder` action
- **THEN** new folder MUST be created at root level and appear as top-level node

### Requirement UI-SCREEN-INVENTORY-FOLDER-TREE-009: Tree visuals SHALL include connector lines for hierarchy clarity
Tree view SHALL render visual connector lines/indent guides to make parent-child hierarchy clear.

#### Scenario: Render hierarchy lines
- **GIVEN** tree has multi-level nesting
- **WHEN** tree renders nodes
- **THEN** UI MUST show clear hierarchical line/guide cues between parent and child nodes

## Implementation recommendation
Preferred component strategy:
- `@react-aria/tree` + `@react-stately/tree` for accessible tree semantics
- optional virtualization for large trees (`@tanstack/react-virtual`)
- optional drag/reorder later via `dnd-kit` (out of scope for initial tree contract)
