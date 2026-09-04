## 1. Tests First

- [x] 1.1 Add API or unit coverage proving item type condition scales are seeded and persisted through profile settings.
- [x] 1.2 Add Inventory UI coverage proving Item Type scopes condition choices.
- [x] 1.3 Add Inventory UI coverage proving Category and Condition compact filters are present and filter rows.

## 2. Settings and API

- [x] 2.1 Add shared taxonomy helpers for item type condition scales and defaults.
- [x] 2.2 Extend grading/taxonomy settings API payloads with item type condition scales while preserving existing flat enum compatibility.
- [x] 2.3 Add or extend Settings UI so users can manage item types and their condition values.

## 3. Inventory UI

- [x] 3.1 Add Item Type field to inventory create/edit/detail forms.
- [x] 3.2 Scope condition controls to the selected Item Type condition scale.
- [x] 3.3 Restore shared compact Category and Condition filters in the inventory browser.
- [x] 3.4 Keep Category as flexible multi-select/free-add metadata independent from Item Type.

## 4. Verification and Delivery

- [x] 4.1 Run OpenSpec validation for the change.
- [x] 4.2 Run targeted API/UI tests.
- [x] 4.3 Build Cabinet, merge to `develop`, close issue, delete branch, and restart the demo.
