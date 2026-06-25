## Purpose
Define the first deliverable slice for unified commerce reconciliation across intent states, purchases, and expected arrivals.

## Requirements
### Requirement COMMERCE-RECONCILIATION-001: Commerce lifecycle SHALL persist profile-scoped intent and purchase states
Cabinet SHALL persist profile-scoped lifecycle entries for `wishlist`, `watchlist`, `cart`, `offer`, and `purchase` states linked to canonical items.

#### Scenario: Record lifecycle state
- **GIVEN** an authenticated user has an active profile and a canonical item exists
- **WHEN** the user submits `POST /api/commerce/lifecycle` with `item_id`, `state`, and optional source/order metadata
- **THEN** Cabinet MUST return `201` with a persisted lifecycle entry linked to the active profile.

### Requirement COMMERCE-RECONCILIATION-002: Purchases SHALL create expected arrivals instead of inventory items directly
Cabinet SHALL model purchases as expected-arrival work first, not immediate inventory reconciliation.

#### Scenario: Purchase creates expected arrival
- **GIVEN** an authenticated user records a `purchase` lifecycle entry for a canonical item
- **WHEN** Cabinet persists the purchase
- **THEN** Cabinet MUST also create an `expected_arrival` record linked to that purchase with status `expected`.

### Requirement COMMERCE-RECONCILIATION-003: Expected arrivals SHALL support reconciliation state transitions
Cabinet SHALL support `expected`, `delivered`, `reconciled`, and `cancelled` expected-arrival states.

#### Scenario: Reconcile delivered purchase
- **GIVEN** an expected arrival exists for the active profile
- **WHEN** the user updates the arrival status to `reconciled` with a delivered date and optional instance reference
- **THEN** Cabinet MUST persist the updated reconciliation state and expose it through `GET /api/commerce/arrivals`.

#### Scenario: Track delivered and cancelled purchase arrival states
- **GIVEN** a purchase lifecycle entry has created an expected arrival for the active profile
- **WHEN** the user updates the arrival status to `delivered` or `cancelled` with the available delivery, instance, or cancellation notes
- **THEN** Cabinet MUST persist the selected state, keep the purchase-linked arrival queryable by item and status, and return the recorded delivery or cancellation evidence through `GET /api/commerce/arrivals`.

### Requirement COMMERCE-RECONCILIATION-004: Purchases SHALL preserve source-backed forwarder provenance
Cabinet SHALL treat forwarder imports as source/provenance evidence for purchase reconciliation rather than as a disconnected primary package workflow.

#### Scenario: Preserve forwarder source evidence for purchase matching
- **GIVEN** an authenticated user has an active profile and Cabinet imports a Stackry or freight-forwarder record with package, sender, shipment, tracking, import-time, and raw-source evidence
- **WHEN** Cabinet stores the imported evidence for reconciliation
- **THEN** Cabinet MUST keep the forwarder source/provenance metadata traceable to purchase and expected-arrival candidates, including source/provider, package or reference id, sender/source text, import timing, raw/source payload reference when available, and the current match-review state.

### Requirement COMMERCE-RECONCILIATION-005: Purchases SHALL expose forwarder match review states
Cabinet SHALL expose forwarder-backed purchase match state from the Purchases review surface so users can inspect unmatched, suggested, confirmed, and rejected or ignored source matches without losing provenance.

#### Scenario: Review and decide a source-backed purchase match
- **GIVEN** a profile has a purchase or expected-arrival candidate and forwarder source evidence that is unmatched, suggested, confirmed, or rejected
- **WHEN** the user reviews the Purchases surface
- **THEN** Cabinet MUST show the match state, enough source evidence to inspect the suggested match, and controls or follow-through paths to confirm a good match or reject/ignore a bad match while preserving both purchase evidence and forwarder provenance.

### Requirement COMMERCE-RECONCILIATION-006: Purchases SHALL present a first-class unified table and add/import entry point
Cabinet SHALL expose a first-class authenticated `/purchases` route, label the purchase-management route as Purchases, and present purchases from manual, CSV, email, desktop-app, eBay, Amazon, and other channel sources in one scannable table-oriented workspace.

#### Scenario: Review and start purchase creation/import
- **GIVEN** the user opens `/purchases`
- **WHEN** the Purchases workspace renders
- **THEN** Cabinet MUST show Purchases as the route title, expose a table shell with purchase/source/price/status/tracking/action columns, and provide a `+` add action that opens a creation/import dialog with New, CSV, and Email modes.
- **AND** primary navigation and command navigation MUST expose the route with the user-facing `Purchases` label.
- **AND** captured-review and purchase-source-match tooling MUST NOT render as standalone primary page actions or default body sections.
- **AND** users MUST be able to deliberately open add/import from the standard page header action area, without a duplicate in-body Purchases title or description block above the table.

### Requirement COMMERCE-RECONCILIATION-007: Purchases SHALL support table filtering and row state actions
Cabinet SHALL let users narrow the Purchases table by purchase text/source/status signals and expose row-level controls for common review state markers before deeper edit/persistence workflows are completed.

