## Purpose
Define license import, validation, and entitlement enforcement behavior.

## Requirements
### Requirement: Licensing SHALL enforce free/pro entitlements
Cabinet SHALL enforce free-tier limits and pro feature unlocks.

#### Scenario: Free-tier limit reached
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** item count exceeds free tier limit
- **THEN** Cabinet SHALL block further limit-gated creation actions

### Requirement: License import SHALL verify signature and validity offline
Cabinet SHALL support offline license validation with signature verification.

#### Scenario: License import
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** user imports signed license file
- **THEN** Cabinet SHALL verify signature and apply entitlement state

### Requirement: License status SHALL be user-visible
Cabinet SHALL expose current license state and feature gate status in settings.

#### Scenario: License status refresh
- **GIVEN** the required preconditions and context for this scenario are satisfied
- **WHEN** settings requests license status
- **THEN** Cabinet SHALL return active entitlement summary
