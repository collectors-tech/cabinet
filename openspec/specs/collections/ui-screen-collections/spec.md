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
- **AND** the active collection context MUST persist across refresh for the signed-in profile

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
- **GIVEN** a non-protected collection row exists in the table
- **WHEN** the user confirms deletion
- **THEN** the row MUST be removed from the table
- **AND** the active management context MUST update deterministically
- **AND** items previously assigned to the deleted collection MUST remain in Cabinet as unassigned items

### Requirement UI-SCREEN-COLLECTIONS-006: Collections table SHALL support deterministic filtering
The shared table surface SHALL support simple filtering without leaving the table workflow.

#### Scenario: Filter collections
- **GIVEN** multiple collections exist
- **WHEN** the user filters the table
- **THEN** only matching collection rows MUST remain visible
- **AND** the filtered count summary MUST update deterministically

### Requirement UI-SCREEN-COLLECTIONS-007: Collections SHALL support assigning items into the selected collection
The collections workflow SHALL allow the user to place a Cabinet item into the currently selected collection from the same management surface.

#### Scenario: Assign item to selected collection
- **GIVEN** a non-protected collection is selected and assignable items exist
- **WHEN** the user assigns an item into that collection
- **THEN** the item MUST appear in the collection members panel
- **AND** the collection row/member state MUST reflect the saved placement without relying on ephemeral in-memory UI only
- **AND** the assignment MUST persist across refresh

### Requirement UI-SCREEN-COLLECTIONS-008: Collections SHALL support moving assigned items between collections
The collections workflow SHALL allow moving an already assigned item from one collection to another from the collections management surface.

#### Scenario: Move assigned item between collections
- **GIVEN** an item is already assigned to a non-protected collection
- **WHEN** the user moves that item to another non-protected collection
- **THEN** the item MUST leave the old collection view
- **AND** the item MUST appear in the destination collection view
- **AND** the moved membership state MUST persist across refresh

### Requirement UI-SCREEN-COLLECTIONS-009: Collections route SHALL retain clear tag-based page identity
Collections navigation and page identity SHALL continue to use the tag iconography for the collections route.

#### Scenario: Collections route iconography
- **GIVEN** the user opens `/collections`
- **WHEN** the navigation and page header render
- **THEN** the Collections route entry MUST remain visible in navigation
- **AND** the page identity area MUST render a visible tag icon

### Requirement UI-SCREEN-COLLECTIONS-010: Collections SHALL expose inventory view navigation from each row
The Collections management summary SHALL NOT expose a standalone Browse action. Each collection row SHALL expose an accessible View action that selects that row as the active collection context before navigating to Inventory.

#### Scenario: View collection from row action
- **GIVEN** collection rows are visible in the table
- **WHEN** the user activates a row-level View action for a collection
- **THEN** the collection MUST become the active collection context
- **AND** the app MUST navigate to Inventory
- **AND** the Inventory screen MUST show that collection as the active context
- **AND** existing row edit and delete actions MUST remain available

### Requirement UI-SCREEN-COLLECTIONS-022: Create collection dialog SHALL submit from Enter
The Create collection dialog SHALL treat Enter in its single-line collection name input as the same create action as the primary save button while preserving the same validation and persistence contract.

#### Scenario: Enter submits a valid collection
- **GIVEN** the Create collection dialog is open
- **WHEN** the user types a valid collection name and presses Enter in the name input
- **THEN** the dialog MUST create the collection through the existing save path
- **AND** the new collection MUST become the active management context
- **AND** the saved profile settings MUST persist the collection and active context

#### Scenario: Enter validates an invalid collection name
- **GIVEN** the Create collection dialog is open
- **WHEN** the user presses Enter with an empty or duplicate collection name
- **THEN** the dialog MUST stay open
- **AND** deterministic validation feedback MUST be shown without creating a collection
- **AND** the pending submit state MUST prevent duplicate collection creation while a submit is in progress

### Requirement UI-SCREEN-COLLECTIONS-023: Collections SHALL support fast create shortcuts and command entry
The Collections screen SHALL expose the Create collection workflow through a page-scoped `Ctrl+N` shortcut and through the global command palette so users can create collection rows without hunting for the header action.

