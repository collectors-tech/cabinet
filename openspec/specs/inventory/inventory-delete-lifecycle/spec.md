## Purpose
Define soft-delete, recycle-bin, restore, and permanent-delete rules for inventory records.

## Requirements
### Requirement INVENTORY-DELETE-LIFECYCLE-001: Delete action SHALL transition records through soft-delete states
Cabinet SHALL implement soft-delete lifecycle using explicit statuses before permanent deletion.

#### Scenario: Soft delete active item
- **GIVEN** item status is not `deleted` and item is visible in active inventory list
- **WHEN** user triggers delete action
- **THEN** item status MUST transition to `deleted` and item MUST be hidden from default active list views

### Requirement INVENTORY-DELETE-LIFECYCLE-002: Second delete SHALL move deleted records to recycle state
Cabinet SHALL move already-deleted records to recycle state on subsequent delete action.

#### Scenario: Move to recycle
- **GIVEN** item status is `deleted`
- **WHEN** user triggers delete action again
- **THEN** item status MUST transition to `recycle` and item MUST appear in recycle list

### Requirement INVENTORY-DELETE-LIFECYCLE-003: Permanent delete SHALL block linked records and provide restore option
Cabinet SHALL block hard delete when linked dependencies exist and SHALL support restore from deleted/recycle states.

#### Scenario: Hard delete with links
- **GIVEN** recycle item has linked dependencies
- **WHEN** user attempts permanent delete
- **THEN** API MUST return `409` with dependency list and item MUST remain recoverable

#### Scenario: Restore from recycle
- **GIVEN** item status is `deleted` or `recycle`
- **WHEN** user selects restore
- **THEN** item MUST return to active status and reappear in active inventory list
