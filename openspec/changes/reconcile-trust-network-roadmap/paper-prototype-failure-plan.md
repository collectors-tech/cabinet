# Trust Network Paper-Prototype Failure Plan

This document defines the #1957 paper-prototype plan for trust-network failure
flows. These prototypes must happen before production P2P, registry, catalogue,
reputation, or signed-receipt implementation issues are opened.

## Prototype Method

- Use paper cards or clickable mockups for participant identities, specimens,
  proposals, reservations, receipts, feedback, mirror manifests, catalogue
  manifests, and recovery actions.
- Each prototype starts from a normal happy-path trade or verification flow, then
  injects one failure card at the named decision point.
- Facilitators record the user's expected next action, confusing copy, missing
  evidence, and whether the user believes Cabinet has already changed ownership,
  payment, reputation, registry, or publication state.
- A prototype passes only when the user can explain what is blocked or degraded,
  what evidence Cabinet preserved, and which recovery actions are available.

## Failure Flow Matrix

| Flow | Failure injection | Evidence shown | User choices | Pass condition | Follow-up vector |
| --- | --- | --- | --- | --- | --- |
| Connection drop | Peer session drops after one side sends a signed reservation or receipt but before acknowledgement. | Outbound queue entry, last peer identity state, signed event ID, previous hash, retry count, and QR/manual export fallback. | Retry, export QR/manual bundle, cancel local draft, or keep pending. | User understands no remote acceptance is assumed and local derived state remains pending. | `TV-PAPER-001-connection-drop` |
| Counterfeit abort | A specimen verification step reveals counterfeit evidence before receipt signing. | Item/specimen refs, failed attestation, photos/evidence refs, unsigned receipt draft, and cancellation reason. | Abort, request more evidence, create dispute note, or continue only as a new explicitly marked proposal. | User does not accidentally sign a clean receipt and can see why confidence is blocked. | `TV-PAPER-002-counterfeit-abort` |
| Broken signature | Receipt, feedback, or ownership event has a missing, malformed, revoked, or wrong-purpose signature. | Object type, signer identity, key purpose, revocation state, signature status, and failing canonical hash. | Reject, ask participant to re-sign, preserve as dispute evidence, or import for manual review only. | User understands reputation/identity score cannot rescue the broken signature. | `TV-PAPER-003-broken-signature` |
| Stale mirror | Git/Radicle mirror is behind the latest signed registry or catalogue root. | Mirror provider, last signed root, freshness timestamp, alternate mirror list, and stale-read warning. | Try alternate mirror, continue with local cached evidence, defer claim, or manually import newer root. | User sees discovery is degraded and does not treat the stale mirror as current authority. | `TV-PAPER-004-stale-mirror` |
| Feedback mismatch | Local feedback cache, public mirror, and participant-supplied feedback object disagree. | Feedback object signatures, receipt refs, withdrawal/moderation records, mirror source, cache timestamp, and current reputation effect. | Refresh source, preserve conflict, view history, request signed correction, or ignore unverified feedback. | User can tell whether feedback is current, withdrawn, forged, or historically retained. | `TV-PAPER-005-feedback-mismatch` |
| Store attestation | Store/custodian attestation is missing, expired, forged, or outside scope. | Store identity, registry state, key purpose, attestation scope, expiry, specimen/catalogue refs, and confidence delta. | Continue without store boost, request a new attestation, attach manual evidence, or block high-confidence claim. | User understands the attestation is scoped evidence, not participant or governance authority. | `TV-PAPER-006-store-attestation` |
| Unknown peer | Peer presents valid transport details but no verified Cabinet identity chain. | Peer transport fingerprint, missing identity fields, capability request, proposed object list, and privacy exposure preview. | Reject, continue with participant-limited manual review, request verified identity, or switch to QR/manual exchange. | User does not share private fields or accept ownership changes from transport identity alone. | `TV-PAPER-007-unknown-peer` |
| Post-event confirmation | After receipt signing, a later event introduces revocation, double-trade evidence, privacy leak, or payment-reference dispute. | Original receipt, new conflicting event, affected fields, current local state, reputation/confidence effect, and required confirmation. | Confirm local resolution, request countersignature, revoke publication, open dispute, or leave unresolved. | User sees prior evidence is preserved and any new local state change requires explicit confirmation. | `TV-PAPER-008-post-event-confirmation` |

## Required Prototype Artifacts

- A one-page script for each flow with setup, failure card, facilitator prompts,
  expected recovery states, and stop conditions.
- A screenshot or scan of the final paper state for each flow.
- A findings note listing copy changes, missing evidence, dangerous assumptions,
  and follow-up implementation issue candidates.
- A traceability link from each implementation issue to the relevant
  `TV-PAPER-*` vector and `TV-THREAT-*` vector.

## Stop Conditions

- Do not open implementation issues for a flow if users believe Cabinet has
  transferred ownership, funds, reputation, registry state, or publication state
  before signed evidence and local acceptance allow it.
- Do not open implementation issues if the prototype hides the preserved failing
  evidence, the exact blocker, or the recovery path.
- Do not open implementation issues if the flow requires live P2P, real payment
  providers, public registry operations, or production catalogues to be
  understandable.
