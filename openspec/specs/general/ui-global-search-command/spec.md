## Purpose
Define global search and command palette behavior for cross-screen navigation and actions.

## Requirements
### Requirement UI-GLOBAL-SEARCH-COMMAND-001: Command palette SHALL be globally invokable via keyboard shortcut
Cabinet SHALL provide a global command palette that can be opened by keyboard shortcut from authenticated routes.

#### Scenario: Open command palette with keyboard
- **GIVEN** authenticated shell is focused
- **WHEN** user presses configured command shortcut (`Ctrl/Cmd + K`)
- **THEN** command palette MUST open and focus command input

### Requirement UI-GLOBAL-SEARCH-COMMAND-002: Global command search SHALL include navigation targets and command actions
Cabinet SHALL return matching routes and supported global commands from command input query.

#### Scenario: Run route command
- **GIVEN** command palette is open
- **WHEN** user searches and selects a navigation command
- **THEN** router MUST navigate to selected destination and close palette

### Requirement UI-GLOBAL-SEARCH-COMMAND-003: Command palette SHALL support non-navigation global actions
Cabinet SHALL support global actions such as theme switching from command palette.

#### Scenario: Execute global theme action
- **GIVEN** command palette is open
- **WHEN** user selects a theme action command
- **THEN** theme state MUST update and palette MUST close
