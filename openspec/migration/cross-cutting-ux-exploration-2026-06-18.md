# Cross-Cutting UX Exploration Audit - 2026-06-18

Issue: #501 `[Exploration] Cross-cutting UX audit and traceability pass`

Mode: exploratory documentation/traceability slice. No product runtime code changed.

## Runtime Preconditions

- Work branch: `issue-501-cross-cutting-ux-audit`.
- Demo lane: demo2-helper at `http://127.0.0.1:17882`.
- `/healthz`: HTTP 200 `ok`.
- `/api/runtime`: `app_version=rev-fe29d5a4cd6e`, `runtime_port=17882`, pid `21864`.
- Browser/profile: required OpenClaw profile is `project-cabinet`.
- Browser limitation: recent adjacent exploration runs show the `project-cabinet` profile can pass CDP doctor checks, but isolated cron browser navigation/open/snapshot actions are blocked by policy. This pass therefore uses repo/runtime/spec/test evidence and does not claim a fresh live element-by-element browser validation.

Evidence logs for this run are under `.work-agent/logs/issue-501-cross-cutting-ux-audit/20260618-1640/`.

## Scope Reviewed

Section:
- Cross-cutting UX behavior across authenticated and public Cabinet surfaces.

Behavior families:
- Toasts, inline feedback, retry, and status surfaces.
- Keyboard shortcuts and keyboard-only workflow completion.
- Focus handling, Escape close, and trigger focus restore for modal/dialog flows.
- Row/detail/edit interaction consistency and checkbox-driven bulk selection.
- Accessibility-visible semantics, icon-button naming, landmarks, headings, and responsive overflow.
- Consistency of status/action wording for successful, blocked, unsupported, and failure states.

Primary source/spec/test anchors:
- `openspec/specs/general/ui-foundation-accessibility/spec.md`
- `openspec/specs/general/ui-foundation-interactions/spec.md`
- `openspec/specs/general/ui-foundation-components/spec.md`
- `openspec/specs/general/ui-keyboard-shortcuts/spec.md`
- `openspec/specs/general/ui-global-search-command/spec.md`
- `openspec/specs/general/ui-foundation-auth-menus-shortcuts/spec.md`
- `ui.web/cypress/e2e/general/ui-foundation-accessibility/spec.cy.ts`
- `ui.web/cypress/e2e/general/ui-foundation-interactions/spec.cy.ts`
- `ui.web/cypress/e2e/general/ui-foundation-components/spec.cy.ts`
- `ui.web/cypress/e2e/general/ui-keyboard-shortcuts/spec.cy.ts`
- `ui.web/cypress/e2e/general/ui-global-search-command/spec.cy.ts`
- `ui.web/cypress/e2e/general/accessibility-responsive-landmarks/spec.cy.ts`
- `tests/ui_foundation_components_contract_test.go`
- current `openspec/traceability.md`

## Findings And Changes

### Finding 1: Cross-cutting foundation contracts are already explicit and traceable

Status: covered by existing specs/tests.

Evidence:
- `UI-FOUNDATION-ACCESSIBILITY-001` covers Escape close and focus behavior for modals/drawers.
- `UI-FOUNDATION-ACCESSIBILITY-002` covers keyboard-only execution for core workflows.
- `UI-FOUNDATION-ACCESSIBILITY-003` covers accessible names for action/icon controls.
- `UI-FOUNDATION-ACCESSIBILITY-004` covers single visible main landmark, headings, header chrome, and responsive overflow.
- `UI-FOUNDATION-INTERACTIONS-001..003` cover row details, thumbnail lightbox separation, explicit checkbox bulk selection, and double-click edit separation.
- `UI-FOUNDATION-COMPONENTS-001..005` cover component state contracts, loading/empty/error/ready states, duplicate-submit prevention, keyboard dialog behavior, and testability artifacts.
- `UI-KEYBOARD-SHORTCUTS-001..003` cover platform-aware shortcut notation, sidebar shortcut behavior, and shortcut collision policy.
- `UI-GLOBAL-SEARCH-COMMAND-001..008` cover command palette keyboard open/focus, navigation/theme actions, loading/error states, and actionable barcode fallback.

Change made:
- Added this repo-tracked #501 audit artifact to bind the cross-cutting UX route-audit section to the existing traceability evidence and record the browser limitation honestly.

### Finding 2: Feedback/status behavior is distributed but covered by route-specific tests

Status: covered by existing specs/tests.

Evidence:
- Settings categories, account, appearance, display, notifications, storage, and operations specs/tests cover deterministic save success, load failure retry, save failure preservation, import/export status, and maintenance action feedback.
- Notification Inbox and shell notification-card specs/tests cover triage status, row/bulk actions, and failure recovery without losing visible queue context.
- Discoveries, Market Watch, Scanner, Media, Purchases, Integrations, and Assistant specs/tests cover failure guidance, retry actions, preview/apply confirmation, unsupported/blocked states, and provenance/status wording.

Residual limit:
- This audit did not run every route-specific Cypress suite. It reconciled existing test titles and traceability rows against the cross-cutting behavior families and targeted foundation coverage.

### Finding 3: No focused product issue was opened from this slice

Status: no new issue needed.

Reason:
- No broken or uncovered cross-cutting UX contract was found in the reviewed source/spec/test evidence.
- The only limitation is the isolated cron browser-policy block, which is a tooling/precondition confidence limit already documented in adjacent exploration artifacts, not a Cabinet product defect.

## Screen And Component Coverage

### Public/auth entry and shell controls

Classification: covered by existing spec/tests.

