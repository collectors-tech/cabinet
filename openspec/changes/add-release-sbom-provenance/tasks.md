## 1. Test-first contracts

- [x] 1.1 Add deterministic CycloneDX generation and tamper-rejection tests that fail against the current package contract.
- [x] 1.2 Add release-workflow contract assertions for explicit attestations, publication verification, SBOM upload, and no PR attestations.

## 2. SBOM generation and verification

- [x] 2.1 Implement a deterministic CycloneDX 1.7 generator for the exact Go module graph and production UI lock graph.
- [x] 2.2 Implement fail-closed CycloneDX identity, source, component, ordering, duplicate, and hash verification.

## 3. Package and candidate evidence chain

- [x] 3.1 Embed the SBOM in the Windows package and emit the same bytes as a standalone versioned file.
- [x] 3.2 Bind SBOM identity and checksums into Cabinet and combined candidate manifests and package verification.
- [x] 3.3 Include SBOM identity in packaged-acceptance candidate verification and fingerprinting.

## 4. Hosted provenance

- [x] 4.1 Add least-privilege build and CycloneDX SBOM attestations to explicit candidate/package workflows only.
- [x] 4.2 Make approved publication verify both attestations and publish the standalone SBOM asset.

## 5. Governance and validation

- [x] 5.1 Add RUNTIME-CORE-026 traceability and user-facing verification/limitation guidance.
- [x] 5.2 Run focused red/green tests, release package contracts, full affected tests, strict OpenSpec validation, and diff checks.
- [x] 5.3 Commit, push, open the issue-linked PR, wait for terminal hosted checks, merge, and reconcile #2550/Project #2.
