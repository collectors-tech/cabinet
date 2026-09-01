## Why

Cabinet release candidates currently carry checksums and file manifests but no
machine-readable dependency inventory or hosted build provenance. GA needs an
exact, independently verifiable link from the Windows package to its source,
build workflow, and runtime dependency inventory without treating that evidence
as release approval or a security guarantee.

## What Changes

- Generate a deterministic CycloneDX 1.7 JSON SBOM from the Go modules recorded
  in the finished binaries and the production `ui.web` lock graph.
- Embed the SBOM in the Windows portable package and emit the same bytes as a
  standalone candidate/release asset.
- Bind the SBOM identity and checksum into Cabinet and combined candidate
  manifests, packaged-acceptance identity, and fail-closed verifiers.
- Create GitHub build-provenance and CycloneDX SBOM attestations only in
  explicitly dispatched packaging/candidate workflows.
- Verify both attestations before approved prerelease publication and publish
  the standalone SBOM with the immutable release files.
- Document verification commands and the limits of SBOM/provenance evidence.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `runtime-core`: Require exact Cabinet release packages to contain and expose a
  deterministic CycloneDX SBOM, bind it through the candidate evidence chain,
  and use hosted attestations that are verified before publication.

## Impact

- Release packaging and verification scripts under `scripts/`.
- Cabinet, combined candidate, and packaged-acceptance manifest contracts.
- Explicit candidate, portable-package, and approved-publication workflows.
- Runtime-core OpenSpec requirements and traceability.
- No product API, database, provider, signing-key, release-publication, or
  `develop`-to-`main` behavior changes.
