## Why

#1957 tracks a post-beta architecture cleanup before Cabinet opens implementation
work for signed trade receipts, public ledgers, local P2P exchange, portable
collector trust, catalogue distribution, and community governance.

The current source pack agrees on local-private defaults and signed Cabinet
objects as the authority, but it mixes identity suites, reputation formulas,
transport roles, registry governance, offline conflict handling, and financial
boundaries. That makes it unsafe to open implementation issues from the pack
without first producing one versioned architecture contract.

## What Changes

- Create a canonical post-beta trust-network roadmap capability.
- Publish a decision and traceability matrix that records each source area,
  canonical rule, rejected alternative, unresolved decision, security/privacy
  impact, and delivery phase.
- Separate identity validation, transaction confidence, and reputation into
  independently explainable models with deterministic test vectors.
- Define signed-object authority, serialization, signature, key rotation,
  revocation, migration, and compatibility expectations before code work.
- Exclude escrow, held balances, checkout, stored value, and payment processing
  from Cabinet core; signed receipts may record only external consideration.
- Keep private inventory, locations, notes, contacts, and values local/private
  by default unless a user explicitly publishes selected claims.
- Treat Git, Radicle, libp2p, DHT, CRDT, and Merkle structures as storage,
  transport, or verification components, never implicit authority.
- Produce dependency-ordered follow-up issues only after the architecture
  decisions and OpenSpec requirements are approved.

## Capabilities

### New Capabilities

- `trust-network-roadmap`: post-beta architecture for Cabinet identity,
  receipts, feedback, public mirrors, local P2P, catalogue distribution,
  reputation, governance, and failure recovery.

### Modified Capabilities

- `security`: binds trust-network architecture to privacy defaults, key
  compromise, revocation, tamper evidence, Sybil/eclipse resistance, and
  redacted failure recovery.

## Impact

- Affected code: none in this architecture slice.
- Affected tests: OpenSpec validation and future deterministic architecture
  test-vector checks.
- Affected documentation: OpenSpec specs, decision matrix, traceability, and
  follow-up issue dependency map.
- Related issues: `#1957`; future child issues stay post-beta/sequenced until
  Max explicitly changes release scope.
