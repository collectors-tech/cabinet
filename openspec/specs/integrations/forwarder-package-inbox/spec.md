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
- **THEN** Cabinet MUST return ordered suggestions with package, item, lifecycle, and expected-arrival IDs, confidence score and label, per-signal evidence, and audit trail text, without creating or modifying any package reconciliation link.

### Requirement INTEGRATION-041: Forwarder package inbox SHALL audit confirm, override, and unlink decisions
Cabinet SHALL persist explicit decision metadata and audit events when a user confirms a suggested package reconciliation, overrides an existing target, or unlinks a package from a purchase arrival.

#### Scenario: Confirm, override, and unlink package reconciliation decisions
- **GIVEN** a profile has a persisted forwarder package, candidate purchase arrivals, and a current package reconciliation link
- **WHEN** a user confirms a suggested target, overrides the link to a different target with explicit override intent, or unlinks the package
- **THEN** Cabinet MUST persist decision/source/notes/audit-trail metadata, reject ambiguous target changes unless override is explicit, preserve durable audit events for confirmed/override/unlinked decisions, remove the active link after unlink, and return the decision/event evidence through the package-link API.
