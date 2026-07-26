## MODIFIED Requirements

### Requirement: Cabinet UI interactions SHALL be keyboard and pointer safe
Cabinet MUST ensure shared UI interactions behave predictably with mouse,
keyboard, touch, and assistive technology across reusable components.

#### Scenario: Row action menus do not trigger row navigation
- **GIVEN** a table row supports row click, row selection, or detail navigation
- **WHEN** the user opens or activates the row action menu with pointer or
  keyboard input
- **THEN** the menu interaction SHALL NOT trigger row navigation or selection
- **AND** closing the menu or any launched dialog SHALL return focus to the row
  action trigger when the trigger is still mounted
- **AND** menu items SHALL be reachable with keyboard navigation and announce a
  clear accessible name

#### Scenario: Dialog focus and submission states are stable
- **GIVEN** a user opens a shared create/edit or destructive confirmation dialog
- **WHEN** the dialog is active
- **THEN** focus SHALL be trapped inside the dialog
- **AND** initial focus SHALL land on the first safe actionable control or
  validation target
- **AND** cancel/close SHALL warn before discarding unsaved changes when the
  form is dirty
- **AND** submit controls SHALL expose loading state and prevent duplicate
  submissions while preserving actionable server errors
- **AND** closing the dialog SHALL restore focus to the invoking control when
  possible
