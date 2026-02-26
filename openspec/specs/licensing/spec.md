## Purpose
Define signed license import, verification, and runtime license state behavior.

## Requirements
### Requirement: License import SHALL verify signature and validity offline
Cabinet SHALL support offline license validation with signature verification.

#### Scenario: License import
- **GIVEN** user has a signed license file
- **WHEN** license is imported
- **THEN** Cabinet SHALL verify signature and apply license state

### Requirement: License status SHALL be user-visible
Cabinet SHALL expose current license state in settings.

#### Scenario: License status refresh
- **GIVEN** settings requests license status
- **WHEN** license status API is called
- **THEN** Cabinet SHALL return active license summary
