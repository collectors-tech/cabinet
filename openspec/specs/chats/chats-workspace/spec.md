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
