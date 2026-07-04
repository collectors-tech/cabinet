# Agent Skill Coverage Matrix

Issue: #1702
Parent: #1701
Related registry epic: #1666

This matrix keeps Cabinet Agent coverage concrete by surface. It is a planning and traceability artifact, not a claim that every listed skill is implemented. Status values are:

- `implemented`: backed by merged code and direct validation evidence.
- `partial`: some backing contract or fixture evidence exists, but runtime or workflow coverage is incomplete.
- `planned`: tracked and ready for implementation.
- `blocked`: known dependency prevents truthful implementation.
- `deferred`: explicitly out of current local registry scope.

Safety levels are `read-only`, `preview-only`, `confirm-required`, `external-write`, and `destructive`.

## Channel and Entry Coverage

| Entry point | Covered context | Status | Linked issues | Validation evidence |
| --- | --- | --- | --- | --- |
| Main Chat `/chats` | Profile, thread, user intent, attachments, preview/apply state | partial | #1703, #1716 | `internal/app/chat_api_test.go`; `docs/validation/agent-acceptance-suite.md` |
| Side-panel Chat | Profile, route, active thread, selected surface context, attachments | partial | #1703, #1714, #1716 | `ui.web/cypress/e2e/chats/assistant-workspace/spec.cy.ts`; #1714 planned context envelope |
| Inbox review | Notification/review item, source thread, preview handoff | planned | #1715 | No executable skill coverage yet; planned under #1715 |
| Telegram/external channel | Authorized sender/chat, source message/media, profile mapping, review link | partial | #1704, #1705, #1706, #1773 | Fixture/proof-packet evidence in #1716; live-channel checklist remains #1773 |
| Skills page detail/actions | Skill id, source, status, safety, enable/disable/import context | planned | #1670 | Planned Cypress: `ui.web/cypress/e2e/settings/agent-skills/spec.cy.ts` |

## Surface Matrix

