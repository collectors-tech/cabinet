## Why

Cabinet has no local Model Context Protocol server, transport, or conformance
surface today. That leaves future Agent Skill exposure without a stable trust
boundary for desktop clients, profile selection, credentials, diagnostics, or
packaged Windows launch.

## What Changes

- Add a local-first Cabinet MCP server foundation using the official Go MCP SDK
  (`github.com/modelcontextprotocol/go-sdk/mcp`), which provides maintained
  server/client/session primitives and stdio transport support.
- Add a packaged stdio launcher for desktop MCP clients and an optional
  loopback-only Streamable HTTP transport that is disabled by default.
- Bind every MCP session to one explicit Cabinet profile and reject ambiguous
  or missing profile selection.
- Store loopback credentials through Cabinet secret storage and redact them from
  logs, Settings status, diagnostics, and protocol errors.
- Add protocol handling for initialize/capability negotiation, unknown methods,
  invalid JSON-RPC messages, cancellation, timeouts, and structured errors.
- Add Settings > Integrations > MCP status, enable/disable, profile selection,
  local client configuration guidance, and diagnostics.
- Add packaging and smoke-test coverage for the Windows desktop application.

## Capabilities

### New Capabilities

- `mcp-local-server`: local Cabinet MCP server lifecycle, transports, profile
  binding, trust boundary, diagnostics, and packaging.

### Modified Capabilities

- `agent-universal-channels`: reserve MCP as a governed Agent entry point that
  inherits the active profile and permission boundary before any tool/resource
  exposure is implemented.

## Impact

- Affected code: `internal/mcp`, `cmd/cabinet-mcp`, `internal/app`,
  `internal/settings`, `internal/secrets`, `ui.web`, and Windows packaging
  scripts.
- Affected tests: Go protocol/unit tests, Settings API/UI tests, MCP inspector
  or client conformance smoke tests, and Windows packaging smoke tests.
- Affected documentation: Settings guidance, OpenSpec specs, and issue evidence.
- Related issues: `#1934`, `#1932`, `#1935`, `#1701`.