#### Scenario: Filter and mark purchase rows
- **GIVEN** the Purchases table contains captured or imported purchase rows
- **WHEN** the user searches purchases, applies a review status filter, or marks a visible row as favorite, arrived, or rated
- **THEN** Cabinet MUST keep the table scannable, show the filtered row count, preserve independent source/status evidence in visible rows, and reflect the selected favorite/arrival/rating state on the affected row without mutating unrelated rows.

### Requirement COMMERCE-RECONCILIATION-008: Purchases SHALL support manual purchase draft creation
Cabinet SHALL let users create a manual purchase draft from the Purchases `+` dialog before the durable purchase API workflow is completed.

#### Scenario: Create a manual purchase draft
- **GIVEN** the user opens the Purchases `+` dialog and selects the New mode
- **WHEN** the user enters a purchase title with optional source, price, and tracking evidence and saves the draft
- **THEN** Cabinet MUST add a manual-draft row to the Purchases table, preserve the entered source/price/tracking evidence in the visible row, expose the row through table search/filtering, and clearly state that durable API persistence remains pending a follow-up slice.

### Requirement COMMERCE-RECONCILIATION-009: Purchases SHALL preview CSV and email imports before confirmation
Cabinet SHALL let users paste CSV rows or order/email text from the Purchases `+` dialog, preview parsed purchase draft fields, and require explicit confirmation before imported purchase drafts appear in the table.

#### Scenario: Preview and confirm purchase imports
- **GIVEN** the user opens the Purchases `+` dialog and selects CSV or Email mode
- **WHEN** the user enters import text and requests a preview
- **THEN** Cabinet MUST show parsed draft fields including source/provenance, title, price/currency, purchase date when available, seller/source/channel, and tracking/delivery evidence when available.
- **AND** Cabinet MUST show actionable empty or parse-failure feedback when no purchase draft can be parsed.
- **AND** Cabinet MUST NOT add imported purchase drafts to the Purchases table until the user explicitly confirms the preview.
- **WHEN** the user confirms a CSV or Email preview
- **THEN** Cabinet MUST add the confirmed import draft rows to the Purchases table with CSV or Email import status and preserve the previewed provenance, price, tracking, and delivery evidence.

### Requirement COMMERCE-RECONCILIATION-010: Purchases SHALL persist add/import drafts through commerce lifecycle records
Cabinet SHALL persist manual, CSV, and email purchase drafts from the Purchases `+` dialog through the commerce lifecycle API so confirmed purchase rows have durable lifecycle and expected-arrival evidence instead of only local UI state.

#### Scenario: Persist confirmed purchase drafts
- **GIVEN** the user opens the Purchases `+` dialog and enters a valid manual purchase or confirms a CSV/email preview
- **WHEN** Cabinet accepts the purchase draft
- **THEN** Cabinet MUST create the purchase item record needed by the commerce lifecycle API and then create a `purchase` lifecycle entry for that item.
- **AND** the commerce lifecycle response MUST include an expected-arrival record linked to the lifecycle entry.
- **AND** the Purchases table MUST show persistence evidence for the added row, including lifecycle and expected-arrival identifiers.
- **AND** parse failures or API persistence failures MUST remain visible and MUST NOT add unpersisted manual, CSV, or email rows to the table.

### Requirement COMMERCE-RECONCILIATION-011: Purchases SHALL expose purchase metadata and order links in table rows
Cabinet SHALL keep purchase date, delivery, and original order-link evidence visible in the Purchases table so cross-platform rows remain scannable without opening a separate detail surface.

#### Scenario: Review row metadata and source order links
- **GIVEN** the Purchases table contains captured, manual, CSV, or email purchase rows
- **WHEN** the row renders
- **THEN** Cabinet MUST show purchase date and delivery evidence when available and a pending state when they are not available.
- **AND** rows with a source order URL MUST expose an external open-order action without replacing the purchase row.
- **AND** table search MUST include purchase date, delivery, and order-link evidence so users can locate rows by those metadata fields.

### Requirement COMMERCE-RECONCILIATION-012: Purchases page actions SHALL be modal-backed and product-approved
Purchases SHALL render primary page header actions as clear modal-backed commands and SHALL NOT expose rejected source-match or captured-review workflows as primary page actions.

#### Scenario: Header actions are icon-only but accessible
- **GIVEN** the user opens `/purchases`
- **WHEN** the Purchases page header action area renders
- **THEN** the Add purchase control MUST render without visible button text.
- **AND** the Add purchase control MUST preserve an accessible label and tooltip describing the command.

#### Scenario: Add purchase opens the purchase creation workflow
- **GIVEN** the Purchases page has loaded
- **WHEN** the user activates the Add purchase action
- **THEN** Cabinet MUST show a modal purchase creation workflow.
- **AND** the modal MUST expose manual/new, CSV import, and email import modes as one purchase creation workflow.

#### Scenario: Rejected review actions are absent from the primary Purchases page
- **GIVEN** the Purchases page has loaded
- **WHEN** the Purchases page header and default page body render
- **THEN** Cabinet MUST NOT render `Review source matches` or `Review captured purchases` as header actions.
- **AND** Cabinet MUST NOT render standalone source-match or captured-review stacked sections as primary page body workflows.
