## Purpose
Define in-app chat copilot behavior, context, and safety model.

## Requirements
### Requirement CHAT-COPILOT-001: Chat copilot SHALL be persistent and globally accessible in workspace
Cabinet SHALL provide right-rail chat panel with open/close controls from header.

#### Scenario: Toggle chat rail
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** user activates open/close chat control
- **THEN** chat rail SHALL toggle visibility without losing current route context

### Requirement CHAT-COPILOT-002: Chat threads and messages SHALL be stored locally per profile
Cabinet SHALL persist thread/message history with profile isolation.

#### Scenario: Reopen thread history
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** user reopens existing thread
- **THEN** prior messages SHALL load from profile-local storage

### Requirement CHAT-COPILOT-003: Chat SHALL support user-selected file attachments
Cabinet SHALL allow attaching local files selected by user for chat context.

#### Scenario: Attach local file
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** user uploads a local attachment in chat
- **THEN** attachment SHALL be stored and linked to chat thread context

### Requirement CHAT-COPILOT-004: Chat-driven mutations SHALL require preview and confirmation
Cabinet SHALL require preview and explicit confirm for mutation actions.

#### Scenario: Apply proposed mutation
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** chat proposes create/update/track/wishlist action
- **THEN** action SHALL not execute until user confirms apply
