## Purpose
Define media upload, derivative generation, and photo interaction behavior.

## Requirements
### Requirement: Cabinet SHALL support desktop and mobile-browser photo uploads
Cabinet SHALL accept local file uploads for item media from desktop and mobile browsers.

#### Scenario: Upload photo
- **WHEN** user uploads supported image file for an item
- **THEN** Cabinet SHALL store original media locally

### Requirement: Cabinet SHALL generate thumbnail and preview derivatives
Cabinet SHALL create and maintain derived thumbnail and preview versions.

#### Scenario: Derivative generation
- **WHEN** original photo is saved
- **THEN** thumbnail and preview derivatives SHALL be generated

### Requirement: Photo ordering and primary selection SHALL be supported
Cabinet SHALL support reorder and primary-image selection per item.

#### Scenario: Set primary photo
- **WHEN** user marks photo as primary
- **THEN** item media state SHALL update primary reference

### Requirement: Fullscreen media viewing SHALL be available
Cabinet SHALL provide fullscreen media view with navigation for applicable collections.

#### Scenario: Open fullscreen media
- **WHEN** user activates fullscreen action on a photo
- **THEN** Cabinet SHALL present fullscreen viewer state
