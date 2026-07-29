## MODIFIED Requirements

### Requirement: OpenAI-backed assistant capabilities SHALL expose truthful setup and readiness states
Cabinet MUST treat OpenAI/API-key/Browser Auth readiness as provider evidence that gates capabilities, not as copy or navigation state.

#### Scenario: Plan governed skill work from provider output
- **GIVEN** the active profile has a healthy configured assistant provider and enabled Agent Skills
- **WHEN** main Chat or side-panel Chat asks for natural-language Agent work
- **THEN** Cabinet SHALL call the assistant-provider runtime through the provider-neutral turn interface
- **AND** Cabinet SHALL provide only enabled and available skill names, descriptions, JSON schemas, safety metadata, and canonical Agent context envelope fields needed for planning
- **AND** provider output SHALL be parsed as structured skill-selection or clarification input to Cabinet's governed planner
- **AND** provider output SHALL NOT be treated as permission to mutate Cabinet state or as direct access to Cabinet skills, database, filesystem, app-control commands, or external-write actions

#### Scenario: Execute planned work through governed skill boundaries
- **GIVEN** the governed planner receives a structured provider selection for an Agent Skill
- **WHEN** the selected work is read-only, local-write, external-write, destructive, disabled, unavailable, ambiguous, or unsupported
- **THEN** Cabinet SHALL evaluate the same profile authority policy, validation, preview, confirmation, idempotency, and audit boundaries used by direct Agent Skill execution
- **AND** read-only work MAY return grounded Cabinet results when allowed by policy
- **AND** local-write work SHALL create a preview and SHALL NOT apply until the user explicitly confirms through the existing apply endpoint
- **AND** external-write and destructive work SHALL NOT bypass per-action confirmation or strong target/impact confirmation
- **AND** disabled, unavailable, ambiguous, unsupported, stale-preview, provider-failure, or tool-failure states SHALL return truthful redacted recovery guidance
