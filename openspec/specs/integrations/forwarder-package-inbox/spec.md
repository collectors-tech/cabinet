## Purpose
Define the first Cabinet package inbox contract for Stackry and future freight-forwarding sources.

## Requirements
### Requirement INTEGRATION-029: Forwarder package imports SHALL normalize package identity and provenance
Cabinet SHALL normalize profile-scoped forwarder package imports from API, capture, email, CSV, and manual sources into a stable package inbox record that preserves provider identity, external package identifiers, shipment/tracking references, status, warehouse metadata, weight, and raw source provenance.

#### Scenario: Normalize a Stackry package import
- **GIVEN** an authenticated user has an active profile and Cabinet receives a Stackry package import with source, external package id, shipment metadata, package status, and raw source payload
- **WHEN** Cabinet normalizes the import
- **THEN** Cabinet MUST trim user/source input, preserve external identifiers and raw provenance, classify the package status into a known inbox state, and assign a deterministic provenance key for duplicate protection.

### Requirement INTEGRATION-030: Forwarder package inbox SHALL prevent duplicate package records
Cabinet SHALL deduplicate package imports by profile, provider, source, and external package id so repeated Stackry/API/email/CSV/manual ingestion updates the existing inbox package instead of creating duplicates.

#### Scenario: Import the same package twice
- **GIVEN** a forwarder package record already exists for a profile, provider, source, and external package id
- **WHEN** Cabinet receives a second import for the same package with updated status or weight fields
- **THEN** Cabinet MUST update the existing package record and keep a single inbox entry for that profile and provenance key.

### Requirement INTEGRATION-031: Forwarder package inbox SHALL expose import and list APIs
Cabinet SHALL persist normalized forwarder package imports and expose an API for package import/upsert and filtered inbox listing without requiring a provider-specific live adapter.

#### Scenario: Import and list persisted forwarder packages
- **GIVEN** a user imports a Stackry or freight-forwarder package with profile, provider, source, external package id, status, and optional raw provenance
- **WHEN** Cabinet accepts the import through the forwarding package API
- **THEN** Cabinet MUST persist the normalized package record, return the saved package identity, allow status-filtered listing, and reject invalid imports without creating a package.

### Requirement INTEGRATION-032: Forwarder package inbox SHALL expose manual import UI
Cabinet SHALL provide a visible package inbox surface that lets users manually import Stackry or freight-forwarder package records, refresh the profile-scoped package list, see normalized provenance, and recover from validation errors without requiring a live provider adapter.

#### Scenario: Import a manual Stackry package from the inbox UI
- **GIVEN** a user is reviewing purchase and forwarding intake work from the Cabinet inbox
- **WHEN** the user submits a manual Stackry package with profile, provider, source, external package id, status, tracking, warehouse, and weight fields
- **THEN** Cabinet MUST call the forwarder package import API, refresh the package list, show the imported package with status and provenance key, and display API validation errors inline when required identity fields are missing.

### Requirement INTEGRATION-033: Forwarder package inbox SHALL parse CSV package imports with row-level validation
Cabinet SHALL parse Stackry/freight-forwarder CSV package rows into normalized package imports while preserving raw row provenance and reporting invalid rows without discarding valid package rows from the same file.

#### Scenario: Parse a mixed-validity CSV package import
- **GIVEN** a user has a CSV export or manually prepared file containing forwarder package ids, statuses, shipment/tracking fields, warehouse metadata, and package weights
- **WHEN** Cabinet parses the CSV for a profile and provider
- **THEN** Cabinet MUST map supported header aliases into normalized CSV package imports, preserve the original row values as raw provenance, return row-specific validation errors for missing identity or invalid weight values, and keep valid rows available for package upsert.

### Requirement INTEGRATION-034: Forwarder package inbox SHALL import CSV rows through API and UI
Cabinet SHALL let users submit Stackry/freight-forwarder CSV package rows from the package inbox UI, persist valid rows through the forwarder package API, refresh the profile package list, and show row-specific validation errors for rows that were not imported.

#### Scenario: Import mixed-validity CSV rows from the inbox UI
- **GIVEN** a user has an active profile and enters CSV package rows with at least one valid package and at least one invalid row
- **WHEN** the user submits the CSV import from the forwarder package inbox
- **THEN** Cabinet MUST upsert valid rows as source `csv`, preserve their provenance keys, keep invalid rows out of the package list, refresh the visible package list, and display row-specific CSV errors inline.

### Requirement INTEGRATION-035: Forwarder package inbox SHALL parse email package notices
Cabinet SHALL parse Stackry/freight-forwarder email package notices into normalized package imports while preserving message provenance and rejecting notices that lack a stable package identity or valid package fields.

