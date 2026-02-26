## Purpose
Define universal interaction, selection, modal, and accessibility behavior across UI screens.

## Requirements
### Requirement: Row-to-detail interaction model SHALL be consistent across data screens
Cabinet SHALL open details on non-interactive row click and reserve thumbnail click for media lightbox where applicable.

#### Scenario: Row click details
- **WHEN** user clicks non-interactive row area
- **THEN** details drawer or modal SHALL open for selected record

#### Scenario: Thumbnail lightbox
- **WHEN** user clicks thumbnail in media-bearing tables
- **THEN** lightbox SHALL open with previous and next navigation in active result order

### Requirement: Bulk mode SHALL be explicit and checkbox-driven
Cabinet SHALL use checkbox controls for selection and SHALL not overload row-click with selection toggles.

#### Scenario: Bulk selection mode
- **WHEN** user selects one or more row checkboxes
- **THEN** selection state SHALL update and bulk action controls SHALL appear

### Requirement: Modal and drawer components SHALL satisfy focus and keyboard contracts
Cabinet SHALL trap and restore focus, support escape close, and provide keyboard-accessible primary actions.

#### Scenario: Dialog keyboard behavior
- **WHEN** modal is open and user presses Escape
- **THEN** modal SHALL close and focus SHALL return to trigger

### Requirement: Accessibility semantics SHALL be non-optional for core workflows
Cabinet SHALL provide explicit labels, landmark roles, and non-color-only status indicators.

#### Scenario: Keyboard-only completion
- **WHEN** user navigates core workflow without mouse
- **THEN** controls SHALL remain reachable and actionable by keyboard
