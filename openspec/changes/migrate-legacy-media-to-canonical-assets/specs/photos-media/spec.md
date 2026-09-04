## MODIFIED Requirements

### Requirement: PHOTOS-MEDIA-001: Cabinet SHALL support desktop and mobile-browser photo uploads
Cabinet SHALL accept local file uploads for item media from desktop and mobile browsers.

#### Scenario: Migrate legacy inventory photos
- **GIVEN** an item photo row references legacy files such as `<mediaDir>/<itemID>/<photoID>_orig.ext`, `<mediaDir>/<itemID>/<photoID>_preview.jpg`, or `<mediaDir>/<itemID>/<photoID>_thumb.jpg`
- **WHEN** Cabinet applies the media migration for that active profile
- **THEN** runtime MUST create or reuse a canonical asset for the original bytes
- **AND** runtime MUST preserve the item relationship, primary-photo state, display order, filename, MIME metadata and user-visible preview/thumbnail behavior
- **AND** runtime MUST record manifest owner/provenance metadata tying the asset to the inventory item and legacy source path class
- **AND** missing or corrupt inventory photo files MUST be reported with item id, photo id, path class and recovery action while unrelated records continue.

### Requirement: PHOTOS-MEDIA-002: Cabinet SHALL generate thumbnail and preview derivatives
Cabinet SHALL create and maintain derived thumbnail and preview versions.

#### Scenario: Preserve or rebuild legacy renditions
- **GIVEN** a legacy inventory photo has existing preview or thumbnail files
- **WHEN** Cabinet migrates the original into the canonical asset folder
- **THEN** runtime MAY preserve verified compatible renditions or regenerate deterministic canonical renditions
- **AND** the manifest MUST record the final rendition paths and generator/provenance metadata
- **AND** a bad or missing legacy rendition MUST NOT block migration when the original verifies successfully and canonical renditions can be regenerated.
