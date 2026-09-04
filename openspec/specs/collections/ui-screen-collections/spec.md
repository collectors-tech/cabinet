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

### Requirement UI-SCREEN-COLLECTIONS-011: Collections SHALL isolate persisted state by active profile
Collections SHALL load collection rows, active collection context, and collection members from the active profile only.

#### Scenario: Switch active profile
- **GIVEN** two profiles have different persisted collection settings
- **WHEN** the active profile changes and the user opens `/collections`
- **THEN** the collections table MUST show the new active profile's collections
- **AND** the active collection context MUST reflect the new active profile's persisted setting
- **AND** members from the previous active profile MUST NOT leak into the new profile context

### Requirement UI-SCREEN-COLLECTIONS-012: Inventory collection create SHALL stay synchronized with Collections
Collections created from the Inventory folder tree SHALL persist through profile settings and remain available to the Collections management route.

#### Scenario: Create collection from Inventory
- **GIVEN** an authenticated user opens Inventory
- **WHEN** the user creates a root collection from the folder tree
- **THEN** the collection MUST become the active Inventory collection context
- **AND** the created collection MUST persist across refresh
- **AND** the same persisted collection state MUST be available to `/collections`

### Requirement UI-SCREEN-COLLECTIONS-013: Wishlist collection create SHALL synchronize into Collections
Collections created from the Wishlist table collection workflow SHALL persist into the shared profile collection state used by Collections.

#### Scenario: Create collection from Wishlist
- **GIVEN** an authenticated user opens Wishlist
- **WHEN** the user creates a collection from the wishlist table collection workflow
- **THEN** the collection MUST be saved through profile settings
- **AND** opening `/collections` MUST show the created collection row
- **AND** the active collection context MUST reflect the created collection

### Requirement UI-SCREEN-COLLECTIONS-014: Collection renames SHALL propagate into compact filters
Renaming a collection from Collections SHALL update downstream compact collection filters that use the shared profile collection state.

#### Scenario: Rename collection and inspect Wishlist filter
- **GIVEN** an authenticated user renames a collection from `/collections`
- **WHEN** the user opens Wishlist and inspects the collection filter
- **THEN** the renamed collection MUST appear in the filter
- **AND** the old collection name MUST NOT remain as a selectable filter option
- **AND** selecting the renamed collection MUST show the renamed filter state

### Requirement UI-SCREEN-COLLECTIONS-015: Collection deletes SHALL propagate into compact filters
Deleting a collection from Collections SHALL remove it from downstream compact collection filters and leave the active context deterministic.

#### Scenario: Delete collection and inspect Wishlist filter
- **GIVEN** an authenticated user deletes a non-protected collection from `/collections`
- **WHEN** the delete is saved through profile settings
- **THEN** the deleted collection MUST be absent from persisted collection settings
- **AND** the active collection context MUST fall back deterministically
- **AND** downstream compact filters MUST NOT show the deleted collection option

### Requirement UI-SCREEN-COLLECTIONS-016: Collections SHALL retire the lower members table
Collections SHALL show collection rows as the primary working surface and SHALL NOT render a separate lower `Collection members` table on the Collections route.

#### Scenario: Open collections without a members table
- **GIVEN** an authenticated user opens `/collections`
- **WHEN** the screen loads
- **THEN** the main Collections table MUST be visible and usable
- **AND** the route MUST NOT render a `Collection members` card or lower members table
- **AND** collection row selection MUST still update the active collection context
- **AND** useful collection information such as live item counts MUST remain visible in the main table

### Requirement UI-SCREEN-COLLECTIONS-017: Collection members table filtering is retired from this route
The selected collection members table filter SHALL be absent because the separate members table is retired from Collections.

#### Scenario: Members filter is absent
- **GIVEN** an authenticated user opens `/collections`
- **WHEN** the screen loads
- **THEN** no collection-members filter input MUST render
- **AND** no collection-members summary MUST render

### Requirement UI-SCREEN-COLLECTIONS-018: All Items count SHALL match live inventory
The protected `All Items` collection row SHALL derive its count from the live inventory catalogue.

#### Scenario: Render All Items count
- **GIVEN** live inventory records are available
- **WHEN** an authenticated user opens `/collections`
- **THEN** the `All Items` row count MUST equal the live inventory item count
- **AND** the members table summary MUST show the same live member total
- **AND** representative live inventory members MUST be visible in the members table

### Requirement UI-SCREEN-COLLECTIONS-019: Collections legacy assignment controls are retired from this route
The pre-table in-route assignment control contract for Collections SHALL be retired and replaced by table/member-surface contracts so closure evidence does not imply active `UI-SCREEN-COLLECTIONS-019` Cypress coverage.

