# OpenSpec Hardening Audit

- timestamp: 2026-02-27T01:24:40.8642210Z
- commit: 9ca1215
- compared-against: d0fd84c
- issue: #184
- validation: `openspec validate --all` => `57 passed, 0 failed`
- phase-2 cleanup: targeted scenario-precondition refinements applied to 10 high-value specs without ID changes

| spec path | requirements total | IDs present before | IDs present after | placeholder GIVEN before | placeholder GIVEN after | IDs added | key executable criteria added | status |
| --- | ---: | ---: | ---: | ---: | ---: | --- | --- | --- |
| openspec/specs/ai-assist/spec.md | 4 | 0 | 4 | 4 | 0 | AI-ASSIST-001, AI-ASSIST-002, AI-ASSIST-003, AI-ASSIST-004 | Concrete GIVEN preconditions with actor/config/data context; Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/auth/spec.md | 3 | 0 | 3 | 0 | 0 | AUTH-001, AUTH-002, AUTH-003 | Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/barcodes/spec.md | 2 | 0 | 2 | 0 | 0 | BARCODES-001, BARCODES-002 | Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/candidates/spec.md | 2 | 0 | 2 | 0 | 0 | CANDIDATES-001, CANDIDATES-002 | Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/chat-copilot/spec.md | 4 | 0 | 4 | 4 | 0 | CHAT-COPILOT-001, CHAT-COPILOT-002, CHAT-COPILOT-003, CHAT-COPILOT-004 | Concrete GIVEN preconditions with actor/config/data context; Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/cloud-auth-billing/spec.md | 5 | 0 | 5 | 5 | 0 | CLOUD-AUTH-BILLING-001, CLOUD-AUTH-BILLING-002, CLOUD-AUTH-BILLING-003, CLOUD-AUTH-BILLING-004, CLOUD-AUTH-BILLING-005 | Concrete GIVEN preconditions with actor/config/data context; Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/collection-domain/spec.md | 4 | 0 | 4 | 4 | 0 | COLLECTION-DOMAIN-001, COLLECTION-DOMAIN-002, COLLECTION-DOMAIN-003, COLLECTION-DOMAIN-004 | Concrete GIVEN preconditions with actor/config/data context; Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/data-management/spec.md | 2 | 0 | 2 | 0 | 0 | DATA-MANAGEMENT-001, DATA-MANAGEMENT-002 | Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/diagnostics/spec.md | 4 | 0 | 4 | 0 | 0 | DIAGNOSTICS-001, DIAGNOSTICS-002, DIAGNOSTICS-003, DIAGNOSTICS-004 | Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/discovery/spec.md | 2 | 0 | 2 | 0 | 0 | DISCOVERY-001, DISCOVERY-002 | Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/documentation-governance/spec.md | 5 | 0 | 5 | 4 | 0 | DOCUMENTATION-GOVERNANCE-001, DOCUMENTATION-GOVERNANCE-002, DOCUMENTATION-GOVERNANCE-003, DOCUMENTATION-GOVERNANCE-004, DOCUMENTATION-GOVERNANCE-005 | Concrete GIVEN preconditions with actor/config/data context; Explicit API status-code acceptance outcomes; Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/entitlements/spec.md | 1 | 0 | 1 | 0 | 0 | ENTITLEMENTS-001 | Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/errors/spec.md | 3 | 0 | 3 | 0 | 0 | ERRORS-001, ERRORS-002, ERRORS-003 | Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/future-hooks/spec.md | 2 | 0 | 2 | 2 | 0 | FUTURE-HOOKS-001, FUTURE-HOOKS-002 | Concrete GIVEN preconditions with actor/config/data context; Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/integrations/spec.md | 2 | 2 | 2 | 0 | 0 | - | No hardening delta in this commit | pass |
| openspec/specs/licensing/spec.md | 2 | 0 | 2 | 0 | 0 | LICENSING-001, LICENSING-002 | Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/logging/spec.md | 4 | 0 | 4 | 0 | 0 | LOGGING-001, LOGGING-002, LOGGING-003, LOGGING-004 | Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/lookup/spec.md | 1 | 0 | 1 | 0 | 0 | LOOKUP-001 | Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/matching/spec.md | 2 | 0 | 2 | 0 | 0 | MATCHING-001, MATCHING-002 | Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/non-functional/spec.md | 2 | 0 | 2 | 0 | 0 | NON-FUNCTIONAL-001, NON-FUNCTIONAL-002 | Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/photos-media/spec.md | 4 | 0 | 4 | 4 | 0 | PHOTOS-MEDIA-001, PHOTOS-MEDIA-002, PHOTOS-MEDIA-003, PHOTOS-MEDIA-004 | Concrete GIVEN preconditions with actor/config/data context; Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/profiles/spec.md | 1 | 0 | 1 | 0 | 0 | PROFILES-001 | Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/provider-amazon/spec.md | 3 | 3 | 3 | 0 | 0 | - | Explicit API status-code acceptance outcomes | pass |
| openspec/specs/provider-au-webshops/spec.md | 3 | 3 | 3 | 0 | 0 | - | No hardening delta in this commit | pass |
| openspec/specs/provider-ebay/spec.md | 3 | 3 | 3 | 0 | 0 | - | No hardening delta in this commit | pass |
| openspec/specs/provider-registry/spec.md | 4 | 4 | 4 | 0 | 0 | - | Explicit API status-code acceptance outcomes | pass |
| openspec/specs/runtime-core/spec.md | 3 | 0 | 3 | 3 | 0 | RUNTIME-CORE-001, RUNTIME-CORE-002, RUNTIME-CORE-003 | Concrete GIVEN preconditions with actor/config/data context; Explicit API status-code acceptance outcomes; Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/scanner/spec.md | 4 | 1 | 4 | 0 | 0 | SCANNER-001, SCANNER-002, SCANNER-003 | Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/search/spec.md | 2 | 0 | 2 | 0 | 0 | SEARCH-001, SEARCH-002 | Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/security/spec.md | 2 | 0 | 2 | 0 | 0 | SECURITY-001, SECURITY-002 | Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/settings/spec.md | 2 | 0 | 2 | 0 | 0 | SETTINGS-001, SETTINGS-002 | Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/ui-data-contract-parity/spec.md | 4 | 0 | 4 | 4 | 0 | UI-DATA-CONTRACT-PARITY-001, UI-DATA-CONTRACT-PARITY-002, UI-DATA-CONTRACT-PARITY-003, UI-DATA-CONTRACT-PARITY-004 | Concrete GIVEN preconditions with actor/config/data context; Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/ui-foundation-accessibility/spec.md | 2 | 0 | 2 | 0 | 0 | UI-FOUNDATION-ACCESSIBILITY-001, UI-FOUNDATION-ACCESSIBILITY-002 | Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/ui-foundation-auth-menus-shortcuts/spec.md | 3 | 0 | 3 | 3 | 0 | UI-FOUNDATION-AUTH-MENUS-SHORTCUTS-001, UI-FOUNDATION-AUTH-MENUS-SHORTCUTS-002, UI-FOUNDATION-AUTH-MENUS-SHORTCUTS-003 | Concrete GIVEN preconditions with actor/config/data context; Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/ui-foundation-components/spec.md | 5 | 0 | 5 | 5 | 0 | UI-FOUNDATION-COMPONENTS-001, UI-FOUNDATION-COMPONENTS-002, UI-FOUNDATION-COMPONENTS-003, UI-FOUNDATION-COMPONENTS-004, UI-FOUNDATION-COMPONENTS-005 | Concrete GIVEN preconditions with actor/config/data context; Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/ui-foundation-interactions/spec.md | 2 | 0 | 2 | 0 | 0 | UI-FOUNDATION-INTERACTIONS-001, UI-FOUNDATION-INTERACTIONS-002 | Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/ui-foundation-shell-navigation/spec.md | 4 | 0 | 4 | 4 | 0 | UI-FOUNDATION-SHELL-NAVIGATION-001, UI-FOUNDATION-SHELL-NAVIGATION-002, UI-FOUNDATION-SHELL-NAVIGATION-003, UI-FOUNDATION-SHELL-NAVIGATION-004 | Concrete GIVEN preconditions with actor/config/data context; Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/ui-foundation-theme-rtl-i18n/spec.md | 4 | 0 | 4 | 4 | 0 | UI-FOUNDATION-THEME-RTL-I18N-001, UI-FOUNDATION-THEME-RTL-I18N-002, UI-FOUNDATION-THEME-RTL-I18N-003, UI-FOUNDATION-THEME-RTL-I18N-004 | Concrete GIVEN preconditions with actor/config/data context; Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/ui-governance-gates/spec.md | 6 | 0 | 6 | 6 | 0 | UI-GOVERNANCE-GATES-001, UI-GOVERNANCE-GATES-002, UI-GOVERNANCE-GATES-003, UI-GOVERNANCE-GATES-004, UI-GOVERNANCE-GATES-005, UI-GOVERNANCE-GATES-006 | Concrete GIVEN preconditions with actor/config/data context; Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/ui-performance/spec.md | 2 | 0 | 2 | 0 | 0 | UI-PERFORMANCE-001, UI-PERFORMANCE-002 | Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/ui-scale/spec.md | 3 | 0 | 3 | 0 | 0 | UI-SCALE-001, UI-SCALE-002, UI-SCALE-003 | Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/ui-screen-chat-copilot/spec.md | 3 | 0 | 3 | 3 | 0 | UI-SCREEN-CHAT-COPILOT-001, UI-SCREEN-CHAT-COPILOT-002, UI-SCREEN-CHAT-COPILOT-003 | Concrete GIVEN preconditions with actor/config/data context; Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/ui-screen-discover/spec.md | 3 | 0 | 3 | 3 | 0 | UI-SCREEN-DISCOVER-001, UI-SCREEN-DISCOVER-002, UI-SCREEN-DISCOVER-003 | Concrete GIVEN preconditions with actor/config/data context; Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/ui-screen-home/spec.md | 3 | 0 | 3 | 5 | 0 | UI-SCREEN-HOME-001, UI-SCREEN-HOME-002, UI-SCREEN-HOME-003 | Concrete GIVEN preconditions with actor/config/data context; Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/ui-screen-inventory-ai-assist/spec.md | 3 | 0 | 3 | 3 | 0 | UI-SCREEN-INVENTORY-AI-ASSIST-001, UI-SCREEN-INVENTORY-AI-ASSIST-002, UI-SCREEN-INVENTORY-AI-ASSIST-003 | Concrete GIVEN preconditions with actor/config/data context; Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/ui-screen-inventory-barcodes/spec.md | 3 | 0 | 3 | 3 | 0 | UI-SCREEN-INVENTORY-BARCODES-001, UI-SCREEN-INVENTORY-BARCODES-002, UI-SCREEN-INVENTORY-BARCODES-003 | Concrete GIVEN preconditions with actor/config/data context; Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/ui-screen-inventory-items/spec.md | 3 | 0 | 3 | 4 | 0 | UI-SCREEN-INVENTORY-ITEMS-001, UI-SCREEN-INVENTORY-ITEMS-002, UI-SCREEN-INVENTORY-ITEMS-003 | Concrete GIVEN preconditions with actor/config/data context; Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/ui-screen-inventory-photos/spec.md | 3 | 0 | 3 | 4 | 0 | UI-SCREEN-INVENTORY-PHOTOS-001, UI-SCREEN-INVENTORY-PHOTOS-002, UI-SCREEN-INVENTORY-PHOTOS-003 | Concrete GIVEN preconditions with actor/config/data context; Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/ui-screen-onboarding-auth/spec.md | 3 | 0 | 3 | 3 | 0 | UI-SCREEN-ONBOARDING-AUTH-001, UI-SCREEN-ONBOARDING-AUTH-002, UI-SCREEN-ONBOARDING-AUTH-003 | Concrete GIVEN preconditions with actor/config/data context; Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/ui-screen-reports/spec.md | 3 | 0 | 3 | 3 | 0 | UI-SCREEN-REPORTS-001, UI-SCREEN-REPORTS-002, UI-SCREEN-REPORTS-003 | Concrete GIVEN preconditions with actor/config/data context; Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/ui-screen-scanner/spec.md | 3 | 0 | 3 | 3 | 0 | UI-SCREEN-SCANNER-001, UI-SCREEN-SCANNER-002, UI-SCREEN-SCANNER-003 | Concrete GIVEN preconditions with actor/config/data context; Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/ui-screen-settings/spec.md | 3 | 0 | 3 | 3 | 0 | UI-SCREEN-SETTINGS-001, UI-SCREEN-SETTINGS-002, UI-SCREEN-SETTINGS-003 | Concrete GIVEN preconditions with actor/config/data context; Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/ui-semantic-component-layer/spec.md | 7 | 0 | 7 | 7 | 0 | UI-SEMANTIC-COMPONENT-LAYER-001, UI-SEMANTIC-COMPONENT-LAYER-002, UI-SEMANTIC-COMPONENT-LAYER-003, UI-SEMANTIC-COMPONENT-LAYER-004, UI-SEMANTIC-COMPONENT-LAYER-005, UI-SEMANTIC-COMPONENT-LAYER-006, UI-SEMANTIC-COMPONENT-LAYER-007 | Concrete GIVEN preconditions with actor/config/data context; Deterministic requirement IDs assigned to missing headings | pass |
| openspec/specs/wishlist-pricing-dashboard/spec.md | 4 | 0 | 4 | 4 | 0 | WISHLIST-PRICING-DASHBOARD-001, WISHLIST-PRICING-DASHBOARD-002, WISHLIST-PRICING-DASHBOARD-003, WISHLIST-PRICING-DASHBOARD-004 | Concrete GIVEN preconditions with actor/config/data context; Deterministic requirement IDs assigned to missing headings | pass |

