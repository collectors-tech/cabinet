## ADDED Requirements

### Requirement: Post-beta trust-network roadmap

Cabinet MUST keep trust-network implementation work sequenced behind an
approved architecture roadmap that defines authority, privacy, identity,
reputation, transport, governance, and failure recovery before code work begins.

#### Scenario: Source conflicts are captured before implementation

- **GIVEN** Cabinet source material describes identity, feedback, receipts,
  public mirrors, local P2P, catalogue distribution, governance, security, and
  failure prototypes
- **WHEN** the roadmap is updated
- **THEN** every source area MUST appear in a committed decision matrix
- **AND** every row MUST record a canonical rule, rejected alternative,
  unresolved decision, security/privacy impact, and delivery phase.

#### Scenario: Signed objects remain authoritative

- **GIVEN** Cabinet publishes or exchanges trust-network data through Git,
  Radicle, libp2p, DHT, CRDT replication, Merkle bundles, QR, or manual handoff
- **WHEN** another device, mirror, or peer consumes that data
- **THEN** versioned signed Cabinet objects MUST be the authority
- **AND** transports, mirrors, queues, and catalogue bundles MUST NOT become
  implicit authority over private local records.

#### Scenario: Payment custody stays out of Cabinet core

- **GIVEN** a trade receipt records value exchanged outside Cabinet
- **WHEN** the receipt is serialized, signed, published, or reconciled
- **THEN** it MAY record external consideration notation
- **AND** it MUST NOT implement escrow, held balances, checkout, stored value,
  payment accounts, or payment processing.

#### Scenario: Trust scores are independently explainable

- **GIVEN** Cabinet calculates identity validation, transaction confidence, or
  ongoing reputation
- **WHEN** users or tests inspect the calculation
- **THEN** each model MUST have an independent purpose, bounded input ranges,
  deterministic test vectors, and user-visible explanation
- **AND** one score MUST NOT silently inherit authority from another.

#### Scenario: Private collection data stays private by default

- **GIVEN** a user has inventory, locations, notes, contact details, collection
  values, attachments, or private metadata
- **WHEN** trust-network features publish, mirror, or exchange data
- **THEN** private local records MUST remain private by default
- **AND** only explicitly selected claims or hash-only proofs MAY leave the
  local workspace.

### Requirement: Trust-network follow-up issue sequencing

Cabinet MUST create only small, dependency-ordered, post-beta implementation
issues from the approved roadmap.

#### Scenario: Follow-up issues preserve sequencing

- **GIVEN** a decision row has enough approved detail for implementation
- **WHEN** a follow-up issue is opened
- **THEN** the issue MUST reference the decision row, delivery phase,
  security/privacy impact, and OpenSpec requirement
- **AND** it MUST remain post-beta/sequenced unless Max explicitly changes the
  release scope.
