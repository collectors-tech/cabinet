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

### Requirement: Canonical trust-network glossary

Cabinet MUST maintain one canonical glossary for the trust-network roadmap so
future implementation issues use consistent terms for identities, items,
signed objects, governance records, and infrastructure roles.

#### Scenario: Required trust-network terms are defined

- **GIVEN** an architecture decision, schema, test vector, or follow-up issue
  references trust-network objects
- **WHEN** the roadmap is validated
- **THEN** the glossary MUST define identity, specimen, catalogue item,
  proposal, reservation, receipt, ownership event, feedback, attestation,
  endorsement, registry, and node
- **AND** each definition MUST state its authority or privacy boundary.

### Requirement: Canonical identity and signature suite

Cabinet MUST define one identity, signature, key-rotation, recovery,
revocation, and compatibility suite before opening trust-network implementation
issues.

#### Scenario: Identity authority does not depend on secure enclaves

- **GIVEN** a user, organization, custodian, or infrastructure node registers a
  trust-network identity
- **WHEN** Cabinet validates that identity or a signed trust object
- **THEN** the canonical suite MUST define the identity object, signing
  envelope, key purpose, rotation event, recovery proof, revocation object, and
  compatibility policy
- **AND** validation MUST NOT require every device to expose a secure enclave.

#### Scenario: External key systems are anchors, not authority

- **GIVEN** a Cabinet identity references PGP, GPG, SSH, `did:key`, wallet,
  hardware, or future external identifiers
- **WHEN** Cabinet evaluates signed trust-network objects
- **THEN** those external identifiers MAY act as signed anchors
- **AND** they MUST NOT replace the Cabinet identity object and signature
  envelope as the canonical authority.

### Requirement: Separate trust scoring models

Cabinet MUST keep identity validation, transaction confidence, and ongoing
reputation as separate explainable models with bounded ranges and deterministic
test vectors.

#### Scenario: Trust models have independent inputs and ranges

- **GIVEN** Cabinet evaluates a trust-network identity, interaction, or history
- **WHEN** a score or explanation is produced
- **THEN** identity validation MUST use the integer range `0..100`
- **AND** transaction confidence MUST use the decimal range `0.00..1.00`
- **AND** ongoing reputation MUST use the decimal range `0.00..1.00`
- **AND** each model MUST name its own inputs and exclusions.

#### Scenario: Strong reputation cannot override invalid evidence

- **GIVEN** a participant has high ongoing reputation or strong identity
  validation
- **WHEN** a transaction has a broken required signature, revoked signing key,
  or unresolved double-trade evidence
- **THEN** transaction confidence MUST be forced or capped by that direct
  evidence
- **AND** the other trust scores MUST NOT silently make the interaction valid.

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
