# Inventory Items Screen Spec

## Use Cases
1. User finds items quickly by search/filter.
2. User creates canonical items with minimal fields.
3. User manages instances without leaving inventory context.

## UI Sections
1. Search/Filter/Sort bar
2. Item list/table
3. Item details panel
4. Quick Add form
5. Advanced metadata form
6. Instances list + form

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

## Test Cases
- `INV-I-001` search by part number.
- `INV-I-002` filter by brand/category.
- `INV-I-003` quick add success.
- `INV-I-004` instance creation and render.

