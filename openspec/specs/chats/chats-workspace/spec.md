## Purpose
Define the dedicated `/chats` workspace as a real conversation surface distinct from Assistant and Inbox.

## Requirements
### Requirement CHATS-WORKSPACE-001: `/chats` SHALL present a real Cabinet conversation workspace rather than placeholder/template inbox content
Chats workspace MUST use Cabinet-specific purpose, copy, and thread semantics instead of generic stock inbox template content.

#### Scenario: Render chats workspace with domain-specific semantics
- **GIVEN** user opens `/chats`
- **WHEN** route renders
- **THEN** headings, empty-state copy, and thread list content MUST reflect Cabinet conversation workflows
- **AND** route MUST NOT render generic placeholder names/copy as if the feature were a stock template inbox

### Requirement CHATS-WORKSPACE-002: Chats send flow SHALL preserve active thread context after send
Sending a message in `/chats` MUST append to the active thread and keep the selected conversation context stable.

#### Scenario: Send message in selected thread
- **GIVEN** user selected a conversation on `/chats`
- **WHEN** user sends a message
- **THEN** outgoing message MUST appear in the active thread
- **AND** route MUST remain on the selected conversation state instead of collapsing to a generic landing/empty state

### Requirement CHATS-WORKSPACE-003: Cabinet SHALL define the boundary between Assistant and `/chats`
Cabinet MUST define the distinct responsibilities of shell Assistant workspace and `/chats` so they are related but non-duplicative.

#### Scenario: Compare Assistant and `/chats` responsibilities
- **GIVEN** shell Assistant workspace and `/chats` route both exist
- **WHEN** product contracts are inspected
- **THEN** Assistant MUST be documented as the AI helper workspace
- **AND** `/chats` MUST be documented as the intentional conversation/thread workspace
- **AND** overlap, if any, MUST be explicitly justified by product behavior

### Requirement CHATS-WORKSPACE-004: `/chats` SHALL preserve the original/example two-pane conversation layout
Chats workspace MUST render as a dark, sparse two-pane conversation surface with a compact conversation rail on the left and the active or empty conversation workspace on the right.

#### Scenario: Render original/example chats layout parity
- **GIVEN** user opens `/chats`
- **WHEN** no conversation is selected
- **THEN** the left rail MUST include a search input before the conversation list
- **AND** conversation rows MUST preserve compact avatar, participant/title, and message preview structure
- **AND** the right workspace MUST show a centered empty-state composition with icon, title, helper text, and primary action affordance
- **AND** the rail and workspace MUST remain separate, readable, and unclipped at normal desktop widths

### Requirement CHATS-WORKSPACE-005: `/chats` SHALL support thread discovery and new-thread entry from the two-pane rail
Chats workspace MUST let users start a new durable thread from both the empty workspace and selected-thread topbar while keeping thread search scoped to discoverable conversation rows.

#### Scenario: Create and filter conversation threads
- **GIVEN** user opens `/chats`
- **WHEN** user creates multiple chat threads and filters the conversation rail
- **THEN** matching thread rows MUST remain visible
- **AND** non-matching thread rows MUST be hidden
- **AND** New Chat / Start a conversation affordances MUST open a labeled new-thread dialog at desktop and compact widths without changing the active route
- **AND** a successfully created thread MUST close the dialog and become the active conversation without requiring access to the conversation rail
- **AND** created threads MUST persist through the Cabinet chat thread API for the active profile

### Requirement CHATS-WORKSPACE-006: `/chats` SHALL closely follow the assistant-ui example chat surface
Chats workspace MUST render close to the provided assistant-ui examples: a dark, compact thread rail, a broad uninterrupted conversation canvas, and a bottom-docked centered composer instead of stacked framed panels.

#### Scenario: Render assistant-ui example visual contract
- **GIVEN** user opens `/chats` and selects a thread
- **WHEN** the selected conversation workspace renders
- **THEN** the layout MUST identify the assistant-ui example visual contract
- **AND** the conversation rail MUST stay compact beside the main surface
- **AND** the message canvas MUST occupy the dominant vertical workspace
- **AND** the composer MUST be docked at the bottom center of the conversation surface
- **AND** the composer MUST NOT be followed by separate custom Attachments or Action Preview panels
- **AND** attachment affordances MUST live with the composer while action previews, when present, MUST appear in the conversation canvas instead of as stacked controls under the composer

### Requirement CHATS-WORKSPACE-008: `/chats` SHALL dispatch normal user text through governed app-control planning
Chats workspace MUST send normal conversation text with profile, thread, route, and assistant context so deterministic app-control requests can return governed route or action results instead of defaulting to Inbox handoff.

