## ADDED Requirements

### Requirement: RUNTIME-CORE-026: Release candidates SHALL carry verifiable SBOM and build provenance
Cabinet SHALL bind each exact Windows release candidate to a deterministic,
machine-readable dependency inventory and hosted build provenance without
treating either as release approval or a security guarantee.

#### Scenario: Generate and package one deterministic SBOM
- **GIVEN** an exact clean Cabinet source commit, canonical release version, commit date, Go build information from the finished binaries, and production UI lock graph
- **WHEN** the Windows portable package is built
- **THEN** packaging MUST generate CycloneDX 1.7 JSON with exact source/version metadata, unique stable component identities, the Go modules recorded in the finished binaries, and non-development UI packages
- **AND** the same source, version, date, binary build information, and lock graph MUST produce byte-identical SBOM output
- **AND** the standalone versioned SBOM and embedded `CABINET-SBOM.cdx.json` MUST contain identical bytes
- **AND** UI development-only packages MUST NOT appear in the release SBOM.

#### Scenario: Bind SBOM evidence through the candidate manifests
- **GIVEN** the package generator produced the portable ZIP and standalone SBOM
- **WHEN** Cabinet writes and combines candidate manifests or creates packaged-acceptance identity
- **THEN** the evidence MUST record the SBOM filename, embedded path, format, specification version, predicate type, SHA-256, byte size, package subject SHA-256, and exact source commit
- **AND** any changed SBOM identity MUST change the candidate acceptance fingerprint
- **AND** missing, malformed, duplicate, source-drifted, checksum-drifted, or standalone-versus-embedded SBOM evidence MUST fail verification.

#### Scenario: Attest only explicit candidate artifacts
- **GIVEN** an operator explicitly dispatches the private package or exact release-candidate workflow
- **WHEN** the exact Windows portable ZIP and SBOM have passed repository verification
- **THEN** the hosted workflow MUST create build-provenance and CycloneDX SBOM attestations for that ZIP using least required token permissions
- **AND** ordinary pull-request, develop, and main validation MUST NOT create attestations or publish a release
- **AND** attestation creation MUST NOT imply code signing, a SLSA level, packaged acceptance, owner approval, or publication.

#### Scenario: Verify provenance before approved publication
- **GIVEN** packaged acceptance, recovery evidence, and exact #1864 approval nominate one successful candidate run and commit
- **WHEN** the approved prerelease workflow downloads the candidate
- **THEN** it MUST reverify the package, embedded and standalone SBOM, Cabinet and combined manifests, default build-provenance attestation, and `https://cyclonedx.org/bom` attestation before publication
- **AND** the immutable prerelease MUST include the standalone SBOM beside the package, checksum, manifests, and notes
- **AND** any missing or invalid attestation MUST stop publication without changing a release or promoting `develop` to `main`.
