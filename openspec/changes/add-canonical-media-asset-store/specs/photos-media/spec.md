## MODIFIED Requirements

### Requirement: PHOTOS-MEDIA-001: Cabinet SHALL support desktop and mobile-browser photo uploads
Cabinet SHALL accept local file uploads for item media from desktop and mobile browsers.

#### Scenario: Upload photo
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** user uploads supported image file for an item
- **THEN** Cabinet SHALL store original media locally
- **AND** new uploads SHALL be promoted atomically into `<media-root>/assets/<asset-id>/` with `original/`, `renditions/`, `variations/` and `manifest.json`
- **AND** the original bytes SHALL be recorded as immutable in the manifest and SHALL NOT be overwritten by rendition or variation generation

### Requirement: PHOTOS-MEDIA-002: Cabinet SHALL generate thumbnail and preview derivatives
Cabinet SHALL create and maintain derived thumbnail and preview versions.

#### Scenario: Derivative generation
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** original photo is saved
- **THEN** thumbnail and preview derivatives SHALL be generated
- **AND** generated derivatives SHALL use deterministic names under `renditions/`
- **AND** the manifest SHALL record each derivative path, generator and generator version so derivatives can be rebuilt without untracked files
