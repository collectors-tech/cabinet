package agentskills

import (
	"slices"
	"testing"
)

func TestSkillRegistryListsAndResolvesBuiltInAndImportedSkills(t *testing.T) {
	t.Parallel()

	registry := NewRegistry([]Skill{{
		ID:              "local.archive.read_only",
		Version:         "0.1.0",
		DisplayName:     "Local read-only archive",
		Description:     "Read local skill fixture metadata.",
		Category:        "testing",
		Status:          StatusAvailable,
		SafetyLevel:     SafetyReadOnly,
		RequiredContext: []string{"profile"},
		Enabled:         true,
	}})

	skills := registry.List()
	if len(skills) < 20 {
		t.Fatalf("expected built-in plus imported skills, got %d", len(skills))
	}
	for _, id := range []string{
		"cabinet.navigate.open_surface",
		"cabinet.inventory.search_items",
		"cabinet.inventory.create_item",
		"cabinet.inventory.update_item",
		"cabinet.inventory.attach_media",
		"cabinet.inventory.assign_to_collection",
		"cabinet.wishlist.search_entries",
		"cabinet.wishlist.create_entry",
		"cabinet.wishlist.update_entry",
		"cabinet.wishlist.mark_purchased",
		"cabinet.wishlist.soft_delete_entry",
		"cabinet.wishlist.restore_entry",
		"cabinet.collections.search",
		"cabinet.collections.create",
		"cabinet.collections.update_metadata",
		"cabinet.collections.assign_item",
		"cabinet.collections.soft_delete",
		"cabinet.collections.move_items_on_delete",
		"cabinet.collection.assign_item",
		"cabinet.guided.inventory.update_item",
		"cabinet.chat.action_timeline.view",
		"cabinet.inbox.search_notifications",
		"cabinet.inbox.summarise_unhandled",
		"cabinet.inbox.open_notification",
		"cabinet.inbox.mark_handled",
		"cabinet.inbox.archive_or_hide",
		"cabinet.inbox.route_to_surface",
		"cabinet.users.search",
		"cabinet.users.invite_user",
		"cabinet.users.resend_invitation",
		"cabinet.users.update_role",
		"cabinet.users.activate_or_deactivate",
		"cabinet.users.remove_user",
		"cabinet.integrations.search_providers",
		"cabinet.integrations.configure_provider",
		"cabinet.integrations.test_connection",
		"cabinet.integrations.repair_provider",
		"cabinet.integrations.disable_provider",
		"cabinet.integrations.explain_required_setup",
		"cabinet.settings.update_profile",
		"cabinet.settings.update_account",
		"cabinet.settings.update_appearance",
		"cabinet.storage.show_status",
		"cabinet.storage.configure_backup",
		"cabinet.data.import_file",
		"cabinet.data.export_bundle",
		"cabinet.data.restore_backup",
		"cabinet.maintenance.run_safe_check",
		"cabinet.market_watch.search_watches",
		"cabinet.market_watch.create_saved_watch",
		"cabinet.market_watch.update_saved_watch",
		"cabinet.market_watch.run_watch",
		"cabinet.market_watch.review_results",
		"cabinet.market_watch.dismiss_result",
		"cabinet.market_watch.handoff_result",
		"cabinet.purchases.search_orders",
		"cabinet.purchases.create_order",
		"cabinet.purchases.add_line_item",
		"cabinet.purchases.receive_order",
		"cabinet.purchases.receive_line_item",
		"cabinet.purchases.reconcile_item",
		"cabinet.purchases.review_purchase",
		"cabinet.media.search",
		"cabinet.media.upload_or_import",
		"cabinet.media.attach_to_item",
		"cabinet.media.review_unlinked",
		"cabinet.media.update_notes",
		"cabinet.media.detach_from_item",
		"cabinet.discoveries.search",
		"cabinet.discoveries.review_result",
		"cabinet.discoveries.dismiss_result",
		"cabinet.discoveries.send_to_wishlist",
		"cabinet.discoveries.create_purchase",
		"cabinet.discoveries.create_or_update_inventory_candidate",
		"local.archive.read_only",
	} {
		if !containsSkill(skills, id) {
			t.Fatalf("expected registry to contain %s", id)
		}
	}

	resolved, ok := registry.Resolve("cabinet.inventory.update_item")
	if !ok {
		t.Fatalf("expected to resolve built-in inventory update skill")
	}
	if resolved.Source != SourceBuiltIn || !resolved.BuiltIn || resolved.Removable {
		t.Fatalf("expected immutable built-in source metadata, got %+v", resolved)
	}
	if !slices.Contains(resolved.Capabilities, "inventory.item.update") || !slices.Contains(resolved.Capabilities, "update_open_item_title") {
		t.Fatalf("expected inventory update capability bindings, got %+v", resolved.Capabilities)
	}
	if resolved.SafetyLevel != SafetyConfirmRequired || !resolved.Permissions.RequiresConfirm || !resolved.Permissions.LocalWrite {
		t.Fatalf("expected confirm-required local write permission metadata, got %+v", resolved)
	}
}

