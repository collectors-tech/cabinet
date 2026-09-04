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

### Requirement DISCOVERY-005: Discoveries SHALL aggregate found opportunities into a collector-facing dashboard
Cabinet SHALL treat Discoveries as the review dashboard for opportunities found by
Wishlist matching, Market Watch output, provider/store scans, scanner runs, and public
or trade-enabled inventory sources. Discoveries SHALL NOT own Market Watch query
configuration or provider run controls; those source workflows publish reviewable
outputs into Discoveries.

#### Scenario: Found-opportunity dashboard aggregation
- **GIVEN** Wishlist, Market Watch, provider/store, scanner, or public/shared inventory workflows emit candidate findings
- **WHEN** Cabinet records candidates for Discoveries review
- **THEN** Discoveries MUST support dashboard summary groups for best deals, wishlist matches, new findings, Market Watch outputs, and provider/store attention
- **AND** Discoveries MUST support source filter groups for all discoveries, wishlist matches, great prices, Market Watch, stores/providers, other public or shared inventories, and ignored or archived candidates
- **AND** Discoveries MUST distinguish Market Watch query controls from Discoveries output review by linking back to the originating query/run surface rather than duplicating query creation or run controls
- **AND** "other inventories" MUST mean public, trade-enabled, or explicitly shared inventory sources only

#### Scenario: Wishlist price-match candidate
- **GIVEN** a Wishlist target has a target price and Cabinet finds a provider/store listing below that target
- **WHEN** the listing becomes a Discoveries candidate
- **THEN** the candidate MUST preserve wishlist id, match reason, target price, observed price, savings delta, source/provider, seller/source label, availability, first/last seen timestamps, source-result URL, and ranking score where available
- **AND** the candidate MUST remain a wanted-opportunity record until the user explicitly promotes it to Wishlist follow-up, Purchase, Inventory, ignore, or archive

#### Scenario: Market Watch output candidate
- **GIVEN** Market Watch executes a saved query or scheduled refresh and emits matching output
- **WHEN** Cabinet sends that output to Discoveries
- **THEN** the Discoveries candidate MUST preserve provider, query set id, query name or run identifier, listing/result identifier, result URL, observed price/currency, seller/source label, availability, source trust or provider-health status, and current review status
- **AND** Discoveries MUST expose a source-result review path and a Market Watch handoff path without presenting the candidate as a new query-run configuration surface

### Requirement DISCOVERY-006: Discoveries SHALL rank and act on candidates without leaking private collection data
Cabinet SHALL rank discovery candidates by collector value while preserving ownership
and privacy boundaries. Private collector inventory, storage location, private notes, and
unpublished collection value SHALL NOT be shown in Discoveries as source evidence or
comparison material.

#### Scenario: Candidate ranking and state model
- **GIVEN** Discoveries has candidates with wishlist, pricing, source, stock, recency, confidence, and status metadata
- **WHEN** the dashboard ranks and renders candidates
- **THEN** wishlist matches and great-price opportunities MUST rank ahead of lower-signal candidates when their metadata supports that ordering
- **AND** ranking MUST consider match type/reason, target price, baseline or market price, price delta amount and percent, availability, recency, source trust status, and confidence/review signal where available
- **AND** each candidate MUST expose one current review status from `new`, `reviewing`, `wishlisted`, `purchase_candidate`, `inventory_candidate`, `ignored`, or `archived`

#### Scenario: Privacy-safe contextual actions
- **GIVEN** a discovery candidate comes from public/shared inventory, store/provider output, Wishlist matching, or Market Watch output
- **WHEN** the user reviews contextual actions
- **THEN** actions MUST be limited to source-result review, Wishlist follow-up, Purchase follow-up, Inventory handoff, ignore, archive, or return-to-source workflow where applicable
- **AND** Inventory/Purchase actions MUST NOT claim ownership until the user explicitly confirms an owned or purchased state
- **AND** Discoveries MUST NOT reveal private collector inventory records, private notes, storage locations, unpublished collection values, or non-shared inventory comparisons as source evidence

