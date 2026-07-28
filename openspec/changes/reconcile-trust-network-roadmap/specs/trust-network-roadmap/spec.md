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

### Requirement: Signed trust objects define authority

Cabinet MUST define versioned signed-object authority for receipts, feedback,
attestations, endorsements, revocations, mirrors, and catalogue manifests before
opening trust-network implementation issues.

#### Scenario: Signed-object authority is verified before use

- **GIVEN** Cabinet receives or loads a receipt, feedback object, attestation,
  endorsement, revocation, mirror manifest, or catalogue manifest
- **WHEN** the object is used as trust-network evidence
- **THEN** Cabinet MUST verify the signing envelope, object schema, issuer key
  purpose, referenced identity state, revocation state, and previous-state hash
- **AND** unknown critical fields, unsupported schema versions, broken
  signatures, revoked issuer keys, or unresolved authority conflicts MUST fail
  closed.

#### Scenario: Carriers never replace signed authority

- **GIVEN** a signed trust object is carried through QR, manual export, local
  queue, peer session, Git mirror, Radicle mirror, DHT advert, catalogue
  bundle, or search/index projection
- **WHEN** Cabinet evaluates that object
- **THEN** the signed Cabinet object MUST remain the authority
- **AND** the carrier MUST NOT mutate private local state, reputation,
  registry membership, or catalogue trust without a valid signed object and the
  required local acceptance or conflict checks.

### Requirement: Privacy and publication boundaries are explicit

Cabinet MUST separate private local records, participant-shared records, public
claims, and hash-only publication before trust-network implementation issues
open.

#### Scenario: Private data remains local unless selected

- **GIVEN** Cabinet stores inventory rows, locations, private notes, condition
  details, purchase costs, values, attachments, contact details, draft
  proposals, disputes, or local trust explanations
- **WHEN** a trust-network flow prepares data for a peer, mirror, catalogue,
  registry, or public profile
- **THEN** Cabinet MUST default those fields to private local records
- **AND** data MUST leave the workspace only after an explicit field-level
  selection, visibility class, recipient or destination, expiry where
  applicable, and signed preview.

#### Scenario: Publication classes restrict sensitive fields

- **GIVEN** a user publishes or shares participant-scoped, public, or hash-only
  trust-network evidence
- **WHEN** Cabinet serializes the selected object
- **THEN** participant-shared records MUST include only data needed for the
  named proposal, reservation, receipt, dispute, recovery, or moderation flow
- **AND** public claims MUST exclude private inventory fields, precise
  locations, private contact details, collection values, hidden attachments,
  full payment details, and unredacted recovery contacts
- **AND** hash-only publication MUST NOT include plaintext private content,
  reversible sensitive identifiers, or enough metadata to reconstruct private
  inventory.

### Requirement: Transport and verification components are non-authoritative

Cabinet MUST document Git, Radicle, libp2p, DHT, CRDT, and Merkle roles as
storage, transport, discovery, reconciliation, or verification components
instead of implicit trust-network authority.

#### Scenario: Component data is validated before trust use

- **GIVEN** Cabinet reads trust-network data from Git, Radicle, libp2p, DHT,
  CRDT replication, or Merkle proofs
- **WHEN** Cabinet evaluates that data for identity, ownership, reputation,
  registry, catalogue, or conflict state
- **THEN** Cabinet MUST validate the referenced signed object, manifest, or hash
  commitment against schema, signature, key purpose, revocation, privacy class,
  freshness, and local acceptance rules
- **AND** component availability, hosting history, peer presence, graph
  membership, CRDT convergence, or Merkle inclusion MUST NOT become authority by
  itself.

#### Scenario: Component failures degrade safely

- **GIVEN** a mirror is stale, a DHT advert is malicious, a peer session drops,
  a CRDT ownership conflict appears, or a Merkle proof mismatches
- **WHEN** Cabinet surfaces the result to a user or follow-up implementation
  test
- **THEN** Cabinet MUST preserve the signed evidence and failing proof context
- **AND** it MUST show degraded discovery, stale mirror, retry/fallback,
  conflict-resolution, or bundle-rejection behaviour instead of silently
  accepting the component state.

### Requirement: Public registry governance is versioned

Cabinet MUST define public-registry bootstrap, governance quorum, revocation,
retirement, appeal, and compromised-key recovery before registry, node, or
community-governance implementation issues open.

#### Scenario: Bootstrap authority cannot hide as community governance

- **GIVEN** Cabinet consumes a public registry root, custodian list, mirror
  list, catalogue release approval, or governance charter
- **WHEN** the record comes from a bootstrap authority
- **THEN** the signed registry record MUST declare bootstrap state, issuer,
  key purpose, expiry, migration target, appeal path, previous root hash, and
  exact authority scope
- **AND** readers MUST surface bootstrap or transition state instead of
  presenting temporary central control as community-owned governance.

#### Scenario: Governance actions require quorum evidence

- **GIVEN** a governance action revokes, retires, restores, migrates, approves,
  or appeals an identity, key, node, mirror, registry record, attestation,
  endorsement, or catalogue release
- **WHEN** Cabinet evaluates the action
- **THEN** the action MUST reference a signed proposal, voting window, voter
  eligibility snapshot, threshold, signer set, evidence refs, and previous
  registry root hash
- **AND** registry-wide decisions MUST require at least three eligible
  governance signers and at least two-thirds approval unless a versioned
  emergency quorum applies only to time-limited containment.

#### Scenario: Revocation and recovery preserve history

- **GIVEN** an identity, key, node, mirror, catalogue release, or registry
  record is revoked, retired, appealed, or recovered
- **WHEN** Cabinet calculates current trust-network state
- **THEN** Cabinet MUST evaluate the ordered chain of signed registry roots,
  revocations, retirements, appeals, and recovery records
- **AND** successful appeal or recovery MAY supersede operational effect but
  MUST NOT erase the original signed evidence or make post-revocation objects
  signed by compromised keys valid.

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
