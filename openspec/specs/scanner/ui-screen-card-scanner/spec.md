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

### Requirement UI-SCREEN-CARD-SCANNER-009: Card Scanner UI SHALL preview scanner review writes before confirmed apply
Card Scanner UI SHALL convert queued quick-scan uploads into scanner recognition review apply requests, request a non-mutating confirmation-required preview first, expose the selected target/candidate/confidence/provenance outcome, and only mark the scan linked after a confirmed apply response plus persistence reload.

#### Scenario: Review and confirm scanner apply from UI
- **GIVEN** a quick-scan upload has recognition candidates and the user selects Inventory or Wishlist as the write target
- **WHEN** the user reviews apply
- **THEN** Card Scanner MUST call scanner review apply with `confirmed=false`
- **AND** the UI MUST show the returned confidence, target, and confirm-before-create state without marking the scan linked
- **WHEN** the user confirms apply
- **THEN** Card Scanner MUST call scanner review apply with `confirmed=true`
- **AND** the UI MUST reload the target item collection and mark the scan linked only after the confirmed response

### Requirement UI-SCREEN-CARD-SCANNER-008: Scanner review apply API SHALL persist only confirmed writes
Scanner review apply API SHALL accept normalized recognition candidates, build the same review preview contract, reject unconfirmed write attempts, and persist confirmed Inventory or Wishlist records with selected candidate, confidence, provenance, manual-override, media, and source metadata retained as auditable item evidence.

#### Scenario: Reject unconfirmed scanner review apply
- **GIVEN** a scanner recognition review has a selected candidate
- **WHEN** the client asks to apply the review without `confirmed=true`
- **THEN** Cabinet MUST return a confirmation-required response and MUST NOT create Inventory or Wishlist records

#### Scenario: Confirm scanner review apply to inventory or wishlist
- **GIVEN** a scanner recognition review has a selected candidate with media and provenance evidence
- **WHEN** the client confirms the apply target
- **THEN** Cabinet MUST create the target Inventory or Wishlist record for the active profile
- **AND** the created item MUST retain scanner media/provenance/confidence evidence and source URL metadata for audit
- **AND** Wishlist targets MUST create both the canonical item and a wishlist entry without marking the item as owned inventory

### Requirement UI-SCREEN-CARD-SCANNER-004: Card Scanner SHALL support deterministic error/retry behavior
Scanner SHALL surface human-readable failure states with retry guidance.

### Requirement UI-SCREEN-CARD-SCANNER-010: Card Scanner SHALL keep failed reads in manual review without linking scans
Card Scanner SHALL keep low-confidence or failed recognition reads in the queued review surface, preserve manual override evidence, and avoid marking scans linked until a confirmed apply succeeds.

#### Scenario: Failed read stays queued for manual review
- **GIVEN** a quick-scan upload is queued and the reviewer selects a manual alternate
- **WHEN** the scanner review preview API returns a failed-read or low-confidence manual-review response
- **THEN** Card Scanner MUST show the failed-read guidance
- **AND** MUST preserve the manual override state
- **AND** MUST keep the scan queued instead of marking it linked or showing a created item result

### Requirement UI-SCREEN-CARD-SCANNER-011: Card Scanner SHALL preserve grading review evidence before scanner writes
Card Scanner SHALL expose the scanner candidate review grading context before any Inventory or Wishlist write, including item type, condition estimate, and grading status, and SHALL include that grading context in the non-mutating review payload.

#### Scenario: Review grading evidence before confirmed apply
- **GIVEN** a quick-scan upload is queued for a trading-card scan
- **WHEN** the reviewer opens the candidate review flow before confirmed apply
- **THEN** Card Scanner MUST show the scan confidence, selected candidate, item type, condition estimate, and grading status
- **AND** the scanner review preview request MUST include the same grading context on each candidate before any confirmed write is requested

### Requirement UI-SCREEN-CARD-SCANNER-012: Card Scanner SHALL support manual-entry intake into scanner review
Card Scanner SHALL allow a reviewer to queue a typed card title as a manual scanner review candidate without camera or upload input, and SHALL keep that manual-entry scan queued and unlinked until the same explicit review/apply confirmation path succeeds.

