## Context

Cabinet's release lanes already freeze a full source commit, reject dirty
worktrees, build a Windows portable ZIP, record package/file checksums, combine
Cabinet and Browser Companion manifests, and reverify the exact candidate before
publication. The missing supply-chain layer is a standard dependency inventory
and a hosted, signed statement that links the produced ZIP to the GitHub build.

The repository is public, so GitHub artifact attestations are available. The
normal PR quality gate must remain read-only and must not create attestations;
only the explicit private package and release-candidate dispatches produce
candidate evidence. Publication still requires the separate #1864 approval.

## Goals / Non-Goals

**Goals:**

- Produce deterministic CycloneDX 1.7 JSON for the exact Go module graph and
  production UI lock graph without adding a mutable package-time dependency.
- Put identical SBOM bytes inside the Windows ZIP and beside it as a candidate
  asset, with hashes bound into every downstream manifest/recorder.
- Fail closed when the SBOM is missing, malformed, source-drifted, duplicated,
  development-only, or checksum-drifted.
- Create and later verify GitHub-hosted build and SBOM attestations for the exact
  portable ZIP in explicit release workflows.

**Non-Goals:**

- Discovering or approving dependency licenses, vulnerabilities, or reachability.
- Claiming code signing, a Windows installer, a SLSA level, GA approval, or
  publication.
- Attesting ordinary PR validation packages.
- Changing product APIs, runtime behavior, provider scope, or user data.

## Decisions

1. **Generate CycloneDX 1.7 with a repository-owned Node script.** The generator
   consumes the Go build information recorded in the finished `cabinet.exe` and
   `cabinet-mcp.exe` binaries plus `ui.web/package-lock.json`, excludes UI
   entries marked development-only, converts available lock/module integrity
   values into CycloneDX hashes, gives every component a stable package URL and
   BOM reference, and sorts all output. The timestamp is the source commit date,
   so the same source/version/date produces identical bytes. This avoids a
   package contract that changes when an external SBOM generator updates.

   Alternatives considered: scanning only the final ZIP with Syft can miss
   dependencies bundled into the embedded UI; installing multiple ecosystem
   generators plus a merge tool adds mutable package-time tools and produces
   harder-to-test cross-tool drift.

2. **Describe shipped application dependencies, not the test toolchain.** Go
   modules recorded in the two shipped binaries and production-reachable npm lock
   entries are included. UI dev dependencies and repository-only test tools are
   excluded. The SBOM does not assert licenses when the authoritative lock/module
   data does not contain them.

3. **Embed and emit one byte-identical SBOM.** `CABINET-SBOM.cdx.json` is placed
   inside the portable archive and a versioned `*-sbom.cdx.json` is emitted next
   to the ZIP. The Cabinet release manifest records the external filename,
   embedded path, format, CycloneDX version, predicate type, SHA-256, size, and
   package subject digest. The combined candidate manifest and acceptance
   fingerprint carry the same object.

4. **Make verification structural and cryptographic.** Repository verification
   checks CycloneDX identity, exact source/version metadata, unique sorted
   components, Go and production-npm coverage, valid hashes, standalone/embedded
   byte equality, and all manifest digests. Unit fixtures exercise deterministic
   generation and fail-closed tampering before production code changes.

5. **Use GitHub's hosted attestations only for explicit artifacts.** The manual
   candidate and private package workflows grant `id-token: write` and
   `attestations: write`, then use a commit-pinned `actions/attest` release once
   for build provenance and once with the CycloneDX SBOM predicate. PR, develop,
   and main validation workflows retain read-only behavior and produce no
   attestations.

6. **Verify before approved publication.** The publication workflow downloads
   the exact successful candidate, reruns package/manifests verification, then
   uses `gh attestation verify` for default build provenance and the
   `https://cyclonedx.org/bom` predicate before attaching the standalone SBOM.
   Attestation success does not replace candidate acceptance or owner approval.

## Risks / Trade-offs

- **Repository-owned generation may lag future CycloneDX changes** → Pin 1.7 in
  the contract, validate its required structure, and upgrade only through a new
  issue/spec change.
- **Lock/module inventories do not prove runtime reachability** → Label the BOM
  as application dependency inventory and keep vulnerability/reachability review
  separate.
- **GitHub attestation service can be unavailable** → Explicit package and
  candidate workflows fail closed; PR validation remains unaffected.
- **New evidence invalidates old candidate artifacts** → Require a new exact
  candidate after merge rather than mutating or blessing an older ZIP.
- **Manifest additions affect acceptance fingerprints** → Include the SBOM in
  the combined manifest and recorder identity so a changed BOM resets evidence.

## Migration Plan

1. Land generator/verifier tests and manifest contract changes.
2. Update package generation, combined manifest, acceptance identity, and
   release workflows.
3. Merge only after the existing hosted package gate passes.
4. Produce a new exact candidate; older candidates remain historical and do not
   satisfy the new GA evidence contract.
5. Roll back by reverting the change and discarding any candidate produced under
   the reverted contract; never rewrite an existing candidate or release asset.

## Open Questions

None for this code lane. Signing keys, installer format, license adjudication,
and the final supported distribution contract remain explicit #2546 decisions.
