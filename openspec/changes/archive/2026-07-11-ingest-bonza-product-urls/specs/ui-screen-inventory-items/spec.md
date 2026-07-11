## ADDED Requirements

### Requirement: Inventory create paste flow SHALL process supported provider URLs
Inventory create paste flow SHALL call provider URL ingestion for supported product URLs and prefill the create-item modal from the normalized item draft.

#### Scenario: Pasted Bonza URL prefills create item modal
- **GIVEN** user opens Inventory and activates the paste create action
- **WHEN** user submits `https://bonzaslotcars.com.au/product/bonza-mug-white/`
- **THEN** the create item modal MUST remain in confirm-before-create mode
- **AND** modal fields MUST be prefilled from Bonza normalized data
- **AND** title MUST be populated with `BONZA MUG WHITE`
- **AND** source URL MUST include the pasted Bonza product URL
- **AND** price, category, item metadata, stock, description, and image URL evidence MUST be available for review before create

#### Scenario: Unsupported pasted URL gives actionable feedback
- **GIVEN** user opens the Inventory create paste flow
- **WHEN** user submits a URL that does not match a supported provider product URL
- **THEN** UI MUST show an actionable unsupported-provider or unsupported-page message
- **AND** UI MUST keep the pasted value available for manual item creation
- **AND** UI MUST NOT discard user input

#### Scenario: Duplicate Bonza URL blocks silent create
- **GIVEN** provider ingestion reports an existing item for the pasted Bonza product URL
- **WHEN** the create modal renders the result
- **THEN** UI MUST show duplicate information
- **AND** UI MUST provide an action to open the existing item
- **AND** UI MUST require explicit confirmation before creating another item from the same source

### Requirement: Inventory create flow SHALL preserve provider provenance for pasted URLs
Inventory item creation from provider-ingested pasted URLs SHALL save source provenance and user-visible evidence with the created item.

#### Scenario: Created item stores Bonza source evidence
- **GIVEN** user confirms creation from a Bonza-ingested product draft
- **WHEN** Cabinet creates the inventory item
- **THEN** the item MUST retain original pasted URL, normalized source URL, provider id, provider family, provider product id, observed timestamp, and extraction method
- **AND** the item detail/editor evidence area MUST be able to display the source link and provider extraction summary
