## Purpose
Define the first-class Notification Inbox surface for operational Cabinet notifications, assistant handoffs, provider/import events, mentions, and system/runtime messages.

## Requirements
### Requirement UI-SCREEN-NOTIFICATION-INBOX-001: Cabinet SHALL expose a first-class Notification Inbox route
Authenticated Cabinet MUST expose `/inbox` as a Notification Inbox page and MUST NOT render the Purchases surface at that route.

#### Scenario: Notification Inbox route renders
- **GIVEN** an authenticated user opens `/inbox`
- **WHEN** the route renders
- **THEN** the page MUST be titled `Notification Inbox`
- **AND** the page MUST show notification triage controls instead of Purchases controls.

### Requirement UI-SCREEN-NOTIFICATION-INBOX-002: Notification Inbox SHALL support triage filters and contextual empty states
The Notification Inbox MUST provide category filters for `All`, `Unread`, `Assistant`, and `System`, with counts and filter-specific empty-state copy.

#### Scenario: Filter notification queue
- **GIVEN** notifications exist in multiple categories
- **WHEN** the user selects each filter
- **THEN** the visible rows MUST match the selected category
- **AND** an empty filter MUST show contextual empty-state text for that filter.

### Requirement UI-SCREEN-NOTIFICATION-INBOX-003: Notification rows SHALL expose state, context, expansion, and linked targets
Each Notification Inbox row MUST show read/unread state, category, title, summary, timestamp, source, and a linked Cabinet target when available. Rows MUST expand to reveal detail/context without losing the row action controls.

#### Scenario: Expand notification detail
- **GIVEN** a notification row has detail context
- **WHEN** the user opens row details
- **THEN** the full detail/context MUST be visible
- **AND** the row target link MUST remain available.

### Requirement UI-SCREEN-NOTIFICATION-INBOX-004: Notification Inbox SHALL support row and bulk read/archive actions
The Notification Inbox MUST allow users to mark rows read/unread, archive rows, select visible rows, bulk mark selected rows read, and bulk archive selected rows.

#### Scenario: Triage selected notifications
- **GIVEN** multiple visible notifications are selected
- **WHEN** the user bulk marks them read or archives them
- **THEN** the UI MUST call the notification update API for each selected row
- **AND** the visible queue MUST update without requiring a full page reload.

### Requirement UI-SCREEN-NOTIFICATION-INBOX-005: Notification Inbox SHALL show loading, retryable error, and refresh states
The Notification Inbox MUST show visible loading, refresh, retryable error, and retry controls for failed notification loads or updates.

#### Scenario: Retry failed notification load
- **GIVEN** the notification API request fails
- **WHEN** the page renders the failure state
- **THEN** the user MUST see a retry action
- **AND** retrying MUST request the notification queue again.
