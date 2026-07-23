## Why

Cabinet media ingestion has historically used separate storage layouts for
inventory photos, chat attachments and Media workspace uploads. That makes
profile relocation, backup/restore, shared ownership and future generated
variations fragile because records can point at machine-specific paths or
duplicated binaries.

## What Changes

- Define one canonical per-asset directory under the active profile media root.
- Require a versioned manifest with immutable original metadata, deterministic
  rendition metadata, owner/provenance links and variation placeholders.
- Store new database references as media-root-relative asset paths and resolve
  them against the active profile media root at runtime.
- Route inventory, Media workspace, Chat attachment and Telegram capture
  ingestion through the shared asset-store contract.
- Preserve canonical asset folders while any database reference remains and
  remove orphan canonical folders only after references are gone.
- Include canonical asset folders and manifests in backup/restore round trips.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `photos-media`: inventory photo ingestion and deletion use canonical asset
  folders and shared retention rules.
- `ui-screen-media`: Media workspace ingestion, linkage and download behavior
  use canonical asset identity and relative paths.
- `data-management`: backup/restore includes canonical media asset trees and
  restores them under the configured data directory.

## Impact

- Affected code: `internal/media`, `internal/app`, `internal/telegramcapture`,
  `internal/backup`.
- Affected data: new media paths are stored relative to the active media root;
  legacy absolute and URL-like paths remain compatible.
- Affected operations: profile relocation and backup/restore preserve canonical
  media references.
- Related issue: `#1936`.
