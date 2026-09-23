## Why

Cabinet's current release controls can only create a `0.1.0-beta.10` private
prerelease. The approved 1.0 GA contract requires a distinct, immutable GA
identity and release path that cannot relabel beta artifacts as production.

## What Changes

- Introduce a canonical `1.0.0` GA release identity for Cabinet and Browser
  Companion portable artifacts.
- **BREAKING** Replace private-beta-only package, manifest, evidence-recorder,
  documentation, approval-marker, and publication assumptions for GA release
  controls.
- Add a fail-closed GA publisher requiring an exact GA candidate and exact
  owner approval before a non-prerelease release is created.
- Retain portable-only Windows x64 distribution and separate protected
  `develop` to `main` promotion approval.

## Capabilities

### New Capabilities

- `ga-release-controls`: Create, verify, approve, and publish an immutable
  Cabinet 1.0 GA portable release without accepting private-beta artifacts.

### Modified Capabilities

- `runtime-core`: Replace the private-beta packaging identity requirement with
  a canonical GA identity and truthful portable-distribution contract.
- `documentation-governance`: Align packaged and in-product guidance with the
  Cabinet 1.0 GA portable-only contract.
- `browser-companion`: Bind Chrome and Edge release package identity to the
  exact GA source and portable release contract.

## Impact

Release configuration, package generators and verifiers, candidate/evidence
controls, release workflows, protected-branch approval checks, Browser
Companion packaging, release notes, Help Center, Privacy/Terms, and their
contracts will change. A fresh exact candidate and separate installed
acceptance/recovery evidence remain required before publication.
