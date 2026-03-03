# Cabinet UI Action Catalog — Wave 1 (Live Crawl)

Date: 2026-03-03
Session: live browser walkthrough as new user (`newuser+demo1@example.com`)

## Method
1. Captured generic shell actions from Home (baseline).
2. Navigated each primary nav page and extracted remaining clickable actions.
3. Expanded Settings and captured subpage actions.
4. Compared screen coverage to OpenSpec + Cypress file presence.

---

## Baseline shell actions (present across most screens)
- Team switcher trigger
- Sidebar links: Dashboard, Inventory, Wishlist, Discoveries, Scanner, Integrations, Chats, Users, Reports
- Settings nav expander + Help Center link
- Profile dropdown trigger
- Header controls: Toggle Sidebar, Search, Language, Theme toggle, Theme settings, Open Chat

## Page-specific actions discovered

### `/` Home
- Refresh Dashboard
- Starter onboarding: Start Setup, Import Existing Collection, Use Sample Data, Back Step, Next Step
- Attention cards: Open links to Discoveries/Wishlist/Pricing/Collection

### `/inventory`
- Add Item
- Add Folder
- Folder quick filters: All Items, Watch List, Wishlist Focus, Store 1, Store 2, Warehouse 1
- Table controls: Status, Priority, View, Rows/Cards, Title sort, pagination controls

### `/wishlist`
- Import
- Create
- Table controls: Status, Priority, View, Rows/Cards, Title sort
- Row actions: Select row, Open menu
- Pagination controls

### `/discoveries`
- Apply Filters (`discover-apply-filters`)

### `/scanner`
- Create Query Set (`scanner-create-query`)

### `/integrations`
- Retry (error-state)
- Integration type selector (All Integrations)
- Rows/Cards toggle

### `/chats`
- Retry (error-state)
- Create thread (`chat-create-thread-button`)
- Send (`chat-send-button`)
- Upload (`chat-upload-attachment-button`)
- Preview Action (`chat-preview-action-button`)
- Apply Action (`chat-apply-action-button`)

### `/users`
- Invite User
- Add User
- Retry (error-state)
- Table controls: Status, Role, View, Username sort, Email sort, pagination

### `/reports`
- Refresh Reports
- Export CSV (`reports-export-button`)
- Retry (error-state)

### `/help-center`
- No screen-specific actions beyond shell nav/header controls

### Settings subpages
#### `/settings` Profile
- Retry (error-state)
- Email selector (`settings-profile-email-trigger`)
- Add URL
- Update profile

#### `/settings/account`
- Retry (error-state)
- Date selector
- Language selector (`settings-account-language-trigger`)
- Update account

#### `/settings/appearance`
- Retry (error-state)
- Update preferences

#### `/settings/notifications`
- Retry (error-state)
- Notification toggles (`settings-notifications-*`)
- Update notifications
- Link to mobile settings

#### `/settings/display`
- Retry (error-state)
- Display location toggles (`settings-display-*`)
- Clear selection
- Update display

#### `/settings/storage`
- Retry (error-state)
- Reindex Search (disabled)
- Rebuild Thumbnails (disabled)

---

## OpenSpec/Cypress coverage check (by screen)
- sign-in/onboarding: spec ✅, e2e ✅
- dashboard-home: spec ✅, e2e ✅
- inventory: spec ✅, e2e ✅
- wishlist: spec ✅, e2e ✅
- discoveries: spec ✅, e2e ✅
- scanner: spec ✅, e2e ✅
- integrations: spec ✅, e2e ✅
- chats: spec ✅, e2e ✅
- users: spec ✅, e2e ✅
- reports: spec ✅, e2e ✅
- help-center: spec ✅, e2e ✅
- settings-main: spec ✅, e2e ✅
- settings-profile: spec ✅, e2e ✅
- settings-account: spec ✅, e2e ✅
- settings-appearance: spec ✅, e2e ✅
- settings-notifications: spec ✅, e2e ✅
- settings-display: spec ✅, e2e ✅
- settings-storage: spec ✅, e2e ❌ (no dedicated storage Cypress spec file found)

---

## Gaps observed while cataloging
- Multiple pages show `active_profile_404`/bootstrap-related retries.
- Users page: `users_fetch_failed_404` observed.
- Chat header still text button (`Open Chat`), not icon-only contract.
- Pin behavior not visible in chat side panel.

Related issues already filed: #239, #258, #259, #260, #261, #262, #263, #264, #265.
