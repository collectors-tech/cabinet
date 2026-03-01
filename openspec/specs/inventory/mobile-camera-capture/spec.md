## Purpose
Define mobile camera capture behavior for item photo intake.

## Requirements
### Requirement MOBILE-CAMERA-001: Inventory photo intake MUST support direct mobile camera capture
Cabinet SHALL support taking item photos from a phone browser camera and attaching them to an item.

#### Scenario: Capture from phone camera
- **GIVEN** a signed-in user opens Inventory Photos on a mobile browser and grants camera permission
- **WHEN** the user selects `Take Photo` and captures an image
- **THEN** upload MUST complete with `201` and create media records:
  - `photo_id`
  - `item_id`
  - `original_path`
  - `thumbnail_path`
  - `preview_path`

### Requirement MOBILE-CAMERA-002: Camera permission failure MUST provide deterministic fallback
Cabinet SHALL provide clear fallback when camera capture is unavailable or denied.

#### Scenario: Permission denied fallback
- **GIVEN** camera access is denied by browser or unavailable on device
- **WHEN** the user attempts `Take Photo`
- **THEN** UI MUST show an actionable message and provide `Upload File` fallback without losing current item context