#### Scenario: Retired assignment control slot
- **GIVEN** the Collections table workflow is active
- **WHEN** traceability is reviewed
- **THEN** `UI-SCREEN-COLLECTIONS-019` MUST be documented as retired/replaced
- **AND** active member-surface behavior MUST be covered by `UI-SCREEN-COLLECTIONS-016`, `017`, `018`, `020`, `021`, and later focused #1078 requirements

### Requirement UI-SCREEN-COLLECTIONS-020: Collections table SHALL stretch to available viewport height
The main Collections table surface SHALL use the available viewport height without causing page-level overflow in normal desktop review dimensions.

#### Scenario: Render single table surface at desktop height
- **GIVEN** an authenticated user opens `/collections` at a desktop viewport
- **WHEN** the Collections workspace renders
- **THEN** the Collections table card MUST consume the available workspace height
- **AND** the shared table surface MUST have usable vertical height
- **AND** the lower members table MUST NOT render
- **AND** the document MUST NOT introduce avoidable page-level vertical overflow

### Requirement UI-SCREEN-COLLECTIONS-021: Collection table cells SHALL truncate long values safely
The single Collections table SHALL prevent long values from overflowing columns while preserving readable clipped text.

#### Scenario: Render long collection table values
- **GIVEN** live inventory records or collection metadata include unusually long text values
- **WHEN** the user opens `/collections`
- **THEN** the Collections table MUST NOT create horizontal table overflow
- **AND** each visible table cell MUST keep its content inside the cell bounds

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
Collections table filters SHALL show deterministic zero-result summaries and empty-row messages while preserving the selected collection context and avoiding profile-settings writes; the retired collection-members filter SHALL remain absent.

#### Scenario: Collections table filter has no matches
- **GIVEN** an authenticated user opens `/collections`
- **WHEN** the user enters a collection filter that matches no collection rows
- **THEN** the collections table MUST show zero matching rows
- **AND** the collections summary MUST show `Showing 0 of <total> collections.`
- **AND** the selected collection context MUST remain unchanged
- **AND** no profile-settings save MUST be sent by filtering alone

#### Scenario: Retired collection members filter is absent
- **GIVEN** an authenticated user opens `/collections`
- **WHEN** the user filters the Collections table
- **THEN** no collection-members filter or empty-row message MUST render
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

### Requirement UI-SCREEN-COLLECTIONS-027: Collections members panel is retired
The Collections members panel SHALL remain absent; selected collection contents are represented by main table row counts and inventory navigation.

#### Scenario: All Items count reflects live inventory without members panel
- **GIVEN** an authenticated user opens `/collections` and live inventory records are available
- **WHEN** the Collections screen loads with `All Items` selected
- **THEN** the `All Items` row count MUST equal the live inventory member count
- **AND** no members panel MUST render

#### Scenario: Selecting a collection updates active context without members panel
- **GIVEN** an authenticated user opens `/collections` with assigned and empty collections available
- **WHEN** the user selects a collection
- **THEN** the active collection context MUST update
- **AND** the persisted active collection setting MUST match the selected collection
- **AND** no members panel MUST render

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
- **WHEN** the user clears the filter to return to the full collections list
- **THEN** the table MUST return to the pagination page containing the selected row

### Requirement UI-SCREEN-COLLECTIONS-029: Collection members pagination is retired
The Collection members table pagination SHALL remain absent because the separate members table is retired from Collections.

#### Scenario: Members pagination is absent
- **GIVEN** an authenticated user opens `/collections`
- **WHEN** the Collections screen loads
- **THEN** no collection-members pagination control MUST render
- **AND** the active collection context MUST remain visible

#### Scenario: Main table pagination remains non-mutating
- **GIVEN** collections require table pagination
- **WHEN** the user changes Collections table pages
- **THEN** the main table page change MUST NOT send a profile-settings save

### Requirement UI-SCREEN-COLLECTIONS-030: Protected default collection SHALL reject rename and delete mutation
The `All Items` default collection SHALL remain a protected management context that cannot be renamed or deleted from row workflows.

#### Scenario: Attempt to rename or delete All Items
- **GIVEN** an authenticated user opens `/collections` with `All Items` selected
- **WHEN** the user attempts to rename `All Items` from the row edit panel
- **THEN** no profile-settings save MUST be sent
- **AND** the `All Items` row MUST remain visible
- **AND** no replacement collection row MUST be created
- **WHEN** the user attempts to delete `All Items`
- **THEN** no profile-settings save MUST be sent
- **AND** the delete confirmation MUST remain blocked from destructive completion
- **AND** the active collection context MUST remain `All Items`

### Requirement UI-SCREEN-COLLECTIONS-031: Collections tables SHALL sort deterministically without passive mutation
The Collections table SHALL expose deterministic column sorting while preserving the current selected collection context and avoiding profile-settings writes; retired collection-members table sorting SHALL remain absent.

