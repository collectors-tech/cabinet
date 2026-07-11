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
  - release builds with an explicit semantic beta version MUST report that version in `app_version` instead of only a revision-derived value

### Requirement RUNTIME-CORE-019: Beta packaging SHALL produce truthful Windows portable artefacts
Cabinet beta packaging SHALL use one canonical private-beta version source and SHALL produce Windows portable package evidence without claiming unsigned installers.

#### Scenario: Windows portable beta package
- **GIVEN** the canonical beta version source names a private beta version
- **WHEN** Cabinet builds release artefacts for Windows beta validation
- **THEN** the package filename SHALL include the beta version and `windows-amd64-portable`
- **AND** the runtime binary SHALL embed the same semantic version, commit revision, and build date for `/api/runtime`
- **AND** packaging SHALL create a SHA-256 checksum file and release notes
- **AND** macOS artefacts SHALL NOT be claimed by the Windows beta package lane until separately signed and validated
- **AND** OpenSpec release guidance SHALL describe install/start, data location, backup/upgrade, rollback/removal, signing limits, and release approval gates

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

### Requirement RUNTIME-CORE-006: Project-local execution SHALL prefer `bin` folder runtime path over ephemeral temp locations
When running from a project workspace, startup and validation workflows MUST prefer executable path under project-local `bin` (or equivalent configured project runtime path) and MUST NOT default to transient template/temp directories.

### Requirement RUNTIME-CORE-007: CLI SHALL support browser auto-open suppression for automation runs
Runtime CLI SHALL provide a flag to suppress browser auto-open on startup (e.g., `--no-open-browser`) for CI/agent/Cypress flows.

#### Scenario: No-open-browser automation startup
- **GIVEN** Cabinet is launched with browser-suppression flag
- **WHEN** startup completes successfully
- **THEN** runtime MUST serve normally without opening a browser window/tab
- **AND** startup output MUST explicitly note browser auto-open is disabled

#### Scenario: Default interactive startup
- **GIVEN** Cabinet is launched without browser-suppression flag
- **WHEN** startup completes successfully
- **THEN** default browser-open behavior MUST remain unchanged unless overridden by config

#### Scenario: Project run-path resolution
- **GIVEN** Cabinet project root is available and contains `bin/cabinet(.exe)`
- **WHEN** run instructions or automation resolves executable path
- **THEN** runtime MUST launch from project-local `bin` executable by default
- **AND** logs/checkpoints MUST record resolved executable path used for run

#### Scenario: Equivalent configured runtime path
- **GIVEN** project defines an explicit runtime executable path different from `bin`
- **WHEN** run instructions resolve launch target
- **THEN** runtime MAY use configured project-local equivalent path
- **AND** transient template/temp-folder executable paths MUST be rejected unless explicitly forced for a test case

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

### Requirement RUNTIME-CORE-008: Runtime CLI SHALL support deterministic startup parameter overrides
Runtime startup SHALL support explicit CLI overrides for address/port, data path, profile context, auth mode, base URL, parallel guard, and log level.

#### Scenario: Launch with explicit startup overrides
- **GIVEN** operator launches Cabinet with `--port`, `--data-dir`, `--profile` (or `--instance-name`), `--auth-mode`, `--base-url`, `--allow-parallel`, and `--log-level`
- **WHEN** startup arguments are parsed before runtime config load
- **THEN** CLI values MUST override process/env defaults deterministically for that process launch
- **AND** runtime MUST reject invalid values with actionable startup validation errors
- **AND** conflicting `--port` and `--listen` values MUST fail fast with deterministic conflict error

### Requirement RUNTIME-CORE-009: Runtime startup SHALL print effective resolved configuration snapshot
After CLI/env resolution and before service run, startup output MUST emit a deterministic effective configuration line.

#### Scenario: Effective startup config line
- **GIVEN** runtime has resolved startup configuration from CLI + environment
- **WHEN** startup output is emitted
- **THEN** output MUST include one line prefixed `CABINET_EFFECTIVE_CONFIG`
- **AND** the line MUST include `addr`, `host`, `port`, `data_dir`, `profile`, `auth_mode`, `base_url`, `allow_parallel`, and `log_level`

