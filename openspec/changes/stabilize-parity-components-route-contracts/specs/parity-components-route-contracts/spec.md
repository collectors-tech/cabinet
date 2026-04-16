## MODIFIED Requirements

### Requirement: General parity contracts SHALL use canonical authenticated routes
Cabinet SHALL keep general parity tests aligned with the canonical authenticated routes currently exposed by the shell.

#### Scenario: Authenticated home parity uses dashboard route
- **GIVEN** a parity test signs into the authenticated shell
- **WHEN** it validates home/dashboard API contracts
- **THEN** it SHALL use the canonical authenticated home route (`/dashboard`) rather than legacy root-route assumptions

#### Scenario: Settings parity uses profile route
- **GIVEN** a parity or foundation-components test targets settings profile behavior
- **WHEN** it boots the authenticated settings surface
- **THEN** it SHALL use the canonical settings profile route (`/settings/profile`) rather than the legacy `/settings/` entry assumption

### Requirement: Foundation component coverage SHALL assert stable visible settings surfaces
Cabinet SHALL keep foundation component tests anchored to stable visible profile-form controls rather than hidden responsive-only mirrors or localization-sensitive copy when proving the settings profile surface is present.

#### Scenario: Settings component surface uses stable visible profile controls
- **GIVEN** a foundation-components test opens the settings profile screen
- **WHEN** it proves the profile surface is rendered
- **THEN** it SHALL assert stable visible controls such as the username field and update action rather than hidden responsive mirrors or fragile copy-only matches
