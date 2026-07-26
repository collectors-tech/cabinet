## MODIFIED Requirements

### Requirement: Cabinet UI SHALL provide shared table and dialog primitives
Cabinet MUST expose reusable UI primitives for table interaction, row-level
record actions, CRUD dialogs, and route-level shell metadata so feature
surfaces do not implement incompatible copies of the same behavior.

#### Scenario: Authenticated routes use canonical metadata
- **GIVEN** an authenticated Cabinet route is visible in the shell
- **WHEN** the application resolves its page metadata
- **THEN** the route SHALL resolve to exactly one canonical title,
  description, icon, navigation group, browser-title eligibility, and stable
  header/icon test identifier
- **AND** page headers, sidebar navigation, command/search navigation, and
  browser document titles SHALL use the same canonical title when they describe
  the same route
- **AND** settings child routes SHALL have specific page titles/icons while
  remaining grouped under Settings navigation
- **AND** unknown, unsupported, and dynamic routes SHALL use an explicit
  fallback that does not hide missing metadata for known authenticated routes
- **AND** existing packaged tests that depend on stable page/icon test IDs
  SHALL keep a compatible identifier or receive an intentional documented
  migration.

#### Scenario: Authenticated route registry covers the route matrix
- **GIVEN** the authenticated route metadata registry is loaded
- **WHEN** the registry enumerates canonical Cabinet routes
- **THEN** it SHALL include the following entries before consumers are migrated:

| Path or pattern | Canonical title | Navigation group | Icon | Document title | Stable test IDs |
| --- | --- | --- | --- | --- | --- |
| `/`, `/dashboard` | Home | General | `LayoutDashboard` | `Cabinet - Home` | `dashboard-header-title`, `dashboard-header-icon`, `sidebar-nav-link-dashboard` |
| `/inventory` | Inventory | General | `ListChecks` | `Cabinet - Inventory` | `inventory-header-title`, `inventory-header-icon`, `sidebar-nav-link-inventory` |
| `/media` | Media | General | `Images` | `Cabinet - Media` | `media-header-title`, `media-header-icon`, `sidebar-nav-link-media` |
| `/collections` | Collections | General | `Tag` | `Cabinet - Collections` | `collections-header-title`, `collections-header-icon`, `sidebar-nav-link-collections` |
| `/wishlist` | Wishlist | General | `Heart` | `Cabinet - Wishlist` | `wishlist-header-title`, `wishlist-header-icon`, `sidebar-nav-link-wishlist` |
| `/discoveries` | Discoveries | General | `Telescope` | `Cabinet - Discoveries` | `discoveries-header-title`, `discoveries-header-icon`, `sidebar-nav-link-discoveries` |
| `/scanner` | Market Watch | General | `ScanSearch` | `Cabinet - Market Watch` | `scanner-header-title`, `scanner-header-icon`, `sidebar-nav-link-market-watch` |
| `/inbox` | Notification Inbox | General | `Bell` | `Cabinet - Notification Inbox` | `inbox-header-title`, `inbox-header-icon`, `sidebar-nav-link-inbox` |
| `/purchases` | Purchases | General | `Inbox` | `Cabinet - Purchases` | `purchases-header-title`, `purchases-header-icon`, `sidebar-nav-link-purchases` |
| `/integrations` | Integrations | General | `PlugZap` | `Cabinet - Integrations` | `integrations-header-title`, `integrations-header-icon`, `sidebar-nav-link-integrations` |
| `/chats` | Chats | General | `MessagesSquare` | `Cabinet - Chats` | `chats-header-title`, `chats-header-icon`, `sidebar-nav-link-chats` |
| `/users` | Users | General | `Users` | `Cabinet - Users` | `users-header-title`, `users-header-icon`, `sidebar-nav-link-users` |
| `/reports` | Reports | General | `ChartColumn` | `Cabinet - Reports` | `reports-header-title`, `reports-header-icon`, `sidebar-nav-link-reports` |
| `/settings`, `/settings/profile` | Profile Settings | Settings | `UserCog` | `Cabinet - Profile Settings` | `settings-header-title`, `settings-header-icon`, `sidebar-nav-link-profile` |
| `/settings/account` | Account Settings | Settings | `Wrench` | `Cabinet - Account Settings` | `settings-header-title`, `settings-header-icon`, `sidebar-nav-link-account` |
| `/settings/appearance` | Appearance Settings | Settings | `Palette` | `Cabinet - Appearance Settings` | `settings-header-title`, `settings-header-icon`, `sidebar-nav-link-appearance` |
| `/settings/billing` | Billing Settings | Settings | `CreditCard` | `Cabinet - Billing Settings` | `settings-header-title`, `settings-header-icon`, `sidebar-nav-link-billing` |
| `/settings/categories` | Category Settings | Settings | `Tag` | `Cabinet - Category Settings` | `settings-header-title`, `settings-header-icon`, `sidebar-nav-link-categories` |
| `/settings/display` | Display Settings | Settings | `Monitor` | `Cabinet - Display Settings` | `settings-header-title`, `settings-header-icon`, `sidebar-nav-link-display` |
| `/settings/integrations` | Integration Settings | Settings | `PlugZap` | `Cabinet - Integration Settings` | `settings-header-title`, `settings-header-icon`, `sidebar-nav-link-integrations` |
| `/settings/notifications` | Notification Settings | Settings | `Bell` | `Cabinet - Notification Settings` | `settings-header-title`, `settings-header-icon`, `sidebar-nav-link-notifications` |
| `/settings/operations` | Operations Settings | Settings | `Settings` | `Cabinet - Operations Settings` | `settings-header-title`, `settings-header-icon`, `sidebar-nav-link-operations` |
| `/settings/skills` | Skills Settings | Settings | `BrainCircuit` | `Cabinet - Skills Settings` | `settings-header-title`, `settings-header-icon`, `sidebar-nav-link-skills` |
| `/settings/storage` | Storage Settings | Settings | `Database` | `Cabinet - Storage Settings` | `settings-header-title`, `settings-header-icon`, `sidebar-nav-link-storage` |
| `/help-center` | Help Center | Other | `HelpCircle` | `Cabinet - Help Center` | `help-center-header-title`, `help-center-header-icon`, `sidebar-nav-link-help-center` |
| `/errors/*` | Error | System | `AlertTriangle` | `Cabinet - Error` | `error-header-title`, `error-header-icon` |
| `/404` | Not Found | System | `CircleHelp` | `Cabinet - Not Found` | `not-found-header-title`, `not-found-header-icon` |

- **AND** each entry SHALL include a non-empty user-facing description suitable
  for page headers and command/search navigation metadata
- **AND** paths not listed in this matrix SHALL resolve through a documented
  fallback rather than silently inheriting a known route's metadata.
