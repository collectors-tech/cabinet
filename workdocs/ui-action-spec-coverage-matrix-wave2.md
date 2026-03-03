# UI Action -> Spec Coverage Matrix (Wave 2)

## dashboard-home
- Spec: `openspec/specs/dashboard/ui-screen-home/spec.md`
- Requirement IDs found: UI-SCREEN-HOME-001, UI-SCREEN-HOME-002, UI-SCREEN-HOME-003
- Action keys observed:
  - Refresh Dashboard
  - Start Setup
  - Import Existing Collection
  - Use Sample Data
  - Next Step

## inventory
- Spec: `openspec/specs/inventory/ui-screen-inventory-items/spec.md`
- Requirement IDs found: UI-SCREEN-INVENTORY-ITEMS-001, UI-SCREEN-INVENTORY-ITEMS-002, UI-SCREEN-INVENTORY-ITEMS-003, UI-SCREEN-INVENTORY-ITEMS-004, UI-SCREEN-INVENTORY-ITEMS-005, UI-SCREEN-INVENTORY-ITEMS-006
- Action keys observed:
  - Add Item
  - Add Folder
  - Status
  - Priority
  - View
  - Rows
  - Cards

## wishlist
- Spec: `openspec/specs/wishlist/ui-screen-wishlist/spec.md`
- Requirement IDs found: UI-SCREEN-WISHLIST-001, UI-SCREEN-WISHLIST-002, UI-SCREEN-WISHLIST-003, UI-SCREEN-WISHLIST-004, UI-SCREEN-WISHLIST-005
- Action keys observed:
  - Import
  - Create
  - Open menu
  - Rows
  - Cards

## discoveries
- Spec: `openspec/specs/dashboard/ui-screen-discover/spec.md`
- Requirement IDs found: UI-SCREEN-DISCOVER-001, UI-SCREEN-DISCOVER-002, UI-SCREEN-DISCOVER-003
- Action keys observed:
  - Apply Filters

## scanner
- Spec: `openspec/specs/integrations/ui-screen-scanner/spec.md`
- Requirement IDs found: UI-SCREEN-SCANNER-001, UI-SCREEN-SCANNER-002, UI-SCREEN-SCANNER-003
- Action keys observed:
  - Create Query Set

## integrations
- Spec: `openspec/specs/integrations/ui-screen-integrations/spec.md`
- Requirement IDs found: UI-SCREEN-INTEGRATIONS-001, UI-SCREEN-INTEGRATIONS-002, UI-SCREEN-INTEGRATIONS-003, UI-SCREEN-INTEGRATIONS-004, UI-SCREEN-INTEGRATIONS-005, UI-SCREEN-INTEGRATIONS-006, UI-SCREEN-INTEGRATIONS-007
- Action keys observed:
  - Retry
  - All Integrations
  - Rows
  - Cards

## chats
- Spec: `openspec/specs/chats/ui-screen-chat-copilot/spec.md`
- Requirement IDs found: UI-SCREEN-CHAT-COPILOT-001, UI-SCREEN-CHAT-COPILOT-002, UI-SCREEN-CHAT-COPILOT-003, UI-SCREEN-CHAT-COPILOT-004, UI-SCREEN-CHAT-COPILOT-005, UI-SCREEN-CHAT-COPILOT-006, UI-SCREEN-CHAT-COPILOT-007, UI-SCREEN-CHAT-COPILOT-008
- Action keys observed:
  - Create
  - Send
  - Upload
  - Preview Action
  - Apply Action
  - Open Chat

## users
- Spec: `openspec/specs/users/ui-screen-users/spec.md`
- Requirement IDs found: UI-SCREEN-USERS-001, UI-SCREEN-USERS-002, UI-SCREEN-USERS-003
- Action keys observed:
  - Invite User
  - Add User
  - Status
  - Role

## reports
- Spec: `openspec/specs/dashboard/ui-screen-reports/spec.md`
- Requirement IDs found: UI-SCREEN-REPORTS-001, UI-SCREEN-REPORTS-002, UI-SCREEN-REPORTS-003
- Action keys observed:
  - Refresh Reports
  - Export CSV

## help-center
- Spec: `openspec/specs/helpcenter/ui-screen-help-center/spec.md`
- Requirement IDs found: UI-SCREEN-HELP-CENTER-001, UI-SCREEN-HELP-CENTER-002
- Action keys observed: (no page-specific actions beyond shell controls)

## settings
- Spec: `openspec/specs/settings/ui-screen-settings/spec.md`
- Requirement IDs found: UI-SCREEN-SETTINGS-001, UI-SCREEN-SETTINGS-002, UI-SCREEN-SETTINGS-003
- Action keys observed:
  - Profile
  - Account
  - Appearance
  - Notifications
  - Display
  - Storage

## settings-profile
- Spec: `openspec/specs/settings/profile/spec.md`
- Requirement IDs found: UI-SCREEN-SETTINGS-PROFILE-001, UI-SCREEN-SETTINGS-PROFILE-002
- Action keys observed:
  - Add URL
  - Update profile

## settings-account
- Spec: `openspec/specs/settings/account/spec.md`
- Requirement IDs found: UI-SCREEN-SETTINGS-ACCOUNT-001, UI-SCREEN-SETTINGS-ACCOUNT-002
- Action keys observed:
  - Update account

## settings-appearance
- Spec: `openspec/specs/settings/appearance/spec.md`
- Requirement IDs found: UI-SCREEN-SETTINGS-APPEARANCE-001, UI-SCREEN-SETTINGS-APPEARANCE-002, UI-SCREEN-SETTINGS-APPEARANCE-003
- Action keys observed:
  - Update preferences

## settings-notifications
- Spec: `openspec/specs/settings/notifications/spec.md`
- Requirement IDs found: UI-SCREEN-SETTINGS-NOTIFICATIONS-001, UI-SCREEN-SETTINGS-NOTIFICATIONS-002
- Action keys observed:
  - Update notifications

## settings-display
- Spec: `openspec/specs/settings/display/spec.md`
- Requirement IDs found: UI-SCREEN-SETTINGS-DISPLAY-001, UI-SCREEN-SETTINGS-DISPLAY-002
- Action keys observed:
  - Clear selection
  - Update display

## settings-storage
- Spec: `openspec/specs/settings/storage/spec.md`
- Requirement IDs found: UI-SCREEN-SETTINGS-STORAGE-001, UI-SCREEN-SETTINGS-STORAGE-002, UI-SCREEN-SETTINGS-STORAGE-003, UI-SCREEN-SETTINGS-STORAGE-004, UI-SCREEN-SETTINGS-STORAGE-005
- Action keys observed:
  - Retry
  - Reindex Search
  - Rebuild Thumbnails

## Current known gaps from live comparison
- Chat header control not icon-only and no pin behavior visible (#239).
- Active profile bootstrap missing causes module retries/404 states (#258, #262, #264).
- Settings storage has no dedicated Cypress screen spec file (#266).
- Settings nav parity incomplete for Operations/Billing (#261).
