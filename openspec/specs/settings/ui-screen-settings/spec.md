## Purpose
Define Settings shell behavior: section navigation, deep-link routing, and shared deterministic state handling.

## Requirements
### Requirement UI-SCREEN-SETTINGS-001: Settings shell SHALL expose canonical section-based navigation with deep-link URLs
Settings SHALL support direct navigation to canonical section routes, with `/settings/profile` treated as the canonical Profile route and `/settings` treated as the entry route that redirects into it. All visible Profile navigation affordances within settings and profile menus MUST target `/settings/profile`.

#### Scenario: Open settings entry route
- **GIVEN** user opens `/settings`
- **WHEN** route resolves
- **THEN** runtime MUST redirect to `/settings/profile`
- **AND** Profile section MUST render with sidebar navigation reflecting active Profile state

#### Scenario: Open section via canonical URL
- **GIVEN** user opens a canonical settings URL (`/settings/profile`, `/settings/account`, `/settings/appearance`, `/settings/notifications`, `/settings/display`, `/settings/storage`, `/settings/operations`, `/settings/billing`)
- **WHEN** route resolves
- **THEN** corresponding section MUST render and sidebar navigation MUST reflect active section

### Requirement UI-SCREEN-SETTINGS-002: Settings shell SHALL maintain canonical section set and labels

#### Scenario: Render canonical settings sections
- **GIVEN** settings screen loads
- **WHEN** sidebar navigation renders
- **THEN** UI MUST include sections `Profile`, `Account`, `Appearance`, `Notifications`, `Display`, `Storage`, `Operations`, and `Billing` with stable route mapping

### Requirement UI-SCREEN-SETTINGS-003: Settings shell SHALL support deterministic route-level states

#### Scenario: Settings shell section load failure
- **GIVEN** section data fetch fails
- **WHEN** section is active
- **THEN** section SHALL show actionable error state without breaking entire settings route

### Requirement UI-SCREEN-SETTINGS-004: Primary navigation rail SHALL expose Storage route
The global left rail MUST expose a visible `Storage` entry as part of settings navigation affordances.

#### Scenario: Open Storage from primary rail
- **GIVEN** user is authenticated and primary left navigation is visible
- **WHEN** user inspects settings-related entries in the rail
- **THEN** rail MUST include visible `Storage` entry
- **AND** selecting `Storage` MUST route to `/settings/storage`

### Requirement UI-SCREEN-SETTINGS-005: Settings forms SHALL block edits when active profile is unavailable
When `GET /api/profiles/active` fails with missing-profile response, settings forms SHALL enter a deterministic blocked state and hide editable submit actions.

#### Scenario: Active profile missing blocks edits and shows remediation
- **GIVEN** an authenticated user opens settings routes and `GET /api/profiles/active` returns `404` with `active_profile_404`
- **WHEN** profile-backed sections (`/settings`, `/settings/notifications`) render
- **THEN** sections MUST show an explicit blocked panel with retry + `Create or Select Profile` action
- **AND** editable submit actions (`Update profile`, `Update notifications`) MUST NOT render while blocked

### Requirement UI-SCREEN-SETTINGS-006: Settings shell SHALL include Operations section route and active-state navigation
Settings shell MUST expose `/settings/operations` in sidebar and resolve deep-link loads deterministically.

#### Scenario: Open Operations section by deep link
- **GIVEN** user loads `/settings/operations`
- **WHEN** route resolves
- **THEN** Operations section MUST render
- **AND** Operations item in settings sidebar MUST be active

### Requirement UI-SCREEN-SETTINGS-007: Settings shell SHALL include Billing section route and active-state navigation
Settings shell MUST expose `/settings/billing` in sidebar and resolve deep-link loads deterministically.

#### Scenario: Open Billing section by deep link
- **GIVEN** user loads `/settings/billing`
- **WHEN** route resolves
- **THEN** Billing section MUST render
- **AND** Billing item in settings sidebar MUST be active

## Notes
Detailed interaction specs are split per section:
- `settings/profile/spec.md`
- `settings/account/spec.md`
- `settings/appearance/spec.md`
- `settings/notifications/spec.md`
- `settings/display/spec.md`
- `settings/storage/spec.md`
