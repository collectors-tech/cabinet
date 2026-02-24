# Inventory Photos Screen Spec

## Use Cases
1. User uploads item photos from file/drag-drop.
2. User captures item image from camera.
3. User sets primary image and previews full screen.

## UI Sections
1. Item selector
2. Upload controls
3. Drag/drop area
4. Camera controls
5. Photo list row actions
6. Fullscreen preview dialog

## State Behavior
- Loading: photo list loading indicator.
- Empty: "No photos yet" + upload CTA.
- Error: upload/camera/list error with recovery guidance.
- Success: photo list with primary marker.

## Acceptance Criteria
- [ ] Upload blocked when item id missing.
- [ ] Camera permission denial shows fallback instruction.
- [ ] Set Primary updates marker in current view.
- [ ] Fullscreen preview closes via button and `Esc`.

## Test Cases
- `INV-P-001` upload valid image.
- `INV-P-002` set primary action.
- `INV-P-003` delete action.
- `INV-P-004` camera denied path.

