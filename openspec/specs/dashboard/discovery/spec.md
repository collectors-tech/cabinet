## Purpose
Define Discoveries as the candidate-item triage inbox for findings that Cabinet has found
through Market Watch, provider imports, scanner runs, or buyer-interest sync before they
become Wishlist, Inventory, Purchase, or archived records.

Discoveries remains a first-class page. It is not a duplicate query-run workspace and it
does not own provider search configuration. Market Watch and provider workflows own query
execution; Discoveries owns accepted candidate findings and the decision trail that turns
those findings into user-managed collection records.

## Requirements
### Requirement DISCOVERY-001: Not-in-collection panel SHALL support actionable triage
Cabinet SHALL support ignore, add-to-wishlist, track-price, and create-item actions.

#### Scenario: Discovery triage action
- **GIVEN** a candidate is in not-in-collection state
- **WHEN** user applies a triage action
- **THEN** Cabinet SHALL persist the requested action outcome

### Requirement DISCOVERY-002: Discovery filters SHALL support price/query/date
Cabinet SHALL provide filtering for discovery queue triage.

#### Scenario: Discovery filtered view
- **GIVEN** discovery queue has candidates
- **WHEN** user applies query and date filters
- **THEN** panel SHALL return filtered not-in-collection candidates

### Requirement DISCOVERY-003: Discoveries SHALL preserve candidate provenance and triage status
Cabinet SHALL persist each discovery candidate with source provenance, candidate metadata,
status, confidence or review notes, and source-result linkage sufficient to audit why the
candidate exists.

#### Scenario: Candidate provenance contract
- **GIVEN** Market Watch, scanner, provider import, or buyer-interest sync emits a candidate finding
- **WHEN** Cabinet records the candidate for Discoveries triage
- **THEN** the persisted discovery record MUST include source/provider identifier, source query or run identifier when available, listing/result identifier, URL, title, observed price, observed currency, seller/source label, first-seen timestamp, last-seen timestamp, confidence or needs-review signal, current triage status, and optional reviewer notes
- **AND** the record MUST retain a source-result link back to the originating query/result surface where that surface exists

### Requirement DISCOVERY-004: Discoveries SHALL define clear handoff destinations
Cabinet SHALL support explicit, auditable handoff decisions from each discovery candidate
to Wishlist, Inventory, Purchase flow, or archive/ignore.

#### Scenario: Discovery handoff boundaries
- **GIVEN** user reviews a discovery candidate
- **WHEN** user promotes or dismisses the candidate
- **THEN** Cabinet MUST persist exactly one current triage status from `new`, `reviewing`, `wishlisted`, `inventory_candidate`, `purchase_candidate`, `ignored`, or `archived`
- **AND** Wishlist promotion MUST create or link a Wishlist entry without claiming ownership
- **AND** Inventory or Purchase promotion MUST create or link a downstream record/workflow only when the user confirms the item is owned or purchased
- **AND** ignore/archive MUST preserve the audit trail and prevent the same source result from reappearing as an unreviewed discovery unless the source materially changes

