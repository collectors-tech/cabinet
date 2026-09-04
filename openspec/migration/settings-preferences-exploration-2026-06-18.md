# Settings / Preferences Exploration Audit - 2026-06-18

Issue: #499 `[Exploration] Settings / preferences audit and traceability pass`

Mode: exploratory documentation/spec-alignment slice. No product runtime code changed.

## Runtime Preconditions

- Work branch: `issue-499-settings-preferences-exploration`.
- Demo lane: demo2-helper at `http://127.0.0.1:17882`.
- `/healthz`: HTTP 200 `ok`.
- `/api/runtime`: `app_version=rev-5cfd6f32297f`, `runtime_port=17882`, pid `52736`.
- Browser/profile: OpenClaw `project-cabinet` profile doctor passed through CDP port `18801`.
- Browser limitation: OpenClaw browser `open` for `http://127.0.0.1:17882/settings/profile` returned `browser navigation blocked by policy`. This pass therefore uses runtime health plus repo/source/spec/test evidence and does not claim fresh live element-by-element browser interaction.

Evidence logs for this run are under `.work-agent/logs/issue-499-settings-preferences-exploration/20260618-1050/`.

## Scope Reviewed

Section:
- Settings / preferences surfaces.

Screens/routes:
- `/settings` redirecting to `/settings/profile`.
- `/settings/profile`, `/settings/account`, `/settings/appearance`, `/settings/notifications`, `/settings/display`, `/settings/storage`, `/settings/categories`, `/settings/operations`, and `/settings/billing`.
- Primary rail Storage affordance and settings sidebar/mobile selector route navigation.

Component/feature surfaces inventoried from source/spec/test evidence:
- Settings header, section sidebar, mobile section selector, active section route state, and deep-link routing.
- Profile form, Telegram catalog capture authorization fields, URL rows, URL validation, retry, save success/failure, and missing-active-profile blocker.
- Account form, required field validation, retry, save success/failure, and missing-active-profile blocker.
- Appearance theme, font, language, update, retry, save failure, missing-profile blocker, first-run dark-theme behavior, and fallback language behavior.
- Notifications scope/channel toggles, guarded security control, update, retry, save failure, and missing-profile blocker.
- Display sidebar item selection, clear selection, minimum-selection validation, update, retry, save failure, and missing-profile blocker.
- Storage media path, write-check/path validation, migration, degraded storage context, retry, maintenance actions, backup table, restore failure, integrity check, export downloads, and degraded export blocking.
- Categories reusable taxonomy settings for inventory categories, packaging grades, and item type condition scales.
- Operations runtime/recovery/import-export/logs/queue controls and Billing static entitlement placeholders.

## Findings And Changes

### Finding 1: Settings shell canonical section list omitted Categories

Status: fixed in this branch.

Expected current contract:
- Settings shell canonical section labels include `Categories` because `/settings/categories` is an implemented first-class settings section with its own OpenSpec, route, sidebar entry, Cypress suite, and traceability rows.

Observed artifact:
- `openspec/specs/settings/ui-screen-settings/spec.md` listed Profile, Account, Appearance, Notifications, Display, Storage, Operations, and Billing under `UI-SCREEN-SETTINGS-002`, but omitted Categories.
- The same file's section notes also omitted `settings/categories/spec.md`.

Evidence:
- `ui.web/src/features/settings/index.tsx` includes a `Categories` sidebar item routed to `/settings/categories`.
- `ui.web/src/routes/_authenticated/settings/categories.tsx` implements the route.
- `openspec/specs/settings/categories/spec.md` defines `UI-SCREEN-SETTINGS-CATEGORIES-001` through `003`.
- `ui.web/cypress/e2e/settings/ui-screen-settings/spec.cy.ts` already expects `/settings/categories` and the `Categories` navigation label.
- `openspec/traceability.md` maps the Categories settings requirements to `ui.web/cypress/e2e/settings/categories/spec.cy.ts`.

Correction made:
- Updated `UI-SCREEN-SETTINGS-002` to include `Categories`.
- Added `settings/categories/spec.md` to the detailed section notes.

Issue linkage:
- Tracked and fixed under #499 as a settings/preferences exploration spec-alignment finding.

## Screen And Component Coverage

### Settings shell and section navigation

Classification: covered by existing spec/tests; spec drift fixed.

Requirement anchors:
- `UI-SCREEN-SETTINGS-001` through `008`.

Evidence:
- `ui.web/src/routes/_authenticated/settings/index.tsx` redirects `/settings` to `/settings/profile`.
- `ui.web/src/features/settings/index.tsx` renders the canonical settings sidebar list and page header.
- `ui.web/src/features/settings/components/sidebar-nav.tsx` routes links/select options through TanStack Router.
- `ui.web/cypress/e2e/settings/ui-screen-settings/spec.cy.ts` covers direct section routes, canonical labels/links including Categories, section error state, Storage primary rail entry, missing-active-profile blockers, Operations/Billing deep links, and keyboard navigation across Notifications, Categories, Operations, and Billing.

Residual limit:
- Fresh live browser interaction was blocked by policy, so this pass does not independently re-execute settings navigation in the browser.

### Profile and Account preferences

Classification: covered by existing spec/tests.

Requirement anchors:
- `UI-SCREEN-SETTINGS-PROFILE-001` through `005`.
- `UI-SCREEN-SETTINGS-ACCOUNT-001` through `007`.

Evidence:
- `openspec/specs/settings/profile/spec.md` and `account/spec.md` define validation, retry, update, save-failure, and missing-profile contracts.
- Cypress suites under `ui.web/cypress/e2e/settings/profile/` and `ui.web/cypress/e2e/settings/account/` cover persistence, invalid/no-save behavior, retry, save failure preservation, and blocker states.

