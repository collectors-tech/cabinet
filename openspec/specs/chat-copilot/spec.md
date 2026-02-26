## Purpose
Define in-app chat copilot behavior, context, and safety model.

## Requirements
### Requirement: Chat copilot SHALL be persistent and globally accessible in workspace
Cabinet SHALL provide right-rail chat panel with open/close controls from header.

#### Scenario: Toggle chat rail
- **WHEN** user activates open/close chat control
- **THEN** chat rail SHALL toggle visibility without losing current route context

### Requirement: Chat threads and messages SHALL be stored locally per profile
Cabinet SHALL persist thread/message history with profile isolation.

#### Scenario: Reopen thread history
- **WHEN** user reopens existing thread
- **THEN** prior messages SHALL load from profile-local storage

### Requirement: Chat SHALL support user-selected file attachments
Cabinet SHALL allow attaching local files selected by user for chat context.

#### Scenario: Attach local file
- **WHEN** user uploads a local attachment in chat
- **THEN** attachment SHALL be stored and linked to chat thread context

### Requirement: Chat-driven mutations SHALL require preview and confirmation
Cabinet SHALL require preview and explicit confirm for mutation actions.

#### Scenario: Apply proposed mutation
- **WHEN** chat proposes create/update/track/wishlist action
- **THEN** action SHALL not execute until user confirms apply
