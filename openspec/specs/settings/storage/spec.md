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

## Use-Case IDs and E2E Mapping
| UC ID | Flow | Expected Result | E2E Mapping |
| --- | --- | --- | --- |
| UC-SET-STR-01 | Reindex Search action | `Reindex Search` triggers search reindex workflow with deterministic feedback | `ui.web/cypress/e2e/settings/storage/spec.cy.ts` `UI-SCREEN-SETTINGS-STORAGE-006 runs Reindex Search and reports deterministic completion feedback` |
| UC-SET-STR-02 | Rebuild Thumbnails action | `Rebuild Thumbnails` triggers thumbnail maintenance workflow with deterministic feedback | `ui.web/cypress/e2e/settings/storage/spec.cy.ts` `UI-SCREEN-SETTINGS-STORAGE-006 runs Rebuild Thumbnails and reports deterministic feedback` |
