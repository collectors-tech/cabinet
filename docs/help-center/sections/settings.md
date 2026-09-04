# Settings

## Use Settings for
- Profile/account preferences
- Display and notifications
- Storage and maintenance controls

## Storage section
- Configure media path
- Create, list, and restore local backups. Keep upgrade or rollback backups outside the extracted portable folder.
- Download active-profile JSON and item CSV exports after the storage context reports ready.
- Run diagnostics and maintenance actions where enabled.

## Diagnostics and retention

Remote diagnostics are disabled by default. With opt-in disabled, runtime and activity diagnostics remain local and Cabinet does not send a remote event. Enabling remote diagnostics also requires a configured remote endpoint; that endpoint applies its own availability and retention rules.

Export diagnostics logs as a recursively redacted snapshot. Cabinet removes known credentials, cookies, authorization headers, tokens, raw session identifiers, sensitive local paths, and private page content before local export or an opt-in remote diagnostics send. Redaction is a safety boundary, not a reason to share an export without reviewing it.

The beta has no fixed automatic retention period for local workspace data, backups, or diagnostics logs. Remove supported records and old backups through Cabinet where a control exists, or close Cabinet and remove the confirmed data directory when intentionally removing the complete local workspace. Deleting only the executable does not reliably delete data.

## Export, deletion, and support

Settings Operations and Settings Storage expose JSON, CSV, backup/restore, and redacted log paths. Exports and backups are new copies under your control; changing or deleting a source record does not recall copies you already saved or data previously sent to a provider, ZITADEL authority, or opted-in remote diagnostics endpoint.

For a beta support request, contact the coordinator who supplied the candidate and include its version, exact source commit, and a reviewed redacted export when useful. Never attach credentials, cookies, tokens, Browser Companion secrets, or an unreviewed database or backup.

## Categories and taxonomy
- Open **Categories** to manage profile-scoped item types, item type condition scales, and packaging grade values.
- Add condition scale values in the order collectors should see them in Inventory and Wishlist editors.
- Keep packaging grade labels consistent across the profile so search, filters, saved views, and API validation use the same taxonomy language.
- Existing records keep their saved values when you update the taxonomy, but new edits must use configured item type, condition, and packaging grade choices.