func TestInventorySkillsExposePreviewBoundaries(t *testing.T) {
	t.Parallel()

	registry := NewRegistry(nil)

	searchItems, ok := registry.Resolve("cabinet.inventory.search_items")
	if !ok {
		t.Fatalf("expected Inventory search skill")
	}
	if searchItems.SafetyLevel != SafetyReadOnly || !searchItems.Executable || searchItems.Permissions.LocalWrite {
		t.Fatalf("Inventory search should be executable read-only metadata, got %+v", searchItems)
	}
	if !slices.Contains(searchItems.IntegrationWorkflows, "inventory.item.search") {
		t.Fatalf("expected Inventory search workflow binding, got %+v", searchItems.IntegrationWorkflows)
	}

	createItem, ok := registry.Resolve("cabinet.inventory.create_item")
	if !ok {
		t.Fatalf("expected Inventory create skill")
	}
	if createItem.SafetyLevel != SafetyConfirmRequired || !createItem.Permissions.RequiresConfirm || !slices.Contains(createItem.InputSchemaRefs, "part_number") {
		t.Fatalf("create item should require item context and confirmation, got %+v", createItem)
	}
	missingCreateContext, err := registry.Preview(PreviewRequest{
		SkillID:    "cabinet.inventory.create_item",
		Parameters: map[string]any{"title": "Agent inventory item"},
	})
	if err != nil {
		t.Fatalf("preview inventory create item: %v", err)
	}
	if missingCreateContext.Allowed || missingCreateContext.Blocker != "inventory_item_context_required" || missingCreateContext.MutationApplied {
		t.Fatalf("expected Inventory item context blocker without mutation, got %+v", missingCreateContext)
	}

	updateItem, ok := registry.Resolve("cabinet.inventory.update_item")
	if !ok {
		t.Fatalf("expected Inventory update skill")
	}
	if updateItem.SafetyLevel != SafetyConfirmRequired || !slices.Contains(updateItem.RequiredContext, "selected_item") {
		t.Fatalf("update item should require selected item and confirmation, got %+v", updateItem)
	}
	missingItem, err := registry.Preview(PreviewRequest{SkillID: "cabinet.inventory.update_item"})
	if err != nil {
		t.Fatalf("preview inventory update item: %v", err)
	}
	if missingItem.Allowed || missingItem.Blocker != "inventory_item_required" {
		t.Fatalf("expected Inventory item blocker, got %+v", missingItem)
	}

	missingMedia, err := registry.Preview(PreviewRequest{
		SkillID:    "cabinet.inventory.attach_media",
		Parameters: map[string]any{"item_id": "item-1"},
	})
	if err != nil {
		t.Fatalf("preview inventory attach media: %v", err)
	}
	if missingMedia.Allowed || missingMedia.Blocker != "inventory_media_required" {
		t.Fatalf("expected Inventory media blocker, got %+v", missingMedia)
	}

	invalidCollection, err := registry.Preview(PreviewRequest{
		SkillID:    "cabinet.inventory.assign_to_collection",
		Parameters: map[string]any{"item_id": "item-1", "collection_name": "Deleted"},
	})
	if err != nil {
		t.Fatalf("preview inventory assign collection: %v", err)
	}
	if invalidCollection.Allowed || invalidCollection.Blocker != "inventory_collection_invalid" {
		t.Fatalf("expected invalid collection blocker, got %+v", invalidCollection)
	}
}

func TestWishlistAndCollectionsSkillsExposePreviewBoundaries(t *testing.T) {
	t.Parallel()

	registry := NewRegistry(nil)

	searchWishlist, ok := registry.Resolve("cabinet.wishlist.search_entries")
	if !ok {
		t.Fatalf("expected Wishlist search skill")
	}
	if searchWishlist.SafetyLevel != SafetyReadOnly || !searchWishlist.Executable || searchWishlist.Permissions.LocalWrite {
		t.Fatalf("Wishlist search should be executable read-only metadata, got %+v", searchWishlist)
	}

	markPurchased, ok := registry.Resolve("cabinet.wishlist.mark_purchased")
	if !ok {
		t.Fatalf("expected Wishlist mark purchased skill")
	}
	if markPurchased.SafetyLevel != SafetyConfirmRequired || !markPurchased.Permissions.RequiresConfirm || !slices.Contains(markPurchased.OutputSchemaRefs, "inventory_quantity_sync") {
		t.Fatalf("mark purchased should expose confirm-required purchase/inventory sync metadata, got %+v", markPurchased)
	}
	missingEntry, err := registry.Preview(PreviewRequest{SkillID: "cabinet.wishlist.mark_purchased"})
	if err != nil {
		t.Fatalf("preview wishlist mark purchased: %v", err)
	}
	if missingEntry.Allowed || missingEntry.Blocker != "wishlist_entry_required" || missingEntry.MutationApplied {
		t.Fatalf("expected Wishlist entry blocker without mutation, got %+v", missingEntry)
	}

	createEntry, err := registry.Preview(PreviewRequest{
		SkillID:    "cabinet.wishlist.create_entry",
		Parameters: map[string]any{"title": "Wanted slot car"},
	})
	if err != nil {
		t.Fatalf("preview wishlist create entry: %v", err)
	}
	if createEntry.Allowed || createEntry.Blocker != "wishlist_item_context_required" {
		t.Fatalf("expected Wishlist item context blocker, got %+v", createEntry)
	}

	searchCollections, ok := registry.Resolve("cabinet.collections.search")
	if !ok {
		t.Fatalf("expected Collections search skill")
	}
	if searchCollections.SafetyLevel != SafetyReadOnly || !searchCollections.Executable || searchCollections.Permissions.LocalWrite {
		t.Fatalf("Collections search should be executable read-only metadata, got %+v", searchCollections)
	}

	assignItem, ok := registry.Resolve("cabinet.collections.assign_item")
	if !ok {
		t.Fatalf("expected Collections assign item skill")
	}
	if assignItem.SafetyLevel != SafetyConfirmRequired || !assignItem.Permissions.RequiresConfirm || !slices.Contains(assignItem.InputSchemaRefs, "item_id") {
		t.Fatalf("assign item should require item context and confirmation, got %+v", assignItem)
	}
	missingItem, err := registry.Preview(PreviewRequest{
		SkillID:    "cabinet.collections.assign_item",
		Parameters: map[string]any{"collection_name": "Display Case"},
	})
	if err != nil {
		t.Fatalf("preview collections assign item: %v", err)
	}
	if missingItem.Allowed || missingItem.Blocker != "collections_item_required" {
		t.Fatalf("expected Collections item blocker, got %+v", missingItem)
	}
	legacyAssignItem, ok := registry.Resolve("cabinet.collection.assign_item")
	if !ok {
		t.Fatalf("expected legacy Collection assign item skill")
	}
	if legacyAssignItem.SafetyLevel != SafetyConfirmRequired ||
		!slices.Contains(legacyAssignItem.IntegrationWorkflows, "collections.item.assign") ||
		!slices.Contains(legacyAssignItem.InputSchemaRefs, "collection_name") {
		t.Fatalf("legacy collection assign skill should expose governed assignment metadata, got %+v", legacyAssignItem)
	}
	legacyMissingCollection, err := registry.Preview(PreviewRequest{
		SkillID:    "cabinet.collection.assign_item",
		Parameters: map[string]any{"item_id": "item-1"},
	})
	if err != nil {
		t.Fatalf("preview legacy collection assign item: %v", err)
	}
	if legacyMissingCollection.Allowed || legacyMissingCollection.Blocker != "collections_target_required" {
		t.Fatalf("expected legacy Collections target blocker, got %+v", legacyMissingCollection)
	}

	protectedAllItems, err := registry.Preview(PreviewRequest{
		SkillID:    "cabinet.collections.soft_delete",
		Parameters: map[string]any{"collection_name": "All Items"},
	})
	if err != nil {
		t.Fatalf("preview collections soft delete: %v", err)
	}
	if protectedAllItems.Allowed || protectedAllItems.Blocker != "collections_all_items_protected" {
		t.Fatalf("expected All Items protection blocker, got %+v", protectedAllItems)
	}

	missingDestination, err := registry.Preview(PreviewRequest{
		SkillID: "cabinet.collections.soft_delete",
		Parameters: map[string]any{
			"collection_name": "Display Case",
			"has_items":       true,
		},
	})
	if err != nil {
		t.Fatalf("preview collections soft delete with items: %v", err)
	}
	if missingDestination.Allowed || missingDestination.Blocker != "collections_delete_destination_required" {
		t.Fatalf("expected missing destination blocker for non-empty collection delete, got %+v", missingDestination)
	}
}

