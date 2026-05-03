## MODIFIED Requirements

### Requirement: Runtime startup SHALL negotiate fallback port when requested port is occupied by a non-Cabinet listener
When the requested runtime port is already in use by a non-Cabinet process, startup MUST bind deterministically to the next available port within the fallback scan window and continue serving.

#### Scenario: Requested port occupied by another process
- **GIVEN** runtime attempts to bind `host:requested_port`
- **AND** that endpoint is occupied by another process that is not a healthy Cabinet runtime
- **WHEN** startup bind resolution occurs
- **THEN** runtime MUST attempt fallback ports (`requested_port+1 ... requested_port+N`) and bind the first available address
- **AND** startup output MUST preserve `requested_port` and report actual `resolved_port`
- **AND** resolved `/healthz` endpoint on the fallback URL MUST return `200`

### Requirement: Runtime startup SHALL use PID-only lock file, attach to setup-metadata runtime, and reuse or restart a healthy requested Cabinet endpoint deterministically
Runtime startup MUST keep `cabinet.pid` as PID-only lock signal, MUST resolve attach/open target from runtime setup metadata (`cabinet.json`) instead of storing endpoint metadata in the PID file, and MUST reuse a healthy Cabinet runtime already serving the requested endpoint before starting a second server process unless explicit restart behavior is requested.

#### Scenario: Attach to already-running runtime in same data directory
- **GIVEN** startup data directory contains `cabinet.pid` with an alive PID and `cabinet.json` runtime metadata (`runtime.resolvedUrl` or `meta.currentUrl`) for a healthy endpoint
- **WHEN** Cabinet launcher starts again for the same data directory
- **THEN** startup MUST attach to the existing runtime URL and open/attach the browser target without starting a second server process
- **AND** runtime logs MUST emit a deterministic attach line containing `url`, `pid`, `data_dir`, and resolved port

#### Scenario: Reuse healthy Cabinet runtime already serving requested endpoint
- **GIVEN** Cabinet launcher starts for a requested `host:port`
- **AND** the requested endpoint is already serving a healthy Cabinet runtime
- **AND** explicit parallel mode is not enabled
- **AND** explicit restart mode is not enabled
- **WHEN** startup attach resolution executes
- **THEN** startup MUST reuse the requested endpoint instead of starting a new server process
- **AND** runtime logs MUST emit a deterministic attach line describing requested-endpoint reuse
- **AND** startup MUST NOT silently fall through to a fallback port in this case

#### Scenario: Restart healthy Cabinet runtime already serving requested endpoint
- **GIVEN** Cabinet launcher starts for a requested `host:port`
- **AND** the requested endpoint is already serving a healthy Cabinet runtime
- **AND** explicit restart mode is enabled
- **WHEN** startup restart resolution executes
- **THEN** startup MUST identify the active Cabinet PID from runtime lifecycle metadata and/or pid file
- **AND** startup MUST stop the existing Cabinet process before binding the new runtime
- **AND** startup MUST wait until the requested endpoint is no longer in use before starting the replacement runtime
- **AND** startup logs MUST emit deterministic restart diagnostics containing the requested endpoint, prior PID, and restart outcome
- **AND** the replacement runtime MUST bind the originally requested port rather than silently port-falling back

#### Scenario: Restart requested but endpoint is not a healthy Cabinet runtime
- **GIVEN** Cabinet launcher starts for a requested `host:port`
- **AND** explicit restart mode is enabled
- **AND** the requested endpoint is occupied by a non-Cabinet listener or cannot be verified as a healthy Cabinet runtime
- **WHEN** startup restart resolution executes
- **THEN** startup MUST NOT terminate the existing listener
- **AND** runtime MUST fail with an actionable restart validation error instead of force-replacing an unknown process

#### Scenario: Explicit parallel mode bypasses requested-endpoint reuse
- **GIVEN** Cabinet launcher starts for a requested `host:port`
- **AND** the requested endpoint is already serving a healthy Cabinet runtime
- **AND** explicit parallel mode is enabled
- **WHEN** startup attach resolution executes
- **THEN** startup MUST bypass requested-endpoint reuse behavior
- **AND** runtime MAY continue into explicit parallel/fallback startup behavior consistent with the launch request

#### Scenario: Stale PID recovery
- **GIVEN** startup data directory contains stale `cabinet.pid` or metadata endpoint health check fails
- **WHEN** attach resolution executes
- **THEN** runtime MUST remove the stale PID file and continue with fresh startup
- **AND** no attach/open action MUST occur for stale lock state

#### Scenario: Startup reporting includes endpoint-occupied status
- **GIVEN** Cabinet launcher starts for a requested `host:port`
- **WHEN** startup preflight determines whether the requested endpoint is free, occupied by Cabinet, or occupied by another listener
- **THEN** startup diagnostics MUST report that endpoint status
- **AND** when attach or restart behavior is selected, diagnostics MUST report the selected action before runtime start/exit completes
