# Integrations Exploration Audit - 2026-06-18

Issue: #496 `[Exploration] Integrations audit and traceability pass`

Mode: exploratory documentation/spec-alignment slice. No product runtime code changed.

## Runtime Preconditions

- Work branch: `issue-496-integrations-exploration`.
- Demo lane: demo2-helper at `http://127.0.0.1:17882`.
- `/healthz`: HTTP 200 `ok`.
- `/api/runtime`: `app_version=rev-38876695dfed`, `runtime_port=17882`, pid `63632`.
- Browser/profile: OpenClaw `project-cabinet` profile was running and browser doctor passed through CDP port `18801`.
- Browser limitation: OpenClaw `open`, `snapshot`, and direct route navigation for the profile returned `browser navigation blocked by policy`. This pass therefore uses repo/runtime/spec/test evidence and does not claim a fresh live element-by-element browser validation.

Evidence logs for this run are under `.work-agent/logs/issue-496-integrations-exploration/20260618-0300/`.

## Scope Reviewed

Section:
- Integrations route and provider configuration/workflow surfaces.

Screens/routes:
- `/integrations/`, implemented by `ui.web/src/routes/_authenticated/integrations/index.tsx`, `ui.web/src/features/integrations/index.tsx`, and `ui.web/src/features/apps/index.tsx`.
- Related provider work surfaces that are exposed from the eBay/OpenAI/Telegram integration dialogs and routed through the provider APIs.

Component/feature surfaces inventoried from source/spec/test evidence:
- Integrations page header.
- Bootstrap loading/error and active-profile recovery surfaces.
- Integration type selector, text filter, sort controls, rows/cards view toggles, query-backed route state, and no-match states.
- Primary provider rows table, pagination, row single-click details dialog, row double-click edit dialog, and nested row action separation.
- Provider cards and API family/support badges.
- Provider detail dialog with setup instructions, health/last-run metadata, disabled/explained Sync guidance, Validate, Save, and write-only token handling.
- OpenAI dialog Browser Auth, API-key, Test OpenAI, empty-token validation, secret save, and disconnect surfaces.
- Telegram assistant capture status panel.
- eBay setup status panel, health readiness aliases, buyer-interest sync/import, seller operations preview/execute, listing lifecycle preview/execute, and landed-cost planner.

## Findings And Changes

### Finding 1: INTEGRATION-022 default-view wording was stale

Status: fixed in this branch.

Expected current contract:
- Integrations defaults to the full-page table/rows view.
- Cards remain available through the `view=cards` route query state.

Evidence:
- `ui.web/src/routes/_authenticated/integrations/index.tsx` accepts `view` values `rows` and `cards`.
- `ui.web/src/features/apps/index.tsx` resolves `viewMode` to rows when no query is present.
- `openspec/specs/integrations/ui-screen-integrations/spec.md` defines `UI-SCREEN-INTEGRATIONS-011` as the primary table requirement and `UI-SCREEN-INTEGRATIONS-014` row interactions.
- `ui.web/cypress/e2e/integrations/ui-screen-integrations/spec.cy.ts` names and validates the current default as table.

Actual stale artifact:
- `openspec/specs/integrations/spec.md` still said INTEGRATION-022 defaults to cards and only activates rows with `view=rows`.
- `openspec/traceability.md` repeated that stale "defaults to cards" summary.

Correction made:
- Updated `INTEGRATION-022` to require table rows by default and cards only when `view=cards` is present.
- Updated the traceability row to match the existing implementation and Cypress proof.

Issue linkage:
- Tracked and fixed under #496 because this is an integrations exploration/spec-alignment finding discovered during the route audit.

## Screen And Component Coverage

### Integrations route shell and provider list

Classification: covered by existing spec/tests; spec drift fixed.

Requirement anchors:
- `INTEGRATION-013`, `INTEGRATION-014`, `INTEGRATION-020`, `INTEGRATION-022`
- `UI-SCREEN-INTEGRATIONS-001`, `005`, `006`, `011`, `012`, `013`, `014`

