# UI Action -> Requirement Text Match (Wave 3)

| Screen | Action | Spec path | Text match in spec |
|---|---|---|---|
| dashboard-home | Refresh Dashboard | $(System.Collections.Hashtable.spec) | ❌ |
| dashboard-home | Start Setup | $(System.Collections.Hashtable.spec) | ❌ |
| dashboard-home | Import Existing Collection | $(System.Collections.Hashtable.spec) | ❌ |
| dashboard-home | Use Sample Data | $(System.Collections.Hashtable.spec) | ❌ |
| dashboard-home | Back Step | $(System.Collections.Hashtable.spec) | ❌ |
| dashboard-home | Next Step | $(System.Collections.Hashtable.spec) | ❌ |
| inventory | Add Item | $(System.Collections.Hashtable.spec) | ❌ |
| inventory | Add Folder | $(System.Collections.Hashtable.spec) | ❌ |
| inventory | Status | $(System.Collections.Hashtable.spec) | ❌ |
| inventory | Priority | $(System.Collections.Hashtable.spec) | ❌ |
| inventory | View | $(System.Collections.Hashtable.spec) | ❌ |
| inventory | Rows | $(System.Collections.Hashtable.spec) | ✅ |
| inventory | Cards | $(System.Collections.Hashtable.spec) | ❌ |
| inventory | Title | $(System.Collections.Hashtable.spec) | ✅ |
| wishlist | Import | $(System.Collections.Hashtable.spec) | ✅ |
| wishlist | Create | $(System.Collections.Hashtable.spec) | ✅ |
| wishlist | Open menu | $(System.Collections.Hashtable.spec) | ✅ |
| wishlist | Rows | $(System.Collections.Hashtable.spec) | ✅ |
| wishlist | Cards | $(System.Collections.Hashtable.spec) | ✅ |
| wishlist | Title | $(System.Collections.Hashtable.spec) | ❌ |
| discoveries | Apply Filters | $(System.Collections.Hashtable.spec) | ❌ |
| scanner | Create Query Set | $(System.Collections.Hashtable.spec) | ❌ |
| integrations | Retry | $(System.Collections.Hashtable.spec) | ✅ |
| integrations | All Integrations | $(System.Collections.Hashtable.spec) | ❌ |
| integrations | Rows | $(System.Collections.Hashtable.spec) | ❌ |
| integrations | Cards | $(System.Collections.Hashtable.spec) | ✅ |
| chats | Open Chat | $(System.Collections.Hashtable.spec) | ✅ |
| chats | Create | $(System.Collections.Hashtable.spec) | ✅ |
| chats | Send | $(System.Collections.Hashtable.spec) | ✅ |
| chats | Upload | $(System.Collections.Hashtable.spec) | ❌ |
| chats | Preview Action | $(System.Collections.Hashtable.spec) | ❌ |
| chats | Apply Action | $(System.Collections.Hashtable.spec) | ✅ |
| chats | Retry | $(System.Collections.Hashtable.spec) | ✅ |
| users | Invite User | $(System.Collections.Hashtable.spec) | ✅ |
| users | Add User | $(System.Collections.Hashtable.spec) | ✅ |
| users | Status | $(System.Collections.Hashtable.spec) | ✅ |
| users | Role | $(System.Collections.Hashtable.spec) | ✅ |
| users | Retry | $(System.Collections.Hashtable.spec) | ❌ |
| reports | Refresh Reports | $(System.Collections.Hashtable.spec) | ❌ |
| reports | Export CSV | $(System.Collections.Hashtable.spec) | ❌ |
| reports | Retry | $(System.Collections.Hashtable.spec) | ✅ |
| settings-profile | Retry | $(System.Collections.Hashtable.spec) | ✅ |
| settings-profile | Add URL | $(System.Collections.Hashtable.spec) | ❌ |
| settings-profile | Update profile | $(System.Collections.Hashtable.spec) | ❌ |
| settings-account | Retry | $(System.Collections.Hashtable.spec) | ✅ |
| settings-account | Update account | $(System.Collections.Hashtable.spec) | ❌ |
| settings-appearance | Retry | $(System.Collections.Hashtable.spec) | ✅ |
| settings-appearance | Update preferences | $(System.Collections.Hashtable.spec) | ❌ |
| settings-notifications | Retry | $(System.Collections.Hashtable.spec) | ❌ |
| settings-notifications | Update notifications | $(System.Collections.Hashtable.spec) | ❌ |
| settings-display | Retry | $(System.Collections.Hashtable.spec) | ❌ |
| settings-display | Clear selection | $(System.Collections.Hashtable.spec) | ❌ |
| settings-display | Update display | $(System.Collections.Hashtable.spec) | ❌ |
| settings-storage | Retry | $(System.Collections.Hashtable.spec) | ✅ |
| settings-storage | Reindex Search | $(System.Collections.Hashtable.spec) | ❌ |
| settings-storage | Rebuild Thumbnails | $(System.Collections.Hashtable.spec) | ❌ |

