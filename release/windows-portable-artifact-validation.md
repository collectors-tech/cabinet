# Windows Portable Artifact Validation

Issue: #1868
Version: 0.1.0-beta.7
Release gate: build private/internal evidence before approval; require #1864 approval before external publication or promotion.

Internal candidate creation does not require final #1864 approval.
Final #1864 approval is required before external prerelease publication or `develop` to `main` promotion.

This checklist records the non-publishing proof required before the private beta package can move to packaged acceptance and then release approval. It validates the portable artifact generated from a clean, frozen Cabinet commit without creating or updating a GitHub prerelease.

## Build

- Run `pwsh -NoLogo -NoProfile -File .\scripts\package-installers.ps1` from a clean worktree.
- Confirm the generated package is `dist/cabinet/cabinet-0.1.0-beta.7-windows-amd64-portable.zip`.
- Confirm the generated checksum file is `dist/cabinet/cabinet-0.1.0-beta.7-windows-amd64-portable.zip.sha256`.
- Confirm release notes are generated at `dist/cabinet/cabinet-0.1.0-beta.7-release-notes.md`.
- Confirm `cabinet-release-manifest.json` records the exact commit, build date, package/checksum filenames and every packaged file SHA-256/size.
- Confirm the ZIP contains `cabinet.exe`, `cabinet-mcp.exe`, `README.md`, and `WINDOWS-PORTABLE-BETA.md`.
- Record the exact #2034 Browser Companion filenames, versions, source commit, release manifest and separate SHA-256 checksums nominated with this Cabinet candidate.
- Record `beta-candidate-bundle-manifest.json` and confirm its Cabinet and Browser Companion components name the same exact source commit.

## Runtime Smoke

- Start the packaged `cabinet.exe` from an extracted artifact folder with an isolated data directory.
- Verify `/healthz` returns `ok`.
- Verify `/api/runtime` reports `app_version=0.1.0-beta.7` and its full lowercase `build_revision` exactly equals `cabinet-release-manifest.json` `source_commit`.
- Record runtime port, process id, data directory, build revision, and build date in the run log.

## Release Gate

- Preserve the successful candidate workflow run ID, package path, SHA-256, Cabinet/companion/combined manifests, release notes path, runtime smoke output, and command logs under `.work-agent/logs/issue-1868-portable-artifact-validation/`.
- Preserve the Cabinet and companion files as private/internal acceptance artefacts for #1869.
- Must not publish a GitHub prerelease, attach artefacts to an external release, or promote `develop` to `main` until final #1864 approval exists after packaged acceptance.