#### Scenario: Ctrl+N opens Create collection
- **GIVEN** the user is on `/collections`
- **AND** focus is not inside a text input, textarea, select, contenteditable, or textbox control
- **WHEN** the user presses `Ctrl+N`
- **THEN** the Create collection dialog MUST open through the same create path as the primary header action
- **AND** submitting a valid collection name MUST persist the new collection and active context

#### Scenario: Command palette opens Create collection
- **GIVEN** the command palette is open
- **WHEN** the user selects the `Create collection` command entry with the `Ctrl+N` shortcut label
- **THEN** Cabinet MUST navigate to `/collections` and open the Create collection dialog
- **AND** submitting a valid collection name MUST persist the new collection and active context

### Requirement UI-SCREEN-COLLECTIONS-024: Collections filters SHALL expose deterministic empty states without mutation
Collections table and collection-members table filters SHALL show deterministic zero-result summaries and empty-row messages while preserving the selected collection context and avoiding profile-settings writes.

#### Scenario: Collections table filter has no matches
- **GIVEN** an authenticated user opens `/collections`
- **WHEN** the user enters a collection filter that matches no collection rows
- **THEN** the collections table MUST show zero matching rows
- **AND** the collections summary MUST show `Showing 0 of <total> collections.`
- **AND** the selected collection context MUST remain unchanged
- **AND** no profile-settings save MUST be sent by filtering alone

#### Scenario: Collection members filter has no matches
- **GIVEN** an authenticated user opens `/collections` with collection members visible
- **WHEN** the user enters a members filter that matches no member rows
- **THEN** the members table MUST show zero matching rows
- **AND** the members summary MUST show `Showing 0 of <selected total> items.`
- **AND** the empty-row message MUST explain that no collection members match the current filter
- **AND** no profile-settings save MUST be sent by filtering alone

### Requirement UI-SCREEN-COLLECTIONS-025: Collections transient workflows SHALL cancel without mutation
Collections create, edit, and delete transient surfaces SHALL close cleanly on explicit cancellation while preserving the current table rows, selected collection context, and persisted profile settings.

#### Scenario: Cancel create collection
- **GIVEN** an authenticated user opens `/collections`
- **WHEN** the user opens Create collection, enters a draft name, and cancels
- **THEN** the create dialog MUST close
- **AND** no draft collection row MUST be added
- **AND** the selected collection context MUST remain unchanged
- **AND** no profile-settings save MUST be sent

#### Scenario: Cancel edit collection
- **GIVEN** an authenticated user opens `/collections` with an editable collection selected
- **WHEN** the user opens the row edit panel, changes the draft name, and cancels
- **THEN** the edit panel MUST close
- **AND** the original collection name MUST remain visible
- **AND** no draft renamed row MUST be added
- **AND** the selected collection context MUST remain unchanged
- **AND** no profile-settings save MUST be sent

#### Scenario: Cancel delete collection
- **GIVEN** an authenticated user opens `/collections` with a deletable collection row selected
- **WHEN** the user opens the delete confirmation and cancels
- **THEN** the delete dialog MUST close
- **AND** the collection row MUST remain visible
- **AND** the selected collection context MUST remain unchanged
- **AND** no profile-settings save MUST be sent

### Requirement UI-SCREEN-COLLECTIONS-026: Collections row side panel SHALL support keyboard-sized record work
Collections row edit side panel SHALL provide deterministic record navigation and validation from the selected visible collection rows without falling back to a blocking edit dialog.

#### Scenario: Double-click opens row edit side panel and navigates records
- **GIVEN** an authenticated user opens `/collections` with editable collection rows visible
- **WHEN** the user double-clicks an editable row
- **THEN** Cabinet MUST open a right-side edit panel for that row instead of a blocking edit dialog
- **AND** next/previous controls MUST move the draft edit context across visible editable records
- **AND** saving a valid rename MUST persist the updated collection name through profile settings and survive refresh

#### Scenario: Duplicate rename remains local to the side panel
- **GIVEN** an authenticated user opens `/collections` with editable collection rows visible
- **WHEN** the user opens the row edit side panel and submits a duplicate collection name
- **THEN** the side panel MUST remain open with the attempted draft value
- **AND** no profile-settings save MUST be sent
- **AND** the original and duplicate source rows MUST remain unchanged
- **AND** the active collection context MUST remain on the row being edited

### Requirement UI-SCREEN-COLLECTIONS-027: Collections members panel SHALL reflect selected collection contents
The Collections members panel SHALL render the selected collection's current inventory members using the same live inventory catalogue and persisted workspace assignment state that drives collection row counts.

