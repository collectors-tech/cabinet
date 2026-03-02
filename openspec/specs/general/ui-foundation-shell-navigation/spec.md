## Purpose
Define global shell, navigation, context pane, and layout ownership behavior.

## Requirements
### Requirement UI-FOUNDATION-SHELL-NAVIGATION-001: App shell SHALL define fixed navigation and scroll ownership
Cabinet SHALL keep primary navigation fixed and assign page-body scroll to content container.

#### Scenario: Page content scroll
- **GIVEN** desktop viewport is active and page body content exceeds available vertical height
- **WHEN** content exceeds viewport height
- **THEN** only content column SHALL scroll while primary nav remains fixed and header actions remain visible at top of content pane

### Requirement UI-FOUNDATION-SHELL-NAVIGATION-002: Primary navigation SHALL support collapse and edit behavior
Cabinet SHALL support nav collapse and configurable ordering and visibility controls.

#### Scenario: Edit navigation order
- **GIVEN** user enters navigation edit mode from shell footer controls with profile-scoped nav preferences available
- **WHEN** user enters nav edit mode and reorders entries
- **THEN** nav order SHALL persist for active profile context and reload with same order and visibility toggles

### Requirement UI-FOUNDATION-SHELL-NAVIGATION-003: Context navigation SHALL support collection-focused workspace context
Cabinet SHALL provide context pane behavior for collection and folder navigation.

#### Scenario: Context pane selection
- **GIVEN** collection context pane is visible and at least two contexts/folders are available
- **WHEN** user changes active collection context
- **THEN** page content SHALL reflect selected context state and context label in header/body SHALL match selected entry

### Requirement UI-FOUNDATION-SHELL-NAVIGATION-004: Version/build metadata SHALL be visible in shell footer
Cabinet SHALL display app version and build date metadata in sidebar footer region.

#### Scenario: Render shell footer metadata
- **GIVEN** runtime metadata endpoint returns `app_version` and `build_date`
- **WHEN** shell renders
- **THEN** version/build metadata SHALL be visible in footer

### Requirement UI-FOUNDATION-SHELL-NAVIGATION-005: Local Workspace area SHALL be collection-centric
Local Workspace entry/region SHALL represent Collections domain explicitly.

#### Scenario: Local Workspace shows collections list
- **GIVEN** user opens Local Workspace area from shell/navigation context
- **WHEN** workspace panel renders
- **THEN** UI MUST show `Collections` heading and list available collections

### Requirement UI-FOUNDATION-SHELL-NAVIGATION-006: Local Workspace SHALL expose Add Collection action
Local Workspace SHALL allow creating a new collection directly from the collections list area.

#### Scenario: Add new collection from Local Workspace
- **GIVEN** user is viewing Local Workspace collections list
- **WHEN** user clicks `Add Collection` and submits valid name/details
- **THEN** new collection MUST be created and appear in collection list without full reload

### Requirement UI-FOUNDATION-SHELL-NAVIGATION-007: Navigation edit dialog SHALL reflect live item order during reordering
When user moves menu items up/down in nav edit mode, edit dialog list order MUST update immediately to match resulting navigation order.

#### Scenario: Move menu item and verify edit list order
- **GIVEN** navigation edit dialog is open with reorder controls
- **WHEN** user moves an item up or down
- **THEN** edit dialog list MUST re-render in new order immediately
- **AND** resulting saved order MUST match what was shown in edit dialog before save