## Wave Evidence
- Wave: `Wave 1 (auth/security/error)`
- Executed: `2026-02-27`
- Traceability delta:
  - `implemented`: `10 -> 22`
  - `partial`: `156 -> 144`
  - Net reduced partial IDs: `12`
- IDs moved to implemented:
  - `AUTH-001`, `AUTH-002`, `AUTH-003`
  - `CLOUD-AUTH-BILLING-001`, `CLOUD-AUTH-BILLING-002`, `CLOUD-AUTH-BILLING-003`
  - `DIAGNOSTICS-001`
  - `ERRORS-001`, `ERRORS-002`
  - `RUNTIME-CORE-001`, `RUNTIME-CORE-003`
  - `SECURITY-002`
- Test evidence:
  - `internal/app/traceability_wave1_auth_security_test.go`
  - `internal/app/auth_lock_runtime_api_test.go`
  - `internal/app/cloud_auth_api_test.go`
  - `internal/app/clerk_billing_webhook_api_test.go`
  - `internal/app/logging_recovery_api_test.go`
  - `internal/app/license_api_test.go`
  - `internal/app/ui_template_contract_test.go`
  - `tests/shop_providers_contract_test.go`

## Wave Evidence
- Wave: `Wave 2 (runtime/diagnostics/security)`
- Executed: `2026-02-27`
- Traceability delta:
  - `implemented`: `22 -> 28`
  - `partial`: `144 -> 138`
  - Net reduced partial IDs: `6`
