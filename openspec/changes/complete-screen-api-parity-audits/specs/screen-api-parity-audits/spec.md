## ADDED Requirements

### Requirement: Authenticated production routes MUST map to real API workflows
The system SHALL ensure each authenticated route has explicit API read/mutation mappings or documented intentional no-data behavior.

#### Scenario: Route parity verification
- **WHEN** a parity audit is executed for an authenticated route
- **THEN** all displayed data and mutations SHALL map to defined Cabinet API endpoints

### Requirement: Placeholder/sample content MUST NOT appear on production routes
The system SHALL not render template/sample placeholder content in authenticated production workflows.

#### Scenario: Placeholder content detection
- **WHEN** a user navigates production routes
- **THEN** screen content SHALL originate from API-backed or explicitly computed runtime state