## Unmatched actions (need spec text/ID refinement)
- dashboard-home: Refresh Dashboard ($(@{screen=dashboard-home; action=Refresh Dashboard; spec=openspec/specs/dashboard/ui-screen-home/spec.md}.spec))
- dashboard-home: Start Setup ($(@{screen=dashboard-home; action=Start Setup; spec=openspec/specs/dashboard/ui-screen-home/spec.md}.spec))
- dashboard-home: Import Existing Collection ($(@{screen=dashboard-home; action=Import Existing Collection; spec=openspec/specs/dashboard/ui-screen-home/spec.md}.spec))
- dashboard-home: Use Sample Data ($(@{screen=dashboard-home; action=Use Sample Data; spec=openspec/specs/dashboard/ui-screen-home/spec.md}.spec))
- dashboard-home: Back Step ($(@{screen=dashboard-home; action=Back Step; spec=openspec/specs/dashboard/ui-screen-home/spec.md}.spec))
- dashboard-home: Next Step ($(@{screen=dashboard-home; action=Next Step; spec=openspec/specs/dashboard/ui-screen-home/spec.md}.spec))
- inventory: Add Item ($(@{screen=inventory; action=Add Item; spec=openspec/specs/inventory/ui-screen-inventory-items/spec.md}.spec))
- inventory: Add Folder ($(@{screen=inventory; action=Add Folder; spec=openspec/specs/inventory/ui-screen-inventory-items/spec.md}.spec))
- inventory: Status ($(@{screen=inventory; action=Status; spec=openspec/specs/inventory/ui-screen-inventory-items/spec.md}.spec))
- inventory: Priority ($(@{screen=inventory; action=Priority; spec=openspec/specs/inventory/ui-screen-inventory-items/spec.md}.spec))
- inventory: View ($(@{screen=inventory; action=View; spec=openspec/specs/inventory/ui-screen-inventory-items/spec.md}.spec))
- inventory: Cards ($(@{screen=inventory; action=Cards; spec=openspec/specs/inventory/ui-screen-inventory-items/spec.md}.spec))
- wishlist: Title ($(@{screen=wishlist; action=Title; spec=openspec/specs/wishlist/ui-screen-wishlist/spec.md}.spec))
- discoveries: Apply Filters ($(@{screen=discoveries; action=Apply Filters; spec=openspec/specs/dashboard/ui-screen-discover/spec.md}.spec))
- scanner: Create Query Set ($(@{screen=scanner; action=Create Query Set; spec=openspec/specs/integrations/ui-screen-scanner/spec.md}.spec))
- integrations: All Integrations ($(@{screen=integrations; action=All Integrations; spec=openspec/specs/integrations/ui-screen-integrations/spec.md}.spec))
- integrations: Rows ($(@{screen=integrations; action=Rows; spec=openspec/specs/integrations/ui-screen-integrations/spec.md}.spec))
- chats: Upload ($(@{screen=chats; action=Upload; spec=openspec/specs/chats/ui-screen-chat-copilot/spec.md}.spec))
- chats: Preview Action ($(@{screen=chats; action=Preview Action; spec=openspec/specs/chats/ui-screen-chat-copilot/spec.md}.spec))
- users: Retry ($(@{screen=users; action=Retry; spec=openspec/specs/users/ui-screen-users/spec.md}.spec))
- reports: Refresh Reports ($(@{screen=reports; action=Refresh Reports; spec=openspec/specs/dashboard/ui-screen-reports/spec.md}.spec))
- reports: Export CSV ($(@{screen=reports; action=Export CSV; spec=openspec/specs/dashboard/ui-screen-reports/spec.md}.spec))
- settings-profile: Add URL ($(@{screen=settings-profile; action=Add URL; spec=openspec/specs/settings/profile/spec.md}.spec))
- settings-profile: Update profile ($(@{screen=settings-profile; action=Update profile; spec=openspec/specs/settings/profile/spec.md}.spec))
- settings-account: Update account ($(@{screen=settings-account; action=Update account; spec=openspec/specs/settings/account/spec.md}.spec))
- settings-appearance: Update preferences ($(@{screen=settings-appearance; action=Update preferences; spec=openspec/specs/settings/appearance/spec.md}.spec))
- settings-notifications: Retry ($(@{screen=settings-notifications; action=Retry; spec=openspec/specs/settings/notifications/spec.md}.spec))
- settings-notifications: Update notifications ($(@{screen=settings-notifications; action=Update notifications; spec=openspec/specs/settings/notifications/spec.md}.spec))
- settings-display: Retry ($(@{screen=settings-display; action=Retry; spec=openspec/specs/settings/display/spec.md}.spec))
- settings-display: Clear selection ($(@{screen=settings-display; action=Clear selection; spec=openspec/specs/settings/display/spec.md}.spec))
- settings-display: Update display ($(@{screen=settings-display; action=Update display; spec=openspec/specs/settings/display/spec.md}.spec))
- settings-storage: Reindex Search ($(@{screen=settings-storage; action=Reindex Search; spec=openspec/specs/settings/storage/spec.md}.spec))
- settings-storage: Rebuild Thumbnails ($(@{screen=settings-storage; action=Rebuild Thumbnails; spec=openspec/specs/settings/storage/spec.md}.spec))
