## Purpose
Define Display settings screen behavior.

## Requirements
### Requirement UI-SCREEN-SETTINGS-DISPLAY-001: Display screen SHALL manage sidebar visibility preferences

#### Scenario: Save display item selection
- **GIVEN** user opens `/settings/display`
- **WHEN** user updates sidebar item selection and saves
- **THEN** runtime MUST persist selected items

### Requirement UI-SCREEN-SETTINGS-DISPLAY-002: Display screen SHALL require minimum one selected item

#### Scenario: Invalid empty selection
- **GIVEN** user deselects all display items
- **WHEN** user submits form
- **THEN** UI MUST block save and show validation requiring at least one selected item
