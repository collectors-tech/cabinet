# Inbox / Communications Exploration Audit - 2026-06-18

Issue: #498 `[Exploration] Inbox / communications audit and traceability pass`

Mode: exploratory documentation/traceability slice. No product runtime code changed.

## Runtime Preconditions

- Work branch: `issue-498-inbox-communications-exploration`.
- Demo lane: demo2-helper at `http://127.0.0.1:17882`.
- `/healthz`: HTTP 200 `ok`.
- `/api/runtime`: `app_version=rev-365cf0800918`, `runtime_port=17882`, pid `67720`.
- Local repo: clean `develop` at `365cf080091809b61cfcec1fa7d14d73d72ef9b1` before branch creation.
- Browser/profile: OpenClaw `project-cabinet` profile doctor passed through CDP port `18801`.
- Browser limitation: OpenClaw browser `open` for `http://127.0.0.1:17882/inbox` returned `browser navigation blocked by policy`. This pass therefore uses runtime health plus repo/source/spec/test evidence and does not claim fresh live element-by-element browser interaction.

Evidence logs for this run are under `.work-agent/logs/issue-498-inbox-communications-exploration/20260618-0500/`.

## Scope Reviewed

Section:
- Inbox / communications surfaces.

Screens/routes:
- First-class Notification Inbox route `/inbox`, implemented by `ui.web/src/routes/_authenticated/inbox/index.tsx` and `ui.web/src/features/notifications/index.tsx`.
- Shell Inbox workspace panel, implemented by `ui.web/src/components/layout/inbox-workspace-panel.tsx`.
- Chats route handoff target for assistant and Telegram capture review links, implemented by `ui.web/src/features/chats/index.tsx`.
- Purchases route forwarding package inbox/reconciliation surfaces, implemented by `ui.web/src/features/purchases/index.tsx`.
- Notification settings route as a communication-preference adjacent surface, implemented by `ui.web/src/features/settings/notifications/notifications-form.tsx`.

Component/feature surfaces inventoried from source/spec/test evidence:
- Notification Inbox header, refresh action, compact Inbox opener, summary count cards, All/Unread/Assistant/System filters, select visible control, bulk read/archive actions, loading state, contextual empty states, retryable error state, notification rows, read/unread/archive row actions, detail expander, linked target URL rendering, assistant/system category derivation, and profile-scoped `/api/chat/inbox` loading/updating.
- Shell Inbox workspace card list, refresh action, loading/empty/error states, read/unread/archive card triage, item links, Telegram capture review links, Open Chats empty-state action, Open Assistant Workspace action, and Open in Assistant action that seeds thread/provider/model local storage before switching workspaces.
- Assistant inbox handoff API path that persists assistant handoff items and companion assistant messages through `/api/chat/messages`.
- Forwarder package inbox manual import, CSV import, email import, list refresh, package detail, link/reconciliation form, confirm/override/unlink controls, non-mutating match suggestions, confidence filters, scoped package suggestions, audit-event rendering, review-state filters, and profile-scoped forwarding APIs.
- Notification settings controls for notification type and communication/social/marketing/security/mobile preferences.

## Findings And Changes

No new product defects or new spec gaps were found in the reviewed no-browser-interaction evidence set.

The durable change for #498 is this audit artifact. It records that the current Inbox / communications surfaces are already covered by focused OpenSpec requirements and Cypress/Go tests for the non-live-provider scope. The report also records the isolated-browser limitation so #498 is not treated as fresh live UI proof.

## Screen And Component Coverage

### First-Class Notification Inbox Route

Route: `/inbox`

Classification: covered by existing spec/tests for the current notification-queue contract.

Requirement anchors:
- `UI-SCREEN-NOTIFICATION-INBOX-001`
- `UI-SCREEN-NOTIFICATION-INBOX-002`
- `UI-SCREEN-NOTIFICATION-INBOX-003`
- `UI-SCREEN-NOTIFICATION-INBOX-004`
- `UI-SCREEN-NOTIFICATION-INBOX-005`
- `UI-SCREEN-NOTIFICATION-INBOX-006`

