# 15 - Template Section Names and Mapping

This document defines canonical names for template sections and the Cabinet-specific mapping we will use going forward.

## 1. Shell Section Names (Canonical)

1. `Header Account Menu`  
Context: top-right profile/avatar dropdown in header.

2. `Sidebar Account Panel`  
Context: bottom-left user block in sidebar footer.

3. `Primary Navigation`  
Context: left application navigation menu.

4. `Context Navigation`  
Context: secondary navigation panel for active workspace/folder/view context.

5. `Top Tabs`  
Context: horizontal header tabs near page title.

6. `Page Header`  
Context: title + quick actions row for current page.

7. `Page Content`  
Context: main scrollable body.

8. `System Pages`  
Context: auth and error routes.

## 2. Header Account Menu (Top Right)

### Include
- Keyboard shortcut list (platform-aware labels):
  - macOS labels: `⌘`, `⌥`, `⇧`
  - Windows labels: `Ctrl`, `Alt`, `Shift`
- Signed-in user identity:
  - display name from auth state
  - email from auth state

### Remove
- `New Team`
- `Billing`

### Canonical naming
- `Header Account Menu`
- `Keyboard Shortcuts`
- `Signed-In Identity`

## 3. Sidebar Account Panel (Bottom Left)

### Include
- profile summary
- account/settings shortcuts

### Remove
- `Upgrade to Pro`

### Note on visual treatment
- if an upsell row exists in template with a divider, remove the row and collapse spacing so no empty highlight/divider remains.

### Canonical naming
- `Sidebar Account Panel`
- `Sidebar Account Actions`

## 4. Internationalization (i18n)

### Template capability
- The template has direction support and structure that can host i18n.
- It does not provide complete Cabinet i18n content packs by default.

### Cabinet decision
- Add explicit i18n framework integration during migration (`i18next` + `react-i18next`).
- Bootstrap with one locale in v1:
  - `en` (default and only active locale)
- Store locale resources under:
  - `ui.web/src/locales/{lang}/{namespace}.json`
- Required namespaces:
  - `common`
  - `nav`
  - `pages`

### UI principles (mandatory)
- No new hard-coded shell labels in components once i18n is wired.
- All top-level shell labels (sidebar groups/items, page header primitives) must resolve through translation keys.
- Translation keys must be stable and semantic (`groups.general`, `items.dashboard`), not presentation-dependent.
- Unknown/missing keys must render safe default text (do not break navigation rendering).
- Locale architecture must support adding new languages without structural refactor.

### Canonical naming
- `Localization Layer`
- `Locale Dictionary`
- `Platform Shortcut Labels`

## 5. Primary Navigation Mapping

## 5.1 Dashboard

Canonical page name: `Home Dashboard`

Primary KPI cards (required):
- `Inventory Count`
- `Inventory Value`
- `Inventory Value Delta (MoM)`
- `Watchlist Count`
- `Watchlist Value`
- `Watchlist Value Delta (MoM)`
- `Recommendations Count`

Additional recommended cards:
- `Discoveries Pending Review`
- `Price Drops (24h/7d)`
- `Low Stock Opportunities`
- `Scanner Health`

## 5.2 Tasks -> Inventory

Canonical page name: `Inventory Workspace`

Baseline behavior:
- Use Tasks-style data table foundation.
- Wire to Cabinet Inventory APIs/data model.

Required columns:
- `Thumbnail`
- `Product Code`
- `Title`
- `Packaging Type`
- `Purchase Date`

Image interaction requirements:
- Hover preview: larger popover preview.
- Click: `Media Lightbox`.
- Lightbox supports left/right navigation across current list (carousel behavior).
- Lightbox footer shows key metadata for active row.

Status model:
- `New`
- `Ungraded`
- `Graded`
- `Deleted`

Rules:
- default priority on new items: `Medium`
- `Deleted` rows hidden by default and visible via status filter.

Additional fields:
- `Packaging Type`: `Loose`, `Blister`
- `Car Grade Type`
- `Packaging Grade Type`
- `Collector Classification`
- `Is Graded` (yes/no)
- `Grader` (text)
- `Grade` (numeric)
- `Slabbed` (yes/no)

