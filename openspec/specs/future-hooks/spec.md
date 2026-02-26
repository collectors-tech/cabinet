## Purpose
Define non-active extension hooks and compatibility boundaries.

## Requirements
### Requirement: Future extension hooks SHALL exist but remain non-active by default
Cabinet SHALL provide disabled scaffolds for additional AI providers, scanner providers, share/compare, for-sale flag, and structured offers.

#### Scenario: Runtime hook listing
- **WHEN** hooks are queried in current version
- **THEN** unsupported hooks SHALL be marked disabled and non-operative

### Requirement: Disabled hooks SHALL not affect active workflows
Cabinet SHALL prevent disabled hook paths from mutating active data flows.

#### Scenario: Disabled hook invocation
- **WHEN** user or system attempts disabled hook action
- **THEN** Cabinet SHALL return explicit not-active response