### Requirement RUNTIME-CORE-010: Default runtime data directory SHALL resolve to executable-local path first
When no explicit data directory override is provided, runtime MUST resolve data storage under executable-local path before falling back to global/shared OS locations.

#### Scenario: Fresh runtime launch without explicit data-dir override
- **GIVEN** Cabinet launches without `CABINET_DATA_DIR` and without CLI `--data-dir`
- **WHEN** startup resolves storage paths
- **THEN** default data directory MUST resolve to `<exe_dir>/data`
- **AND** diagnostics/startup output MUST surface resolved data directory
- **AND** explicit `--data-dir`/`CABINET_DATA_DIR` override MUST continue to take precedence

### Requirement RUNTIME-CORE-011: Runtime startup SHALL negotiate fallback port when requested port is occupied
When requested runtime port is already in use, startup MUST bind deterministically to the next available port within fallback scan window and continue serving.

#### Scenario: Requested port occupied on startup
- **GIVEN** runtime attempts to bind `host:requested_port` and that port is occupied by another process
- **WHEN** startup bind occurs
- **THEN** runtime MUST attempt fallback ports (`requested_port+1 ... requested_port+N`) and bind first available address
- **AND** startup output MUST preserve `requested_port` and report actual `resolved_port`
- **AND** resolved `/healthz` endpoint on fallback URL MUST return `200`

### Requirement RUNTIME-CORE-012: Runtime startup SHALL use PID-only lock file and attach to running endpoint from setup metadata
Runtime startup MUST keep `cabinet.pid` as PID-only lock signal and MUST resolve attach/open target from runtime setup metadata (`cabinet.json`) instead of storing endpoint metadata in PID file.

#### Scenario: Attach to already-running runtime in same data directory
- **GIVEN** startup data directory contains `cabinet.pid` with alive PID and `cabinet.json` runtime metadata (`runtime.resolvedUrl` or `meta.currentUrl`) for a healthy endpoint
- **WHEN** Cabinet launcher starts again for the same data directory
- **THEN** startup MUST attach to existing runtime URL and open/attach browser target without starting a second server process
- **AND** runtime logs MUST emit deterministic attach line containing `url`, `pid`, `data_dir`, and resolved port

#### Scenario: Stale PID recovery
- **GIVEN** startup data directory contains stale `cabinet.pid` or metadata endpoint health check fails
- **WHEN** attach resolution executes
- **THEN** runtime MUST remove stale PID file and continue with fresh startup
- **AND** no attach/open action MUST occur for stale lock state

### Requirement RUNTIME-CORE-013: Runtime SHALL persist lifecycle metadata in `cabinet.json` after setup configuration exists
When runtime setup config exists, Cabinet MUST enrich `cabinet.json` with last-known lifecycle provenance for startup and shutdown without over-claiming current liveness.

#### Scenario: Lifecycle metadata after clean run
- **GIVEN** `cabinet.json` already exists for the runtime data directory
- **WHEN** Cabinet starts, serves requests, and shuts down cleanly
- **THEN** `cabinet.json` MUST persist `startedAt`, `startedBy`, `launchSource`, `lastKnownPid`, `lastKnownUrl`, `lastHeartbeatAt`, `lastShutdownAt`, `lastShutdownReason`, and `lastRunClean`
- **AND** `lastRunClean` MUST be `true` after clean shutdown
- **AND** metadata MUST be treated as last-known run facts, not proof of live process state

### Requirement RUNTIME-CORE-014: Runtime SHALL write structured durable runtime/access/error logs under the active data directory
Cabinet MUST write machine-readable JSONL log files for runtime lifecycle, request access, and runtime/server errors.

