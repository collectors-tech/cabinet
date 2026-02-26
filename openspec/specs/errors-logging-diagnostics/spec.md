## Purpose
Define error handling, recovery, logging, and diagnostics behavior.

## Requirements
### Requirement: Cabinet SHALL provide clear, recoverable error flows
Cabinet SHALL present user-readable errors and recovery actions for known failure states.

#### Scenario: Scanner operation error
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** scanner provider call fails
- **THEN** Cabinet SHALL show clear failure message with retry path

### Requirement: Logging SHALL include activity and debug controls
Cabinet SHALL support activity logs, debug toggle, and export functionality.

#### Scenario: Export diagnostics log
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** user requests log export
- **THEN** Cabinet SHALL return diagnostics bundle output

### Requirement: Sensitive data SHALL be redacted in logs
Cabinet SHALL redact API keys, tokens, and credentials from log outputs.

#### Scenario: Log redaction
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** sensitive values are present in event context
- **THEN** persisted/exported logs SHALL contain redacted values

### Requirement: Recovery state SHALL surface abnormal shutdown context
Cabinet SHALL expose crash/recovery indicators to guide user remediation.

#### Scenario: Abnormal shutdown detected
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** runtime starts after abnormal termination
- **THEN** recovery state SHALL indicate diagnostics recommendation
