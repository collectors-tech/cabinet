# 11 - Sortly Pattern Mapping (Semantic + Workflow)

This document converts reviewed Sortly references into strict Cabinet requirements.

Reviewed references:
- https://help.sortly.com/hc/en-us/articles/360036515592-Sortly-Product-Overview
- https://help.sortly.com/hc/en-us/articles/360030575572-Adding-Photos-to-Items-or-Folders
- https://help.sortly.com/hc/en-us/articles/360001995132-How-to-Update-Your-Inventory
- https://help.sortly.com/hc/en-us/articles/360016384492-Bulk-Edit-Items-and-Folders

## 1. Reference Semantic Structure (from overview screenshot)

```text
main.app-shell
  aside.primary-nav (global app navigation + utility links)
  section.collection-browser
    header.browser-toolbar
      h1 ("All Items")
      actions (Add Item, Add Folder)
      search (collection-level)
      view/summarize controls
    aside.folder-tree
      search (folders)
      tree list (folders/subfolders)
      utility actions (history/trash)
    section.results-region
      header.results-summary
        folders count
        items count
        total quantity
        total value
      section.results-grid-or-table
        folder/item cards with thumbnail, title, value, quantity
      footer.pagination
```

## 2. UX Principles to Adopt

1. Navigation stability
- Left nav remains fixed; page body owns vertical scroll.

2. Progressive density
- Start with scan/search/actions and summary; reveal deep metadata in context, not upfront.

3. Folder-first browsing
- Support hierarchical browse + search + quick totals in one workspace.

4. Explicit item operations
- Add, update, and bulk edit actions should be visible and proximity-based to selected rows/cards.

5. Photo-first inventory trust
- Photo workflows must be obvious and resilient: add, preview, set primary, retry on failure.

## 3. Workflow Requirements from Sortly Help Flows

### 3.1 Adding Photos
- Entry points:
  - Add photos while creating a new item/folder.
  - Add photos from existing item detail.
- Actions:
  - Upload one or multiple photos.
  - Preview before save.
  - Set primary image.
  - Delete/replace image.
- Failure handling:
  - Permission/error guidance with clear fallback.

### 3.2 Updating Inventory
- Update paths:
  - Inline update from list/detail.
  - Full edit from item detail.
- Fields:
  - quantity, price, notes, location, metadata fields.
- Constraints:
  - Immediate success/error feedback.
  - Deterministic state refresh without full page reset.

### 3.3 Bulk Edit Items/Folders
- Selection model:
  - Multi-select from list/grid.
  - Select-all current results.
- Bulk actions:
  - Edit shared fields/tags/location/status.
  - Move items/folders.
  - Delete with guardrails.
- Safety:
  - Clear preview of impact count.
  - Confirm destructive changes.

## 4. Cabinet Mapping (Strict)

## 4.1 Dashboard
- Keep dashboard action-first; avoid raw technical blocks in default view.
- Show only outcome-focused panels: attention queue, quick actions, recent activity.

## 4.2 Collection (Items)
- Add `Collection Browser` layout:
  - left folder tree pane
  - top action/search row
  - summary strip
  - results area with grid/table toggle
- Add row/card selection model powering bulk actions.
- Add inline quick edit drawer for high-frequency updates.

## 4.3 Photos
- Support multi-upload and staged preview before commit.
- Keep camera and file flows side-by-side with clear fallback.
- Add primary image badge and reorder support.

## 4.4 Shared
- Use sticky page header inside content area.
- Keep global nav fixed.
- Keep body scroll isolated to main content container.

## 5. Acceptance Additions (Issue Gate)

1. Collection workspace can be used in under 60 seconds for: search -> select -> bulk edit.
2. Photos workflow supports: upload -> preview -> set primary -> verify in card/list.
3. Inline inventory update persists and updates visible totals without full reload.
4. Bulk edit always shows selected count and change preview before apply.
5. Desktop scroll behavior: fixed nav, sticky content header, scrolling content body.

## 6. Test Additions

1. `COLL-BULK-001`: select multiple items, apply shared field update, verify all changed.
2. `COLL-BULK-002`: destructive bulk action requires explicit confirm.
3. `PHOTO-UX-001`: multi-image upload and primary image set persists after reload.
4. `INV-UPD-001`: inline quantity edit updates list row and summary strip.
5. `LAYOUT-001`: desktop shell keeps sidebar fixed while content scrolls.
