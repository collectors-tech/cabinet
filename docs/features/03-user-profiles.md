# 03 User Profiles

## Feature
- Name: User Profiles
- ID: 03
- Status: draft

## Goal
- Support multiple local profiles with isolated preferences and credentials.

## Scope
- In scope: Create/list/activate profile, per-profile settings/secrets/license.
- Out of scope: Non-v1 enhancements unless explicitly listed in feature backlog.

## User Stories
- As a collector, I need user profiles to be reliable and easy to use.

## Functional Requirements
- FR-1: Deliver all capabilities listed in docs/FULL_FEATURE_LIST.md for this feature area.
- FR-2: Keep behavior local-first and profile-aware where relevant.
- FR-3: Provide deterministic errors and recovery guidance for failure conditions.

## API and Integration Touchpoints
- Endpoints/services: /api/profiles*
- External providers: eBay/OpenAI/OS features as applicable.

## Data Model Touchpoints
- Tables/entities: profiles, profile_settings, profile_secrets, profile_licenses.
- Settings/secrets: Profile settings and secret storage where applicable.

## UX Flow
- Entry point: Main application workspace and related navigation section.
- Primary path: Happy-path action from list -> details/actions -> confirmation.
- Failure/recovery path: Show actionable message, allow retry, preserve user progress.

## Acceptance Criteria
- [ ] AC-1: Feature capabilities map to full feature list requirements.
- [ ] AC-2: API and persistence behavior are covered by tests.
- [ ] AC-3: UX is consistent with onboarding-first strategy.

## Test Strategy
- Unit: domain/service logic and validation.
- API: handler contract tests for success/error paths.
- E2E: primary user flow plus at least one failure/retry path.

## Dependencies
- Internal: Shared UI components, auth/profile context, persistence layer.
- External: Provider credentials/services where feature requires integration.

## Risks
- Risk: Scope creep and inconsistent behavior across feature screens.
- Mitigation: enforce acceptance criteria and shared component patterns.

## Telemetry and Diagnostics
- Events/logs: activity log entries for major state transitions and failures.
- Error signals: structured error codes for UI handling and diagnostics export.

## Open Questions
- Q1: Which v1.0 cutoff items remain mandatory vs. deferred for this feature?
