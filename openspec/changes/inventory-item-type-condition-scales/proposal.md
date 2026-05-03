## Why

Inventory needs Category and Condition filters restored, but condition values are not universal across collector domains. Slot cars, trading cards, toys, and other item types need different condition scales, so Cabinet needs one clear source of truth for which condition values apply to an item.

## What Changes

- Introduce `Item Type` as a single-select inventory attribute that drives condition choices.
- Keep `Category` as flexible multi-select metadata for grouping, tagging, and filtering.
- Add profile/admin-managed item type condition scales seeded with Slot Cars and Trading Cards defaults.
- Scope inventory create/edit condition choices to the selected item type.
- Restore Inventory Category and Condition filters using the shared compact filter pattern.
- Preserve existing category option behavior and existing grading/default settings where possible.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `inventory-grading`: extend configurable condition enums so they can be managed per single-select item type while category remains independent flexible metadata.
- `ui-screen-inventory-items`: require Inventory forms and filters to expose item type driven condition choices and shared Category/Condition filter controls.

## Impact

- Profile settings APIs for inventory grading/taxonomy configuration.
- Inventory create/edit form state and persistence payloads.
- Inventory item/instance condition controls and filters.
- Settings screens for reusable inventory taxonomy values.
- Cypress coverage for settings persistence and inventory scoped condition behavior.
