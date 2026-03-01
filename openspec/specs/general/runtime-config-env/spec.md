## Purpose
Define environment-driven runtime configuration for URL/port and related host settings.

## Requirements
### Requirement RUNTIME-CONFIG-ENV-001: Runtime host and port SHALL be configurable through environment variables
Cabinet SHALL read runtime host/port from environment values and use defaults when unset.

#### Scenario: Load host and port from environment
- **GIVEN** `.env` or process environment provides `CABINET_HOST` and `CABINET_PORT`
- **WHEN** runtime starts
- **THEN** server MUST bind to configured host/port and expose the same values in runtime metadata endpoint

### Requirement RUNTIME-CONFIG-ENV-002: Invalid environment configuration SHALL fail with actionable diagnostics
Cabinet SHALL fail fast for invalid host/port values and return actionable startup diagnostics.

#### Scenario: Invalid port value
- **GIVEN** `CABINET_PORT` is non-numeric or out of supported range
- **WHEN** runtime initializes server configuration
- **THEN** startup MUST fail with explicit config error message naming the invalid variable
