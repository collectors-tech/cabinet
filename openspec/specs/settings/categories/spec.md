## Purpose
Define Categories settings screen behavior for reusable inventory taxonomy controls.

## Requirements
### Requirement UI-SCREEN-SETTINGS-CATEGORIES-001: Categories screen SHALL persist reusable taxonomy settings

#### Scenario: Save taxonomy settings
- **GIVEN** user opens `/settings/categories`
- **WHEN** user adds or removes inventory categories, packaging grades, or item type condition scales and saves
- **THEN** runtime MUST persist the taxonomy settings for the active profile
- **AND** screen MUST render deterministic success feedback only after persistence succeeds

### Requirement UI-SCREEN-SETTINGS-CATEGORIES-002: Categories screen SHALL preserve editable taxonomy values on save failure
When taxonomy settings save fails, the screen SHALL show deterministic error feedback, keep the edited category, packaging-grade, and item-type values available for retry, and avoid displaying success feedback until persistence succeeds.

#### Scenario: Taxonomy save failure keeps retry context
- **GIVEN** user changes categories, packaging grades, and item type condition scales
- **WHEN** runtime rejects the taxonomy settings update
- **THEN** UI MUST render deterministic save-error feedback
- **AND** edited taxonomy values MUST remain visible for retry
- **AND** success feedback MUST NOT render

## Use-Case IDs and E2E Mapping
| UC ID | Flow | Expected Result | E2E Mapping |
| --- | --- | --- | --- |
| UC-SET-CAT-01 | Save taxonomy settings | Categories, packaging grades, and item type condition scales persist for the active profile | `ui.web/cypress/e2e/settings/categories/spec.cy.ts` `UI-SCREEN-SETTINGS-CATEGORIES-001 manages reusable taxonomy settings for the active profile` |
| UC-SET-CAT-02 | Taxonomy save failure | Save failure shows error feedback and preserves edited taxonomy values | `ui.web/cypress/e2e/settings/categories/spec.cy.ts` `UI-SCREEN-SETTINGS-CATEGORIES-002 preserves taxonomy edits when save fails` |
