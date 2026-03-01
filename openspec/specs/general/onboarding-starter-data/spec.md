## Purpose
Define onboarding wizard behavior for starter setup and sample data bootstrapping.

## Requirements
### Requirement ONBOARDING-STARTER-DATA-001: Onboarding SHALL remain simple for first-time users
Cabinet SHALL guide first-time users through minimal required setup before showing advanced forms.

#### Scenario: First profile onboarding
- **GIVEN** active profile has no completed onboarding state
- **WHEN** user opens home workspace
- **THEN** UI MUST present guided onboarding steps and MUST NOT require full advanced-form completion on first view

### Requirement ONBOARDING-STARTER-DATA-002: Starter flow SHALL support sample data seeding after identity completion
Cabinet SHALL support one-click sample data import in onboarding after identity/auth requirements are satisfied.

#### Scenario: Seed sample data
- **GIVEN** identity setup is complete and sample data is not yet seeded for profile
- **WHEN** user selects `Use Sample Data` in onboarding wizard
- **THEN** runtime MUST create starter dataset and return seed summary (`folders_created`, `items_created`, `media_created`)

### Requirement ONBOARDING-STARTER-DATA-003: Starter flow SHALL support import-existing alternative
Cabinet SHALL provide import-existing-collection option as an alternative to sample data.

#### Scenario: Choose import existing path
- **GIVEN** onboarding wizard is on starter-data step
- **WHEN** user selects `Import Existing Collection`
- **THEN** wizard MUST route to import flow without auto-seeding sample data
