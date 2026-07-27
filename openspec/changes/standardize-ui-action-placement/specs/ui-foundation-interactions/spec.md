## ADDED Requirements

### Requirement: Action placement interaction consistency

Moved Cabinet actions MUST preserve keyboard access, accessible names, loading
and disabled states, focus behavior, responsive layout, and user workflow state.

#### Scenario: Header actions overflow accessibly

- **GIVEN** a page header has more actions than fit at a narrow viewport
- **WHEN** the viewport width is constrained
- **THEN** secondary actions move into an accessible overflow menu
- **AND** the primary action remains reachable by keyboard and screen-reader
  users
- **AND** action text, icons, and shell utilities do not overlap.

#### Scenario: Action order and labels are stable

- **GIVEN** a page exposes multiple actions in a canonical region
- **WHEN** the user scans or tabs through the region
- **THEN** actions follow the documented order for that region
- **AND** each action has a stable icon, visible label or tooltip, accessible
  name, loading state, disabled state, and test ID.

#### Scenario: Moving actions preserves workflow state

- **GIVEN** a user has unsaved form state, active filters, selected rows, or
  visible error feedback
- **WHEN** an action is moved into its canonical region
- **THEN** triggering, cancelling, retrying, or completing that action preserves
  the same state semantics that existed before the move.

#### Scenario: Duplicate page content is removed after parity

- **GIVEN** a route header already provides the page title, description, icon,
  and whole-page action region
- **WHEN** local page content repeats the same title, description, or action
  controls
- **THEN** the duplicated local block is removed after tests prove behavior
  parity
- **AND** no user loses keyboard, pointer, or screen-reader access to the
  action.
