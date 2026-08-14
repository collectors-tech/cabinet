## Purpose
Define Cabinet Agent as the persistent, governed conversational control surface for context-aware help and app actions across the compact shell and full Chat workspace.

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

#### Scenario: Compact Agent panel uses icon-only chrome controls
- **GIVEN** issue #1438 requires the shell Assistant panel to match the compact workspace chrome direction
- **WHEN** the shell Assistant side-panel renders
- **THEN** the panel header MUST use the visible title `Cabinet Agent`
- **AND** header actions for new thread, mute/quiet state, and close MUST be icon-only controls with accessible names/tooltips
- **AND** the panel MUST expose a compact assistant identity/status card, conversation selector, readable dark message area, fixed bottom composer with icon-only send, and a collapsed `Action Timeline` disclosure row
- **AND** the panel MUST preserve Cabinet chat thread persistence, route context, provider/model metadata, and governed preview/confirm/apply action boundaries

### Requirement ASSISTANT-WORKSPACE-007: Assistant SHALL expose governed action execution results with persistence proof
Assistant workspace capability/action cards MUST make preview, confirmation, execution state, and applied result links visible in the compact side-panel while preserving Cabinet's preview-before-apply mutation boundary.

#### Scenario: Preview and apply a governed action from the Assistant workspace
- **GIVEN** the shell Assistant workspace has an active profile-scoped Cabinet chat thread
- **WHEN** the user previews a governed item-creation action from the action card
- **THEN** the preview card MUST show the pending action payload
- **AND** the target item MUST NOT exist in inventory before explicit confirmation
- **WHEN** the user confirms the apply dialog
- **THEN** the execution state MUST report success
- **AND** the result card MUST expose an item result link
- **AND** a refreshed inventory data query MUST include the applied item under the active profile
- **AND** the active chat thread MUST persist an assistant/action audit message that records the confirmed mutation outcome

### Requirement ASSISTANT-WORKSPACE-008: Assistant SHALL keep guided walkthroughs open across main-route navigation
The shell Assistant side-panel MUST remain open and connected to the active guided walkthrough while Chat-driven app-control commands navigate the main Cabinet route, highlight targets, or wait for user confirmation.

#### Scenario: Preserve side-panel walkthrough during navigation
- **GIVEN** a guided walkthrough is active in the shell Assistant side-panel
- **WHEN** the shell command bus navigates the main app from one authenticated route to another
- **THEN** the Assistant side-panel MUST stay open with the same active thread and workflow run selected
- **AND** the compact Action Timeline MUST show the navigation step, target highlight step, and next required action
- **AND** route changes MUST NOT clear pending preview, confirmation, recipe, or target state for the active profile/thread
- **AND** closing the side-panel MUST not apply or cancel a mutating step unless the user explicitly chooses that action

### Requirement ASSISTANT-WORKSPACE-009: Assistant side-panel SHALL dispatch normal user text through governed app-control planning
The shell Assistant side-panel MUST send normal user text with route, profile, thread, selection, and assistant context so deterministic app-control requests can return governed route or action results instead of defaulting to an Inbox handoff.

#### Scenario: Dispatch a route-opening request from the Assistant side-panel
- **GIVEN** the shell Assistant side-panel is open on an authenticated route
- **WHEN** user sends `open media`
- **THEN** the chat message API response MUST include a `navigate.open_surface` app-control result for `/media`
- **AND** the response MUST include workflow-run audit evidence for the governed dispatch
- **AND** the side-panel MUST render a route action card without navigating the page until the user chooses the action
- **AND** the handled app-control request MUST NOT create a default assistant Inbox handoff

### Requirement ASSISTANT-WORKSPACE-010: Assistant side-panel SHALL render setup-needed app-control guidance
The shell Assistant side-panel MUST turn provider-backed app-control setup-needed results into visible guidance so users can tell the requested assistant action is blocked by provider readiness.

#### Scenario: Provider-backed request shows setup-needed guidance
- **GIVEN** the shell Assistant side-panel is open on an authenticated route
- **WHEN** user sends a provider-backed request that cannot run without provider readiness
- **THEN** the chat message API response MUST include `setup_needed=true`
- **AND** the side-panel MUST render visible provider setup-needed guidance
- **AND** the side-panel MUST NOT render a route action card for the unavailable provider-backed request

