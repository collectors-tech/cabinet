# Cabinet 0.1 Packaged Core-Workflow Acceptance Checklist

Issue: #1869
Package lane: Windows portable beta package from #1868 plus the exact #2034 Browser Companion package
Release gate: #1864 approval required before external prerelease publication or `develop` to `main`

This checklist runs against the private/internal candidate produced after the
source-ready gate. Internal candidate creation does not require final #1864 approval.
Final #1864 approval is required before external prerelease publication or `develop` to `main` promotion.
The final approval follows packaged acceptance and #1867 data-safety evidence.

For a practical Windows operator sequence, evidence-folder layout, Chrome/Edge
split, stop conditions, and handback steps, use
`openspec/migration/cabinet-1.0-ga-second-pc-test-plan.md`. The checklist on this
page remains the stable recorder contract.

## Resumable evidence recorder

Use `scripts/record-beta-acceptance.mjs` (or `npm run acceptance:record --`) to
create the repository-owned JSON and Markdown evidence pack. `init` verifies the
three exact candidate manifests, the Cabinet archive, both Chrome and Edge
Browser Companion archives, and every separate `.sha256` file before it reads or
resumes prior evidence. The generated candidate fingerprint binds the manifests,
source commit, versions, package filenames, package bytes, candidate-gate run and
artifact name. A stale candidate is archived under a fingerprinted filename and
all 51 rows restart as `not_run`; an in-place checksum mismatch fails closed.

Each stable row below is represented exactly once as `not_run`, `blocked`, `pass`, or `fail`.
Use `record --row <stable-id>` to update one row. `pass` and `fail` require one or
more non-secret `--evidence` references plus operator `--notes`. `blocked` requires an exact `--unblock`
condition. A human workflow can reach `pass` only with `--operator-confirmed`;
the recorder never operates the browser, provider or packaged UI and therefore
cannot auto-pass Frontline, Bonza, install, UI or recovery evidence.

The recorder redacts credentials, bearer/cookie material, explicitly marked
private page content and sensitive local paths. The raw isolated data-directory
path is represented by a redacted display value plus SHA-256 identity. Run
`validate` before attaching the deterministic JSON and Markdown outputs to
#1869, #1867 and #1864. The CLI has no release, branch-promotion, provider
interaction or browser-automation operation.
Overall output is exactly `not_run`, `fail_with_blockers`, or `pass`; the
per-row state preserves whether the blocker was `blocked` or `fail`.

```text
node scripts/record-beta-acceptance.mjs --help
node scripts/record-beta-acceptance.mjs init <candidate and environment options>
node scripts/record-beta-acceptance.mjs record --json <state.json> --markdown <summary.md> --row <id> --status <status> <evidence options>
node scripts/record-beta-acceptance.mjs validate --json <state.json>
node scripts/record-beta-acceptance.mjs render --json <state.json> --markdown <summary.md>
```

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
- [ ] `/api/runtime.app_version` and full `/api/runtime.build_revision`, build date, runtime port, and pid are recorded; build revision equals the Cabinet manifest `source_commit`.
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

The #2062 real-runtime Cypress preflight proves that both unconfigured providers can be selected with native pointer/keyboard input, saved as enabled active-profile integration instances, reopened with retained non-secret values, scoped in a saved Market Watch query, and projected into the paired Browser Companion module registry. This is configuration-path evidence only; it does not satisfy either external user-present live-search checkbox below.

- [ ] Install the exact Chrome and Edge packages through the documented beta path without developer source tools.
- [ ] Upgrade and rollback use only verified versioned packages, preserve visible pending jobs, revoke stale origins and fail closed on checksum or protocol mismatch.
- [ ] Pair to Cabinet through #2033 and verify reconnect, credential rotation, revoke-one and revoke-all.
- [ ] Enabled browser-capable integration changes propagate from Cabinet without rebuilding the extension.
- [ ] Cabinet/provider open-focus and ready, login-required, action-required, partial, selector-drift and disconnected states are truthful.
- [ ] A real saved Market Watch and Discovery hand-off pass independently for Voglers.
- [ ] A real saved Market Watch and Discovery hand-off pass independently for Hobbytech.
- [ ] A user-present real Frontline search persists an observation, appears through `GET /api/discovery/not-in-collection`, accepts reviewed `add_to_wishlist`, and persists exactly one linked Wishlist row visible through `GET /api/wishlist`.
- [ ] A user-present real Bonza search after normal browser interaction persists an observation, appears through `GET /api/discovery/not-in-collection`, accepts reviewed `add_to_wishlist`, and persists exactly one linked Wishlist row visible through `GET /api/wishlist`.
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
