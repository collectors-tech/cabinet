# Assistant / AI Exploration Audit - 2026-06-18

Issue: #497 `[Exploration] Assistant / AI surfaces audit and traceability pass`

Mode: exploratory documentation/traceability slice. No product runtime code changed.

## Runtime Preconditions

- Work branch: `issue-497-assistant-ai-exploration`.
- Demo lane: demo2-helper at `http://127.0.0.1:17882`.
- `/healthz`: HTTP 200 `ok`.
- `/api/runtime`: `app_version=rev-a74e5e9a5c26`, `runtime_port=17882`, pid `5856`.
- Browser/profile: OpenClaw `project-cabinet` profile was running and browser doctor passed through CDP port `18801`.
- Browser limitation: OpenClaw browser `open` for `http://127.0.0.1:17882/chats` returned `browser navigation blocked by policy`. This pass therefore uses repo/runtime/spec/test evidence and does not claim a fresh live element-by-element browser validation.

Evidence logs for this run are under `.work-agent/logs/issue-497-assistant-ai-exploration/20260618-0400/`.

## Scope Reviewed

Section:
- Assistant / AI surfaces.

Screens/routes:
- Shell Assistant workspace side panel, implemented by `ui.web/src/components/layout/assistant-workspace-panel.tsx`.
- Full `/chats` route, implemented by `ui.web/src/routes/_authenticated/chats/index.tsx` and `ui.web/src/features/chats/index.tsx`.
- Cabinet assistant-ui adapter primitives in `ui.web/src/features/chats/assistant-ui-adapter.tsx`.
- Assistant execution, AI assist, assistant inbox handoff, and provider readiness contracts in `openspec/specs/chats/*`.
- OpenAI provider readiness surfaces in Integrations insofar as they gate Assistant capability claims.

Component/feature surfaces inventoried from source/spec/test evidence:
- Shell workspace Assistant toggle and compact side-panel modal.
- Assistant thread selector, new/reset thread actions, route/profile/selection context chips, provider/model selectors, message list, composer, navigation action card, preview/apply card, confirmation dialog, permission boundary, execution state, and result link.
- Full `/chats` thread rail, search, thread creation, thread rows, assistant-ui message canvas, empty states, top actions, composer, prompt chips, attachment control, model/default row, action preview controls, apply/cancel confirmation, apply result, and error states.
- Assistant capability registry, workflow run records, preview/apply/cancel APIs, AI test/suggest/apply APIs, and OpenAI provider setup/readiness gates.

## Findings And Changes

### Finding 1: Remaining assistant execution traceability rows had no current open follow-up

Status: tracked by new issue #1337.

Expected current contract:
- Every non-implemented Assistant execution traceability row has a current open issue, blocker reason, or closure path that names the missing contract and verification target.

