# 08 Scanner Engine

## Feature
- Name: Scanner Engine
- ID: 08
- Status: draft

## Goal
- Run query sets against providers to collect candidate listings.
- Capture stock availability/count signals when provider pages expose them.

## Scope
- In scope: Query-set management, run now/scheduled run, candidate persistence, failures.
- Out of scope: Non-v1 enhancements unless explicitly listed in feature backlog.

## User Stories
- As a collector, I need scanner engine to be reliable and easy to use.

## Functional Requirements
- FR-1: Deliver all capabilities listed in docs/FULL_FEATURE_LIST.md for this feature area.
- FR-2: Keep behavior local-first and profile-aware where relevant.
- FR-3: Provide deterministic errors and recovery guidance for failure conditions.

## API and Integration Touchpoints
- Endpoints/services: /api/scanner/*, /api/provider/health
- External providers: eBay/OpenAI/OS features as applicable.

## Data Model Touchpoints
- Tables/entities: scanner_query_sets, scanner_candidates, scanner_failures, provider_health.
- Candidate stock fields: `stock_status`, `stock_count`, `stock_observed_at`, `stock_source_note`.
- Settings/secrets: Profile settings and secret storage where applicable.

## Stock Signal Requirements
- SR-1: Scanner candidate ingest attempts to read stock status from provider payload/page.
- SR-2: If quantity exists, persist normalized integer `stock_count`.
- SR-3: If stock signal is unavailable, persist `stock_status=unknown` (never guess).
- SR-4: Preserve last observed stock timestamp for buy-timing analysis.
- SR-5: Candidate list surfaces stock state inline for triage decisions.

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
