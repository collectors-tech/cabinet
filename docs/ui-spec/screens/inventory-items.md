# Inventory Items Screen Spec

## Use Cases
1. User finds items quickly by search/filter.
2. User creates canonical items with minimal fields.
3. User manages instances without leaving inventory context.

## UI Sections
1. Search/Filter/Sort bar
2. Collection Browser split layout (Folder tree pane + Results pane)
3. Summary strip (folders/items/quantity/value)
4. Item list/table or card grid
5. Item details panel
6. Quick Add form
7. Advanced metadata form
8. Instances list + form
9. Bulk action bar (visible when selection > 0)

## State Behavior
- Loading: list skeleton.
- Empty: "No items found" + `Add Item`.
- Error: item load/create error + retry.
- Success: list + details + instance panels.

## Acceptance Criteria
- [ ] Search supports title and part number.
- [ ] Quick Add requires only minimum fields (part number, title).
- [ ] Advanced metadata is optional and collapsed by default.
- [ ] Instance add/update is visible immediately after save.
- [ ] Results support multi-select with visible selected count.
- [ ] Bulk edit supports shared field update with confirmation preview.
- [ ] Inline quick edit updates row values without full page reload.

## Test Cases
- `INV-I-001` search by part number.
- `INV-I-002` filter by brand/category.
- `INV-I-003` quick add success.
- `INV-I-004` instance creation and render.
- `INV-I-005` bulk edit selected items.
- `INV-I-006` inline update quantity/value refresh.
