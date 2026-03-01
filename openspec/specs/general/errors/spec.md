## Purpose
Define application error surfacing, user recovery actions, and failure-state UX contracts.

## Requirements
### Requirement ERRORS-001: Cabinet SHALL provide clear app error flows
Cabinet SHALL present user-readable errors with contextual recovery actions for known failure states across scanner, import, auth, and integrations workflows.

#### Scenario: Scanner operation error
- **GIVEN** scanner provider call fails during execution
- **WHEN** failure state is returned to UI
- **THEN** Cabinet SHALL show clear error message, reason category, and retry action

### Requirement ERRORS-002: Error screens SHALL preserve safe navigation
Cabinet SHALL provide safe route recovery controls (`go back`, `go home`, and `retry`) without data-destructive side effects.

#### Scenario: Route-level runtime error
- **GIVEN** a route throws runtime error
- **WHEN** error boundary renders fallback
- **THEN** user SHALL be offered safe navigation and retry options

### Requirement ERRORS-003: Error taxonomy SHALL map to deterministic user actions
Cabinet SHALL classify errors into validation, provider, connectivity, authorization, and internal categories with deterministic next action guidance.

#### Scenario: Error category guidance
- **GIVEN** an error is classified as provider failure
- **WHEN** user views error details
- **THEN** UI SHALL show provider-oriented remediation guidance

