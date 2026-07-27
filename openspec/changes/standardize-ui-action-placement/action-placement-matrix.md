# Action Placement Matrix

This matrix captures the intended #1941 placement contract for authenticated
Cabinet surfaces. It is the baseline for migration slices; individual page
changes must preserve behavior while moving controls into their canonical
region.

| Surface | Page header actions | Table/list toolbar actions | Row actions | Dialog footer actions | Migration notes |
| --- | --- | --- | --- | --- | --- |
| Dashboard | Global refresh or create shortcuts only when they affect the whole dashboard. | Dashboard-local filters or time ranges. | No row actions unless a repeated activity list exposes records. | Dialog-scoped cancel/confirm only. | Pages without meaningful page actions remain balanced without placeholders. |
| Inventory | New item, import, export, and whole-page refresh. | Search, filters, saved views, view switch, selection, and bulk operations. | Edit, duplicate when supported, restore, archive/delete, permanent delete through shared row menu. | Create/edit/delete confirmations only. | Preserve folder tree, saved views, selected record, and media dialogs. |
| Collections | New collection and whole-page refresh/export when available. | Collection filters and view controls. | Open/view, edit, delete via shared row menu. | Destination-aware delete and edit/create controls only. | Remove exposed row icon clusters after row-menu parity exists. |
| Wishlist | Add/import wanted item and whole-page refresh/export. | Search, filters, saved views, view switch, selection, and bulk operations. | Edit, promote/purchase workflow actions, restore, delete/permanent delete through shared row menu. | Wishlist create/edit/import confirmation controls only. | Bulk delete stays in toolbar, not per-row menu. |
| Media | Add/upload media and whole-page refresh/import/export. | Asset filters, linkage filters, search, view switch, and bulk operations. | Open, analyze, assign, archive/delete through shared row menu when row capabilities allow. | Upload, metadata, analyze, assign, and archive confirmation controls only. | Capability-gated Analyze/Assign availability must stay truthful. |
| Purchases | New purchase, import/export, and whole-page refresh. | Purchase filters, reconciliation filters, selection, and bulk actions. | Order and line-item operations through shared row/detail action menu. | Draft, review, reconcile, and queue-action confirms only. | Name whether the target is an order or line item. |
| Integrations | Add/connect provider and whole-page refresh. | Provider search, filters, and status controls. | View details, edit/connect, validate, disconnect when supported through shared row menu. | Provider configuration and disconnect confirmation controls only. | Manifest actions remain capability-gated and must not mix with shell utilities. |
| Chats | New chat and whole-page refresh/export when available. | Conversation search, filters, and selection controls. | Conversation-level archive/delete/export through row menu when surfaced as a list. | Chat compose and confirmation dialog controls only. | Composer send controls are content actions, not page actions. |
| Inbox | Mark all read, clear/archive selected, and whole-page refresh. | Notification filters, search, selection, and bulk actions. | Mark read/unread, open source, archive/delete through shared row menu. | Confirmation controls only. | Bulk notification state changes remain toolbar actions. |
| Discoveries | Run discovery, import/export, and whole-page refresh. | Provider/query filters, result status filters, selection, and bulk actions. | Review source, restore, ignore/archive, promote to Wishlist, purchase follow-up, and Inventory handoff through row menu. | Promotion, handoff, and confirmation controls only. | Unsupported handoff paths are omitted rather than disabled without reason. |
| Reports | Refresh, export, and whole-page report actions. | Report filters, date ranges, grouping, and table-specific refresh. | Report row drilldown/export only when a repeated record table exists. | Report export/confirmation controls only. | Remove duplicated title/description blocks once header actions own the page controls. |
| Market Watch | New query, run selected/global query set, import/export, and whole-page refresh. | Query filters, provider/scope filters, search, view controls, and bulk query actions. | Inspect output, run now, pause/resume, edit, delete through shared row menu. | Query create/edit, output inspection, and destructive confirmation controls only. | Align visible title with Market Watch and remove duplicate local action blocks. |
| Settings Profile | Save/reset profile-level settings when they apply to the whole child page. | Section filters or repeated-list controls only. | Repeated record rows, such as profile-scoped lists, use shared row menu. | Settings form submit/cancel belongs to the active form/dialog. | Settings child routes retain their specific header metadata from #1940. |
| Settings Account | Account save/reset page actions only. | Section filters or repeated-list controls only. | Repeated account records use shared row menu. | Dialog/footer controls only. | Security-sensitive actions keep confirmation copy near the active dialog. |
| Settings Appearance | Save/reset appearance and display actions only. | Sidebar item multi-select and view controls stay in the form region. | Repeated display rows use shared row menu only if introduced. | Form submit/cancel controls only. | Do not convert preference controls into page-header buttons. |
| Settings Billing | Billing refresh/manage actions that apply to the whole page. | Billing history filters and list controls. | Invoice/subscription row operations through shared row menu when present. | Billing dialog controls only. | External billing redirects require explicit labels. |
| Settings Categories | Save taxonomy settings and add high-level taxonomy groups. | Category, condition, and packaging list controls. | Remove/edit repeated taxonomy rows through row menu when migrated. | Form/dialog controls only. | Inline remove-before-save behavior may stay local until row-menu migration is safe. |
| Settings Integrations | Save provider settings or refresh setup state for the child page. | Provider setup filters and section controls. | Provider rows use shared row action contract. | Provider setup dialog controls only. | Avoid duplicating the main Integrations page action model. |
| Settings Notifications | Save/reset notification preferences and test delivery when page-wide. | Notification channel filters and repeated-list controls. | Channel rows use shared row menu when present. | Dialog/footer controls only. | Test delivery must expose loading and result states. |
| Settings Operations | Save/reset operations settings, pause/resume all workers when whole-page. | Worker/queue filters and repeated queue controls. | Queue or worker row operations through row menu when present. | Confirmation controls only. | Dangerous runtime controls need clear confirmation copy. |
| Settings Skills | Import skill, refresh, and whole-page actions. | Skill filters, search, and selection/bulk controls. | Enable/disable, details, quarantine/remove through shared row menu when row capability supports it. | Skill import/details confirmation controls only. | Preserve governed skill lifecycle boundaries. |
| Settings Storage | Backup/export/restore actions that apply to storage page state. | Backup list filters and selection/bulk controls. | Download, restore, delete backup through shared row menu. | Restore/delete confirmation controls only. | Restore remains confirmation-gated. |
| Users | Add/invite user and whole-page refresh/export. | User filters, search, role/status filters, selection, and bulk actions. | View/edit/delete through shared row menu, with protected owner capability omission. | Invite, edit, and delete confirmation controls only. | Owner safeguards remain capability-driven. |
| Help Center | No placeholder page action unless help-wide search/export exists. | Article search, category filters, and table-of-contents controls. | Article row/open actions only where repeated lists expose row operations. | Feedback/contact dialog controls only. | Help pages should stay readable rather than button-heavy. |

## Shared Action Region Rules

- Whole-page actions must appear in the global page header and have one
  implementation.
- Table/list toolbars own filters, search, sort, view switches, selection,
  bulk actions, and table-scoped refresh.
- Row actions must use the shared #1938/#1939 record action menu when the
  surface is a repeated record list or table.
- Dialog footers own cancel, confirm, apply, save, and destructive confirmation
  controls for the active dialog only.
- Shell utilities such as language, theme, configuration, profile, and sidebar
  controls are not page actions.
- Responsive overflow must preserve primary action access, focus order,
  accessible names, disabled/loading states, and non-overlap.
