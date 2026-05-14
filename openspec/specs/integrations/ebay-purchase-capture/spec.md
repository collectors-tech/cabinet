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
