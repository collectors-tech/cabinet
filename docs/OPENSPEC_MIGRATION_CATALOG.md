# OpenSpec Migration Catalog (Docs -> OpenSpec)

Last updated: 2026-02-26  
Tracking issue: `#160`

## Purpose
Catalog every spec-like document under `docs/`, map its OpenSpec migration status, and identify remaining migration work.

## Status Legend
- `MIGRATED`: covered by one or more `openspec/specs/*/spec.md` files
- `PARTIAL`: partially covered; requires additional OpenSpec spec detail
- `PENDING`: not yet migrated into OpenSpec baseline specs
- `REFERENCE_ONLY`: non-spec operational/marketing artifact; not targeted for OpenSpec migration

## OpenSpec Baseline Specs (Current)
- `openspec/specs/runtime-core/spec.md`
- `openspec/specs/auth-and-profiles/spec.md`
- `openspec/specs/collection-domain/spec.md`
- `openspec/specs/photos-media/spec.md`
- `openspec/specs/barcodes-and-lookup/spec.md`
- `openspec/specs/ai-assist/spec.md`
- `openspec/specs/scanner-and-candidates/spec.md`
- `openspec/specs/matching-and-discovery/spec.md`
- `openspec/specs/wishlist-pricing-dashboard/spec.md`
- `openspec/specs/search-and-data-management/spec.md`
- `openspec/specs/licensing-and-entitlements/spec.md`
- `openspec/specs/settings-and-integrations/spec.md`
- `openspec/specs/errors-logging-diagnostics/spec.md`
- `openspec/specs/future-hooks/spec.md`
- `openspec/specs/non-functional-and-security/spec.md`
- `openspec/specs/chat-copilot/spec.md`
- `openspec/specs/ui-foundation-shell-navigation/spec.md`
- `openspec/specs/ui-foundation-auth-menus-shortcuts/spec.md`
- `openspec/specs/ui-foundation-theme-rtl-i18n/spec.md`
- `openspec/specs/ui-foundation-interactions-and-accessibility/spec.md`
- `openspec/specs/ui-screen-home/spec.md`
- `openspec/specs/ui-screen-onboarding-auth/spec.md`
- `openspec/specs/ui-screen-inventory-items/spec.md`
- `openspec/specs/ui-screen-inventory-photos/spec.md`
- `openspec/specs/ui-screen-inventory-barcodes/spec.md`
- `openspec/specs/ui-screen-inventory-ai-assist/spec.md`
- `openspec/specs/ui-screen-scanner/spec.md`
- `openspec/specs/ui-screen-discover/spec.md`
- `openspec/specs/ui-screen-reports/spec.md`
- `openspec/specs/ui-screen-settings/spec.md`
- `openspec/specs/ui-screen-chat-copilot/spec.md`

## Catalog: Root `docs/`
| Source Doc | Type | OpenSpec Coverage | Status | Action |
| --- | --- | --- | --- | --- |
| `docs/FULL_FEATURE_LIST.md` | feature baseline | multiple capability specs | MIGRATED | keep as human-readable source index |
| `docs/SPEC.md` | product intent | runtime/auth/collection/dashboard/cross-domain specs | MIGRATED | maintain alignment during future deltas |
| `docs/USE_CASES_AND_SCENARIOS.md` | cross-feature use cases | capability + screen specs | PARTIAL | add formal UC IDs + scenario mapping into OpenSpec |
| `docs/APP_COMPLETION_ANALYSIS.md` | execution blueprint | represented via seeded OpenSpec changes | PARTIAL | convert remaining execution gates into OpenSpec governance spec/change |
| `docs/ARCHITECTURE.md` | technical architecture | distributed across capabilities | PARTIAL | add dedicated architecture-contract spec if behaviorally normative |
| `docs/FORM_ARCHITECTURE.md` | forms architecture | not directly represented | PENDING | add `ui-foundation-forms` OpenSpec spec |
| `docs/UI_ENDPOINT_PARITY.md` | route/endpoint mapping | parity change + screen specs | PARTIAL | add generated parity contract spec with requirement-level endpoint assertions |
| `docs/UI_INTUITIVE_PLANNING.md` | UX strategy | dashboard + ui-foundation specs | PARTIAL | migrate attention model detail to dedicated `ui-screen-dashboard-attention` spec delta |
| `docs/ROADMAP_90_DAYS.md` | delivery roadmap | not a runtime behavior spec | REFERENCE_ONLY | keep as planning artifact |
| `docs/SHOP_PROVIDERS.md` | provider catalog | settings/integrations + scanner specs | PARTIAL | add provider capability contract (availability/stock scraping/limits) |
| `docs/MARKETING.md` | marketing | n/a | REFERENCE_ONLY | no OpenSpec migration needed |
| `docs/OPENSPEC_WORKFLOW.md` | process | n/a | REFERENCE_ONLY | keep updated |

