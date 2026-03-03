## Purpose
Define human-friendly Cabinet startup console output while preserving machine-readable startup telemetry.

## Requirements

### Requirement STARTUP-CONSOLE-001: Startup output SHALL include a human-readable banner
On successful runtime bind, console output SHALL include a readable startup banner with key runtime fields.

#### Scenario: Render startup banner
- **GIVEN** Cabinet starts successfully
- **WHEN** runtime bind completes
- **THEN** console MUST print a startup banner containing URL, instance, profile, data dir, and bind/port details

### Requirement STARTUP-CONSOLE-002: Startup output SHALL include machine-parseable structured line
Startup output SHALL include one structured machine line for automation parsing.

#### Scenario: Structured startup payload
- **GIVEN** Cabinet starts successfully
- **WHEN** startup output is emitted
- **THEN** console MUST include `CABINET_STARTUP_JSON { ... }` (or equivalent stable structured key/value JSON payload)

### Requirement STARTUP-CONSOLE-003: TTY/non-TTY rendering SHALL be deterministic
When attached to TTY, output MAY include color/icon formatting. When not TTY, output MUST degrade to plain text with no data loss.

#### Scenario: Non-TTY fallback
- **GIVEN** stdout is non-interactive (CI/log collector)
- **WHEN** startup output is emitted
- **THEN** startup banner MUST be plain text and preserve all key fields

## Use-Case IDs and E2E Mapping
| UC ID | Flow | Expected Result | E2E Mapping |
| --- | --- | --- | --- |
| UC-STC-01 | Successful startup | Human-readable banner prints required fields | planned: `tests/startup_console_banner_test.go` `startup-banner-fields` |
| UC-STC-02 | Structured startup line | Machine payload emitted for parser compatibility | planned: `tests/startup_console_banner_test.go` `startup-json-line` |
| UC-STC-03 | Non-TTY mode | Plain-text fallback emitted with same fields | planned: `tests/startup_console_banner_test.go` `startup-banner-non-tty` |
