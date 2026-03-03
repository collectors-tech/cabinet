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

### Requirement UI-FOUNDATION-SHELL-NAVIGATION-005: Local Workspace switcher SHALL remain DB/profile-only
Local Workspace shell switcher SHALL represent workspace database/profile context only and SHALL NOT render collections management widget above primary navigation.

#### Scenario: Sidebar top area renders DB/profile only
- **GIVEN** authenticated shell sidebar is rendered
- **WHEN** Local Workspace switcher is visible
- **THEN** top area MUST show DB/profile switcher context only
- **AND** collections list/add widget MUST NOT appear above primary nav items

### Requirement UI-FOUNDATION-SHELL-NAVIGATION-006: Collections management SHALL live in Collections section and inline pickers
Collections creation/list management SHALL be provided via dedicated Collections section and inline picker quick-create flows, not via sidebar-top widget.

#### Scenario: Collections management placement
- **GIVEN** user needs to manage collections
- **WHEN** user uses navigation or picker flows
- **THEN** collections list/create MUST be available in Collections section and relevant pickers
- **AND** sidebar top area remains uncluttered

### Requirement UI-FOUNDATION-SHELL-NAVIGATION-007: Navigation edit dialog SHALL reflect live item order during reordering
When user moves menu items up/down in nav edit mode, edit dialog list order MUST update immediately to match resulting navigation order.

#### Scenario: Move menu item and verify edit list order
- **GIVEN** navigation edit dialog is open with reorder controls
- **WHEN** user moves an item up or down
- **THEN** edit dialog list MUST re-render in new order immediately
- **AND** resulting saved order MUST match what was shown in edit dialog before save
