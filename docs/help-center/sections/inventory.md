# Inventory

## Use Inventory for
- Managing owned items
- Folder browsing and collection context
- Photos, barcodes, AI assist

## Common actions
- **New**: create an inventory item
- **Create menu**: quick-create related entities
- Filter/sort rows and switch Rows/Cards views

## Item taxonomy and grading
- Use **Item type** to pick the configured collector class for the item, such as Slot Cars, Trading Cards, or a custom type from Settings > Categories.
- Use **Condition** after selecting the item type; condition choices come from that item type's configured condition scale.
- Use **Packaging grade** for boxed, sealed, loose, or other packaging states maintained in the active profile taxonomy.
- Inventory search matches item type, condition, category, and packaging grade values, and the item type and packaging grade filters are preserved in saved view URLs.
- If an API or import sends a taxonomy value that is not configured for the active profile, Cabinet returns `invalid_taxonomy_value` with the field that needs correction.
