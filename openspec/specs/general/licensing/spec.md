## Purpose
Define signed license import, verification, and runtime license state behavior.

## Requirements
### Requirement LICENSING-001: License import SHALL verify signature and validity offline
Cabinet SHALL support offline license validation with signature verification.

#### Scenario: License import
- **GIVEN** user has a signed license file
- **WHEN** license is imported
- **THEN** Cabinet SHALL verify signature and apply license state

### Requirement LICENSING-002: License status SHALL be user-visible
Cabinet SHALL expose current license state in settings.

#### Scenario: License status refresh
- **GIVEN** settings requests license status
- **WHEN** license status API is called
- **THEN** Cabinet SHALL return active license summary

#### Scenario: Founding license import from settings
- **GIVEN** a user has an active profile and a signed founding license payload
- **WHEN** the user imports the signed license from Settings Billing
- **THEN** Cabinet SHALL submit the signed payload to the license import API
- **AND** the refreshed Settings Billing state SHALL show the resulting signed license status