### Appearance, Notifications, Display, and Categories

Classification: covered by existing spec/tests.

Requirement anchors:
- `UI-SCREEN-SETTINGS-APPEARANCE-001` through `007`.
- `UI-SCREEN-SETTINGS-NOTIFICATIONS-001` through `005`.
- `UI-SCREEN-SETTINGS-DISPLAY-001` through `006`.
- `UI-SCREEN-SETTINGS-CATEGORIES-001` through `003`.

Evidence:
- Appearance specs/tests cover theme/font/language persistence, Chinese/Japanese language choices, retry, save failure without unpersisted side effects, and missing-profile blocking.
- Notifications specs/tests cover scope/channel persistence, guarded security control behavior, retry/update actions, save failure preservation, and missing-profile blocking.
- Display specs/tests cover sidebar item selection, clear selection, minimum-selection validation, update/retry, save failure preservation, and missing-profile blocking.
- Categories specs/tests cover reusable taxonomy persistence, save failure preservation, and missing-profile blocking.

### Storage, Operations, and Billing

Classification: covered by existing spec/tests.

Requirement anchors:
- `UI-SCREEN-SETTINGS-STORAGE-001` through `011`.
- `UI-SCREEN-SETTINGS-006` and operations route contracts from #495.
- `UI-SCREEN-SETTINGS-BILLING-001`.

Evidence:
- Storage specs/tests cover media path management, degraded storage context, retry, maintenance actions, backup table, restore/integrity failure handling, and export downloads only in ready state.
- Operations route behavior was reconciled in #495 and remains covered by `ui.web/cypress/e2e/settings/operations/spec.cy.ts` plus Go UI template contract tests.
- Billing specs/tests cover static plan/license placeholders and disabled billing portal controls until cloud billing is available.

## Scenario Matrix

| Scenario class | Result | Evidence |
| --- | --- | --- |
| Happy path settings entry | Covered by tests | `/settings` redirect and direct section routing covered by `UI-SCREEN-SETTINGS-001` Cypress. |
| Section navigation and keyboard path | Covered by tests | Sidebar labels/links and keyboard activation covered by `UI-SCREEN-SETTINGS-002` and `008`; spec now includes Categories. |
| Form update and persistence | Covered by tests | Profile, Account, Appearance, Notifications, Display, Categories, and Storage suites cover update paths. |
| Cancellation/abort path | Not broadly applicable | Settings forms use explicit update/retry actions rather than modal cancellation; transient restore confirmation behavior is covered under Storage. |
| Invalid input path | Covered by tests | Profile invalid URL, Account invalid fields, Display empty selection, Storage invalid path, and taxonomy/settings save failures are covered. |
| Empty/degraded state | Covered by tests | Missing active profile blockers and degraded storage context have focused Cypress coverage. |
| Loading/retry state | Covered by tests | Section retry behavior is covered across Profile, Account, Appearance, Notifications, Display, Storage, and Operations. |
| Error/failure state | Covered by tests | Save failure and failed update preservation are covered by focused settings suites. |
| Refresh/back-forward persistence | Covered by route/persistence tests where applicable | Direct route hydration and persisted settings reload behavior are represented in Cypress; browser history was not freshly exercised due policy. |
| Permission/role variance | Not applicable beyond active-profile context | Missing active profile blockers cover the current settings permission boundary. |

## Element Inventory And Classification

| Element/surface | Classification | Notes |
| --- | --- | --- |
| Settings header and page identity | Covered | Source renders resolved user-facing copy and settings icon. |
| Sidebar/mobile section navigation | Covered; spec drift fixed | Categories is implemented/tested and now included in shell spec. |
| Profile form and Telegram capture authorization | Covered | Profile Cypress and spec traceability cover persistence and validation. |
| Account form | Covered | Account Cypress covers save, retry, invalid no-save, save failure, and blocker state. |
| Appearance controls | Covered | Appearance Cypress covers update/retry/failure/blocker; language expansion and dark first paint are specified. |
| Notifications controls | Covered | Notifications Cypress covers update/retry/failure/blocker and guarded controls. |
| Display item controls | Covered | Display Cypress covers clear/update/minimum-selection/failure/blocker. |
| Storage diagnostics/export/backup controls | Covered | Storage Cypress covers maintenance, backup, restore/integrity, and export contracts. |
| Categories taxonomy controls | Covered | Categories Cypress and Go taxonomy contract tests cover reusable taxonomy settings. |
| Operations controls | Covered | #495 reconciled the operations route and tests. |
| Billing controls | Covered | Billing Cypress covers disabled non-mutating state. |

## Traceability Summary

Existing traceability already maps the reviewed settings contracts:
- `UI-SCREEN-SETTINGS-001` through `008` -> `ui.web/cypress/e2e/settings/ui-screen-settings/spec.cy.ts`.
- Profile, Account, Appearance, Notifications, Display, Storage, Categories, and Billing section specs -> focused Cypress suites under `ui.web/cypress/e2e/settings/`.
- Categories taxonomy also maps to `TestSettingsCategoriesTaxonomyControlsContract` in `internal/app/ui_template_contract_test.go`.

The only traceability-relevant change was updating the settings shell spec text to align with existing implementation and tests for Categories.

## Follow-Up Recommendation

#499 can close after this report/spec alignment is merged and demo2 is recycled from `develop`. The next route-ordered exploration issue is #500 Runtime / operational states unless newer higher-priority product feedback is added.

Completeness label: Complete with browser-policy limitation. Source/spec/test coverage was reconciled, a stale shell spec omission was fixed, and fresh live browser interaction remains blocked by OpenClaw browser policy in this isolated cron context.
