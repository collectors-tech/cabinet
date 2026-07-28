# Trust Network Registry and Governance

This document defines the post-beta public-registry and governance baseline for
#1957. It resolves the central-whitelist versus community-owned-governance
conflict by making bootstrap authority explicit, temporary, signed, and
migratable before any registry implementation work opens.

## Registry Scope

- The public registry is a versioned signed set of identity, node, custodian,
  governance, revocation, retirement, appeal, mirror, and catalogue release
  records.
- Registry inclusion helps discovery and validation scoring, but it does not
  replace signed identity objects, signed trust objects, local private records,
  or user acceptance.
- Registry records may be mirrored through Git, Radicle, self-hosted Git, or
  other approved carriers only as signed registry snapshots or manifests.
- Private inventory, locations, notes, contact details, collection values,
  payment credentials, and unredacted recovery contacts are never registry
  fields.

## Bootstrap and Migration

- Bootstrap phase: an explicitly named Cabinet bootstrap authority may sign the
  first registry root, custodian list, mirror list, and governance charter.
- Bootstrap records must declare `bootstrap=true`, issuer identity, key purpose,
  expiry, migration target, appeal contact path, and the exact authority scope.
- Bootstrap authority expires unless renewed by the approved governance quorum
  or migrated to community governance.
- Migration to community governance requires a signed transition proposal,
  public review window, quorum approval, final transition record, and previous
  registry root hash.
- Readers must show bootstrap/transition state so a temporary central registry
  cannot silently present itself as community-owned.

## Governance Quorum

- Governance actions use signed proposals with object type, scope, rationale,
  evidence refs, voting window, threshold, voter eligibility snapshot, and
  previous registry root hash.
- Baseline quorum for registry-wide decisions is at least three eligible
  governance signers and at least two-thirds approval among participating
  eligible signers.
- Emergency actions may use a smaller pre-declared emergency quorum only for
  compromised-key, poisoned-catalogue, malicious-node, or privacy-leakage
  containment, and must expire into normal review.
- Custodian or store attestations are scoped evidence. They are not governance
  votes unless the custodian identity is also eligible in the voter snapshot.
- Quorum rules must be versioned and immutable for the proposal window; changing
  quorum requires its own governance proposal.

## Revocation, Retirement, and Appeal

- Revocation is append-only signed evidence naming the affected identity, key,
  node, registry record, attestation, endorsement, mirror, or catalogue release.
- Retirement is a non-punitive signed record for identities, nodes, mirrors, or
  catalogue releases that are intentionally ending service or support.
- Appeal is a signed proposal that references the revocation or retirement,
  presents evidence, redacts private recovery contacts, and defines requested
  outcome.
- Successful appeal may supersede the operational effect of a revocation but
  must not erase the original signed record or evidence trail.
- Registry readers must evaluate current state from the ordered chain of signed
  registry roots, revocations, retirements, appeals, and recovery records.

## Compromised-Key Recovery

- Compromised-key recovery may be accepted through a pre-declared recovery key,
  emergency revocation key, governance quorum, or scoped custodian recovery
  flow from the identity suite.
- Recovery records must name compromised keys, effective time, replacement keys,
  affected object scopes, evidence refs, previous identity hash, previous
  registry root hash when public, and redacted recovery proof.
- New objects signed by compromised keys after the revocation effective time
  must fail closed.
- Historical objects signed before the effective time remain evidence and must
  be evaluated with timestamp, revocation, appeal, and conflict context.
- Node key compromise revokes node transport, relay, mirror, or attestation
  authority without automatically revoking the operator's personal identity
  unless the operator identity is explicitly included.

## User-Visible Requirements

- Registry state must show whether an identity, key, node, mirror, or catalogue
  release is current, bootstrapped, community-governed, retired, revoked,
  under appeal, or recovered.
- Users must see why trust evidence is blocked or degraded: expired bootstrap,
  missing quorum, active appeal, compromised key, retired node, stale registry
  root, or conflicting governance records.
- Follow-up implementation issues must include deterministic test vectors for
  bootstrap expiry, quorum approval/rejection, emergency revocation expiry,
  successful appeal, failed appeal, and compromised-key recovery.