Evidence:
- `GET /api/providers/registry` is the list source-of-truth in `internal/app/app.go`.
- `providerRegistryPayload` includes OpenAI, Telegram, eBay, Amazon, and configured AU webshop provider records with health/last-run/API support metadata.
- Cypress covers table default, filter/sort/view switching, pagination, query hydration, zero-result rows/cards states, row details/edit/action separation, provider cards, and API family badges.
- Go UI template tests cover integration route/query usage and labeled config fields.

Residual limit:
- Fresh live browser element validation was blocked by the isolated browser policy.

### Provider detail and configuration dialog

Classification: covered by existing spec/tests.

Requirement anchors:
- `UI-SCREEN-INTEGRATIONS-002`, `003`, `004`, `007`, `008`, `009`, `010`
- `INTEGRATION-020`, `INTEGRATION-024`

Evidence:
- Source exposes provider detail metadata, setup instructions, health/last-run state, Validate/Save, disabled Sync guidance, token replacement, labels, and API family/support profile fields.
- Cypress covers provider detail open, settings persistence, write-only replace-token workflow, validation health reconciliation, eBay readiness states, and config field labels.
- `internal/app/ui_template_contract_test.go` guards provider config labels, validation health reconciliation, eBay setup status, seller operations, listing lifecycle, and landed-cost panel contracts.

Residual limit:
- This run did not mutate shared demo credentials. Persistence claims remain delegated to existing Cypress/API contract tests.

### OpenAI / ChatGPT integration

Classification: covered by existing spec/tests for current non-live-provider scope.

Requirement anchors:
- `PROVIDER-OPENAI-UX-005`, `008`, `009`, `010`
- `ASSISTANT-EXECUTION-006` remains partial in traceability for broader assistant workflow readiness and is outside this #496 route audit closure.

Evidence:
- Cypress `provider-openai-chatgpt-ux/spec.cy.ts` covers clean card/dialog setup, Browser Auth setup-needed state, API-key secret separation, empty-token validation, health preflight blocking, and API-key disconnect preserving Browser Auth state.
- `internal/app/provider_openai_config_api_test.go` covers OpenAI registry and profile-scoped health readiness.

Residual limit:
- No live OpenAI request or Browser Auth proof was executed in this pass.

### Telegram capture integration

Classification: covered by existing spec/tests.

Requirement anchor:
- `TELEGRAM-CATALOG-CAPTURE-025`

Evidence:
- Registry and UI surface Telegram as a sender/chat-authorized assistant capture channel.
- Cypress validates the capture status panel and setup state.
- Go registry tests cover the Telegram capture channel setup state.

Residual limit:
- No real Telegram account/channel mutation was performed in this pass.

### eBay provider setup, Market Watch handoff, buyer-interest, seller operations, and listing lifecycle

Classification: covered for non-credentialed safety-gated states; live marketplace actions remain externally blocked.

Requirement anchors:
- `INTEGRATION-005`, `006`, `007`, `025`, `026`, `027`, `028`
- `COMMERCE-LANDED-COST-001`, `003`

Evidence:
- eBay provider, OpenAPI, traceability, and Cypress tests cover provider auth/search error envelopes, health readiness aliases, setup status, Market Watch provenance/handoff, buyer-interest import without write-back claims, seller operations preview/execute without false remote writes, listing lifecycle draft/local-only and adapter-required blocker states, and landed-cost non-mutating previews.
- Recent issue handoffs for #827, #841, and #842 classify live eBay adapter/write work as blocked pending verified eBay credentials/capability evidence.

Residual limit:
- This pass does not claim live eBay seller, buyer-interest, or listing lifecycle marketplace writes.

## Scenario Matrix