#### Scenario: Sort collections and members tables
- **GIVEN** an authenticated user opens `/collections` with multiple collections and collection members visible
- **WHEN** the user sorts the collections table by collection name descending
- **THEN** the collection rows MUST render in deterministic descending name order
- **AND** the active collection context MUST remain unchanged
- **AND** no profile-settings save MUST be sent by sorting alone
- **AND** the retired collection-members table MUST remain absent

### Requirement UI-SCREEN-COLLECTIONS-032: Collection delete SHALL soft-delete and reconcile assigned items
Collections SHALL hide deleted collection rows from active views by default, retain deleted metadata for review, and make assigned-item outcomes explicit before saving.

#### Scenario: Soft-delete and review deleted collection
- **GIVEN** an authenticated user opens `/collections` with a non-protected collection row visible
- **WHEN** the user confirms deletion
- **THEN** the collection MUST be marked deleted rather than removed from persisted collection state
- **AND** active collection tables and selectors MUST hide the deleted collection by default
- **AND** the deleted filter MUST show the deleted row with an explicit deleted state

#### Scenario: Delete populated collection with reassignment choice
- **GIVEN** a non-protected collection has assigned inventory items
- **WHEN** the user opens the delete confirmation
- **THEN** the dialog MUST show the assigned item count
- **AND** the dialog MUST explain that choosing no destination removes only collection membership
- **WHEN** the user chooses another active collection and confirms deletion
- **THEN** the original collection MUST be marked deleted
- **AND** assigned items MUST move to the chosen destination collection
- **WHEN** the user confirms deletion with no destination selected
- **THEN** assigned items MUST remain in Cabinet with no collection assignment and remain visible under `All Items`

#### Scenario: Edit collection metadata
- **GIVEN** a non-protected collection row exists
- **WHEN** the user opens the edit side panel
- **THEN** the panel MUST expose persisted collection metadata fields for name, scope, status, and description
- **AND** saving valid metadata MUST persist and render the updated row metadata

### Requirement UI-SCREEN-COLLECTIONS-033: Collection row actions SHALL be icon-only and accessible
Collections row action controls SHALL use compact icon-only buttons while preserving understandable labels, tooltips, and the existing row action behavior.

