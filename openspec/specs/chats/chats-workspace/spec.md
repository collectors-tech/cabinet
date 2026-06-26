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
- **AND** New Chat / Start a conversation affordances MUST keep the new-thread entry available without changing the active route
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
- **AND** action-preview and attachment controls MUST remain visually secondary to the chat composer

### Requirement CHATS-WORKSPACE-008: `/chats` SHALL dispatch normal user text through governed app-control planning
Chats workspace MUST send normal conversation text with profile, thread, route, and assistant context so deterministic app-control requests can return governed route or action results instead of defaulting to Inbox handoff.

#### Scenario: Dispatch route-opening chat text without creating Inbox noise
- **GIVEN** user opens `/chats` with an active profile and selected thread
- **WHEN** user sends `open media`
- **THEN** the chat message API response MUST include a `navigate.open_surface` app-control result for `/media`
- **AND** the route-opening action MUST create durable workflow-run audit evidence for the selected thread
- **AND** the response MUST NOT create a default assistant Inbox handoff for the handled app-control request