func TestMarketWatchAndPurchasesSkillsExposePreviewAndProvenanceBoundaries(t *testing.T) {
	t.Parallel()

	registry := NewRegistry(nil)

	searchWatches, ok := registry.Resolve("cabinet.market_watch.search_watches")
	if !ok {
		t.Fatalf("expected Market Watch search skill")
	}
	if searchWatches.SafetyLevel != SafetyReadOnly || !searchWatches.Executable || searchWatches.Permissions.LocalWrite {
		t.Fatalf("Market Watch search should be executable read-only metadata, got %+v", searchWatches)
	}
	if !slices.Contains(searchWatches.RequiredProviders, "provider-registry") || !slices.Contains(searchWatches.IntegrationWorkflows, "market_watch.watch.search") {
		t.Fatalf("expected provider-backed Market Watch workflow binding, got providers=%+v workflows=%+v", searchWatches.RequiredProviders, searchWatches.IntegrationWorkflows)
	}

	runWatch, ok := registry.Resolve("cabinet.market_watch.run_watch")
	if !ok {
		t.Fatalf("expected Market Watch run skill")
	}
	if runWatch.SafetyLevel != SafetyConfirmRequired || !runWatch.Permissions.RequiresConfirm || !runWatch.Permissions.ExternalWrite {
		t.Fatalf("run watch should be confirmation-gated external-read/write metadata, got %+v", runWatch)
	}
	missingProvider, err := registry.Preview(PreviewRequest{
		SkillID:    "cabinet.market_watch.run_watch",
		Parameters: map[string]any{"watch_id": "watch-1"},
	})
	if err != nil {
		t.Fatalf("preview run watch missing provider: %v", err)
	}
	if missingProvider.Allowed || missingProvider.Blocker != "market_watch_provider_required" || missingProvider.MutationApplied {
		t.Fatalf("expected provider readiness blocker without mutation, got %+v", missingProvider)
	}

	missingResult, err := registry.Preview(PreviewRequest{
		SkillID:    "cabinet.market_watch.handoff_result",
		Parameters: map[string]any{"provider_id": "ebay"},
	})
	if err != nil {
		t.Fatalf("preview handoff missing result: %v", err)
	}
	if missingResult.Allowed || missingResult.Blocker != "market_watch_result_required" {
		t.Fatalf("expected Market Watch result blocker, got %+v", missingResult)
	}

	searchOrders, ok := registry.Resolve("cabinet.purchases.search_orders")
	if !ok {
		t.Fatalf("expected Purchases search skill")
	}
	if searchOrders.SafetyLevel != SafetyReadOnly || !searchOrders.Executable || searchOrders.Permissions.LocalWrite {
		t.Fatalf("Purchases search should be executable read-only metadata, got %+v", searchOrders)
	}

	addLineItem, ok := registry.Resolve("cabinet.purchases.add_line_item")
	if !ok {
		t.Fatalf("expected Purchases add line item skill")
	}
	if addLineItem.SafetyLevel != SafetyConfirmRequired || !slices.Contains(addLineItem.RequiredContext, "target_order") || !addLineItem.Permissions.RequiresConfirm {
		t.Fatalf("add line item should require target order and confirmation, got %+v", addLineItem)
	}
	missingOrder, err := registry.Preview(PreviewRequest{
		SkillID:    "cabinet.purchases.add_line_item",
		Parameters: map[string]any{"item_id": "item-1"},
	})
	if err != nil {
		t.Fatalf("preview purchase add line item: %v", err)
	}
	if missingOrder.Allowed || missingOrder.Blocker != "purchases_order_required" {
		t.Fatalf("expected purchase order blocker without mutation, got %+v", missingOrder)
	}

	missingItem, err := registry.Preview(PreviewRequest{
		SkillID:    "cabinet.purchases.reconcile_item",
		Parameters: map[string]any{"order_id": "order-1"},
	})
	if err != nil {
		t.Fatalf("preview purchase reconcile item: %v", err)
	}
	if missingItem.Allowed || missingItem.Blocker != "purchases_item_required" {
		t.Fatalf("expected purchase item blocker without mutation, got %+v", missingItem)
	}
}

