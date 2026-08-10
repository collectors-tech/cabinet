# Cabinet 0.1 Packaged Core-Workflow Acceptance Checklist

Issue: #1869
Package lane: Windows portable beta package from #1868 plus the exact #2034 Browser Companion package
Release gate: #1864 approval required before external prerelease publication or `develop` to `main`

This checklist runs against the private/internal candidate produced after the
source-ready gate. Internal candidate creation does not require final #1864 approval.
Final #1864 approval is required before external prerelease publication or `develop` to `main` promotion.
The final approval follows packaged acceptance and #1867 data-safety evidence.

## Candidate Identity

- [ ] OS version and host profile are recorded.
- [ ] Cabinet package filename is recorded.
- [ ] Cabinet package SHA-256 is recorded and matches its `.sha256` file.
- [ ] Cabinet source commit SHA is recorded.
- [ ] Successful Beta Release Candidate Gate run ID and exact artifact name are recorded.
- [ ] Cabinet, Browser Companion and combined candidate manifest paths all name the same source commit.
- [ ] Browser name/version and Browser Companion package filename are recorded.
- [ ] Browser Companion package SHA-256, source commit, extension version and release-manifest path are recorded.
- [ ] Browser Companion production identity, target, protocol compatibility and immutable candidate version match the release manifest; the Development source build is not used.
- [ ] `/api/runtime.app_version`, build date, runtime port, and pid are recorded.
- [ ] Cabinet and Browser Companion release notes paths are recorded.

## Required Collector Journey

- [ ] Fresh start and onboarding/profile setup complete from a clean Windows data directory.
- [ ] Inventory item can be created, edited, searched, filtered, reloaded, and verified after restart.
- [ ] Media can be attached, marked primary, and verified after restart.
- [ ] #1937 media migration evidence records discovered, migrated, already-migrated, duplicate, skipped, failed, and orphan counts from the packaged or explicit maintenance smoke run.
- [ ] Wishlist item can be created, reprioritised, status-updated, and marked purchased into Inventory.
- [ ] Collection can be created/edited, receive/move an item, soft-delete safely, and protect All Items.
- [ ] Data export and backup both complete with non-secret artefacts.
- [ ] Backup restore into an isolated target preserves core record counts and relationships.
- [ ] Discovery review can hand an item to Wishlist or Inventory without ownership confusion.
- [ ] One failed provider and one invalid import/restore input show useful recovery/error behavior.

## Required Provider and Companion Journey

- [ ] Install the exact Chrome and Edge packages through the documented beta path without developer source tools.
- [ ] Upgrade and rollback use only verified versioned packages, preserve visible pending jobs, revoke stale origins and fail closed on checksum or protocol mismatch.
- [ ] Pair to Cabinet through #2033 and verify reconnect, credential rotation, revoke-one and revoke-all.
- [ ] Enabled browser-capable integration changes propagate from Cabinet without rebuilding the extension.
- [ ] Cabinet/provider open-focus and ready, login-required, action-required, partial, selector-drift and disconnected states are truthful.
- [ ] A real saved Market Watch and Discovery hand-off pass independently for Voglers.
- [ ] A real saved Market Watch and Discovery hand-off pass independently for Hobbytech.
- [ ] A user-present real search, persisted observation and Discovery hand-off pass independently for Frontline.
- [ ] A user-present real search after normal browser interaction, persisted observation and Discovery hand-off pass independently for Bonza.
- [ ] A stalled or unavailable Frontline request returns within the bounded provider timeout, records no candidates or false success, and leaves the next provider run usable.
- [ ] A stalled or unavailable Bonza request returns within the bounded provider timeout, records no candidates or false success, and leaves the next provider run usable.
- [ ] Failure of one provider does not prevent, mutate or corrupt another provider's watches or observations.
- [ ] Replaying one capture proves item and media idempotency with transport/module/schema provenance.
- [ ] One durable protected-provider image uses the canonical asset manifest/layout and survives restart, backup, relocation and restore.
- [ ] Browser-closed, Cabinet-restart and extension-service-worker recovery resume without duplicate observations.

## Cross-Cutting Proof

- [ ] Persistence is verified after reload and application restart.
- [ ] Active-profile isolation is verified for at least one created record, companion session and export/restore path.
- [ ] No raw translation keys, placeholder security claims, or unsigned-installer claims appear in release UI/docs.
- [ ] No cookie/token export, challenge solving, hidden crawling or silent inventory mutation occurs.
- [ ] Empty and error states are useful enough for a beta user to recover or report the issue.
- [ ] Exact Cabinet and extension versions/commits are visible in recorded evidence.

## Failure Handling

- [ ] Every failure creates or links a focused GitHub issue with route/surface, expected behavior, actual behavior, repro steps, evidence, requirement link, and planned validation target.
- [ ] Release-blocking failures are linked back to #1864 and #1869 before rerun.
- [ ] The acceptance pack is rerun after release-blocking fixes against a new exact candidate commit.
- [ ] Final evidence explicitly states pass, fail with blockers, or not run, without using visual toasts or redirects as persistence proof.
- [ ] If all gates pass, the proposed #1864 approval comment records `APPROVE CABINET 0.1 PRIVATE BETA <exact-commit>`; publication is not invoked by this checklist.

## Prohibited Shortcuts

- [ ] Final packaged acceptance uses the packaged binary, not a dev server.
- [ ] Final packaged acceptance does not require test-only hooks.
- [ ] Final packaged acceptance does not rely on a dirty worktree or unpublished local changes.
- [ ] Final packaged acceptance does not merge `develop` into `main` or publish a release without #1864 approval.

The source precondition for the two timeout rows is #2054 / `PROVIDER-FAMILY-011`:
the shared commerce-provider client must pass the stalled-header, partial-body,
request-cancellation, full-bounds and same-client recovery contract tests before
an exact candidate is nominated. Packaged evidence remains required; source tests
alone do not check either row.