Evidence:
- `ui.web/src/routes/_authenticated/inbox/index.tsx` binds `/inbox` to `NotificationInbox`, not Purchases.
- `ui.web/src/features/notifications/index.tsx` implements counts, filters, row detail, target links, row triage, bulk triage, refresh, loading, empty, and retryable error behavior.
- `ui.web/cypress/e2e/chats/notification-inbox/spec.cy.ts` covers first-class route rendering, category filters, contextual empty states, detail expansion, linked target navigation, row/bulk triage, retryable API failure, and failed-update row preservation.
- `openspec/traceability.md` maps `UI-SCREEN-NOTIFICATION-INBOX-001` through `006` to the Cypress coverage above.

Element inventory and classification:
- Header title and page icon: covered by `UI-SCREEN-NOTIFICATION-INBOX-001`.
- Refresh action: covered by `UI-SCREEN-NOTIFICATION-INBOX-005`.
- Compact Inbox opener: source verified; already part of shell workspace integration, no separate defect found.
- Count cards for All/Unread/Assistant/System: covered by `UI-SCREEN-NOTIFICATION-INBOX-002`.
- Filter tabs: covered by `UI-SCREEN-NOTIFICATION-INBOX-002`.
- Select visible, bulk mark read, bulk archive: covered by `UI-SCREEN-NOTIFICATION-INBOX-004`.
- Loading state, empty states, error alert, retry action: covered by `UI-SCREEN-NOTIFICATION-INBOX-005`.
- Row checkbox, title, status badge, category badge, source, timestamp, linked target, read/unread, archive, details: covered by `UI-SCREEN-NOTIFICATION-INBOX-003`, `004`, and `006`.

Scenario status:
- Happy path: Validated by Cypress/source evidence for route render, list display, filters, detail expansion, target links, and triage.
- Cancellation/abort path: Not applicable; no modal or destructive confirm flow exists in the Notification Inbox route.
- Invalid input path: Not applicable for this route because users do not submit free-form input.
- Empty state: Validated by `UI-SCREEN-NOTIFICATION-INBOX-002` and `005` coverage.
- Loading state: Validated by `UI-SCREEN-NOTIFICATION-INBOX-005` coverage.
- Error/failure state: Validated by `UI-SCREEN-NOTIFICATION-INBOX-005` and `006` coverage.
- Keyboard-only path: Covered indirectly by native button/link/checkbox/tab semantics in the component and Cypress visible-control coverage; no fresh manual keyboard pass is claimed.
- Refresh/back-forward persistence path: Refresh action covered; browser back/forward was not freshly exercised in this run due browser policy.
- Post-action verification path: Validated by row/bulk update assertions and failed-update preservation coverage.
- Permission/role variance: Not applicable beyond authenticated active-profile scoping for this route.

Residual limit:
- Fresh live browser interaction was blocked by policy, so this pass does not independently re-execute `/inbox` in the browser.

### Shell Inbox Workspace

Surface: authenticated shell workspace panel.

Classification: covered by existing spec/tests for catch-up card and assistant handoff routing behavior.

Requirement anchors:
- `UI-SCREEN-INBOX-NOTIFICATION-CARDS-001`
- `UI-SCREEN-INBOX-NOTIFICATION-CARDS-002`
- `UI-SCREEN-INBOX-NOTIFICATION-CARDS-003`
- `UI-SCREEN-INBOX-NOTIFICATION-CARDS-004`
- `ASSISTANT-INBOX-001`
- `ASSISTANT-INBOX-002`
- `ASSISTANT-INBOX-003`

Evidence:
- `ui.web/src/components/layout/inbox-workspace-panel.tsx` loads `/api/chat/inbox`, filters archived items out of the catch-up list, renders card state/source/age/target links, and supports read/unread/archive updates through `/api/chat/inbox/:id`.
- Telegram capture review links preserve `thread_id` and `preview_id` as `/chats` query state.
- `Open in Assistant` stores the selected assistant thread/provider/model for the active profile before switching the shell workspace to Assistant.
- `ui.web/cypress/e2e/chats/inbox-notification-cards/spec.cy.ts` covers catch-up cards, read/unread/archive triage, Telegram capture review URLs, and failed update preservation.
- `internal/app/chat_api_test.go` covers assistant handoff item creation and inbox status lifecycle.

