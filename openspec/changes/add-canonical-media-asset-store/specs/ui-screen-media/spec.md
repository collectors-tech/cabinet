## MODIFIED Requirements

### Requirement: MEDIA-STORAGE-001: Upload pipeline SHALL store assets in a global media registry
Cabinet SHALL ingest uploaded binaries into a shared media library before linkage to inventory or wishlist targets.

#### Scenario: Upload inventory image into global library
- **GIVEN** user uploads image from inventory workflow
- **WHEN** upload is accepted
- **THEN** runtime MUST persist media asset in global registry and return canonical `asset_id` used for downstream linking
- **AND** runtime MUST write new assets through a staging path, validate/hash the original, write a versioned manifest, and atomically promote a completed `assets/<asset-id>/` folder
- **AND** database records for new canonical assets MUST store media-root-relative `assets/<asset-id>/...` paths and resolve them against the active profile media root at runtime

#### Scenario: Upload evidence from Chat, Media workspace, or channel capture
- **GIVEN** Chat attachments, Media workspace uploads, or supported channel captures provide media bytes for the active profile
- **WHEN** ingestion succeeds
- **THEN** runtime MUST store the bytes through the same canonical asset service and owner/provenance model used by inventory photos
- **AND** runtime MUST create a manifest owner reference for the source thread, inventory item, or other supported owner

### Requirement: MEDIA-STORAGE-003: Inventory and wishlist assignment SHALL use link records instead of binary duplication
Cabinet SHALL associate media assets to inventory/wishlist via linkage records.

#### Scenario: Assign shared asset to multiple targets
- **GIVEN** asset already exists in media registry
- **WHEN** user assigns asset to inventory item and wishlist entry
- **THEN** runtime MUST create/update link records per target and MUST NOT create duplicate media binaries
- **AND** deleting one target reference MUST NOT remove the canonical asset folder while another `item_photos`, `chat_attachments`, or `media_asset_links` reference remains

#### Scenario: Remove orphan canonical asset
- **GIVEN** a canonical asset folder exists under the active media root
- **WHEN** the final database reference to that asset is removed
- **THEN** runtime MAY remove the complete `assets/<asset-id>/` folder as one cleanup unit
- **AND** cleanup MUST stay scoped inside the active media root

### Requirement: MEDIA-STORAGE-004: Derived variants SHALL be generated once per canonical asset
Cabinet SHALL generate preview/thumbnail variants at asset level and reuse them across linked targets.

#### Scenario: Variant reuse across links
- **GIVEN** asset is linked to multiple inventory/wishlist targets
- **WHEN** target views request preview/thumbnail
- **THEN** runtime MUST serve shared derived variants for same `asset_id` without per-link regeneration
- **AND** user-generated or AI-generated alternatives MUST be tracked under `variations/` with parent/provenance links instead of overwriting `original/`
