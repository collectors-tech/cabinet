# Trust Network Follow-up Issue Map

This map is the approval-gated child issue plan for #1957. It records the
small implementation slices that may be opened after the
`reconcile-trust-network-roadmap` OpenSpec change is approved. Until then, no
child issue below authorizes production P2P, public registry, reputation,
catalogue distribution, escrow, checkout, or payment processing work.

## Creation Gate

- Parent: #1957.
- Gate: open the child issues only after the #1957 roadmap change is approved.
- Default scope label: `status/sequenced`.
- Default release scope: post-beta.
- Every child issue body must reference its decision row, delivery phase,
  security/privacy impact, OpenSpec requirement, and relevant test vectors.
- Any issue touching payment-like evidence must state `notation-only`,
  `evidence-import`, or a separately approved non-core commerce integration,
  and must prove custody, checkout, held balance, payment credentials, and
  payment-processing states are absent by default.

## Dependency Order

| Order | Issue title | Depends on | Decision row | Delivery phase | Required vectors | OpenSpec requirement |
| --- | --- | --- | --- | --- | --- | --- |
| 1 | `architecture(trust-network): codify canonical signed-envelope schemas` | #1957 approval | User and infrastructure-node identity/registration; Signed trade receipts and QR/manual handoff; Signed feedback vaults | 2. paper prototype/schema | `TV-TRUST-003-broken-signature`, `TV-THREAT-004-tampered-object`, `TV-THREAT-006-key-compromise` | Canonical identity and signature suite; Signed trust objects define authority |
| 2 | `architecture(trust-network): add identity rotation and recovery test vectors` | signed-envelope schemas | User and infrastructure-node identity/registration; Community governance, custodians, revocation, and appeal | 2. paper prototype/schema | `TV-TRUST-005-compromised-key`, `TV-THREAT-006-key-compromise` | Canonical identity and signature suite; Public registry governance is versioned |
| 3 | `prototype(trust-network): build receipt and feedback paper flows` | signed-envelope schemas | Signed trade receipts and QR/manual handoff; Signed feedback vaults; Financial boundary | 2. paper prototype/schema; 3. signed receipts/feedback | `TV-PAPER-002-counterfeit-abort`, `TV-PAPER-003-broken-signature`, `TV-PAPER-005-feedback-mismatch`, `TV-THREAT-005-feedback-forgery-delete` | Paper prototypes cover trust failure flows; Financial capability stays outside Cabinet core |
| 4 | `prototype(trust-network): build mirror and catalogue failure flows` | signed-envelope schemas | Public Git/Radicle ledgers and mirror comparison; Catalogue bundle distribution and Merkle verification | 2. paper prototype/schema; 4. Git/Radicle publication; 6. catalogue distribution | `TV-PAPER-004-stale-mirror`, `TV-THREAT-003-sybil-eclipse`, `TV-THREAT-007-catalogue-poisoning` | Transport and verification components are non-authoritative; Paper prototypes cover trust failure flows |
| 5 | `prototype(trust-network): build local peer and offline conflict flows` | signed-envelope schemas; identity rotation vectors | Direct local P2P exchange; Offline transaction queues and conflicts | 2. paper prototype/schema; 5. local P2P sessions | `TV-PAPER-001-connection-drop`, `TV-PAPER-007-unknown-peer`, `TV-THREAT-001-fake-peer`, `TV-TRUST-004-double-trade` | Offline exchange conflicts are explicit and idempotent; Transport and verification components are non-authoritative |
| 6 | `architecture(trust-network): specify public registry governance records` | identity rotation vectors | Community governance, custodians, revocation, and appeal; User and infrastructure-node identity/registration | 6. governance/network nodes | `TV-THREAT-002-fake-store`, `TV-THREAT-006-key-compromise`, `TV-PAPER-006-store-attestation` | Public registry governance is versioned |
| 7 | `architecture(trust-network): specify bounded reputation calculations` | receipt and feedback paper flows; registry governance records | Reputation, validation scoring, and web-of-trust formulas; Signed feedback vaults | 3. signed receipts/feedback; 6. governance/network nodes | `TV-TRUST-001-new-valid-user`, `TV-TRUST-002-store-backed-receipt`, `TV-TRUST-004-double-trade`, `TV-THREAT-003-sybil-eclipse` | Separate trust scoring models |
| 8 | `architecture(trust-network): specify public publication and redaction tests` | mirror/catalogue failure flows; registry governance records | Privacy and publication boundaries; Public Git/Radicle ledgers and mirror comparison; Security threats | 4. Git/Radicle publication; 6. governance/network nodes | `TV-THREAT-008-privacy-leakage`, `TV-PAPER-008-post-event-confirmation` | Privacy and publication boundaries are explicit; Security requirements SHALL bind decisions to threat evidence |

## Child Issue Body Requirements

Each issue opened from this map must include:

- Parent/gate: "Post-beta child of #1957; blocked until the approved
  `reconcile-trust-network-roadmap` OpenSpec change is accepted."
- Decision references: exact decision matrix row names and delivery phase.
- Privacy/security impact copied or narrowed from the decision row.
- OpenSpec requirement references and at least one deterministic test vector.
- Acceptance criteria proving signed Cabinet objects, not transports or mirrors,
  are authoritative wherever trust evidence crosses a boundary.
- A negative acceptance criterion for privacy leakage when data leaves the
  local workspace.
- A negative acceptance criterion for payment custody language or state when
  receipt, value, or consideration fields are touched.

## Non-Issues

Do not open child implementation issues for these before a separate approval:

- Production P2P network service.
- Production public identity registry.
- Production reputation score rollout.
- Catalogue torrent or node infrastructure.
- Escrow, held balance, wallet, checkout, refund, settlement, or payment
  processor integration inside Cabinet core.