| Surface | Expected skill ids | User-facing request examples | Safety | Required context/selection | Required setup | Bound capability ids | Bound guided workflow ids | Status | Missing issue/PR links | Validation evidence | Blocked/deferred reason |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Dashboard | `cabinet.navigate.open_surface`, `cabinet.dashboard.summarise_activity` | "Show my dashboard", "Summarise what changed today" | read-only | Active profile | None | navigation/open-surface, dashboard-summary planned | None | planned | #1714, follow-up needed for dashboard summary skill | None yet | Summary skill is not yet specified as a child issue |
| Inventory | `cabinet.inventory.search_items`, `cabinet.inventory.create_item`, `cabinet.inventory.update_item`, `cabinet.inventory.attach_media`, `cabinet.inventory.assign_to_collection`, `cabinet.guided.inventory.update_item` | "Find boxed kits", "Create this item", "Update the selected item", "Attach this photo" | read-only, preview-only, confirm-required | Profile; item selection for update/media/assignment; collection selection for assignment | Media context for attachment | inventory-search planned, inventory-create/update built-in registry seeds, media-attachment planned | `guided.inventory.update_item` | partial | #1707; #1513 for guided update walkthrough | `internal/agentskills/registry_test.go` covers seed/status; execution tests planned | Guided update remains blocked until #1513 |
| Wishlist | `cabinet.wishlist.search_entries`, `cabinet.wishlist.create_entry`, `cabinet.wishlist.update_entry`, `cabinet.wishlist.mark_purchased`, `cabinet.wishlist.soft_delete_entry`, `cabinet.wishlist.restore_entry` | "Add this to wishlist", "Mark this wanted item purchased" | read-only, preview-only, confirm-required | Profile; wishlist entry or item intent | Inventory sync rules for purchase handoff | wishlist-create built-in registry seed, wishlist mutation planned | None | planned | #1708 | Registry seed only for `cabinet.wishlist.create_entry`; execution tests planned | Purchase sync must prove no duplicate inventory quantity increments |
| Collections | `cabinet.collections.search`, `cabinet.collections.create`, `cabinet.collections.update_metadata`, `cabinet.collections.assign_item`, `cabinet.collections.soft_delete`, `cabinet.collections.move_items_on_delete` | "Create a collection", "Move selected item to this collection", "Delete this collection safely" | read-only, preview-only, confirm-required, destructive | Profile; collection/item selection | Collection state and `All Items` protection | collection-assign planned | None | planned | #1708 | Existing collection API/UI tests cover product behavior, not Agent skill execution | Agent skill wrapper still missing |
| Media | `cabinet.media.search`, `cabinet.media.upload_or_import`, `cabinet.media.attach_to_item`, `cabinet.media.review_unlinked`, `cabinet.media.update_notes`, `cabinet.media.detach_from_item` | "Find unlinked images", "Attach this media to the selected item" | read-only, preview-only, confirm-required | Profile; media selection; target item selection | Explicit upload/import or existing media | media-search/attach planned | None | planned | #1709 | Attachment plumbing covered through #1716; media-specific Agent skill tests planned | Media provenance handoff must be proven before apply claims |
| Discoveries | `cabinet.discoveries.search`, `cabinet.discoveries.review_result`, `cabinet.discoveries.dismiss_result`, `cabinet.discoveries.send_to_wishlist`, `cabinet.discoveries.create_purchase`, `cabinet.discoveries.create_or_update_inventory_candidate` | "Review new discoveries", "Send this result to wishlist" | read-only, preview-only, confirm-required | Profile; discovery/result selection | Provider/source result provenance | discovery-review/handoff planned | None | planned | #1709 | Provider/discovery product coverage exists outside Agent execution | Skill wrapper and provenance-preserving handoff remain planned |
| Market Watch / Scanner | `cabinet.market_watch.search_watches`, `cabinet.market_watch.create_saved_watch`, `cabinet.market_watch.update_saved_watch`, `cabinet.market_watch.run_watch`, `cabinet.market_watch.review_results`, `cabinet.market_watch.dismiss_result`, `cabinet.market_watch.handoff_result` | "Run this saved watch", "Create a watch for this search" | read-only, preview-only, confirm-required, external-write | Profile; saved watch/result selection | Provider readiness and health | market-watch planned, scanner-provider planned | None | planned | #1710 | Provider/scanner tests cover underlying behavior, not Agent skill execution | Provider health/setup must gate run/handoff |
| Purchases | `cabinet.purchases.search_orders`, `cabinet.purchases.create_order`, `cabinet.purchases.add_line_item`, `cabinet.purchases.receive_order`, `cabinet.purchases.receive_line_item`, `cabinet.purchases.reconcile_item`, `cabinet.purchases.review_purchase` | "Create a purchase order", "Receive this line item" | read-only, preview-only, confirm-required | Profile; order/line/item selection | Purchase workflow state | purchases planned | None | planned | #1710 | Purchase product tests exist outside Agent execution | Receiving/reconciliation needs preview/apply and item identity proof |
| Integrations | `cabinet.integrations.search_providers`, `cabinet.integrations.configure_provider`, `cabinet.integrations.test_connection`, `cabinet.integrations.repair_provider`, `cabinet.integrations.disable_provider`, `cabinet.integrations.explain_required_setup` | "Explain eBay setup", "Test this provider", "Disable this integration" | read-only, preview-only, confirm-required, external-write | Profile; provider selection | Provider credentials and health where applicable | integrations-provider planned | None | planned | #1711; registry migration blockers #1463-#1482 | Existing provider tests cover setup/status; Agent skill tests planned | Secrets must be redacted in chat and audit evidence |
| Inbox | `cabinet.inbox.search_notifications`, `cabinet.inbox.summarise_unhandled`, `cabinet.inbox.open_notification`, `cabinet.inbox.mark_handled`, `cabinet.inbox.archive_or_hide`, `cabinet.inbox.route_to_surface` | "Summarise unhandled Inbox items", "Mark this handled" | read-only, preview-only, confirm-required | Profile; notification selection for mutation | None | notification-inbox planned | None | planned | #1715 | Existing Inbox UI/API tests cover product behavior, not Agent skill execution | Skill execution must prove preview/confirm/apply for state changes |
| Users / workspace admin | `cabinet.users.search`, `cabinet.users.invite_user`, `cabinet.users.resend_invitation`, `cabinet.users.update_role`, `cabinet.users.activate_or_deactivate`, `cabinet.users.remove_user` | "Invite this user", "Deactivate this account" | read-only, confirm-required, destructive | Profile/workspace admin context; exact target user | Backend support and role permissions | users-admin planned | None | planned | #1715 | No Agent skill tests yet | Protected owner/local admin rules must be enforced before mutation |
| Settings / Profile | `cabinet.settings.update_profile` | "Update my profile display settings" | preview-only, confirm-required | Active profile | None | settings-profile planned | None | planned | #1711 | Settings product tests exist outside Agent execution | Agent wrapper and confirmation copy missing |
| Settings / Account | `cabinet.settings.update_account` | "Change account preferences" | preview-only, confirm-required, destructive where applicable | Authenticated account context | Account capability availability | settings-account planned | None | planned | #1711 | None yet | Risk varies by account operation and must be split before execution |
| Settings / Appearance | `cabinet.settings.update_appearance` | "Switch appearance mode" | preview-only, confirm-required | Active profile | None | settings-appearance planned | None | planned | #1711 | Settings preference tests exist outside Agent execution | Skill wrapper missing |
| Settings / Storage | `cabinet.storage.show_status`, `cabinet.storage.configure_backup` | "Show storage status", "Configure backup" | read-only, preview-only, confirm-required | Active profile/storage context | Local storage/backup availability | storage-status planned | None | planned | #1711 | Data-management/storage tests exist outside Agent execution | Backup setup must explain write target before apply |
| Import / Export / Backup / Restore / Maintenance | `cabinet.data.import_file`, `cabinet.data.export_bundle`, `cabinet.data.restore_backup`, `cabinet.maintenance.run_safe_check` | "Export my data", "Import this file", "Run a safe maintenance check" | read-only, preview-only, confirm-required, destructive | Profile; selected file for import/restore | Explicit file selection; storage readiness | data-management planned | None | planned | #1711 | Data-management tests cover underlying APIs; Agent skill tests planned | Import/restore are high-risk and require strong confirmation |
| Chats / Agent itself | `cabinet.chat.action_timeline.view`, `cabinet.navigate.open_surface`, `cabinet.agent.explain_available_work` | "What can you do here?", "Show the action timeline", "Open inventory" | read-only, preview-only | Profile; thread; route/source context | None | chat-action-timeline planned, navigate-open-surface built-in seed | None | partial | #1703, #1714, #1716 | `docs/validation/agent-acceptance-suite.md`; `internal/app/chat_api_test.go` | Universal context envelope remains planned in #1714 |

## Current Built-In Registry Seed Coverage

The current built-in registry evidence covers only a seed subset of the full matrix:

- `cabinet.navigate.open_surface`
- `cabinet.inventory.create_item`
- `cabinet.inventory.update_item`
- `cabinet.wishlist.create_entry`
- `cabinet.guided.inventory.update_item`

These seed entries are not enough to close #1701 or any child skill issue by themselves. Child issues #1707, #1708, #1709, #1710, #1711, and #1715 own execution behavior, preview/apply tests, UI/channel coverage where applicable, and closure evidence for each surface group.

## Explicit Gaps

- Dashboard summary skill has no focused child issue yet and should be split before implementation.
- Main Chat, side-panel Chat, Inbox review, and Telegram/external channel entry points have different context and proof requirements; validation must not collapse them into one generic Agent path.
- Live Telegram production-channel validation is tracked by #1773 and remains separate from fixture/proof-packet evidence.
- Marketplace discovery/publishing/payments/reviews remain deferred and must not be implied by local skill import or the Skills page.
