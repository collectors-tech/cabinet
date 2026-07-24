## MODIFIED Requirements

### Requirement UI-SCREEN-SETTINGS-STORAGE-003: Storage screen SHALL support optional migration
Storage migration controls SHALL make canonical media migration state inspectable and recoverable without requiring log-file access.

#### Scenario: Report canonical media migration readiness
- **GIVEN** user opens `/settings/storage` for an active profile that may contain legacy inventory photos or Chat attachments
- **WHEN** Storage migration status loads
- **THEN** UI MUST show the current migration state as one of `not_needed`, `ready`, `dry_run_complete`, `applying`, `needs_repair`, `blocked`, `rollback_available`, or `completed`
- **AND** UI MUST show discovered, pending, migrated, already-migrated, duplicate, skipped, failed and orphan counts when available
- **AND** UI MUST show the latest checkpoint identifier and timestamp when a recoverable checkpoint exists
- **AND** UI MUST NOT expose raw absolute source paths by default.

#### Scenario: Surface repair and rollback actions safely
- **GIVEN** migration status is `needs_repair`, `blocked`, or `rollback_available`
- **WHEN** user reviews Settings > Storage
- **THEN** UI MUST show record identifiers, redacted path class, failure category and recovery guidance for each actionable problem
- **AND** repair and rollback actions MUST require explicit confirmation before mutating data
- **AND** repair MUST re-run verification before changing references
- **AND** rollback MUST only be enabled when a recoverable checkpoint is present and verified.

#### Scenario: Orphan audit is report-only
- **GIVEN** migration status includes orphaned legacy files
- **WHEN** user reviews the orphan audit on Settings > Storage
- **THEN** UI MUST show orphan count, inferred type, redacted location summary and recovery guidance
- **AND** UI MUST NOT offer silent delete as part of migration, repair, rollback or audit actions.