Element inventory and classification:
- Shell Inbox intro card: covered by source and workspace Cypress entry coverage.
- Refresh button: covered by load/error path in source and card Cypress setup.
- Loading text and empty-state Open Chats / Open Assistant Workspace actions: source verified; no defect found.
- Notification cards, status badges, source/age text, item/review links: covered by `UI-SCREEN-INBOX-NOTIFICATION-CARDS-001` and `003`.
- Mark read, mark unread, archive: covered by `UI-SCREEN-INBOX-NOTIFICATION-CARDS-002` and `004`.
- Open in Assistant: source verified for thread/provider/model workspace handoff; assistant handoff behavior is covered by `ASSISTANT-INBOX-*` backend and chat workspace tests.
- Error message: covered by `UI-SCREEN-INBOX-NOTIFICATION-CARDS-004`.

Scenario status:
- Happy path: Validated by Cypress/source evidence for visible cards, Telegram review links, item links, and assistant workspace handoff setup.
- Cancellation/abort path: Not applicable; no modal confirmation is used in shell Inbox.
- Invalid input path: Not applicable; users do not submit free-form input.
- Empty state: Source verified for empty copy and Open Chats/Open Assistant Workspace actions.
- Loading state: Source verified.
- Error/failure state: Validated by failed-update Cypress coverage.
- Keyboard-only path: Covered indirectly by native button/link semantics; no fresh manual keyboard pass is claimed.
- Refresh/back-forward persistence path: Shell workspace state is covered by broader shell workspace tests, but not freshly exercised here due browser policy.
- Post-action verification path: Validated by Cypress assertions that status changes and archived cards leave the visible list.
- Permission/role variance: Not applicable beyond active-profile scoping in request payloads.

Residual limit:
- Fresh live shell workspace interaction was blocked by browser policy.

### Chats / Assistant Handoff Target

Route: `/chats`

Classification: covered by existing chat and assistant-inbox traceability for inbox-originating links.

Requirement anchors:
- `ASSISTANT-INBOX-001`
- `ASSISTANT-INBOX-002`
- `ASSISTANT-INBOX-003`
- `CHAT-COPILOT-*`
- `UI-SCREEN-CHAT-COPILOT-*`

Evidence:
- `/api/chat/messages` can create assistant handoff inbox items with linked thread metadata.
- `ui.web/cypress/e2e/chats/inbox-notification-cards/spec.cy.ts` proves Telegram capture review URLs open `/chats` with the requested thread selected.
- Chat workspace and assistant workspace specs cover thread/message rendering, assistant workflow previews, and governed apply/cancel behavior.

Scenario status:
- Happy path: Validated by targeted Cypress and Go tests for thread selection and assistant handoff item creation.
- Error/failure state: Covered in chat/assistant surface-specific tests, not re-executed in this #498 report.
- Respond flow: Current product contract routes users to Chats/Assistant for response/continuation instead of embedding reply composition directly in Inbox. No defect is filed because the linked continuation path is the implemented and traced contract.

Residual limit:
- No live OpenAI provider or Telegram external runtime success is claimed.

### Forwarder Package Inbox / Purchase Source Matches

Route: `/purchases`, expandable Purchase Source Matches / forwarder package inbox panel.

Classification: covered by existing spec/tests for forwarding package inbox import and reconciliation workflows.

Requirement anchors:
- `INTEGRATION-029` through `INTEGRATION-062`
- `EBAY-PURCHASE-CAPTURE-*` where purchase capture review intersects inbox intake.

