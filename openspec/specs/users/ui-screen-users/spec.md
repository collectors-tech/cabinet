## Purpose
Define Users screen behavior for table management and user action dialogs.

## Requirements
### Requirement UI-SCREEN-USERS-001: Users screen SHALL support filter/sort/pagination table workflows
Users screen SHALL provide searchable and filterable user list with sortable columns and pagination.

#### Scenario: Filter and paginate users
- **GIVEN** users route is open with table data loaded
- **WHEN** user applies username/status/role filters and navigates pages
- **THEN** table rows MUST reflect active filters and pagination state without route failure

### Requirement UI-SCREEN-USERS-002: Users screen SHALL support invite and add user dialogs
Users screen SHALL provide `Invite User` and `Add User` actions from primary controls.

#### Scenario: Open invite/add dialogs
- **GIVEN** users route is loaded
- **WHEN** user clicks `Invite User` or `Add User`
- **THEN** corresponding dialog MUST open and remain actionable until saved or canceled

### Requirement UI-SCREEN-USERS-003: Users screen SHALL support edit/delete actions from row context
Users screen SHALL support row-level edit and delete actions with confirmation flow.

#### Scenario: Edit or delete selected row
- **GIVEN** a user row is selected from table row actions
- **WHEN** user triggers edit or delete action
- **THEN** edit dialog or delete confirmation MUST open with selected user context and close cleanly on completion/cancel

## Use-Case IDs and E2E Mapping
| UC ID | Flow | Expected Result | E2E Mapping |
| --- | --- | --- | --- |
| UC-USR-01 | Filter by role/status | Table reflects filtered rows | planned: `cypress/e2e/ui/users.cy.ts` `users-filtering` |
| UC-USR-02 | Add user | Add dialog opens and saves | planned: `cypress/e2e/ui/users.cy.ts` `users-add` |
| UC-USR-03 | Invite user | Invite dialog opens and submits | planned: `cypress/e2e/ui/users.cy.ts` `users-invite` |
| UC-USR-04 | Edit/delete user | Row action dialogs open with selected context | planned: `cypress/e2e/ui/users.cy.ts` `users-row-actions` |
