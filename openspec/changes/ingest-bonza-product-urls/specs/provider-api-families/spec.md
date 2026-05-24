## ADDED Requirements

### Requirement: Provider URL router SHALL detect known product URLs deterministically
Cabinet SHALL parse pasted URLs, normalize host/path values, and route known provider product URLs to the matching provider family without AI inference.

#### Scenario: Bonza product URL routes to WooCommerce provider
- **GIVEN** user input contains `https://bonzaslotcars.com.au/product/bonza-mug-white/`
- **WHEN** Cabinet detects the pasted URL
- **THEN** runtime MUST classify the provider as `bonzaslotcars`
- **AND** runtime MUST classify the provider family as `woocommerce`
- **AND** runtime MUST extract product slug `bonza-mug-white`
- **AND** runtime MUST select product URL ingestion as the next action

#### Scenario: Known provider non-product URL is rejected clearly
- **GIVEN** user input contains a Bonza URL that is not under `/product/`
- **WHEN** Cabinet detects the pasted URL
- **THEN** runtime MUST return a supported-provider unsupported-page response
- **AND** response MUST NOT create or mutate an inventory item

### Requirement: WooCommerce product URL ingestion SHALL use Store API first
WooCommerce-backed product URL ingestion SHALL resolve product detail from the public Store API before attempting page metadata or HTML fallback extraction.

#### Scenario: Product detail resolved through Store API
- **GIVEN** a WooCommerce product URL with slug `bonza-mug-white`
- **WHEN** ingestion runs
- **THEN** runtime MUST query the Store API product surface using a slug-derived search term or equivalent provider-supported lookup
- **AND** runtime MUST match the returned product by exact slug or normalized permalink
- **AND** runtime MUST return a normalized product draft only when a deterministic match is found

#### Scenario: Store API fields are normalized consistently
- **GIVEN** Store API returns product title, price, currency, description, categories, attributes, images, and stock values
- **WHEN** ingestion normalizes the product
- **THEN** output MUST include common fields for provider id, provider family, provider product id, source URL, title, description, price, currency, category values, attribute values, stock state, stock count, image URLs, and evidence metadata

#### Scenario: Page fallback is limited to missing fields
- **GIVEN** Store API lookup succeeds but selected optional fields are missing
- **WHEN** fallback extraction runs
- **THEN** runtime MAY use product page metadata or HTML to fill missing title, image, price, category, description, or stock fields
- **AND** fallback evidence MUST identify the field source
