# 12 Price Tracking

## Feature
- Name: Price Tracking
- ID: 12
- Status: draft

## Goal
- Persist daily price snapshots and provide graph/export views.
- Persist stock snapshots by source so pricing can be combined with availability trends.

## Scope
- In scope: Track item, history, source breakdown, export.
- In scope: Source stock count/state observation history and change detection.
- Out of scope: Non-v1 enhancements unless explicitly listed in feature backlog.

## User Stories
- As a collector, I need price tracking to be reliable and easy to use.

## Functional Requirements
- FR-1: Deliver all capabilities listed in docs/FULL_FEATURE_LIST.md for this feature area.
- FR-2: Keep behavior local-first and profile-aware where relevant.
- FR-3: Provide deterministic errors and recovery guidance for failure conditions.

## API and Integration Touchpoints
- Endpoints/services: /api/pricing/*
- External providers: eBay/OpenAI/OS features as applicable.

## Data Model Touchpoints
- Tables/entities: tracked_items, price_snapshots.
- Additional entities/fields: stock_snapshots (or stock columns on price snapshots), source_stock_events.
- Settings/secrets: Profile settings and secret storage where applicable.

## Stock-Aware Requirements
- SR-1: Snapshot includes `stock_status` and optional `stock_count` per source.
- SR-2: Compute `in_stock_change` event when state transitions between in/out of stock.
- SR-3: Source breakdown UI exposes latest stock state next to latest price.
- SR-4: Export includes stock fields for downstream buy-timing analysis.

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
