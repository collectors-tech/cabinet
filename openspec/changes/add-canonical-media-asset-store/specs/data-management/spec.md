## MODIFIED Requirements

### Requirement: DATA-MANAGEMENT-004: Backup and restore SHALL report verifiable outcomes and require restore confirmation
Cabinet SHALL require explicit restore confirmation and SHALL return readable backup, list, and restore metadata including selected path, file name, size, timestamp, archive format, download URL, and integrity-check outcome.

#### Scenario: Backup run and restore
- **GIVEN** a user creates or selects a database backup
- **WHEN** Cabinet runs backup, lists backups, or restores a confirmed backup
- **THEN** responses SHALL include user-verifiable backup or restore metadata
- **AND** restore SHALL fail without explicit confirmation before replacing the active database
- **AND** newly created backups SHALL be timestamped ZIP archives containing the active database and app-owned backup metadata
- **AND** confirmed restores SHALL take and report a pre-restore ZIP backup of the active database before replacing it
- **AND** restore replacement SHALL reject active-database alias paths and SHALL preserve the active database when replacement cannot proceed
- **AND** corrupt or incomplete restore inputs SHALL fail before replacement and SHALL NOT create misleading pre-restore or success metadata
- **AND** the Settings backup flow SHALL expose the generated ZIP filename and a download action
- **AND** the Settings backup list SHALL render backup metadata in a sortable table that distinguishes generated ZIP archives from legacy database snapshots
- **AND** backup archives SHALL include canonical media asset folders and manifests needed to restore `media/assets/<asset-id>/...`
- **AND** confirmed restores SHALL copy canonical media entries back under the configured data directory without allowing archive path traversal outside that directory
