## Why

Cabinet can now write new media into the canonical per-asset layout, but
existing user stores may still contain legacy inventory photo folders, Chat
attachment files, and absolute database paths. Release upgrade evidence cannot
be credible while those older references rely on machine-specific paths or
parallel storage layouts.

## What Changes

- Define an idempotent, resumable legacy media migration with preflight and
  dry-run reporting before mutation.
- Migrate inventory photos and Chat attachments into `media/assets/<asset-id>/`
  with manifests, owner/provenance metadata, relative database paths, and
  preserved user-visible relationships.
- Verify source and migrated originals by content hash before changing database
  references.
- Keep legacy files until the database commit, verification, and a recoverable
  checkpoint succeed.
- Report missing, corrupt, duplicate, locked, already-migrated, and orphan
  inputs with record identifiers and recovery guidance.
- Require backup/restore and profile relocation evidence after migration.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `ui-screen-media`: Media storage migration reporting, idempotence and orphan
  audit behavior.
- `photos-media`: legacy inventory photo paths migrate into canonical assets
  without changing primary photo or display-order behavior.
- `data-management`: backup, restore and upgrade evidence must preserve
  migrated canonical media under relocated Windows profile paths.

## Impact

- Affected code: `internal/media`, `internal/app`, `internal/backup`, Settings
  Storage UI and upgrade/package smoke fixtures.
- Affected data: legacy `item_photos` and `chat_attachments` rows may be
  rewritten from absolute or legacy-relative paths to media-root-relative
  canonical asset paths.
- Operational risk: data migration must remain dry-runable, resumable and
  conservative about deletion.
- Related issue: `#1937`.
