## Purpose
Define first-run onboarding and authentication screen flow.

## Requirements
### Requirement: Onboarding/auth screen SHALL enforce WebAuthn-first identity setup
The onboarding/auth screen SHALL require credential registration before advanced workspace access.

#### Scenario: Use case - first-time identity setup
- **WHEN** first-time user authenticates
- **THEN** onboarding SHALL guide user through required identity completion

### Requirement: Onboarding/auth screen SHALL persist and resume progress
Onboarding/auth progress SHALL survive reload/restart until completion.

#### Scenario: Use case - resume onboarding
- **WHEN** user exits mid-onboarding and returns
- **THEN** screen SHALL resume from last incomplete step
