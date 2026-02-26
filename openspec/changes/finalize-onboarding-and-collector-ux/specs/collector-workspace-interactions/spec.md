## ADDED Requirements

### Requirement: Inventory and Wishlist SHALL share a standardized interaction model
The system SHALL use the same interaction behavior across inventory and wishlist data views.

#### Scenario: Row click behavior
- **WHEN** a user clicks a non-interactive row area
- **THEN** the system SHALL open a details drawer for the selected record

#### Scenario: Thumbnail behavior
- **WHEN** a user clicks a record thumbnail
- **THEN** the system SHALL open a lightbox with adjacent navigation across current filtered/sorted dataset order

#### Scenario: Bulk mode behavior
- **WHEN** one or more checkboxes are selected
- **THEN** row interactions SHALL prioritize selection semantics and SHALL not trigger details open by default

### Requirement: Deleted-state lifecycle SHALL be visible and recoverable
The system SHALL support `Deleted` visibility through filtering and SHALL provide restore actions consistent with recycle policy.

#### Scenario: Deleted item restoration
- **WHEN** a user filters to deleted records and selects restore
- **THEN** the record SHALL return to active lists with original metadata intact
