## MODIFIED Requirements

### Requirement: MEDIA-STORAGE-001: Upload pipeline SHALL store assets in a global media registry
Cabinet SHALL ingest uploaded binaries into a shared media library before linkage to inventory or wishlist targets.

#### Scenario: Migrate legacy media into global library
- **GIVEN** the active profile contains legacy inventory photo files, legacy Chat attachment files, or canonical assets already under `media/assets/<asset-id>/`
- **WHEN** Cabinet runs the media migration preflight or apply operation
- **THEN** runtime MUST classify each discovered file as `pending`, `already_migrated`, `duplicate`, `missing`, `corrupt`, `locked`, `failed`, `orphan`, or `migrated`
- **AND** dry-run/preflight MUST report counts and affected record identifiers without mutating files or database rows
- **AND** apply MUST create canonical asset folders and manifests for eligible legacy files before rewriting references to media-root-relative `assets/<asset-id>/...` paths
- **AND** running apply repeatedly MUST NOT create duplicate canonical assets or change already-migrated references.

### Requirement: MEDIA-STORAGE-003: Inventory and wishlist assignment SHALL use link records instead of binary duplication
Cabinet SHALL associate media assets to inventory/wishlist via linkage records.

#### Scenario: Preserve legacy source until verified checkpoint
- **GIVEN** a legacy media file is eligible for canonical migration
- **WHEN** Cabinet applies the migration
- **THEN** runtime MUST hash the source original and migrated original before changing database references
- **AND** runtime MUST keep the legacy source file until the canonical asset, manifest, database commit, hash verification and recoverable checkpoint have succeeded
- **AND** interruption before commit MUST leave the legacy source as the active reference
- **AND** interruption after commit MUST allow the next run to resume from canonical references without exposing partial staging folders.

#### Scenario: Audit orphans without deletion
- **GIVEN** files exist under legacy media or Chat attachment locations with no matching active database reference
- **WHEN** Cabinet runs migration preflight, apply, repair, or audit
- **THEN** runtime MUST report the files as orphans with path, inferred type and recovery guidance
- **AND** runtime MUST NOT delete orphan files silently.
