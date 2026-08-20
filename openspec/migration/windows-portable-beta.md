# Cabinet Windows Portable Beta

Cabinet `{{CABINET_BETA_VERSION}}` is packaged as a Windows portable ZIP until a signed installer is validated.

## Install and Start

1. Extract `{{CABINET_PORTABLE_FILENAME}}` into a writable folder.
2. Run `cabinet.exe`.
3. Open the printed local runtime URL if the browser does not open automatically.

## Data Location

For a normal portable launch, Cabinet creates a `data` directory beside `cabinet.exe`. This executable-local directory is the default runtime root and normally contains the database, runtime configuration, media, backups, and logs.

`CABINET_DATA_DIR` can override the runtime root before startup, `CABINET_DB_PATH` can override the database file, and first-run setup can record another writable storage location. The `data_dir` reported by `/api/runtime` is authoritative for the running process. Check it before upgrade, backup, restore, relocation, rollback, or removal instead of assuming the default path.

## Backup and Upgrade

Before replacing an existing beta build, run a backup from Settings Storage and keep the generated ZIP outside the extracted Cabinet folder. Start the new portable build against the same confirmed data directory only after recording the package checksum and commit. Confirm the active profile and representative Inventory, Wishlist, Collection, and saved-view data after restart.

## Uninstall or Rollback

Close Cabinet before moving or deleting files. With the executable-local layout, deleting the whole extracted folder also deletes the default `data` directory inside it; deleting only `cabinet.exe` does not reliably remove data from an overridden or custom location. Keep or move the confirmed data directory as needed.

To roll back, extract the prior portable package into a separate writable folder and start it only with a compatible confirmed data directory. Restore from the saved backup if the data directory was changed during validation.

## Signing and Release Gate

This package is not a signed installer. Do not describe it as an installer and do not promote it to `main` or a public release without #1864 approval.
