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
