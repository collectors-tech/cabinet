## 1. Architecture Contract

- [x] 1.1 Create the OpenSpec proposal, task plan, and initial decision matrix
  for #1957.
- [x] 1.2 Add canonical glossary entries for identity, specimen, catalogue
  item, proposal, reservation, receipt, ownership event, feedback,
  attestation, endorsement, registry, and node.
- [x] 1.3 Complete the source-area decision matrix with canonical rule,
  rejected alternative, unresolved decision, security/privacy impact, and
  delivery phase for every source area listed in #1957.

## 2. Trust and Authority Models

- [ ] 2.1 Define the canonical identity, signature, key-rotation, revocation,
  recovery, and compatibility suite without assuming secure enclave support.
- [ ] 2.2 Separate identity validation, transaction confidence, and reputation
  models with bounded ranges and deterministic test vectors.
- [ ] 2.3 Define signed-object authority for receipts, feedback,
  attestations, endorsements, revocations, mirrors, and catalogue manifests.
- [ ] 2.4 Define privacy/publication boundaries for private local records,
  participant-shared records, public claims, and hash-only publication.

## 3. Transport, Ledger, and Governance Decisions

- [ ] 3.1 Document Git, Radicle, libp2p, DHT, CRDT, and Merkle roles as
  storage, transport, or verification components, not implicit authority.
- [ ] 3.2 Define public-registry bootstrap, governance quorum, revocation,
  retirement, appeal, and compromised-key recovery.
- [ ] 3.3 Define offline idempotency, reservation expiry, double-trade
  detection, conflict evidence, and manual resolution.
- [ ] 3.4 Define the financial boundary so Cabinet records external payment
  notation only and excludes escrow, held balance, checkout, and payment
  processing.

## 4. Evidence and Follow-up

- [ ] 4.1 Map threat-model requirements to tests and user-visible recovery for
  fake peer/store, Sybil/eclipse, tampering, forged/deleted feedback, key
  compromise, catalogue poisoning, and privacy leakage.
- [ ] 4.2 Produce the paper-prototype plan for connection drop, counterfeit
  abort, broken signature, stale mirror, feedback mismatch, store attestation,
  unknown peer, and post-event confirmation.
- [ ] 4.3 Add OpenSpec traceability rows for the approved roadmap and
  deterministic test-vector requirements.
- [ ] 4.4 Create dependency-ordered follow-up implementation issues only after
  the decisions are approved.
- [ ] 4.5 Run strict OpenSpec validation and record the final #1957 handoff.
