# Cabinet OpenSpec Baseline

This directory is the primary behavioral source of truth for Cabinet product and UI specifications.

Legacy markdown under `docs/` has been migrated and removed.

## Capability Specs
- `runtime-core`
- `auth`
- `profiles`
- `collection-domain`
- `photos-media`
- `barcodes`
- `lookup`
- `ai-assist`
- `scanner`
- `candidates`
- `matching`
- `discovery`
- `wishlist-pricing-dashboard`
- `search`
- `data-management`
- `licensing`
- `entitlements`
- `settings`
- `integrations`
- `errors-logging-diagnostics`
- `future-hooks`
- `non-functional`
- `security`
- `chat-copilot`
- `cloud-auth-billing`
- `provider-registry`
- `provider-ebay`
- `provider-amazon`
- `provider-au-webshops`
- `documentation-governance`

### UI Foundations
- `ui-foundation-shell-navigation`
- `ui-foundation-auth-menus-shortcuts`
- `ui-foundation-theme-rtl-i18n`
- `ui-foundation-interactions`
- `ui-foundation-accessibility`
- `ui-foundation-components`
- `ui-semantic-component-layer`

### UI Contracts and Gates
- `ui-data-contract-parity`
- `ui-scale`
- `ui-performance`
- `ui-governance-gates`

## Screen Specs
- `ui-screen-home`
- `ui-screen-onboarding-auth`
- `ui-screen-inventory-items`
- `ui-screen-inventory-photos`
- `ui-screen-inventory-barcodes`
- `ui-screen-inventory-ai-assist`
- `ui-screen-scanner`
- `ui-screen-discover`
- `ui-screen-reports`
- `ui-screen-settings`
- `ui-screen-chat-copilot`

## Use-Case Namespace Standard
Canonical IDs:
- Cross-feature use cases from legacy document: `UC-01` .. `UC-23`
- Screen use cases: `UC-<SCREEN>-<NN>`
  - Home: `UC-HOME-<NN>`
  - Onboarding/Auth: `UC-ONB-<NN>`
  - Inventory Items: `UC-INV-<NN>`
  - Inventory Photos: `UC-PHO-<NN>`
  - Inventory Barcodes: `UC-BAR-<NN>`
  - Inventory AI: `UC-AI-<NN>`
  - Scanner: `UC-SCN-<NN>`
  - Discover: `UC-DIS-<NN>`
  - Reports: `UC-REP-<NN>`
  - Settings: `UC-SET-<NN>`
  - Chat: `UC-CHAT-<NN>`

Rules:
1. Every critical flow SHALL have a UC ID.
2. Every UC SHALL have deterministic expected result.
3. Every UC SHALL map to automated tests (existing or planned).

## Mapping: UC-01..UC-23 to OpenSpec
| Legacy UC | Primary OpenSpec Capability | Supporting Screen Spec(s) |
| --- | --- | --- |
| UC-01 Install and launch | `runtime-core` | `ui-screen-onboarding-auth` |
| UC-02 Create profile and first credential | `auth` | `ui-screen-onboarding-auth` |
| UC-03 Unlock app and auto-lock | `auth` | `ui-screen-onboarding-auth` |
| UC-04 Recover after credential loss | `auth` | `ui-screen-onboarding-auth`, `ui-screen-settings` |
| UC-05 Switch between profiles | `profiles` | `ui-screen-onboarding-auth`, `ui-screen-settings` |
| UC-06 Create canonical item and instances | `collection-domain` | `ui-screen-inventory-items` |
| UC-07 Import existing collection | `data-management` | `ui-screen-settings` |
| UC-08 Manage photos | `photos-media` | `ui-screen-inventory-photos` |
| UC-09 Add/resolve barcodes | `barcodes` | `ui-screen-inventory-barcodes` |
| UC-10 Configure scanner query sets | `scanner` | `ui-screen-scanner` |
| UC-11 Execute scanner and persist candidates | `candidates` | `ui-screen-scanner` |
| UC-12 Match candidates | `matching` | `ui-screen-scanner`, `ui-screen-discover` |
| UC-13 Not-in-collection actions | `discovery` | `ui-screen-discover` |
| UC-14 Wishlist and targets | `wishlist-pricing-dashboard` | `ui-screen-discover`, `ui-screen-reports` |
| UC-15 Price tracking trends | `wishlist-pricing-dashboard` | `ui-screen-reports` |
| UC-16 Search/filter/sort | `search` | `ui-screen-inventory-items` |
| UC-17 Dashboard weekly review | `wishlist-pricing-dashboard` | `ui-screen-home` |
| UC-18 AI assist suggestions | `ai-assist` | `ui-screen-inventory-ai-assist` |
| UC-19 License and gating | `entitlements` | `ui-screen-settings` |
| UC-20 Logs and diagnostics export | `errors-logging-diagnostics` | `ui-screen-settings` |
| UC-21 Restore backup | `data-management` | `ui-screen-settings` |
| UC-22 Safe upgrade | `runtime-core` | `ui-screen-settings` |
| UC-23 Chat copilot actions | `chat-copilot` | `ui-screen-chat-copilot` |

## Capability Test Mapping (Integration/API)
| Capability | Required Test Type | Current/Planned Mapping |
| --- | --- | --- |
| runtime-core | API integration | `internal/app/ui_root_test.go`, startup smoke |
| auth | API integration | `internal/app/auth_*_api_test.go` |
| profiles | API integration | `internal/app/profile_*_api_test.go` |
| collection-domain | API integration | `internal/app/items_api_test.go` |
| photos-media | API integration | `internal/app/photos_api_test.go` |
| barcodes | API integration | `internal/app/barcodes_api_test.go` |
| lookup | API integration | `internal/app/barcodes_api_test.go` |
| ai-assist | API integration | `internal/app/ai_api_test.go` |
| scanner | API integration | `internal/app/scanner_api_test.go` |
| candidates | API integration | `internal/app/scanner_api_test.go` |
| matching | API integration | `internal/app/matching_api_test.go` |
| discovery | API integration | `internal/app/discovery_wishlist_pricing_api_test.go` |
| wishlist-pricing-dashboard | API integration | `internal/app/discovery_wishlist_pricing_api_test.go`, `dashboard_api_test.go` |
| search | API integration | `internal/app/search_api_test.go` |
| data-management | API integration | `internal/app/data_import_api_test.go`, `backup_api_test.go` |
| licensing | API integration | `internal/app/license_api_test.go` |
| entitlements | API integration | `internal/app/license_api_test.go` |
| settings | API integration + E2E | planned parity suites for settings |
| integrations | API integration + E2E | planned parity suites for integrations |
| errors-logging-diagnostics | API integration | `internal/app/logging_recovery_api_test.go` |
| chat-copilot | API integration + E2E | `internal/app/chat_api_test.go`, planned ui chat suite |
| cloud-auth-billing | API integration | `internal/app/cloud_*_api_test.go`, `clerk_billing_webhook_api_test.go` |

## Screen Test Mapping (Critical-flow E2E)
All screen specs include critical UC-to-E2E mapping tables:
- `ui-screen-home`
- `ui-screen-onboarding-auth`
- `ui-screen-inventory-items`
- `ui-screen-inventory-photos`
- `ui-screen-inventory-barcodes`
- `ui-screen-inventory-ai-assist`
- `ui-screen-scanner`
- `ui-screen-discover`
- `ui-screen-reports`
- `ui-screen-settings`
- `ui-screen-chat-copilot`


