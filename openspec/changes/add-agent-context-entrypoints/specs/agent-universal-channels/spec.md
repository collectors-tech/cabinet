## MODIFIED Requirements

### Requirement: Cabinet Agent SHALL expose universal governed entry points
Cabinet SHALL make Agent available from the main `/chats` workspace, side-panel
Chat, relevant table/detail screens, Inbox/action review flows, and approved
external channels without changing the safety boundary for the requested work.
Each in-app or approved external launch SHALL normalize context into a
canonical Agent context envelope before skill routing, clarification, preview,
apply, route-change, workflow, or audit handling.

#### Scenario: Open Agent from supported surfaces
- **GIVEN** an authenticated user is on `/chats`, a side-panel Chat surface, a supported table/detail screen, or an Inbox/action review flow
- **WHEN** the user opens Cabinet Agent
- **THEN** Agent SHALL preserve the active profile, route, thread, selected entity, and source surface context needed to answer or guide the request
- **AND** missing route, selection, profile, provider, permission, or setup context SHALL be reported as setup-needed or clarification-required guidance instead of guessed
- **AND** side-panel Chat SHALL remain available while non-mutating navigation or highlighting changes the main app route

#### Scenario: Normalize Agent launch context
- **GIVEN** a user opens Agent from main Chat, side-panel Chat, a supported table/detail surface, an Inbox/action-review entry, or an approved external channel
- **WHEN** Cabinet creates or dispatches the Agent request
- **THEN** Cabinet SHALL normalize the launch into an Agent context envelope containing profile/workspace ID, route or surface ID, selected record type and ID when available, active thread or conversation ID, user intent text, attachment or media IDs when available, source channel, permission and setup state, and workflow or audit IDs when present
- **AND** the envelope SHALL preserve only non-secret source metadata and SHALL NOT treat context presence as permission to bypass the active-profile authority policy or preview/apply boundary
- **AND** missing required fields SHALL produce deterministic clarification or setup guidance before any target-specific mutation is previewed or applied

#### Scenario: Explain available work before execution
- **GIVEN** Agent is opened from a supported surface
- **WHEN** the user asks what Agent can do
- **THEN** Agent SHALL explain available skills and capabilities for the current context
- **AND** it SHALL distinguish read-only, preview-only, confirm-required, external-write, blocked, and unavailable work
- **AND** unavailable or unsupported work SHALL include the next safe user action rather than silent omission or hallucinated capability

#### Scenario: Preserve route-change continuity
- **GIVEN** an Agent thread or workflow has an active Agent context envelope
- **WHEN** Agent performs a governed route change, surface switch, or selected-record highlight
- **THEN** Cabinet SHALL preserve the active profile, thread, workflow or audit identifier, source surface, and prior intent context needed to continue the workflow
- **AND** the side-panel Chat state SHALL remain available after the route change
- **AND** the changed route or selected record context SHALL be recorded as an updated envelope value rather than losing the original thread context

#### Scenario: Record context evidence for review
- **GIVEN** Agent uses a context envelope to explain, route, preview, apply, clarify, or change route
- **WHEN** Cabinet records workflow, Action Timeline, audit, or diagnostic evidence
- **THEN** the evidence SHALL include non-secret profile, source surface/channel, route, selected record reference, thread/conversation, skill/capability, and workflow/audit identifiers sufficient to review what context Agent used
- **AND** the evidence SHALL omit secrets, raw provider credentials, arbitrary local file paths, and unrelated profile context
