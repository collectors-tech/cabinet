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

### Requirement UI-GLOBAL-SEARCH-COMMAND-004: Command palette SHALL surface local catalog matches
Cabinet SHALL search local inventory catalog records from the command palette and provide a direct filtered Inventory handoff for selected results.

#### Scenario: Open filtered Inventory from local catalog result
- **GIVEN** command palette is open
- **WHEN** user enters a local catalog query and selects a matching inventory result
- **THEN** router MUST navigate to Inventory with the selected part or title reflected in the filter query

### Requirement UI-GLOBAL-SEARCH-COMMAND-005: Command palette SHALL make unresolved barcode lookups actionable
Cabinet SHALL detect barcode-like command queries and offer a scanner/market-watch handoff when no local barcode match exists.

#### Scenario: Open scanner from unresolved barcode
- **GIVEN** command palette is open
- **WHEN** user enters a barcode-like query with no local match
- **THEN** command palette MUST expose an action that opens Scanner with the barcode prefilled as the query context

### Requirement UI-GLOBAL-SEARCH-COMMAND-006: Search trigger SHALL render platform-aware shortcut hint
Cabinet SHALL render the visible search trigger with a shortcut hint matching the user's platform convention.

#### Scenario: Display shortcut hint
- **GIVEN** authenticated shell is rendered
- **WHEN** user views the global search trigger
- **THEN** the trigger MUST show a platform-appropriate command shortcut hint

### Requirement UI-GLOBAL-SEARCH-COMMAND-007: Command palette SHALL expose local catalog loading, empty, and error states
Cabinet SHALL show deterministic local catalog search progress, empty, and unavailable states while command input queries the catalog API.

#### Scenario: Local catalog search states
- **GIVEN** command palette is open
- **WHEN** local catalog search is pending, returns no matches, or fails
- **THEN** command palette MUST show loading, empty, or unavailable feedback matching the catalog API outcome

### Requirement UI-GLOBAL-SEARCH-COMMAND-008: Command palette SHALL expose barcode lookup loading and error states
Cabinet SHALL show deterministic barcode lookup progress and unavailable states while barcode-like command input queries barcode lookup APIs.

#### Scenario: Barcode lookup states
- **GIVEN** command palette is open
- **WHEN** local catalog search has no barcode match and barcode lookup is pending or fails
- **THEN** command palette MUST show lookup progress and unavailable feedback for the barcode query
