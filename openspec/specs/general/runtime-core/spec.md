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

### Requirement RUNTIME-CORE-004: Startup console output SHALL report resolved runtime endpoint and execution context
After successful listener bind, Cabinet MUST print a machine-parseable startup line containing resolved URL and runtime context.

#### Scenario: Startup console line after bind
- **GIVEN** runtime starts with configured listen address and writable data directory
- **WHEN** listener bind succeeds and resolved address is known
- **THEN** console output MUST include `url=<resolved-url>` matching actual bound endpoint
- **AND** output MUST include `instance`, `profile`, and `data_dir` context values
- **AND** when requested port and resolved port differ, output MUST include both `requested_port` and `resolved_port`

### Requirement RUNTIME-CORE-005: Startup console output SHALL include human banner and structured JSON line
After successful listener bind, Cabinet MUST emit human-readable startup lines and a structured JSON line while preserving existing key-value machine output.

#### Scenario: Human startup banner lines
- **GIVEN** runtime starts and listener bind succeeds with resolved address known
- **WHEN** startup console output is emitted
- **THEN** output MUST include human-readable lines containing `Cabinet Started`, `URL`, `Instance`, `Profile`, `Data Dir`, `Port`, and `Bind`
- **AND** when stdout is TTY, banner title MAY include emoji decoration
- **AND** when stdout is non-TTY, banner title MUST fall back to plain text without emoji requirement

#### Scenario: Structured startup JSON line
- **GIVEN** runtime starts and listener bind succeeds
- **WHEN** startup console output is emitted
- **THEN** output MUST include exactly one line prefixed `CABINET_STARTUP_JSON `
- **AND** the JSON payload MUST include keys `url`, `requested_addr`, `resolved_addr`, `instance`, `profile`, `data_dir`, `requested_port`, and `resolved_port`
- **AND** existing key-value startup line `CABINET_STARTUP ...` MUST remain present for backwards compatibility
