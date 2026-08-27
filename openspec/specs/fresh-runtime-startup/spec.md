# fresh-runtime-startup Specification

## Purpose
Define the validation-safe startup contract for initializing a fresh Cabinet runtime without avoidable migration overhead.
## Requirements
### Requirement: Fresh runtime startup SHALL remain validation-safe on empty databases
Cabinet SHALL initialize a fresh tempdir runtime database quickly enough that startup-bound validation suites do not fail due to avoidable migration overhead in the normal app startup path.

#### Scenario: Fresh validation startup on empty database
- **GIVEN** Cabinet starts against a new empty SQLite database in a temp data directory
- **WHEN** `app.New()` runs during local validation or CI
- **THEN** schema creation and migration SHALL complete within the normal startup timeout without requiring ad hoc timeout overrides

#### Scenario: Startup NFR gate measures real startup path
- **GIVEN** the startup non-functional gate executes against a fresh tempdir runtime
- **WHEN** the gate measures runtime startup
- **THEN** it SHALL exercise the real startup path and fail only for genuine startup regressions rather than avoidable migration batching overhead

