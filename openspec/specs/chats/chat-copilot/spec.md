## Purpose
Define in-app chat copilot behavior, context, and safety model.

## Requirements
### Requirement CHAT-COPILOT-001: Chat copilot SHALL be persistent and globally accessible in workspace
Cabinet SHALL provide right-rail chat panel with open/close controls from header.

#### Scenario: Toggle chat rail
- **GIVEN** authenticated user is on any workspace route and global shell header is rendered
- **WHEN** user activates open/close chat control
- **THEN** chat rail SHALL toggle visibility without losing current route context and route URL/query state MUST remain unchanged

### Requirement CHAT-COPILOT-002: Chat threads and messages SHALL be stored locally per profile
Cabinet SHALL persist thread/message history with profile isolation.

#### Scenario: Reopen thread history
- **GIVEN** active profile has existing thread records in local chat storage
- **WHEN** user reopens existing thread
- **THEN** prior messages SHALL load from profile-local storage in chronological order with preserved author role labels

### Requirement CHAT-COPILOT-003: Chat SHALL support user-selected file attachments
Cabinet SHALL allow attaching local files selected by user for chat context.

#### Scenario: Attach local file
- **GIVEN** chat thread is open and user selects a local file via attachment control
- **WHEN** user uploads a local attachment in chat
- **THEN** runtime MUST persist attachment via chat attachment API and link it to the current thread with attachment metadata (`name`, `size_bytes`, `mime_type`)

### Requirement CHAT-COPILOT-004: Chat-driven mutations SHALL require preview and confirmation
Cabinet SHALL require preview and explicit confirm for mutation actions.

#### Scenario: Apply proposed mutation
- **GIVEN** chat action preview exists and no explicit user confirmation has been submitted
- **WHEN** chat proposes create/update/track/wishlist action
- **THEN** action SHALL not execute until user confirms apply
