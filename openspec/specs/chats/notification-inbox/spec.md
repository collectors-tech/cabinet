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

### Requirement UI-SCREEN-NOTIFICATION-INBOX-006: Notification Inbox SHALL preserve rows when triage updates fail
When a read, unread, or archive request fails, the Notification Inbox MUST show a retryable update error and MUST keep the affected row visible with its previous status so users can retry without losing queue context.

#### Scenario: Failed row update keeps queue context
- **GIVEN** an unread notification row is visible
- **WHEN** the user marks the row read and the update API fails
- **THEN** the update error MUST be visible
- **AND** the row MUST remain in the queue with its unread status.

### Requirement UI-SCREEN-NOTIFICATION-INBOX-007: Notification Inbox SHALL provide dense recoverable triage
The Notification Inbox MUST render a compact two-pane operating layout with filter counts, a table-style paginated message list, total count, selected notification detail, and icon-only page actions. Clearing visible notifications MUST archive/hide the currently visible records without dropping them from the durable Inbox state, and Show hidden MUST reveal those archived records again.

#### Scenario: Recover cleared notifications
- **GIVEN** the Inbox has read, unread, assistant, system, and archived notifications
- **WHEN** the user clears the currently visible queue
- **THEN** the visible queue MUST hide those notifications
- **AND** the hidden notifications MUST remain available when Show hidden is enabled.

#### Scenario: Dense table and detail layout
- **GIVEN** the Inbox has more rows than one page can show
- **WHEN** the route renders
- **THEN** it MUST show filter buttons with inline counts, stat counters, a search field, table-style paginated rows, a total message count below the table, and a selected notification detail pane
- **AND** page actions such as refresh, mark all visible as read, clear all, and show hidden MUST be icon-only with accessible names/tooltips.

### Requirement UI-SCREEN-NOTIFICATION-INBOX-008: Notification-like UI events SHALL be preserved in Inbox history
Cabinet notification-like UI events, including toast messages, promise-based success/failure feedback, shared confirmation or warning dialogs, and inline status/banner messages from operational settings, storage maintenance, taxonomy settings, Collections workspace, and Integrations provider health surfaces, MUST be captured into the Notification Inbox history so immediate feedback is not the only record. Captured records MUST include a source label, event time, level/category metadata, title, and detail or lifecycle summary sufficient for later review from the Inbox route. Once Cabinet has an active profile, captured local notification history MUST be promoted into the server-backed Inbox store and deduplicated by local capture ID so the record survives local history loss or reload.

#### Scenario: Toast lifecycle events appear in Inbox
- **GIVEN** a user triggers a promise-based UI feedback flow
- **WHEN** the feedback shows loading and then settles
- **THEN** the Notification Inbox MUST include the captured feedback lifecycle records
- **AND** the records MUST show source, time, level/category metadata, and detail sufficient to identify the event after the toast disappears.

#### Scenario: Shared confirmation dialogs appear in Inbox
- **GIVEN** a user opens a shared confirmation dialog for a destructive or warning action
- **WHEN** the dialog is dismissed and the user later opens the Notification Inbox
- **THEN** the Notification Inbox MUST include a captured dialog history record
- **AND** the record MUST show source, time, level/category metadata, title, and detail sufficient to identify the dialog after it disappears.

#### Scenario: Inline status banners appear in Inbox
- **GIVEN** a user triggers an operational settings, storage maintenance, taxonomy settings, Collections workspace, or Integrations provider health action that renders inline or toast-backed success or failure status copy
- **WHEN** the user later opens the Notification Inbox
- **THEN** the Notification Inbox MUST include a captured status/banner history record
- **AND** the record MUST show source, time, level/category metadata, title, and detail sufficient to identify the operational status after it disappears or is replaced.

#### Scenario: Local notification history is promoted to durable Inbox storage
- **GIVEN** local notification history contains a captured notification-like UI event and an active profile is available
- **WHEN** the Notification Inbox syncs history with `/api/chat/inbox`
- **THEN** Cabinet MUST create a server-backed Inbox record with source, time, state, category/type, title, detail, and source label metadata
- **AND** repeated syncs of the same local capture ID MUST return the existing Inbox record rather than creating duplicate rows
- **AND** the promoted record MUST remain visible from `/api/chat/inbox` even if local browser history is cleared.
