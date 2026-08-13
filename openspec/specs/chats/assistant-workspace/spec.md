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

#### Scenario: Compact Assistant panel uses icon-only chrome controls
- **GIVEN** issue #1438 requires the shell Assistant panel to match the compact workspace chrome direction
- **WHEN** the shell Assistant side-panel renders
- **THEN** the panel header MUST use the visible title `Chat`
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

### Requirement ASSISTANT-WORKSPACE-012: Assistant side-panel SHALL dispatch governed Agent Skills with source context
The shell Assistant side-panel Agent Skill card MUST support governed skill execution for Integrations, Market Watch, and Purchases without losing the active profile, thread, source channel, source surface, or preview-before-apply confirmation boundary.

#### Scenario: Dispatch Market Watch skill from the Assistant side-panel
- **GIVEN** the shell Assistant side-panel is open with an active profile-scoped thread
- **WHEN** the user selects `cabinet.market_watch.run_watch`, enters provider and saved-watch context, previews the skill, and confirms apply
- **THEN** the preview request MUST include the active profile, thread, `source_channel=in-app`, `source_surface=market_watch.saved_watch.row`, provider ID, and saved watch ID
- **AND** the apply request MUST repeat that same source context with `confirm=true`
- **AND** the side-panel MUST show preview and result state without applying an unconfirmed mutation

#### Scenario: Dispatch Purchases skill from the Assistant side-panel
- **GIVEN** the shell Assistant side-panel is open with an active profile-scoped thread
- **WHEN** the user selects `cabinet.purchases.create_order`, enters purchase source, item, and source URL context, previews the skill, and confirms apply
- **THEN** the preview request MUST include the active profile, thread, `source_channel=in-app`, `source_surface=purchases.inbox.capture`, purchase source, item ID, and source URL
- **AND** the apply request MUST repeat that same source context with `confirm=true`
- **AND** the side-panel MUST show preview and result state without applying an unconfirmed mutation

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

### Requirement ASSISTANT-WORKSPACE-017: Agent controls SHALL remain keyboard reachable at compact and zoom-equivalent layouts
The shell Agent workspace MUST keep its native launcher, composer, attachments, results, and guarded action controls reachable without forced events at a 640 by 360 compact desktop or 200-percent-zoom-equivalent viewport, and MUST provide a thread-preserving handoff to full Chat.

#### Scenario: Use contextual Agent and continue in full Chat with the keyboard
- **GIVEN** an authenticated Cabinet route is rendered at a 640 by 360 viewport
- **WHEN** the user activates the native Agent launcher with Enter, reaches contextual composer and attachment controls, and activates the full-Chat handoff with Space
- **THEN** every operated control MUST be visible within the viewport or its intended scroll container
- **AND** full Chat MUST open the same profile-scoped thread
- **AND** keyboard focus MUST move to the full Chat composer
- **AND** the flow MUST NOT require forced events, hidden controls, or test-only production behavior
