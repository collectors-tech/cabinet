# Scanner Screen Spec

## Use Cases
1. User creates/edits scanner query sets.
2. User runs scanner now or scheduled.
3. User inspects failures and retries query runs.
4. User checks provider health and matching summary.

## UI Sections
1. Query set editor and list
2. Run controls
3. Failure list + retry
4. Provider health status
5. Matching summary

## State Behavior
- Loading: query sets/failures loading states.
- Empty: no query sets/failures messaging.
- Error: scanner API errors with retry.
- Success: query state + run status messages.

## Acceptance Criteria
- [ ] Run now and scheduled show deterministic status text.
- [ ] Failures include query set id, reason, attempts.
- [ ] Retry is available only when query set id exists.
- [ ] Provider health shows healthy/state indicators.

## Test Cases
- `SCAN-001` create and edit query set.
- `SCAN-002` run now and scheduled status.
- `SCAN-003` failure load and retry.
- `SCAN-004` provider health load.

