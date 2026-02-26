## ADDED Requirements

### Requirement: Cypress startup on Windows MUST be deterministic
The system SHALL provide a single reliable command path to start Cypress-compatible runtime dependencies on Windows.

#### Scenario: Local startup on Windows
- **WHEN** a developer executes the documented Cypress startup command on Windows
- **THEN** required app services SHALL start successfully without unsupported option errors

#### Scenario: CI startup on Windows runner
- **WHEN** CI executes the same startup contract
- **THEN** Cypress preconditions SHALL be met and tests SHALL run without startup flag failures

### Requirement: Startup failures MUST fail fast with actionable diagnostics
The system SHALL emit clear, actionable error output when startup cannot satisfy Cypress preconditions.

#### Scenario: Startup command invalid
- **WHEN** a startup command uses unsupported arguments
- **THEN** the workflow SHALL fail fast and indicate the supported command path