#### Scenario: Parse a package email notice
- **GIVEN** a user has a forwarded package notification email containing package id, status, shipment/tracking fields, received timestamp, sender, warehouse location, and package weight
- **WHEN** Cabinet parses the email for a profile and provider
- **THEN** Cabinet MUST map supported label aliases into a normalized source `email` package import, preserve the source message id and raw body as provenance, return deterministic validation errors for missing identity or invalid weight values, and produce an email provenance key for package upsert.

### Requirement INTEGRATION-036: Forwarder package inbox SHALL import email notices through API and UI
Cabinet SHALL let users submit Stackry/freight-forwarder package email notice text from the package inbox UI, persist the parsed notice through the forwarder package API, refresh the profile package list, and show deterministic validation errors inline when the notice cannot be normalized.

#### Scenario: Import a package email notice from the inbox UI
- **GIVEN** a user has an active profile and enters a package notification email containing package id, status, shipment/tracking fields, sender, warehouse location, and package weight
- **WHEN** the user submits the email import from the forwarder package inbox
- **THEN** Cabinet MUST upsert the notice as source `email`, preserve its email provenance key, refresh the visible package list, and display parser/API validation errors inline without creating invalid package records.

### Requirement INTEGRATION-037: Forwarder package inbox SHALL expose package and shipment detail
Cabinet SHALL provide a visible detail view for each forwarder package record that exposes package identity, shipment/tracking fields, source timestamps, and raw provenance without requiring a provider-specific adapter.

#### Scenario: Review a forwarder package detail record
- **GIVEN** a user has imported or refreshed a Stackry/freight-forwarder package record with shipment, tracking, timestamp, and raw provenance fields
- **WHEN** the user opens the package detail view from the forwarder package inbox
- **THEN** Cabinet MUST show package identity, shipment id, tracking number, received/created/updated timestamps, and raw source provenance while keeping missing optional values explicit as pending.

### Requirement INTEGRATION-038: Forwarder package inbox SHALL reconcile packages to purchase arrivals
Cabinet SHALL let a profile-scoped forwarder package be linked to one confirmed inventory item and purchase/expected-arrival target while preserving review provenance and rejecting ambiguous relinks to different targets.

#### Scenario: Link a forwarder package to an expected purchase arrival
- **GIVEN** a user has an active profile with a persisted forwarder package, inventory item, purchase lifecycle entry, and expected arrival
- **WHEN** Cabinet receives a reconciliation link for that package, item, lifecycle entry, and arrival
- **THEN** Cabinet MUST verify every target belongs to the active profile, persist a durable package-to-purchase link with source/notes provenance, return the saved link through the API, and reject attempts to relink the same package to a different item or arrival.

### Requirement INTEGRATION-039: Forwarder package inbox SHALL expose reconciliation links in the UI
Cabinet SHALL let a user review and create package-to-purchase-arrival reconciliation links from the forwarder package inbox detail view, using the profile-scoped package-link API and surfacing ambiguous-link failures inline.

