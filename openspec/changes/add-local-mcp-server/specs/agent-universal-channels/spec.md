# agent-universal-channels Specification

## Modified Requirements

### Requirement: Cabinet Agent SHALL expose universal governed entry points

Cabinet SHALL make Agent available from the main `/chats` workspace, side-panel
Chat, relevant table/detail screens, Inbox/action review flows, local MCP, and
approved external channels without changing the safety boundary for the
requested work.

#### Scenario: Open Agent from supported surfaces

- **GIVEN** an authenticated user is on `/chats`, a side-panel Chat surface, a
  supported table/detail screen, an Inbox/action review flow, or an initialized
  local MCP session
- **WHEN** the user opens Cabinet Agent or invokes an Agent entry point through
  local MCP
- **THEN** Agent SHALL preserve the active profile, route, thread, selected
  entity, session identity, and source surface context needed to answer or guide
  the request
- **AND** missing route, selection, profile, provider, permission, session
  authority, or setup context SHALL be reported as setup-needed or
  clarification-required guidance instead of guessed
- **AND** side-panel Chat SHALL remain available while non-mutating navigation or
  highlighting changes the main app route

#### Scenario: Explain available work before execution

- **GIVEN** Agent is opened from a supported surface
- **WHEN** the user asks what Agent can do
- **THEN** Agent SHALL explain available skills and capabilities for the current context
- **AND** it SHALL distinguish read-only, preview-only, confirm-required, external-write, blocked, and unavailable work
- **AND** unavailable or unsupported work SHALL include the next safe user action rather than silent omission or hallucinated capability
