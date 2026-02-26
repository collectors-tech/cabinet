## Purpose
Define global shell, navigation, context pane, and layout ownership behavior.

## Requirements
### Requirement: App shell SHALL define fixed navigation and scroll ownership
Cabinet SHALL keep primary navigation fixed and assign page-body scroll to content container.

#### Scenario: Page content scroll
- **WHEN** content exceeds viewport height
- **THEN** only content column SHALL scroll while primary nav remains fixed

### Requirement: Primary navigation SHALL support collapse and edit behavior
Cabinet SHALL support nav collapse and configurable ordering and visibility controls.

#### Scenario: Edit navigation order
- **WHEN** user enters nav edit mode and reorders entries
- **THEN** nav order SHALL persist for active profile context

### Requirement: Context navigation SHALL support collection-focused workspace context
Cabinet SHALL provide context pane behavior for collection and folder navigation.

#### Scenario: Context pane selection
- **WHEN** user changes active collection context
- **THEN** page content SHALL reflect selected context state

### Requirement: Version/build metadata SHALL be visible in shell footer
Cabinet SHALL display app version and build date metadata in sidebar footer region.

#### Scenario: Render shell footer metadata
- **WHEN** shell renders
- **THEN** version/build metadata SHALL be visible in footer
