# OpenSpec Migration Catalog (Finalized)

Last updated: 2026-02-26  
Tracking issue: `#171`

## Purpose
Record migration status for legacy docs and lock the source of truth.

## Source of Truth Policy
OpenSpec (`openspec/specs/*/spec.md`) is the normative source of product, UI, and behavior requirements.

Legacy docs under `docs/` are retained only as:
- historical context
- implementation notes
- operational references (API docs, auth setup, workflow guidance)

## Status Legend
- `MIGRATED`: normative requirements moved to OpenSpec specs
- `REFERENCE_ONLY`: non-normative support material
- `SUPERSEDED`: kept for history; replaced by OpenSpec coverage

## Baseline OpenSpec Coverage
- Runtime/Core: `openspec/specs/runtime-core/spec.md`
- Auth/Profiles/WebAuthn: `openspec/specs/auth-and-profiles/spec.md`
- Collection Domain: `openspec/specs/collection-domain/spec.md`
- Photos/Media: `openspec/specs/photos-media/spec.md`
- Barcodes: `openspec/specs/barcodes-and-lookup/spec.md`
- AI Assist: `openspec/specs/ai-assist/spec.md`
- Scanner/Candidates: `openspec/specs/scanner-and-candidates/spec.md`
- Matching/Discovery: `openspec/specs/matching-and-discovery/spec.md`
- Wishlist/Pricing/Dashboard: `openspec/specs/wishlist-pricing-dashboard/spec.md`
- Search/Data Management: `openspec/specs/search-and-data-management/spec.md`
- Licensing: `openspec/specs/licensing-and-entitlements/spec.md`
- Settings/Integrations: `openspec/specs/settings-and-integrations/spec.md`
- Logging/Diagnostics: `openspec/specs/errors-logging-diagnostics/spec.md`
- Future Hooks: `openspec/specs/future-hooks/spec.md`
- Non-Functional/Security: `openspec/specs/non-functional-and-security/spec.md`
- Chat Copilot: `openspec/specs/chat-copilot/spec.md`
- UI Foundation:
  - `openspec/specs/ui-foundation-shell-navigation/spec.md`
  - `openspec/specs/ui-foundation-auth-menus-shortcuts/spec.md`
  - `openspec/specs/ui-foundation-theme-rtl-i18n/spec.md`
  - `openspec/specs/ui-foundation-interactions-and-accessibility/spec.md`
  - `openspec/specs/ui-foundation-components/spec.md`
  - `openspec/specs/ui-semantic-component-layer/spec.md`
- UI Screen Specs:
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
- Governance/Parity/Scale:
  - `openspec/specs/ui-data-contract-parity/spec.md`
  - `openspec/specs/ui-scale-and-performance/spec.md`
  - `openspec/specs/ui-governance-gates/spec.md`
  - `openspec/specs/cloud-auth-billing/spec.md`

## Legacy Docs Disposition
See `docs/LEGACY_DOCS_STATUS.md` for per-file migration and retention status.