### Requirement ASSISTANT-WORKSPACE-012: Assistant side-panel SHALL dispatch governed Agent Skills from conversation context
The shell Assistant side-panel MUST support governed skill execution for Integrations, Market Watch, and Purchases from natural-language conversation without losing the active profile, thread, source channel, source surface, or preview-before-apply confirmation boundary. The customer-facing Chat surface MUST NOT require a raw skill selector or client-carried provider, setup-step, secret, or mutation parameter form.

#### Scenario: Dispatch Market Watch skill from the Assistant side-panel
- **GIVEN** the shell Assistant side-panel is open with an active profile-scoped thread
- **WHEN** the user asks Cabinet to run a saved watch and reviews the server-owned preview before confirming apply
- **THEN** the preview request MUST include the active profile, thread, `source_channel=in-app`, `source_surface=market_watch.saved_watch.row`, provider ID, and saved watch ID
- **AND** the apply request MUST repeat that same source context with `confirm=true`
- **AND** the side-panel MUST show preview and result state without applying an unconfirmed mutation

#### Scenario: Dispatch Purchases skill from the Assistant side-panel
- **GIVEN** the shell Assistant side-panel is open with an active profile-scoped thread
- **WHEN** the user asks Cabinet to create a purchase order and reviews the server-owned preview before confirming apply
- **THEN** the preview request MUST include the active profile, thread, `source_channel=in-app`, `source_surface=purchases.inbox.capture`, purchase source, item ID, and source URL
- **AND** the apply request MUST repeat that same source context with `confirm=true`
- **AND** the side-panel MUST show preview and result state without applying an unconfirmed mutation

### Requirement ASSISTANT-WORKSPACE-016: Cabinet Agent SHALL preserve one governed conversation across compact and full workspaces
Cabinet MUST present one `Cabinet Agent` identity in the shell and `/chats`, preserve the active profile-scoped thread and pending review state when moving between those surfaces, and render registry-derived capability and planner states without exposing arbitrary provider payloads or secret-bearing parameters.

#### Scenario: Expand a contextual Agent thread into the full workspace
- **GIVEN** the contextual Cabinet Agent has an active thread with messages or a pending action preview
- **WHEN** the user activates `Open this thread in full Cabinet Agent` by pointer or keyboard
- **THEN** `/chats` MUST select that exact profile-scoped thread
- **AND** its messages and pending preview MUST remain reviewable
- **AND** reopening the contextual Agent MUST return to the same thread without creating a duplicate conversation or dispatching an action

#### Scenario: Explain current Agent capabilities consistently
- **GIVEN** Cabinet has an active profile authority policy and Agent Skill registry
- **WHEN** the user asks what Cabinet Agent can do from either Agent surface
- **THEN** the UI MUST show bounded counts for available, confirmation-required, setup-required, and policy-blocked skills
- **AND** the same persisted capability explanation MUST remain visible after expanding into `/chats`
- **AND** the UI MUST link to Agent Skill settings without treating registry presence as proof that a skill is executable

#### Scenario: Keep compact Agent controls reachable
- **GIVEN** the Cabinet Agent is open at a compact desktop viewport
- **WHEN** the identity and context rail plus a structured result exceed the available height
- **THEN** the identity rail and conversation area MUST remain independently scrollable
- **AND** keyboard launch and expand controls MUST remain operable with native `Enter` or `Space`

#### Scenario: Show a recoverable planner failure truthfully
- **GIVEN** the selected assistant provider cannot return a valid governed skill plan
- **WHEN** Cabinet records a recoverable planner failure
- **THEN** compact and full Agent surfaces MUST show the failure and deterministic next action
- **AND** they MUST NOT render a pending preview, applied result, or false success state
- **AND** the explanation MUST retain Cabinet as authority and dispatch owner

### Requirement ASSISTANT-WORKSPACE-016: Assistant thread bootstrap SHALL remain lint-clean without stale context
The shell Assistant side-panel MUST keep React Hook dependency semantics explicit so route context, selected record context, provider/model defaults, and thread bootstrap logic remain stable without duplicate thread creation or stale provider/model state.

