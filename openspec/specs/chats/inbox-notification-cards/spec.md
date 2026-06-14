## Purpose
Define the authenticated shell Inbox workspace notification-card behavior for catch-up cards, read-state triage, assistant routing, and linked target routing.

## Requirements
### Requirement UI-SCREEN-INBOX-NOTIFICATION-CARDS-001: Shell Inbox SHALL show actionable catch-up notification cards
The shell Inbox workspace MUST render non-archived notification cards from `/api/chat/inbox`, including title, summary, source, age, read/unread state, and available target links.

#### Scenario: Render catch-up cards
- **GIVEN** an authenticated user has unread, read, and archived Inbox notifications
- **WHEN** the user opens the shell Inbox workspace
- **THEN** the workspace MUST show non-archived catch-up notification cards
- **AND** each visible card MUST expose title, summary, source label, age, read/unread state, and item or review links when metadata provides them.

### Requirement UI-SCREEN-INBOX-NOTIFICATION-CARDS-002: Shell Inbox SHALL support card read-state triage
The shell Inbox workspace MUST let users mark a notification card read, mark it unread, and archive it through `/api/chat/inbox/:id` without a full route reload.

#### Scenario: Triage card read state
- **GIVEN** an unread notification card is visible in the shell Inbox workspace
- **WHEN** the user marks it read, marks it unread, and archives it
- **THEN** each action MUST send the active profile id and requested status to `/api/chat/inbox/:id`
- **AND** the card state MUST update in place
- **AND** archived cards MUST leave the visible catch-up list.

### Requirement UI-SCREEN-INBOX-NOTIFICATION-CARDS-003: Shell Inbox SHALL route Telegram capture reviews to the requested Chat thread
Telegram catalog capture Inbox cards MUST expose capture review URLs as Cabinet links and the linked Chats route MUST select the requested capture thread from the review URL.

#### Scenario: Open Telegram capture review
- **GIVEN** a Telegram catalog capture Inbox item has `metadata.review_url`, `thread_id`, and `preview_id`
- **WHEN** the user follows the Inbox review link
- **THEN** the link MUST point to the Cabinet Chats review URL
- **AND** the Chats route MUST select the requested Telegram capture thread instead of defaulting to another thread.

### Requirement UI-SCREEN-INBOX-NOTIFICATION-CARDS-004: Shell Inbox SHALL preserve cards when triage updates fail
When a shell Inbox read, unread, or archive request fails, Cabinet MUST show a deterministic update error and keep the affected notification card visible with its previous status.

#### Scenario: Failed card update keeps retry context
- **GIVEN** an unread shell Inbox notification card is visible
- **WHEN** the user marks the card read and `/api/chat/inbox/:id` fails
- **THEN** the shell Inbox MUST show update failure feedback
- **AND** the card MUST remain visible with its previous unread status
- **AND** the read action MUST be available for retry.
