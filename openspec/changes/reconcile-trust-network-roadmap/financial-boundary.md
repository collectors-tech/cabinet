# Trust Network Financial Boundary

This document defines the post-beta financial boundary for #1957. Cabinet may
record that participants acknowledged external consideration, but it must not
become a wallet, escrow service, payment processor, checkout system, or stored
value product.

## Allowed External Consideration Notation

- A proposal, reservation, receipt, dispute, or manual resolution may include an
  external consideration notation chosen by the participants.
- Allowed notation fields are limited to type label, amount, currency, external
  payment method label, external reference supplied by the user, acknowledgement
  timestamp, participant acknowledgements, and optional redacted evidence refs.
- Consideration notation is evidence that parties claim an external exchange was
  agreed, attempted, completed, disputed, refunded, or cancelled. It is not proof
  that funds moved unless the parties or a separate signed attestation say so.
- Public or participant-shared views must let users redact amount, currency,
  external references, and payment method labels when they are not necessary for
  the selected visibility class.

## Forbidden Cabinet Financial Capabilities

Cabinet core must not implement or imply any of these capabilities:

- escrow, staged release, held balances, stored value, wallet balances, account
  top-ups, or internal credits;
- card, bank, PayID, crypto, marketplace, or payment-service-provider checkout;
- payment authorization, capture, refund, chargeback, payout, settlement,
  reconciliation, or ledger-of-funds workflows;
- custody, recovery, freezing, seizure, or transfer of user funds;
- automatic ownership transfer based on a payment webhook, external payment
  status, or balance event;
- storage of payment credentials, full account numbers, full card data, seed
  phrases, private wallet keys, or payment-provider secret tokens in trust
  objects.

## Receipt and Ownership Boundary

- Signed receipts record participant acknowledgements and external consideration
  notation only.
- Ownership events may reference receipt evidence, but they must be signed
  Cabinet objects accepted through the local trust workflow, not consequences of
  a payment system event.
- A payment reference may support transaction confidence only as scoped evidence.
  It cannot override broken signatures, revoked keys, privacy violations,
  expired reservations, double-trade conflicts, or missing local acceptance.
- Disputes may reference external payment evidence in redacted form, but Cabinet
  does not arbitrate funds or instruct a processor to move money.

## UI and Language Rules

- UI copy must use terms such as "external payment noted",
  "consideration acknowledged", "payment reference", or "outside Cabinet".
- UI copy must not use terms that imply custody or processing, including
  "Cabinet balance", "top up", "pay with Cabinet", "checkout", "release funds",
  "hold funds", "escrow protected", or "refund through Cabinet".
- Any trust-network flow that records consideration must show that payment is
  external, optional for barter/gift/loan flows where applicable, and separate
  from Cabinet's signed ownership evidence.
- Follow-up implementation issues must include tests proving forbidden terms and
  payment states are absent from schemas, API payloads, UI copy, logs, and
  public/participant-shared serialized objects.

## Integration Boundary

- Cabinet may deep-link to, import a user-supplied reference from, or store a
  non-secret label for an external payment provider only in a separately scoped
  integration issue.
- External payment integrations must remain read-only/reference-only for
  trust-network receipts unless Max explicitly approves a new non-core scope.
- Provider webhooks, screenshots, emails, or exported payment records can be
  attached as redacted evidence refs, but they cannot mutate ownership state
  without signed participant acceptance.
- Any future commerce integration must be tracked outside this roadmap's core
  trust-network work and must preserve the no-custody/no-checkout boundary by
  default.