#### Scenario: Link a package from the inbox detail panel
- **GIVEN** a user has opened a persisted forwarder package detail record from the inbox
- **WHEN** the user reviews existing link state and submits an item, lifecycle entry, expected arrival, source, and notes
- **THEN** Cabinet MUST create the package reconciliation link through \`/api/forwarding/package-links\`, display the saved link result, refresh the package link state, and show API validation errors inline when the package is already linked to a different target.

### Requirement INTEGRATION-040: Forwarder package inbox SHALL suggest purchase-arrival matches with confidence explanations
Cabinet SHALL compute deterministic, non-mutating match suggestions between unlinked forwarder packages and profile-scoped purchase expected arrivals, including confidence labels, scored signal explanations, and audit-ready provenance for later manual confirmation or override.

#### Scenario: Suggest a package-to-purchase-arrival match
- **GIVEN** a profile has an unlinked forwarder package plus purchase lifecycle and expected-arrival records with matching tracking, package/order references, item text, seller/source, quantity, and date signals
- **WHEN** Cabinet requests forwarder package match suggestions
- **THEN** Cabinet MUST return ordered suggestions with package, item, lifecycle, and expected-arrival IDs, UI-readable `confidence_score` and `confidence_label`, per-signal `score` and evidence, and audit trail text, without creating or modifying any package reconciliation link.

### Requirement INTEGRATION-041: Forwarder package inbox SHALL audit confirm, override, and unlink decisions
Cabinet SHALL persist explicit decision metadata and audit events when a user confirms a suggested package reconciliation, overrides an existing target, or unlinks a package from a purchase arrival.

#### Scenario: Confirm, override, and unlink package reconciliation decisions
- **GIVEN** a profile has a persisted forwarder package, candidate purchase arrivals, and a current package reconciliation link
- **WHEN** a user confirms a suggested target, overrides the link to a different target with explicit override intent, or unlinks the package
- **THEN** Cabinet MUST persist decision/source/notes/audit-trail metadata, reject ambiguous target changes unless override is explicit, preserve durable audit events for confirmed/override/unlinked decisions, remove the active link after unlink, and return the decision/event evidence through the package-link API.

### Requirement INTEGRATION-042: Forwarder package inbox UI SHALL expose link decision audit controls
Cabinet SHALL let users confirm, override, and unlink package reconciliation targets from the forwarder package detail panel while showing the active decision metadata and durable audit event history returned by the package-link API.

#### Scenario: Review and update package link decisions from the inbox UI
- **GIVEN** a user has opened a persisted forwarder package detail record with existing or pending reconciliation link decisions
- **WHEN** the user confirms a link, overrides the target with explicit override intent, or unlinks the package from the detail panel
- **THEN** Cabinet MUST send the corresponding decision/source/notes/audit-trail payload to `/api/forwarding/package-links`, refresh active link and event state after each mutation, show confirmed/override/unlinked decision evidence inline, and surface API validation errors without losing the user's entered target fields.

### Requirement INTEGRATION-043: Forwarder package inbox UI SHALL expose match suggestions before confirmation
Cabinet SHALL let users request non-mutating forwarder package match suggestions from the package inbox, inspect confidence labels, scored signal evidence, and audit-ready explanation text, then use a suggestion to prefill the existing reconciliation confirmation form without creating a link until the user confirms.

#### Scenario: Review and accept a suggested package match
- **GIVEN** a user has a persisted unlinked forwarder package and Cabinet has computed package-to-purchase-arrival suggestions
- **WHEN** the user requests match suggestions, opens a package detail panel, and chooses a suggested target
- **THEN** Cabinet MUST show the suggestion confidence, score, signal evidence, expected-arrival target, and audit trail text, prefill item/lifecycle/arrival/source fields from the suggestion, and only call `/api/forwarding/package-links` after the user explicitly confirms the link.

### Requirement INTEGRATION-044: Forwarder package inbox UI SHALL expose complete decision audit event evidence
Cabinet SHALL show the durable package-link audit event details returned by the package-link API so a reviewer can understand the current target, prior target, timestamp, source, notes, and audit-trail evidence for confirmed, override, and unlink decisions.

#### Scenario: Review complete package decision event history
- **GIVEN** a user has opened a forwarder package detail panel with confirmed, override, or unlink audit events
- **WHEN** Cabinet refreshes package link state from `/api/forwarding/package-links`
- **THEN** Cabinet MUST show each event's action, current item/lifecycle/arrival target when present, previous item/lifecycle/arrival target when present, source, created timestamp, notes, and audit-trail entries without hiding the active link state.

### Requirement INTEGRATION-045: Forwarder package inbox UI SHALL summarize reconciliation review state
Cabinet SHALL show a compact forwarder package review summary so a reviewer can see package volume, known linked/unlinked reconciliation state, loaded audit event volume, and loaded match suggestion volume before drilling into individual package records.

#### Scenario: Review package reconciliation summary
- **GIVEN** the forwarder package inbox has loaded package records, package-link state, decision audit events, or match suggestions
- **WHEN** the user reviews the forwarder package inbox
- **THEN** Cabinet MUST show package, linked, unlinked, audit-event, and suggestion counts that update after package refreshes, link-state refreshes, link/unlink decisions, and match suggestion loads.

### Requirement INTEGRATION-046: Forwarder package inbox UI SHALL filter reconciliation review states
Cabinet SHALL let reviewers filter the forwarder package inbox by all packages, linked packages, unlinked packages, and packages with loaded match suggestions using the currently loaded reconciliation evidence.

#### Scenario: Filter package reconciliation review states
- **GIVEN** the forwarder package inbox has loaded package records plus package-link or match-suggestion evidence
- **WHEN** the user chooses a reconciliation review-state filter
- **THEN** Cabinet MUST show only the packages matching that filter, preserve the non-mutating review evidence, and show an empty filtered-state message when no packages match without hiding the original package count.

### Requirement INTEGRATION-048: Forwarder package inbox UI SHALL label per-package reconciliation evidence
Cabinet SHALL show package-row reconciliation evidence labels derived from the currently loaded active link, match suggestion, and audit event state so reviewers can decide which package rows need drill-in review without relying only on aggregate counts.

#### Scenario: Review row-level reconciliation evidence labels
- **GIVEN** the forwarder package inbox has loaded package records plus package-link, match-suggestion, or decision audit event evidence
- **WHEN** the user reviews the visible package rows after refreshing link or suggestion evidence
- **THEN** Cabinet MUST show each visible package row's loaded active-link, suggestion, and audit-event counts, and MUST show an explicit no-evidence state when no reconciliation evidence is loaded for that package.

### Requirement INTEGRATION-049: Forwarder package APIs SHALL be documented in OpenAPI
Cabinet SHALL document the forwarder package import, CSV/email ingestion, reconciliation link decision, unlink, audit event, and match suggestion APIs in docs/api/openapi.yaml so integration clients can rely on the same request and response contract covered by app tests.

#### Scenario: Review the forwarding API contract
- **GIVEN** Cabinet exposes forwarding package endpoints for imports, links, audit events, and match suggestions
- **WHEN** a developer reviews the OpenAPI specification
- **THEN** the specification MUST include /api/forwarding/packages, /api/forwarding/packages/import-csv, /api/forwarding/packages/import-email, /api/forwarding/package-links, and /api/forwarding/package-match-suggestions with the persisted package, link, event, non-mutating suggestion, and row-error schemas used by the tested API responses.

### Requirement INTEGRATION-050: Forwarder package match suggestions SHALL summarize confidence buckets
Cabinet SHALL include deterministic review-summary metadata with non-mutating forwarder package match suggestions so reviewers and UI clients can see total candidate count, scoped-package state, and high/medium/low confidence buckets without re-scoring suggestion rows client-side.

#### Scenario: Review confidence-bucketed match suggestions
- **GIVEN** Cabinet has computed zero or more non-mutating package-to-purchase-arrival match suggestions for a profile or a scoped package
- **WHEN** a client requests `/api/forwarding/package-match-suggestions`
- **THEN** Cabinet MUST return `summary.count`, `summary.high_confidence`, `summary.medium_confidence`, `summary.low_confidence`, and `summary.scoped_packages` values derived from the returned suggestions while keeping the suggestion request non-mutating.

### Requirement INTEGRATION-051: Forwarder package inbox UI SHALL surface suggestion confidence summary
Cabinet SHALL display the non-mutating match-suggestion summary in the forwarder package inbox so reviewers can see total candidates, scoped packages, and high/medium/low confidence buckets before drilling into package rows.

#### Scenario: Review match-suggestion confidence buckets in the inbox
- **GIVEN** the forwarder package inbox has loaded package records and match-suggestion summary metadata
- **WHEN** the user requests match suggestions from the inbox
- **THEN** Cabinet MUST show candidate count, scoped package count, high-confidence count, medium-confidence count, and low-confidence count from the API response without mutating package reconciliation links.

### Requirement INTEGRATION-052: Forwarder package match suggestions SHALL support confidence-label filtering
Cabinet SHALL let API clients request non-mutating package match suggestions filtered to a specific confidence label so review queues can focus on high, medium, or low confidence candidates without client-side re-scoring.

#### Scenario: Filter package match suggestions by confidence label
- **GIVEN** Cabinet has computed non-mutating package-to-purchase-arrival match suggestions across multiple confidence labels
- **WHEN** a client requests `/api/forwarding/package-match-suggestions?confidence_label=medium`
- **THEN** Cabinet MUST return only medium-confidence suggestions, set `confidence_filter` to `medium`, derive summary bucket counts from the filtered result set, reject unknown confidence labels with a validation error, and avoid creating or modifying reconciliation links.

### Requirement INTEGRATION-053: Forwarder package inbox UI SHALL filter match suggestions by confidence
Cabinet SHALL let reviewers choose all, high, medium, or low confidence before loading non-mutating forwarder package match suggestions, send the selected confidence label to the suggestions API, and show the active filter with summary counts derived from the returned filtered result set.

#### Scenario: Review high-confidence package match suggestions
- **GIVEN** the reviewer is in the forwarder package inbox and wants to focus reconciliation review
- **WHEN** the reviewer selects the high-confidence suggestion filter and requests matches
- **THEN** Cabinet MUST request `/api/forwarding/package-match-suggestions?confidence_label=high`, render only returned high-confidence suggestions, display the active filter and summary counts, and avoid creating or modifying package reconciliation links until the reviewer explicitly confirms a link.

### Requirement INTEGRATION-054: Forwarder package inbox UI SHALL scope match suggestions to an opened package
Cabinet SHALL let reviewers request non-mutating match suggestions for a single opened forwarder package while preserving the selected confidence filter and returned summary evidence.

#### Scenario: Review scoped package match suggestions
- **GIVEN** the reviewer has opened a forwarder package detail panel and selected a suggestion confidence filter
- **WHEN** the reviewer requests matches for that package
- **THEN** Cabinet MUST request `/api/forwarding/package-match-suggestions` with both `package_id` and the selected `confidence_label`, update the visible suggestion summary and result text from the scoped response, and avoid creating or modifying package reconciliation links until the reviewer explicitly confirms a link.
