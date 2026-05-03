## MODIFIED Requirements

### Requirement: Inventory browser controls SHALL support filter/sort/view switching
Inventory browser controls SHALL expose Category, Condition, View, and list/card toggles with deterministic state changes. Category and Condition filters SHALL use the shared compact filter pattern.

#### Scenario: Filter and view controls available
- **GIVEN** user is on `/inventory`
- **WHEN** collection browser renders
- **THEN** `Category`, `Condition`, and `View` controls MUST be available and operable
- **AND** `Rows`/`Cards` mode toggles MUST switch presentation consistently

#### Scenario: Category filter uses multi-value item category metadata
- **GIVEN** inventory items have category metadata
- **WHEN** user applies a Category filter
- **THEN** the inventory table MUST only show rows matching that category

#### Scenario: Condition filter uses instance condition metadata
- **GIVEN** inventory items have instance condition metadata
- **WHEN** user applies a Condition filter
- **THEN** the inventory table MUST only show rows matching that condition

### Requirement: Inventory item forms SHALL separate item type from category
Inventory create and edit forms SHALL expose Item Type as a single-select field and Category as flexible multi-select metadata.

#### Scenario: Item type controls condition choices
- **GIVEN** user opens an inventory create or edit form
- **WHEN** user selects an Item Type
- **THEN** the condition field MUST show values from that Item Type condition scale

#### Scenario: Category remains flexible
- **GIVEN** user opens an inventory create or edit form
- **WHEN** user adds or selects category values
- **THEN** category values MUST be saved independently from Item Type
