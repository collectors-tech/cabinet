## Purpose
Define diagnostics collection, recovery-state visibility, and optional session telemetry integration.

## Requirements
### Requirement DIAGNOSTICS-001: Diagnostics SHALL expose abnormal shutdown recovery state
Cabinet SHALL expose crash/recovery indicators and diagnostics recommendation when runtime detects abnormal previous termination.

#### Scenario: Abnormal shutdown detected
- **GIVEN** runtime starts after abnormal termination
- **WHEN** recovery state is evaluated
- **THEN** diagnostics recommendation SHALL be surfaced with path to logs and recovery actions

### Requirement DIAGNOSTICS-002: Diagnostics telemetry SHALL be explicit opt-in
Cabinet SHALL default diagnostics telemetry to opt-out and require explicit user opt-in before any remote diagnostics session events are sent.

#### Scenario: Diagnostics opt-in control
- **GIVEN** diagnostics telemetry setting is disabled by default
- **WHEN** user enables diagnostics opt-in
- **THEN** runtime SHALL allow remote diagnostics session event upload

### Requirement DIAGNOSTICS-003: Diagnostics provider SHALL support Sentry-compatible event model
Cabinet SHALL support Sentry-compatible error/session envelopes (or equivalent provider abstraction) for remote diagnostics when opt-in is enabled, with the same recursive redaction boundary used for local diagnostics storage and export.

#### Scenario: Sentry-compatible user session diagnostics
- **GIVEN** diagnostics opt-in is enabled and provider configuration is valid
- **WHEN** unhandled error occurs in active user session
- **THEN** diagnostics pipeline SHALL emit session and error events using Sentry-compatible schema
- **AND** the remote envelope SHALL NOT include raw cookies, authorization headers, passwords, tokens, API keys, secrets, raw session identifiers, private page content, or sensitive local paths

### Requirement DIAGNOSTICS-004: Diagnostics SHALL remain local-only when opt-in is disabled
Cabinet SHALL keep diagnostics local when opt-in is disabled, without remote event transmission.

#### Scenario: Opt-in disabled
- **GIVEN** diagnostics opt-in is disabled
- **WHEN** runtime logs errors for a user session
- **THEN** diagnostics data SHALL remain local and no remote telemetry call SHALL be executed

