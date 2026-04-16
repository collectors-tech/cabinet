## Purpose
Define global login and session-entry behavior that is not tied to a single feature section.

## Requirements
### Requirement UI-LOGIN-SESSION-001: Login routes SHALL gate access to authenticated workspace routes
Cabinet SHALL redirect unauthenticated users to login and preserve intended destination for post-login return.

#### Scenario: Redirect to sign-in with return target
- **GIVEN** user is unauthenticated and requests an authenticated route
- **WHEN** router resolves route guards
- **THEN** UI MUST redirect to sign-in and preserve redirect target for successful session bootstrap

### Requirement UI-LOGIN-SESSION-002: Login state SHALL support deterministic error and retry behavior
Cabinet SHALL surface actionable auth failure states and allow retry without hard refresh.

#### Scenario: Invalid credential or auth bootstrap failure
- **GIVEN** sign-in request fails due to invalid credential or auth bootstrap error
- **WHEN** login form submits
- **THEN** UI MUST display inline error guidance and keep form state available for retry

### Requirement UI-LOGIN-SESSION-003: Session entry SHALL support profile-aware activation
Cabinet SHALL support selecting/activating profile context after successful authentication when multiple profiles exist.

#### Scenario: Select profile context after login
- **GIVEN** authenticated session exists and profile list has at least two profiles
- **WHEN** user selects an active profile
- **THEN** workspace context MUST switch to selected profile and subsequent API calls MUST use active profile scope

### Requirement UI-LOGIN-SESSION-004: Session entry SHALL avoid active-profile missing errors on first-run core screens
First-run signed-in sessions SHALL resolve a usable active profile context so core routes render without `active_profile_404`/`active_profile_not_set` failures.

#### Scenario: First-run core route sweep
- **GIVEN** a newly signed-in first-run user and no previously selected active profile context
- **WHEN** user navigates core routes (`/settings/display`, `/chats`, `/integrations`, `/reports`, `/users`)
- **THEN** routes MUST render usable states without surfacing raw `active_profile_404` or `active_profile_not_set` errors

### Requirement UI-LOGIN-SESSION-005: Root unauthenticated entry SHALL use a clean sign-in route while preserving deep-link returns
Cabinet SHALL avoid attaching a redundant `redirect=%2F` query when an unauthenticated user opens the base app entry, while still preserving redirect targets for protected deep links.

#### Scenario: Root entry redirects cleanly to sign-in
- **GIVEN** user is unauthenticated and requests the base app URL `/`
- **WHEN** router resolves the unauthenticated entry redirect
- **THEN** UI MUST land on `/sign-in` without a redundant `redirect=%2F` query
- **AND** sign-in from that entry MUST continue to the canonical dashboard destination `/dashboard`

#### Scenario: Protected deep link still preserves return target
- **GIVEN** user is unauthenticated and requests a protected deep link such as `/inventory/`
- **WHEN** router resolves the unauthenticated entry redirect
- **THEN** UI MUST land on `/sign-in` with the intended protected return target preserved in search state

### Requirement UI-LOGIN-SESSION-006: Session exit SHALL clear auth state and return users to sign-in
Cabinet SHALL provide a concrete `/sign-out` route that resets local auth state and redirects users to sign-in so protected routes no longer remain reachable under the prior session.

#### Scenario: Direct sign-out route clears session and re-gates protected routes
- **GIVEN** an authenticated local session and a reachable protected route
- **WHEN** user visits `/sign-out`
- **THEN** Cabinet MUST clear local auth state, redirect to `/sign-in`, and re-gate the next protected-route request through sign-in instead of rendering the prior authenticated workspace

### Requirement UI-LOGIN-SESSION-007: Dashboard SHALL not remain reachable after sign-out
After a local sign-out, Cabinet SHALL re-apply authenticated-route gating to the canonical dashboard entry so the previous authenticated dashboard shell cannot be reopened without signing in again.

#### Scenario: Sign-out removes dashboard access until re-authenticated
- **GIVEN** an authenticated local session currently able to reach `/dashboard`
- **WHEN** the user signs out and then requests `/dashboard` again
- **THEN** Cabinet MUST redirect to `/sign-in`
- **AND** the previous authenticated dashboard content MUST not render until the user signs in again
