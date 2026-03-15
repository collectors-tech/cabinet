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

### Requirement UI-SCREEN-COLLECTIONS-004: Collections screen SHALL show and confirm active collection context changes
The dedicated Collections screen SHALL let users understand which collection is currently active and confirm the effect of switching to another collection.

#### Scenario: Switch active collection from Collections screen
- **GIVEN** collections section is open and at least two collections exist
- **WHEN** user selects a different collection entry
- **THEN** the newly active collection MUST be visually distinguished on the page
- **AND** the active context change MUST be communicated with explicit on-screen state and persistence semantics

### Requirement UI-SCREEN-COLLECTIONS-005: Collections screen SHALL support rename and remove management actions
The dedicated Collections screen SHALL provide direct management actions for existing collections beyond creation and activation.

#### Scenario: Rename or remove an existing collection
- **GIVEN** collections section is open and a non-default collection exists
- **WHEN** user chooses rename or remove for that collection
- **THEN** the screen MUST expose the corresponding action
- **AND** the action MUST complete through a clear confirmation/edit flow with deterministic result messaging

### Requirement UI-SCREEN-COLLECTIONS-006: Collections screen SHALL expose collection details and metadata summaries
Users SHALL be able to understand what a collection represents before choosing it.

#### Scenario: Review collection details from list
- **GIVEN** collections section is open
- **WHEN** user scans or opens a collection entry
- **THEN** the UI MUST expose useful metadata for that collection such as summary/details/counts/status as defined by the product contract

### Requirement UI-SCREEN-COLLECTIONS-007: Collections screen SHALL support search, filtering, and ordering tools for collection management
When multiple collections exist, the dedicated management surface SHALL help users find and organize them efficiently.

#### Scenario: Locate and organize a collection
- **GIVEN** collections section contains many entries
- **WHEN** user needs to find or organize a specific collection
- **THEN** the screen MUST provide supported search/filtering and ordering controls for collection management workflows

### Requirement UI-SCREEN-COLLECTIONS-008: Collections screen SHALL communicate create-action outcomes and available create paths clearly
Collection creation entry points SHALL make their behavior obvious and communicate whether creation succeeded, failed, or requires additional input.

#### Scenario: Use visible create actions on Collections screen
- **GIVEN** collections section is open
- **WHEN** user uses `New` or `Create`
- **THEN** the resulting workflow/options MUST be obvious from the page state
- **AND** success, validation, and failure outcomes MUST be communicated with visible feedback

#### Scenario: Blank collection-name submission shows inline validation
- **GIVEN** collections section inline create panel is open
- **WHEN** user clicks `Save` with an empty `Collection name`
- **THEN** the inline create panel MUST remain open
- **AND** the screen MUST show visible required-field guidance
- **AND** only `Cancel` MAY silently dismiss the inline create panel without validation feedback
