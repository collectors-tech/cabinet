# Trust Network Scoring Models

This document separates the three trust-network calculations required by
#1957. The numbers are architecture baselines for deterministic tests and
paper prototypes, not production tuning.

## Identity Validation

Purpose: explain how much evidence supports an identity's control and claims.

- Range: integer `0..100`.
- Inputs: signed identity object, non-revoked current key, recovery path,
  optional external anchors, optional custodian attestations, and optional
  public registry state.
- Exclusions: trade outcome, feedback rating, reputation propagation, and item
  value never increase identity validation.
- Formula baseline:
  - `40` for a valid `cabinet.identity.v1` object with a current key.
  - `20` for at least one tested recovery path.
  - `15` for a non-revoked external anchor.
  - `15` for a scoped custodian or store attestation.
  - `10` for current non-retired registry inclusion.
  - Cap at `100`; subtract `60` for compromised current key until recovery.

## Transaction Confidence

Purpose: explain confidence in one proposed or completed interaction.

- Range: decimal `0.00..1.00`.
- Inputs: participant identity validation, object signature validity,
  reservation status, receipt countersignatures, conflict evidence, expiry, and
  transport freshness.
- Exclusions: long-term reputation cannot make an invalid signature valid, and
  identity validation cannot hide a conflict.
- Formula baseline:
  - `0.30` if all required object signatures validate.
  - `0.20` if both participant identities have validation score `>= 60`.
  - `0.20` if reservation or proposal state is current and unexpired.
  - `0.20` if receipt or handoff is countersigned by all required parties.
  - `0.10` if no stale mirror, double-trade, or conflict evidence exists.
  - Clamp to `0.00..1.00`; force `0.00` for broken required signature.
  - Cap at `0.40` when unresolved double-trade evidence exists.

## Ongoing Reputation

Purpose: summarize durable interaction history without replacing direct
validation or per-transaction confidence.

- Range: decimal `0.00..1.00`.
- Inputs: signed feedback outcomes, revocations, appeal state, endorsement
  edges, recency, and anti-Sybil attenuation.
- Exclusions: identity validation and transaction confidence are display
  context only; they are not reputation points.
- Formula baseline:
  - Start at `0.50` for a validated identity with no usable history.
  - Add up to `0.30` from signed positive feedback, weighted by transaction
    confidence and recency.
  - Subtract up to `0.40` from signed negative feedback, disputes, or revoked
    feedback fraud.
  - Add up to `0.10` from endorsed graph paths after attenuation.
  - Clamp to `0.00..1.00`; each graph hop applies at least `0.50`
    attenuation and any path through a revoked key contributes `0.00`.

## Deterministic Test Vectors

| Vector ID | Identity validation | Transaction confidence | Reputation | Expected result |
| --- | --- | --- | --- | --- |
| TV-TRUST-001-new-valid-user | Valid identity, current key, recovery path, no anchors, no attestations, no registry: `60`. | Valid proposal signatures, both participants `>= 60`, current reservation, no receipt yet, no conflicts: `0.60`. | No signed history: `0.50`. | New user can propose but not claim high reputation. |
| TV-TRUST-002-store-backed-receipt | Valid identity plus recovery, external anchor, store attestation, registry: `100`. | Valid signatures, both identities `>= 60`, current reservation, countersigned receipt, no conflicts: `1.00`. | Positive signed feedback with high-confidence receipt: `0.74`. | Strong identity and clean receipt raise transaction confidence; reputation remains separate. |
| TV-TRUST-003-broken-signature | Identity evidence otherwise scores `85`. | Broken required receipt signature: forced `0.00`. | Prior reputation `0.80`. | Reputation and identity do not rescue a broken signature. |
| TV-TRUST-004-double-trade | Identity evidence scores `75`. | Valid signatures and receipt but unresolved double-trade evidence: capped `0.40`. | Mixed signed feedback after dispute: `0.42`. | Conflict evidence caps the transaction while reputation records history. |
| TV-TRUST-005-compromised-key | Compromised current key before recovery: base `85 - 60 = 25`. | Any new object signed after revocation effective time: `0.00`. | Graph paths through revoked key: `0.00` contribution. | Compromise blocks new authority until recovery, while old evidence needs timestamp context. |