- IDs moved to implemented:
  - `CLOUD-AUTH-BILLING-005`
  - `DIAGNOSTICS-002`, `DIAGNOSTICS-003`, `DIAGNOSTICS-004`
  - `ERRORS-003`
  - `SECURITY-001`
- Runtime behavior implemented:
  - Strict cloud auth token verification mode (`CABINET_CLOUD_AUTH_ENFORCE_SIGNED_TOKENS=1`) with HS256 signature validation (`CABINET_CLOUD_AUTH_HS256_SECRET`).
  - Diagnostics API contracts: `/api/diagnostics/config` and `/api/diagnostics/event` with explicit opt-in and local-only behavior.
  - Deterministic error taxonomy contract: `/api/errors/classify`.
  - Secret fallback hardening: encrypted fallback storage (`enc:v1:`) when explicit insecure fallback is enabled.
- Test evidence:
  - `internal/app/traceability_wave2_priority_test.go`
  - `internal/profile/secrets_fallback_security_test.go`

## Wave Evidence
- Wave: `Wave 3 (cloud gating/runtime update)`
- Executed: `2026-02-27`
- Traceability delta:
  - `implemented`: `28 -> 30`
  - `partial`: `138 -> 136`
  - Net reduced partial IDs: `2`
- IDs moved to implemented:
  - `CLOUD-AUTH-BILLING-004`
  - `RUNTIME-CORE-002`