## 5.3 Tasks Copy -> Wishlist

Canonical page name: `Wishlist Workspace`

Baseline behavior:
- Reuse Inventory Workspace table mechanics.
- Wire to Wishlist APIs/data model.

Required columns:
- `Thumbnail`
- `Media Code`
- `Title`
- `Scheduled` (yes/no)
- `Scheduled Date`

Status model:
- `Discovered`
- `Wishlist`
- `Deleted`

Rules:
- default priority on new wishlist entries: `Medium`
- `Deleted` rows hidden by default and visible via status filter.
- maintain same descriptive fields as Inventory where applicable to articulate target item characteristics.

## 5.4 Apps

Canonical page name: `Integrations`

Keep:
- `Gmail`
- `Telegram`
- `WhatsApp`
- `GitHub`
- `Mr Toys`
- `Voglers`
- `WA Slot Cars`
- `Bonza Slot Cars`
- `K & K Creative Toys`
- `Hobbyco`
- `Frontline Hobbies`
- `Metro Hobbies`

## 5.5 Chats

Canonical page name: `Chat Workspace`

Decision:
- Keep this page.
- Add API contract for external ingestion (for example AI assistant/chat events).

## 5.6 Users

Canonical page name: `User Management`

Decision:
- Keep page.
- Single-user mode initially.
- Must be wired to Cabinet APIs.
- `Invite User` and `Add User` actions require API endpoints and full flow wiring.

## 5.7 Error Pages

Canonical section name: `System Error Pages`

Decision:
- Keep routes and wire into Go app error handling.
- Remove from primary navigation.

## 5.8 Auth / Secure by Clerk

Canonical section name: `Cloud Authentication`

Decision:
- Keep and wire into Cabinet auth flow.
- Remove redundant auth demo entries from primary navigation.

## 5.9 Settings

Canonical page name: `Settings Workspace`

Decision:
- Merge existing template sections with Cabinet settings domains.
- Add Cabinet-specific sections where no template equivalent exists.

## 5.10 Help Center

Canonical page name: `Help Center`

Decision:
- Keep page.
- Use cards layout for Cabinet documentation and guides.

## 6. Grading and Classification Vocabulary (Canonical)

## 6.1 Car Condition Grade
- `M` - Mint (Unrun)
- `NM` - Near Mint
- `EX` - Excellent
- `VG` - Very Good
- `G` - Good
- `F` - Fair
- `P` - Poor

## 6.2 Packaging Grade
- `Sealed Mint`
- `Sealed Good`
- `Opened Complete`
- `Loose`

## 6.3 Collector Classification
- `Standard Release`
- `Collector Series`
- `Limited Edition`
- `Clear Series`
- `Tribute Edition`
- `Set Car`
- `Store Exclusive`
- `Vintage Mega G`
- `Mega G+`
- `Racemasters Era`

## 6.4 Grading Status
- `Ungraded`
- `Graded`

## 6.5 Grading Company
Examples:
- `VSS`
- `CGA`
- `Other` (free text)

## 6.6 Numerical Grade Scale
- `10.0` - Gem Mint
- `9.5` - Near Mint+
- `9.0` - Mint
- `8.0` - NM
- `7.0` - Excellent
- `6.0 and below` - wear present

## 7. Immediate Implementation Labels (for tickets and code)

Use these labels exactly in issues, specs, and components:
- `Header Account Menu`
- `Sidebar Account Panel`
- `Home Dashboard`
- `Inventory Workspace`
- `Wishlist Workspace`
- `Integrations`
- `Chat Workspace`
- `User Management`
- `System Error Pages`
- `Cloud Authentication`
- `Settings Workspace`
- `Help Center`

## 8. Row-to-Detail Interaction Standard

Applies to:
- `Inventory Workspace`
- `Wishlist Workspace`
- `Integrations` (table/list variants)

Scope rule:
- Drawer + URL selection behavior is mandatory across all table-based screens.
- Lightbox/carousel behavior is only enabled for media-bearing tables.

### 8.1 Base interaction model
- Use split interaction model for fast/predictable behavior.

