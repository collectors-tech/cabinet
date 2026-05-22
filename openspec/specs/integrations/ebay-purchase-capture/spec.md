## Purpose
Define eBay purchase-history capture behavior for Cabinet browser companion modules.

## Requirements
### Requirement EBAY-PURCHASE-CAPTURE-001: eBay purchase capture MUST preserve listing title separately from purchased item metadata
Cabinet SHALL treat eBay purchase-history cards as structured purchase records rather than a flat title export.

#### Scenario: Capture card variant metadata from purchase card
- **GIVEN** an eBay AU purchase-history card has a broad listing title and item-card aspect metadata
- **WHEN** Cabinet parses the purchase card capture payload
- **THEN** the broad listing title SHALL be preserved as listing title
- **AND** aspect metadata such as `Choose Single Or Playset`, `Card`, and `Quantity` SHALL be preserved as structured key/value metadata
- **AND** the `Card` aspect SHALL be available as the purchased item identity candidate when present
- **AND** quantity SHALL be parsed from card metadata when present

### Requirement EBAY-PURCHASE-CAPTURE-002: eBay purchase capture MUST preserve seller, order, transaction, and action metadata separately
Cabinet SHALL preserve capture metadata needed for reconciliation and future explicit workflows without executing eBay actions during passive capture.

#### Scenario: Capture passive purchase-card actions
- **GIVEN** an eBay AU purchase-history card contains seller links, note actions, feedback links, and More Actions menu entries
- **WHEN** Cabinet parses the purchase card capture payload
- **THEN** seller username and seller profile URL SHALL be captured when visible
- **AND** order ID, transaction ID, listing ID, and variation ID SHALL be parsed from card URLs/attributes where visible
- **AND** note, feedback, contact seller, return, buy-again/view-item, hide-order, and help/report actions SHALL be captured as passive action metadata only
- **AND** passive capture MUST NOT execute note, feedback, contact, return, buy-again, hide-order, help/report, or other menu actions

### Requirement EBAY-PURCHASE-CAPTURE-003: eBay purchase capture MUST group purchased items under purchase orders
Cabinet SHALL represent captured eBay purchases as order parent records with child purchased item records, rather than flattening all item cards into one list.

#### Scenario: Group one or more item cards under the captured order
- **GIVEN** multiple captured eBay purchase-history item cards share the same order ID
- **WHEN** Cabinet groups the parsed purchase cards for the Purchase Inbox
- **THEN** one purchase order parent record SHALL be produced for that order ID
- **AND** each purchased item card SHALL remain a child record under the order parent
- **AND** order-level metadata such as order total, currency, seller set, status, destination marker, costs, and order detail URL SHALL be preserved on the parent when available
- **AND** item-level metadata such as listing ID, variation ID, transaction ID, purchased card/aspect metadata, quantity, item price, image, item status, tracking status, note capability, and passive action metadata SHALL remain on the child item

#### Scenario: Merge repeated captures without duplicating child items
- **GIVEN** the same eBay purchase card is captured more than once
- **WHEN** Cabinet groups the repeated captures by order
- **THEN** the repeated captures SHALL merge into the same order parent
- **AND** matching child items SHALL be merged by transaction ID, listing/variation ID, or purchased identity fallback
- **AND** updated item metadata from later captures SHALL be preserved without creating duplicate child rows

### Requirement EBAY-PURCHASE-CAPTURE-004: Purchase Inbox review records MUST expose safe item actions before mutation
Cabinet SHALL convert grouped purchase captures into review records that expose order/item status, missing fields, and suggested next actions without linking or creating inventory records automatically.

#### Scenario: Build review actions for a ready purchase item
- **GIVEN** a captured purchase item has stable identity, quantity, and price evidence
- **WHEN** Cabinet prepares Purchase Inbox review records
- **THEN** the item SHALL expose link-existing-item and convert-to-inventory suggested actions
- **AND** each mutating suggested action SHALL require explicit confirmation before inventory state changes
- **AND** the action target SHALL use a stable purchase item key derived from captured order/item metadata

#### Scenario: Flag incomplete purchase item fields
- **GIVEN** a captured purchase item is missing quantity or price evidence
- **WHEN** Cabinet prepares Purchase Inbox review records
- **THEN** the order and item SHALL be marked as needs-review
- **AND** missing fields SHALL be listed for the item
- **AND** Cabinet SHALL suggest completing missing fields rather than offering a mutating link or convert action

### Requirement EBAY-PURCHASE-CAPTURE-005: Purchase Inbox API MUST prepare review records without inventory mutation
Cabinet SHALL expose a profile-scoped Purchase Inbox API boundary that prepares captured eBay purchase cards for Inbox review while preserving confirmation-before-mutation behavior.

#### Scenario: Prepare Purchase Inbox reviews from captured cards
- **GIVEN** Cabinet has an active profile and receives captured eBay purchase cards for Purchase Inbox review
- **WHEN** the client posts the cards to the Purchase Inbox review API
- **THEN** Cabinet SHALL return review records grouped by purchase order for the active profile
- **AND** ready items SHALL expose confirmation-required link-existing-item and convert-to-inventory actions
- **AND** incomplete items SHALL expose missing field details and non-mutating completion actions
- **AND** the API response SHALL identify the source as ebay_purchase_capture
- **AND** the API MUST NOT create, link, or update inventory records while preparing review records

### Requirement EBAY-PURCHASE-CAPTURE-006: Purchase Inbox UI MUST review captured purchases before confirmed mutation actions
Cabinet SHALL expose a Purchase Inbox UI surface for captured eBay purchase reviews that covers empty, loading, error, and ready states while keeping link and convert actions confirmation-gated.

#### Scenario: Review captured purchase order and item states
- **GIVEN** Cabinet has captured eBay purchase cards ready for review
- **WHEN** the user opens the Purchase Inbox and prepares review records
- **THEN** the UI SHALL show order-level review cards with seller, total, currency, status, and child purchased items
- **AND** ready child items SHALL expose link-existing-inventory-item and convert-to-inventory-item actions
- **AND** incomplete child items SHALL list missing fields and expose a non-mutating completion action

#### Scenario: Confirm mutation-gated purchase item actions
- **GIVEN** a ready Purchase Inbox item exposes a mutating link or convert action
- **WHEN** the user selects that action
- **THEN** Cabinet SHALL show an explicit confirmation dialog before queueing the action
- **AND** the UI SHALL not present the action as applied until the user confirms

#### Scenario: Recover from review API failure
- **GIVEN** the Purchase Inbox review API is unavailable or returns an error
- **WHEN** the user prepares review records
- **THEN** the UI SHALL show an error state that keeps the page usable and retryable