func TestMediaAndDiscoveriesSkillsExposePreviewAndProvenanceBoundaries(t *testing.T) {
	t.Parallel()

	registry := NewRegistry(nil)

	searchMedia, ok := registry.Resolve("cabinet.media.search")
	if !ok {
		t.Fatalf("expected Media search skill")
	}
	if searchMedia.SafetyLevel != SafetyReadOnly || !searchMedia.Executable || searchMedia.Permissions.LocalWrite {
		t.Fatalf("Media search should be executable read-only metadata, got %+v", searchMedia)
	}
	if !slices.Contains(searchMedia.Capabilities, "media.workflow") || !slices.Contains(searchMedia.IntegrationWorkflows, "media.search") {
		t.Fatalf("expected Media workflow binding, got capabilities=%+v workflows=%+v", searchMedia.Capabilities, searchMedia.IntegrationWorkflows)
	}

	attachMedia, ok := registry.Resolve("cabinet.media.attach_to_item")
	if !ok {
		t.Fatalf("expected Media attach skill")
	}
	if attachMedia.SafetyLevel != SafetyConfirmRequired || !attachMedia.Permissions.RequiresConfirm || !slices.Contains(attachMedia.RequiredContext, "target_item") {
		t.Fatalf("attach media should require target item and confirmation, got %+v", attachMedia)
	}
	missingMedia, err := registry.Preview(PreviewRequest{
		SkillID:    "cabinet.media.attach_to_item",
		Parameters: map[string]any{"item_id": "item-1"},
	})
	if err != nil {
		t.Fatalf("preview attach media missing media: %v", err)
	}
	if missingMedia.Allowed || missingMedia.Blocker != "media_target_required" || missingMedia.MutationApplied {
		t.Fatalf("expected missing media blocker without mutation, got %+v", missingMedia)
	}

	missingItem, err := registry.Preview(PreviewRequest{
		SkillID:    "cabinet.media.detach_from_item",
		Parameters: map[string]any{"media_id": "media-1"},
	})
	if err != nil {
		t.Fatalf("preview detach media missing item: %v", err)
	}
	if missingItem.Allowed || missingItem.Blocker != "media_item_required" {
		t.Fatalf("expected missing item blocker without mutation, got %+v", missingItem)
	}

	searchDiscoveries, ok := registry.Resolve("cabinet.discoveries.search")
	if !ok {
		t.Fatalf("expected Discoveries search skill")
	}
	if searchDiscoveries.SafetyLevel != SafetyReadOnly || !searchDiscoveries.Executable || searchDiscoveries.Permissions.LocalWrite {
		t.Fatalf("Discoveries search should be executable read-only metadata, got %+v", searchDiscoveries)
	}
	if !slices.Contains(searchDiscoveries.RequiredProviders, "provider-registry") || !searchDiscoveries.Permissions.ExternalRead {
		t.Fatalf("expected provider-backed Discoveries read metadata, got providers=%+v permissions=%+v", searchDiscoveries.RequiredProviders, searchDiscoveries.Permissions)
	}

	sendToWishlist, ok := registry.Resolve("cabinet.discoveries.send_to_wishlist")
	if !ok {
		t.Fatalf("expected Discoveries wishlist handoff skill")
	}
	if sendToWishlist.SafetyLevel != SafetyConfirmRequired || !sendToWishlist.Permissions.RequiresConfirm {
		t.Fatalf("discovery handoff should require confirmation, got %+v", sendToWishlist)
	}
	missingProvider, err := registry.Preview(PreviewRequest{
		SkillID:    "cabinet.discoveries.send_to_wishlist",
		Parameters: map[string]any{"result_id": "result-1"},
	})
	if err != nil {
		t.Fatalf("preview discovery handoff missing provider: %v", err)
	}
	if missingProvider.Allowed || missingProvider.Blocker != "discoveries_provider_required" {
		t.Fatalf("expected provider blocker without mutation, got %+v", missingProvider)
	}

	missingResult, err := registry.Preview(PreviewRequest{
		SkillID:    "cabinet.discoveries.create_purchase",
		Parameters: map[string]any{"provider_id": "ebay"},
	})
	if err != nil {
		t.Fatalf("preview discovery purchase missing result: %v", err)
	}
	if missingResult.Allowed || missingResult.Blocker != "discoveries_result_required" {
		t.Fatalf("expected discovery result blocker without mutation, got %+v", missingResult)
	}
}

