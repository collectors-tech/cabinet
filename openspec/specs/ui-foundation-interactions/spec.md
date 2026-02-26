## Purpose
Define cross-screen row, selection, and modal interaction behavior.

## Requirements
### Requirement UI-FOUNDATION-INTERACTIONS-001: Row-to-detail interaction model SHALL be consistent across data screens
Cabinet SHALL open details on non-interactive row click and reserve thumbnail click for media lightbox where applicable.

#### Scenario: Row click details
- **GIVEN** data row contains non-interactive surface area
- **WHEN** user clicks row surface
- **THEN** details drawer or modal SHALL open for selected record

#### Scenario: Thumbnail lightbox
- **GIVEN** media-bearing row contains thumbnail
- **WHEN** user clicks thumbnail
- **THEN** lightbox SHALL open with previous/next navigation in active result order

### Requirement UI-FOUNDATION-INTERACTIONS-002: Bulk mode SHALL be explicit and checkbox-driven
Cabinet SHALL use checkbox controls for selection and SHALL not overload row-click with selection toggles.

#### Scenario: Bulk selection mode
- **GIVEN** row checkboxes are visible
- **WHEN** user selects one or more checkboxes
- **THEN** selection state SHALL update and bulk action controls SHALL appear

