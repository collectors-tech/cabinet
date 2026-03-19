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

### Requirement UI-SCREEN-INVENTORY-FOLDER-TREE-010: Tree rows SHALL provide clear disclosure, iconography, and hierarchy cues
The tree MUST make hierarchy and current state obvious at a glance with dedicated disclosure affordances and node/leaf icon treatment beyond bare text rows, while keeping selection visually clear without flattening the surrounding hierarchy.

#### Scenario: Scan and understand tree structure quickly
- **GIVEN** user opens the inventory tree with mixed parent and leaf nodes
- **WHEN** they scan the visible hierarchy without interacting deeply
- **THEN** parent rows MUST expose a clear disclosure affordance distinct from selection
- **AND** disclosure activation MUST expand/collapse without implicitly changing the selected folder context
- **AND** rows MUST show node/leaf visual treatment that communicates expandable vs terminal state
- **AND** mixed parent and leaf rows MUST preserve consistent alignment cues that improve quick scanning
- **AND** the active selection MUST be visually distinct without overwhelming the rest of the tree

### Requirement UI-SCREEN-INVENTORY-FOLDER-TREE-011: Tree SHALL support contextual row actions without crowding the primary selection flow
Tree rows MUST support secondary actions in a way that does not force the user to choose between selection and management.

#### Scenario: Invoke row-level management actions
- **GIVEN** user is browsing folders in the tree
- **WHEN** they need to manage a specific node
- **THEN** the row MUST expose contextual actions such as add child and future management affordances through inline or overflow actions
- **AND** invoking those actions MUST NOT accidentally change the primary selected context unless the action explicitly requires it

### Requirement UI-SCREEN-INVENTORY-FOLDER-TREE-012: Tree SHALL preserve and restore navigation context predictably
Tree exploration state MUST survive normal in-app navigation and keep users oriented when returning to the tree.

#### Scenario: Return to previously explored branch
- **GIVEN** user has expanded branches and selected a deep node
- **WHEN** they navigate away and return to the inventory workspace or refresh within supported persistence scope
- **THEN** the tree MUST restore the selected node and expanded ancestor path according to product persistence rules
- **AND** if the active node is not visible, the tree MUST reveal it by expanding required ancestors

### Requirement UI-SCREEN-INVENTORY-FOLDER-TREE-013: Tree rows SHALL support richer metadata rendering for inventory context
The tree MUST support custom row rendering for inventory context such as counts, status badges, or secondary labels when available.

#### Scenario: Review folder context from the tree alone
- **GIVEN** inventory folders have useful secondary metadata such as item counts or status indicators
- **WHEN** the tree renders rows
- **THEN** rows MUST be able to display that metadata in a structured way without breaking hierarchy readability or keyboard accessibility

### Requirement UI-SCREEN-INVENTORY-FOLDER-TREE-014: Tree SHALL support hierarchy re-organization through drag-drop or equivalent move workflow
Users MUST be able to reorganize the folder hierarchy without rebuilding it manually from scratch.

#### Scenario: Move folder within the hierarchy
- **GIVEN** user wants to re-parent or reorder a folder within the tree
- **WHEN** they perform the supported move interaction
- **THEN** the tree MUST provide a deterministic move workflow with clear feedback, valid drop/move constraints, and correct hierarchy updates after completion

## Implementation recommendation
Preferred component strategy:
- `@react-aria/tree` + `@react-stately/tree` for accessible tree semantics
- optional virtualization for large trees (`@tanstack/react-virtual`)
- drag/reorder via `dnd-kit` or equivalent once move semantics are specified
- use a row-renderer model that supports node icons, badges/counts, and contextual actions without breaking tree semantics
