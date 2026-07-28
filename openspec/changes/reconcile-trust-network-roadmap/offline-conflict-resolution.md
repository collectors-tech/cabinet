# Trust Network Offline Conflict Resolution

This document defines the post-beta offline transaction model for #1957. It
keeps local private state, signed trade objects, and peer transport state
separate so Cabinet can reconcile disconnected trade activity without pretending
that distributed atomic ownership updates exist.

## Canonical Event Chain

- Trade state is represented by append-only signed events: proposal,
  reservation, reservation expiry, receipt, ownership event, cancellation,
  dispute, supersession, and manual resolution.
- Each signed event has a stable event ID derived from canonical serialized
  bytes, object type, schema version, issuer identity, sequence nonce, and
  previous event hash where applicable.
- The local database may cache derived state, but the signed event chain remains
  the portable authority for post-beta trust-network exchange.
- A device may not rewrite or delete another device's signed event. Corrections
  are new signed supersession, dispute, or resolution events.

## Idempotency Rules

- Replaying the same signed event with the same event ID is idempotent and must
  not create duplicate reservations, receipts, ownership events, feedback, or
  audit records.
- A retried send must carry the same event ID and previous-state reference until
  the user intentionally creates a new event.
- Receipt and ownership application must check identity state, key revocation,
  privacy class, specimen/catalogue reference, previous event hash, reservation
  state, and local acceptance before updating derived state.
- Imported queue entries that fail validation remain evidence with a blocked
  state, reason, and recovery action instead of disappearing silently.

## Reservation Expiry

- Reservations are signed participant-shared events with an explicit expiry time,
  specimen or catalogue scope, participants, consideration notation class, and
  previous-state hash.
- Expired reservations cannot produce a valid receipt unless a new signed
  reservation or explicit participant-signed extension exists.
- Expiry is evaluated locally from the signed event timestamp, allowed clock-skew
  policy, and current device time. Disagreements are surfaced as stale or
  clock-skew warnings rather than automatic ownership transfer.
- Reservation expiry releases only Cabinet's local hold/intent state. It does
  not revoke signed evidence that a negotiation occurred.

## Double-Trade Detection

- A double-trade conflict exists when two or more valid signed receipt or
  ownership chains claim incompatible transfer state for the same specimen,
  catalogue serial identity, or mutually exclusive reservation window.
- Detection compares signed event IDs, previous-state hashes, specimen IDs,
  catalogue item references, participant identities, reservation scopes, expiry
  windows, and receipt timestamps.
- Reputation and identity validation may explain risk, but they must not choose
  the winner of an ownership conflict.
- Conflicting chains must remain visible as evidence until both parties sign a
  correction or the user records a manual resolution with rationale.

## Manual Resolution

- Manual resolution is a signed local decision event that records selected
  chain, rejected chain refs, rationale, evidence refs, actor identity, timestamp,
  and privacy class.
- A manual resolution may update the local derived state, but it does not erase
  the conflicting signed objects or force remote peers to accept the same result.
- Resolution UI must show the affected specimen/catalogue item, parties,
  reservation/receipt timestamps, signature status, revocation status, and
  consequences before confirmation.
- Follow-up implementation issues must include deterministic vectors for replay,
  expired reservation, double-trade conflict, broken previous hash, queue retry,
  manual resolution, and post-resolution import of the rejected chain.

## Transport and Queue Boundaries

- Offline queues, QR/manual handoff, libp2p sessions, CRDT replicas, and Git or
  Radicle mirrors carry signed events. They do not approve local state changes.
- A dropped session preserves queued outbound and inbound signed events with
  retry, export, discard, or manual-review actions.
- CRDT convergence may help synchronize drafts or queue metadata, but ownership
  and receipt conflicts stay explicit until signed evidence and local acceptance
  resolve them.
