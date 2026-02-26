## Purpose
Define Settings screen behavior and admin/maintenance use cases.

## Requirements
### Requirement: Settings screen SHALL provide API-backed configuration workflows
The screen SHALL load and persist profile settings and secrets via API.

#### Scenario: Use case - update profile settings
- **WHEN** user updates settings and saves
- **THEN** changes SHALL persist and survive reload

### Requirement: Settings screen SHALL provide maintenance and diagnostics controls
The screen SHALL provide backup/restore, reindex/repair, logging, and license workflows.

#### Scenario: Use case - restore backup
- **WHEN** user confirms restore for selected backup
- **THEN** app SHALL execute restore and report status
