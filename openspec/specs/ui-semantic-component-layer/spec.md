## Purpose
Define the semantic UI component layering contract so Cabinet UI is composable, testable, and implementation-independent from low-level UI libraries.

## Requirements
### Requirement: Semantic component architecture SHALL be structured into L0-L4 layers
Cabinet UI SHALL follow a layered model:
- L0 tokens/primitives
- L1 app shell
- L2 workspace components
- L3 domain blocks
- L4 overlays/interactions

#### Scenario: Component placement review
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** a new UI component is introduced
- **THEN** it SHALL be assigned to a valid semantic layer with clear ownership boundaries

### Requirement: L0 primitives SHALL provide shared visual and state consistency
L0 primitives SHALL include page/card/form/status primitives and SHALL define common styling and state conventions.

#### Scenario: Primitive reuse
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** a new screen section is implemented
- **THEN** it SHALL compose from L0 primitives rather than duplicating raw style behavior

### Requirement: L1 app shell SHALL own global layout responsibilities
L1 SHALL define app shell, primary nav, context pane, sticky page header, and main content scroll ownership.

#### Scenario: Shell behavior validation
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** user navigates across top-level screens
- **THEN** shell ownership (fixed nav, sticky header, body scroll) SHALL remain stable

### Requirement: L2 workspace components SHALL model screen-level workflows with deterministic states
L2 workspaces SHALL expose loading, empty, error, and ready states for all data-bound regions.

#### Scenario: Workspace data state transition
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** data fetch resolves from loading to ready or error
- **THEN** workspace SHALL transition through deterministic state rendering behavior

### Requirement: L3 domain blocks SHALL be reusable across workspaces
L3 blocks (toolbars, summary strips, grids/tables, media pickers, status panels) SHALL be reusable and context-configurable.

#### Scenario: Domain block reuse
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** inventory and wishlist need shared table behavior
- **THEN** both SHALL consume the same semantic table block contract with contextual props

### Requirement: L4 overlays SHALL enforce interaction safety and focus behavior
L4 overlays (dialogs, drawers, command palette, chat rail) SHALL enforce focus management, explicit open/close controls, and destructive action guardrails.

#### Scenario: Confirm dialog guardrail
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** user initiates destructive action
- **THEN** confirm dialog SHALL require explicit confirmation before mutation execution

### Requirement: Semantic layer SHALL enforce accessibility contract
All layers SHALL comply with landmark semantics, input labeling, keyboard navigation, and non-color-only state signaling.

#### Scenario: Keyboard-only workflow
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** user navigates shell and overlays without mouse
- **THEN** all primary interactions SHALL remain reachable and operable

## Acceptance Criteria
1. L0-L4 contract requirements are explicit and scenario-testable.
2. Accessibility and deterministic state model requirements are explicitly defined.
3. Semantic wrapper rule is explicit (business/screen code uses semantic components, not raw primitive imports directly).
4. Delivery sequence and incremental extraction path are represented as enforceable requirements.

## Success Criteria
1. UI implementation remains consistent across screens as features expand.
2. Refactoring from page-level monoliths to semantic modules does not regress shell behavior.
3. New engineers can place/implement components correctly without external guidance.

## E2E and Test Implications
Minimum required test coverage tied to this spec:
- Shell stability tests: fixed nav, sticky header, scroll ownership
- Workspace state tests: loading/empty/error/ready per L2 workspace
- Overlay tests: dialog/drawer focus trap and close behavior
- Reuse tests: shared L3 block behavior in at least two screens

## Source Mapping
- Semantic component definitions are canonically captured in this spec.