- Runtime behavior implemented:
  - Deterministic pro-gated runtime denial across mutation endpoints for `ai_assist`, `price_tracking`, and `scanner_automation` via consistent `403` envelope and no-side-effect denial path.
  - Runtime update harness endpoint `/api/runtime/update/install` for signature verification + release-channel compatibility with deterministic error envelopes:
    - `INVALID_UPDATE_SIGNATURE` (`400`)
    - `UPDATE_CHANNEL_MISMATCH` (`409`)
- Test evidence:
  - `internal/app/traceability_wave3_priority_test.go`

## Wave Evidence
- Wave: `Wave 4 (provider/scanner contracts)`
- Executed: `2026-02-27`
- Traceability delta:
  - `implemented`: `30 -> 39`
  - `partial`: `136 -> 127`
  - Net reduced partial IDs: `9`
- IDs moved to implemented:
  - `CANDIDATES-001`, `CANDIDATES-002`
  - `INTEGRATION-008`, `INTEGRATION-009`, `INTEGRATION-012`, `INTEGRATION-013`, `INTEGRATION-014`, `INTEGRATION-015`
  - `SCANNER-002`
- Runtime behavior implemented:
  - Added canonical provider registry endpoint `/api/providers/registry` with Amazon mode metadata and AU webshop catalog domains.
  - Added `/api/providers/amazon/run` contract path with deterministic disabled-mode `409` envelope and normalized candidate payload when enabled.
  - Added `/api/providers/au-webshops/parse-stock` to normalize stock signal extraction into deterministic fields.
  - Expanded scheduled scanner summary contract fields: `run_id`, `query_sets_executed`, `candidates_collected`, `failures`.
- Test evidence:
  - `internal/app/traceability_wave4_provider_scanner_test.go`

## Wave Evidence
- Wave: `Wave 5 (collection/discovery/pricing/matching workflows)`
- Executed: `2026-02-27`
- Traceability delta:
  - `implemented`: `39 -> 49`
  - `partial`: `127 -> 117`
  - Net reduced partial IDs: `10`
- IDs moved to implemented:
  - `SEARCH-001`, `SEARCH-002`
  - `DISCOVERY-001`, `DISCOVERY-002`
  - `WISHLIST-PRICING-DASHBOARD-001`, `WISHLIST-PRICING-DASHBOARD-002`, `WISHLIST-PRICING-DASHBOARD-003`, `WISHLIST-PRICING-DASHBOARD-004`
  - `MATCHING-001`, `MATCHING-002`
- Runtime behavior implemented:
  - No runtime code changes required; existing API behavior already satisfied these contracts.
- Test evidence:
  - `internal/app/traceability_wave5_collection_workflows_test.go`
