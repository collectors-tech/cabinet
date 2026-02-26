## Purpose
Define activity and operational logging policy including explicit API request logging and error logging triggers.

## Requirements
### Requirement LOGGING-001: API request logging SHALL be explicit and bounded
Cabinet SHALL log API request metadata (timestamp, route, method, profile, status code, duration, correlation ID) for auditable operations.

#### Scenario: API request logging trigger
- **GIVEN** an API request enters runtime handler
- **WHEN** request completes (success or failure)
- **THEN** request metadata SHALL be recorded in activity log with correlation ID

### Requirement LOGGING-002: Error logging SHALL trigger on failed operations
Cabinet SHALL emit error log records for unhandled exceptions, failed provider calls, and non-2xx API responses that represent actionable failures.

#### Scenario: Error logging trigger
- **GIVEN** API call returns actionable failure state
- **WHEN** runtime finalizes response
- **THEN** error logging SHALL persist failure context with redacted sensitive values

### Requirement LOGGING-003: Logging SHALL support debug mode and export
Cabinet SHALL support debug-level logging toggle and diagnostics export bundle generation.

#### Scenario: Export diagnostics log
- **GIVEN** log records exist for active profile/runtime
- **WHEN** user requests log export
- **THEN** Cabinet SHALL generate export bundle containing activity and error logs

### Requirement LOGGING-004: Sensitive data SHALL be redacted in logs
Cabinet SHALL redact tokens, API keys, credentials, and secrets from persisted and exported logs.

#### Scenario: Log redaction
- **GIVEN** sensitive values are present in event context
- **WHEN** log record is written or exported
- **THEN** sensitive fields SHALL be redacted before storage/output

