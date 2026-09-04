## ADDED Requirements

### Requirement: Cabinet SHALL define reproducible isolated container deployments
Cabinet SHALL define one canonical runtime service and authoritative local,
demo and production deployment plans.

#### Scenario: Resolve an environment deployment
- **WHEN** deployment tooling reads the service catalogue and an environment
  deployment plan
- **THEN** every service, Compose path, dependency and layer reference SHALL
  resolve
- **AND** the active Cabinet deployment SHALL appear in exactly one layer

#### Scenario: Isolate environment state
- **WHEN** local, demo and production deployments are configured
- **THEN** each deployment SHALL use a distinct profile, instance name and
  persistent data volume
- **AND** no environment SHALL mount another environment's data

### Requirement: SQLite-backed deployments SHALL run one non-overlapping replica
Cabinet SHALL run exactly one active container against each persistent SQLite
workspace.

#### Scenario: Replace a running deployment
- **WHEN** Coolify upgrades or rolls back demo or production
- **THEN** overlapping zero-downtime replacement SHALL be disabled
- **AND** the old container SHALL stop before the replacement writes to the
  shared data volume

### Requirement: Remote deployments SHALL use immutable and attributable images
Demo and production SHALL use an immutable image digest with runtime revision
evidence.

#### Scenario: Verify a deployed image
- **WHEN** an operator checks a deployed demo or production resource
- **THEN** `/healthz` SHALL return `200`
- **AND** `/api/runtime` SHALL report version and build metadata matching the
  locked image source revision

### Requirement: Remote local-device workspaces SHALL require an access boundary
Cabinet SHALL not represent local-device mode as remote user authentication.

#### Scenario: Expose demo or production through Coolify
- **WHEN** a remote Cabinet route is configured
- **THEN** it SHALL remain behind an approved remote identity or edge access
  gate
- **AND** the Compose deployment SHALL not publish the application port
  directly on the host

### Requirement: Container deployments SHALL consume shared identity by reference
Cabinet SHALL consume the approved shared ZITADEL capability without deploying
a duplicate identity service.

#### Scenario: Configure environment identity
- **WHEN** local, demo or production application authentication is provisioned
- **THEN** the deployment plan SHALL reference a pinned shared identity
  contract and environment-specific Cabinet application configuration
- **AND** the configuration SHALL define separate redirect URIs, audiences and
  Cabinet roles for that environment
