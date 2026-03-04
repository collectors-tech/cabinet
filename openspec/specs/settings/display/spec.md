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

### Requirement UI-SCREEN-SETTINGS-DISPLAY-003: Display screen SHALL expose Retry action for fetch/bootstrap failures

#### Scenario: Retry display load failure
- **GIVEN** display section is in error state after failed fetch/bootstrap
- **WHEN** user clicks `Retry`
- **THEN** display section MUST re-attempt load and render deterministic ready/empty/error state

### Requirement UI-SCREEN-SETTINGS-DISPLAY-004: Display screen SHALL expose Clear selection and Update display actions

#### Scenario: Clear selection action
- **GIVEN** display items are currently selected
- **WHEN** user clicks `Clear selection`
- **THEN** UI MUST clear selected items and keep form editable without route transition

#### Scenario: Update display action
- **GIVEN** display form selection satisfies validation rules
- **WHEN** user clicks `Update display`
- **THEN** runtime MUST persist display preferences and render deterministic success feedback

## Use-Case IDs and E2E Mapping
| UC ID | Flow | Expected Result | E2E Mapping |
| --- | --- | --- | --- |
| UC-SET-DIS-01 | Retry display load failure | `Retry` re-attempts display fetch deterministically | planned: `ui.web/cypress/e2e/settings/display/spec.cy.ts` `settings-display-retry` |
| UC-SET-DIS-02 | Clear selection action | `Clear selection` clears selected display items | planned: `ui.web/cypress/e2e/settings/display/spec.cy.ts` `settings-display-clear-selection` |
| UC-SET-DIS-03 | Update display action | `Update display` persists display selections | planned: `ui.web/cypress/e2e/settings/display/spec.cy.ts` `settings-display-update` |
