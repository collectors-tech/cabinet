## Purpose
Define safe import/export flows and maintenance operations for collection data stores.

## Requirements
### Requirement DATA-MANAGEMENT-001: Data import SHALL support safe dry-run and explicit conflict choices
Cabinet SHALL support JSON/CSV dry-run preview and merge/create/skip conflict resolution on apply.

#### Scenario: Import dry-run
- **GIVEN** import payload is submitted
- **WHEN** user runs dry-run import
- **THEN** Cabinet SHALL report conflicts without mutating persisted records

### Requirement DATA-MANAGEMENT-002: Maintenance operations SHALL include reindex and repair
Cabinet SHALL support search reindex and database repair endpoints.

#### Scenario: Reindex operation
- **GIVEN** maintenance operation is requested
- **WHEN** user triggers reindex
- **THEN** Cabinet SHALL execute and report maintenance outcome

### Requirement DATA-MANAGEMENT-003: Import apply SHALL report explicit mutation counts
Cabinet SHALL require import apply flows to return readable created, merged, skipped, and failed counts so users can verify the result of confirmed JSON/CSV imports.

#### Scenario: Apply confirmed import
- **GIVEN** a user has reviewed an import dry-run and submits an apply request with conflict choices
- **WHEN** Cabinet applies the import
- **THEN** the response SHALL include total item, created, merged, skipped, and failed counts
- **AND** the failed count SHALL be zero only when all requested item actions were committed successfully

### Requirement DATA-MANAGEMENT-004: Backup and restore SHALL report verifiable outcomes and require restore confirmation
Cabinet SHALL require explicit restore confirmation and SHALL return readable backup, list, and restore metadata including selected path, file name, size, timestamp, archive format, download URL, and integrity-check outcome.

#### Scenario: Backup run and restore
- **GIVEN** a user creates or selects a database backup
- **WHEN** Cabinet runs backup, lists backups, or restores a confirmed backup
- **THEN** responses SHALL include user-verifiable backup or restore metadata
- **AND** restore SHALL fail without explicit confirmation before replacing the active database
- **AND** newly created backups SHALL be timestamped ZIP archives containing the active database and app-owned backup metadata
- **AND** the Settings backup flow SHALL expose the generated ZIP filename and a download action

### Requirement DATA-MANAGEMENT-005: Data export SHALL expose explicit download affordances
Cabinet SHALL expose profile-scoped JSON snapshot and CSV item exports as user-downloadable actions with deterministic download filenames.

#### Scenario: Export snapshot and item CSV
- **GIVEN** a user is reviewing Settings Storage for the active profile
- **WHEN** Cabinet offers JSON snapshot and CSV item exports
- **THEN** the UI SHALL provide clear download actions for both formats
- **AND** the export endpoints SHALL return attachment filenames that identify the Cabinet snapshot or item CSV export

### Requirement DATA-MANAGEMENT-006: Import apply UI SHALL report mutation counts
Cabinet SHALL show created, merged, skipped, and failed counts after a confirmed JSON or CSV import apply so users can verify the result before continuing recovery or portability work.

#### Scenario: Import apply count feedback
- **GIVEN** a user has reviewed a JSON or CSV import dry-run and confirms apply
- **WHEN** the apply request succeeds
- **THEN** the UI SHALL report total item, created, merged, skipped, and failed counts returned by the API
- **AND** the user SHALL remain on the import operations screen for follow-up review

### Requirement DATA-MANAGEMENT-007: Maintenance endpoints SHALL report operation metadata
Cabinet SHALL return explicit operation metadata for maintenance actions so users and diagnostics can verify what ran, whether it completed, and when it completed.

#### Scenario: Reindex and repair maintenance
- **GIVEN** a user triggers search reindex or database integrity check maintenance
- **WHEN** Cabinet completes the requested operation
- **THEN** the response SHALL include `ok`, `operation`, and `completed_at` metadata
- **AND** reindex SHALL report that the search index was rebuilt
- **AND** repair SHALL report the integrity-check result

### Requirement DATA-MANAGEMENT-008: Import failure feedback SHALL preserve recovery context
Cabinet SHALL explain import dry-run and apply failures without implying partial mutation, and SHALL preserve reviewed dry-run summaries after failed apply attempts so users can retry safely.

#### Scenario: Import failure recovery
- **GIVEN** a user is running a JSON or CSV import dry-run or applying a reviewed import
- **WHEN** the dry-run or apply request fails
- **THEN** the UI SHALL state that no records were changed
- **AND** failed apply feedback SHALL keep the reviewed dry-run summary visible for retry or conflict-action changes
