## MODIFIED Requirements

### Requirement: Cabinet UI SHALL provide shared table and dialog primitives
Cabinet MUST expose reusable UI primitives for table interaction, row-level
record actions, and CRUD dialogs so feature surfaces do not implement
incompatible copies of the same behavior.

#### Scenario: Shared record actions are capability-driven
- **GIVEN** a table row represents a Cabinet record
- **WHEN** the surface provides row-level actions
- **THEN** the surface SHALL use the shared record action menu contract instead
  of page-specific row action copies
- **AND** the trigger SHALL be an icon-only kebab action in the final table
  action column with an accessible label and tooltip
- **AND** supported actions SHALL follow the standard order: View/Open, Edit,
  Duplicate, Archive/Delete, Restore, Permanent delete
- **AND** actions SHALL be driven by the record's capabilities and permissions
  rather than by static page assumptions
- **AND** unsupported operations SHALL be omitted or truthfully disabled with a
  short reason, never displayed as functional when no handler exists
- **AND** bulk actions SHALL remain in the table toolbar rather than inside
  individual row menus

#### Scenario: Shared CRUD dialogs expose consistent structure
- **GIVEN** a surface creates, edits, duplicates, archives, deletes, restores,
  or permanently deletes a Cabinet record
- **WHEN** the user opens the corresponding dialog
- **THEN** the dialog SHALL use the shared CRUD or destructive confirmation
  contract rather than page-specific modal behavior
- **AND** create/edit dialogs SHALL provide a title, description, icon,
  validation area, server-error area, cancel/close behavior, dirty-state
  protection, submitting state, and double-submit prevention
- **AND** destructive confirmations SHALL name the record, describe the
  consequence, and distinguish archive or soft delete from permanent deletion
- **AND** the shared contract SHALL support domain-equivalent labels such as
  Archive, Remove, Restore, Disable, or Revoke when plain Delete is inaccurate
