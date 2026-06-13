## Purpose
Define deterministic route-to-spec coverage for authenticated Cabinet screens.

## Requirements
### Requirement UI-ROUTE-COVERAGE-001: Every authenticated route SHALL map to at least one screen spec
Cabinet SHALL maintain explicit route coverage so each user-facing route is governed by a corresponding UI spec.

#### Scenario: Validate authenticated route coverage
- **GIVEN** authenticated routes include dashboard, inventory, wishlist, integrations, chats, users, help-center, settings sections, and errors
- **WHEN** route coverage is reviewed
- **THEN** each route MUST map to at least one `ui-screen-*` spec (or explicit placeholder spec)

### Requirement UI-ROUTE-COVERAGE-002: Settings section routes SHALL be split one-spec-per-screen
Cabinet SHALL model settings routes with separate specs per section screen.

#### Scenario: Settings screen split validation
- **GIVEN** settings routes include `/settings`, `/settings/account`, `/settings/appearance`, `/settings/notifications`, `/settings/display`, `/settings/storage`
- **WHEN** settings specs are reviewed
- **THEN** each route MUST have a dedicated settings screen spec and shell/navigation behavior MUST remain in shared settings-shell spec

### Requirement UI-ROUTE-COVERAGE-003: Placeholder routes SHALL have explicit placeholder specs
Cabinet SHALL not leave placeholder routes undocumented.

#### Scenario: Help-center placeholder route
- **GIVEN** `/help-center` is intentionally placeholder/coming-soon
- **WHEN** route contract is evaluated
- **THEN** spec MUST define deterministic placeholder behavior, shell controls, and transition expectation for future full implementation

### Requirement UI-ROUTE-COVERAGE-004: Protected routes SHALL redirect unauthenticated users with deep-link state
Cabinet SHALL guard authenticated routes by sending signed-out users to sign-in while preserving the requested route and query string as a safe redirect target.

#### Scenario: Unauthenticated protected route access
- **GIVEN** a signed-out user opens an authenticated Cabinet route with query state
- **WHEN** route guards evaluate the request
- **THEN** Cabinet MUST show the sign-in route and preserve the original path/query in the redirect search parameter

### Requirement UI-ROUTE-COVERAGE-005: Route error pages SHALL expose deterministic recovery actions
Cabinet SHALL render deterministic public and authenticated route-error pages with user-readable status copy and non-destructive navigation recovery.

#### Scenario: Public unknown route recovery
- **GIVEN** a user opens an unknown public route
- **WHEN** the not-found boundary renders
- **THEN** Cabinet MUST show the 404 state with Go Back and Back to Home recovery controls

#### Scenario: Authenticated error route taxonomy
- **GIVEN** an authenticated user opens a known error taxonomy route
- **WHEN** the error screen renders
- **THEN** Cabinet MUST show the matching error status, readable guidance, and safe recovery controls inside the authenticated shell

### Requirement UI-ROUTE-COVERAGE-006: Document titles SHALL remain product-prefixed across route classes
Cabinet SHALL use deterministic `Cabinet - <Page Title>` document titles for authenticated workspace routes and documented route-error surfaces.

#### Scenario: Route title coverage
- **GIVEN** authenticated workspace and route-error screens are opened
- **WHEN** navigation completes
- **THEN** `document.title` MUST use the Cabinet product prefix and the route-specific page title without leaking route module names

## Route Coverage Map
| Route | Primary Spec |
| --- | --- |
| `/` | `openspec/specs/dashboard/ui-screen-home/spec.md` |
| `/inventory` | `openspec/specs/inventory/ui-screen-inventory-items/spec.md` |
| `/wishlist` | `openspec/specs/wishlist/ui-screen-wishlist/spec.md` |
| `/integrations` | `openspec/specs/integrations/ui-screen-integrations/spec.md` |
| `/chats` | `openspec/specs/chats/ui-screen-chat-copilot/spec.md` |
| `/users` | `openspec/specs/users/ui-screen-users/spec.md` |
| `/help-center` | `openspec/specs/helpcenter/ui-screen-help-center/spec.md` |
| `/settings` | `openspec/specs/settings/profile/spec.md` |
| `/settings/account` | `openspec/specs/settings/account/spec.md` |
| `/settings/appearance` | `openspec/specs/settings/appearance/spec.md` |
| `/settings/notifications` | `openspec/specs/settings/notifications/spec.md` |
| `/settings/display` | `openspec/specs/settings/display/spec.md` |
| `/settings/storage` | `openspec/specs/settings/storage/spec.md` |
| `/errors/$error` | `openspec/specs/general/errors/spec.md` |
