# Cabinet Windows Portable GA

Cabinet `{{CABINET_BETA_VERSION}}` is distributed as a Windows portable ZIP. It is not an installer and is not code signed.

## Install, verify, and start

1. Compare the supplied SHA-256 checksum with `{{CABINET_PORTABLE_FILENAME}}` before extraction.
2. Extract `{{CABINET_PORTABLE_FILENAME}}` into a writable folder.
3. Run `cabinet.exe`, then use the printed local runtime URL if a browser does not open automatically.

## Data, backup, upgrade, and rollback

For a normal portable launch, Cabinet creates `data` beside `cabinet.exe`. `/api/runtime` reports the authoritative active `data_dir`; check it before upgrade, backup, restore, relocation, rollback, or removal.

Before replacing a release, make a Settings Storage backup and retain it outside the extracted Cabinet folder. To roll back, extract the prior portable package into a separate writable folder and use a compatible confirmed data directory or restore the saved backup.

## SBOM and provenance

The release includes `cabinet-{{CABINET_BETA_VERSION}}-sbom.cdx.json` beside the ZIP and identical bytes as `CABINET-SBOM.cdx.json` inside it. The manifest records the package and SBOM SHA-256 values and exact source commit. Hosted attestation verifies workflow provenance; it does not turn this portable ZIP into a signed installer.
