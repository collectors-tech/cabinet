## Purpose
Define Users screen behavior for table management and user action dialogs.

## Requirements
### Requirement UI-SCREEN-USERS-001: Users screen SHALL support filter/sort/pagination table workflows
Users screen SHALL load list data from `GET /api/users` and provide searchable/filterable/sortable table workflows with pagination.

#### Scenario: Filter and paginate users
- **GIVEN** an authenticated user opens `/users` and `GET /api/users` returns `200` with a `users[]` payload
- **WHEN** user applies username/status/role filters and navigates pages
- **THEN** table rows MUST reflect active filters and pagination state without route failure

### Requirement UI-SCREEN-USERS-002: Users screen SHALL support invite and add user dialogs
Users screen SHALL provide `Invite User` and `Add User` actions that persist through Cabinet API.

#### Scenario: Open invite/add dialogs
- **GIVEN** users route is loaded
- **WHEN** user submits `Add User` (`POST /api/users`) or `Invite User` (`POST /api/users/invite`)
- **THEN** API response MUST be successful (`201`/`200`) and table MUST refresh with the new/updated user row

### Requirement UI-SCREEN-USERS-003: Users screen SHALL support row selection, details, edit, and delete actions from row context
Users screen SHALL support row selection, details viewing, and row-level role/status edit and delete actions with API persistence.

#### Scenario: Select a row and open user details
- **GIVEN** an authenticated desktop user has the Users table loaded and at least two visible rows are available
- **WHEN** the user single-clicks a row
- **THEN** Cabinet MUST mark that user as the selected row without opening the details dialog or invoking bulk checkbox selection
- **AND** the selected-row `View user` action MUST open the details dialog for that selected user.

#### Scenario: Open details from row double-click
- **GIVEN** an authenticated desktop user has the Users table loaded and at least two visible rows are available
- **WHEN** the user double-clicks a user row
- **THEN** Cabinet MUST open the user details dialog for the double-clicked row, preserve the selected row context, and avoid invoking edit/delete row actions.

### Requirement UI-SCREEN-USERS-004: Users screen SHALL expose actionable retry when list fetch fails
When `GET /api/users` fails, Users screen SHALL render deterministic error state with `Retry` control.

#### Scenario: Retry users list after fetch failure
- **GIVEN** users list request fails and error state is visible
- **WHEN** user clicks `Retry`
- **THEN** screen MUST re-attempt users list fetch and recover to ready/empty/error state deterministically

### Requirement UI-SCREEN-USERS-005: Users screen SHALL remain available when active profile is missing
When no active profile is currently set for the authenticated session, `GET /api/users` SHALL return a deterministic default-scope users payload instead of `404`.

#### Scenario: Users route loads without an active profile
- **GIVEN** an authenticated user lands on `/users` and profile runtime returns no active profile id
- **WHEN** the screen requests `GET /api/users`
- **THEN** API response MUST be `200` with `users[]` payload and the screen MUST render the `Users` shell title without `users_fetch_failed_404`

### Requirement UI-SCREEN-USERS-006: Users screen SHALL expose deterministic loading and empty states
Users screen SHALL keep the route shell and primary controls visible while list data loads, and SHALL render a deterministic empty table state when the Cabinet users API returns no rows.

#### Scenario: Loading and empty users list states
- **GIVEN** an authenticated user opens `/users` and `GET /api/users` is pending or returns `200` with an empty `users[]` payload
- **WHEN** the route renders the pending state and then receives the empty result
- **THEN** the screen MUST show `Loading users...` without hiding the route purpose or primary actions while pending
- **AND** the ready state MUST show the users table empty row without rendering a fetch error or stale user row

#### Scenario: Edit or delete selected row
- **GIVEN** a user row is selected from table row actions
- **WHEN** user saves edits (`PUT /api/users/{id}`) or confirms deletion (`DELETE /api/users/{id}`)
- **THEN** API response MUST be successful and table MUST reflect persisted role/status changes or row removal
- **AND** an edit save MUST refresh the list and continue to show the updated row after a route refresh without retaining the stale pre-edit username/email

### Requirement UI-SCREEN-USERS-007: Users screen SHALL support hiding optional table columns without losing data context
Users screen SHALL expose a table view-options control for hideable columns and keep non-hideable identity/action columns available when optional columns are hidden.

#### Scenario: Hide optional table columns
- **GIVEN** an authenticated desktop user has the Users table loaded
- **WHEN** the user opens `View` and hides the optional `email` column
- **THEN** the table MUST hide the email header and cells while preserving username/status/role/action context
- **AND** the route MUST remain on `/users` without losing the current users rows

### Requirement UI-SCREEN-USERS-008: Users screen SHALL persist bulk toolbar actions through Cabinet API
Users screen SHALL make selected-row bulk invite, status, and delete actions durable through Cabinet API calls and SHALL refresh the table from the resulting users source of truth.

