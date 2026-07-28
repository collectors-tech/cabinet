# Trust Network Identity Suite

This suite is the canonical #1957 identity baseline for future schema,
prototype, and implementation issues. It is deliberately post-beta and does
not authorize production registry, reputation, or P2P code.

## Canonical Suite

- Identity objects: `cabinet.identity.v1` for users, organizations, custodians,
  and infrastructure nodes. The object records subject type, display claims,
  public signing keys, key purpose, creation time, rotation history, recovery
  references, and revocation status.
- Signing envelope: every trust-network object uses one versioned Cabinet
  envelope with canonical JSON serialization, an object type, schema version,
  key ID, signature algorithm, content hash, issued time, optional expiry, and
  detached signature bytes.
- Baseline algorithm: Ed25519 is the default software-key signing algorithm for
  portable identity and trust objects. Hardware-backed keys may protect the same
  Ed25519 key material or a wrapped local key, but secure enclave availability
  is not required.
- Compatibility bridge: PGP, GPG, SSH, `did:key`, and future wallet or hardware
  identifiers may appear only as signed external anchors. They do not replace
  the Cabinet identity object or signature envelope as authority.
- Key purposes: keys must declare one or more scoped purposes: identity-root,
  object-signing, node-signing, recovery, revocation, catalogue-signing, or
  governance-signing. A key cannot exercise a purpose that is absent from the
  current signed identity state.

## Rotation and Recovery

- Rotation is an append-only signed event linking the old key ID, new key ID,
  effective time, reason, and previous identity state hash.
- Normal rotation requires a valid current identity-root signature or an
  approved recovery path.
- Recovery may use pre-declared recovery keys, recovery contacts, printed
  recovery codes, or an approved governance/custodian flow. Recovery proofs
  must be signed and redacted enough to avoid exposing private contacts.
- A recovered identity must preserve prior public history and mark which keys
  are retired, compromised, or superseded.

## Revocation and Compromise

- Revocation is a signed object that identifies the affected identity, key, or
  node, revocation reason, effective time, replacement or appeal reference, and
  previous state hash.
- Compromised-key revocation must be accepted from a valid recovery key,
  governance/custodian quorum, or a previously published emergency revocation
  key.
- Revoked keys must not validate new trust objects after the revocation
  effective time, but older objects remain historical evidence and must be
  evaluated with timestamp and revocation context.
- Node revocation removes node transport, relay, mirror, or attestation
  authority without revoking the operator's personal identity unless the
  operator identity is explicitly named.

## Compatibility Policy

- Every identity, rotation, recovery, and revocation object must include a
  schema version and migration policy before implementation issues open.
- Readers must fail closed on unknown critical fields and may ignore unknown
  non-critical fields.
- Future key algorithms may be added only through a versioned compatibility
  decision with deterministic test vectors.
- Private inventory, locations, notes, contact details, and values are never
  implied identity claims and must not be published by identity registration.
