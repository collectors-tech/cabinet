## Purpose
Define runtime, packaging, update, and health behavior for Cabinet desktop execution.

## Requirements
### Requirement: Cabinet SHALL run as a desktop-first local application
Cabinet SHALL run with embedded UI and local database storage on Windows and macOS.

#### Scenario: Runtime boot
- **WHEN** the user starts Cabinet
- **THEN** the app SHALL start local runtime services and serve embedded UI

### Requirement: Cabinet SHALL support signed updates and release channels
Cabinet SHALL support signed update verification and stable/beta channel preferences.

#### Scenario: Update signature verification
- **WHEN** an update package is evaluated
- **THEN** Cabinet SHALL verify signature before allowing install

### Requirement: Cabinet SHALL expose runtime diagnostics endpoints
Cabinet SHALL expose runtime and health diagnostics for local supportability.

#### Scenario: Runtime endpoint access
- **WHEN** `/healthz` and `/api/runtime` are requested
- **THEN** Cabinet SHALL return runtime health payloads