#### Scenario: Bootstrap uses stable assistant thread callbacks
- **GIVEN** the shell Assistant side-panel opens for an active profile with stored or default provider/model settings
- **WHEN** React effects bootstrap the assistant thread and load messages
- **THEN** the bootstrap effect MUST use stable callback dependencies
- **AND** it MUST NOT create duplicate assistant workspace threads for the same active profile/defaults
- **AND** loaded thread metadata MUST continue to set the visible provider and model.

#### Scenario: Default provider sync observes provider/model metadata explicitly
- **GIVEN** an active assistant workspace thread uses `assistant_workspace_session` semantics
- **WHEN** integration-level assistant defaults change from the stored provider/model
- **THEN** the side-panel default sync MUST compare the current visible provider/model and the thread metadata provider/model explicitly
- **AND** it MUST update local storage and visible provider/model only for assistant workspace session threads
- **AND** route/search/message re-renders MUST NOT cause unrelated context recomputation or thread reset.

### Requirement ASSISTANT-WORKSPACE-017: Assistant SHALL render the latest server-owned Agent response state
The compact Assistant workspace SHALL render the same normalized Agent response contract as full Chat and SHALL NOT recover a stale planner, capability, navigation, or preview card by scanning backward past a newer assistant message.

#### Scenario: Replace stale structured state with the latest response
- **GIVEN** an active thread contains an older Agent preview or capability response
- **WHEN** a newer assistant message has no Agent response state
- **THEN** the compact workspace MUST remove the older structured Agent card
- **AND** refresh or thread switching MUST retain only the latest server-owned state for that exact profile and thread

#### Scenario: Failure never presents as success
- **GIVEN** the server returns setup, authority, unsupported, provider, expiry, stale-target, cancelled, or failed state
- **WHEN** the compact Agent response card renders
- **THEN** it MUST NOT show Apply, Completed, Applied, or other success language
- **AND** Retry MUST appear only when the server state is retryable and includes the bounded original intent

### Requirement ASSISTANT-WORKSPACE-017: Customer Chat SHALL be conversation-first
Cabinet's compact and full Chat surfaces MUST expose one natural-language composer plus structured server-owned response cards. Manual inventory preview fields and raw Agent Skill/provider/setup/secret playground controls MUST NOT render in production. Configuration and credential entry MUST remain in the governed Settings or Integrations destination.

#### Scenario: Production compact Chat keeps the composer actionable
- **GIVEN** Cabinet is running without E2E hooks at a compact desktop viewport
- **WHEN** the user opens Chat from an authenticated product route
- **THEN** the composer and Action Timeline MUST remain reachable without forced interaction
- **AND** the manual part-number/title preview form MUST NOT render
- **AND** the raw Agent Skill/provider/setup/secret form MUST NOT render
- **AND** the Chat rail MUST use bounded independent scrolling rather than displacing the composer

#### Scenario: Natural-language mutation uses a server-owned confirmation card
- **GIVEN** an active profile-scoped Chat thread
- **WHEN** the user asks Cabinet to perform a governed local mutation
- **THEN** Cabinet MUST return a structured server-owned preview bound to the profile, thread, source, and target
- **AND** the UI MUST offer explicit Apply and Cancel controls only for that trusted preview
- **AND** Apply MUST execute at most once while Cancel MUST execute no mutation
- **AND** raw skill identifiers, credential values, and client-carried mutation parameters MUST NOT be requested in the primary conversation

#### Scenario: Setup-required response links to the owning product surface
- **GIVEN** an Agent request cannot proceed because a provider, credential, or permission is missing
- **WHEN** Cabinet renders the setup-required response
- **THEN** the response MUST derive an allowlisted destination from structured server context and link to the exact Settings or Integrations surface that owns setup
- **AND** arbitrary next-action text MUST NOT become a link
- **AND** the response MUST NOT render or echo a secret value

