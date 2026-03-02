## Purpose
Define Profile settings screen behavior.

## Requirements
### Requirement UI-SCREEN-SETTINGS-PROFILE-001: Profile screen SHALL support validated profile editing
Profile screen SHALL allow editing username, display email, bio, and URL list with inline validation.

#### Scenario: Save profile details
- **GIVEN** user opens `/settings`
- **WHEN** user submits valid profile values
- **THEN** runtime MUST persist values and UI MUST show deterministic success state

### Requirement UI-SCREEN-SETTINGS-PROFILE-002: Profile screen SHALL handle deterministic error states

#### Scenario: Profile save failure
- **GIVEN** profile API update fails
- **WHEN** user submits profile form
- **THEN** UI MUST show actionable error state and preserve entered values for retry
