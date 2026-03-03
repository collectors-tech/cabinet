## Purpose
Define top-level Collections screen behavior and create workflows.

## Requirements
### Requirement UI-SCREEN-COLLECTIONS-001: Collections SHALL be a top-level managed section with dedicated create flow
Collections management SHALL exist as a first-class section and support list + create operations.

#### Scenario: Open collections section
- **GIVEN** authenticated user navigates to Collections section
- **WHEN** section renders
- **THEN** UI MUST show collection list and dedicated `New` create action for collection entries

### Requirement UI-SCREEN-COLLECTIONS-002: Collections screen SHALL expose New and Create-menu actions
Collections screen SHALL provide a dedicated `New` button for primary entity creation and adjacent `Create` menu for quick create actions.

#### Scenario: New + Create menu behavior
- **GIVEN** collections section is open
- **WHEN** user clicks `New`
- **THEN** primary create-collection flow MUST open
- **WHEN** user clicks adjacent `Create` menu
- **THEN** menu MUST show configured quick-create actions

### Requirement UI-SCREEN-COLLECTIONS-003: Collection picker SHALL support inline quick-create
Where collection picker is used (inventory/wishlist detail forms), picker SHALL allow inline create without leaving current workflow.

#### Scenario: Quick-create from collection picker
- **GIVEN** user opens collection picker in item/wishlist details
- **WHEN** user selects `+ New Collection` and submits valid name
- **THEN** new collection MUST be created and auto-selected in the current picker context
