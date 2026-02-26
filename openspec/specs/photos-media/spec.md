## Purpose
Define media upload, derivative generation, and photo interaction behavior.

## Requirements
### Requirement PHOTOS-MEDIA-001: Cabinet SHALL support desktop and mobile-browser photo uploads
Cabinet SHALL accept local file uploads for item media from desktop and mobile browsers.

#### Scenario: Upload photo
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** user uploads supported image file for an item
- **THEN** Cabinet SHALL store original media locally

### Requirement PHOTOS-MEDIA-002: Cabinet SHALL generate thumbnail and preview derivatives
Cabinet SHALL create and maintain derived thumbnail and preview versions.

#### Scenario: Derivative generation
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** original photo is saved
- **THEN** thumbnail and preview derivatives SHALL be generated

### Requirement PHOTOS-MEDIA-003: Photo ordering and primary selection SHALL be supported
Cabinet SHALL support reorder and primary-image selection per item.

#### Scenario: Set primary photo
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** user marks photo as primary
- **THEN** item media state SHALL update primary reference

### Requirement PHOTOS-MEDIA-004: Fullscreen media viewing SHALL be available
Cabinet SHALL provide fullscreen media view with navigation for applicable collections.

#### Scenario: Open fullscreen media
- **GIVEN** an authenticated actor with the required role is operating an active local profile, required capability configuration is enabled, and scenario fixture data exists for execution
- **WHEN** user activates fullscreen action on a photo
- **THEN** Cabinet SHALL present fullscreen viewer state
