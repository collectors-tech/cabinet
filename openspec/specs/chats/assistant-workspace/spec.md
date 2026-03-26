## Purpose
Define Cabinet Assistant as a persistent in-app workspace for context-aware AI help, thread continuity, and governed action assistance.

## Requirements
### Requirement ASSISTANT-WORKSPACE-001: Assistant SHALL preserve thread context across route and workspace changes
Assistant MUST preserve active thread state while users move through authenticated routes and switch between Navigation, Assistant, and Inbox workspaces.

#### Scenario: Return to prior assistant thread after route change
- **GIVEN** user has active assistant thread on `/inventory`
- **WHEN** user navigates to `/wishlist` and later returns to Assistant workspace
- **THEN** prior thread/messages MUST still be present unless an explicit reset boundary occurred

### Requirement ASSISTANT-WORKSPACE-002: Assistant SHALL receive deterministic route and selection context
Assistant message requests MUST include normalized route/profile/selection context so the assistant behaves like an in-app helper instead of a detached chatbot.

#### Scenario: Send message with current route context
- **GIVEN** user is on `/inventory` with active profile and selected collection/item context
- **WHEN** user submits an assistant message
- **THEN** runtime MUST attach deterministic route/profile/selection context to the assistant request envelope

### Requirement ASSISTANT-WORKSPACE-003: Assistant SHALL support provider and model selection with explicit thread semantics
Assistant workspace MUST expose provider/model selection and define what happens when the user changes provider/model relative to thread continuity.

#### Scenario: Change provider in assistant workspace
- **GIVEN** active assistant thread exists with provider `openai`
- **WHEN** user changes provider or model
- **THEN** UI MUST show deterministic thread behavior (for example new thread, forked thread, or annotated continuation)
- **AND** resulting thread metadata MUST record provider/model context

### Requirement ASSISTANT-WORKSPACE-004: Assistant SHALL define explicit reset boundaries
Assistant context MUST only reset on explicit boundaries such as logout, active database/profile switch, manual clear/new thread, or provider/session invalidation.

#### Scenario: Reset on active profile change
- **GIVEN** user has active assistant thread in `Primary DB`
- **WHEN** user switches to another active database/profile
- **THEN** assistant workspace MUST apply the configured reset/isolation policy and MUST NOT leak prior profile thread state across profiles