#### Scenario: Queue manual-entry scan for review
- **GIVEN** user is on `/scanner`
- **WHEN** user queues a non-empty manual card title
- **THEN** Card Scanner MUST add the title to the scanner review queue with candidate suggestions and queued status
- **AND** MUST keep the item unlinked until confirmed apply succeeds
- **WHEN** user attempts to queue an empty manual title
- **THEN** Card Scanner MUST show actionable validation feedback and MUST NOT create a queued scan

### Requirement UI-SCREEN-CARD-SCANNER-007: Scanner recognition review SHALL normalize candidates before writes
Scanner recognition review SHALL normalize candidate payloads into a non-mutating preview that preserves top match, alternates, confidence label, provenance, media evidence, manual override state, target record type, and a required confirm-before-create boundary.

#### Scenario: Normalize scan candidates for review
- **GIVEN** recognition returns one or more candidate matches with confidence, provenance, and media evidence
- **WHEN** Cabinet builds the scanner review preview
- **THEN** Cabinet MUST choose the highest-confidence top candidate, retain alternates, classify confidence, preserve source/provenance/media evidence, and set `confirm_before_create=true` without writing Inventory or Wishlist records

#### Scenario: Preserve manual override in review preview
- **GIVEN** a reviewer manually selects an alternate recognition candidate
- **WHEN** Cabinet builds the scanner review preview
- **THEN** Cabinet MUST keep the automatic top match as evidence, select the manual override, mark manual review as required, preserve the requested Inventory/Wishlist target, and still require explicit confirmation before any create/update write

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
| UC-CS-05 | View recent unlinked scans | Quick-category panel shows most-recent unlinked results in Cards and Table modes | implemented: `ui.web/cypress/e2e/scanner/ui-screen-card-scanner/spec.cy.ts` `UI-SCREEN-CARD-SCANNER-005 shows recent unlinked scans in cards/table with deterministic newest-first ordering` |
| UC-CS-06 | Quick Scan mobile/desktop capture | `Quick Scan` launches rapid capture flow on mobile and desktop with fallback path | implemented: `ui.web/cypress/e2e/scanner/ui-screen-card-scanner/spec.cy.ts` `UI-SCREEN-CARD-SCANNER-006 provides quick-scan action for mobile and desktop with deterministic intake feedback` |
| UC-CS-07 | Apply reviewed scan | API rejects unconfirmed writes and persists confirmed Inventory/Wishlist records with scanner evidence | implemented: `TestScannerRecognitionReviewApplyRequiresConfirmationAndDoesNotMutate`, `TestScannerRecognitionReviewApplyCreatesWishlistItemWithEvidence` (`internal/app/scanner_api_test.go`) |
| UC-CS-09 | UI review and confirmed apply | Quick-scan UI previews scanner review apply, confirms explicit Wishlist/Inventory write, reloads persistence, and then marks scan linked | implemented: `ui.web/cypress/e2e/scanner/ui-screen-card-scanner/spec.cy.ts` `UI-SCREEN-CARD-SCANNER-009 reviews and confirms scanner apply through the API before marking linked` |
| UC-CS-10 | Failed read manual review | Failed/low-confidence review preview preserves manual override evidence and keeps scan queued/unlinked | implemented: `ui.web/cypress/e2e/scanner/ui-screen-card-scanner/spec.cy.ts` `UI-SCREEN-CARD-SCANNER-010 keeps failed reads in manual review without linking the scan` |
| UC-CS-11 | Grading review evidence | Candidate review shows and sends item type, condition estimate, and grading status before confirmed writes | implemented: `ui.web/cypress/e2e/scanner/ui-screen-card-scanner/spec.cy.ts` `UI-SCREEN-CARD-SCANNER-011 preserves grading context in candidate review before writes` |
| UC-CS-12 | Manual-entry intake | Typed card title queues a scanner review candidate without linking or mutating records until confirmed apply | implemented: `ui.web/cypress/e2e/scanner/ui-screen-card-scanner/spec.cy.ts` `UI-SCREEN-CARD-SCANNER-012 queues manual-entry scans for review before writes` |
