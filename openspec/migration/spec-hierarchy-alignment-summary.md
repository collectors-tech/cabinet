# Spec Hierarchy Alignment Summary

Date: 2026-03-01

## Objective
Reorganized OpenSpec capabilities from a flat layout into UI section hierarchy folders:
`general`, `dashboard`, `inventory`, `wishlist`, `integrations`, `chats`, `users`, `settings`, `helpcenter`.

## What Changed
1. Moved flat spec folders into section sub-spec folders while keeping requirement IDs unchanged.
2. Kept section-root specs in place where they already existed:
- `openspec/specs/integrations/spec.md`
- `openspec/specs/settings/spec.md`
3. Updated references in OpenSpec artifacts and traceability to new paths.
4. Added section READMEs with purpose, contained sub-specs, and ID namespace groupings.
5. Updated `openspec/specs/README.md` to describe the new hierarchy.

## Section Mapping (by destination)
- `general`: auth, cloud-auth-billing, data-management, diagnostics, documentation-governance, entitlements, errors, future-hooks, licensing, logging, non-functional, runtime-core, security, ui-data-contract-parity, ui-foundation-accessibility, ui-foundation-auth-menus-shortcuts, ui-foundation-components, ui-foundation-interactions, ui-foundation-shell-navigation, ui-foundation-theme-rtl-i18n, ui-governance-gates, ui-performance, ui-scale, ui-screen-onboarding-auth, ui-semantic-component-layer
- `dashboard`: discovery, ui-screen-discover, ui-screen-home, ui-screen-reports
- `inventory`: barcodes, collection-domain, lookup, matching, photos-media, search, ui-screen-inventory-ai-assist, ui-screen-inventory-barcodes, ui-screen-inventory-items, ui-screen-inventory-photos
- `wishlist`: wishlist-pricing-dashboard
- `integrations`: candidates, provider-amazon, provider-au-webshops, provider-ebay, provider-registry, scanner, ui-screen-scanner (+ existing section root spec)
- `chats`: ai-assist, chat-copilot, ui-screen-chat-copilot
- `users`: profiles
- `settings`: ui-screen-settings (+ existing section root spec)
- `helpcenter`: section README scaffold (no sub-specs yet)

## Validation
Command run:
- `openspec validate --all`

Result:
- `Totals: 5 passed, 0 failed (5 items)`

## Coverage/ID Safety
- All existing requirement IDs preserved unchanged.
- No missing spec path references detected for `openspec/specs/**/spec.md` paths.

## Notes
- Existing empty `openspec/specs/ui/*` placeholder directories remain non-normative and contain no spec files.