func TestIntegrationsAndSettingsSkillsExposeSetupPreviewBoundaries(t *testing.T) {
	t.Parallel()

	registry := NewRegistry(nil)

	searchProviders, ok := registry.Resolve("cabinet.integrations.search_providers")
	if !ok {
		t.Fatalf("expected integrations search providers skill")
	}
	if searchProviders.SafetyLevel != SafetyReadOnly || !searchProviders.Executable || searchProviders.Permissions.LocalWrite {
		t.Fatalf("search providers should be executable read-only metadata, got %+v", searchProviders)
	}
	if !slices.Contains(searchProviders.RequiredProviders, "provider-registry") || !slices.Contains(searchProviders.IntegrationWorkflows, "integrations.provider.search") {
		t.Fatalf("expected provider registry workflow bindings, got providers=%+v workflows=%+v", searchProviders.RequiredProviders, searchProviders.IntegrationWorkflows)
	}

	configure, ok := registry.Resolve("cabinet.integrations.configure_provider")
	if !ok {
		t.Fatalf("expected configure provider skill")
	}
	if configure.SafetyLevel != SafetyConfirmRequired || !configure.Permissions.RequiresConfirm || !configure.Permissions.ExternalWrite || !configure.Permissions.SecretAccess {
		t.Fatalf("configure provider should declare confirm-required external/secret safety, got %+v", configure)
	}

	configurePreview, err := registry.Preview(PreviewRequest{
		SkillID: "cabinet.integrations.configure_provider",
		Parameters: map[string]any{
			"provider_id": "ebay",
			"api_key":     "secret-value",
		},
	})
	if err != nil {
		t.Fatalf("preview configure provider: %v", err)
	}
	if configurePreview.Allowed || configurePreview.MutationApplied || configurePreview.Blocker != "confirmation_required" {
		t.Fatalf("expected non-mutating confirmation preview, got %+v", configurePreview)
	}
	if _, leaked := configurePreview.Target["api_key"]; leaked {
		t.Fatalf("preview target must not echo secrets, got %+v", configurePreview.Target)
	}

	missingProvider, err := registry.Preview(PreviewRequest{SkillID: "cabinet.integrations.test_connection"})
	if err != nil {
		t.Fatalf("preview missing provider test connection: %v", err)
	}
	if missingProvider.Allowed || missingProvider.Blocker != "integrations_provider_required" {
		t.Fatalf("expected provider selection blocker, got %+v", missingProvider)
	}

	showStorage, ok := registry.Resolve("cabinet.storage.show_status")
	if !ok {
		t.Fatalf("expected storage status skill")
	}
	if showStorage.SafetyLevel != SafetyReadOnly || !showStorage.Executable || showStorage.Permissions.LocalWrite {
		t.Fatalf("storage status should be executable read-only metadata, got %+v", showStorage)
	}

	restoreBackup, ok := registry.Resolve("cabinet.data.restore_backup")
	if !ok {
		t.Fatalf("expected restore backup skill")
	}
	if restoreBackup.SafetyLevel != SafetyDestructive || !restoreBackup.Permissions.Destructive || !restoreBackup.Permissions.RequiresConfirm {
		t.Fatalf("restore backup should be destructive and confirmation-gated, got %+v", restoreBackup)
	}
	restoreMissingTarget, err := registry.Preview(PreviewRequest{SkillID: "cabinet.data.restore_backup"})
	if err != nil {
		t.Fatalf("preview restore backup: %v", err)
	}
	if restoreMissingTarget.Allowed || restoreMissingTarget.Blocker != "data_backup_target_required" {
		t.Fatalf("expected restore target blocker without mutation, got %+v", restoreMissingTarget)
	}
}

func TestInboxAndUsersAdminSkillsExposeSafetyAndExecutionBoundaries(t *testing.T) {
	t.Parallel()

	registry := NewRegistry(nil)

	inboxSearch, ok := registry.Resolve("cabinet.inbox.search_notifications")
	if !ok {
		t.Fatalf("expected inbox search skill")
	}
	if inboxSearch.SafetyLevel != SafetyReadOnly || inboxSearch.Permissions.LocalWrite || !inboxSearch.Executable {
		t.Fatalf("inbox search should be executable read-only metadata, got %+v", inboxSearch)
	}

	inboxMutation, ok := registry.Resolve("cabinet.inbox.mark_handled")
	if !ok {
		t.Fatalf("expected inbox mark handled skill")
	}
	if inboxMutation.SafetyLevel != SafetyConfirmRequired || !inboxMutation.Permissions.RequiresConfirm || !inboxMutation.Permissions.LocalWrite {
		t.Fatalf("inbox mutation should declare confirm-required local write safety, got %+v", inboxMutation)
	}
	if inboxMutation.Status != StatusAvailable || !inboxMutation.Executable || inboxMutation.NextAction != "" {
		t.Fatalf("inbox mutation should be executable after handler binding, got %+v", inboxMutation)
	}

	userSearch, ok := registry.Resolve("cabinet.users.search")
	if !ok {
		t.Fatalf("expected users search skill")
	}
	if userSearch.SafetyLevel != SafetyReadOnly || !slices.Contains(userSearch.RequiredContext, "admin_session") || !userSearch.Executable {
		t.Fatalf("users search should be executable read-only admin metadata, got %+v", userSearch)
	}

	removeUser, ok := registry.Resolve("cabinet.users.remove_user")
	if !ok {
		t.Fatalf("expected remove user skill")
	}
	if removeUser.SafetyLevel != SafetyDestructive || !removeUser.Permissions.Destructive || !removeUser.Permissions.RequiresConfirm {
		t.Fatalf("remove user should declare destructive confirmation safety, got %+v", removeUser)
	}
	if removeUser.Status != StatusAvailable || !removeUser.Executable || removeUser.NextAction != "" {
		t.Fatalf("remove user should be executable after protected admin handlers are bound, got %+v", removeUser)
	}
}

func TestSkillPreviewBlocksUnboundInboxAndUsersMutations(t *testing.T) {
	t.Parallel()

	registry := NewRegistry(nil)

	inboxMissingTarget, err := registry.Preview(PreviewRequest{SkillID: "cabinet.inbox.mark_handled"})
	if err != nil {
		t.Fatalf("preview inbox skill: %v", err)
	}
	if inboxMissingTarget.Allowed || inboxMissingTarget.MutationApplied || inboxMissingTarget.Blocker != "inbox_notification_target_required" {
		t.Fatalf("expected missing Inbox target blocker without mutation, got %+v", inboxMissingTarget)
	}

	inboxTargeted, err := registry.Preview(PreviewRequest{
		SkillID:    "cabinet.inbox.archive_or_hide",
		Parameters: map[string]any{"notification_id": "notice-1", "action": "hide"},
	})
	if err != nil {
		t.Fatalf("preview inbox targeted skill: %v", err)
	}
	if inboxTargeted.Allowed || !inboxTargeted.ConfirmationRequired || inboxTargeted.Blocker != "confirmation_required" {
		t.Fatalf("expected Inbox confirmation blocker, got %+v", inboxTargeted)
	}

	usersMissingTarget, err := registry.Preview(PreviewRequest{SkillID: "cabinet.users.update_role"})
	if err != nil {
		t.Fatalf("preview users skill: %v", err)
	}
	if usersMissingTarget.Allowed || usersMissingTarget.Blocker != "users_admin_target_required" {
		t.Fatalf("expected users admin missing target blocker, got %+v", usersMissingTarget)
	}

	protectedRemove, err := registry.Preview(PreviewRequest{
		SkillID: "cabinet.users.remove_user",
		Parameters: map[string]any{
			"target_user":         "owner@example.test",
			"target_role_current": "owner",
		},
	})
	if err != nil {
		t.Fatalf("preview remove user skill: %v", err)
	}
	if protectedRemove.Allowed || protectedRemove.MutationApplied || protectedRemove.Blocker != "users_admin_protected_owner_remove_blocked" {
		t.Fatalf("expected protected owner removal blocker without mutation, got %+v", protectedRemove)
	}
}

