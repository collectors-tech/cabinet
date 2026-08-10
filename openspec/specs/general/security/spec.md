## Purpose
Define security and privacy controls for secrets and offline trust.

## Requirements
### Requirement SECURITY-001: Secrets SHALL never be stored in plaintext SQLite records
Cabinet SHALL store sensitive keys in OS-backed secure storage.

#### Scenario: Secret persistence
- **GIVEN** API key is saved for a profile
- **WHEN** persistence operation completes
- **THEN** plaintext secret SHALL not be persisted in SQLite tables

### Requirement SECURITY-002: License verification SHALL function offline
Cabinet SHALL verify license state without requiring cloud access.

#### Scenario: Offline license check
- **GIVEN** runtime is offline with existing license
- **WHEN** license validation executes
- **THEN** license SHALL validate through local verification path

### Requirement SECURITY-003: Browser requests SHALL respect runtime host and origin boundaries
Cabinet SHALL reject browser-originated API requests that do not target a trusted runtime host and same origin while preserving approved CLI/MCP and self-authenticated Browser Companion transports.

#### Scenario: Cross-site loopback mutation fails closed
- **GIVEN** a foreign website sends a state-changing request to a Cabinet loopback API
- **WHEN** the request carries a foreign `Origin`, cross-site fetch metadata, an untrusted `Host`, or a simple `text/plain` browser body
- **THEN** Cabinet MUST reject it before the route handler runs
- **AND** no record or configuration mutation may be persisted

#### Scenario: Approved clients remain usable
- **GIVEN** the request comes from the same-origin Cabinet UI, an origin-less CLI/MCP client targeting the configured runtime host, or a Browser Companion endpoint with its own credential and exact-origin policy
- **WHEN** the request otherwise satisfies authentication and route requirements
- **THEN** the global request boundary MUST allow the request to reach its existing authorization handler