#### Scenario: Structured runtime log files during local execution
- **GIVEN** Cabinet starts successfully with writable data directory
- **WHEN** runtime lifecycle events occur and HTTP requests are served
- **THEN** Cabinet MUST create a fresh timestamped JSONL log set under the active data directory's `logs/` subfolder for runtime, access, and error streams on every process start
- **AND** runtime log entries MUST append only to that process start's own runtime log file and include timestamp, level, event/type, and process/runtime context
- **AND** access log entries MUST include method, path, status, and duration
- **AND** fixed shared filenames like `cabinet.runtime.log`, `cabinet.access.log`, and `cabinet.error.log` MUST NOT be reused across multiple starts, and the timestamped files MUST live under that `logs/` subfolder

### Requirement RUNTIME-CORE-015: Runtime startup SHALL remain deterministic under parallel local validation load
Parallel local validation runs that create many fresh runtime instances concurrently MUST avoid self-inflicted migration deadline failures.

#### Scenario: Parallel fresh-db startup in local validation
- **GIVEN** multiple local validation/test workers create fresh Cabinet data directories and SQLite paths concurrently
- **WHEN** each worker opens and migrates its own fresh runtime database
- **THEN** runtime migration/open behavior MUST complete without deterministic deadline failures caused only by local parallel startup pressure

### Requirement RUNTIME-CORE-016: Runtime run loop SHALL fast-exit when startup context is already canceled
If the supplied runtime context is already canceled before run-loop startup begins, Cabinet MUST not bind/listen/start the server just to shut it down again.

#### Scenario: Pre-canceled startup context
- **GIVEN** caller invokes runtime `Run` with a context that is already canceled
- **WHEN** run-loop startup begins
- **THEN** runtime MUST return quickly without binding a listener or writing startup lifecycle artifacts
- **AND** local NFR/startup checks MUST observe the fast-exit behavior rather than an unnecessary shutdown timeout path

### Requirement RUNTIME-CORE-017: E2E reset hooks SHALL tolerate legacy/shared schema drift
E2E-only reset hooks MUST clear supported runtime data without failing solely because a known reset table is absent from a legacy or shared managed test database.

#### Scenario: Reset with missing known table
- **GIVEN** an E2E-enabled managed runtime uses a database where a known reset table is absent
- **WHEN** `/api/test/reset` or the reset hook clears E2E state
- **THEN** reset MUST skip the absent table and continue clearing present reset tables
- **AND** reset MUST still fail on real delete/query errors for tables that exist

### Requirement RUNTIME-CORE-018: Cypress runner SHALL prepare dependencies and persist execution logs
Cabinet Cypress execution scripts MUST perform required local preparation before invoking Cypress and MUST persist progress/output logs for traceability.

#### Scenario: Cypress execution from clean or temporary worktree
- **GIVEN** a Cabinet worktree has the Cypress spec and runtime config but does not have prepared `ui.web/node_modules`
- **WHEN** the Cypress runner starts
- **THEN** it MUST prepare UI dependencies by reusing a configured/local Cabinet `node_modules` install when available or by running deterministic install from the UI lockfile
- **AND** it MUST build static UI assets and the project-local `bin/cabinet(.exe)` runtime before starting the validation server unless an explicit build skip is requested
- **AND** it MUST recycle an existing listener on the target base URL by default so validation runs against the current worktree, unless explicit server reuse is requested
- **AND** it MUST log each preparation, runtime, and Cypress execution step to a timestamped run log
- **AND** it MUST write a machine-readable run summary containing the spec, browser, base URL, runtime path, exit code, and ordered step list
- **AND** it MUST retain existing runtime health, E2E hook, stale-port recycling, and project-local runtime path protections

#### Scenario: Cypress runner fails on stale runtime app version
- **GIVEN** the Cypress runner has resolved the current Git `source_commit`
- **AND** the managed or reused runtime returns `/api/runtime.app_version`
- **WHEN** the app version does not equal `rev-<source_commit first 12 chars>`
- **THEN** the runner MUST fail before executing the browser spec with a diagnostic stale-runtime mismatch error
- **AND** the run summary MUST record whether stale runtime app versions were explicitly allowed
- **AND** the runner MAY proceed only when an explicit stale-runtime baseline override is passed