Evidence:
- Sign-in action/icon naming is covered by `UI-FOUNDATION-ACCESSIBILITY-003`.
- Auth menu shortcut labels and unsupported action pruning are covered by `ui-foundation-auth-menus-shortcuts`.
- Responsive shell landmarks and headings are covered by `UI-FOUNDATION-ACCESSIBILITY-004`.
- Sidebar shortcut behavior is covered by `UI-KEYBOARD-SHORTCUTS-002`.

### Global command/search surfaces

Classification: covered by existing spec/tests.

Evidence:
- Ctrl/Cmd+K command palette open/focus, navigation commands, theme actions, local catalog search, loading/error states, barcode lookup loading/error states, and unresolved barcode handoff are covered by `ui-global-search-command`.

### Data tables, rows, and bulk selection

Classification: covered by existing spec/tests.

Evidence:
- Inventory row click versus thumbnail lightbox behavior is covered by `UI-FOUNDATION-INTERACTIONS-001`.
- Explicit checkbox-driven bulk selection is covered by `UI-FOUNDATION-INTERACTIONS-002`.
- Inventory, Users, and Integrations row detail/edit separation is covered by `UI-FOUNDATION-INTERACTIONS-003` and route-specific Cypress suites.

### Dialogs, modals, drawers, and focus behavior

Classification: covered by existing spec/tests.

Evidence:
- Foundation accessibility and component specs cover Escape close and focus-return contracts.
- Inventory create dialog, row details/edit dialogs, Users edit dialogs, Integrations detail/edit/action dialogs, Wishlist modals, Collections side panels/dialogs, Media edit/add dialogs, and Assistant preview/apply flows have route-specific Cypress coverage for the relevant open/close/mutate/cancel states.

### Feedback, status, and retry states

Classification: covered by existing spec/tests.

Evidence:
- Route-specific traceability rows cover success/error feedback, retry actions, disabled/blocked controls, unsupported states, and preservation of user edits or row context after failures.
- Notable examples include Settings Storage restore failure, Settings Operations import/export feedback, Market Watch provider failure guidance, Notification Inbox triage failure recovery, Discoveries action failure stability, Media upload/assignment error handling, and Assistant preview/apply confirmation.

## Scenario Matrix

| Scenario class | Result | Evidence |
| --- | --- | --- |
| Happy path cross-cutting workflows | Covered by tests, not freshly live-validated | Foundation and route-specific Cypress coverage. |
| Cancellation/abort path | Covered by tests | Escape close, dialog cancel/no-mutation, preview cancellation, and route-specific cancel specs. |
| Invalid input path | Covered by tests | Settings/account/category/provider forms, setup wizard, provider workflows, and command/search error states. |
| Empty state | Covered by tests | Foundation components, users, settings, inbox, media, discoveries, collections, Market Watch, and route-specific empty rows/cards. |
| Loading state | Covered by tests | Foundation components, global search, inbox, runtime/settings routes, and API-backed route states. |
| Error/failure state | Covered by tests | Retry/failure traceability rows across settings, inbox, integrations, Market Watch, storage, media, and assistant workflows. |
| Keyboard-only path | Covered by tests | Foundation accessibility, keyboard shortcuts, command palette, folder tree, settings section navigation. |
| Refresh/back-forward/query persistence | Covered by route tests where applicable | Global search, Integrations query state, Collections pagination/filtering, Market Watch query state, and route-specific persistence tests. |
| Post-action verification | Covered by route tests where applicable | Save/import/export/persisted action tests verify resulting data or refreshed state rather than toast-only outcomes. |
| Permission/role variance | Covered outside this slice where applicable | Auth/permissions, route guard, profile-scope, and users tests; not re-audited as a standalone permission pass here. |

## Element Inventory And Classification

| Element/surface | Classification | Notes |
| --- | --- | --- |
| Icon-only action buttons | Covered | Accessible name guard for sign-in and mobile header plus route-specific labels. |
| Main landmarks/headings | Covered | Responsive landmark Cypress coverage. |
| Sidebar and profile shortcut labels | Covered | Platform-aware shortcut and auth-menu specs. |
| Command palette | Covered | Keyboard open/focus, actions, loading/error states. |
| Row surfaces | Covered | Single-click details, double-click edit, URL selected context, no accidental destructive actions. |
| Checkbox bulk selection | Covered | Explicit checkbox and toolbar behavior. |
| Thumbnail/media lightbox | Covered | Distinct from row details, with previous/next/close controls. |
| Dialog/drawer Escape close | Covered | Foundation and route-specific modal tests. |
| Mutating submit controls | Covered | Duplicate-submit prevention, inline errors, edit preservation, persistence checks. |
| Toast/status/feedback surfaces | Covered through route-specific tests | No single global toast-only contract found; behavior is correctly verified by downstream state where mutating. |
| Browser-policy live route pass | Blocked by prerequisite | Recorded as confidence limitation, not a product issue. |

## Follow-Up Disposition

New product issues opened: none.

Spec/traceability changes made in this branch:
- Added this #501 cross-cutting UX exploration artifact.

No uncovered product behavior remained untracked for this audited source/spec/test scope. Future live manual validation should be rerun when `project-cabinet` browser navigation/snapshot actions are available in the isolated cron context.

## Completeness

Completeness label: Complete with browser-policy limitation.

The cross-cutting behavior families were reconciled against source, OpenSpec, traceability, and focused Cypress/Go coverage. Fresh live browser element validation remains blocked by isolated cron browser policy, so no additional live interaction claims are made.

## Next Recommendation

After #501 is reviewed/merged and demo2 is recycled from `develop`, return to #502 for final traceability/backlog closure or leave #502 open if the exploration sweep needs another closure reconciliation after validator review.