#### Scenario: Dispatch route-opening chat text without creating Inbox noise
- **GIVEN** user opens `/chats` with an active profile and selected thread
- **WHEN** user sends `open media`
- **THEN** the chat message API response MUST include a `navigate.open_surface` app-control result for `/media`
- **AND** the route-opening action MUST create durable workflow-run audit evidence for the selected thread
- **AND** the response MUST NOT create a default assistant Inbox handoff for the handled app-control request

### Requirement CHATS-WORKSPACE-009: `/chats` SHALL preview normal-text mutations before apply
Chats workspace MUST let deterministic app-control planning convert normal user text into preview-required mutation actions without mutating records before explicit confirmation.

#### Scenario: Dispatch item-create chat text as a pending preview
- **GIVEN** user opens `/chats` with an active profile and selected thread
- **WHEN** user sends `create an inventory item ...`
- **THEN** the chat message API response MUST include an `inventory.item.create` app-control result with a `create_inventory_item` preview
- **AND** the preview MUST include the parsed item payload and pending confirmation workflow-run audit evidence
- **AND** inventory MUST NOT include the item before the user explicitly applies the preview
- **AND** the response MUST NOT create a default assistant Inbox handoff for the handled app-control request

### Requirement CHATS-WORKSPACE-010: `/chats` SHALL render governed app-control route and setup-needed results
Chats workspace MUST expose deterministic app-control outcomes from normal chat messages as visible, user-activated UI instead of relying only on API metadata.

#### Scenario: Activate a route-opening result and expose provider setup-needed guidance
- **GIVEN** user opens `/chats` with an active profile and selected thread
- **WHEN** user sends `open media`
- **THEN** the workspace MUST render a route action card for `/media`
- **AND** Cabinet MUST stay on `/chats` until the user activates the card
- **AND** activating the card MUST navigate to `/media`
- **WHEN** user sends a provider-backed request that cannot run without provider readiness
- **THEN** the workspace MUST show visible setup-needed guidance instead of pretending success

### Requirement CHATS-WORKSPACE-011: Chat messages SHALL answer normal text without default Inbox handoff noise
Chats workspace and shell Assistant messages MUST produce a direct assistant response for ordinary non-action text and MUST reserve durable Inbox handoffs for background, review-required, queued, or failure work.

#### Scenario: Respond to greeting without durable Inbox item
- **GIVEN** user sends `hello` with profile, thread, route, and assistant context
- **WHEN** `/api/chat/messages` accepts the message
- **THEN** the response MUST include a direct `assistant_response` thread message
- **AND** the assistant thread MUST show the direct response instead of `Assistant handoff queued in Inbox.`
- **AND** `/api/chat/inbox` MUST NOT receive an `assistant_handoff` item for that normal message

#### Scenario: Preserve explicit handoff durability
- **GIVEN** user sends a message that asks Cabinet to follow up, queue, review, notify, or run background work
- **WHEN** `/api/chat/messages` accepts the message
- **THEN** Cabinet MAY create an `assistant_handoff` Inbox item with thread metadata
- **AND** Inbox triage status lifecycle MUST remain durable for that queued handoff

### Requirement CHATS-WORKSPACE-012: Full Chat Agent controls SHALL avoid compact-layout overflow traps
Full Chat MUST keep its message result, composer, attachment, Retry, Apply, and Cancel controls keyboard reachable through intended scrolling at compact desktop and 200-percent-zoom-equivalent layouts.

#### Scenario: Operate a governed result at 200-percent-zoom equivalent
- **GIVEN** full Chat is rendered at a 640 by 360 viewport with a selected thread
- **WHEN** a governed Agent result exposes retry or preview actions
- **THEN** the result and composer regions MUST use bounded independent scrolling without fixed minimum-height or fixed-width overflow
- **AND** attachment, Retry, Apply, and Cancel controls MUST remain visible and keyboard focusable when applicable
- **AND** focus MUST remain on the initiating action after a terminal dialog is canceled

### Requirement CHATS-WORKSPACE-013: Public Chat ingress SHALL reject client-authored trusted Agent evidence
The public `/api/chat/messages` endpoint MUST accept only user-authored messages and MUST fail closed when request context carries trusted Agent result, preview, execution, authority, capability, or success evidence. Server-side planner, provider, dispatcher, preview, apply, cancel, and audit paths MAY still persist assistant/system evidence through internal service calls.

#### Scenario: Reject forged public Chat evidence
- **GIVEN** a client calls `POST /api/chat/messages`
- **WHEN** the request uses `role=assistant` or `role=system`
- **THEN** Cabinet MUST reject the request before storing a chat message
- **WHEN** a `role=user` request carries `agent_planner`, `agent_capabilities`, preview, execution, authority, assistant response, assistant handoff, admin-session, or success evidence in public context fields
- **THEN** Cabinet MUST reject the request before storing a chat message
- **AND** a normal user message MAY still trigger server-owned planner/provider/dispatcher persistence for trusted assistant evidence
