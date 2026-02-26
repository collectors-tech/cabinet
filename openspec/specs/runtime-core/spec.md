## Purpose
Define runtime, packaging, update, and health behavior for Cabinet desktop execution.

## Requirements
### Requirement RUNTIME-CORE-001: Cabinet SHALL run as a desktop-first local application
Cabinet SHALL run with embedded UI and local database storage on Windows and macOS.

#### Scenario: Runtime boot
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** the user starts Cabinet
- **THEN** the app SHALL start local runtime services and serve embedded UI

### Requirement RUNTIME-CORE-002: Cabinet SHALL support signed updates and release channels
Cabinet SHALL support signed update verification and stable/beta channel preferences.

#### Scenario: Update signature verification
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** an update package is evaluated
- **THEN** Cabinet SHALL verify signature before allowing install

### Requirement RUNTIME-CORE-003: Cabinet SHALL expose runtime diagnostics endpoints
Cabinet SHALL expose runtime and health diagnostics for local supportability.

#### Scenario: Runtime endpoint access
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** `/healthz` and `/api/runtime` are requested
- **THEN** Cabinet SHALL return runtime health payloads
- API outcome MUST be explicit: `200` on success, `4xx` for validation/auth conflicts, and `5xx` for unexpected failures
