# Trust Network Threat Model Mapping

This document maps #1957 trust-network threats to required mitigations,
architecture test vectors, and user-visible recovery behavior. It is a
post-beta architecture contract only; implementation issues must reuse these
rows instead of inventing silent trust behavior.

## Threat Mapping Matrix

| Threat | Applies to | Required mitigation | Architecture test vector | User-visible recovery or warning |
| --- | --- | --- | --- | --- |
| Fake peer | libp2p sessions, QR/manual handoff, offline queues, direct local exchange. | Cabinet identity handshake, signed object verification, scoped capability list, replay protection, and local acceptance before state mutation. | `TV-THREAT-001-fake-peer`: an unknown peer presents valid transport metadata but no valid Cabinet identity/signature chain; signed objects are rejected or held for review and no local state mutates. | Show "unknown peer" or "identity not verified", preserve received evidence, offer retry with verified identity, QR/manual fallback, or discard. |
| Fake store/custodian | Store attestations, identity validation, catalogue trust, dispute evidence. | Store/custodian attestations are signed scoped objects tied to current registry and key-purpose state; they never replace participant signatures. | `TV-THREAT-002-fake-store`: a forged store attestation or unregistered custodian cannot raise identity validation or transaction confidence. | Show "attestation could not be verified", keep the trade usable without store boost, and offer manual evidence review. |
| Sybil/eclipse | DHT discovery, peer graph, reputation propagation, endorsement graph, mirror discovery. | DHT adverts are locator-only, graph edges are signed and attenuated, registry quorum is versioned, and discovery diversity is required before claiming public state. | `TV-THREAT-003-sybil-eclipse`: many new identities or one DHT neighborhood cannot inflate reputation, hide revocations, or become authoritative. | Show degraded discovery, limited reputation confidence, mirror diversity warning, and allow alternate mirror/manual import. |
| Tampering | Signed receipts, feedback, revocations, mirror manifests, catalogue bundles, Merkle proofs, queued events. | Canonical serialization, signatures, previous-state hashes, Merkle inclusion/non-inclusion proofs, freshness checks, and fail-closed schema validation. | `TV-THREAT-004-tampered-object`: a bit-flipped receipt, stale mirror root, or mismatched Merkle proof is rejected while the failing proof is retained. | Show "tamper evidence detected", block trust use, keep a copy for dispute/debugging, and suggest another mirror or original participant. |
| Forged or deleted feedback | Feedback vaults, reputation cache, moderation, public mirrors. | Feedback is append-only signed evidence with receipt/proposal refs; deletion is a signed withdrawal/tombstone or moderation action, not physical erasure of proof. | `TV-THREAT-005-feedback-forgery-delete`: unsigned feedback and deleted mirror entries do not change reputation; signed withdrawal changes current view while preserving history. | Show feedback provenance, withdrawal/moderation state, and explain why a rating is ignored or still historically visible. |
| Key compromise | Identity objects, receipts, feedback, endorsements, registry, catalogue and node signatures. | Effective-time revocation, recovery key/governance/custodian recovery, key-purpose checks, timestamp context, and blocking post-revocation objects from compromised keys. | `TV-THREAT-006-key-compromise`: objects signed after revocation effective time fail closed; pre-revocation evidence remains timestamped but downgraded where relevant. | Show compromised-key warning, recovery/appeal state, affected object scope, and require re-signing or manual resolution. |
| Catalogue poisoning | Catalogue manifests, bundle distribution, mirror releases, Merkle roots, search projections. | Catalogue signing keys, governance approval where required, Merkle root verification, rollback/freshness policy, compatibility checks, and poisoned-release revocation. | `TV-THREAT-007-catalogue-poisoning`: a bundle with a valid transport but invalid manifest signature, stale root, or mismatched item hash is rejected. | Show bundle rejection, release revocation or stale-version warning, preserve hash evidence, and offer another release/mirror. |
| Privacy leakage | Public mirrors, registry records, hash-only publication, participant sharing, recovery/appeal records, logs. | Field-level visibility classes, signed preview, recipient/destination selection, redaction, salted/non-reversible hash commitments, and leakage revocation workflow. | `TV-THREAT-008-privacy-leakage`: private location/contact/value/payment/recovery data included in a public or hash-only object fails publication validation. | Block publication, name the offending field class, offer redaction, withdrawal/revocation, and safer hash-only or participant-scoped alternatives. |

## Follow-up Test Rules

- Every implementation issue opened from #1957 must reference at least one
  `TV-THREAT-*` row when it touches peer exchange, registry, mirrors,
  signatures, feedback, catalogue bundles, reputation, publication, or recovery.
- Tests must assert both the machine decision and the user-facing state: block,
  degraded, retry, manual review, revocation, appeal, redaction, or alternate
  mirror/import.
- Threat tests must use deterministic fake peer/store, fake mirror, fake DHT,
  fake catalogue bundle, and fake registry inputs. They must not require live
  public networks or real payment/provider credentials.
- Logs, serialized evidence, and user-visible explanations must avoid secrets,
  private inventory fields, private contacts, precise locations, collection
  values, full payment details, recovery contacts, and raw provider payloads.

## Recovery State Vocabulary

- `blocked_invalid_signature`
- `blocked_unknown_peer`
- `blocked_unverified_attestation`
- `blocked_privacy_leakage`
- `blocked_compromised_key`
- `degraded_discovery`
- `degraded_stale_mirror`
- `manual_review_required`
- `redaction_required`
- `revocation_or_appeal_pending`
- `alternate_mirror_required`
