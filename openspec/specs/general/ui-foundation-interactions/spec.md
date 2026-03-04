## Purpose
Define cross-screen row, selection, and modal interaction behavior.

## Requirements
### Requirement UI-FOUNDATION-INTERACTIONS-001: Row-to-detail interaction model SHALL be consistent across data screens
Cabinet SHALL open details on non-interactive row click and reserve thumbnail click for media lightbox where applicable.

#### Scenario: Row click details
- **GIVEN** inventory rows are rendered in rows view and a non-interactive row cell is focused
- **WHEN** user single-clicks the row surface cell
- **THEN** row details modal/drawer SHALL open for the selected record

#### Scenario: Thumbnail lightbox
- **GIVEN** inventory photos are available for the selected record
- **WHEN** user clicks a photo thumbnail
- **THEN** fullscreen lightbox SHALL open
- **AND** previous/next actions SHALL navigate within the active photo result order

### Requirement UI-FOUNDATION-INTERACTIONS-002: Bulk mode SHALL be explicit and checkbox-driven
Cabinet SHALL use checkbox controls for selection and SHALL not overload row-click with selection toggles.

#### Scenario: Bulk selection mode
- **GIVEN** a rows view exposes a checkbox column
- **WHEN** user selects one or more row checkboxes
- **THEN** bulk actions toolbar SHALL appear with selected count
- **AND** row cell click SHALL NOT implicitly toggle checkbox selection state

### Requirement UI-FOUNDATION-INTERACTIONS-003: Row and media interactions SHALL support drawer/lightbox/modal model
Cabinet SHALL use split interaction behavior across inventory and integrations workspaces.

#### Scenario: Row and double-click behavior
- **GIVEN** a list/table view is rendered and row is not in bulk selection mode
- **WHEN** user single-clicks row surface
- **THEN** details drawer/modal SHALL open and URL state SHALL preserve selected record context

#### Scenario: Double-click edit behavior
- **GIVEN** row details are available for current record
- **WHEN** user double-clicks row
- **THEN** edit modal SHALL open with previous/next navigation over current filtered result set
- **AND** double-click MUST NOT trigger unrelated launch/publish/delete actions or error-toasts

