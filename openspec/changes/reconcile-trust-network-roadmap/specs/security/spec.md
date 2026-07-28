## MODIFIED Requirements

### Requirement: Security requirements SHALL bind decisions to threat evidence

Cabinet security architecture MUST map trust-network decisions to threat-model
coverage and user-visible recovery behavior before implementation work begins.

#### Scenario: Trust-network threats have mapped mitigations

- **GIVEN** a trust-network decision authorizes future implementation
- **WHEN** the decision is recorded
- **THEN** it MUST identify relevant threats, including fake peer/store,
  Sybil/eclipsing, tampering, forged or deleted feedback, key compromise,
  catalogue poisoning, and privacy leakage where applicable
- **AND** it MUST name the expected mitigation, architecture test vector, and
  user-visible recovery or warning behavior.
