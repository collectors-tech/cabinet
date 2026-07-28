# Trust Network Transport, Storage, and Verification Roles

This document assigns post-beta trust-network roles for Git, Radicle, libp2p,
DHT, CRDT replication, and Merkle structures. These components are implementation
options for moving, storing, discovering, reconciling, or verifying signed
Cabinet objects. They are never authority by themselves.

## Component Role Matrix

| Component | Approved role | Not authority for | Required guardrails | Delivery phase |
| --- | --- | --- | --- | --- |
| Git repositories | Append-friendly public mirror storage for selected signed objects, mirror manifests, hash-only publications, catalogue release metadata, and traceable history. | Identity validity, private inventory state, ownership transfer, reputation totals, registry membership, or catalogue truth. | Signed mirror manifests, provider-replaceable layout, stale-ref detection, withdrawal/revocation guidance, and no private fields by default. | 4. Git/Radicle publication |
| Radicle | Peer-owned Git-style mirror and discovery option for the same selected signed objects that may appear in hosted Git mirrors. | Radicle identity, project membership, or peer graph as Cabinet trust authority. | Same signed-object validation as Git, mirror equivalence checks, stale mirror warnings, and fallback to hosted/self-hosted Git. | 4. Git/Radicle publication |
| libp2p | Session transport for explicitly selected signed objects, peer capability negotiation, and local exchange prototypes. | Peer identity, trust score, local database access, conflict resolution, or acceptance of ownership changes. | Cabinet identity handshake, scoped capability list, per-object consent, replay protection, size/rate limits, and QR/manual fallback. | 5. local P2P sessions |
| DHT | Optional discovery and announcement layer for mirror manifests, peer rendezvous, and catalogue availability hints. | Object content, registry inclusion, reputation, catalogue release validity, or peer trust. | Store hashes/locator hints only, verify signed manifests after retrieval, expire adverts, resist eclipse/Sybil assumptions, and avoid private metadata leakage. | 5. local P2P sessions; 6. network nodes |
| CRDT replication | Local reconciliation structure for drafts, queues, or collaborative metadata where commutative state helps user experience. | Signed object validity, final trade state, private record publication, or conflict-free ownership transfer. | CRDT entries must reference signed events, conflicts remain visible, local acceptance is required, and distributed atomic updates are not assumed. | 5. local P2P sessions |
| Merkle structures | Verification structure for catalogue bundles, object sets, hash-only commitments, mirror snapshots, and tamper evidence. | Truth of claims, identity validation, reputation, ownership, or publication permission. | Signed root manifest, deterministic canonical hashing, rollback/freshness checks, inclusion and non-inclusion proof policy, and sensitive-field salt/nonce rules. | 4. Git/Radicle publication; 6. catalogue distribution |

## Cross-Component Rules

- Every component must consume or emit versioned signed Cabinet objects, signed
  manifests, or hash commitments from the approved schemas.
- A component may be unavailable, stale, replaced, forked, or malicious without
  changing Cabinet's authority model. The result is degraded discovery or
  warning state, not silent trust transfer.
- Readers must validate signatures, key purposes, revocation state, schema
  compatibility, privacy class, freshness, and local acceptance before using
  component data.
- Mirrors and discovery systems may prove that data was published or advertised,
  but they do not prove the underlying claim is true.
- Peer transports may move selected objects, but they cannot grant ambient
  access to the local database, filesystem, tools, private notes, contacts, or
  collection values.

## Failure Modes

- Stale Git/Radicle mirror: show mirror freshness warning, continue evaluating
  any locally available signed objects with timestamp context, and require a
  newer manifest before claiming current public state.
- Malicious DHT advert: discard locator-only data unless the retrieved signed
  manifest validates; never display DHT presence as trust.
- CRDT ownership conflict: keep both signed claims as conflict evidence and
  require manual resolution instead of auto-selecting a winner.
- Merkle rollback or mismatch: reject the bundle or snapshot, preserve the
  failing root/proof as evidence, and point users to another mirror or release.
- libp2p session drop: preserve the local outbound/inbound signed-object queue
  and require retry, QR/manual fallback, or explicit cancellation.
