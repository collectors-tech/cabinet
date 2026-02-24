# 05 Test Matrix (Strict)

## Purpose
Define mandatory verification for IA, workflows, and attention logic.

## Test Layers
1. Unit/UI tests (Vitest + Testing Library)
2. API tests (Go handlers)
3. E2E tests (Playwright and Cypress where available)

## Matrix
1. IA and navigation
- Validate top-level nav entries and active state.
- Validate mobile drawer open/close and navigation parity.
- Validate guard states (no profile, onboarding incomplete, onboarding complete).

2. Onboarding and auth
- Validate 5-step onboarding happy path.
- Validate identity failure and retry path.
- Validate WebAuthn begin/finish flows.
- Validate onboarding completion unlocks workspace.

3. Home attention cards
- Validate card ordering by severity and ranking logic.
- Validate thresholds for price-change triggers.
- Validate snooze and dismiss persistence.
- Validate card actions route to correct screens.

4. Inventory tabs
- Items CRUD/search/filter states.
- Photos upload/camera/primary/delete/fullscreen.
- Barcodes add/load/lookup/external search.
- AI assist enable/suggest/apply with explicit confirmation.

5. Discover and scanner
- Discover filters and actions.
- Scanner run now/scheduled.
- Failures list and retry.
- Provider health rendering and errors.

6. Reports and settings
- Pricing trend/history/export.
- Backup list/run/restore guarded path.
- License import/status.
- Diagnostics and runtime/recovery status.

## Pass/Fail Gate
1. No critical screen lacks empty/error state tests.
2. No nav destination lacks mount/navigation tests.
3. Attention ranking logic has deterministic tests.
4. End-to-end happy path and one failure path per core area pass.

## Evidence Template
- Command
- Result (pass/fail)
- Test file(s)
- Related issue number

## Acceptance Criteria
- [ ] Matrix covers all v1 IA destinations.
- [ ] Each destination has at least one E2E scenario.
- [ ] Critical decision logic has deterministic unit coverage.