#### Scenario: All Items reflects live inventory members
- **GIVEN** an authenticated user opens `/collections` and live inventory records are available
- **WHEN** the Collections screen loads with `All Items` selected
- **THEN** the `All Items` row count MUST equal the live inventory member count
- **AND** the members table summary MUST show all live inventory members
- **AND** each visible member row MUST show its assigned or unassigned collection label

#### Scenario: Selecting a collection updates the members panel
- **GIVEN** an authenticated user opens `/collections` with assigned and empty collections available
- **WHEN** the user selects a collection that has assigned members
- **THEN** the members panel MUST show only the selected collection's assigned members
- **AND** the persisted active collection setting MUST match the selected collection
- **WHEN** the user selects a collection with no assigned members
- **THEN** the members panel MUST show the empty state for that selected collection
- **AND** it MUST NOT display members from the previously selected collection

### Requirement UI-SCREEN-COLLECTIONS-028: Collections table pagination SHALL preserve selection and avoid passive mutation
Collections table pagination SHALL let users reach collections beyond the first page while keeping profile writes limited to explicit collection selection.

#### Scenario: Select collection from a paginated page
- **GIVEN** an authenticated user opens `/collections` with enough collections to require multiple table pages
- **WHEN** the user navigates to a later collections page
- **THEN** later-page collection rows MUST render without saving profile settings
- **WHEN** the user selects a later-page collection row
- **THEN** the selected collection context MUST update visibly
- **AND** profile settings MUST persist the selected collection
- **AND** the selected context MUST survive refresh

#### Scenario: Filter after paginated selection
- **GIVEN** a later-page collection is selected
- **WHEN** the user filters the collections table to that row
- **THEN** the selected row MUST remain reachable and selected
- **AND** filtering MUST NOT send an additional profile-settings save

### Requirement UI-SCREEN-COLLECTIONS-029: Collection members pagination SHALL preserve selected context without passive mutation
The Collection members table SHALL let users reach selected-collection members beyond the first page while preserving the active collection context and keeping profile writes limited to explicit collection selection or item assignment actions.

#### Scenario: Page selected collection members
- **GIVEN** an authenticated user opens `/collections` with a selected collection containing enough assigned inventory members to require multiple members-table pages
- **WHEN** the user navigates to a later members page
- **THEN** later-page member rows MUST render for the selected collection
- **AND** the active collection context MUST remain unchanged
- **AND** no profile-settings save MUST be sent by members pagination alone
- **WHEN** the user refreshes the route
- **THEN** the same selected collection context MUST be restored from profile settings
- **AND** the later-page members MUST remain reachable

#### Scenario: Filter after members pagination
- **GIVEN** a selected collection has paginated member rows
- **WHEN** the user filters the members table to a later-page member
- **THEN** the matching member MUST remain visible for the selected collection
- **AND** the members summary MUST reflect the filtered result count
- **AND** filtering MUST NOT send a profile-settings save

## Acceptance Criteria
- Collections uses one practical table-driven management surface.
- Create, rename, delete, assign, and move workflows all happen from the collections route.
- Assignment and move outcomes survive refresh for the signed-in profile.

## Success Criteria
- Collections behaves like a practical operational surface, not a static taxonomy page.
- Item placement can be managed directly from Collections before moving on to later wishlist/owned-state work.

