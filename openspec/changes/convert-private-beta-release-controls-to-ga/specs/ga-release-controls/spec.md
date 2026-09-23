## ADDED Requirements

### Requirement: Exact GA artifact identity
Cabinet SHALL generate GA artifacts only when Cabinet and Browser Companion
manifests bind the exact source commit to version `1.0.0`, channel `ga`, and a
non-published GA candidate state.

#### Scenario: Beta artifact is supplied to the GA publisher
- **WHEN** a candidate manifest has a private-beta channel or version
- **THEN** publication SHALL fail before a release or tag is created

### Requirement: Exact GA publication approval
Cabinet SHALL create a non-prerelease GA release only after a successful exact
GA candidate and a trusted `APPROVE CABINET 1.0 GA <sha>` approval comment.

#### Scenario: Exact approval and candidate are present
- **WHEN** the approved SHA matches a successful retained GA candidate
- **THEN** the publisher SHALL create immutable `v1.0.0` assets for that SHA
