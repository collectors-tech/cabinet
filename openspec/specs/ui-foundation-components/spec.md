## Purpose
Define normative UI component contracts so any engineer can implement Cabinet screens consistently without external knowledge.

## Requirements
### Requirement: Every foundation component family SHALL define explicit contract surface
Each component family SHALL define required inputs, outputs, state model, accessibility requirements, and deterministic error behavior.

#### Scenario: Implementing a new component instance
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** an engineer implements a Button, Input, Select, Dialog, Drawer, or domain component
- **THEN** implementation SHALL follow the contract shape for inputs, outputs, states, and accessibility

### Requirement: Foundation components SHALL support loading, empty, error, and ready states where data-bound
Data-bound components SHALL expose deterministic rendering behavior for all primary operational states.

#### Scenario: Data-bound component enters error state
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** a data request fails for a component
- **THEN** the component SHALL render explicit error state with retry-capable action where applicable

### Requirement: Mutating component flows SHALL prevent double submit and race actions
Mutating actions in forms and action panels SHALL enforce busy-state locking and idempotent intent handling.

#### Scenario: Repeated submit clicks
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** user clicks submit repeatedly while request is in-flight
- **THEN** component SHALL block duplicate submissions until current request resolves

### Requirement: Accessibility behavior SHALL be non-optional for foundation components
Inputs, dialogs, drawers, and action controls SHALL meet keyboard and semantic accessibility behavior.

#### Scenario: Keyboard-only dialog interaction
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** user opens dialog with keyboard and exits with Escape
- **THEN** focus SHALL be trapped while open and restored to trigger on close

### Requirement: Component-level contracts SHALL include testability artifacts
Each major component family SHALL define acceptance criteria, success criteria, and test mapping requirements.

#### Scenario: Component readiness review
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** a component family is marked implementation-ready
- **THEN** acceptance and success criteria plus test mapping SHALL be present in spec and linked to automated tests

## Acceptance Criteria
### Global primitives
- Button contract covers: variants, sizes, disabled/loading behavior, keyboard activation, focus-visible state.
- Input contract covers: label/aria requirements, invalid state messaging, blur/change behavior.
- Select contract covers: option change state and accessible labeling.
- Dialog/Drawer contracts cover: focus trap, escape close, and trigger focus restore.

### Domain component groups
- Home components define action emission and card-state behavior.
- Inventory components define table/form/detail interaction boundaries.
- Photos/Barcodes/AI components define domain-specific state and mutation constraints.
- Scanner/Discover/Reports/Settings components define workflow-specific action contracts.

## Success Criteria
1. Any new engineer can implement a component from this spec without referencing external notes.
2. Critical component regressions are caught by contract tests before merge.
3. Accessibility regressions are detectable by deterministic keyboard/semantic assertions.
4. Data-bound components consistently render loading/empty/error/ready states.

## E2E and Integration Test Mapping Requirements
### Required mapping model
- Each component family SHALL map to:
  - at least one component-level automated test (unit/integration where applicable)
  - at least one E2E assertion in a user workflow

### Minimum E2E coverage expectations
- Dialog/Drawer: open/close/focus restore/escape behaviors.
- Inventory table row interactions: row click vs checkbox selection vs interactive control guard.
- Form submit flows: submit lock while pending and inline error behavior.
- Scanner and Discover row actions: correct payload/action routing and retry flows.
- Settings maintenance flows: guarded destructive/restore actions.

## Data Profiles (Sample and Bulk)
### Sample profile (developer sanity)
- 20 inventory rows
- 5 scanner candidates
- 3 wishlist entries
- 2 backups
Use for rapid component behavior checks.

### Bulk profile (table/grid stress)
- 10,000 inventory items
- 50,000 instances
- 200,000 scanner candidates
- 2,000 tracked pricing rows
Use for pagination, virtualization, and interaction latency validation.

## Notes for Implementation
- Component contracts in this spec are the canonical baseline for implementation and tests.
- This capability is normative for all current and future `ui-screen-*` specs.