#### Scenario: Render collection row actions
- **GIVEN** collection rows are visible in the table
- **WHEN** the row action cell renders
- **THEN** View, Edit, and Delete controls MUST render as icon-only buttons without visible text labels
- **AND** each action control MUST expose an accessible label and tooltip/title naming the action
- **AND** activating a row action MUST keep the existing View, Edit, or Delete workflow behavior
- **AND** double-clicking an explicit row action control MUST NOT trigger the row double-click side-panel handler

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
| UC-COL-11 | Switch active profile | Collections state and members switch to the active profile's persisted settings | `ui.web/cypress/e2e/collections/ui-screen-collections/spec.cy.ts` `UI-SCREEN-COLLECTIONS-011 switches collection state with the active profile` |
| UC-COL-12 | Inventory creates collection | Inventory folder-tree collection create persists and stays available to Collections | `ui.web/cypress/e2e/collections/ui-screen-collections/spec.cy.ts` `UI-SCREEN-COLLECTIONS-012 keeps inventory collection create in the folder tree` |
| UC-COL-13 | Wishlist creates collection | Wishlist table collection create persists into Collections manager state | `ui.web/cypress/e2e/collections/ui-screen-collections/spec.cy.ts` `UI-SCREEN-COLLECTIONS-013 reflects wishlist table collection create inside the collections manager` |
| UC-COL-14 | Rename propagates to Wishlist | Collection rename updates Wishlist compact collection filter options | `ui.web/cypress/e2e/collections/ui-screen-collections/spec.cy.ts` `UI-SCREEN-COLLECTIONS-014 propagates rename into the wishlist table collection filter` |
| UC-COL-15 | Delete propagates to filters | Collection delete removes the option from compact filters and falls back active context | `ui.web/cypress/e2e/collections/ui-screen-collections/spec.cy.ts` `UI-SCREEN-COLLECTIONS-015 removes deleted collections from compact filters` |
| UC-COL-16 | Retire members table | Collections renders one main working table and no lower members table | `ui.web/cypress/e2e/collections/ui-screen-collections/spec.cy.ts` `UI-SCREEN-COLLECTIONS-016 retires the lower members table and keeps the main table useful` |
| UC-COL-17 | Retire members filter | Members table filter and summary are absent | `ui.web/cypress/e2e/collections/ui-screen-collections/spec.cy.ts` `UI-SCREEN-COLLECTIONS-017 retires the members table filter from Collections` |
| UC-COL-18 | All Items live count | All Items row count and members summary match live inventory records | `ui.web/cypress/e2e/collections/ui-screen-collections/spec.cy.ts` `UI-SCREEN-COLLECTIONS-018 keeps All Items count aligned with inventory members` |
| UC-COL-19 | Retired assignment controls | Legacy in-route assignment controls are documented as retired/replaced by table/navigation contracts | retired/replaced by `UI-SCREEN-COLLECTIONS-016`, `017`, `018`, `020`, `021`, and Inventory assignment workflows |
| UC-COL-20 | Viewport-height table | Collections table stretches to available viewport height without page overflow | `ui.web/cypress/e2e/collections/ui-screen-collections/spec.cy.ts` `UI-SCREEN-COLLECTIONS-020 stretches the main table to the available viewport height` |
| UC-COL-21 | Long value truncation | Collections table clips long values inside cells | `ui.web/cypress/e2e/collections/ui-screen-collections/spec.cy.ts` `UI-SCREEN-COLLECTIONS-021 keeps long collection values inside the single table` |
| UC-COL-22 | Press Enter in create dialog | Valid Enter submit persists; invalid Enter shows inline validation | `ui.web/cypress/e2e/collections/ui-screen-collections/spec.cy.ts` `UI-SCREEN-COLLECTIONS-022 submits Create collection with Enter and validates invalid Enter` |
| UC-COL-23 | Shortcut and command create | `Ctrl+N` and command palette entry open the same create dialog and persisted create flow | `ui.web/cypress/e2e/collections/ui-screen-collections/spec.cy.ts` `UI-SCREEN-COLLECTIONS-023 opens Create collection from Ctrl+N and command entry` |
| UC-COL-24 | Zero-result filters | Collections and members filters show deterministic empty states without profile-settings writes | `ui.web/cypress/e2e/collections/collections-filter-empty-states/spec.cy.ts` `UI-SCREEN-COLLECTIONS-024 shows deterministic zero-result filter states without saving settings` |
| UC-COL-25 | Cancel transient workflows | Create/edit/delete cancellation closes transient surfaces without adding, renaming, deleting, changing active context, or saving settings | `ui.web/cypress/e2e/collections/collections-cancel-no-mutation/spec.cy.ts` `UI-SCREEN-COLLECTIONS-025 cancels create edit and delete workflows without saving settings` |
| UC-COL-26 | Row edit side panel | Double-click opens the right-side edit panel, record navigation changes draft context, valid rename persists, and duplicate rename does not save | `ui.web/cypress/e2e/collections/collections-row-side-panel/spec.cy.ts` `UI-SCREEN-COLLECTIONS-026 opens and validates collection row side-panel workflows` |
| UC-COL-27 | Retired members panel | Separate members panel remains absent while main table counts stay useful | `ui.web/cypress/e2e/collections/collections-members-panel/spec.cy.ts` `UI-SCREEN-COLLECTIONS-027 retires the separate members panel from Collections` |
| UC-COL-28 | Paginate collections table | Later-page rows render without passive settings writes; selected later-page context persists across refresh and remains selected after filtering | `ui.web/cypress/e2e/collections/collections-pagination/spec.cy.ts` `UI-SCREEN-COLLECTIONS-028 preserves paginated collection selection without passive settings writes` |
| UC-COL-29 | Retired members pagination | Members-table pagination remains absent | `ui.web/cypress/e2e/collections/collections-members-pagination/spec.cy.ts` `UI-SCREEN-COLLECTIONS-029 retires members-table pagination from Collections` |
| UC-COL-30 | Protect All Items | Rename and delete attempts do not write profile settings, create replacement rows, or leave the active `All Items` context | `ui.web/cypress/e2e/collections/collections-protected-all-items/spec.cy.ts` `UI-SCREEN-COLLECTIONS-030 keeps All Items protected from row rename and delete actions` |
| UC-COL-31 | Sort collections | Collection rows sort deterministically without saving settings or changing the active collection context | `ui.web/cypress/e2e/collections/collections-sorting-no-mutation/spec.cy.ts` `UI-SCREEN-COLLECTIONS-031 sorts collections without passive settings writes` |
| UC-COL-32 | Soft-delete and reconcile collection items | Deleted rows are hidden by default, visible in the deleted filter, assigned items move to a chosen destination or become unassigned, and metadata edits persist | `ui.web/cypress/e2e/collections/ui-screen-collections/spec.cy.ts` `UI-SCREEN-COLLECTIONS-032 soft-deletes with deleted filter, reassignment choices, and editable metadata` |
| UC-COL-33 | Icon-only row actions | Row View/Edit/Delete actions render as compact icon-only accessible controls without changing row workflows | `ui.web/cypress/e2e/collections/collections-row-side-panel/spec.cy.ts` `renders collection row actions as icon-only accessible controls` |
