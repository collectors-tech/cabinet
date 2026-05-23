## ADDED Requirements

### Requirement: Bonza product URL ingestion SHALL populate Cabinet item draft data
Bonza provider ingestion SHALL support direct product URL ingestion for `bonzaslotcars.com.au/product/<slug>/` and return a normalized Cabinet item draft.

#### Scenario: Bonza mug URL extracts structured product data
- **GIVEN** user submits `https://bonzaslotcars.com.au/product/bonza-mug-white/`
- **WHEN** Bonza product ingestion runs
- **THEN** normalized output MUST include title `BONZA MUG WHITE`
- **AND** output MUST include provider product id `19603` when returned by the Store API
- **AND** output MUST include source URL `https://bonzaslotcars.com.au/product/bonza-mug-white/`
- **AND** output MUST include AUD price `9.95`
- **AND** output MUST include stock count `3` and stock state `in_stock` when the provider reports `3 in stock`
- **AND** output MUST include the product description text
- **AND** output MUST include categories and attributes returned by the provider
- **AND** output MUST include at least one product image URL when provider images are available

#### Scenario: Bonza categories and attributes map to item metadata
- **GIVEN** Bonza Store API returns categories and attributes for a product
- **WHEN** product data is normalized for Cabinet
- **THEN** categories MUST map to Cabinet category draft values
- **AND** Brand, Scale, and Type attributes MUST map to item metadata or evidence fields
- **AND** Type MAY map to Cabinet Item Type when a matching configured item type exists

#### Scenario: Bonza evidence records extraction source
- **GIVEN** Bonza product ingestion succeeds
- **WHEN** normalized output is returned
- **THEN** evidence MUST include provider `bonzaslotcars`, family `woocommerce`, extraction method `store_api`, product id, original pasted URL, normalized source URL, and observed timestamp

### Requirement: Bonza product URL ingestion SHALL protect against duplicates
Bonza product URL ingestion SHALL check existing item source evidence before allowing a duplicate provider product to be created silently.

#### Scenario: Existing Bonza product source is detected
- **GIVEN** an inventory item already has source evidence for Bonza product id `19603` or the normalized Bonza mug source URL
- **WHEN** the same Bonza product URL is processed again
- **THEN** runtime MUST return duplicate candidate information
- **AND** Inventory UI MUST offer to open the existing item or continue only with explicit user confirmation
