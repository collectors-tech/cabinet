# Trust Network Signed-Object Authority

This document defines which signed Cabinet object is authoritative for each
post-beta trust-network object type. Transports, mirrors, indexes, catalogues,
and peer sessions may carry or verify these objects, but they do not become the
authority themselves.

## Shared Authority Rules

- Every authority object uses the Cabinet signing envelope from
  `identity-suite.md`: canonical JSON, object type, schema version, key ID,
  algorithm, content hash, issued time, optional expiry, and detached signature.
- Each object records `object_id`, `object_type`, `schema_version`,
  `subject_refs`, `issuer_identity_id`, `issuer_key_id`, `issued_at`,
  `previous_state_hash`, and optional `supersedes`, `expires_at`,
  `revocation_ref`, and `appeal_ref` fields.
- Readers must verify the envelope, issuer key purpose, object schema,
  referenced identity state, revocation state, and previous-state hash before
  treating the object as usable evidence.
- Unknown critical fields, unsupported schema versions, broken signatures,
  revoked issuer keys, or unresolved authority conflicts must fail closed.
- Deleting a mirror row, repository commit, DHT entry, local queue item, or peer
  copy never deletes the signed authority object if another valid copy exists.

## Object Authority Matrix

| Object type | Canonical authority | Required signer purpose | Supersession and revocation | Non-authoritative carriers |
| --- | --- | --- | --- | --- |
| `cabinet.receipt.v1` | Countersigned receipt object records participant acknowledgements, item/specimen refs, transfer state, and external consideration notation only. | Participant object-signing keys; optional custodian/store attestation references are separate objects. | Superseded by correction or cancellation receipt that names the previous receipt hash; revoked only by signed dispute, fraud, or compromised-key evidence. | QR payloads, manual export files, peer session copies, local queue rows, Git/Radicle mirrors. |
| `cabinet.feedback.v1` | Signed feedback object tied to a receipt or proposal ref, outcome, rating band, evidence refs, and visibility scope. | Feedback author's object-signing key. | Superseded by amended feedback; revoked by author withdrawal, moderation/governance decision, fraud proof, or compromised-key evidence. | Feedback vault index, reputation cache, mirror commit, search projection. |
| `cabinet.attestation.v1` | Scoped claim from a store, custodian, catalogue maintainer, or governance actor about identity, item, receipt, node, or catalogue state. | Attestation, governance-signing, catalogue-signing, or node-signing key matching the issuer role and declared scope. | Superseded by newer attestation for the same scope or revoked by issuer/governance revocation. | Store profile page, catalogue metadata, registry projection, mirror index. |
| `cabinet.endorsement.v1` | Directed signed trust edge from one identity to another with scope, weight, expiry, and anti-Sybil attenuation metadata. | Endorser object-signing or governance-signing key with endorsement scope. | Superseded by a new edge version; revoked by endorser withdrawal, expiry, compromised key, or governance action. | Web-of-trust graph cache, reputation summary, registry index. |
| `cabinet.revocation.v1` | Signed revocation object naming the affected identity, key, node, object, attestation, endorsement, catalogue release, or mirror. | Revocation, recovery, governance-signing, catalogue-signing, or issuer object-signing key allowed by the affected object scope. | Revocation is append-only; appeal or recovery may supersede its operational effect but cannot erase the historical record. | Registry row, mirror tombstone, local warning cache, notification inbox. |
| `cabinet.mirror.v1` | Signed mirror manifest declaring provider, repository/ref, published object hashes, freshness, retention, and retirement metadata. | Mirror operator node-signing key plus any required governance/catalogue-signing approval for public registries. | Superseded by a newer mirror manifest or revoked/retired by operator or governance authority. | GitHub/GitLab/Codeberg/Radicle repository state, DHT advert, local mirror cache. |
| `cabinet.catalogue_manifest.v1` | Signed catalogue release manifest with version, bundle hash tree, item hash refs, maintainer identity, compatibility, rollback policy, and mirror refs. | Catalogue maintainer catalogue-signing key; optional governance-signing approval for official community releases. | Superseded by a higher compatible release; revoked by maintainer or governance action for poisoning, rollback, or compromise. | Download bundle, torrent/IPFS-like shard, Git release, Radicle mirror, Merkle proof cache. |

## Boundary Decisions

- A local private database remains authoritative for private inventory fields
  until the user explicitly signs and exports a selected claim.
- A signed receipt is authoritative for trade evidence, but it does not mutate
  another participant's local ownership state without local acceptance and
  conflict checks.
- Feedback and endorsements may inform ongoing reputation only after the
  referenced signed objects validate and any revocation/appeal state is applied.
- Attestations prove only the issuer's scoped claim; they do not transfer item
  ownership, identity control, or registry membership by themselves.
- Mirror manifests prove publication/freshness for a selected object set; they
  do not make the hosting provider or Git history authoritative.
- Catalogue manifests prove bundle integrity and maintainer intent; Merkle
  proofs verify inclusion, not truth of the underlying claims.
