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

### Requirement UI-SCREEN-USERS-003: Users screen SHALL support edit/delete actions from row context
Users screen SHALL support row-level role/status edit and delete actions with API persistence.

### Requirement UI-SCREEN-USERS-004: Users screen SHALL expose actionable retry when list fetch fails
When `GET /api/users` fails, Users screen SHALL render deterministic error state with `Retry` control.

#### Scenario: Retry users list after fetch failure
- **GIVEN** users list request fails and error state is visible
- **WHEN** user clicks `Retry`
- **THEN** screen MUST re-attempt users list fetch and recover to ready/empty/error state deterministically

#### Scenario: Edit or delete selected row
- **GIVEN** a user row is selected from table row actions
- **WHEN** user saves edits (`PUT /api/users/{id}`) or confirms deletion (`DELETE /api/users/{id}`)
- **THEN** API response MUST be successful and table MUST reflect persisted role/status changes or row removal

## Use-Case IDs and E2E Mapping
| UC ID | Flow | Expected Result | E2E Mapping |
| --- | --- | --- | --- |
| UC-USR-01 | Filter by role/status | Table reflects filtered rows | planned: `cypress/e2e/ui/users.cy.ts` `users-filtering` |
| UC-USR-02 | Add user | Add dialog opens and saves | planned: `cypress/e2e/ui/users.cy.ts` `users-add` |
| UC-USR-03 | Invite user | Invite dialog opens and submits | planned: `cypress/e2e/ui/users.cy.ts` `users-invite` |
| UC-USR-04 | Edit/delete user | Row action dialogs open with selected context | planned: `cypress/e2e/ui/users.cy.ts` `users-row-actions` |
| UC-USR-05 | Users fetch failure retry | Error state `Retry` re-attempts list fetch deterministically | planned: `ui.web/cypress/e2e/users/ui-screen-users/spec.cy.ts` `users-error-retry` |
