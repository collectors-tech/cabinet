# Record Action Surface Matrix

This matrix captures the current Cabinet table-like record surfaces for #1938
before shared row action menus and CRUD dialogs are implemented. It is a
migration baseline, not the final component API.

| Surface | Current row actions | Current create/edit dialog | Current destructive or state-change confirmation | Migration notes |
| --- | --- | --- | --- | --- |
| Inventory | Kebab row menu from `tasks` table for Edit, disabled Make a copy, disabled Favorite, Delete. Restore is a separate inline row button for deleted records. | Existing task mutation drawer/dialog flow for create/edit plus row detail dialogs. | Existing task delete confirmation; deleted inventory rows can restore from the inline control. | Move Edit, Duplicate, Delete, Restore, and Permanent delete into shared capability-driven action definitions; remove static disabled placeholders unless backed by a truthful reason. |
| Wishlist | Kebab row menu from `tasks` table for Edit and Delete or Delete permanently. Restore is a separate inline row button for deleted records. | Wishlist create/edit drawer plus import/paste/screenshot/barcode dialogs. | Wishlist delete confirmation distinguishes deleted entries via permanent-delete label; bulk delete lives in toolbar. | Use the shared menu ordering while keeping bulk actions in the toolbar; restore and permanent delete need record capability flags. |
| Users | Kebab row menu for Edit and Delete; protected owner rows omit Delete. Details also expose sidebar Edit/Delete actions. | Add New User and Edit User dialog; invite dialog is separate. | `ConfirmDialog` based single and bulk delete flows; owner safeguard covered by capability omission. | Convert owner protection to explicit row capabilities so Delete is omitted or disabled with a reason consistently. |
| Collections | Exposed icon buttons for View in inventory, Edit, and Delete in the row action cell. | Inline page-level Create collection and edit collection dialogs. | Custom Delete collection dialog with destination handling. | Consolidate three icon buttons into the shared kebab menu ordered View/Open, Edit, Delete; map delete copy to destructive confirmation with destination requirement. |
| Media | Exposed icon buttons for Open, Analyze, Assign, and Archive in rows and cards; Analyze and Assign are disabled by record state. | Add media, Analyze media, Assign media, and metadata edit dialogs. | Archive is exposed as an action but currently has no shared destructive confirmation contract. | Use capabilities for Analyze/Assign availability; Archive must use destructive/state-change copy that names the asset and clarifies soft archive behavior. |
| Purchases | Detail pane buttons/actions drive Reconcile and related purchase workflow actions rather than a compact row menu. | Purchase draft, line item, review, and detail action dialogs. | Detail action confirmation is a generic Queue action dialog with target and notes. | Treat order and line-item actions as record actions when rendered in table/detail rows; shared dialog copy must name whether the target is an order or line item. |
| Market Watch | Query table exposes inline buttons for Run Now, Pause/Resume, Edit, Inspect Output, and Delete. Row double-click opens output details. | Query create/edit form/dialog and output detail panel. | Delete runs directly through the page action path; Pause/Resume are state changes without shared confirmation. | Convert row action buttons into shared menu actions ordered Open/Inspect, Edit, state change, Delete while preserving double-click output details and row-click isolation. |
| Discoveries | Row action cell exposes icon buttons for Review source, Restore for review, Ignore/archive, Promote to Wishlist, Purchase follow-up, and Inventory handoff. | Promotion/handoff work is action-driven rather than a standard CRUD dialog. | Ignore/archive and restore run as state changes; no shared confirmation copy. | Model each available action as a capability; unsupported promotion paths stay omitted. Archive/restore labels should remain domain-specific. |
| Integrations | Configured integrations table exposes one row button: Edit or Connect. Row single-click opens details; row double-click opens edit. Provider detail modal lists manifest actions. | Provider selector and provider-specific configuration dialog; row details and row edit modals. | Disconnect appears inside provider configuration, not as a row confirmation. | Shared row menu should provide View/Open details and Edit/Connect without triggering row click; provider manifest actions need capability mapping before exposing more menu items. |
| Settings | Settings has several record-like tables/lists: Skills rows expose Enable/Disable and Details, Storage backup rows expose Download and Restore, Categories lists expose Remove controls. | Settings forms are mostly page-local; Skills uses a details sheet and import panel, Storage uses a restore dialog. | Storage restore has a confirmation dialog; Categories removal is inline before Save; Skill enable/disable is direct. | Only migrate settings record lists that have repeated row actions. Keep global settings forms out of scope unless they become table row actions. |

## Shared Migration Constraints

- Row actions that mutate, archive, restore, delete, disable, revoke, or hand off
  records must be capability-driven per row.
- Unsupported actions must be omitted unless a disabled state gives a truthful,
  short reason.
- Bulk operations remain toolbar actions and must not move into per-row menus.
- Row action triggers must stop pointer/keyboard propagation so row selection,
  row click, and double-click detail navigation remain isolated.
- Dialogs launched from a row action must restore focus to the invoking trigger
  when that trigger is still mounted.