Observed artifact:
- `ASSISTANT-EXECUTION-006` remains `partial`.
- `ASSISTANT-EXECUTION-008` remains `planned`.
- The rows reference older assistant workflow planning/delivery issues (#847, #856, #857, #858), but those issues are closed.

Action taken:
- Opened #1337 `[Assistant] Close remaining assistant execution traceability rows`.
- #1337 asks for either executable implementation evidence and status closure, or explicit blocker reasoning with linked validation targets.

No runtime code or OpenSpec requirement text changed in this branch. The durable change for #497 is this audit artifact plus the linked follow-up issue.

## Screen And Component Coverage

### Shell Assistant workspace side panel

Classification: covered by existing spec/tests for current non-live-provider scope.

Requirement anchors:
- `ASSISTANT-WORKSPACE-001`, `002`, `003`, `004`, `005`, `007`
- `UI-SCREEN-CHAT-COPILOT-010`, `011`, `018`

Evidence:
- Source renders assistant-ui anchored modal, thread selector, provider/model selectors, route/profile/selection context, Cabinet-governed composer, preview/apply controls, permission boundary, and confirm-before-apply dialog.
- Cypress `ui.web/cypress/e2e/chats/assistant-workspace/spec.cy.ts` covers thread continuity, route/profile/selection context, provider/model fork semantics, reset boundaries, chat selection/new-chat/navigation action, assistant-ui adapter primitives, and governed action result persistence proof.
- Go chat APIs persist threads/messages/action previews and enforce confirmation boundaries in `internal/app/chat_api_test.go` and `internal/chat/service_test.go`.

Residual limit:
- Fresh live browser element validation was blocked by isolated browser policy.
- No live OpenAI provider call is claimed.

### Full `/chats` workspace

Classification: covered by existing spec/tests for current non-live-provider scope.

Requirement anchors:
- `CHAT-COPILOT-001`, `002`, `003`, `004`
- `CHATS-WORKSPACE-001` through `006`
- `UI-SCREEN-CHAT-COPILOT-001` through `019`

Evidence:
- Source renders a dedicated dark chat workspace with thread rail, assistant-ui message primitives, bottom-centered composer, prompt chips, model/default row, attachments, and explicit preview/apply/cancel controls.
- Cypress `ui.web/cypress/e2e/chats/ui-screen-chat-copilot/spec.cy.ts` covers thread/message loading, unavailable bootstrap gating, assistant-ui full route primitives, dark shell structure, preview/apply/cancel, provider defaults, collection assignment, failed mutation states, thread context reset, pending preview restore, mobile image attachment, and `/inbox` route reachability.
- Cypress `ui.web/cypress/e2e/chats/chats-workspace/spec.cy.ts` covers route semantics, active-thread preservation, Assistant versus Chats boundary copy, two-pane layout parity, assistant-ui visual contract, filtering, and new-thread stability.
- Cypress `ui.web/cypress/e2e/chats/ui-screen-chat-copilot/header-trigger-icon-only.cy.ts` covers the icon-only header trigger contract.

Residual limit:
- This run did not execute Cypress because no product behavior changed and the selected mode is exploration/reporting.
- Live browser element validation was blocked by policy.

### Assistant action execution and capability registry

Classification: mostly covered; residual traceability gap now tracked by #1337.

Requirement anchors:
- `ASSISTANT-EXECUTION-001` through `009`
- `AI-ASSIST-001` through `004`

Evidence:
- `internal/app/assistant_capabilities.go` exposes governed capabilities with read-only, preview-only, confirm-required, setup-needed, and unavailable states.
- `internal/app/chat_api_test.go` covers capability discovery, content/listing generation preview-first behavior, workflow run persistence, image capability source/audit preservation, and Cabinet Agent app-control tools.
- `internal/app/ai_api_test.go` covers AI apply confirmation-required behavior.
- `internal/app/traceability_wave7_ai_settings_lookup_test.go` covers missing-key, suggest, toggle, and profile-scoped AI settings behavior.
- Cypress `ui.web/cypress/e2e/chats/assistant-execution-surfaces/spec.cy.ts` covers preview-before-apply, confirmation, lifecycle state, and visible permission guidance.

Residual tracked work:
- #1337 covers closure of `ASSISTANT-EXECUTION-006` and `ASSISTANT-EXECUTION-008` traceability state.

### Assistant Inbox handoff

Classification: covered by existing spec/tests.

Requirement anchors:
- `ASSISTANT-INBOX-001`, `002`, `003`
- related Inbox notification card requirements.

Evidence:
- Cypress `ui.web/cypress/e2e/chats/assistant-inbox-handoff/spec.cy.ts` covers persisted assistant handoff items and reopening linked threads from Inbox.
- Cypress `ui.web/cypress/e2e/chats/inbox-notification-cards/spec.cy.ts` covers Telegram capture review links back into Chats with requested thread and preview context.
- `internal/app/chat_api_test.go` covers chat inbox item lifecycle.

Residual limit:
- No external Telegram/OpenAI provider event was executed in this pass.

### OpenAI readiness and provider/model defaults

Classification: covered for setup-needed/non-live-provider behavior; live provider readiness remains outside this pass.

Requirement anchors:
- `AI-ASSIST-001`, `002`, `004`
- `ASSISTANT-EXECUTION-006`
- provider OpenAI integration contracts.

Evidence:
- Provider defaults are loaded from profile settings and shown/preserved in Assistant and `/chats` preview contexts.
- `internal/app/provider_openai_config_api_test.go` covers OpenAI registry setup-needed state, persisted active method, and profile-scoped health readiness.
- Cypress `ui.web/cypress/e2e/integrations/provider-openai-chatgpt-ux/spec.cy.ts` covers Browser Auth/API-key setup-needed and readiness UX.

Residual limit:
- No live OpenAI request, Browser Auth proof, or real provider credential validation is claimed by this audit.

## Scenario Matrix

| Scenario class | Result | Evidence |
| --- | --- | --- |
| Happy path Assistant side-panel render | Covered by tests, not freshly live-validated | `assistant-workspace/spec.cy.ts`; browser navigation blocked in this run. |
| Happy path `/chats` render | Covered by tests, not freshly live-validated | `ui-screen-chat-copilot/spec.cy.ts`; `chats-workspace/spec.cy.ts`. |
| Thread selection and new-thread flow | Covered by tests | Assistant workspace and Chats Cypress suites. |
| Message send and reload | Covered by tests | Cypress request assertions and `internal/app/chat_api_test.go`. |
| Provider/model defaults and fork semantics | Covered by tests | Assistant workspace Cypress; profile settings API tests. |
| Route/profile/selection context | Covered by tests | Assistant workspace Cypress request assertions. |
| Attachment path | Covered by tests | Chat copilot Cypress and chat API tests. |
| Preview-before-apply | Covered by tests | Chat copilot, Assistant workspace, assistant execution Cypress; chat service/API tests. |
| Apply confirmation | Covered by tests | UI confirm dialogs and backend confirmation-required tests. |
| Cancel/abort path | Covered by tests | Chat copilot cancel-preview Cypress and service tests. |
| Invalid/stale thread/profile apply | Covered by tests | Chat service tests for cross-profile/cross-thread rejection. |
| Empty state | Covered by tests | `/chats` empty workspace/thread Cypress. |
| Loading/error/retry state | Covered by tests | Bootstrap/message error handling and unavailable bootstrap Cypress. |
| Keyboard/accessibility path | Partially covered | Header icon accessible name and labeled controls are covered; live keyboard traversal was blocked by browser policy. |
| Refresh/back-forward/persistence | Covered by tests | Pending preview restore after route return/reload Cypress. |
| Live OpenAI provider action | Not claimed | Requires verified provider credentials/capability evidence; tracked by readiness/spec rows and #1337. |

## Element Inventory And Classification

| Element/surface | Classification | Notes |
| --- | --- | --- |
| Shell Assistant workspace toggle/panel | Covered | Assistant workspace Cypress covers open/route continuity. |
| Assistant thread selector | Covered | Existing chat selection and thread switching tests. |
| New/reset thread controls | Covered | Reset boundary and new-chat Cypress coverage. |
| Route/profile/selection chips | Covered | Request envelope assertions verify deterministic context. |
| Provider/model selects | Covered | Provider/model changes fork thread with metadata. |
| Assistant message list | Covered | assistant-ui adapter renders Cabinet persisted messages. |
| Assistant composer | Covered | Sends through `/api/chat/messages` with Cabinet context. |
| Navigation action card | Covered | Layout prompt exposes explicit screen-opening action. |
| Agent action preview card | Covered | Preview remains non-mutating until confirm. |
| Apply confirmation dialog | Covered | Confirm/cancel flows covered. |
| Permission boundary text | Covered | Assistant execution Cypress checks explicit guidance. |
| Apply result/result link | Covered | Assistant workspace persistence proof and result link coverage. |
| `/chats` thread rail/search/new thread | Covered | Chats workspace Cypress. |
| `/chats` dark assistant-ui canvas | Covered | `UI-SCREEN-CHAT-COPILOT-019` and `CHATS-WORKSPACE-006`. |
| `/chats` prompt chips/model row/composer | Covered | Chat copilot and chats workspace Cypress. |
| Attachment upload controls | Covered | Chat copilot Cypress and chat API tests. |
| Action mode/target fields | Covered | Create/update/wishlist/collection preview tests. |
| Failed mutation states | Covered | Missing target and stale apply tests. |
| Assistant Inbox handoff rows | Covered | Assistant inbox handoff and notification card Cypress. |
| Capability registry | Covered, with residual rows tracked | Implemented registry tests pass; #1337 tracks remaining non-implemented rows. |
| OpenAI live readiness | Tracked residual | Setup-needed UX covered; live provider proof not claimed. |

## Follow-Up Disposition

New product/traceability issue opened:
- #1337 `[Assistant] Close remaining assistant execution traceability rows`.

Spec/traceability changes made in this branch:
- None beyond this migration audit artifact. The residual traceability issue is deliberately tracked in #1337 because it needs a focused implementation/closure decision rather than a silent status edit.

No untracked broken product behavior remains from the repo-driven portion of this audited scope. Fresh live element-level validation remains blocked by OpenClaw browser policy in this isolated cron context.

## Completeness

Completeness label: Complete with browser-policy limitation and one tracked traceability follow-up.

The route, screen, and component inventory was reconciled against source, specs, traceability, issue state, and focused tests. A residual non-implemented traceability gap was converted into #1337. Fresh live browser validation remains blocked and is not claimed.

## Next Recommendation

After #497 is reviewed/merged and demo2 is recycled from `develop`, continue route-ordered exploration at #498 Inbox / communications surfaces unless a higher-priority Cabinet issue is added. Prioritize #1337 when the queue returns to Assistant execution traceability closure.