func TestImportedSkillCannotOverrideBuiltInSkillID(t *testing.T) {
	t.Parallel()

	registry := NewRegistry([]Skill{{
		ID:          "cabinet.navigate.open_surface",
		DisplayName: "Unsafe override",
		Enabled:     true,
	}})

	skills := registry.List()
	var count int
	for _, skill := range skills {
		if skill.ID == "cabinet.navigate.open_surface" {
			count++
			if skill.Source != SourceBuiltIn || !skill.BuiltIn || skill.Removable {
				t.Fatalf("override changed built-in metadata: %+v", skill)
			}
		}
	}
	if count != 1 {
		t.Fatalf("expected exactly one effective built-in navigate skill, got %d", count)
	}
	if err := registry.ValidateImportedSkill(Skill{ID: "cabinet.navigate.open_surface"}); err == nil {
		t.Fatalf("expected duplicate built-in id validation error")
	}
}

func TestSkillStatusAndSafetyDerivation(t *testing.T) {
	t.Parallel()

	registry := NewRegistry(nil)
	guided, ok := registry.Resolve("cabinet.guided.inventory.update_item")
	if !ok {
		t.Fatalf("expected guided inventory skill")
	}
	if guided.Status != StatusRequiresImplementation || guided.Executable {
		t.Fatalf("guided inventory update must not be executable before #1513, got %+v", guided)
	}
	if !slices.Contains(guided.GuidedWorkflows, "inventory.item.update") {
		t.Fatalf("expected guided workflow binding, got %+v", guided.GuidedWorkflows)
	}
	if guided.NextAction == "" {
		t.Fatalf("expected blocked skill to expose next action")
	}

	readOnly, ok := registry.Resolve("cabinet.chat.action_timeline.view")
	if !ok {
		t.Fatalf("expected action timeline skill")
	}
	if readOnly.SafetyLevel != SafetyReadOnly || readOnly.Permissions.LocalWrite || readOnly.Permissions.RequiresConfirm {
		t.Fatalf("expected read-only no-write timeline skill, got %+v", readOnly)
	}
}

func TestSkillInvocationReportsMissingContextAndPermissions(t *testing.T) {
	t.Parallel()

	registry := NewRegistry(nil)

	missing, err := registry.ReviewRequirements(PreviewRequest{SkillID: "cabinet.inventory.attach_media"})
	if err != nil {
		t.Fatalf("review inventory attach requirements: %v", err)
	}
	for _, context := range []string{"profile", "workspace", "thread", "selected_item", "selected_media"} {
		if !slices.Contains(missing.MissingContext, context) {
			t.Fatalf("expected missing context %q, got %+v", context, missing.MissingContext)
		}
	}
	if missing.Allowed || missing.Blocker != "missing_context" || !missing.ConfirmationRequired {
		t.Fatalf("expected missing-context blocker before inventory mutation, got %+v", missing)
	}
	if !missing.Permissions.LocalWrite || !missing.Permissions.RequiresConfirm || missing.Permissions.SecretAccess {
		t.Fatalf("expected explicit non-secret local write permission boundary, got %+v", missing.Permissions)
	}
	if missing.AuditBehavior == "" {
		t.Fatalf("expected audit behavior to be exposed")
	}

	ready, err := registry.ReviewRequirements(PreviewRequest{
		SkillID:        "cabinet.inventory.attach_media",
		ProfileID:      "profile-a",
		SourceThreadID: "thread-a",
		Parameters: map[string]any{
			"workspace_id": "workspace-a",
			"item_id":      "item-1",
			"media_id":     "media-1",
		},
	})
	if err != nil {
		t.Fatalf("review ready inventory attach requirements: %v", err)
	}
	if len(ready.MissingContext) != 0 || ready.Blocker != "confirmation_required" || ready.Allowed {
		t.Fatalf("expected complete context to advance to confirmation requirement, got %+v", ready)
	}

	readOnly, err := registry.ReviewRequirements(PreviewRequest{
		SkillID:   "cabinet.storage.show_status",
		ProfileID: "profile-a",
		Parameters: map[string]any{
			"workspace_id": "workspace-a",
			"storage":      "local",
		},
	})
	if err != nil {
		t.Fatalf("review storage status requirements: %v", err)
	}
	if !readOnly.Allowed || readOnly.Blocker != "" || readOnly.ConfirmationRequired || readOnly.Permissions.LocalWrite {
		t.Fatalf("expected read-only storage skill to be allowed with context, got %+v", readOnly)
	}
}

