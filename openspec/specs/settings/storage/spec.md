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

### Requirement UI-SCREEN-SETTINGS-STORAGE-008: Storage screen SHALL report restore failures without navigation loss
Restore failures MUST keep the user on `/settings/storage`, close the confirmation dialog, and show deterministic retry-safe feedback.

#### Scenario: Restore failure
- **GIVEN** backup metadata is available and the user confirms a restore
- **WHEN** `/api/backup/restore` rejects the selected backup
- **THEN** the Storage screen MUST remain on `/settings/storage`
- **AND** the restore confirmation dialog MUST close
- **AND** the UI MUST show `Backup restore failed.` feedback without implying data mutation

### Requirement UI-SCREEN-SETTINGS-STORAGE-009: Storage screen SHALL run database integrity checks
The Storage diagnostics area MUST expose a database integrity check action with readable healthy-result feedback.

#### Scenario: Healthy integrity check
- **GIVEN** storage information and backup list state are loaded
- **WHEN** user clicks `Run Integrity Check`
- **THEN** runtime MUST call the database repair/integrity endpoint
- **AND** the UI MUST report the returned integrity result

### Requirement UI-SCREEN-SETTINGS-STORAGE-010: Storage screen SHALL report integrity-check failures without navigation loss
Integrity-check failures MUST leave the Storage route available and show deterministic recovery feedback.

#### Scenario: Integrity check failure
- **GIVEN** storage information and backup list state are loaded
- **WHEN** the database repair/integrity endpoint fails
- **THEN** the Storage screen MUST remain on `/settings/storage`
- **AND** the UI MUST show failure feedback that directs the user to retry after diagnostics recovery

### Requirement UI-SCREEN-SETTINGS-STORAGE-011: Storage screen SHALL expose active-profile export downloads only in ready state
Data export actions MUST provide profile-scoped JSON snapshot and item CSV downloads after storage context loads, and MUST not expose live download links while storage context is degraded.

#### Scenario: Export downloads ready
- **GIVEN** storage information and backup list state are loaded
- **WHEN** user reviews Data exports on `/settings/storage`
- **THEN** JSON Snapshot MUST link to `/api/data/export/json` with download filename `cabinet-data-snapshot.json`
- **AND** Item CSV MUST link to `/api/data/export/csv/items` with download filename `cabinet-items.csv`

#### Scenario: Export downloads blocked while degraded
- **GIVEN** storage information fails to load for the active profile
- **WHEN** user reviews Data exports on `/settings/storage`
- **THEN** JSON Snapshot and Item CSV actions MUST be disabled buttons
- **AND** live export download links MUST NOT render until storage context recovers

### Requirement UI-SCREEN-SETTINGS-STORAGE-012: Storage screen SHALL use the route header as the visible page title
Storage route title hierarchy MUST remain owned by the shared Settings route header rather than duplicating a second page-level heading inside the content body.

#### Scenario: Route header title
- **GIVEN** user opens `/settings/storage`
- **WHEN** storage information and backup list state are loaded
- **THEN** the shared Settings header MUST show `Storage Settings`
- **AND** the Storage content body MUST NOT render a duplicate `h1`

### Requirement UI-SCREEN-SETTINGS-STORAGE-013: Storage actions SHALL remain scoped to local cards, rows, and dialogs
Storage actions MUST stay near the data or workflow they affect instead of creating a global header action region.

#### Scenario: Local storage actions
- **GIVEN** storage information and backup list state are loaded
- **WHEN** user reviews `/settings/storage`
- **THEN** no global storage header actions region MUST render
- **AND** the backup creation action MUST remain inside the Backups card
- **AND** backup download and restore actions MUST remain scoped to each backup row
- **AND** restore confirmation actions MUST render inside the restore confirmation dialog

## Use-Case IDs and E2E Mapping
| UC ID | Flow | Expected Result | E2E Mapping |
| --- | --- | --- | --- |
| UC-SET-STR-01 | Reindex Search action | `Reindex Search` triggers search reindex workflow with deterministic completion/error feedback | `ui.web/cypress/e2e/settings/storage/spec.cy.ts` `UI-SCREEN-SETTINGS-STORAGE-006 runs Reindex Search and reports deterministic completion feedback`, `UI-SCREEN-SETTINGS-STORAGE-006 reports Reindex Search failure without route reload` |
| UC-SET-STR-02 | Rebuild Thumbnails action | `Rebuild Thumbnails` triggers thumbnail maintenance workflow with deterministic completion/error feedback | `ui.web/cypress/e2e/settings/storage/spec.cy.ts` `UI-SCREEN-SETTINGS-STORAGE-006 runs Rebuild Thumbnails and reports deterministic feedback`, `UI-SCREEN-SETTINGS-STORAGE-006 reports Rebuild Thumbnails failure without route reload` |
| UC-SET-STR-03 | Backup table inspection | Backup archives render in a sortable metadata table with supported download/restore actions | `ui.web/cypress/e2e/settings/storage/spec.cy.ts` `UI-SCREEN-SETTINGS-STORAGE-007 creates, lists, and restores backups from Settings Storage` |
| UC-SET-STR-04 | Restore failure recovery | Failed restore closes confirmation, stays on Storage, and shows deterministic failure feedback | `ui.web/cypress/e2e/settings/storage/spec.cy.ts` `UI-SCREEN-SETTINGS-STORAGE-008 reports restore failure without route reload` |
| UC-SET-STR-05 | Integrity check | Database integrity check reports healthy and failure outcomes without route loss | `ui.web/cypress/e2e/settings/storage/spec.cy.ts` `UI-SCREEN-SETTINGS-STORAGE-009 runs storage integrity check and shows healthy result`, `UI-SCREEN-SETTINGS-STORAGE-010 reports integrity-check failure without route reload` |
| UC-SET-STR-06 | Export downloads | JSON snapshot and item CSV actions expose deterministic download targets only after ready storage context | `ui.web/cypress/e2e/settings/storage-export/spec.cy.ts` `UI-SCREEN-SETTINGS-STORAGE-011 exposes JSON snapshot and item CSV download actions`, `UI-SCREEN-SETTINGS-STORAGE-011 disables export downloads while storage context is degraded` |
| UC-SET-STR-07 | Storage title hierarchy | The shared Settings header owns the visible Storage page title without a duplicate body `h1` | `ui.web/cypress/e2e/settings/storage/spec.cy.ts` `UI-SCREEN-SETTINGS-STORAGE-012 uses the route header as the only visible page title` |
| UC-SET-STR-08 | Local storage actions | Backup and restore controls stay scoped to the Backups card, row, and restore dialog without global header actions | `ui.web/cypress/e2e/settings/storage/spec.cy.ts` `UI-SCREEN-SETTINGS-STORAGE-013 keeps storage actions scoped to cards, rows, and dialogs` |
