## Purpose
Define Inventory Barcodes screen behavior for add, lookup, and variant-safe resolution.

## Requirements
### Requirement: Inventory Barcodes SHALL support add and local lookup workflows
The screen SHALL support item barcode add and local lookup display.

#### Scenario: Add and lookup barcode
- **WHEN** user adds barcode and performs lookup
- **THEN** barcode list and lookup result SHALL update

### Requirement: Inventory Barcodes SHALL support external fallback search
If local matches are absent, screen SHALL provide external search pathway.

#### Scenario: No local match fallback
- **WHEN** local lookup has no results
- **THEN** external search action SHALL be available

### Requirement: Inventory Barcodes SHALL support deterministic state handling
The screen SHALL support loading, empty, error, and ready states.

#### Scenario: Barcode lookup error
- **WHEN** barcode lookup endpoint fails
- **THEN** screen SHALL show actionable error state with retry

## Acceptance Criteria
- UC IDs cover add, lookup, fallback, and error behavior.
- E2E mappings defined for local and external resolution paths.

## Success Criteria
- Barcode workflows complete without route errors.
- Users can always proceed from no-match state.

## Use-Case IDs and E2E Mapping
| UC ID | Flow | Expected Result | E2E Mapping |
| --- | --- | --- | --- |
| UC-BAR-01 | Add barcode to item | Barcode attached and listed | planned: `cypress/e2e/ui/barcodes.cy.ts` `barcode-add` |
| UC-BAR-02 | Local lookup | Local matches returned when present | planned: `cypress/e2e/ui/barcodes.cy.ts` `barcode-local-lookup` |
| UC-BAR-03 | No local match | External search action appears | planned: `cypress/e2e/ui/barcodes.cy.ts` `barcode-external-fallback` |
| UC-BAR-04 | Lookup failure | Error + retry shown | planned: `cypress/e2e/ui/barcodes.cy.ts` `barcode-error-state` |