### 8.2 Row click opens Details Drawer
- Click anywhere on non-interactive row area opens right-side `Details Drawer`.
- Drawer includes:
  - full metadata
  - schedule info
  - linked media/posts
  - audit fields
  - row-level actions
- URL includes selected record id:
  - `?selected=<id>`
- Refresh/back/forward preserves selected context.
- App should also persist last selected context/view per screen so reload restores previous working state when possible.

### 8.3 Thumbnail click opens Lightbox (media-bearing tables only)
- Clicking row image opens modal `Media Lightbox` (not drawer).
- Left/right navigates adjacent items in current filtered/sorted result set.
- Keyboard:
  - `Enter` from focused row opens drawer
  - `Space` on focused thumbnail opens lightbox

### 8.4 Selection mode for bulk actions
- Checkbox column is explicit selection control.
- In bulk mode, row click behavior remains unchanged from normal mode.
- Selection actions are checkbox-driven (row click does not toggle selection).
- `Esc` exits selection mode.
- `Open details` action remains available for single selected row.

### 8.5 Interaction guards
- Clicks on interactive controls do not trigger row open:
  - checkbox
  - menu button
  - status chip
  - link
- Double-click behavior is optional:
  - open edit modal directly
  - modal supports previous/next record navigation within current table result set

### 8.6 Mobile behavior
- Row tap opens full-screen detail sheet.
- Thumbnail tap opens lightbox.
- Bulk mode is entered via explicit `Select` action button.

### 8.7 Unsaved changes guard
- If details form has unsaved changes and user changes row or closes drawer:
  - show `Save / Discard / Cancel` confirmation.

### 8.8 Hidden/deleted behavior
- `Deleted` rows are hidden by default and shown via status filter.
- If direct URL references hidden/deleted row, open details with status warning banner and `Restore` CTA.

### 8.9 Integrations-specific application
- `Integrations` rows open a details modal that includes:
  - provider details
  - available actions
  - setup instructions
  - credentials/keys storage section
- Lightbox is disabled for integrations unless a media thumbnail column is explicitly introduced in that screen.
- If integrations are rendered as table/list rows, selection URL state still uses `?selected=<id>`.

### 8.10 Carousel order source
- Lightbox and row-to-row navigation order must always follow the active table source order:
  - current filters
  - current sort
  - current visible result set

## 9. Keyboard Shortcuts Standard

Standard shortcuts:
- `Cmd/Ctrl+K` open global search/command
- `Enter` on focused row opens details drawer/modal
- `Space` on focused thumbnail opens lightbox
- `Esc` closes drawer/lightbox and exits bulk mode

Display requirement:
- Show platform-specific labels in UI:
  - macOS: `⌘`, `⌥`, `⇧`
  - Windows: `Ctrl`, `Alt`, `Shift`

## 10. User and Database Access Model

### 10.1 Scope
- Users are scoped to the current database.
- If a user has their own account/database, that is a separate database scope.

### 10.2 Initial roles
- `View`
- `Admin`

### 10.3 Invite/Add behavior (initial)
- `Add User` is active and writes to local DB user store.
- `Invite User` does not depend on email delivery initially; user records can be added for same-network access scenarios.

## 11. Status and Grading Configurability

### 11.1 Initial status sets (default)
- Inventory: `New | Ungraded | Graded | Deleted`
- Wishlist: `Discovered | Wishlist | Deleted`

### 11.2 Configurability
- Status enums must be configurable (global per database).
- Grading enums must be configurable (global per database).

### 11.3 Grading field type
- `Car Grade Type`: enum list (configurable)
- `Packaging Grade Type`: enum list (configurable)
- Use provided defaults from this document as seed values.

## 12. Delete and Recycle Lifecycle Policy

### 12.1 Soft delete flow
- First delete action sets status to `Deleted` (hidden from default list views).
- Items with status `Deleted` can be moved to `Recycle` on explicit delete action.

### 12.2 Recycle behavior
- `Recycle` is a separate list/workspace.
- Recycle supports:
  - restore item
  - permanent delete one-by-one
  - permanent delete all (subject to link constraints)

### 12.3 Link constraints
- Items in `Recycle` cannot be permanently deleted while linked dependencies exist.

### 12.4 Retention
- Recycle retention is indefinite for now.
