# First Launch, Sign-in, and Data Setup

## Before starting

Verify the candidate ZIP checksum, extract it into a writable stable folder, and run `cabinet.exe`. The private beta is a Windows portable ZIP, not an installer.

For the default portable layout, Cabinet creates an extracted folder-local `data` directory beside `cabinet.exe`. A `CABINET_DATA_DIR` environment override or custom first-run storage choice can use another directory. Open `/api/runtime` and use its `data_dir` value as the authoritative path for the running process.

## Choose a login and authentication mode

- **Local mode** is the default for an ordinary local beta workspace. Create or select a local profile and follow the displayed passkey or recovery requirements if enabled.
- **ZITADEL mode** appears only for a runtime configured with the deployment's ZITADEL issuer, client, audience, and redirect settings. The external authority handles sign-in; Cabinet uses the returned identity/session context for the local workspace.

Do not enter a made-up provider key or cloud token to make setup appear complete. If ZITADEL was expected but is unavailable, record the runtime version and contact the beta coordinator or deployment operator.

## Select the working profile

1. Open Cabinet and complete the displayed sign-in or local unlock flow.
2. In the top-left sidebar switcher, choose **Database**.
3. Select an existing profile or create a profile for this collection.
4. Confirm the selected profile before importing or creating real records.
5. Check Inventory, Wishlist, and Collections to verify the expected data context.

Profiles are workspace boundaries within the configured Cabinet runtime. Collections are grouping buckets inside a profile; they are not separate databases.

## Back up before changing builds

Use Settings Storage to create a backup and keep the backup outside the extracted Cabinet folder. Record the exact candidate version, source commit, and checksum. Deleting or replacing the whole portable folder can delete the default `data` directory contained in it.