## Catalog: `docs/features/`
| Source Folder | OpenSpec Coverage | Status | Action |
| --- | --- | --- | --- |
| `docs/features/01-application-core.md` to `docs/features/23-in-app-chat-copilot.md` | mapped to corresponding capability specs | MIGRATED | maintain 1:1 references from features to OpenSpec spec IDs |
| `docs/features/README.md` | reference index | REFERENCE_ONLY | keep |
| `docs/features/SPEC_TEMPLATE.md` | template | REFERENCE_ONLY | optionally replace with OpenSpec template usage note |

## Catalog: `docs/ui-spec/` (Foundation)
| Source Doc | OpenSpec Coverage | Status | Action |
| --- | --- | --- | --- |
| `docs/ui-spec/01-IA-NAV-STRICT.md` | `ui-foundation-shell-navigation` | PARTIAL | add explicit nav edit persistence + mobile drawer contracts |
| `docs/ui-spec/02-SCREEN-SPECS.md` | screen specs set | PARTIAL | migrate any missing screen section definitions |
| `docs/ui-spec/03-DASHBOARD-ATTENTION-STRICT.md` | `ui-screen-home` + `wishlist-pricing-dashboard` | PARTIAL | add UC IDs and test mapping |
| `docs/ui-spec/04-DATA-CONTRACTS-UI.md` | parity + capability specs | PARTIAL | add endpoint-by-endpoint SHALL requirements |
| `docs/ui-spec/05-TEST-MATRIX-UI.md` | none formalized | PENDING | migrate to OpenSpec testability matrix artifact |
| `docs/ui-spec/06-SCREEN-DETAIL-SPECS.md` | screen specs | PARTIAL | merge missing details into `ui-screen-*` specs |
| `docs/ui-spec/07-SCALABILITY-DATA-PLAN.md` | `non-functional-and-security` partial | PARTIAL | add dedicated `ui-scale-and-performance` spec |
| `docs/ui-spec/08-GAP-AND-INTUITIVENESS-REVIEW.md` | open change proposals | PARTIAL | convert key findings to requirements/deltas |
| `docs/ui-spec/09-COMPONENT-SPECS-STRICT.md` | ui foundation partial | PARTIAL | add `ui-foundation-components` spec |
| `docs/ui-spec/10-COMPONENT-CONTRACT-IMPLEMENTATION-MAP.md` | none formalized | PENDING | migrate mapping into requirements + task traceability |
| `docs/ui-spec/10-SCREEN-TEMPLATES-SEMANTIC-TREE.md` | ui foundation partial | PARTIAL | merge remaining semantic tree constraints |
| `docs/ui-spec/11-REFERENCE-PATTERN-MAPPING.md` | ui foundation partial | PARTIAL | retain only neutral patterns; migrate normative parts |
| `docs/ui-spec/12-PERF-VALIDATION-S2-S3.md` and `.json` | non-functional partial | PARTIAL | move benchmark assertions into OpenSpec requirements |
| `docs/ui-spec/13-UI-UX-STRATEGY-GATE.md` | none formalized | PENDING | migrate into `ui-governance-gates` spec |
| `docs/ui-spec/14-SEMANTIC-COMPONENT-LAYER.md` | ui foundation partial | PARTIAL | migrate semantic layer requirement blocks |
| `docs/ui-spec/15-TEMPLATE-SECTION-NAMES-AND-MAPPING.md` | auth menu/nav/interaction specs | PARTIAL | migrate remaining labels and lifecycle policies |
| `docs/ui-spec/README.md` | reference index | REFERENCE_ONLY | keep as index |