func TestProfileAgentPermissionPolicyGuardsSkillAuthority(t *testing.T) {
	t.Parallel()

	registry := NewRegistry(nil)

	readOnlySearch, err := registry.ReviewAuthority(PreviewRequest{
		SkillID:   "cabinet.inventory.search_items",
		ProfileID: "profile-a",
		Parameters: map[string]any{
			"workspace_id": "workspace-a",
		},
	}, AgentAuthorityPolicy{
		ProfileID:  "profile-a",
		Mode:       AgentAuthorityReadOnly,
		EntryPoint: "chat",
	})
	if err != nil {
		t.Fatalf("review read-only authority: %v", err)
	}
	if !readOnlySearch.Allowed || !readOnlySearch.PreviewAllowed || !readOnlySearch.ApplyAllowed || readOnlySearch.Blocker != "" {
		t.Fatalf("expected read-only skill to remain allowed in read-only mode, got %+v", readOnlySearch)
	}

	blockedMutation, err := registry.ReviewAuthority(PreviewRequest{
		SkillID:        "cabinet.inventory.create_item",
		ProfileID:      "profile-a",
		SourceThreadID: "thread-a",
		Confirm:        true,
		Parameters: map[string]any{
			"workspace_id": "workspace-a",
			"part_number":  "AFX-1",
			"title":        "Agent blocked item",
		},
	}, AgentAuthorityPolicy{
		ProfileID:  "profile-a",
		Mode:       AgentAuthorityReadOnly,
		EntryPoint: "direct-api",
	})
	if err != nil {
		t.Fatalf("review blocked mutation authority: %v", err)
	}
	if blockedMutation.Allowed || blockedMutation.PreviewAllowed || blockedMutation.ApplyAllowed ||
		blockedMutation.Blocker != "agent_authority_read_only" ||
		blockedMutation.Decision != "blocked" {
		t.Fatalf("expected read-only profile to block crafted mutation, got %+v", blockedMutation)
	}

	defaultMutationPreview, err := registry.ReviewAuthority(PreviewRequest{
		SkillID:        "cabinet.inventory.create_item",
		ProfileID:      "profile-a",
		SourceThreadID: "thread-a",
		Parameters: map[string]any{
			"workspace_id": "workspace-a",
			"part_number":  "AFX-2",
			"title":        "Agent preview item",
		},
	}, AgentAuthorityPolicy{
		ProfileID:  "profile-a",
		Mode:       AgentAuthorityAskBeforeLocalChanges,
		EntryPoint: "assistant-side-panel",
	})
	if err != nil {
		t.Fatalf("review default mutation authority: %v", err)
	}
	if defaultMutationPreview.Allowed || !defaultMutationPreview.PreviewAllowed || defaultMutationPreview.ApplyAllowed ||
		!defaultMutationPreview.ConfirmationRequired || defaultMutationPreview.Blocker != "confirmation_required" {
		t.Fatalf("expected default mode to permit preview but require confirmation before apply, got %+v", defaultMutationPreview)
	}

	externalWriteBlocked, err := registry.ReviewAuthority(PreviewRequest{
		SkillID:        "cabinet.market_watch.run_watch",
		ProfileID:      "profile-a",
		SourceThreadID: "thread-a",
		Confirm:        true,
		Parameters: map[string]any{
			"workspace_id": "workspace-a",
			"provider_id":  "ebay",
			"watch_id":     "watch-a",
		},
	}, AgentAuthorityPolicy{
		ProfileID:  "profile-a",
		Mode:       AgentAuthorityAskBeforeLocalChanges,
		EntryPoint: "telegram",
	})
	if err != nil {
		t.Fatalf("review external write authority: %v", err)
	}
	if externalWriteBlocked.Allowed || externalWriteBlocked.ApplyAllowed || externalWriteBlocked.Blocker != "agent_authority_external_write_not_approved" {
		t.Fatalf("expected external write to require separate approval, got %+v", externalWriteBlocked)
	}

	externalWriteApproved, err := registry.ReviewAuthority(PreviewRequest{
		SkillID:        "cabinet.market_watch.run_watch",
		ProfileID:      "profile-a",
		SourceThreadID: "thread-a",
		Confirm:        true,
		Parameters: map[string]any{
			"workspace_id": "workspace-a",
			"provider_id":  "ebay",
			"watch_id":     "watch-a",
		},
	}, AgentAuthorityPolicy{
		ProfileID:              "profile-a",
		Mode:                   AgentAuthorityApprovedExternalActions,
		ExternalWriteApproved:  true,
		EntryPoint:             "telegram",
		StrongConfirmationText: "Run saved watch watch-a with eBay",
	})
	if err != nil {
		t.Fatalf("review approved external write authority: %v", err)
	}
	if !externalWriteApproved.Allowed || !externalWriteApproved.ApplyAllowed || !externalWriteApproved.ConfirmationRequired || externalWriteApproved.Blocker != "" {
		t.Fatalf("expected approved external write to pass only with confirmation, got %+v", externalWriteApproved)
	}

	profileMismatch, err := registry.ReviewAuthority(PreviewRequest{
		SkillID:   "cabinet.storage.show_status",
		ProfileID: "profile-b",
		Parameters: map[string]any{
			"workspace_id": "workspace-b",
			"storage":      "local",
		},
	}, AgentAuthorityPolicy{
		ProfileID:  "profile-a",
		Mode:       AgentAuthorityAskBeforeLocalChanges,
		EntryPoint: "mcp",
	})
	if err != nil {
		t.Fatalf("review profile mismatch authority: %v", err)
	}
	if profileMismatch.Allowed || profileMismatch.Blocker != "agent_authority_profile_mismatch" {
		t.Fatalf("expected profile mismatch blocker, got %+v", profileMismatch)
	}
}

