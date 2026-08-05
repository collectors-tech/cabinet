# Cabinet 0.1 beta data-safety evidence matrix

Issue: #1867 `P0(beta): validate database upgrade, backup, export and restore round trip`

Status as of 2026-08-06: implementation and automated regression evidence is in place for the release data-safety contracts. #1867 remains in release review until #1869 tests the exact private/internal Cabinet + Browser Companion candidate and attaches binary-level evidence to #1864.

Internal candidate creation does not require final #1864 approval. Final #1864 approval follows #1869/#1867 evidence and is required before external prerelease publication or `develop` to `main` promotion.

## Acceptance matrix

| #1867 acceptance item | Current evidence | Status |
| --- | --- | --- |
| Tests fail first for missing round-trip guarantee | PR #1905 recorded a red pre-implementation proof for saved filter snapshot round-trip before implementing `Snapshot.SavedFilters`; log `.work-agent/logs/issue-1867-saved-filter-roundtrip/go-test-datamgmt-saved-filter-red-20260712-0356.log`. | Covered |
| Current-`main`/prior-release fixture upgrades without data loss | PR #1903 added `TestOpenAndMigratePreservesRepresentativeLegacyReleaseData`, covering profile settings, license state, saved filters, inventory item/barcode/instance/photo relationships, wishlist state, Market Watch query/results, required migration columns, and scanner candidate relationship repair. | Covered |
| Backup is taken or explicitly offered before destructive migration/restore | PR #1896 added pre-restore backup creation/reporting on confirmed restore; `DATA-MANAGEMENT-004` now requires pre-restore ZIP backup metadata. | Covered |
| Exported data can be imported/restored and key counts/relationships match | PR #1901 covered item/barcode/instance relationship round-trip, PR #1902 covered media/photo references, and PR #1905 covered saved filter/view definitions through export -> clean import -> export. | Covered |
| Secrets are not leaked into normal export/diagnostic artifacts | PR #1900 covered active-profile data export scope plus non-leakage of stored profile secret values and raw license material through `TestDataExportsDoNotLeakProfileSecretsOrLicenses`. | Covered |
| Restore failure leaves prior workspace recoverable | PR #1904 added `TestRestoreRejectsIncompleteArchiveWithoutChangingActiveDatabase`; PR #1899 covered atomic restore replacement/recovery behavior. | Covered |
| Windows packaged-binary evidence is attached to #1864 | This is intentionally release-lane evidence, not a source-only PR claim. It remains owned by #1868/#1869 on the exact Cabinet + Browser Companion candidate, files, checksums and commit. | Pending packaged release acceptance |
| Zero data-loss defects remain open for the release candidate | No open #1867 data-loss implementation blocker is known after merged PRs #1896, #1897, #1899, #1900, #1901, #1902, #1903, #1904, and #1905. Final confirmation belongs to #1869 packaged core-workflow acceptance on the release candidate. | Pending packaged release acceptance |

## Merged proof slices

- PR #1896: pre-restore backup/restore recoverability.
- PR #1897: import dry-run/apply identity parity.
- PR #1899: atomic restore replacement recovery.
- PR #1900: profile-scoped exports and secret/license-safe export behavior.
- PR #1901: JSON/CSV export relationship round-trip.
- PR #1902: media reference round-trip.
- PR #1903: representative prior-release database upgrade fixture.
- PR #1904: corrupt/incomplete restore input recovery.
- PR #1905: saved filter/view snapshot round-trip.

## Release dependency

#1867 can be treated as source-level data-safety complete once this matrix is reviewed, but it should not be used to claim full beta release readiness until:

1. #1868 produces the versioned Windows beta artifact and checksums.
2. #1869 runs packaged core-workflow acceptance against that artifact.
3. #1869/#1867 verify canonical media manifests and links through companion image sync, restart, relocation, backup and restore.
4. #1864 receives the exact release commit, Cabinet and companion artifacts/checksums, packaged data-safety, and rollback/recovery evidence before the final approval decision.