## Catalog: `docs/ui-spec/screens/` (Per Screen)
| Source Screen Doc | OpenSpec Screen Spec | Status | Action |
| --- | --- | --- | --- |
| `docs/ui-spec/screens/home.md` | `openspec/specs/ui-screen-home/spec.md` | MIGRATED | add UC IDs + Cypress mapping |
| `docs/ui-spec/screens/onboarding-auth.md` | `openspec/specs/ui-screen-onboarding-auth/spec.md` | MIGRATED | add UC IDs + Cypress mapping |
| `docs/ui-spec/screens/inventory-items.md` | `openspec/specs/ui-screen-inventory-items/spec.md` | MIGRATED | add UC IDs + Cypress mapping |
| `docs/ui-spec/screens/inventory-photos.md` | `openspec/specs/ui-screen-inventory-photos/spec.md` | MIGRATED | add UC IDs + Cypress mapping |
| `docs/ui-spec/screens/inventory-barcodes.md` | `openspec/specs/ui-screen-inventory-barcodes/spec.md` | MIGRATED | add UC IDs + Cypress mapping |
| `docs/ui-spec/screens/inventory-ai-assist.md` | `openspec/specs/ui-screen-inventory-ai-assist/spec.md` | MIGRATED | add UC IDs + Cypress mapping |
| `docs/ui-spec/screens/scanner.md` | `openspec/specs/ui-screen-scanner/spec.md` | MIGRATED | add UC IDs + Cypress mapping |
| `docs/ui-spec/screens/discover.md` | `openspec/specs/ui-screen-discover/spec.md` | MIGRATED | add UC IDs + Cypress mapping |
| `docs/ui-spec/screens/reports.md` | `openspec/specs/ui-screen-reports/spec.md` | MIGRATED | add UC IDs + Cypress mapping |
| `docs/ui-spec/screens/settings.md` | `openspec/specs/ui-screen-settings/spec.md` | MIGRATED | add UC IDs + Cypress mapping |
| `docs/ui-spec/screens/chat-copilot.md` | `openspec/specs/ui-screen-chat-copilot/spec.md` | MIGRATED | add UC IDs + Cypress mapping |
| `docs/ui-spec/screens/README.md` | reference index | REFERENCE_ONLY | keep |

## Catalog: Other `docs/` Subfolders
| Source Doc | Type | Status | Action |
| --- | --- | --- | --- |
| `docs/api/openapi.yaml` | API contract source | REFERENCE_ONLY | remains source of API contract; enforce parity checks |
| `docs/api/README.md` | api docs usage | REFERENCE_ONLY | keep |
| `docs/api/redoc.html` | generated/static docs | REFERENCE_ONLY | keep generated |
| `docs/auth/CLERK_BILLING_SETUP.md` | operational auth integration guide | PARTIAL | add `cloud-auth-billing` OpenSpec spec if billing behavior is normative |

## Migration Gaps (Priority)
1. Formal testability migration:
- All `ui-screen-*` specs need explicit `UC-<screen>-<n>` scenario IDs and Cypress mapping references.
2. UI component contract migration:
- `09-COMPONENT-SPECS-STRICT.md` and semantic layer docs need dedicated OpenSpec capability specs.
3. Performance and gate migration:
- Perf validation and UX strategy gate docs need normative OpenSpec requirements.
4. Data-contract parity migration:
- Endpoint-level UI parity requirements need direct OpenSpec contract definitions.
