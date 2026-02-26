## Purpose
Define Inventory Photos screen behavior and media management use cases.

## Requirements
### Requirement: Inventory Photos screen SHALL support upload and media management lifecycle
The screen SHALL support item selection, upload, derivative view, primary selection, and delete.

#### Scenario: Use case - upload and set primary photo
- **WHEN** user uploads a photo and sets primary
- **THEN** photo list SHALL update and reflect primary state

### Requirement: Inventory Photos screen SHALL support fullscreen viewing interactions
The screen SHALL provide fullscreen media view for selected photos.

#### Scenario: Use case - inspect media fullscreen
- **WHEN** user opens fullscreen on selected photo
- **THEN** viewer SHALL render media in fullscreen mode
