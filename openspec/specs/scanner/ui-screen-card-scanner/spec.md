## Purpose
Define Card Scanner behavior for camera/photo scanning, recognition, and inventory/wishlist record creation.

## Requirements
### Requirement UI-SCREEN-CARD-SCANNER-001: Card Scanner SHALL support camera and photo-upload capture
Card Scanner SHALL allow live camera capture and file upload for card recognition.

### Requirement UI-SCREEN-CARD-SCANNER-006: Card Scanner SHALL provide quick-scan capture on mobile and desktop
Card Scanner SHALL expose a `Quick Scan` action for fast capture workflows on both mobile and desktop surfaces.

#### Scenario: Quick Scan on mobile
- **GIVEN** user opens scanner on mobile
- **WHEN** user taps `Quick Scan`
- **THEN** camera capture flow MUST open immediately for rapid photo capture and queue scan processing

#### Scenario: Quick Scan on desktop
- **GIVEN** user opens scanner on desktop
- **WHEN** user clicks `Quick Scan`
- **THEN** app MUST offer camera capture (if available) and immediate upload fallback for rapid scan intake
- **AND** quick-scan entry MUST remain reachable by keyboard (Tab + Enter/Space)

#### Scenario: Capture or upload card image
- **GIVEN** user is on `/scanner`
- **WHEN** user captures a photo or uploads an image
- **THEN** image MUST be ingested for recognition workflow

### Requirement UI-SCREEN-CARD-SCANNER-002: Card Scanner SHALL perform recognition with confidence and candidate list
Recognition SHALL return top match, alternates, and confidence score with manual override support.

#### Scenario: Review recognition candidates
- **GIVEN** scan completes
- **WHEN** candidates are returned
- **THEN** UI MUST show top candidate, alternates, and confidence with manual selection option

### Requirement UI-SCREEN-CARD-SCANNER-003: Card Scanner SHALL support confirm-before-create for inventory/wishlist writes
Scanner SHALL never silently mutate data; user confirmation is required before creating/updating records.

#### Scenario: Confirm scan result to inventory
- **GIVEN** recognized card candidate is selected
- **WHEN** user confirms create/update target (inventory or wishlist)
- **THEN** runtime MUST persist record with media linkage and show auditable outcome

### Requirement UI-SCREEN-CARD-SCANNER-004: Card Scanner SHALL support deterministic error/retry behavior
Scanner SHALL surface human-readable failure states with retry guidance.

### Requirement UI-SCREEN-CARD-SCANNER-005: Scanner quick-category panel SHALL support card-list and table views for unlinked recent scans
Scanner SHALL provide a quick-category area showing most recently added scan results not yet linked to inventory, with toggleable `Cards` and `Table` views.

#### Scenario: Review recent unlinked scans in quick-category views
- **GIVEN** recent scan results exist and are not yet linked to inventory records
- **WHEN** user opens quick-category panel and toggles `Cards`/`Table`
- **THEN** scanner MUST render the same unlinked recent set in selected view mode with deterministic ordering by most-recent first
- **AND** linked inventory records MUST NOT appear in the quick-category dataset

#### Scenario: Recognition failure
- **GIVEN** recognition request fails or low-confidence ambiguity occurs
- **WHEN** scanner returns failure/ambiguous result
- **THEN** UI MUST show actionable retry/manual-selection guidance

## Use-Case IDs and E2E Mapping
| UC ID | Flow | Expected Result | E2E Mapping |
| --- | --- | --- | --- |
| UC-CS-01 | Capture/upload image | Image accepted for recognition | planned: `ui.web/cypress/e2e/scanner/ui-screen-card-scanner/spec.cy.ts` `card-scanner-capture-upload` |
| UC-CS-02 | Review candidates | Confidence + alternates + override shown | planned: `ui.web/cypress/e2e/scanner/ui-screen-card-scanner/spec.cy.ts` `card-scanner-candidate-review` |
| UC-CS-03 | Confirm create/update | Confirm-before-apply write with media linkage | planned: `ui.web/cypress/e2e/scanner/ui-screen-card-scanner/spec.cy.ts` `card-scanner-confirm-write` |
| UC-CS-04 | Recognition error | Retry/manual guidance shown | planned: `ui.web/cypress/e2e/scanner/ui-screen-card-scanner/spec.cy.ts` `card-scanner-error-guidance` |
| UC-CS-05 | View recent unlinked scans | Quick-category panel shows most-recent unlinked results in Cards and Table modes | planned: `ui.web/cypress/e2e/scanner/ui-screen-card-scanner/spec.cy.ts` `card-scanner-recent-unlinked-cards-table` |
| UC-CS-06 | Quick Scan mobile/desktop capture | `Quick Scan` launches rapid capture flow on mobile and desktop with fallback path | planned: `ui.web/cypress/e2e/scanner/ui-screen-card-scanner/spec.cy.ts` `card-scanner-quick-scan-mobile-desktop` |
