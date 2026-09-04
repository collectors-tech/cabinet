## MODIFIED Requirements

### Requirement: Cabinet Agent SHALL expose universal governed entry points
Cabinet SHALL make Agent available from the main `/chats` workspace, side-panel Chat, relevant table/detail screens, Inbox/action review flows, and approved external channels without changing the safety boundary for the requested work.

#### Scenario: Open Agent from supported surfaces
- **GIVEN** an authenticated user is on `/chats`, a side-panel Chat surface, a supported table/detail screen, or an Inbox/action review flow
- **WHEN** the user opens Cabinet Agent
- **THEN** Agent SHALL preserve the active profile, route, thread, selected entity, and source surface context needed to answer or guide the request
- **AND** missing route, selection, profile, provider, permission, or setup context SHALL be reported as setup-needed or clarification-required guidance instead of guessed
- **AND** side-panel Chat SHALL remain available while non-mutating navigation or highlighting changes the main app route

#### Scenario: Explain available work before execution
- **GIVEN** Agent is opened from a supported surface
- **WHEN** the user asks what Agent can do
- **THEN** Agent SHALL explain available skills and capabilities for the current context
- **AND** it SHALL distinguish read-only, preview-only, confirm-required, external-write, blocked, and unavailable work
- **AND** unavailable or unsupported work SHALL include the next safe user action rather than silent omission or hallucinated capability

#### Scenario: Reuse context envelope for natural-language skill planning
- **GIVEN** Agent receives a natural-language request from main Chat, side-panel Chat, a selected app surface, or an approved external channel
- **WHEN** Cabinet asks the governed planner to choose or clarify a skill
- **THEN** the planner input SHALL use the canonical Agent context envelope for profile, route, selected record, active thread, user intent, attachment or media IDs, source channel, permission/setup state, and workflow/audit identifiers
- **AND** context SHALL be sufficient for selected-record work such as renaming the current inventory item when the selection is present
- **AND** missing or ambiguous context SHALL produce clarification or setup guidance before any target-specific mutation preview
- **AND** context presence SHALL NOT bypass the active profile authority policy, preview/apply boundary, external-write approval, destructive confirmation, or profile isolation
