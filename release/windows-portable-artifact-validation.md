# Windows Portable Artifact Validation

Issue: #1868
Version: 0.1.0-beta.1
Release gate: #1864 approval required before publishing, attaching prerelease artifacts, or promoting `develop` to `main`.

This checklist records the non-publishing proof required before the private beta package can move to release approval. It validates the portable artifact generated from a clean Cabinet commit without creating or updating a GitHub prerelease.

## Build

- Run `pwsh -NoLogo -NoProfile -File .\scripts\package-installers.ps1` from a clean worktree.
- Confirm the generated package is `dist/cabinet-0.1.0-beta.1-windows-amd64-portable.zip`.
- Confirm the generated checksum file is `dist/cabinet-0.1.0-beta.1-windows-amd64-portable.zip.sha256`.
- Confirm release notes are generated at `dist/cabinet-0.1.0-beta.1-release-notes.md`.
- Confirm the ZIP contains `cabinet.exe`, `README.md`, and `WINDOWS-PORTABLE-BETA.md`.

## Runtime Smoke

- Start the packaged `cabinet.exe` from an extracted artifact folder with an isolated data directory.
- Verify `/healthz` returns `ok`.
- Verify `/api/runtime` reports `app_version=0.1.0-beta.1`.
- Record runtime port, process id, data directory, build revision, and build date in the run log.

## Release Gate

- Preserve package path, SHA-256, release notes path, runtime smoke output, and command logs under `.work-agent/logs/issue-1868-portable-artifact-validation/`.
- Must not publish a GitHub prerelease, attach artifacts to a release, or claim final Windows install/start acceptance until #1864 approval exists.
