## Purpose
Define theme token behavior, localization, language switching, and RTL layout support.

## Requirements
### Requirement: Theme system SHALL be token-driven and profile-persistent
Cabinet SHALL support light and dark themes and density options through consistent tokenized styling.

#### Scenario: Theme persistence
- **WHEN** user changes theme and reloads app
- **THEN** selected theme SHALL persist for profile

### Requirement: Localization layer SHALL support translatable shell and page labels
Cabinet SHALL resolve top-level UI labels through translation keys with safe fallback behavior.

#### Scenario: Missing translation key
- **WHEN** a translation key is absent for active locale
- **THEN** UI SHALL render safe fallback text without layout break

### Requirement: RTL capability SHALL be supported at layout contract level
Cabinet SHALL support RTL layout direction behavior for shell, nav, and content alignment.

#### Scenario: RTL direction enabled
- **WHEN** locale direction is RTL
- **THEN** nav and content alignment SHALL mirror according to RTL contract

### Requirement: Locale-switch controls SHALL be available in header controls
Cabinet SHALL provide language selection control in header action region.

#### Scenario: Change locale from header
- **WHEN** user selects locale in header control
- **THEN** UI strings SHALL update to selected locale resources
