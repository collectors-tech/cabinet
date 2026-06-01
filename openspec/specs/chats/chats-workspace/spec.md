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
