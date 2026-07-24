## 1. Contract

- [x] 1.1 Define the legacy inventory-photo and Chat-attachment migration
  contract for #1937.
- [x] 1.2 Define operator-visible Settings Storage reporting, repair, rollback
  and orphan-audit states.

## 2. Implementation Evidence

- [x] 2.1 Add migration preflight/dry-run reporting for legacy inventory photo
  folders and Chat attachment files.
- [x] 2.2 Add idempotent migration apply for inventory photos, preserving
  primary-photo/display-order and source hash evidence.
- [x] 2.3 Add idempotent migration apply for Chat attachments, preserving
  thread/message attachment links and filename/MIME metadata.
- [ ] 2.4 Add interruption recovery so staged assets are not exposed and repeat
  runs resume without duplicating records.
- [ ] 2.5 Add duplicate, missing, corrupt, locked, already-migrated and orphan
  classification without silently deleting unknown files.

## 3. UI and Operations

- [ ] 3.1 Surface migration status, counts, blockers, recovery actions and
  rollback guidance in Settings > Storage.
- [ ] 3.2 Record package/upgrade smoke evidence with discovered, migrated,
  skipped, failed and orphan counts.

## 4. Validation

- [ ] 4.1 Cover mixed legacy/new stores, duplicate bytes, missing files, corrupt
  files, Unicode names and Windows path edge cases.
- [ ] 4.2 Verify backup/export/restore and profile relocation after migration.
- [ ] 4.3 Run strict OpenSpec validation and record results on #1937.