## Use-Case IDs and E2E Mapping
| UC ID | Flow | Expected Result | E2E Mapping |
| --- | --- | --- | --- |
| UC-COL-01 | Open collections table | Shared table surface renders | `ui.web/cypress/e2e/collections/ui-screen-collections/spec.cy.ts` `UI-SCREEN-COLLECTIONS-001 renders shared collections management table` |
| UC-COL-02 | Select row | Selected panel and active context update | `ui.web/cypress/e2e/collections/ui-screen-collections/spec.cy.ts` `UI-SCREEN-COLLECTIONS-002 selects a row and persists active context across refresh` |
| UC-COL-03 | Create collection | New row appears and survives refresh | `ui.web/cypress/e2e/collections/ui-screen-collections/spec.cy.ts` `UI-SCREEN-COLLECTIONS-003 creates a collection from the table workflow and persists after refresh` |
| UC-COL-04 | Rename collection | Updated row survives refresh | `ui.web/cypress/e2e/collections/ui-screen-collections/spec.cy.ts` `UI-SCREEN-COLLECTIONS-004 renames a collection from the row workflow and persists after refresh` |
| UC-COL-05 | Delete collection | Row removed and items released deterministically | `ui.web/cypress/e2e/collections/ui-screen-collections/spec.cy.ts` `UI-SCREEN-COLLECTIONS-005 deletes a collection and releases assigned items` |
| UC-COL-06 | Filter collections | Matching rows remain visible | `ui.web/cypress/e2e/collections/ui-screen-collections/spec.cy.ts` `UI-SCREEN-COLLECTIONS-006 filters collections within the shared table surface` |
| UC-COL-07 | Assign item | Item appears in selected collection and survives refresh | `ui.web/cypress/e2e/collections/ui-screen-collections/spec.cy.ts` `UI-SCREEN-COLLECTIONS-007 assigns an item into the selected collection and persists after refresh` |
| UC-COL-08 | Move item | Item leaves source and appears in destination after refresh | `ui.web/cypress/e2e/collections/ui-screen-collections/spec.cy.ts` `UI-SCREEN-COLLECTIONS-008 moves an assigned item between collections and persists after refresh` |
| UC-COL-09 | Route iconography | Tag icon remains visible for collections route | `ui.web/cypress/e2e/collections/ui-screen-collections/spec.cy.ts` `UI-SCREEN-COLLECTIONS-009 retains tag iconography for collections route identity` |
| UC-COL-10 | View row in Inventory | Row View selects collection and navigates to Inventory | `ui.web/cypress/e2e/collections/ui-screen-collections/spec.cy.ts` `UI-SCREEN-COLLECTIONS-010 moves Browse into row-level View actions` |
| UC-COL-22 | Press Enter in create dialog | Valid Enter submit persists; invalid Enter shows inline validation | `ui.web/cypress/e2e/collections/ui-screen-collections/spec.cy.ts` `UI-SCREEN-COLLECTIONS-022 submits Create collection with Enter and validates invalid Enter` |
| UC-COL-23 | Shortcut and command create | `Ctrl+N` and command palette entry open the same create dialog and persisted create flow | `ui.web/cypress/e2e/collections/ui-screen-collections/spec.cy.ts` `UI-SCREEN-COLLECTIONS-023 opens Create collection from Ctrl+N and command entry` |
| UC-COL-24 | Zero-result filters | Collections and members filters show deterministic empty states without profile-settings writes | `ui.web/cypress/e2e/collections/collections-filter-empty-states/spec.cy.ts` `UI-SCREEN-COLLECTIONS-024 shows deterministic zero-result filter states without saving settings` |
| UC-COL-25 | Cancel transient workflows | Create/edit/delete cancellation closes transient surfaces without adding, renaming, deleting, changing active context, or saving settings | `ui.web/cypress/e2e/collections/collections-cancel-no-mutation/spec.cy.ts` `UI-SCREEN-COLLECTIONS-025 cancels create edit and delete workflows without saving settings` |
| UC-COL-26 | Row edit side panel | Double-click opens the right-side edit panel, record navigation changes draft context, valid rename persists, and duplicate rename does not save | `ui.web/cypress/e2e/collections/collections-row-side-panel/spec.cy.ts` `UI-SCREEN-COLLECTIONS-026 opens and validates collection row side-panel workflows` |
| UC-COL-27 | Members panel contents | Selected collection drives member rows, live All Items count, assignment labels, and empty selected-collection states | `ui.web/cypress/e2e/collections/collections-members-panel/spec.cy.ts` `UI-SCREEN-COLLECTIONS-027 reflects selected collection members and empty states` |
| UC-COL-28 | Paginate collections table | Later-page rows render without passive settings writes; selected later-page context persists across refresh and remains selected after filtering | `ui.web/cypress/e2e/collections/collections-pagination/spec.cy.ts` `UI-SCREEN-COLLECTIONS-028 preserves paginated collection selection without passive settings writes` |
| UC-COL-29 | Paginate collection members | Later-page member rows render for the selected collection without passive settings writes; selected context survives refresh and member filtering remains non-mutating | `ui.web/cypress/e2e/collections/collections-members-pagination/spec.cy.ts` `UI-SCREEN-COLLECTIONS-029 preserves paginated collection members without passive settings writes` |
