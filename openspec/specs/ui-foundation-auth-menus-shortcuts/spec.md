## Purpose
Define account menu content, identity display, and shortcut notation behavior.

## Requirements
### Requirement: Header account menu SHALL be identity-backed and platform-aware
Cabinet SHALL display current user identity and keyboard shortcuts with OS-specific labels.

#### Scenario: Header menu shortcut labels
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** platform is macOS or Windows
- **THEN** displayed shortcut notation SHALL match platform conventions

### Requirement: Header account menu SHALL exclude unsupported template actions
Cabinet SHALL not show template actions not in Cabinet scope such as New Team or Billing.

#### Scenario: Header menu render
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** header account menu is opened
- **THEN** out-of-scope template actions SHALL be absent

### Requirement: Sidebar account panel SHALL exclude upsell rows
Cabinet SHALL not show upgrade prompt entries in sidebar account panel.

#### Scenario: Sidebar account panel render
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** sidebar footer account panel is rendered
- **THEN** upgrade placeholder rows SHALL not be present