func TestProfileScopedInstalledSkillEnableDisableAndInvalidState(t *testing.T) {
	t.Parallel()

	imported := []Skill{
		{
			ID:          "local.archive.disabled_writer",
			Version:     "0.1.0",
			DisplayName: "Disabled writer",
			Status:      StatusAvailable,
			SafetyLevel: SafetyConfirmRequired,
			Enabled:     true,
		},
		{
			ID:          "local.archive.invalid_reader",
			Version:     "0.1.0",
			DisplayName: "Invalid reader",
			Status:      StatusAvailable,
			SafetyLevel: SafetyReadOnly,
			Enabled:     true,
		},
	}

	registry := NewProfileRegistry("profile-a", imported, []InstalledSkillState{
		{ProfileID: "profile-a", SkillID: "local.archive.disabled_writer", Enabled: false},
		{
			ProfileID:          "profile-a",
			SkillID:            "local.archive.invalid_reader",
			Enabled:            true,
			Status:             StatusInvalid,
			ValidationErrors:   []string{"missing capability: local.reader"},
			ValidationWarnings: []string{"schema version is deprecated"},
		},
		{ProfileID: "profile-b", SkillID: "local.archive.disabled_writer", Enabled: true},
		{ProfileID: "profile-a", SkillID: "cabinet.navigate.open_surface", Enabled: false, Status: StatusDisabled},
	})

	disabled, ok := registry.Resolve("local.archive.disabled_writer")
	if !ok {
		t.Fatalf("expected disabled imported skill")
	}
	if disabled.Status != StatusDisabled || disabled.Executable || !disabled.Removable || disabled.BuiltIn {
		t.Fatalf("disabled imported skill should stay visible and non-executable, got %+v", disabled)
	}
	if disabled.NextAction == "" {
		t.Fatalf("expected disabled imported skill guidance")
	}

	invalid, ok := registry.Resolve("local.archive.invalid_reader")
	if !ok {
		t.Fatalf("expected invalid imported skill")
	}
	if invalid.Status != StatusInvalid || invalid.Executable {
		t.Fatalf("invalid imported skill should not be executable, got %+v", invalid)
	}
	if !slices.Contains(invalid.ValidationErrors, "missing capability: local.reader") ||
		!slices.Contains(invalid.ValidationWarnings, "schema version is deprecated") {
		t.Fatalf("expected profile-scoped validation details on invalid skill, got %+v", invalid)
	}

	builtIn, ok := registry.Resolve("cabinet.navigate.open_surface")
	if !ok {
		t.Fatalf("expected built-in navigation skill")
	}
	if builtIn.Status == StatusDisabled || !builtIn.Enabled || !builtIn.BuiltIn || builtIn.Removable {
		t.Fatalf("profile installed state must not disable or remove built-ins, got %+v", builtIn)
	}

	otherProfile := NewProfileRegistry("profile-b", imported, []InstalledSkillState{
		{ProfileID: "profile-a", SkillID: "local.archive.disabled_writer", Enabled: false},
		{ProfileID: "profile-b", SkillID: "local.archive.disabled_writer", Enabled: true},
	})
	enabledForOtherProfile, ok := otherProfile.Resolve("local.archive.disabled_writer")
	if !ok {
		t.Fatalf("expected imported writer for second profile")
	}
	if enabledForOtherProfile.Status != StatusAvailable || !enabledForOtherProfile.Executable {
		t.Fatalf("expected installed state to remain profile scoped, got %+v", enabledForOtherProfile)
	}
}

func TestInstalledSkillStorePersistsProfileScopedState(t *testing.T) {
	t.Parallel()

	store := NewInstalledSkillStore([]InstalledSkillState{
		{ProfileID: " profile-a ", SkillID: " local.archive.writer ", Enabled: true, Status: StatusAvailable},
		{
			ProfileID:          "profile-a",
			SkillID:            "local.archive.invalid",
			Enabled:            true,
			Status:             StatusInvalid,
			ValidationErrors:   []string{"missing capability: local.writer"},
			ValidationWarnings: []string{"deprecated schema"},
		},
		{ProfileID: "profile-b", SkillID: "local.archive.writer", Enabled: false, Status: StatusDisabled},
	})

	profileA := store.List("profile-a")
	if len(profileA) != 2 {
		t.Fatalf("expected two profile-a installed states, got %+v", profileA)
	}
	if profileA[0].SkillID != "local.archive.invalid" || profileA[1].SkillID != "local.archive.writer" {
		t.Fatalf("expected deterministic skill ordering, got %+v", profileA)
	}
	profileA[0].ValidationErrors[0] = "mutated"
	again := store.List("profile-a")
	if again[0].ValidationErrors[0] != "missing capability: local.writer" {
		t.Fatalf("installed skill store must protect validation slices from caller mutation, got %+v", again[0])
	}
	allStates := store.ListAll()
	if len(allStates) != 3 || allStates[0].ProfileID != "profile-a" || allStates[0].SkillID != "local.archive.invalid" || allStates[2].ProfileID != "profile-b" {
		t.Fatalf("expected deterministic full installed state snapshot, got %+v", allStates)
	}
	allStates[0].ValidationErrors[0] = "mutated again"
	if store.List("profile-a")[0].ValidationErrors[0] != "missing capability: local.writer" {
		t.Fatalf("installed skill full snapshot must protect validation slices")
	}

	disabled, err := store.SetEnabled("profile-a", "local.archive.writer", false)
	if err != nil {
		t.Fatalf("disable installed skill: %v", err)
	}
	if disabled.Status != StatusDisabled || disabled.Enabled {
		t.Fatalf("expected disabled persisted state, got %+v", disabled)
	}

	enabled, err := store.SetEnabled("profile-a", "local.archive.writer", true)
	if err != nil {
		t.Fatalf("enable installed skill: %v", err)
	}
	if enabled.Status != StatusAvailable || !enabled.Enabled {
		t.Fatalf("expected re-enabled available state, got %+v", enabled)
	}

	imported := []Skill{
		{ID: "local.archive.writer", Version: "0.1.0", DisplayName: "Writer", Status: StatusAvailable, SafetyLevel: SafetyConfirmRequired, Enabled: true},
		{ID: "local.archive.invalid", Version: "0.1.0", DisplayName: "Invalid", Status: StatusAvailable, SafetyLevel: SafetyReadOnly, Enabled: true},
	}
	registry := NewProfileRegistry("profile-a", imported, store.List("profile-a"))
	invalid, ok := registry.Resolve("local.archive.invalid")
	if !ok {
		t.Fatalf("expected invalid imported skill")
	}
	if invalid.Status != StatusInvalid || invalid.Executable || !slices.Contains(invalid.ValidationWarnings, "deprecated schema") {
		t.Fatalf("expected persisted invalid state to make imported skill safe, got %+v", invalid)
	}

	if !store.Remove("profile-a", "local.archive.invalid") {
		t.Fatalf("expected installed state removal to report true")
	}
	if len(store.List("profile-a")) != 1 || len(store.List("profile-b")) != 1 {
		t.Fatalf("expected removal to stay profile scoped, profile-a=%+v profile-b=%+v", store.List("profile-a"), store.List("profile-b"))
	}
	if _, err := store.Save(InstalledSkillState{ProfileID: "profile-a"}); err == nil {
		t.Fatalf("expected required skill id error")
	}
}

func containsSkill(skills []Skill, id string) bool {
	for _, skill := range skills {
		if skill.ID == id {
			return true
		}
	}
	return false
}
