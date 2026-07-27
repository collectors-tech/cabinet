## ADDED Requirements

### Requirement: Authenticated page action regions

Authenticated Cabinet pages MUST place actions in canonical regions so users can
predict whether an action affects the whole page, a table/list, one record, or
the active dialog.

#### Scenario: Whole-page actions live in the global page header

- **GIVEN** an authenticated route has a create, add, import, export, refresh,
  run, invite, backup, restore, or other action that applies to the whole page
- **WHEN** the route renders its page chrome
- **THEN** the action is available from the global page header action region
- **AND** there is no duplicate implementation of the same whole-page action in
  page body title blocks or unrelated toolbars.

#### Scenario: List controls remain in list toolbars

- **GIVEN** a page renders a table, grid, list, or repeated record collection
- **WHEN** the page provides search, filters, sort, view switch, selection,
  bulk actions, or table-scoped refresh
- **THEN** those controls remain in the relevant table/list toolbar
- **AND** they do not move into the global page header unless the action applies
  to the whole page.

#### Scenario: Record operations use the shared row menu

- **GIVEN** a repeated record row supports view, edit, duplicate, archive,
  restore, delete, handoff, validate, run, pause, resume, download, or other
  single-record operations
- **WHEN** the row renders its action affordance
- **THEN** those operations are exposed through the shared record action menu
  contract from #1938/#1939
- **AND** unsupported operations are omitted unless a disabled state provides a
  truthful short reason.

#### Scenario: Dialog actions stay in the active dialog footer

- **GIVEN** a modal, sheet, drawer, or confirmation dialog is active
- **WHEN** the user needs to cancel, confirm, apply, save, delete, archive, or
  restore within that dialog
- **THEN** those controls are presented in the dialog footer
- **AND** page header or table toolbar actions do not duplicate the same active
  dialog confirmation.

#### Scenario: Pages without primary actions remain balanced

- **GIVEN** an authenticated page has no meaningful whole-page action
- **WHEN** the global page header renders
- **THEN** the page does not add artificial placeholder buttons
- **AND** the layout remains visually balanced with shell utilities kept
  separate from page actions.
