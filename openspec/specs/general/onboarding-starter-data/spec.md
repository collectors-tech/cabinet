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

#### Scenario: Seed representative category coverage
- **GIVEN** active profile has no prior onboarding sample dataset
- **WHEN** runtime seeds onboarding sample data
- **THEN** sample items MUST cover multiple distinct collecting categories so inventory/category views do not look empty or repetitive
- **AND** MUST include a mix of immediately-owned examples and wishlist-target examples
- **AND** rerunning the same seed for the same profile MUST remain idempotent

### Requirement ONBOARDING-STARTER-DATA-003: Starter flow SHALL support import-existing alternative
Cabinet SHALL provide import-existing-collection option as an alternative to sample data.

#### Scenario: Choose import existing path
- **GIVEN** setup wizard has completed and starter setup flow is active
- **WHEN** user selects `Import Existing Collection`
- **THEN** flow MUST route to import path without auto-seeding sample data

### Requirement ONBOARDING-STARTER-DATA-004: Sample data SHALL be explicitly identified as showcase data
Cabinet SHALL mark seeded onboarding sample data as showcase/example records in the API result and active profile settings so the dataset cannot be confused with a real working collection.

#### Scenario: Sample seed provenance
- **GIVEN** an active profile has no prior onboarding sample dataset
- **WHEN** runtime seeds onboarding sample data
- **THEN** the seed response MUST include `dataset_kind=sample_showcase`, a sample-oriented dataset label, and an explicit sample-data disclosure
- **AND** the active profile settings MUST persist the sample dataset kind and disclosure for downstream profile/database context surfaces
