## 1. Contract

- [x] 1.1 Define canonical per-asset folder, manifest and relative-path rules
  for #1936.
- [x] 1.2 Define lifecycle rules for immutable originals, deterministic
  renditions, shared references, orphan cleanup and restore.

## 2. Implementation Evidence

- [x] 2.1 Inventory photo uploads create `assets/<asset-id>/` folders with
  `original/`, `renditions/`, `variations/` and `manifest.json`.
- [x] 2.2 Media workspace and Chat attachment uploads use the shared media
  service and store media-root-relative paths.
- [x] 2.3 Telegram capture media routes through canonical asset ingestion when
  the media service is supplied.
- [x] 2.4 Backup/restore preserves canonical media asset folders and manifests.
- [x] 2.5 Inventory photo deletion preserves referenced canonical folders and
  removes orphan canonical folders.

## 3. Validation

- [x] 3.1 Run focused API/media/telegramcapture/backup validations recorded on
  issue #1936.
- [x] 3.2 Run strict OpenSpec validation for this change.
