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
