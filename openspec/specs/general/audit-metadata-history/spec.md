## Purpose
Define record-level audit metadata and immutable audit history contracts across Cabinet entities.

## Requirements
### Requirement AUDIT-METADATA-001: Core entities SHALL persist created/updated actor and timestamp metadata
Core records SHALL include `created_at`, `created_by`, `updated_at`, and `updated_by` fields.

#### Scenario: Create then update entity
- **GIVEN** user creates a new core entity record
- **WHEN** record is created and later updated by a user/action
- **THEN** record MUST persist `created_at/by` at creation and `updated_at/by` on subsequent updates

### Requirement AUDIT-METADATA-002: Soft-delete capable entities SHALL persist delete actor and timestamp metadata
Where soft-delete lifecycle exists, records SHALL include `deleted_at` and `deleted_by` metadata.

#### Scenario: Soft delete record
- **GIVEN** entity supports soft-delete lifecycle
- **WHEN** user performs delete action
- **THEN** runtime MUST persist `deleted_at` and `deleted_by` while preserving recoverable record state

### Requirement AUDIT-HISTORY-001: Cabinet SHALL maintain immutable append-only audit events
Runtime SHALL persist append-only audit events for important mutations.

#### Scenario: Mutation audit event
- **GIVEN** create/update/delete/restore or critical state change occurs
- **WHEN** mutation is committed
- **THEN** runtime MUST append an `audit_event` containing actor, action, entity type/id, timestamp, and source context

### Requirement AUDIT-HISTORY-002: Audit events SHALL capture before/after change context for tracked fields
For tracked mutations, audit events SHALL include structured before/after snapshots or diffs.

#### Scenario: Update with tracked fields
- **GIVEN** tracked entity fields are modified
- **WHEN** audit event is written
- **THEN** event MUST include deterministic before/after values (or diff structure) for changed tracked fields

### Requirement AUDIT-HISTORY-003: Audit history SHALL be queryable by entity and timeline
Audit history API/query layer SHALL support entity-scoped and time-ordered retrieval.

#### Scenario: Fetch entity audit timeline
- **GIVEN** entity has prior audit events
- **WHEN** user/system requests audit timeline for entity ID
- **THEN** runtime MUST return ordered events with actor/action/timestamp and change context