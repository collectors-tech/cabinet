## Purpose
Define theme token behavior, localization, language switching, and RTL layout support.

## Requirements
### Requirement UI-FOUNDATION-THEME-RTL-I18N-001: Theme system SHALL be token-driven and profile-persistent
Cabinet SHALL support light and dark themes and density options through consistent tokenized styling.

#### Scenario: Theme persistence
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** user changes theme and reloads app
- **THEN** selected theme SHALL persist for profile

### Requirement UI-FOUNDATION-THEME-RTL-I18N-002: Localization layer SHALL support translatable shell and page labels
Cabinet SHALL resolve top-level UI labels through translation keys with safe fallback behavior.

#### Scenario: Missing translation key
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** a translation key is absent for active locale
- **THEN** UI SHALL render safe fallback text without layout break

### Requirement UI-FOUNDATION-THEME-RTL-I18N-003: RTL capability SHALL be supported at layout contract level
Cabinet SHALL support RTL layout direction behavior for shell, nav, and content alignment.

#### Scenario: RTL direction enabled
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** locale direction is RTL
- **THEN** nav and content alignment SHALL mirror according to RTL contract

### Requirement UI-FOUNDATION-THEME-RTL-I18N-004: Locale-switch controls SHALL be available in header controls
Cabinet SHALL provide language selection control in header action region.

#### Scenario: Change locale from header
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** user selects locale in header control
- **THEN** UI strings SHALL update to selected locale resources
