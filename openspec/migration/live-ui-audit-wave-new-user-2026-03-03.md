# Live UI Audit Wave — New User (2026-03-03)

## Scope
Manual live walkthrough after fresh sign-in using `newuser+demo1@example.com`.

## Gaps requiring OpenSpec updates
1. First-run profile bootstrap contract missing/insufficient
   - Affects chats, integrations, reports, settings.
2. Users screen error-state contract for missing profile/user context
3. Inventory/Wishlist semantic model contract drift (task model leakage)
4. Sign-in UX contract lacks explicit create-account entry path
5. Settings error-state contract allows editable controls during `active_profile_404`
6. Chat header trigger/icon/panel pin contract mismatch
7. Settings nav parity expansion (Operations/Billing)

## Issue bindings
- #239 chat contract mismatch
- #258 profile bootstrap 404 fanout
- #259 inventory/wishlist semantic drift
- #260 settings submenu storage nav gap
- #261 settings operations/billing nav parity
- #262 users fetch 404
- #263 sign-in missing create-account CTA
- #264 settings editable-on-error state

## Required OpenSpec actions
- Add/extend requirements under:
  - `openspec/specs/general/ui-screen-onboarding-auth/spec.md`
  - `openspec/specs/chats/ui-screen-chat-copilot/spec.md`
  - `openspec/specs/users/ui-screen-users/spec.md`
  - `openspec/specs/integrations/ui-screen-integrations/spec.md`
  - `openspec/specs/settings/ui-screen-settings/spec.md`
  - `openspec/specs/inventory/ui-screen-inventory-items/spec.md`
  - `openspec/specs/wishlist/ui-screen-wishlist/spec.md`
  - `openspec/specs/dashboard/ui-screen-reports/spec.md`

## Evidence requirement
No requirement moved to implemented without executable Cypress/contract proof and traceability linkage.
