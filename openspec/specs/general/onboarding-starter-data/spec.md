## Purpose
Define onboarding wizard behavior for starter setup and sample data bootstrapping.

## Requirements
### Requirement ONBOARDING-STARTER-DATA-001: Onboarding SHALL remain simple for first-time users
Cabinet SHALL guide first-time users through minimal required setup before showing advanced forms.

#### Scenario: First profile onboarding
- **GIVEN** runtime setup configuration is missing for the local instance
- **WHEN** user opens sign-in route
- **THEN** UI MUST present a full-screen setup wizard before auth form and MUST NOT render setup controls inside authenticated Home

### Requirement ONBOARDING-STARTER-DATA-002: Starter flow SHALL support sample data seeding after identity completion
Cabinet SHALL support one-click sample data import in onboarding after identity/auth requirements are satisfied.

#### Scenario: Seed sample data
- **GIVEN** setup wizard has completed runtime config and identity setup is complete for active profile
- **WHEN** user selects `Use Sample Data` from starter setup flows
- **THEN** runtime MUST create starter dataset and return seed summary (`folders_created`, `items_created`, `media_created`)

### Requirement ONBOARDING-STARTER-DATA-003: Starter flow SHALL support import-existing alternative
Cabinet SHALL provide import-existing-collection option as an alternative to sample data.

#### Scenario: Choose import existing path
- **GIVEN** setup wizard has completed and starter setup flow is active
- **WHEN** user selects `Import Existing Collection`
- **THEN** flow MUST route to import path without auto-seeding sample data
