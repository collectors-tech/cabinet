# Windows Portable Install and Upgrade Validation

Issue: #1868
Version: 0.1.0-beta.7
Release gate: validate the private/internal candidate before approval; require #1864 approval before external publication or promotion.

Internal candidate creation does not require final #1864 approval.
Final #1864 approval is required before external prerelease publication or `develop` to `main` promotion.

This checklist records the non-publishing Windows validation required before #1869 acceptance can pass. It complements `release/windows-portable-artifact-validation.md` by proving the exact frozen-commit artefact can start cleanly and can be run against an existing data directory with backup and rollback evidence.

## Candidate Identity

- Build or nominate `dist/cabinet/cabinet-0.1.0-beta.7-windows-amd64-portable.zip` from a clean commit.
- Record the matching `.sha256` value, release notes path, commit SHA, build date, and package extraction path.
- Record the Cabinet release manifest, Browser Companion release manifest, combined candidate manifest and successful exact candidate workflow run ID.
- Record the paired #2034 Browser Companion filename, SHA-256, source commit, manifest and browser version.
- Confirm the extracted package contains `cabinet.exe`, `README.md`, and `WINDOWS-PORTABLE-BETA.md`.

## Clean Install and Start

- Extract the ZIP into a new writable folder on Windows.
- Start `cabinet.exe` with a new isolated data directory.
- Verify `/healthz` returns `200 ok`.
- Verify `/api/runtime` reports `app_version=0.1.0-beta.7`, the runtime port, process id, data directory, build date, and a full lowercase `build_revision` that exactly equals the Cabinet manifest `source_commit`.
- Complete the first-run/onboarding path far enough to prove the UI can create or open the local profile without test-only hooks.

## Existing Data Directory Upgrade

- Start from a representative existing Cabinet data directory and record its source commit/version.
- Create a backup before replacing or reusing the existing data directory.
- Start the portable beta candidate against that same data directory.
- Verify the runtime opens successfully and `/api/runtime` reports the beta version.
- Verify core data remains readable after restart: inventory item count, wishlist item count, collection membership count, saved filter/view count, backup/export availability, and profile identity.
- Record any migration or recovery warning exactly, then file a focused follow-up issue before rerun if data is missing or unreadable.

## Rollback and Evidence

- Preserve the pre-upgrade backup outside the extracted package folder.
- Confirm rollback instructions identify closing Cabinet, restoring the backup or prior data directory, and re-running the prior package if needed.
- Store command output, runtime JSON, checksum proof, screenshots or notes, and pass/fail summary under `.work-agent/logs/issue-1868-portable-upgrade-validation/`.
- Link the private/internal proof to #1869 and #1867 before the #1864 decision.
- After those gates pass, record the trusted #1864 comment ID containing `APPROVE CABINET 0.1 PRIVATE BETA <exact-commit>`; only the explicit prerelease workflow may consume it.
- Do not publish a GitHub prerelease, attach artefacts to an external release, or promote `develop` to `main` until final #1864 approval exists after packaged acceptance.
