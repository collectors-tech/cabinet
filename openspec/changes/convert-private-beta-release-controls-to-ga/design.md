## Context

Current candidate and publisher controls are intentionally private-beta-only.
The approved GA contract is a Windows x64 portable `1.0.0` release; a beta
artifact must never be eligible for GA publication.

## Goals / Non-Goals

**Goals:** establish one GA identity across generated artifacts, verify it in
the candidate and publisher, and preserve exact approval and promotion gates.

**Non-Goals:** signed installers, browser stores, automatic updates, provider
challenge bypass, or replacing installed acceptance/recovery evidence.

## Decisions

- Use `release/cabinet-release-version.json` as the GA identity source. This
  prevents a publisher-only version from drifting from package manifests.
- Retain a distinct beta workflow and add a GA publisher that verifies manifest
  channel, version, publication state, exact candidate SHA, and exact approval.
- Publish portable-only Windows artifacts as a non-prerelease only after a
  fresh GA candidate; do not convert historical beta artifacts.

## Risks / Trade-offs

- [Version/config drift] → package, bundle, recorder, and publisher contracts
  assert the same canonical identity.
- [Accidental beta publication] → GA publisher rejects non-GA manifests before
  release creation.
- [Candidate invalidation] → every conversion merge requires a fresh candidate
  and new installed acceptance/recovery evidence.

## Migration Plan

Merge the controls under protected `develop`, build a fresh GA candidate, run
installed acceptance/recovery, record exact approval, publish, independently
replay assets, then promote to `main` only through its separate approval gate.
