# Cabinet Windows Portable Beta

Cabinet `0.1.0-beta.1` is packaged as a Windows portable ZIP until a signed installer is validated.

## Install and Start

1. Extract `cabinet-0.1.0-beta.1-windows-amd64-portable.zip` into a writable folder.
2. Run `cabinet.exe`.
3. Open the printed local runtime URL if the browser does not open automatically.

## Data Location

Cabinet stores local data under the configured runtime data directory. For default Windows runs, use the Settings storage screen and `/api/runtime` to verify the active `data_dir` before upgrade, backup, restore, or removal.

## Backup and Upgrade

Before replacing an existing beta build, run a backup from Settings Storage and keep the generated ZIP outside the extracted Cabinet folder. Start the new portable build against the same data directory only after recording the package checksum and commit.

## Uninstall or Rollback

Close Cabinet, keep or move the data directory as needed, and delete the extracted portable folder. To roll back, extract the prior portable package and restore from the saved backup if the data directory was changed during validation.

## Signing and Release Gate

This package is not a signed installer. Do not describe it as an installer and do not promote it to `main` or a public release without #1864 approval.
