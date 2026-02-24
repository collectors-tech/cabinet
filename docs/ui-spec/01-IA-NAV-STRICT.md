# 01 IA and Navigation Strict Spec

## Scope
Defines final v1 web information architecture, navigation behavior, and screen guards.

## Top-Level Web IA (v1)
1. Home
2. Inventory
3. Discover
4. Scanner
5. Reports
6. Settings

## Inventory Sub-Navigation (tabs)
1. Items
2. Photos
3. Barcodes
4. AI Assist

## Mobile Navigation
- Pattern: left drawer.
- Top-level entries match web IA.
- Inventory tabs become segmented control at top of Inventory screen.

## Global Navigation Rules
1. Navigation labels are stable and never hidden behind icon-only controls.
2. At most one primary nav item is active (`aria-current="page"`).
3. If a section is unavailable, show disabled state with explicit reason.
4. Keyboard support: `Tab` focus order, `Enter` activate, `Esc` close drawer/modal.

## App State Guards
1. `No profile`:
- Only onboarding shell is interactive.
- Top-level nav disabled with message: `Create a profile to continue`.

2. `Profile active but onboarding incomplete`:
- Show onboarding flow only.
- Top-level nav disabled with message: `Complete onboarding to unlock workspace`.

3. `Profile active and onboarding complete`:
- Full IA enabled.

## Onboarding and Auth Placement
- Onboarding is full-screen guided flow before full workspace unlock.
- WebAuthn identity steps are part of onboarding stage, not hidden in advanced tools.

## IA Mapping from Current Feature Areas
- Dashboard -> Home
- Collection + Photos + Barcodes + AI -> Inventory
- Not In Collection panel -> Discover
- Scanner engine + failures + provider health -> Scanner
- Pricing history + exports + trends -> Reports
- Runtime, license, backups, diagnostics -> Settings

## Acceptance Criteria
- [ ] Six top-level destinations exist and are reachable from desktop and mobile.
- [ ] Inventory exposes exactly four sub-tabs.
- [ ] Disabled navigation includes explicit reason text.
- [ ] Guard transitions are deterministic across profile/onboarding states.

