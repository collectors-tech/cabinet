# Trust Network Privacy and Publication Boundaries

This document defines which trust-network data may remain local, be shared with
participants, become public claims, or be published only as hashes. The default
rule is local privacy: data leaves a Cabinet workspace only through an explicit
user action that selects fields, visibility, expiry, and recipients.

## Visibility Classes

| Class | Boundary | Allowed contents | Forbidden contents | Required controls |
| --- | --- | --- | --- | --- |
| Private local record | Stored only in the user's local Cabinet database and backups. | Inventory rows, collection locations, private notes, condition details, purchase costs, insurance/value notes, attachments, contact details, draft proposals, unresolved disputes, and local scoring explanations. | Publication, peer sync, mirror indexing, catalogue inclusion, or implicit use as public identity claims. | Local profile access controls, export preview, field-level redaction, and no network publication by default. |
| Participant-shared record | Shared only with named participants in a proposal, reservation, receipt, dispute, recovery, or moderation flow. | Selected item/specimen claims, proposal terms, reservation expiry, receipt acknowledgements, external consideration notation, feedback drafts, dispute evidence, and contact method chosen for that flow. | Full inventory, unrelated collection values, home/storage locations, unrelated contacts, private notes, unrelated attachments, or ambient database access. | Recipient list, purpose, expiry where applicable, signature preview, revocation/dispute path, and local acceptance before state mutation. |
| Public claim | Published to a registry, mirror, catalogue release, public profile, or community governance record. | Explicit identity display claims, public key material, registry membership state, scoped attestations, endorsements, selected receipt/feedback summaries, catalogue item references, mirror manifests, and revocation records. | Private inventory fields, precise locations, private contact details, collection values, hidden attachments, full payment details, and unredacted recovery contacts. | Public preview, durable warning, selected-field manifest, hash list, issuer signature, schema version, and withdrawal/revocation guidance. |
| Hash-only publication | Publishes a content hash, Merkle proof, commitment, or manifest entry without revealing the underlying private content. | Object hash, field commitment, bundle hash tree entry, timestamp, schema version, issuer or mirror ref when intentionally public, and optional expiry. | Plaintext private content, reversible identifiers, unsalted sensitive field hashes, or enough metadata to reconstruct private inventory. | Salt/nonce policy for sensitive commitments, documented verification purpose, collision/rollback handling, and no claim that a hash proves truth by itself. |

## Field-Level Rules

- Inventory titles, catalogue references, specimen identifiers, condition
  summaries, photos, and provenance may be shared only when selected for a
  participant flow or public claim.
- Exact location, storage notes, private owner notes, acquisition cost,
  insurance value, collection value, unrelated attachments, and private contact
  details remain private unless a future approved requirement names the exact
  exception.
- External payment notation on receipts is limited to acknowledged
  consideration labels, amount/currency where the user chooses to disclose it,
  external reference text, and participant acknowledgement. Cabinet must not
  store payment credentials, escrow state, held balances, checkout state, or
  payment-processing events.
- Recovery contacts and custodian evidence must be redacted by default. Public
  recovery or appeal records may include proof references, not private contact
  details.
- Reputation summaries may publish aggregate scores or explanation bands only
  after the referenced signed evidence validates and any redaction, appeal, or
  revocation state is applied.

## Publication Workflow

1. Select object and fields.
2. Choose visibility class, recipients or public destination, expiry, and
   revocation/withdrawal path.
3. Render a deterministic preview of included and excluded fields.
4. Sign the selected claim or hash manifest.
5. Publish through the chosen carrier only after signature verification passes.
6. Record local evidence of publication, carrier, object hash, and withdrawal
   instructions.

## Failure and Recovery

- If a mirror, peer, or catalogue exposes private-only fields, Cabinet must
  treat it as privacy leakage evidence, stop trusting that publication surface,
  and show withdrawal, revocation, and rotation guidance.
- If a participant shares a participant-only object publicly, Cabinet must keep
  the original signed object as evidence but must not make that publication the
  canonical visibility rule.
- If a hash-only claim becomes linkable to sensitive content, future
  implementations must support replacement commitments and user-visible
  leakage warnings.
- Revocation or withdrawal does not erase already copied public data; the UI
  must explain that it changes Cabinet trust state and mirror guidance rather
  than guaranteeing deletion from external hosts.
