## Purpose
Define media-library ingestion, analysis, and assignment workflows across inventory and wishlist.

## Requirements
### Requirement MEDIA-AI-ASSIGNMENT-001: Cabinet SHALL support multi-asset upload per inventory item
Cabinet SHALL support uploading multiple image assets per inventory item with stable ordering and primary designation.

#### Scenario: Multi-file upload for item
- **GIVEN** item exists and user selects multiple image files
- **WHEN** user uploads files to item media workflow
- **THEN** runtime MUST persist all assets, assign display order, and expose primary-asset controls

### Requirement MEDIA-AI-ASSIGNMENT-002: Cabinet SHALL generate preview and thumbnail variants for uploaded assets
Cabinet SHALL generate and store derived preview/thumbnail variants for each uploaded image.

#### Scenario: Variant generation on upload
- **GIVEN** upload request stores original media file
- **WHEN** media pipeline completes processing
- **THEN** runtime MUST provide resolvable `original`, `preview`, and `thumbnail` variants for asset retrieval

### Requirement MEDIA-AI-ASSIGNMENT-003: Cabinet SHALL support AI-assisted media analysis before assignment
Cabinet SHALL support analysis workflow that produces suggested metadata and assignment candidates from uploaded assets.

#### Scenario: Analyze uploaded asset
- **GIVEN** unassigned media asset exists
- **WHEN** user triggers analysis action
- **THEN** runtime MUST return suggested labels/metadata with confidence and confirmation requirement flags

### Requirement MEDIA-AI-ASSIGNMENT-004: Cabinet SHALL support assignment of analyzed assets to inventory and wishlist targets
Cabinet SHALL support assigning analyzed assets to inventory items and wishlist entries while preserving auditability.

#### Scenario: Assign analyzed asset
- **GIVEN** analysis result and user-selected target (`inventory_item_id` or `wishlist_entry_id`)
- **WHEN** assignment is confirmed
- **THEN** runtime MUST persist assignment links and expose updated media state on both inventory and wishlist views

### Requirement MEDIA-AI-ASSIGNMENT-005: Cabinet SHALL enforce deterministic media state transitions
Cabinet SHALL maintain explicit state transitions for `uploaded`, `analyzed`, `assigned`, and `archived` asset lifecycle states.

#### Scenario: State transition validation
- **GIVEN** asset is in `uploaded` state
- **WHEN** assignment is requested before analysis
- **THEN** runtime MUST reject transition with deterministic error and guidance to run analysis first
