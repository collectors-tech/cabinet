# mcp-local-server Specification

## Purpose

Define Cabinet's local Model Context Protocol server foundation, trust
boundary, transports, diagnostics, and packaging requirements before tools or
resources are exposed.

## ADDED Requirements

### Requirement: Cabinet SHALL expose a local MCP server with explicit identity

Cabinet SHALL provide a local MCP server identity that describes the Cabinet
instance, selected profile, app version digest, and supported protocol
capabilities without granting ambient access to other profiles.

#### Scenario: Initialize over stdio

- **GIVEN** a packaged Cabinet MCP launcher is started by a local desktop MCP
  client with an explicit profile selection
- **WHEN** the client sends `initialize` over stdio
- **THEN** Cabinet SHALL complete protocol-version negotiation with server
  identity and capability advertisement
- **AND** the advertised identity SHALL include non-secret Cabinet instance,
  selected profile, and app/version evidence
- **AND** capability advertisement SHALL NOT imply permission to access any
  profile, tool, or resource that has not been authorized for that session

#### Scenario: Reject ambiguous profile binding

- **GIVEN** a client starts an MCP session without an explicit profile, with an
  unknown profile, or with a profile that does not belong to the active Cabinet
  data directory
- **WHEN** Cabinet initializes the session
- **THEN** Cabinet SHALL reject the session with a structured protocol error
- **AND** it SHALL NOT silently fall back to a default or previously active
  profile

### Requirement: MCP transports SHALL be local-first and credential protected

Cabinet SHALL support stdio by default and MAY support Streamable HTTP only when
explicitly enabled for loopback clients with generated credentials.

#### Scenario: Optional HTTP transport is loopback-only

- **GIVEN** the user enables the optional MCP HTTP transport
- **WHEN** Cabinet starts or updates the listener
- **THEN** it SHALL bind only to loopback addresses
- **AND** it SHALL reject non-loopback host configuration
- **AND** it SHALL require the generated session credential before accepting
  protocol messages
- **AND** disabling the transport SHALL stop accepting new HTTP sessions
  immediately

#### Scenario: Credentials are stored and redacted

- **GIVEN** Cabinet generates or rotates an MCP transport credential
- **WHEN** settings, diagnostics, logs, errors, backup/export, or UI status are
  rendered
- **THEN** the credential SHALL remain inside Cabinet secret storage
- **AND** no credential value, provider key, or sensitive tool payload SHALL be
  written to ordinary settings, logs, diagnostic events, or protocol errors

### Requirement: MCP protocol failures SHALL be structured and non-fatal

Cabinet SHALL handle invalid messages, unknown methods, cancellation, and
timeouts without crashing the application or leaking sensitive state.

#### Scenario: Invalid protocol input returns compliant errors

- **GIVEN** an MCP client sends malformed JSON-RPC, an unknown method, invalid
  parameters, or a request that exceeds the configured timeout
- **WHEN** Cabinet processes the message
- **THEN** it SHALL return a protocol-compliant structured error
- **AND** the server process SHALL remain available for subsequent valid
  messages when the transport permits it
- **AND** the diagnostic receipt SHALL identify the operation class and outcome
  without storing secret values or full sensitive payloads

#### Scenario: Cancellation is scoped to one operation

- **GIVEN** a client cancels an in-flight MCP operation
- **WHEN** Cabinet receives the cancellation
- **THEN** it SHALL cancel only that operation for the owning session
- **AND** it SHALL preserve a redacted receipt with operation id, session
  identity, selected profile, capability, Cabinet version digest, redacted input
  class, and outcome

### Requirement: Settings SHALL report MCP state truthfully

Cabinet SHALL expose MCP configuration and diagnostics in Settings >
Integrations without overstating readiness or exposing secrets.

#### Scenario: Show MCP runtime status and recovery guidance

- **GIVEN** an authenticated user opens Settings > Integrations > MCP
- **WHEN** Cabinet reports the MCP integration state
- **THEN** it SHALL distinguish running, stopped, disabled, misconfigured, and
  error states
- **AND** it SHALL show the selected profile, available local transports,
  non-secret client configuration guidance, last diagnostic outcome, and next
  recovery action
- **AND** it SHALL NOT mark MCP ready when no profile is selected or a required
  transport credential is missing

### Requirement: Windows packaging SHALL include the MCP launcher

Cabinet SHALL package and smoke-test the local MCP launcher with the Windows
desktop application.

#### Scenario: Packaged launcher initializes

- **GIVEN** a packaged Windows Cabinet build is produced
- **WHEN** the MCP launcher is invoked with a test profile and a supported MCP
  inspector or client smoke harness
- **THEN** the client SHALL complete stdio initialize and capability negotiation
- **AND** package validation SHALL fail if the launcher is missing, cannot find
  Cabinet runtime dependencies, or emits non-protocol output on stdio
