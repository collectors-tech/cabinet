## Purpose
Define top-level Collections screen behavior as a shared Cabinet table management surface.

## Requirements
### Requirement UI-SCREEN-COLLECTIONS-001: Collections SHALL use the shared table management pattern
Collections SHALL render as a practical row-based table surface using the shared Cabinet table pattern instead of a custom mixed create/list layout.

#### Scenario: Collections table renders
- **GIVEN** an authenticated user opens `/collections`
- **WHEN** the screen loads
- **THEN** the primary collections surface MUST render as a shared table
- **AND** the table MUST show collection rows with management-oriented columns
- **AND** the screen MUST expose one primary create action for adding a collection row

### Requirement UI-SCREEN-COLLECTIONS-002: Collections SHALL support row selection with visible management context
The screen SHALL keep row selection and the current collection management context visible while the user works from the table.

#### Scenario: Select collection row
- **GIVEN** collection rows are visible in the table
- **WHEN** the user selects a collection row
- **THEN** the row MUST become the active collection context
- **AND** a visible selected-collection panel MUST reflect the selected row details

### Requirement UI-SCREEN-COLLECTIONS-003: Collections SHALL support create from the table workflow
The shared table surface SHALL support creating a collection through a single explicit create flow.

#### Scenario: Create collection from table surface
- **GIVEN** the user is on `/collections`
- **WHEN** the user triggers the primary create action and submits a valid collection name
- **THEN** the collection MUST be added to the table
- **AND** the new collection MUST become the active management context
- **AND** the result MUST persist across refresh

### Requirement UI-SCREEN-COLLECTIONS-004: Collections SHALL support row edit workflow
The shared table surface SHALL support editing a collection row through the standard row-management workflow.

#### Scenario: Rename collection from row workflow
- **GIVEN** a collection row exists in the table
- **WHEN** the user edits the row and saves a valid new name
- **THEN** the table MUST update to show the new collection name
- **AND** the updated name MUST persist across refresh

### Requirement UI-SCREEN-COLLECTIONS-005: Collections SHALL support row delete workflow
The shared table surface SHALL support removing collection rows through an explicit delete confirmation flow.

#### Scenario: Delete collection from row workflow
- **GIVEN** a collection row exists in the table
- **WHEN** the user confirms deletion
- **THEN** the row MUST be removed from the table
- **AND** the active management context MUST update deterministically

### Requirement UI-SCREEN-COLLECTIONS-006: Collections table SHALL support deterministic filtering
The shared table surface SHALL support simple filtering without leaving the table workflow.

#### Scenario: Filter collections
- **GIVEN** multiple collections exist
- **WHEN** the user filters the table
- **THEN** only matching collection rows MUST remain visible
- **AND** the filtered count summary MUST update deterministically

## Acceptance Criteria
- Collections no longer depends on a custom mixed create/list layout.
- The shared table pattern drives create, edit, delete, selection, and filtering behaviors.
- Refresh-persistence evidence exists for create and rename flows.

## Success Criteria
- Collections behaves like a practical management surface, not a one-off custom page.
- The next collections issue can build assignment/move behavior on top of this table base cleanly.

## Use-Case IDs and E2E Mapping
| UC ID | Flow | Expected Result | E2E Mapping |
| --- | --- | --- | --- |
| UC-COL-01 | Open collections table | Shared table surface renders | `ui.web/cypress/e2e/collections/ui-screen-collections/spec.cy.ts` `renders collections as shared management table` |
| UC-COL-02 | Select row | Selected collection panel updates | `ui.web/cypress/e2e/collections/ui-screen-collections/spec.cy.ts` `selects a collection row and updates management context` |
| UC-COL-03 | Create collection | New row appears and survives refresh | `ui.web/cypress/e2e/collections/ui-screen-collections/spec.cy.ts` `creates a collection from the table workflow and persists after refresh` |
| UC-COL-04 | Rename collection | Updated row survives refresh | `ui.web/cypress/e2e/collections/ui-screen-collections/spec.cy.ts` `renames a collection from the row workflow and persists after refresh` |
| UC-COL-05 | Delete collection | Row removed from table | `ui.web/cypress/e2e/collections/ui-screen-collections/spec.cy.ts` `deletes a collection from the row workflow` |
| UC-COL-06 | Filter collections | Matching rows remain visible | `ui.web/cypress/e2e/collections/ui-screen-collections/spec.cy.ts` `filters collections within the shared table surface` |
