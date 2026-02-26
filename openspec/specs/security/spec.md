## Purpose
Define security and privacy controls for secrets and offline trust.

## Requirements
### Requirement: Secrets SHALL never be stored in plaintext SQLite records
Cabinet SHALL store sensitive keys in OS-backed secure storage.

#### Scenario: Secret persistence
- **GIVEN** API key is saved for a profile
- **WHEN** persistence operation completes
- **THEN** plaintext secret SHALL not be persisted in SQLite tables

### Requirement: License verification SHALL function offline
Cabinet SHALL verify license state without requiring cloud access.

#### Scenario: Offline license check
- **GIVEN** runtime is offline with existing license
- **WHEN** license validation executes
- **THEN** license SHALL validate through local verification path

