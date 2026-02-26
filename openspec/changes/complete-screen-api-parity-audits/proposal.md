## Why

Several key screens still risk partial wiring, fallback sample behavior, or mismatched data contracts. Completing API parity audits prevents disconnected UI flows and ensures screens are production-backed.

## What Changes

- Define parity completion requirements for Users, Settings, and Chat screens.
- Enforce route-by-route API mapping checks and test coverage expectations.
- Remove tolerance for sample/static placeholder behavior in production paths.
- Add closure evidence standards for parity issues.

## Capabilities

### New Capabilities

- `screen-api-parity-audits`: Capability to define and validate route-level UI/API parity requirements.
- `users-settings-chat-wiring`: Capability ensuring Users, Settings, and Chat are fully backed by Cabinet APIs.

### Modified Capabilities

- None.

## Impact

- Affected code: `ui.web/src/routes/_authenticated/*`, feature modules for users/settings/chats, API client layer.
- Affected tests: E2E and API contract checks for parity routes.
- Related issues: `#143`, `#144`, `#145`, content cleanup `#152`.
