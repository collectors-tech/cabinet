## Purpose
Define runtime, packaging, update, and health behavior for Cabinet desktop execution.

## Requirements
### Requirement RUNTIME-CORE-001: Cabinet SHALL run as a desktop-first local application
Cabinet SHALL run with embedded UI and local database storage on Windows and macOS.

#### Scenario: Runtime boot
- **GIVEN** Cabinet binary is launched on supported host OS (Windows/macOS) with writable data directory and SQLite path
- **WHEN** the user starts Cabinet
- **THEN** the app SHALL start local runtime services, serve embedded UI, and return `200` for `GET /healthz` within startup timeout

### Requirement RUNTIME-CORE-002: Cabinet SHALL support signed updates and release channels
Cabinet SHALL support signed update verification and stable/beta channel preferences.

#### Scenario: Update signature verification
- **GIVEN** update channel is configured and downloaded update includes payload plus signature
- **WHEN** an update package is evaluated
- **THEN** Cabinet SHALL verify signature before allowing install

### Requirement RUNTIME-CORE-003: Cabinet SHALL expose runtime diagnostics endpoints
Cabinet SHALL expose runtime and health diagnostics for local supportability.

#### Scenario: Runtime endpoint access
- **GIVEN** runtime server is started and health endpoints are reachable on configured local address
- **WHEN** `/healthz` and `/api/runtime` are requested
- **THEN** Cabinet SHALL return runtime health payloads
  - `GET /healthz` MUST return `200` with body `ok`
  - `GET /api/runtime` MUST return `200` with `app_version` and `build_date`
