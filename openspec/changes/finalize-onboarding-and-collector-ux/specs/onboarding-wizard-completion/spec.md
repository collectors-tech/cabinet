## ADDED Requirements

### Requirement: Onboarding SHALL guide first-time users through a minimal completion path
The system SHALL provide a stepwise onboarding flow that covers identity, starter data choice, first item, and preferences before exposing advanced workspace operations.

#### Scenario: First run onboarding sequence
- **WHEN** a first-time user finishes authentication
- **THEN** the system SHALL present onboarding steps in guided order with clear progress

#### Scenario: Onboarding completion
- **WHEN** all required onboarding steps are completed
- **THEN** the system SHALL unlock advanced workspace navigation and persist completion state

### Requirement: Onboarding state SHALL survive reload and restart
The system SHALL persist onboarding progress and restore the current step after app reload unless onboarding is complete.

#### Scenario: Resume onboarding after restart
- **WHEN** a user closes and reopens the app mid-onboarding
- **THEN** the flow SHALL resume at the last incomplete step
