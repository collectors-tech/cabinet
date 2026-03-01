## Purpose
Define feature-gate entitlement enforcement behavior.

## Requirements
### Requirement ENTITLEMENTS-001: Entitlements SHALL enforce free/pro limits
Cabinet SHALL enforce free-tier limits and pro feature unlocks.

#### Scenario: Free-tier limit reached
- **GIVEN** item count exceeds free tier limit
- **WHEN** user attempts gated creation action
- **THEN** Cabinet SHALL block the action and expose entitlement guidance

