## 1. Contract

- [x] 1.1 Select the maintained Go MCP protocol implementation for #1934.
- [x] 1.2 Define the local MCP trust boundary, transports, diagnostics, and
  packaging requirements in OpenSpec.
- [ ] 1.3 Add failing protocol tests for stdio initialize/capability
  negotiation, malformed messages, unknown methods, cancellation, and timeout.

## 2. Server foundation

- [x] 2.1 Add the Cabinet MCP server package and packaged stdio launcher.
- [ ] 2.2 Bind sessions to one explicit profile and reject missing or mismatched
  profile authority.
- [ ] 2.3 Add structured redacted receipts for material protocol operations.
- [ ] 2.4 Add optional loopback-only HTTP transport guarded by generated secret
  storage and disabled by default.

## 3. Settings and packaging

- [ ] 3.1 Add Settings > Integrations > MCP status, enable/disable, selected
  profile, client configuration guidance, and diagnostics.
- [ ] 3.2 Package the launcher with the Windows desktop application.

## 4. Evidence

- [ ] 4.1 Run targeted Go protocol and Settings tests.
- [ ] 4.2 Run MCP inspector/client stdio initialize smoke validation.
- [ ] 4.3 Run Windows package smoke validation for the launcher.
- [ ] 4.4 Run strict OpenSpec validation and record results on #1934.
