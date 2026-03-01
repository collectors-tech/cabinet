## Purpose
Define global theme and display density controls available across the application shell.

## Requirements
### Requirement UI-THEME-SELECTION-001: Theme selector SHALL support light, dark, and system modes
Cabinet SHALL provide a global theme selector and apply selected mode across shell and pages.

#### Scenario: Change theme mode
- **GIVEN** authenticated workspace is loaded
- **WHEN** user selects `light`, `dark`, or `system` theme mode
- **THEN** UI MUST apply selected mode immediately and persist preference for subsequent sessions

### Requirement UI-THEME-SELECTION-002: Theme controls SHALL be accessible from global header and configuration panel
Cabinet SHALL expose theme controls in globally available UI controls.

#### Scenario: Open global theme controls
- **GIVEN** user is on any authenticated page
- **WHEN** user opens header theme control or config drawer
- **THEN** equivalent theme options MUST be available and stay synchronized

### Requirement UI-THEME-SELECTION-003: Density controls SHALL update layout spacing without route reload
Cabinet SHALL support density changes (for example compact/default/full) as a live layout preference.

#### Scenario: Change density
- **GIVEN** authenticated workspace is loaded
- **WHEN** user changes density setting
- **THEN** shell/content spacing MUST update without full-page reload and remain active after navigation
