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

### Requirement UI-SCREEN-SETTINGS-STORAGE-004: Storage screen SHALL degrade gracefully when storage info endpoint is unavailable
Storage section MUST avoid hard-fail UX and provide actionable recovery with retained baseline path context.

#### Scenario: Storage info unavailable
- **GIVEN** storage info endpoint fails or active profile context is temporarily unavailable
- **WHEN** user opens `/settings/storage`
- **THEN** UI MUST show non-blocking error state with retry action
- **AND** last-known database/media locations (if available) MUST remain visible
- **AND** diagnostics-only actions MUST be disabled with explicit reason, not generic failure

### Requirement UI-SCREEN-SETTINGS-STORAGE-005: Retry action SHALL re-attempt storage info fetch without full page reload

#### Scenario: Retry storage fetch
- **GIVEN** storage info error state is visible
- **WHEN** user clicks `Retry`
- **THEN** runtime MUST re-attempt fetch and render ready state immediately on success without route reload

### Requirement UI-SCREEN-SETTINGS-STORAGE-003: Storage screen SHALL support optional migration

#### Scenario: Migrate existing media on path change
- **GIVEN** user changes storage root path and enables migration
- **WHEN** migration runs
- **THEN** runtime MUST preserve linkage integrity and return migration summary

### Requirement UI-SCREEN-SETTINGS-STORAGE-006: Storage screen SHALL expose Reindex Search and Rebuild Thumbnails maintenance actions

#### Scenario: Reindex Search action
- **GIVEN** storage section is ready and maintenance actions are available
- **WHEN** user clicks `Reindex Search`
- **THEN** runtime MUST start reindex workflow and return deterministic completion/error feedback

#### Scenario: Rebuild Thumbnails action
- **GIVEN** storage section is ready and thumbnail rebuild action is available
- **WHEN** user clicks `Rebuild Thumbnails`
- **THEN** runtime MUST start thumbnail rebuild workflow and return deterministic completion/error feedback

### Requirement UI-SCREEN-SETTINGS-STORAGE-007: Storage screen SHALL show backups in a sortable inspection table
Backup archives MUST render as a dense table so users can inspect backup identity, freshness, source type, size, validity, and supported actions before restore.

#### Scenario: Backup table inspection and sorting
- **GIVEN** backup metadata is available from `/api/backup/list`
- **WHEN** user opens `/settings/storage`
- **THEN** the backup area MUST render a table with filename, created timestamp, backup source, archive size, status, and actions columns
- **AND** generated timestamped ZIP archives MUST be visually distinguished from legacy database snapshots or unsupported artifacts
- **AND** users MUST be able to sort by filename and created timestamp
- **AND** long filenames and paths MUST remain readable without overlapping restore or download actions

## Use-Case IDs and E2E Mapping
| UC ID | Flow | Expected Result | E2E Mapping |
| --- | --- | --- | --- |
| UC-SET-STR-01 | Reindex Search action | `Reindex Search` triggers search reindex workflow with deterministic completion/error feedback | `ui.web/cypress/e2e/settings/storage/spec.cy.ts` `UI-SCREEN-SETTINGS-STORAGE-006 runs Reindex Search and reports deterministic completion feedback`, `UI-SCREEN-SETTINGS-STORAGE-006 reports Reindex Search failure without route reload` |
| UC-SET-STR-02 | Rebuild Thumbnails action | `Rebuild Thumbnails` triggers thumbnail maintenance workflow with deterministic completion/error feedback | `ui.web/cypress/e2e/settings/storage/spec.cy.ts` `UI-SCREEN-SETTINGS-STORAGE-006 runs Rebuild Thumbnails and reports deterministic feedback`, `UI-SCREEN-SETTINGS-STORAGE-006 reports Rebuild Thumbnails failure without route reload` |
| UC-SET-STR-03 | Backup table inspection | Backup archives render in a sortable metadata table with supported download/restore actions | `ui.web/cypress/e2e/settings/storage/spec.cy.ts` `UI-SCREEN-SETTINGS-STORAGE-007 creates, lists, and restores backups from Settings Storage` |
