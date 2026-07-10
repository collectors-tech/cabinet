## Purpose
Define feature-gate entitlement enforcement behavior.

## Requirements
### Requirement ENTITLEMENTS-001: Entitlements SHALL enforce Free/Plus/Pro beta limits
Cabinet SHALL enforce the canonical Free/Plus/Pro beta entitlement matrix and normalize legacy runtime aliases into those user-visible plans.

#### Scenario: Free-tier limit reached
- **GIVEN** item count exceeds the 250-item Free tier limit
- **WHEN** user attempts gated creation action
- **THEN** Cabinet SHALL block the action and expose entitlement guidance

#### Scenario: Legacy plan aliases normalize to beta plans
- **GIVEN** cloud auth or billing sends legacy aliases (`mvp`, `creator`, `teams`, or `paid`)
- **WHEN** Cabinet resolves the current entitlement plan
- **THEN** `mvp` SHALL resolve as `free`, `creator` SHALL resolve as `plus`, and `teams`/`paid` SHALL resolve as `pro`
- **AND** API responses and feature gates SHALL use the canonical `free`, `plus`, or `pro` plan names

#### Scenario: Beta feature matrix
- **GIVEN** the user is on Free, Plus, or Pro
- **WHEN** Cabinet evaluates feature access
- **THEN** Free SHALL retain collection-core and export/read access up to the Free item limit
- **AND** Plus SHALL unlock beta operational automation without Assistant access
- **AND** Pro SHALL unlock Plus capabilities plus Assistant/AI capabilities

#### Scenario: Downgrade and expiry preserve owned data access
- **GIVEN** a user owns Cabinet data created before downgrade or license expiry
- **WHEN** the account resolves to Free because billing was downgraded or the signed license expired
- **THEN** Cabinet SHALL keep existing owned data readable
- **AND** Cabinet SHALL keep JSON and item CSV data export available
- **AND** Cabinet SHALL NOT delete or hide owned records because of the entitlement transition

