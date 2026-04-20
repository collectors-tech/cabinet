## Purpose
Define the authenticated Cabinet shell workspace model so Navigation, Assistant, and Inbox are first-class workspaces instead of ad-hoc panels.

## Requirements
### Requirement UI-SHELL-WORKSPACES-001: Cabinet SHALL provide left-rail workspace switching for Navigation, Assistant, and Inbox
Authenticated shell MUST provide a deterministic workspace switcher that exposes Navigation, Assistant, and Inbox as top-level shell workspaces.

#### Scenario: Switch shell workspaces
- **GIVEN** authenticated shell is visible
- **WHEN** user selects `Navigation`
- **THEN** left workspace region MUST show primary app navigation
- **WHEN** user selects `Assistant`
- **THEN** left workspace region MUST show assistant thread/composer workspace
- **WHEN** user selects `Inbox`
- **THEN** left workspace region MUST show notifications/event inbox workspace

### Requirement UI-SHELL-WORKSPACES-002: Header assistant trigger SHALL activate Assistant workspace instead of a competing rail/panel model
Cabinet header assistant/chat trigger MUST switch shell state into Assistant workspace and MUST NOT introduce a conflicting temporary right-rail pattern.

#### Scenario: Activate Assistant from header
- **GIVEN** user is on any authenticated route
- **WHEN** user clicks header assistant control
- **THEN** shell MUST visibly activate Assistant workspace
- **AND** current route URL/query state MUST remain unchanged

### Requirement UI-SHELL-WORKSPACES-003: Shell workspace state SHALL persist across authenticated route changes
Selected shell workspace MUST persist while user navigates between authenticated routes until an explicit reset boundary occurs.

#### Scenario: Preserve selected workspace during navigation
- **GIVEN** user has Assistant workspace active
- **WHEN** user navigates from `/inventory` to `/wishlist`
- **THEN** Assistant workspace MUST remain active after navigation

### Requirement UI-SHELL-WORKSPACES-004: Shell workspace semantics SHALL distinguish Navigation, Assistant, Inbox, and optional Chats
Cabinet MUST maintain distinct responsibilities for shell workspaces so Assistant, Inbox, and any `/chats` route do not collapse into one vague surface.

#### Scenario: Distinct workspace semantics
- **GIVEN** Cabinet renders shell workspace controls and `/chats` route exists
- **WHEN** user opens Assistant workspace, Inbox workspace, and `/chats`
- **THEN** Assistant MUST behave as AI helper workspace
- **AND** Inbox MUST behave as notification/event workspace
- **AND** `/chats` MUST behave as intentional conversation workspace rather than duplicate Inbox or placeholder assistant shell

### Requirement UI-SHELL-WORKSPACES-005: Inbox empty state SHALL expose actionable next steps
When Inbox has no items, the workspace MUST provide clear actions so users are not left at a dead end.

#### Scenario: Inbox empty state actions
- **GIVEN** Inbox workspace is open and no inbox items exist
- **WHEN** the empty state renders
- **THEN** the workspace MUST show at least one explicit refresh or navigation affordance
- **AND** users MUST be able to open a related communications surface without guessing
