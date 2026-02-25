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
- Initial locales:
  - `en-AU` (default)
  - `en-US` (secondary)

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

### 8.3 Thumbnail click opens Lightbox
- Clicking row image opens modal `Media Lightbox` (not drawer).
- Left/right navigates adjacent items in current filtered/sorted result set.
- Keyboard:
  - `Enter` from focused row opens drawer
  - `Space` on focused thumbnail opens lightbox

### 8.4 Selection mode for bulk actions
- Checkbox column is explicit selection control.
- When any checkbox is selected, row click toggles selection and does not open drawer.
- `Esc` exits selection mode.
- `Open details` action remains available for single selected row.

### 8.5 Interaction guards
- Clicks on interactive controls do not trigger row open:
  - checkbox
  - menu button
  - status chip
  - link
- Double-click behavior is optional:
  - may open edit dialog directly
  - can be deferred if interaction noise is high

### 8.6 Mobile behavior
- Row tap opens full-screen detail sheet.
- Thumbnail tap opens lightbox.
- Bulk mode is entered via explicit `Select` action button.

### 8.7 Unsaved changes guard
- If details form has unsaved changes and user changes row or closes drawer:
  - show `Save / Discard / Cancel` confirmation.

### 8.8 Hidden/deleted behavior
- `Deleted` rows are hidden by default and shown via status filter.
- If direct URL references hidden/deleted row, open details with status warning banner.
