# Trust Network Roadmap Glossary

These definitions are canonical for #1957 until the roadmap is approved or
revised. They distinguish Cabinet's signed trust objects from transports,
mirrors, local database rows, and future governance services.

| Term | Canonical definition | Authority and privacy boundary |
| --- | --- | --- |
| Identity | A versioned Cabinet identity object that binds a person, organization, or infrastructure operator to signing keys, rotation history, recovery metadata, and revocation state. | Authoritative only when its signature chain and revocation status validate; it does not publish private collection data by itself. |
| Specimen | A user's local record for one physical or digital collected item, including private notes, location, value, attachments, and provenance references. | The local private database is authoritative unless the user explicitly signs and shares selected claims about the specimen. |
| Catalogue item | A reference entry from a catalogue release or bundle that describes a known item type, taxonomy, edition, or verification hint. | A signed catalogue manifest and Merkle/hash proof verify the reference data; catalogue data does not override a user's private specimen record. |
| Proposal | A signed request or offer to start a trade, loan, reservation, feedback, attestation, endorsement, governance, or publication workflow. | A proposal is non-final until accepted or expired under its versioned rules; it may reveal only the fields intentionally included by its creator. |
| Reservation | A time-bounded signed claim that a specimen or proposal is held for a specific counterparty or workflow. | It is conflict evidence, not ownership transfer; expiry and cancellation must be deterministic and visible during reconciliation. |
| Receipt | A signed record that participants acknowledged a completed external exchange, transfer, loan, or other agreed outcome. | It may record external consideration notation but never escrow, stored value, checkout, payment accounts, or Cabinet-managed funds. |
| Ownership event | A signed event that records a claim about custody, ownership, transfer, return, cancellation, or dispute state for a specimen. | Events are append-only conflict evidence; local state derives from validated events and manual resolution, not from a remote ledger alone. |
| Feedback | A signed participant statement about a completed or aborted interaction, including context, rating inputs where approved, revocation, and appeal evidence. | Feedback remains auditable even when redacted or revoked; aggregate reputation must not use mutable totals as authority. |
| Attestation | A signed statement from an identity or infrastructure node asserting a bounded fact such as store verification, catalogue validation, event witnessing, or item inspection. | The attestation proves only the asserted fact, issuer, scope, timestamp, and revocation state; it does not grant broad trust. |
| Endorsement | A signed trust statement from one identity about another identity, node, catalogue source, or attestation issuer. | It contributes only to explicitly approved reputation or confidence models with bounded attenuation and does not replace direct validation. |
| Registry | A versioned signed set of public identity, node, governance, revocation, retirement, and appeal records. | The registry is a discovery and verification source; bootstrap, quorum, and migration rules must be approved before it can be authoritative. |
| Node | A Cabinet participant or infrastructure endpoint that stores, mirrors, validates, relays, or exchanges signed objects. | A node may transport or verify signed objects but must not gain ambient authority over private local databases or unpublished records. |
