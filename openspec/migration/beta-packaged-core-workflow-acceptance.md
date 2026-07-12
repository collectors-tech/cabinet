# Cabinet 0.1 Packaged Core-Workflow Acceptance Checklist

Issue: #1869
Package lane: Windows portable beta package from #1868
Release gate: #1864 approval required before prerelease publication or `develop` to `main`

This checklist is the stable packaged-candidate acceptance pack for Cabinet 0.1 private beta. It must be run against the exact Windows artefact, checksum, commit SHA, and app version intended for release.

## Candidate Identity

- [ ] OS version and host profile are recorded.
- [ ] Package filename is recorded.
- [ ] Package SHA-256 is recorded and matches the `.sha256` file.
- [ ] Source commit SHA is recorded.
- [ ] `/api/runtime.app_version`, build date, runtime port, and pid are recorded.
- [ ] Release notes path or prerelease draft URL is recorded.

## Required Journey

- [ ] Fresh start and onboarding/profile setup complete from a clean Windows data directory.
- [ ] Inventory item can be created, edited, searched, filtered, reloaded, and verified after restart.
- [ ] Media can be attached, marked primary, and verified after restart.
- [ ] Wishlist item can be created, reprioritised, status-updated, and marked purchased into Inventory.
- [ ] Collection can be created/edited, receive/move an item, soft-delete safely, and protect All Items.
- [ ] Data export and backup both complete with non-secret artefacts.
- [ ] Backup restore into an isolated target preserves core record counts and relationships.
- [ ] A saved Market Watch can run against the chosen live beta provider with non-secret provider evidence.
- [ ] Discovery review can hand an item to Wishlist or Inventory without ownership confusion.
- [ ] One failed provider and one invalid import/restore input show useful recovery/error behavior.

## Cross-Cutting Proof

- [ ] Persistence is verified after reload and application restart.
- [ ] Active-profile isolation is verified for at least one created record and one export/restore path.
- [ ] No raw translation keys, placeholder security claims, or unsigned-installer claims appear in release UI/docs.
- [ ] Empty and error states are useful enough for a beta user to recover or report the issue.
- [ ] Exact package version and commit are visible in runtime UI or `/api/runtime` evidence.

## Failure Handling

- [ ] Every failure creates or links a focused GitHub issue with route/surface, expected behavior, actual behavior, repro steps, evidence, requirement link, and planned validation target.
- [ ] Release-blocking failures are linked back to #1864 and #1869 before rerun.
- [ ] The acceptance pack is rerun after release-blocking fixes.
- [ ] Final evidence explicitly states pass, fail with blockers, or not run, without using visual toasts or redirects as persistence proof.

## Prohibited Shortcuts

- [ ] Final packaged acceptance uses the packaged binary, not a dev server.
- [ ] Final packaged acceptance does not require test-only hooks.
- [ ] Final packaged acceptance does not rely on a dirty worktree or unpublished local changes.
- [ ] Final packaged acceptance does not merge `develop` into `main` or publish a release without #1864 approval.

