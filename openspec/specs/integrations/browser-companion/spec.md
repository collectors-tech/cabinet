## Purpose
Define the Cabinet browser companion host/module contract for passive capture modules, payload envelopes, local sync, and safety boundaries.

## Requirements
### Requirement INTEGRATION-047: Browser companion modules MUST submit only authenticated passive capture payloads
Cabinet SHALL expose a minimal browser companion registry and payload sync endpoint so companion modules can capture marketplace evidence without hidden write actions.

#### Scenario: Passive module registry is discoverable
- **GIVEN** the browser companion asks Cabinet for available modules
- **WHEN** `GET /api/companion/modules` is requested
- **THEN** Cabinet MUST return registered module identifiers, supported capture actions, target site, and `passive_only=true`.

#### Scenario: Companion payload sync requires a profile-scoped bearer token
- **GIVEN** a browser companion module submits a captured marketplace payload
- **WHEN** `POST /api/companion/payloads` is requested without `Authorization: Bearer companion:<profile_id>` matching the payload profile
- **THEN** Cabinet MUST reject the payload as unauthorized.

#### Scenario: Companion payload sync rejects hidden write attempts
- **GIVEN** a browser companion module submits a captured marketplace payload
- **WHEN** the payload is not marked passive or reports an attempted write
- **THEN** Cabinet MUST reject the payload with a passive-capture error
- **AND** Cabinet MUST NOT create inventory, wishlist, purchase, or provider write records from that payload.

#### Scenario: Companion payload sync accepts validated passive evidence
- **GIVEN** a registered module submits a passive payload with a valid capture URL, payload type, confidence score, and matching profile-scoped bearer token
- **WHEN** Cabinet accepts the payload
- **THEN** the response MUST report `sync_mode=passive_capture`
- **AND** the response MUST report `remote_write=false`
- **AND** the response MUST include audit trail evidence naming the module and passive sync mode.
