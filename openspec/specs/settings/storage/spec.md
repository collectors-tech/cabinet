## Purpose
Define Storage settings screen behavior for media path management.

## Requirements
### Requirement UI-SCREEN-SETTINGS-STORAGE-001: Storage screen SHALL manage media root path

#### Scenario: Save media root path
- **GIVEN** user opens `/settings/storage`
- **WHEN** user saves a writable path
- **THEN** runtime MUST persist `storage.media_root_path` for active profile

### Requirement UI-SCREEN-SETTINGS-STORAGE-002: Storage screen SHALL validate path deterministically

#### Scenario: Invalid storage path
- **GIVEN** configured path is invalid or not writable
- **WHEN** user saves or runs write-check
- **THEN** UI MUST show deterministic validation error and runtime MUST reject update

### Requirement UI-SCREEN-SETTINGS-STORAGE-003: Storage screen SHALL support optional migration

#### Scenario: Migrate existing media on path change
- **GIVEN** user changes storage root path and enables migration
- **WHEN** migration runs
- **THEN** runtime MUST preserve linkage integrity and return migration summary
