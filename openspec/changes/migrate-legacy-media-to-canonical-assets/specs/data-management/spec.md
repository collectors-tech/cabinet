## MODIFIED Requirements

### Requirement: DATA-MANAGEMENT-004: Backup and restore SHALL report verifiable outcomes and require restore confirmation
Cabinet SHALL require explicit restore confirmation and SHALL return readable backup, list, and restore metadata including selected path, file name, size, timestamp, archive format, download URL, and integrity-check outcome.

#### Scenario: Restore migrated media after profile relocation
- **GIVEN** a backup was created after legacy media migration completed for the active profile
- **WHEN** Cabinet restores that backup under a different Windows data directory
- **THEN** restored database rows MUST resolve migrated inventory photos and Chat attachments from the restored `media/assets/<asset-id>/...` tree
- **AND** restored media paths MUST remain media-root-relative rather than preserving source-machine absolute paths
- **AND** restore MUST reject archive traversal or misplaced media entries before replacing active data.

### Requirement: DATA-MANAGEMENT-009: Database upgrade SHALL preserve representative release data
Cabinet SHALL upgrade a representative database created by the prior release baseline without losing core local-first collection, profile, recovery, or market-watch data.

#### Scenario: Upgrade representative store with legacy media
- **GIVEN** a prior-release Cabinet store contains legacy inventory photos, Chat attachment files, mixed already-canonical assets, duplicate bytes, Unicode filenames and Windows-style absolute paths
- **WHEN** the current app runs the #1937 media migration during upgrade or explicit maintenance
- **THEN** upgrade evidence MUST report discovered, migrated, already-migrated, duplicate, skipped, failed and orphan counts
- **AND** user-visible inventory primary photos, display order, Chat attachment links, backup/restore behavior and profile isolation MUST remain intact after migration
- **AND** release smoke evidence MUST keep #1868 and #1869 acceptance blocked until this migration evidence is complete and reviewed.
