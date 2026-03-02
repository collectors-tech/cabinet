## Purpose
Define Settings shell behavior: section navigation, deep-link routing, and shared deterministic state handling.

## Requirements
### Requirement UI-SCREEN-SETTINGS-001: Settings shell SHALL expose section-based navigation with deep-link URLs
Settings SHALL support direct navigation to section routes.

#### Scenario: Open section via URL
- **GIVEN** user opens a direct settings URL (`/settings`, `/settings/account`, `/settings/appearance`, `/settings/notifications`, `/settings/display`, `/settings/storage`)
- **WHEN** route resolves
- **THEN** corresponding section MUST render and sidebar navigation MUST reflect active section

### Requirement UI-SCREEN-SETTINGS-002: Settings shell SHALL maintain canonical section set and labels

#### Scenario: Render canonical settings sections
- **GIVEN** settings screen loads
- **WHEN** sidebar navigation renders
- **THEN** UI MUST include sections `Profile`, `Account`, `Appearance`, `Notifications`, `Display`, and `Storage` with stable route mapping

### Requirement UI-SCREEN-SETTINGS-003: Settings shell SHALL support deterministic route-level states

#### Scenario: Settings shell section load failure
- **GIVEN** section data fetch fails
- **WHEN** section is active
- **THEN** section SHALL show actionable error state without breaking entire settings route

## Notes
Detailed interaction specs are split per section:
- `settings/profile/spec.md`
- `settings/account/spec.md`
- `settings/appearance/spec.md`
- `settings/notifications/spec.md`
- `settings/display/spec.md`
- `settings/storage/spec.md`
