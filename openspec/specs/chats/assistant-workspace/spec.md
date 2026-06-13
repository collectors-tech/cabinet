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

### Requirement ASSISTANT-WORKSPACE-005: Assistant SHALL adopt assistant-ui through a Cabinet-owned transport adapter
Assistant workspace MUST compare direct assistant-ui AI SDK runtime adoption against a Cabinet-specific assistant-ui adapter, and the first implementation slice MUST adapt assistant-ui's compact anchored modal primitives into the existing shell Assistant sidepanel while keeping Cabinet Go chat APIs as the source of truth for profile-scoped threads, persisted messages, route context, provider/model metadata, attachment context, and confirm-before-apply mutation safety.

The 2026-06-13 Cabinet product direction further requires the side-panel Assistant UI to use assistant-ui.com / `https://www.assistant-ui.com/examples/ai-sdk` as the primary UI reference alongside the main chat UI. Future Assistant workspace work MUST bind to `UI-SCREEN-CHAT-COPILOT-018` when it changes side-panel composer, message, thread/status, or tool/result interaction behavior.

#### Scenario: Render assistant-ui-compatible shell primitives without replacing Cabinet persistence
- **GIVEN** the shell Assistant workspace has an active profile-scoped Cabinet chat thread
- **WHEN** the anchored modal frame, message list, and composer render through the Cabinet assistant-ui adapter
- **THEN** the user MUST still be able to send a message through `/api/chat/messages`
- **AND** the request MUST preserve route/profile/selection/provider/model context
- **AND** the rendered adapter surface MUST expose assistant-ui anchor/trigger/content, message, and composer primitives without moving persistence or provider secrets into a Next.js/Vercel AI SDK sample backend
- **AND** action preview/apply controls MUST remain explicit Cabinet controls outside automatic chat tool-call mutation

#### Scenario: Choose the first assistant-ui implementation slice
- **GIVEN** direct AI SDK runtime adoption would require moving persistence/provider flow into the sample backend
- **WHEN** Cabinet evaluates the assistant-ui adoption path
- **THEN** the first implementation slice MUST target the shell Assistant workspace panel before broad `/chats` replacement
- **AND** direct AI SDK runtime adoption MUST remain a later option only if Cabinet-owned transport, persistence, and confirmation boundaries are preserved

#### Scenario: Future Assistant side-panel changes name the assistant-ui direction
- **GIVEN** a future issue, spec, or PR changes the shell Assistant workspace or compact Assistant panel
- **WHEN** the change affects side-panel composer, message, thread/status, or tool/result interaction behavior
- **THEN** it MUST state how the side-panel Assistant UI follows the assistant-ui AI SDK example
- **AND** any divergence MUST link to a Cabinet issue/spec that records the reason, affected surface, and validation expectation

#### Scenario: Compact side-panel aligns with the dark chat shell language
- **GIVEN** `/chats` uses the assistant-ui-inspired dark shell from `UI-SCREEN-CHAT-COPILOT-019`
- **WHEN** the shell Assistant side-panel renders in compact form
- **THEN** it MUST reuse the same Cabinet assistant visual language for header, thread controls, composer, prompt/action affordances, and governed tool/action cards without crowding the current page context
- **AND** compact visual alignment MUST NOT weaken persisted thread/message ownership, provider/model readiness, route context, or explicit preview/confirm/apply boundaries
