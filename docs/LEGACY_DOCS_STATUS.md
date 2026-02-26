# Legacy Docs Status (OpenSpec Migration Lock-In)

Last updated: 2026-02-26  
Tracking issue: `#171`

## Policy
- `MIGRATED_TO_OPENSPEC`: requirements moved; this file is non-normative context.
- `REFERENCE_ONLY`: operational/support content, not requirement source.
- `SUPERSEDED_BY_OPENSPEC`: retained for history; do not implement from this file.

## Root Docs
| File | Status | OpenSpec Mapping / Notes |
| --- | --- | --- |
| `docs/FULL_FEATURE_LIST.md` | MIGRATED_TO_OPENSPEC | Covered across capability specs and UI screen specs |
| `docs/SPEC.md` | MIGRATED_TO_OPENSPEC | Covered by baseline OpenSpec capabilities |
| `docs/USE_CASES_AND_SCENARIOS.md` | MIGRATED_TO_OPENSPEC | UC mapping normalized in OpenSpec and test matrix |
| `docs/APP_COMPLETION_ANALYSIS.md` | SUPERSEDED_BY_OPENSPEC | Converted into governance/spec deltas |
| `docs/ARCHITECTURE.md` | REFERENCE_ONLY | Engineering reference; behavior contracts live in OpenSpec |
| `docs/FORM_ARCHITECTURE.md` | MIGRATED_TO_OPENSPEC | Covered by foundation/screen specs and component contracts |
| `docs/UI_ENDPOINT_PARITY.md` | MIGRATED_TO_OPENSPEC | Covered by `ui-data-contract-parity` + screen specs |
| `docs/UI_INTUITIVE_PLANNING.md` | SUPERSEDED_BY_OPENSPEC | Migrated to UI foundation and screen specs |
| `docs/ROADMAP_90_DAYS.md` | REFERENCE_ONLY | Delivery planning artifact |
| `docs/SHOP_PROVIDERS.md` | MIGRATED_TO_OPENSPEC | Covered in integrations/scanner/provider requirements |
| `docs/MARKETING.md` | REFERENCE_ONLY | Non-technical collateral |
| `docs/OPENSPEC_WORKFLOW.md` | REFERENCE_ONLY | Process guide |
| `docs/OPENSPEC_MIGRATION_CATALOG.md` | REFERENCE_ONLY | Migration record |
| `docs/OPENSPEC_MIGRATION_TODO.md` | REFERENCE_ONLY | Migration completion checklist |
| `docs/OPENSPEC_TEST_MATRIX.md` | REFERENCE_ONLY | UC/test traceability report |

## UI Spec Docs
| File | Status | OpenSpec Mapping / Notes |
| --- | --- | --- |
| `docs/ui-spec/01-IA-NAV-STRICT.md` | MIGRATED_TO_OPENSPEC | `ui-foundation-shell-navigation` |
| `docs/ui-spec/02-SCREEN-SPECS.md` | MIGRATED_TO_OPENSPEC | `ui-screen-*` specs |
| `docs/ui-spec/03-DASHBOARD-ATTENTION-STRICT.md` | MIGRATED_TO_OPENSPEC | home/dashboard screen + wishlist/pricing/dashboard specs |
| `docs/ui-spec/04-DATA-CONTRACTS-UI.md` | MIGRATED_TO_OPENSPEC | `ui-data-contract-parity` |
| `docs/ui-spec/05-TEST-MATRIX-UI.md` | MIGRATED_TO_OPENSPEC | `docs/OPENSPEC_TEST_MATRIX.md` |
| `docs/ui-spec/06-SCREEN-DETAIL-SPECS.md` | MIGRATED_TO_OPENSPEC | `ui-screen-*` specs |
| `docs/ui-spec/07-SCALABILITY-DATA-PLAN.md` | MIGRATED_TO_OPENSPEC | `ui-scale-and-performance` |
| `docs/ui-spec/08-GAP-AND-INTUITIVENESS-REVIEW.md` | MIGRATED_TO_OPENSPEC | UI governance + semantic layer requirements |
| `docs/ui-spec/09-COMPONENT-SPECS-STRICT.md` | MIGRATED_TO_OPENSPEC | `ui-foundation-components` |
| `docs/ui-spec/10-COMPONENT-CONTRACT-IMPLEMENTATION-MAP.md` | MIGRATED_TO_OPENSPEC | foundation/spec mapping in OpenSpec |
| `docs/ui-spec/10-SCREEN-TEMPLATES-SEMANTIC-TREE.md` | MIGRATED_TO_OPENSPEC | semantic layer + foundation specs |
| `docs/ui-spec/11-REFERENCE-PATTERN-MAPPING.md` | MIGRATED_TO_OPENSPEC | normalized into component/foundation requirements |
| `docs/ui-spec/12-PERF-VALIDATION-S2-S3.md` | MIGRATED_TO_OPENSPEC | `ui-scale-and-performance` |
| `docs/ui-spec/13-UI-UX-STRATEGY-GATE.md` | MIGRATED_TO_OPENSPEC | `ui-governance-gates` |
| `docs/ui-spec/14-SEMANTIC-COMPONENT-LAYER.md` | MIGRATED_TO_OPENSPEC | `ui-semantic-component-layer` |
| `docs/ui-spec/15-TEMPLATE-SECTION-NAMES-AND-MAPPING.md` | MIGRATED_TO_OPENSPEC | foundation/screen naming contracts |
| `docs/ui-spec/README.md` | REFERENCE_ONLY | legacy index |
| `docs/ui-spec/screens/*.md` | MIGRATED_TO_OPENSPEC | consolidated into `openspec/specs/ui-screen-*/spec.md` |

## Feature Docs
| Path | Status | OpenSpec Mapping / Notes |
| --- | --- | --- |
| `docs/features/*.md` | MIGRATED_TO_OPENSPEC | Capability specs under `openspec/specs/*` |
| `docs/features/README.md` | REFERENCE_ONLY | legacy index |
| `docs/features/SPEC_TEMPLATE.md` | REFERENCE_ONLY | historical template only |

## API/Auth Docs
| File | Status | OpenSpec Mapping / Notes |
| --- | --- | --- |
| `docs/api/openapi.yaml` | REFERENCE_ONLY | API contract source; validated for parity |
| `docs/api/redoc.html` | REFERENCE_ONLY | generated rendering asset |
| `docs/api/README.md` | REFERENCE_ONLY | usage notes |
| `docs/auth/CLERK_BILLING_SETUP.md` | MIGRATED_TO_OPENSPEC | `cloud-auth-billing` spec; doc retained as runbook |

