## Purpose
Define Appearance settings screen behavior.

## Requirements
### Requirement UI-SCREEN-SETTINGS-APPEARANCE-001: Appearance screen SHALL manage theme and font preferences

#### Scenario: Update appearance preferences
- **GIVEN** user opens `/settings/appearance`
- **WHEN** user changes theme/font and saves
- **THEN** UI MUST apply preferences immediately and persist for next session
