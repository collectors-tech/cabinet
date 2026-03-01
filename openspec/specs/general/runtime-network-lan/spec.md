## Purpose
Define local network access behavior for Cabinet runtime.

## Requirements
### Requirement RUNTIME-NETWORK-LAN-001: Runtime SHALL support configurable LAN bind mode
Cabinet SHALL support local-only and LAN-access bind modes.

#### Scenario: Enable LAN bind
- **GIVEN** runtime config sets `bind_mode=lan` and `host=0.0.0.0`
- **WHEN** Cabinet starts runtime server
- **THEN** HTTP server MUST listen on configured LAN interface and expose health endpoint to same-network clients

### Requirement RUNTIME-NETWORK-LAN-002: LAN mode SHALL retain auth protections
Cabinet SHALL enforce the same authentication and authorization policy for LAN clients as local clients.

#### Scenario: Unauthorized LAN request
- **GIVEN** runtime is bound in LAN mode and request is made without valid session
- **WHEN** client calls protected API endpoint
- **THEN** API MUST return `401` or `403` and MUST NOT mutate profile data