| Scenario class | Result | Evidence |
| --- | --- | --- |
| Happy path route render | Covered by tests, not freshly live-validated | Cypress table/default provider list coverage; browser navigation blocked in this run. |
| Filtering/sorting/type/view controls | Covered by tests | `UI-SCREEN-INTEGRATIONS-001/012/013` Cypress cases. |
| Rows pagination and row interactions | Covered by tests | `UI-SCREEN-INTEGRATIONS-011/014` Cypress cases. |
| Cards view and no-match cards state | Covered by tests | `view=cards` query and zero-result Cypress cases. |
| Bootstrap loading/error/retry | Covered by tests | Registry failure and active-profile recovery Cypress cases. |
| Provider detail open/cancel | Covered by tests | Provider detail/action Cypress and UI template contracts. |
| Save provider settings | Covered by tests | Settings/secrets Cypress and app tests; not mutated live in this pass. |
| Token write-only behavior | Covered by tests | Replace-token and empty-token validation Cypress cases. |
| Provider health validation | Covered by tests | Validation reconciliation and readiness alias Cypress/API contracts. |
| OpenAI setup and disconnect | Covered by tests | Provider OpenAI UX Cypress and provider health tests. |
| Telegram capture setup state | Covered by tests | Telegram capture Cypress and registry tests. |
| eBay setup and safety-gated workflows | Covered by tests for non-credentialed state | eBay Cypress/API/OpenAPI/traceability guard evidence. |
| Live external provider calls | Blocked/not claimed | Missing verified provider credentials/capabilities; not required for this exploration slice. |
| Keyboard/accessibility path | Covered where current specs name it | Route-level live keyboard pass was blocked by browser policy; config field labels are guarded by tests. |
| Refresh/back-forward/query persistence | Covered by tests | Direct route query hydration and preserved query context Cypress cases. |

## Element Inventory And Classification

| Element/surface | Classification | Notes |
| --- | --- | --- |
| Header/title/description | Covered | `UI-SCREEN-INTEGRATIONS-005` guards resolved copy; route uses `pages` translation keys through `Integrations`. |
| Type selector, filter, sort, rows/cards toggle | Covered | Cypress exercises URL-backed state and presentation switching. |
| Provider rows table | Covered | Default table columns, pagination, and row actions are covered. |
| Provider cards | Covered | Cards view remains available and tested through `view=cards`. |
| Row details and row edit dialogs | Covered | Single-click, double-click, and nested action separation covered. |
| Provider detail dialog | Covered | Instructions, health, last-run, Validate, Save, disabled Sync guidance. |
| Credential fields | Covered | Visible/programmatic labels and write-only token handling. |
| Bootstrap error/recovery | Covered | Registry retry, profile selection, and inline profile creation are covered. |
| OpenAI Browser Auth/API key/Test sections | Covered for setup-needed/non-live-provider behavior | Live provider call not claimed. |
| Telegram capture panel | Covered | Status/setup state shown from registry/profile authorization settings. |
| eBay setup status panel | Covered | Auth mode, marketplace, token state, health, next action. |
| eBay buyer-interest panel | Covered for local import/write-back blocker | No remote write-back claimed. |
| eBay seller operations panel | Covered for preview/local sync/adapter blocker | No remote seller write claimed. |
| eBay listing lifecycle panel | Covered for local draft/confirmed adapter blocker | No remote listing write claimed. |
| eBay landed-cost planner | Covered for non-mutating preview | No inventory/cost mutation claimed. |

## Follow-Up Disposition

New product issues opened: none.

Spec/traceability changes made in this branch:
- Aligned `INTEGRATION-022` default-view contract with the current table-first UI and Cypress evidence.
- Aligned the `INTEGRATION-022` traceability summary with the same table-first behavior.

No uncovered product behavior remained untracked for this audited scope. Broader live provider work remains tracked in #827, #841, and #842 and remains blocked on external credential/capability evidence before any live marketplace claims.

## Completeness

Completeness label: Complete with browser-policy limitation.

The route, screen, and component inventory was reconciled against source, specs, traceability, and focused tests. One stale spec/traceability contract was corrected. Fresh live element-by-element browser validation remains blocked by OpenClaw browser policy in this isolated cron context.

## Next Recommendation

After #496 is reviewed/merged and demo2 is recycled from `develop`, continue route-ordered exploration at #497 Assistant / AI surfaces unless a higher-priority Cabinet issue is added.