Evidence:
- `ui.web/src/features/purchases/index.tsx` implements the forwarder package inbox panel, package import/listing, CSV/email import, detail view, link/reconciliation forms, non-mutating suggestions, confidence filters, scoped package suggestion requests, audit events, and review-state summaries.
- `openspec/specs/integrations/forwarder-package-inbox/spec.md` defines the normalized import, API, UI, reconciliation, suggestion, filtering, OpenAPI, and active-profile contracts.
- `ui.web/cypress/e2e/purchases/purchase-inbox/spec.cy.ts` covers the Purchases route, purchase source matches expander, forwarding package import/list/reconciliation surfaces, suggestions, confidence filtering, and audit-event rendering.
- Go tests in `internal/forwarding/package_inbox_test.go`, `internal/app/forwarder_package_api_test.go`, and OpenAPI parity tests cover API normalization, persistence, active-profile isolation, and documented contracts.

Element inventory and classification:
- Purchase Source Matches expander: covered by Cypress.
- Manual package import fields and validation: covered by `INTEGRATION-032`.
- CSV import and row-level errors: covered by `INTEGRATION-033` and `034`.
- Email import and parser/API errors: covered by `INTEGRATION-035` and `036`.
- Package detail, raw provenance, link forms, active link state, confirm/override/unlink, audit events: covered by `INTEGRATION-037` through `044`.
- Review summary, state filter, row-level evidence labels, match suggestions, confidence filters, package-scoped suggestions: covered by `INTEGRATION-045`, `046`, `048`, and `050` through `055`.
- OpenAPI contracts and active-profile isolation: covered by `INTEGRATION-049`, `056`, `057`, `060`, `061`, and `062`.

Scenario status:
- Happy path: Validated by Cypress/Go coverage for manual/CSV/email import, package list refresh, suggestions, and link decisions.
- Cancellation/abort path: Covered where dialogs/forms expose cancel or require explicit confirmation in Purchases tests; no new gap found.
- Invalid input path: Validated by API/UI validation-error tests for imports and links.
- Empty state: Covered by Purchases route and package-panel Cypress coverage.
- Loading state: Source verified; not freshly browser-exercised here.
- Error/failure state: Covered by forwarding API/UI tests for validation and scoped error rendering.
- Keyboard-only path: Not freshly claimed in this policy-blocked run.
- Refresh/back-forward persistence path: Package refresh and route stability are covered by Cypress; browser history was not freshly exercised.
- Post-action verification path: Validated by Cypress/Go assertions that imports refresh the list and link decisions/audit events persist.
- Permission/role variance: Active-profile isolation is covered by forwarding API tests.

Residual limit:
- This pass does not claim live external Stackry/freight-forwarder provider sync.

### Notification Settings

Route: `/settings/notifications`

Classification: adjacent communication-preference surface; covered by settings-specific route/tests, not a primary #498 inbox route.

Requirement anchors:
- `openspec/specs/settings/notifications/spec.md`

Evidence:
- `ui.web/src/features/settings/notifications/notifications-form.tsx` exposes notification type, communication email, marketing email, social email, security email, and mobile notification controls with saved profile settings.
- `ui.web/cypress/e2e/settings/notifications/spec.cy.ts` covers the settings form behavior.

Residual limit:
- #498 does not broaden into settings exploration; #499 remains the settings/preferences exploration issue.

## Traceability Summary

Existing traceability already maps the reviewed Inbox / communications contracts:
- `UI-SCREEN-NOTIFICATION-INBOX-001` through `006` -> `ui.web/cypress/e2e/chats/notification-inbox/spec.cy.ts`.
- `UI-SCREEN-INBOX-NOTIFICATION-CARDS-001` through `004` -> `ui.web/cypress/e2e/chats/inbox-notification-cards/spec.cy.ts`.
- `ASSISTANT-INBOX-*` -> assistant/chat API and UI handoff coverage.
- `INTEGRATION-029` through `062` -> forwarding package Go/API/OpenAPI/Cypress coverage.
- Notification settings -> `ui.web/cypress/e2e/settings/notifications/spec.cy.ts`.

No traceability rows were changed in this run because the current rows already point to focused executable coverage and no new gap was found.

## Follow-Up Recommendation

#498 can close after this report is merged if reviewer accepts repo/source/spec/test evidence plus the recorded browser-policy limitation. The next route-ordered exploration issue is #499 Settings / preferences, unless newer higher-priority product feedback is added.