#### Scenario: Run selected-row bulk actions
- **GIVEN** an authenticated desktop user has selected multiple users from the Users table
- **WHEN** the user invokes bulk invite, bulk deactivate or activate, and bulk delete actions from the selected-row toolbar
- **THEN** invite actions MUST call `POST /api/users/invite` for each selected user with the selected user's email and role
- **AND** status actions MUST call `PUT /api/users/{id}` for each selected user and refresh the table with the persisted status
- **AND** delete actions MUST require explicit confirmation, call `DELETE /api/users/{id}` for each selected user, refresh the table, and remove deleted rows from the visible source-of-truth state

### Requirement UI-SCREEN-USERS-009: Users screen SHALL use compact global header actions with table controls in the toolbar
Users screen SHALL present route identity and primary user actions in the shared Cabinet shell header while keeping search, filters, and table view controls visually attached to the table toolbar.

#### Scenario: Render Users actions in the shell header action lane
- **GIVEN** an authenticated desktop user opens `/users`
- **WHEN** the Users screen renders ready, loading, or empty table states
- **THEN** the sticky shell header MUST expose `Users`, `Invite User`, and `Add User` in the global header action lane
- **AND** the workspace MUST NOT render a duplicate `User List` heading block above the table
- **AND** the table toolbar MUST retain the username search, status filter, role filter, and `View` column-control action next to the table
- **AND** keyboard focus order MUST reach shell actions before table toolbar controls.

## Use-Case IDs and E2E Mapping
| UC ID | Flow | Expected Result | E2E Mapping |
| --- | --- | --- | --- |
| UC-USR-01 | Filter by username | Table reflects filtered rows and URL state | implemented: `ui.web/cypress/e2e/users/ui-screen-users/spec.cy.ts` `UI-SCREEN-USERS-001 reads users table from Cabinet API and supports filter/sort/pagination workflows` |
| UC-USR-02 | Add user | Add dialog opens, saves, and refreshed table shows the new user | implemented: `ui.web/cypress/e2e/users/ui-screen-users/spec.cy.ts` `UI-SCREEN-USERS-002 persists add and invite actions through Cabinet API` |
| UC-USR-03 | Invite user | Invite dialog opens, submits, and refreshed table shows the invited user | implemented: `ui.web/cypress/e2e/users/ui-screen-users/spec.cy.ts` `UI-SCREEN-USERS-002 persists add and invite actions through Cabinet API` |
| UC-USR-04 | View/edit/delete user | Row selection, selected-row View details, row double-click details, explicit row action edit dialog, edit save, and delete persistence are covered | implemented: `ui.web/cypress/e2e/users/ui-screen-users/spec.cy.ts` `UI-SCREEN-USERS-003 selects rows and opens details from View user and row double-click`; `UI-SCREEN-USERS-003 persists edit saves through Cabinet API and refreshes the edited row`; `UI-SCREEN-USERS-003 persists delete actions through Cabinet API row context` |
| UC-USR-05 | Users fetch failure retry | Error state `Retry` re-attempts list fetch deterministically | implemented: `ui.web/cypress/e2e/users/ui-screen-users/spec.cy.ts` `UI-SCREEN-USERS-004 retries users list after a fetch failure` |
| UC-USR-06 | Missing active profile fallback | Users screen loads list without 404 fallback error | `ui.web/cypress/e2e/users/ui-screen-users/fallback-profile-scope.cy.ts` |
| UC-USR-07 | Loading and empty list states | Pending users list shows loading feedback; empty API response renders table empty state without stale rows or error banner | implemented: `ui.web/cypress/e2e/users/ui-screen-users/spec.cy.ts` `UI-SCREEN-USERS-006 renders deterministic loading and empty states` |
| UC-USR-08 | Table view options | View menu hides optional columns while preserving core row context | implemented: `ui.web/cypress/e2e/users/ui-screen-users/spec.cy.ts` `UI-SCREEN-USERS-007 hides optional table columns from the View menu` |
| UC-USR-09 | Bulk user actions | Selected-row invite, status, and delete actions call Cabinet APIs, refresh the users list, and prove persisted table outcomes | implemented: `ui.web/cypress/e2e/users/ui-screen-users/spec.cy.ts` `UI-SCREEN-USERS-008 persists bulk invite, status, and delete actions through Cabinet API` |
| UC-USR-10 | Compact global header actions | Users title, Invite User, and Add User render in the shell header while search/filter/View controls remain in the table toolbar and duplicate page chrome is absent | implemented: `ui.web/cypress/e2e/users/ui-screen-users/spec.cy.ts` `UI-SCREEN-USERS-009 keeps primary actions in the Users shell header and table controls in the toolbar` |