### Requirement ASSISTANT-WORKSPACE-018: Public Chat ingress SHALL reject client-authored Agent success evidence
Public `POST /api/chat/messages` requests MUST accept only user-authored messages and MUST NOT persist client-supplied planner, capability, app-control, handoff, assistant-response, provider, execution, mutation, preview, or workflow evidence as trusted Agent output. Assistant and system messages with trusted evidence MUST be persisted only by Cabinet's in-process planner, provider, or dispatcher after governed execution.

#### Scenario: Reject client-authored assistant and planner evidence
- **GIVEN** a public client can submit a message to an existing profile-scoped thread
- **WHEN** it submits role `assistant` or `system`, or adds server-owned Agent evidence to a user message
- **THEN** Cabinet MUST reject the request before persisting any message or evidence
- **AND** unknown evidence nested in the client Agent context MUST be removed by the normalized Agent envelope
- **AND** an authenticated owner session MUST NOT bypass the public authorship boundary

#### Scenario: Persist trusted non-live E2E planner evidence from a user request
- **GIVEN** explicit E2E hooks register Cabinet's embedded synthetic Agent provider
- **WHEN** a role `user` message requests a governed local change
- **THEN** the production planner and dispatcher MUST create the preview and assistant message
- **AND** persisted provider provenance MUST state `network=disabled`, `test_provider=true`, and `live_provider=false`
- **AND** Apply or Cancel MUST consume only the server-created opaque preview identity
- **AND** the synthetic provider MUST NOT be registered in a normal non-E2E runtime

### Requirement ASSISTANT-WORKSPACE-019: Agent attachments SHALL persist across compact and full surfaces without crossing scope
Cabinet MUST represent persisted message attachments as safe structured metadata and render them from server state in both compact Assistant and full Chats.

#### Scenario: Attachment survives compact/full handoff and reload exactly once
- **GIVEN** a user attaches a file to an active profile-scoped thread in the compact Assistant
- **WHEN** the message is sent, opened as that exact thread in full Chats, reloaded, and reopened in the compact Assistant
- **THEN** both surfaces MUST render the same persisted attachment exactly once
- **AND** the rendered metadata MUST contain only its safe name, media type, size, source, and opaque attachment identity
- **AND** the handoff and reload MUST NOT create a replacement thread or duplicate attachment binding

#### Scenario: Attachment binding and staged state fail closed across scope changes
- **GIVEN** a staged or persisted attachment belongs to one profile-scoped thread
- **WHEN** the user changes thread, profile, provider, or model, or attempts to bind that attachment to another scoped thread
- **THEN** Cabinet MUST clear staged attachment state before the new context is used
- **AND** the API MUST reject cross-profile, cross-thread, missing, or duplicate attachment identities
- **AND** the rejected request MUST NOT persist a partial message or attachment binding

#### Scenario: Telegram media uses the canonical message attachment model
- **GIVEN** an allowed Telegram update contains supported media for a mapped Cabinet profile
- **WHEN** Cabinet captures the update into the mapped Agent thread
- **THEN** Cabinet MUST persist the media in its governed attachment store and bind it to the canonical Chat message
- **AND** the binding MUST carry safe `telegram_media` attachment provenance and `telegram` source metadata
- **AND** the persisted message MUST remain isolated to its mapped profile and thread after restart or reload

### Requirement ASSISTANT-WORKSPACE-020: Agent controls SHALL remain keyboard reachable at compact and zoom-equivalent layouts
The shell Agent workspace MUST keep its native launcher, composer, attachments, results, and guarded action controls reachable without forced events at a 640 by 360 compact desktop or 200-percent-zoom-equivalent viewport, and MUST provide a thread-preserving handoff to full Chat.

#### Scenario: Use contextual Agent and continue in full Chat with the keyboard
- **GIVEN** an authenticated Cabinet route is rendered at a 640 by 360 viewport
- **WHEN** the user activates the native Agent launcher with Enter, reaches contextual composer and attachment controls, and activates the full-Chat handoff with Space
- **THEN** every operated control MUST be visible within the viewport or its intended scroll container
- **AND** full Chat MUST open the same profile-scoped thread
- **AND** keyboard focus MUST move to the full Chat composer
- **AND** the flow MUST NOT require forced events, hidden controls, or test-only production behavior
