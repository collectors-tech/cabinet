## Purpose
Define global keyboard shortcut behavior and platform-aware shortcut presentation.

## Requirements
### Requirement UI-KEYBOARD-SHORTCUTS-001: Shell shortcuts SHALL support platform-aware notation
Cabinet SHALL render keyboard shortcut labels using OS-appropriate notation for macOS and Windows.

#### Scenario: Render profile menu shortcuts
- **GIVEN** profile menu is opened in authenticated shell
- **WHEN** platform is detected as macOS or Windows
- **THEN** shortcut labels MUST use platform-appropriate notation

### Requirement UI-KEYBOARD-SHORTCUTS-002: Sidebar toggle shortcut SHALL be globally available in authenticated shell
Cabinet SHALL support keyboard shortcut for sidebar open/collapse behavior.

#### Scenario: Toggle sidebar with keyboard
- **GIVEN** authenticated shell has keyboard focus
- **WHEN** user presses sidebar toggle shortcut
- **THEN** sidebar state MUST toggle and remain consistent with current layout mode

### Requirement UI-KEYBOARD-SHORTCUTS-003: Shortcut map SHALL avoid collisions with core browser/app shortcuts
Cabinet SHALL maintain a managed shortcut map that avoids common collision patterns and remains configurable.

#### Scenario: Validate shortcut registration
- **GIVEN** shortcut map is loaded at app startup
- **WHEN** duplicate or reserved shortcut is registered
- **THEN** registration MUST be rejected with diagnostic entry and fallback shortcut policy
