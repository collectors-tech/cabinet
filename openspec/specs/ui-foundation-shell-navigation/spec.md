## Purpose
Define global shell, navigation, context pane, and layout ownership behavior.

## Requirements
### Requirement UI-FOUNDATION-SHELL-NAVIGATION-001: App shell SHALL define fixed navigation and scroll ownership
Cabinet SHALL keep primary navigation fixed and assign page-body scroll to content container.

#### Scenario: Page content scroll
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** content exceeds viewport height
- **THEN** only content column SHALL scroll while primary nav remains fixed

### Requirement UI-FOUNDATION-SHELL-NAVIGATION-002: Primary navigation SHALL support collapse and edit behavior
Cabinet SHALL support nav collapse and configurable ordering and visibility controls.

#### Scenario: Edit navigation order
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** user enters nav edit mode and reorders entries
- **THEN** nav order SHALL persist for active profile context

### Requirement UI-FOUNDATION-SHELL-NAVIGATION-003: Context navigation SHALL support collection-focused workspace context
Cabinet SHALL provide context pane behavior for collection and folder navigation.

#### Scenario: Context pane selection
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** user changes active collection context
- **THEN** page content SHALL reflect selected context state

### Requirement UI-FOUNDATION-SHELL-NAVIGATION-004: Version/build metadata SHALL be visible in shell footer
Cabinet SHALL display app version and build date metadata in sidebar footer region.

#### Scenario: Render shell footer metadata
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** shell renders
- **THEN** version/build metadata SHALL be visible in footer
