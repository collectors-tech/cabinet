package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/collectors-tech/cabinet/internal/chat"
	"github.com/collectors-tech/cabinet/internal/collection"
	"github.com/collectors-tech/cabinet/internal/commerce"
	"github.com/collectors-tech/cabinet/internal/dashboard"
	"github.com/collectors-tech/cabinet/internal/discovery"
	"github.com/collectors-tech/cabinet/internal/ebay"
	"github.com/collectors-tech/cabinet/internal/media"
	"github.com/collectors-tech/cabinet/internal/profile"
	"github.com/collectors-tech/cabinet/internal/scanner"
	"github.com/collectors-tech/cabinet/internal/wishlist"
	"github.com/google/uuid"
)

func applyAgentSkill(ctx context.Context, conn *sql.DB, chatSvc *chat.Service, skillID, profileID string, params map[string]any) (map[string]any, string, error) {
	skillID = strings.TrimSpace(skillID)
	profileID = strings.TrimSpace(profileID)
	if params == nil {
		params = map[string]any{}
	}
	switch skillID {
	case "cabinet.inbox.mark_handled":
		return applyAgentInboxSkill(ctx, chatSvc, profileID, params, "read")
	case "cabinet.inbox.archive_or_hide":
		return applyAgentInboxSkill(ctx, chatSvc, profileID, params, "archived")
	case "cabinet.inventory.search_items",
		"cabinet.inventory.create_item",
		"cabinet.inventory.update_item",
		"cabinet.inventory.attach_media",
		"cabinet.inventory.assign_to_collection":
		return applyAgentInventorySkill(ctx, conn, skillID, profileID, params)
	case "cabinet.integrations.search_providers",
		"cabinet.integrations.configure_provider",
		"cabinet.integrations.test_connection",
		"cabinet.integrations.repair_provider",
		"cabinet.integrations.disable_provider",
		"cabinet.integrations.explain_required_setup":
		return applyAgentIntegrationSkill(ctx, conn, skillID, profileID, params)
	case "cabinet.market_watch.search_watches",
		"cabinet.market_watch.create_saved_watch",
		"cabinet.market_watch.update_saved_watch",
		"cabinet.market_watch.run_watch",
		"cabinet.market_watch.review_results",
		"cabinet.market_watch.dismiss_result",
		"cabinet.market_watch.handoff_result":
		return applyAgentMarketWatchSkill(ctx, conn, skillID, profileID, params)
	case "cabinet.purchases.search_orders",
		"cabinet.purchases.create_order",
		"cabinet.purchases.add_line_item",
		"cabinet.purchases.receive_order",
		"cabinet.purchases.receive_line_item",
		"cabinet.purchases.reconcile_item",
		"cabinet.purchases.review_purchase":
		return applyAgentPurchasesSkill(ctx, conn, skillID, profileID, params)
	case "cabinet.wishlist.search_entries",
		"cabinet.wishlist.create_entry",
		"cabinet.wishlist.update_entry",
		"cabinet.wishlist.mark_purchased",
		"cabinet.wishlist.soft_delete_entry",
		"cabinet.wishlist.restore_entry":
		return applyAgentWishlistSkill(ctx, conn, skillID, profileID, params)
	case "cabinet.collections.search",
		"cabinet.collections.create",
		"cabinet.collections.update_metadata",
		"cabinet.collections.assign_item",
		"cabinet.collection.assign_item",
		"cabinet.collections.soft_delete",
		"cabinet.collections.move_items_on_delete":
		return applyAgentCollectionsSkill(ctx, conn, skillID, profileID, params)
	case "cabinet.dashboard.summarise_activity":
		return applyAgentDashboardSkill(ctx, conn, profileID, params)
	case "cabinet.media.search",
		"cabinet.media.upload_or_import",
		"cabinet.media.attach_to_item",
		"cabinet.media.review_unlinked",
		"cabinet.media.update_notes",
		"cabinet.media.detach_from_item":
		return applyAgentMediaSkill(ctx, conn, skillID, profileID, params)
	case "cabinet.discoveries.search",
		"cabinet.discoveries.review_result",
		"cabinet.discoveries.dismiss_result",
		"cabinet.discoveries.send_to_wishlist",
		"cabinet.discoveries.create_purchase",
		"cabinet.discoveries.create_or_update_inventory_candidate":
		return applyAgentDiscoveriesSkill(ctx, conn, skillID, profileID, params)
	case "cabinet.settings.update_profile",
		"cabinet.settings.update_account",
		"cabinet.settings.update_appearance",
		"cabinet.storage.show_status",
		"cabinet.storage.configure_backup",
		"cabinet.data.import_file",
		"cabinet.data.export_bundle",
		"cabinet.data.restore_backup",
		"cabinet.maintenance.run_safe_check":
		return applyAgentSettingsDataSkill(ctx, conn, skillID, profileID, params)
	case "cabinet.users.invite_user", "cabinet.users.resend_invitation":
		email := stringMapParam(params, "target_email")
		if email == "" {
			email = stringMapParam(params, "target_user")
		}
		user, err := inviteRuntimeUser(ctx, conn, profileID, email, stringMapParam(params, "target_role"))
		if err != nil {
			return nil, agentUserBlocker(err), err
		}
		return map[string]any{"user": user}, "", nil
	case "cabinet.users.update_role":
		targetID, blocker, err := resolveAgentSkillUserTarget(ctx, conn, profileID, params)
		if err != nil {
			return nil, blocker, err
		}
		role := stringMapParam(params, "target_role")
		if strings.TrimSpace(role) == "" {
			return nil, "users_admin_target_role_required", fmt.Errorf("target_role required")
		}
		user, err := updateRuntimeUser(ctx, conn, profileID, targetID, "", "", "", "", "", "", role)
		if err != nil {
			return nil, agentUserBlocker(err), err
		}
		return map[string]any{"user": user}, "", nil
	case "cabinet.users.activate_or_deactivate":
		targetID, blocker, err := resolveAgentSkillUserTarget(ctx, conn, profileID, params)
		if err != nil {
			return nil, blocker, err
		}
		status := stringMapParam(params, "target_status")
		if strings.TrimSpace(status) == "" {
			return nil, "users_admin_target_status_required", fmt.Errorf("target_status required")
		}
		user, err := updateRuntimeUser(ctx, conn, profileID, targetID, "", "", "", "", "", status, "")
		if err != nil {
			return nil, agentUserBlocker(err), err
		}
		return map[string]any{"user": user}, "", nil
	case "cabinet.users.remove_user":
		targetID, blocker, err := resolveAgentSkillUserTarget(ctx, conn, profileID, params)
		if err != nil {
			return nil, blocker, err
		}
		if err := deleteRuntimeUser(ctx, conn, profileID, targetID); err != nil {
			return nil, agentUserBlocker(err), err
		}
		return map[string]any{"removed_user_id": targetID}, "", nil
	default:
		return nil, "skill_apply_not_supported", fmt.Errorf("skill apply not supported")
	}
}

func applyAgentDashboardSkill(ctx context.Context, conn *sql.DB, profileID string, params map[string]any) (map[string]any, string, error) {
	if conn == nil {
		return agentDashboardUnavailableResult(profileID, params, "dashboard_store_unavailable"), "", nil
	}
	summary, err := dashboard.NewService(conn).Summary(ctx, profileID)
	if err != nil {
		return agentDashboardUnavailableResult(profileID, params, "dashboard_summary_unavailable"), "", nil
	}
	recentItems, err := agentDashboardRecentItems(ctx, conn, profileID, 5)
	dependencyWarnings := []string{}
	dependencyStatus := "available"
	recentItemsUnavailable := false
	if err != nil {
		dependencyStatus = "partial"
		recentItemsUnavailable = true
		dependencyWarnings = append(dependencyWarnings, "Recent Dashboard item identifiers are unavailable; current totals and attention counts still come from the canonical Dashboard service.")
	}
	requestedWindow := firstNonEmptyString(stringMapParam(params, "window"), stringMapParam(params, "time_window"), "current")
	attentionSignals := []map[string]any{
		agentDashboardSignal("new_discoveries", "New discoveries", summary.NewDiscoveries, "/discoveries"),
		agentDashboardSignal("wishlist_hits", "Wishlist hits", summary.WishlistHits, "/wishlist"),
		agentDashboardSignal("price_drops", "Price drops", summary.PriceDrops, "/pricing"),
		agentDashboardSignal("low_stock_discoveries", "Low-stock discoveries", summary.LowStockDiscoveries, "/discoveries"),
		agentDashboardSignal("restocks", "Restocks", summary.Restocks, "/pricing"),
	}
	nothingNeedsAttention := summary.NewDiscoveries == 0 &&
		summary.WishlistHits == 0 &&
		summary.PriceDrops == 0 &&
		summary.LowStockDiscoveries == 0 &&
		summary.Restocks == 0 &&
		len(recentItems) == 0
	return map[string]any{
		"profile_id":               profileID,
		"operation":                "dashboard.activity.summary",
		"read_only":                true,
		"nothing_needs_attention":  nothingNeedsAttention,
		"empty_state":              agentDashboardEmptyState(nothingNeedsAttention),
		"collection":               agentDashboardCollectionSummary(summary.Collection),
		"attention_signals":        attentionSignals,
		"recent_items":             recentItems,
		"recent_items_unavailable": recentItemsUnavailable,
		"destination_links":        agentDashboardDestinationLinks(summary.Cards),
		"dependency_state": map[string]any{
			"status":      dependencyStatus,
			"warnings":    dependencyWarnings,
			"next_action": agentDashboardDependencyNextAction(dependencyStatus),
		},
		"time_window": map[string]any{
			"requested_window": requestedWindow,
			"evidence_backed":  false,
			"snapshot_only":    true,
			"caveat":           "Dashboard currently exposes current snapshot values; historical time-window changes require event history that is not yet available.",
		},
		"record_identifiers": map[string]any{
			"recent_items":                  recentItems,
			"recent_item_identifiers_found": len(recentItems),
		},
		"warnings":    []string{"Dashboard activity summary is read-only and reports current snapshots, not historical deltas."},
		"next_action": "Open Dashboard, Discoveries, Wishlist, Pricing, or Collections using the destination links to inspect the underlying records.",
	}, "", nil
}

func agentDashboardUnavailableResult(profileID string, params map[string]any, reason string) map[string]any {
	requestedWindow := firstNonEmptyString(stringMapParam(params, "window"), stringMapParam(params, "time_window"), "current")
	warning := "Dashboard activity data is unavailable right now; no totals or attention counts were inferred."
	return map[string]any{
		"profile_id":               profileID,
		"operation":                "dashboard.activity.summary",
		"read_only":                true,
		"nothing_needs_attention":  false,
		"empty_state":              map[string]any{"active": false},
		"collection":               map[string]any{"unavailable": true},
		"attention_signals":        []map[string]any{},
		"recent_items":             []map[string]any{},
		"recent_items_unavailable": true,
		"destination_links":        agentDashboardFallbackDestinationLinks(),
		"dependency_state": map[string]any{
			"status":      "unavailable",
			"reason":      reason,
			"warnings":    []string{warning},
			"next_action": agentDashboardDependencyNextAction("unavailable"),
		},
		"time_window": map[string]any{
			"requested_window": requestedWindow,
			"evidence_backed":  false,
			"snapshot_only":    true,
			"caveat":           "Dashboard currently exposes current snapshot values when dependencies are available; historical time-window changes require event history that is not yet available.",
		},
		"record_identifiers": map[string]any{
			"recent_items":                  []map[string]any{},
			"recent_item_identifiers_found": 0,
			"unavailable":                   true,
		},
		"warnings":    []string{warning, "Dashboard activity summary is read-only and reports current snapshots, not historical deltas."},
		"next_action": agentDashboardDependencyNextAction("unavailable"),
	}
}

func agentDashboardDependencyNextAction(status string) string {
	if status == "available" {
		return "Open Dashboard, Discoveries, Wishlist, Pricing, or Collections using the destination links to inspect the underlying records."
	}
	return "Open Dashboard or run a maintenance safe check to verify the local data store before relying on this summary."
}

func agentDashboardEmptyState(nothingNeedsAttention bool) map[string]any {
	if !nothingNeedsAttention {
		return map[string]any{"active": false}
	}
	return map[string]any{
		"active":  true,
		"message": "Nothing needs attention in the current Dashboard snapshot.",
	}
}

func agentDashboardCollectionSummary(stats dashboard.CollectionStats) map[string]any {
	return map[string]any{
		"total_items":     stats.TotalItems,
		"total_instances": stats.TotalInstances,
		"estimated_value": stats.EstimatedValue,
	}
}

func agentDashboardSignal(id, label string, count int, route string) map[string]any {
	return map[string]any{
		"id":               id,
		"label":            label,
		"count":            count,
		"destination_link": route,
	}
}

func agentDashboardDestinationLinks(cards []dashboard.Card) []map[string]any {
	links := make([]map[string]any, 0, len(cards))
	for _, card := range cards {
		if strings.TrimSpace(card.Link) == "" {
			continue
		}
		links = append(links, map[string]any{
			"id":    agentDashboardLinkID(card.Title),
			"label": card.Title,
			"value": card.Value,
			"route": card.Link,
		})
	}
	return links
}

func agentDashboardFallbackDestinationLinks() []map[string]any {
	return []map[string]any{
		{"id": "dashboard", "label": "Dashboard", "route": "/dashboard"},
		{"id": "discoveries", "label": "Discoveries", "route": "/discoveries"},
		{"id": "wishlist", "label": "Wishlist", "route": "/wishlist"},
		{"id": "pricing", "label": "Pricing", "route": "/pricing"},
		{"id": "collections", "label": "Collections", "route": "/collections"},
	}
}

func agentDashboardLinkID(label string) string {
	id := strings.ToLower(strings.TrimSpace(label))
	id = strings.ReplaceAll(id, " ", "_")
	id = strings.ReplaceAll(id, "-", "_")
	return id
}

func agentDashboardRecentItems(ctx context.Context, conn *sql.DB, profileID string, limit int) ([]map[string]any, error) {
	if limit <= 0 {
		limit = 5
	}
	query := `
		SELECT id, title
		FROM canonical_items
		WHERE COALESCE(status, '') != 'deleted'
	`
	args := []any{}
	if profileID != "" {
		query += " AND profile_id = ?"
		args = append(args, profileID)
	}
	query += " ORDER BY datetime(created_at) DESC, id DESC LIMIT ?"
	args = append(args, limit)
	rows, err := conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []map[string]any{}
	for rows.Next() {
		var id, title string
		if err := rows.Scan(&id, &title); err != nil {
			return nil, err
		}
		items = append(items, map[string]any{
			"item_id":          id,
			"title":            title,
			"destination_link": "/collections",
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func applyAgentInventorySkill(ctx context.Context, conn *sql.DB, skillID, profileID string, params map[string]any) (map[string]any, string, error) {
	if conn == nil {
		return nil, "inventory_store_required", fmt.Errorf("database connection required")
	}
	repo := collection.NewRepository(conn)
	result := map[string]any{"profile_id": profileID}
	switch skillID {
	case "cabinet.inventory.search_items":
		items, err := agentInventorySearch(ctx, conn, profileID, stringMapParam(params, "query"))
		if err != nil {
			return nil, "inventory_search_failed", err
		}
		result["operation"] = "inventory.item.search"
		result["read_only"] = true
		result["query"] = stringMapParam(params, "query")
		result["items"] = items
		result["total"] = len(items)
		result["next_action"] = "Select an inventory item before applying update, media, or collection actions."
	case "cabinet.inventory.create_item":
		item, err := agentInventoryItemFromParams(params, collection.Item{
			Brand:    firstNonEmptyString(stringMapParam(params, "brand"), "Unknown"),
			Category: firstNonEmptyString(stringMapParam(params, "category"), "Inventory"),
			Status:   firstNonEmptyString(stringMapParam(params, "status"), "active"),
		})
		if err != nil {
			return nil, "inventory_item_context_required", err
		}
		created, err := repo.CreateItemForProfile(ctx, profileID, item)
		if err != nil {
			return nil, "inventory_item_persist_failed", err
		}
		result["operation"] = "inventory.item.create"
		result["status"] = "confirmed"
		result["item_id"] = created.ID
		result["item"] = created
		result["inventory_persisted"] = true
		result["next_action"] = "Open Inventory to verify the new item and add instances or media if needed."
	case "cabinet.inventory.update_item":
		itemID := stringMapParam(params, "item_id")
		if itemID == "" {
			return nil, "inventory_item_required", fmt.Errorf("inventory item required")
		}
		existing, err := repo.GetItemByID(ctx, itemID)
		if err != nil || strings.TrimSpace(existing.ID) == "" || !agentCollectionItemBelongsToProfile(ctx, conn, profileID, itemID) {
			return nil, "inventory_item_required", fmt.Errorf("inventory item not found")
		}
		updated, err := agentInventoryItemFromParams(params, existing)
		if err != nil {
			return nil, "inventory_item_context_required", err
		}
		updated.ID = existing.ID
		if _, err := repo.UpdateItem(ctx, updated.ID, updated); err != nil {
			return nil, "inventory_item_persist_failed", err
		}
		reloaded, err := repo.GetItemByID(ctx, itemID)
		if err != nil {
			return nil, "inventory_item_required", err
		}
		result["operation"] = "inventory.item.update"
		result["status"] = "confirmed"
		result["item_id"] = reloaded.ID
		result["item"] = reloaded
		result["inventory_persisted"] = true
		result["next_action"] = "Refresh Inventory or item detail to verify the confirmed field updates."
	case "cabinet.inventory.attach_media":
		itemID := stringMapParam(params, "item_id")
		mediaID := firstNonEmptyString(stringMapParam(params, "media_id"), stringMapParam(params, "attachment_id"))
		if itemID == "" {
			return nil, "inventory_item_required", fmt.Errorf("inventory item required")
		}
		if mediaID == "" {
			return nil, "inventory_media_required", fmt.Errorf("media required")
		}
		if !agentCollectionItemBelongsToProfile(ctx, conn, profileID, itemID) {
			return nil, "inventory_item_required", fmt.Errorf("inventory item not found")
		}
		assignment, err := media.NewService(conn, "").ApplyAssignment(ctx, profileID, mediaID, "inventory", itemID)
		if err != nil {
			return nil, "inventory_media_attach_failed", err
		}
		result["operation"] = "inventory.media.attach"
		result["status"] = "confirmed"
		result["item_id"] = itemID
		result["media_id"] = mediaID
		result["attachment_persisted"] = assignment.Applied
		result["provenance_preserved"] = true
		result["assignment"] = assignment
		result["next_action"] = "Open the Inventory item detail to verify the selected media is linked."
	case "cabinet.inventory.assign_to_collection":
		itemID := stringMapParam(params, "item_id")
		collectionName := agentCollectionName(params)
		if itemID == "" {
			return nil, "inventory_item_required", fmt.Errorf("inventory item required")
		}
		if collectionName == "" {
			return nil, "inventory_collection_required", fmt.Errorf("collection required")
		}
		if strings.EqualFold(collectionName, "Deleted") || strings.EqualFold(collectionName, "Trash") {
			return nil, "inventory_collection_invalid", fmt.Errorf("collection invalid")
		}
		item, err := repo.GetItemByID(ctx, itemID)
		if err != nil || strings.TrimSpace(item.ID) == "" || !agentCollectionItemBelongsToProfile(ctx, conn, profileID, itemID) || strings.TrimSpace(item.Status) == "deleted" {
			return nil, "inventory_item_required", fmt.Errorf("inventory item not found")
		}
		state, err := loadAgentCollectionsWorkspace(ctx, conn, profileID)
		if err != nil {
			return nil, "collections_workspace_load_failed", err
		}
		state.Collections = ensureAgentCollectionName(state.Collections, "All Items")
		state.Collections = ensureAgentCollectionName(state.Collections, collectionName)
		state.ActiveCollection = collectionName
		state.Items = upsertAgentCollectionItem(state.Items, item, collectionName)
		if err := persistAgentCollectionsWorkspace(ctx, conn, profileID, state); err != nil {
			return nil, "collections_workspace_persist_failed", err
		}
		result["operation"] = "inventory.collection.assign"
		result["status"] = "confirmed"
		result["item_id"] = itemID
		result["collection_name"] = collectionName
		result["collection_persisted"] = true
		result["workspace"] = state
		result["next_action"] = "Open Collections or Inventory to verify the item appears in the selected collection."
	}
	return result, "", nil
}

func agentInventorySearch(ctx context.Context, conn *sql.DB, profileID, query string) ([]collection.Item, error) {
	rows, err := conn.QueryContext(ctx, `
		SELECT id, brand, category, COALESCE(item_type,''), COALESCE((SELECT condition FROM instances WHERE item_id = canonical_items.id ORDER BY created_at ASC LIMIT 1), ''), part_number, title, status, priority, grading_status, grader, grade_numeric, slabbed, collector_classification, car_grade_type, packaging_grade_type, make, model, year, scale, series, description, COALESCE(notes,''), tags_json, COALESCE(source_urls_json,'[]'), created_at, created_by, updated_at, updated_by, COALESCE(deleted_at,''), COALESCE(deleted_by,'')
		FROM canonical_items
		WHERE profile_id = ? AND COALESCE(deleted_at,'') = ''
		ORDER BY updated_at DESC, created_at DESC, title ASC
	`, strings.TrimSpace(profileID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	needle := strings.ToLower(strings.TrimSpace(query))
	var items []collection.Item
	for rows.Next() {
		var item collection.Item
		var tagsJSON, sourceURLsJSON string
		if err := rows.Scan(&item.ID, &item.Brand, &item.Category, &item.ItemType, &item.Condition, &item.PartNumber, &item.Title, &item.Status, &item.Priority, &item.GradingStatus, &item.Grader, &item.GradeNumeric, &item.Slabbed, &item.CollectorClassification, &item.CarGradeType, &item.PackagingGradeType, &item.Make, &item.Model, &item.Year, &item.Scale, &item.Series, &item.Description, &item.Notes, &tagsJSON, &sourceURLsJSON, &item.CreatedAt, &item.CreatedBy, &item.UpdatedAt, &item.UpdatedBy, &item.DeletedAt, &item.DeletedBy); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(tagsJSON), &item.Tags)
		_ = json.Unmarshal([]byte(sourceURLsJSON), &item.SourceURLs)
		if needle != "" && !agentInventoryItemMatches(item, needle) {
			continue
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func agentInventoryItemMatches(item collection.Item, needle string) bool {
	haystack := strings.ToLower(strings.Join([]string{
		item.ID,
		item.Brand,
		item.Category,
		item.PartNumber,
		item.Title,
		item.Status,
		item.Priority,
		item.Description,
		item.Notes,
		item.Make,
		item.Model,
		item.Series,
	}, " "))
	return strings.Contains(haystack, needle)
}

func agentInventoryItemFromParams(params map[string]any, base collection.Item) (collection.Item, error) {
	out := base
	if partNumber := stringMapParam(params, "part_number"); partNumber != "" {
		out.PartNumber = partNumber
	}
	if title := stringMapParam(params, "title"); title != "" {
		out.Title = title
	}
	if strings.TrimSpace(out.PartNumber) == "" || strings.TrimSpace(out.Title) == "" {
		return collection.Item{}, fmt.Errorf("inventory item context required")
	}
	if brand := stringMapParam(params, "brand"); brand != "" {
		out.Brand = brand
	}
	if category := stringMapParam(params, "category"); category != "" {
		out.Category = category
	}
	if status := stringMapParam(params, "status"); status != "" {
		out.Status = status
	}
	if priority := stringMapParam(params, "priority"); priority != "" {
		out.Priority = priority
	}
	if notes := stringMapParam(params, "notes"); notes != "" {
		out.Notes = notes
	}
	if description := stringMapParam(params, "description"); description != "" {
		out.Description = description
	}
	if sourceURL := stringMapParam(params, "source_url"); sourceURL != "" && !stringSliceContainsFold(out.SourceURLs, sourceURL) {
		out.SourceURLs = append(out.SourceURLs, sourceURL)
	}
	return out, nil
}

func applyAgentIntegrationSkill(ctx context.Context, conn *sql.DB, skillID, profileID string, params map[string]any) (map[string]any, string, error) {
	providerID := stringMapParam(params, "provider_id")
	if providerID == "" {
		providerID = stringMapParam(params, "provider_name")
	}
	if skillID != "cabinet.integrations.search_providers" && providerID == "" {
		return nil, "integrations_provider_required", fmt.Errorf("provider required")
	}
	result := map[string]any{
		"provider_id":            providerID,
		"external_write_claimed": false,
		"secret_redacted":        containsSecretParameter(params),
	}
	switch skillID {
	case "cabinet.integrations.search_providers":
		result["operation"] = "integrations.provider.search"
		result["read_only"] = true
		result["providers"] = []map[string]any{{
			"id":             "provider-registry",
			"status":         "available",
			"setup_required": true,
		}}
	case "cabinet.integrations.test_connection":
		result["operation"] = "integrations.provider.test_connection"
		result["read_only"] = true
		health := agentSkillProviderHealthSnapshot(ctx, conn, providerID)
		result["provider_health"] = health
		result["connection_status"] = health["category"]
		result["next_action"] = health["next_action"]
	case "cabinet.integrations.explain_required_setup":
		result["operation"] = "integrations.provider.explain_setup"
		result["read_only"] = true
		result["setup_required"] = true
		result["next_action"] = "Open Integrations settings for provider-specific credential and permission setup."
	case "cabinet.integrations.configure_provider":
		result["operation"] = "integrations.provider.configure"
		result["status"] = "confirmed"
		result["setup_step"] = stringMapParam(params, "setup_step")
		persisted, secretPersisted, err := persistAgentProviderSettings(ctx, conn, profileID, providerID, params)
		if err != nil {
			return nil, "integrations_provider_settings_persist_failed", err
		}
		result["settings_persisted"] = persisted
		result["secret_persisted"] = secretPersisted
		result["next_action"] = "Run provider health validation from Integrations before routing live provider workflows."
	case "cabinet.integrations.repair_provider":
		result["operation"] = "integrations.provider.repair"
		result["status"] = "confirmed"
		if err := persistAgentProfileSettings(ctx, conn, profileID, map[string]string{
			"integration." + providerSettingSlug(providerID) + ".repair_status": "reviewed",
		}); err != nil {
			return nil, "integrations_provider_settings_persist_failed", err
		}
		result["settings_persisted"] = []string{"integration." + providerSettingSlug(providerID) + ".repair_status"}
		result["next_action"] = "Run a provider health check after reviewing repaired setup steps."
	case "cabinet.integrations.disable_provider":
		result["operation"] = "integrations.provider.disable"
		result["status"] = "confirmed"
		keys := providerSettingsKeys(providerID)
		if err := persistAgentProfileSettings(ctx, conn, profileID, map[string]string{
			keys.EnabledKey: "false",
		}); err != nil {
			return nil, "integrations_provider_settings_persist_failed", err
		}
		result["settings_persisted"] = []string{keys.EnabledKey}
		result["next_action"] = "Confirm provider disabled state in Integrations before routing provider-backed workflows."
	}
	return result, "", nil
}

func applyAgentMarketWatchSkill(ctx context.Context, conn *sql.DB, skillID, profileID string, params map[string]any) (map[string]any, string, error) {
	providerID := stringMapParam(params, "provider_id")
	if providerID == "" {
		providerID = stringMapParam(params, "provider_name")
	}
	if providerID == "" {
		return nil, "market_watch_provider_required", fmt.Errorf("provider required")
	}
	if conn == nil {
		return nil, "market_watch_store_required", fmt.Errorf("database connection required")
	}
	scannerSvc := scanner.NewService(conn)
	result := map[string]any{
		"provider_id":              providerID,
		"external_write_claimed":   false,
		"provenance_preserved":     false,
		"provider_health":          agentSkillProviderHealthSnapshot(ctx, conn, providerID),
		"live_provider_dispatched": false,
	}
	switch skillID {
	case "cabinet.market_watch.search_watches":
		watches, err := agentMarketWatchSearch(ctx, scannerSvc, profileID, providerID, stringMapParam(params, "query"))
		if err != nil {
			return nil, "market_watch_search_failed", err
		}
		result["operation"] = "market_watch.watch.search"
		result["read_only"] = true
		result["query"] = stringMapParam(params, "query")
		result["watches"] = watches
		result["total"] = len(watches)
		result["next_action"] = "Review saved watches and provider health before running or changing a watch."
	case "cabinet.market_watch.review_results":
		result["operation"] = "market_watch.results.review"
		result["read_only"] = true
		result["watch_id"] = stringMapParam(params, "watch_id")
		result["next_action"] = "Select a result before dismissing or handing it off."
	case "cabinet.market_watch.create_saved_watch":
		query := stringMapParam(params, "watch_query")
		if query == "" {
			query = stringMapParam(params, "query")
		}
		if query == "" {
			return nil, "market_watch_query_required", fmt.Errorf("watch query required")
		}
		created, err := scannerSvc.CreateQuerySetForProfile(ctx, profileID, agentMarketWatchQuerySet(providerID, query, params))
		if err != nil {
			return nil, "market_watch_watch_persist_failed", err
		}
		result["operation"] = "market_watch.watch.create"
		result["status"] = "confirmed"
		result["watch_id"] = created.ID
		result["watch_query"] = query
		result["watch_persisted"] = true
		result["saved_watch"] = created
		result["next_action"] = "Review provider health before running the saved watch."
	case "cabinet.market_watch.update_saved_watch", "cabinet.market_watch.run_watch":
		watchID := stringMapParam(params, "watch_id")
		if watchID == "" {
			return nil, "market_watch_watch_required", fmt.Errorf("watch required")
		}
		if skillID == "cabinet.market_watch.update_saved_watch" {
			existing, err := scannerSvc.GetQuerySetForProfile(ctx, profileID, watchID)
			if err != nil {
				return nil, "market_watch_watch_required", err
			}
			updated, err := scannerSvc.UpdateQuerySetForProfile(ctx, profileID, watchID, agentMarketWatchUpdatedQuerySet(existing, providerID, params))
			if err != nil {
				return nil, "market_watch_watch_persist_failed", err
			}
			result["operation"] = "market_watch.watch.update"
			result["status"] = "confirmed"
			result["watch_persisted"] = true
			result["saved_watch"] = updated
			result["query"] = strings.Join(updated.Keywords, " ")
		} else {
			watch, err := scannerSvc.GetQuerySetForProfile(ctx, profileID, watchID)
			if err != nil {
				return nil, "market_watch_watch_required", err
			}
			result["operation"] = "market_watch.watch.run"
			result["status"] = "confirmed_provider_ready_check"
			result["saved_watch"] = watch
			result["query"] = strings.Join(watch.Keywords, " ")
			if runResult, candidates, dispatched, err := runAgentMarketWatchProvider(ctx, conn, scannerSvc, profileID, providerID, watch); err != nil {
				return nil, "market_watch_provider_run_failed", err
			} else if dispatched {
				result["status"] = "confirmed_provider_run"
				result["live_provider_dispatched"] = true
				result["run"] = runResult
				result["candidates"] = candidates
				result["candidate_count"] = len(candidates)
				result["next_action"] = "Review the persisted Market Watch provider results before dismissing or handing off candidates."
			}
		}
		result["watch_id"] = watchID
		if _, ok := result["next_action"]; !ok {
			result["next_action"] = "Review provider-health evidence before treating any live result sync as complete."
		}
	case "cabinet.market_watch.dismiss_result", "cabinet.market_watch.handoff_result":
		resultID := stringMapParam(params, "result_id")
		if resultID == "" {
			return nil, "market_watch_result_required", fmt.Errorf("result required")
		}
		result["result_id"] = resultID
		result["provenance_preserved"] = true
		result["source_url"] = stringMapParam(params, "source_url")
		if skillID == "cabinet.market_watch.dismiss_result" {
			result["operation"] = "market_watch.result.dismiss"
			result["status"] = "confirmed_preview"
			result["next_action"] = "Confirm the result dismissal in the Market Watch review queue."
			break
		}
		destination := stringMapParam(params, "destination")
		if destination == "" {
			return nil, "market_watch_destination_required", fmt.Errorf("destination required")
		}
		result["operation"] = "market_watch.result.handoff"
		result["destination"] = destination
		applied, err := applyAgentMarketWatchHandoff(ctx, conn, profileID, providerID, resultID, destination, params)
		if err != nil {
			return nil, "market_watch_handoff_apply_failed", err
		}
		for key, value := range applied {
			result[key] = value
		}
	}
	return result, "", nil
}

func runAgentMarketWatchProvider(ctx context.Context, conn *sql.DB, scannerSvc *scanner.Service, profileID, providerID string, watch scanner.QuerySet) (scanner.RunResult, []scanner.Candidate, bool, error) {
	switch strings.ToLower(strings.TrimSpace(providerID)) {
	case "ebay":
		settings, err := profile.NewRepository(conn).GetSettings(ctx, strings.TrimSpace(profileID))
		if err != nil {
			return scanner.RunResult{}, nil, false, err
		}
		if strings.TrimSpace(settings["ebay_bearer_token"]) == "" || strings.TrimSpace(settings["ebay_marketplace"]) == "" {
			return scanner.RunResult{}, nil, false, nil
		}
		if raw := strings.TrimSpace(settings[providerSettingsKeys("ebay").ItemsPerPageKey]); raw != "" {
			value, err := strconv.Atoi(raw)
			if err != nil || value <= 0 {
				return scanner.RunResult{}, nil, false, fmt.Errorf("invalid ebay items_per_page setting")
			}
			watch.ItemsPerPage = value
			if _, err := scannerSvc.UpdateQuerySetForProfile(ctx, strings.TrimSpace(profileID), watch.ID, watch); err != nil {
				return scanner.RunResult{}, nil, false, err
			}
		}
		provider := ebay.NewProvider(ebay.ProviderConfig{
			BaseURL:     settings["ebay_base_url"],
			BearerToken: settings["ebay_bearer_token"],
			Marketplace: settings["ebay_marketplace"],
		})
		run, err := scannerSvc.RunNowForProfile(ctx, strings.TrimSpace(profileID), watch.ID, provider)
		if err != nil {
			return scanner.RunResult{}, nil, false, err
		}
		candidates, err := scannerSvc.ListCandidatesByProfile(ctx, strings.TrimSpace(profileID), watch.ID)
		if err != nil {
			return scanner.RunResult{}, nil, false, err
		}
		return run, candidates, true, nil
	default:
		return scanner.RunResult{}, nil, false, nil
	}
}

func applyAgentMarketWatchHandoff(ctx context.Context, conn *sql.DB, profileID, providerID, resultID, destination string, params map[string]any) (map[string]any, error) {
	destination = strings.ToLower(strings.TrimSpace(destination))
	itemID, createdItem, err := ensureAgentMarketWatchHandoffItem(ctx, conn, profileID, providerID, resultID, destination, params)
	if err != nil {
		return nil, err
	}
	out := map[string]any{
		"status":              "confirmed",
		"handoff_persisted":   true,
		"destination_applied": destination,
		"item_id":             itemID,
		"created_item":        createdItem,
		"next_action":         "Open the destination surface to review the persisted Market Watch handoff.",
	}
	switch destination {
	case "wishlist":
		entry, err := wishlist.NewService(conn).CreateForProfile(ctx, profileID, wishlist.Entry{
			ItemID:         itemID,
			TargetPrice:    floatMapParam(params, "target_price"),
			Priority:       firstNonEmptyString(stringMapParam(params, "priority"), "medium"),
			Notes:          agentMarketWatchHandoffNotes(providerID, resultID, params),
			HighlightHit:   true,
			Quantity:       intMapParam(params, "quantity"),
			NeededQuantity: intMapParam(params, "needed_quantity"),
		})
		if err != nil {
			return nil, err
		}
		out["wishlist_entry_id"] = entry.ID
	case "purchases":
		orderID := firstNonEmptyString(stringMapParam(params, "order_id"), resultID)
		created, arrival, err := commerce.NewService(conn).CreateLifecycleForProfile(ctx, profileID, commerce.LifecycleEntry{
			ItemID:      itemID,
			State:       "purchase",
			Source:      "market_watch",
			ExternalRef: orderID,
			Quantity:    intMapParam(params, "quantity"),
			Amount:      floatMapParam(params, "amount"),
			Currency:    stringMapParam(params, "currency"),
			Notes:       agentMarketWatchHandoffNotes(providerID, resultID, params),
		})
		if err != nil {
			return nil, err
		}
		out["order_id"] = orderID
		out["lifecycle_entry_id"] = created.ID
		out["expected_arrival_id"] = created.ExpectedArrivalID
		if arrival != nil {
			out["arrival_status"] = arrival.Status
		}
	case "inventory":
		instanceID := uuid.NewString()
		quantity := intMapParam(params, "quantity")
		if quantity <= 0 {
			quantity = 1
		}
		_, err := conn.ExecContext(ctx, `
			INSERT INTO instances(id, item_id, condition, status, quantity, storage_location, acquisition_price, acquisition_date, notes)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, instanceID, itemID, firstNonEmptyString(stringMapParam(params, "condition"), "custom"), firstNonEmptyString(stringMapParam(params, "inventory_status"), "loose"), quantity, stringMapParam(params, "storage_location"), floatMapParam(params, "amount"), stringMapParam(params, "acquisition_date"), agentMarketWatchHandoffNotes(providerID, resultID, params))
		if err != nil {
			return nil, err
		}
		out["instance_id"] = instanceID
	default:
		return nil, fmt.Errorf("unsupported handoff destination")
	}
	return out, nil
}

func ensureAgentMarketWatchHandoffItem(ctx context.Context, conn *sql.DB, profileID, providerID, resultID, destination string, params map[string]any) (string, bool, error) {
	if itemID := stringMapParam(params, "item_id"); itemID != "" {
		return itemID, false, nil
	}
	itemID := uuid.NewString()
	sourceURL := stringMapParam(params, "source_url")
	sourceURLs := "[]"
	if sourceURL != "" {
		encoded, _ := json.Marshal([]string{sourceURL})
		sourceURLs = string(encoded)
	}
	partNumber := firstNonEmptyString(stringMapParam(params, "part_number"), strings.ToUpper(providerSettingSlug(providerID)+"-"+providerSettingSlug(resultID)+"-"+providerSettingSlug(destination)))
	title := firstNonEmptyString(stringMapParam(params, "title"), stringMapParam(params, "listing_title"), "Market Watch result "+resultID)
	_, err := conn.ExecContext(ctx, `
		INSERT INTO canonical_items(id, profile_id, brand, category, part_number, title, status, notes, source_urls_json, created_by, updated_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 'agent.market_watch.handoff', 'agent.market_watch.handoff')
	`, itemID, strings.TrimSpace(profileID), firstNonEmptyString(stringMapParam(params, "brand"), providerID), firstNonEmptyString(stringMapParam(params, "category"), "Market Watch"), partNumber, title, agentMarketWatchItemStatus(destination), agentMarketWatchHandoffNotes(providerID, resultID, params), sourceURLs)
	if err != nil {
		return "", false, err
	}
	return itemID, true, nil
}

func agentMarketWatchItemStatus(destination string) string {
	if strings.EqualFold(destination, "wishlist") {
		return "wishlist"
	}
	return "active"
}

func agentMarketWatchHandoffNotes(providerID, resultID string, params map[string]any) string {
	notes := []string{
		"source=market_watch",
		"provider_id=" + strings.TrimSpace(providerID),
		"result_id=" + strings.TrimSpace(resultID),
	}
	if sourceURL := stringMapParam(params, "source_url"); sourceURL != "" {
		notes = append(notes, "source_url="+sourceURL)
	}
	if seller := stringMapParam(params, "seller"); seller != "" {
		notes = append(notes, "seller="+seller)
	}
	if tracking := stringMapParam(params, "tracking"); tracking != "" {
		notes = append(notes, "tracking="+tracking)
	}
	if note := stringMapParam(params, "notes"); note != "" {
		notes = append(notes, "notes="+note)
	}
	return strings.Join(notes, " ")
}

func applyAgentPurchasesSkill(ctx context.Context, conn *sql.DB, skillID, profileID string, params map[string]any) (map[string]any, string, error) {
	result := map[string]any{
		"profile_id":           profileID,
		"provenance_preserved": false,
	}
	switch skillID {
	case "cabinet.purchases.search_orders":
		if conn == nil {
			return nil, "purchases_store_required", fmt.Errorf("database connection required")
		}
		orders, err := commerce.NewService(conn).ListPurchaseOrdersByProfile(ctx, profileID, stringMapParam(params, "status"), stringMapParam(params, "query"), intMapParam(params, "page"), intMapParam(params, "page_size"))
		if err != nil {
			return nil, "purchases_order_search_failed", err
		}
		result["operation"] = "purchases.orders.search"
		result["read_only"] = true
		result["query"] = stringMapParam(params, "query")
		result["orders"] = orders.Orders
		result["total"] = orders.Total
		result["next_action"] = "Select an order or line item before applying purchase state."
	case "cabinet.purchases.review_purchase":
		result["operation"] = "purchases.order.review"
		result["read_only"] = true
		result["order_id"] = stringMapParam(params, "order_id")
		result["review_status"] = stringMapParam(params, "review_status")
	case "cabinet.purchases.create_order":
		if conn == nil {
			return nil, "purchases_store_required", fmt.Errorf("database connection required")
		}
		source := stringMapParam(params, "purchase_source")
		if source == "" {
			source = stringMapParam(params, "source")
		}
		if source == "" {
			return nil, "purchases_source_required", fmt.Errorf("purchase source required")
		}
		itemID, itemCreated, err := ensureAgentPurchaseOrderItem(ctx, conn, profileID, params)
		if err != nil {
			return nil, "purchases_item_required", err
		}
		orderID := firstNonEmptyString(
			stringMapParam(params, "order_id"),
			stringMapParam(params, "external_ref"),
			stringMapParam(params, "result_id"),
			"agent-order-"+uuid.NewString(),
		)
		created, arrival, err := commerce.NewService(conn).CreateLifecycleForProfile(ctx, profileID, commerce.LifecycleEntry{
			ItemID:      itemID,
			State:       "purchase",
			Source:      source,
			ExternalRef: orderID,
			Quantity:    intMapParam(params, "quantity"),
			Amount:      floatMapParam(params, "amount"),
			Currency:    stringMapParam(params, "currency"),
			Notes:       purchaseAgentSkillNotes(params),
		})
		if err != nil {
			return nil, "purchases_order_persist_failed", err
		}
		result["operation"] = "purchases.order.create"
		result["purchase_source"] = source
		result["order_id"] = orderID
		result["item_id"] = itemID
		result["created_item"] = itemCreated
		result["status"] = "confirmed"
		result["purchase_persisted"] = true
		result["provenance_preserved"] = true
		result["lifecycle_entry_id"] = created.ID
		result["expected_arrival_id"] = created.ExpectedArrivalID
		if arrival != nil {
			result["arrival_status"] = arrival.Status
		}
	case "cabinet.purchases.add_line_item":
		if conn == nil {
			return nil, "purchases_store_required", fmt.Errorf("database connection required")
		}
		orderID := stringMapParam(params, "order_id")
		if orderID == "" {
			return nil, "purchases_order_required", fmt.Errorf("order required")
		}
		itemID := stringMapParam(params, "item_id")
		if itemID == "" {
			itemID = stringMapParam(params, "line_item_id")
		}
		if itemID == "" {
			return nil, "purchases_item_required", fmt.Errorf("item required")
		}
		source := stringMapParam(params, "source")
		if source == "" {
			source = "agent_skill"
		}
		created, arrival, err := commerce.NewService(conn).CreateLifecycleForProfile(ctx, profileID, commerce.LifecycleEntry{
			ItemID:      itemID,
			State:       "purchase",
			Source:      source,
			ExternalRef: orderID,
			Quantity:    intMapParam(params, "quantity"),
			Amount:      floatMapParam(params, "amount"),
			Currency:    stringMapParam(params, "currency"),
			Notes:       purchaseAgentSkillNotes(params),
		})
		if err != nil {
			return nil, "purchases_line_item_persist_failed", err
		}
		result["operation"] = "purchases.order.add_line_item"
		result["order_id"] = orderID
		result["item_id"] = itemID
		result["source"] = source
		result["result_id"] = stringMapParam(params, "result_id")
		result["status"] = "confirmed"
		result["purchase_persisted"] = true
		result["provenance_preserved"] = true
		result["lifecycle_entry_id"] = created.ID
		result["expected_arrival_id"] = created.ExpectedArrivalID
		if arrival != nil {
			result["arrival_status"] = arrival.Status
		}
	case "cabinet.purchases.receive_order":
		if conn == nil {
			return nil, "purchases_store_required", fmt.Errorf("database connection required")
		}
		orderID := stringMapParam(params, "order_id")
		if orderID == "" {
			return nil, "purchases_order_required", fmt.Errorf("order required")
		}
		updates, err := receiveAgentPurchaseOrder(ctx, commerce.NewService(conn), profileID, orderID, params)
		if err != nil {
			return nil, "purchases_order_receive_failed", err
		}
		result["operation"] = "purchases.order.receive"
		result["order_id"] = orderID
		result["status"] = "confirmed"
		result["purchase_persisted"] = true
		result["received_count"] = len(updates)
		result["received_arrivals"] = updates
	case "cabinet.purchases.receive_line_item":
		if conn == nil {
			return nil, "purchases_store_required", fmt.Errorf("database connection required")
		}
		orderID := stringMapParam(params, "order_id")
		if orderID == "" {
			return nil, "purchases_order_required", fmt.Errorf("order required")
		}
		lineItemID := stringMapParam(params, "line_item_id")
		if lineItemID == "" {
			return nil, "purchases_line_item_required", fmt.Errorf("line item required")
		}
		arrival, err := updateAgentPurchaseLine(ctx, commerce.NewService(conn), profileID, orderID, lineItemID, "delivered", params)
		if err != nil {
			return nil, "purchases_line_item_receive_failed", err
		}
		result["operation"] = "purchases.line_item.receive"
		result["order_id"] = orderID
		result["line_item_id"] = lineItemID
		result["status"] = "confirmed"
		result["purchase_persisted"] = true
		result["expected_arrival_id"] = arrival.ID
		result["arrival_status"] = arrival.Status
		result["delivered_on"] = arrival.DeliveredOn
	case "cabinet.purchases.reconcile_item":
		if conn == nil {
			return nil, "purchases_store_required", fmt.Errorf("database connection required")
		}
		orderID := stringMapParam(params, "order_id")
		itemID := stringMapParam(params, "item_id")
		if orderID == "" {
			return nil, "purchases_order_required", fmt.Errorf("order required")
		}
		if itemID == "" {
			return nil, "purchases_item_required", fmt.Errorf("item required")
		}
		lineItemID := stringMapParam(params, "line_item_id")
		if lineItemID == "" {
			lineItemID = itemID
		}
		arrival, err := updateAgentPurchaseLine(ctx, commerce.NewService(conn), profileID, orderID, lineItemID, "reconciled", params)
		if err != nil {
			return nil, "purchases_item_reconcile_failed", err
		}
		result["operation"] = "purchases.item.reconcile"
		result["order_id"] = orderID
		result["item_id"] = itemID
		result["status"] = "confirmed"
		result["purchase_persisted"] = true
		result["reconciliation_persisted"] = true
		result["provenance_preserved"] = true
		result["expected_arrival_id"] = arrival.ID
		result["arrival_status"] = arrival.Status
		result["reconciled_instance_id"] = arrival.ReconciledInstanceID
	}
	return result, "", nil
}

func applyAgentMediaSkill(ctx context.Context, conn *sql.DB, skillID, profileID string, params map[string]any) (map[string]any, string, error) {
	if conn == nil {
		return nil, "media_store_required", fmt.Errorf("database connection required")
	}
	svc := media.NewService(conn, "")
	result := map[string]any{
		"profile_id":             profileID,
		"provenance_preserved":   false,
		"external_write_claimed": false,
	}
	switch skillID {
	case "cabinet.media.search", "cabinet.media.review_unlinked":
		filter := "all"
		if skillID == "cabinet.media.review_unlinked" {
			filter = "unlinked"
		}
		list, err := svc.ListWorkspaceAssets(ctx, profileID, filter)
		if err != nil {
			return nil, "media_search_failed", err
		}
		assets := filterAgentMediaAssets(list.Assets, stringMapParam(params, "query"))
		result["operation"] = map[bool]string{true: "media.review_unlinked", false: "media.search"}[skillID == "cabinet.media.review_unlinked"]
		result["read_only"] = true
		result["query"] = stringMapParam(params, "query")
		result["assets"] = assets
		result["total"] = len(assets)
		result["summary"] = list.Summary
		result["next_action"] = "Select media before applying attachment, metadata, or detachment changes."
	case "cabinet.media.upload_or_import":
		source := firstNonEmptyString(stringMapParam(params, "source_url"), stringMapParam(params, "file_path"))
		if source == "" {
			return nil, "media_source_required", fmt.Errorf("media source required")
		}
		assetID, err := persistAgentMediaImport(ctx, conn, profileID, source, params)
		if err != nil {
			return nil, "media_import_persist_failed", err
		}
		result["operation"] = "media.upload_or_import"
		result["status"] = "confirmed"
		result["media_id"] = assetID
		result["source_url"] = source
		result["filename"] = firstNonEmptyString(stringMapParam(params, "filename"), "agent-media-import")
		result["media_persisted"] = true
		result["provenance_preserved"] = true
		result["next_action"] = "Review imported media before linking it to inventory or wishlist records."
	case "cabinet.media.attach_to_item":
		mediaID := firstNonEmptyString(stringMapParam(params, "media_id"), stringMapParam(params, "attachment_id"))
		itemID := firstNonEmptyString(stringMapParam(params, "item_id"), stringMapParam(params, "target_item"))
		if mediaID == "" {
			return nil, "media_target_required", fmt.Errorf("media target required")
		}
		if itemID == "" {
			return nil, "media_item_required", fmt.Errorf("media item required")
		}
		assignment, err := svc.ApplyAssignment(ctx, profileID, mediaID, "inventory", itemID)
		if err != nil {
			return nil, "media_attachment_persist_failed", err
		}
		result["operation"] = "media.attach_to_item"
		result["status"] = "confirmed"
		result["media_id"] = mediaID
		result["item_id"] = itemID
		result["attachment_persisted"] = true
		result["provenance_preserved"] = true
		result["assignment"] = assignment
		result["next_action"] = "Open Media or the item detail surface to verify the persisted attachment."
	case "cabinet.media.update_notes":
		mediaID := firstNonEmptyString(stringMapParam(params, "media_id"), stringMapParam(params, "attachment_id"))
		if mediaID == "" {
			return nil, "media_target_required", fmt.Errorf("media target required")
		}
		updated, err := updateAgentMediaNotes(ctx, svc, profileID, mediaID, stringMapParam(params, "notes"))
		if err != nil {
			return nil, "media_metadata_persist_failed", err
		}
		result["operation"] = "media.update_notes"
		result["status"] = "confirmed"
		result["media_id"] = mediaID
		result["notes"] = updated.Notes
		result["metadata_persisted"] = true
		result["provenance_preserved"] = true
		result["asset"] = updated
	case "cabinet.media.detach_from_item":
		mediaID := firstNonEmptyString(stringMapParam(params, "media_id"), stringMapParam(params, "attachment_id"))
		itemID := firstNonEmptyString(stringMapParam(params, "item_id"), stringMapParam(params, "target_item"))
		if mediaID == "" {
			return nil, "media_target_required", fmt.Errorf("media target required")
		}
		if itemID == "" {
			return nil, "media_item_required", fmt.Errorf("media item required")
		}
		removed, err := detachAgentMediaFromItem(ctx, conn, profileID, mediaID, itemID)
		if err != nil {
			return nil, "media_detachment_persist_failed", err
		}
		result["operation"] = "media.detach_from_item"
		result["status"] = "confirmed"
		result["media_id"] = mediaID
		result["item_id"] = itemID
		result["detachment_persisted"] = removed
		result["provenance_preserved"] = true
		result["next_action"] = "Refresh Media or item detail to verify the attachment is no longer linked."
	}
	return result, "", nil
}

func filterAgentMediaAssets(assets []media.WorkspaceAsset, query string) []media.WorkspaceAsset {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return assets
	}
	out := make([]media.WorkspaceAsset, 0, len(assets))
	for _, asset := range assets {
		haystack := strings.ToLower(strings.Join([]string{
			asset.ID,
			asset.Title,
			asset.Filename,
			asset.Source,
			asset.Notes,
			asset.ItemID,
			asset.WishlistID,
		}, " "))
		if strings.Contains(haystack, query) {
			out = append(out, asset)
		}
	}
	return out
}

func persistAgentMediaImport(ctx context.Context, conn *sql.DB, profileID, source string, params map[string]any) (string, error) {
	threadID, err := ensureAgentMediaThread(ctx, conn, profileID)
	if err != nil {
		return "", err
	}
	assetID := uuid.NewString()
	filename := firstNonEmptyString(stringMapParam(params, "filename"), stringMapParam(params, "title"), "agent-media-import")
	mimeType := firstNonEmptyString(stringMapParam(params, "mime_type"), "application/octet-stream")
	if _, err := conn.ExecContext(ctx, `
		INSERT INTO chat_attachments(id, profile_id, thread_id, filename, mime_type, size_bytes, stored_path)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, assetID, profileID, threadID, filename, mimeType, intMapParam(params, "size_bytes"), source); err != nil {
		return "", fmt.Errorf("persist imported media attachment: %w", err)
	}
	contextJSON, err := json.Marshal(map[string]any{
		"source":            "agent.media.upload_or_import",
		"asset_id":          assetID,
		"title":             firstNonEmptyString(stringMapParam(params, "title"), filename),
		"origin":            source,
		"filename":          filename,
		"download_filename": firstNonEmptyString(stringMapParam(params, "download_filename"), filename),
		"notes":             stringMapParam(params, "notes"),
		"source_url":        source,
		"provenance":        "agent media skill confirmed import",
	})
	if err != nil {
		return "", fmt.Errorf("marshal media import metadata: %w", err)
	}
	if _, err := conn.ExecContext(ctx, `
		INSERT INTO chat_messages(id, profile_id, thread_id, role, content, attachments_json, context_json)
		VALUES (?, ?, ?, 'user', 'Media asset added from Media workspace.', '[]', ?)
	`, uuid.NewString(), profileID, threadID, string(contextJSON)); err != nil {
		return "", fmt.Errorf("persist media import metadata: %w", err)
	}
	return assetID, nil
}

func ensureAgentMediaThread(ctx context.Context, conn *sql.DB, profileID string) (string, error) {
	var threadID string
	err := conn.QueryRowContext(ctx, `
		SELECT id
		FROM chat_threads
		WHERE profile_id = ? AND title = 'Agent Media Imports'
		ORDER BY created_at ASC, id ASC
		LIMIT 1
	`, profileID).Scan(&threadID)
	if err == nil {
		return threadID, nil
	}
	if err != sql.ErrNoRows {
		return "", fmt.Errorf("find agent media thread: %w", err)
	}
	threadID = uuid.NewString()
	metadataJSON, err := json.Marshal(map[string]any{"source": "agent.media"})
	if err != nil {
		return "", fmt.Errorf("marshal agent media thread metadata: %w", err)
	}
	if _, err := conn.ExecContext(ctx, `
		INSERT INTO chat_threads(id, profile_id, title, metadata_json)
		VALUES (?, ?, 'Agent Media Imports', ?)
	`, threadID, profileID, string(metadataJSON)); err != nil {
		return "", fmt.Errorf("create agent media thread: %w", err)
	}
	return threadID, nil
}

func updateAgentMediaNotes(ctx context.Context, svc *media.Service, profileID, mediaID, notes string) (media.WorkspaceAsset, error) {
	list, err := svc.ListWorkspaceAssets(ctx, profileID, "all")
	if err != nil {
		return media.WorkspaceAsset{}, err
	}
	for _, asset := range list.Assets {
		if asset.ID != mediaID {
			continue
		}
		return svc.UpdateWorkspaceAssetMetadata(ctx, profileID, mediaID, media.WorkspaceAssetMetadataUpdate{
			Title:            asset.Title,
			Filename:         asset.Filename,
			Source:           asset.Source,
			DownloadFilename: asset.DownloadFilename,
			Notes:            notes,
		})
	}
	return media.WorkspaceAsset{}, fmt.Errorf("media asset not found")
}

func detachAgentMediaFromItem(ctx context.Context, conn *sql.DB, profileID, mediaID, itemID string) (bool, error) {
	result, err := conn.ExecContext(ctx, `
		DELETE FROM media_asset_links
		WHERE profile_id = ? AND asset_id = ? AND target_type = 'inventory' AND target_id = ?
	`, profileID, mediaID, itemID)
	if err != nil {
		return false, fmt.Errorf("detach media link: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("inspect detached media link: %w", err)
	}
	return affected > 0, nil
}

func applyAgentDiscoveriesSkill(ctx context.Context, conn *sql.DB, skillID, profileID string, params map[string]any) (map[string]any, string, error) {
	if conn == nil {
		return nil, "discoveries_store_required", fmt.Errorf("database connection required")
	}
	providerID := firstNonEmptyString(stringMapParam(params, "provider_id"), stringMapParam(params, "provider_name"))
	if providerID == "" {
		return nil, "discoveries_provider_required", fmt.Errorf("provider required")
	}
	svc := discovery.NewService(conn)
	result := map[string]any{
		"profile_id":             profileID,
		"provider_id":            providerID,
		"provenance_preserved":   false,
		"external_write_claimed": false,
	}
	switch skillID {
	case "cabinet.discoveries.search":
		items, err := agentDiscoveryItems(ctx, svc, providerID, stringMapParam(params, "query"), false)
		if err != nil {
			return nil, "discoveries_search_failed", err
		}
		result["operation"] = "discoveries.search"
		result["read_only"] = true
		result["query"] = stringMapParam(params, "query")
		result["items"] = items
		result["total"] = len(items)
		result["next_action"] = "Select a discovery result before dismissing, handing off, or creating destination state."
	case "cabinet.discoveries.review_result":
		resultID := firstNonEmptyString(stringMapParam(params, "result_id"), stringMapParam(params, "candidate_id"))
		if resultID == "" {
			return nil, "discoveries_result_required", fmt.Errorf("discovery result required")
		}
		item, err := findAgentDiscoveryItem(ctx, svc, providerID, resultID)
		if err != nil {
			return nil, "discoveries_result_required", err
		}
		result["operation"] = "discoveries.review_result"
		result["read_only"] = true
		result["result_id"] = resultID
		result["item"] = item
		result["provenance_preserved"] = true
		result["next_action"] = "Confirm a discovery action only after reviewing provider, listing, and destination provenance."
	case "cabinet.discoveries.dismiss_result",
		"cabinet.discoveries.send_to_wishlist",
		"cabinet.discoveries.create_purchase",
		"cabinet.discoveries.create_or_update_inventory_candidate":
		resultID := firstNonEmptyString(stringMapParam(params, "result_id"), stringMapParam(params, "candidate_id"))
		if resultID == "" {
			return nil, "discoveries_result_required", fmt.Errorf("discovery result required")
		}
		actionType := agentDiscoveryActionType(skillID)
		actionResult, err := svc.ApplyActionWithResult(ctx, discovery.Action{
			CandidateID: resultID,
			Type:        actionType,
			Payload:     agentDiscoveryActionPayload(params, providerID, skillID),
		})
		if err != nil {
			return nil, "discoveries_action_persist_failed", err
		}
		result["operation"] = agentDiscoveryOperation(skillID)
		result["status"] = "confirmed"
		result["result_id"] = resultID
		result["action"] = string(actionType)
		result["action_result"] = actionResult
		result["discovery_persisted"] = true
		result["provenance_preserved"] = true
		result["next_action"] = "Open Discoveries and the destination surface to verify the persisted handoff state."
	}
	return result, "", nil
}

func agentDiscoveryItems(ctx context.Context, svc *discovery.Service, providerID, query string, includeArchived bool) ([]discovery.Item, error) {
	items, err := svc.ListNotInCollection(ctx, discovery.Filter{Query: query, IncludeArchived: includeArchived})
	if err != nil {
		return nil, err
	}
	providerID = strings.ToLower(strings.TrimSpace(providerID))
	out := make([]discovery.Item, 0, len(items))
	for _, item := range items {
		if providerID != "" && strings.ToLower(strings.TrimSpace(item.SourceProvider)) != providerID {
			continue
		}
		out = append(out, item)
	}
	return out, nil
}

func findAgentDiscoveryItem(ctx context.Context, svc *discovery.Service, providerID, resultID string) (discovery.Item, error) {
	items, err := agentDiscoveryItems(ctx, svc, providerID, "", true)
	if err != nil {
		return discovery.Item{}, err
	}
	for _, item := range items {
		if item.CandidateID == resultID || item.ListingID == resultID {
			return item, nil
		}
	}
	return discovery.Item{}, fmt.Errorf("discovery result not found")
}

func agentDiscoveryActionType(skillID string) discovery.ActionType {
	switch skillID {
	case "cabinet.discoveries.dismiss_result":
		return discovery.ActionIgnore
	case "cabinet.discoveries.send_to_wishlist":
		return discovery.ActionAddWishlist
	case "cabinet.discoveries.create_purchase":
		return discovery.ActionMarkPurchased
	case "cabinet.discoveries.create_or_update_inventory_candidate":
		return discovery.ActionCreateItem
	default:
		return discovery.ActionReview
	}
}

func agentDiscoveryOperation(skillID string) string {
	switch skillID {
	case "cabinet.discoveries.dismiss_result":
		return "discoveries.dismiss_result"
	case "cabinet.discoveries.send_to_wishlist":
		return "discoveries.send_to_wishlist"
	case "cabinet.discoveries.create_purchase":
		return "discoveries.create_purchase"
	case "cabinet.discoveries.create_or_update_inventory_candidate":
		return "discoveries.create_or_update_inventory_candidate"
	default:
		return "discoveries.review_result"
	}
}

func agentDiscoveryActionPayload(params map[string]any, providerID, skillID string) map[string]any {
	payload := map[string]any{
		"source":      "agent.discoveries",
		"provider_id": providerID,
		"notes":       stringMapParam(params, "notes"),
	}
	for _, key := range []string{"decision", "destination", "quantity", "target_price", "priority", "source_url"} {
		if value, ok := params[key]; ok {
			payload[key] = value
		}
	}
	if skillID == "cabinet.discoveries.create_or_update_inventory_candidate" {
		payload["ownership_confirmed"] = true
	}
	return payload
}

func applyAgentWishlistSkill(ctx context.Context, conn *sql.DB, skillID, profileID string, params map[string]any) (map[string]any, string, error) {
	if conn == nil {
		return nil, "wishlist_store_required", fmt.Errorf("database connection required")
	}
	svc := wishlist.NewService(conn)
	result := map[string]any{
		"profile_id":                   profileID,
		"purchase_sync_provenance":     false,
		"inventory_quantity_sync_safe": false,
	}
	switch skillID {
	case "cabinet.wishlist.search_entries":
		entries, err := svc.ListByProfileDeleted(ctx, profileID, boolMapParam(params, "deleted"))
		if err != nil {
			return nil, "wishlist_search_failed", err
		}
		filtered := filterAgentWishlistEntries(entries, stringMapParam(params, "query"))
		result["operation"] = "wishlist.entry.search"
		result["read_only"] = true
		result["query"] = stringMapParam(params, "query")
		result["entries"] = filtered
		result["total"] = len(filtered)
		result["next_action"] = "Select a wishlist entry before applying updates, purchase state, delete, or restore."
	case "cabinet.wishlist.create_entry":
		itemID, createdItem, err := ensureAgentWishlistItem(ctx, conn, profileID, params)
		if err != nil {
			return nil, "wishlist_item_context_required", err
		}
		entry, err := svc.CreateForProfile(ctx, profileID, agentWishlistEntryFromParams(params, wishlist.Entry{ItemID: itemID}))
		if err != nil {
			return nil, "wishlist_entry_persist_failed", err
		}
		result["operation"] = "wishlist.entry.create"
		result["status"] = "confirmed"
		result["item_id"] = itemID
		result["created_item"] = createdItem
		result["wishlist_entry_id"] = entry.ID
		result["wishlist_entry"] = entry
		result["wishlist_persisted"] = true
		result["next_action"] = "Open Wishlist to review the persisted entry before purchase or delete actions."
	case "cabinet.wishlist.update_entry", "cabinet.wishlist.mark_purchased":
		entryID := agentWishlistEntryID(params)
		if entryID == "" {
			return nil, "wishlist_entry_required", fmt.Errorf("wishlist entry required")
		}
		existing, err := svc.GetByIDForProfile(ctx, profileID, entryID)
		if err != nil {
			return nil, "wishlist_entry_required", err
		}
		updated := agentWishlistEntryFromParams(params, existing)
		if skillID == "cabinet.wishlist.mark_purchased" {
			updated.Owned = true
			updated.Delivered = boolMapParam(params, "delivered")
		}
		if err := svc.UpdateForProfile(ctx, profileID, updated); err != nil {
			return nil, "wishlist_entry_persist_failed", err
		}
		reloaded, err := svc.GetByIDForProfile(ctx, profileID, entryID)
		if err != nil {
			return nil, "wishlist_entry_required", err
		}
		result["operation"] = agentWishlistOperation(skillID)
		result["status"] = "confirmed"
		result["item_id"] = reloaded.ItemID
		result["wishlist_entry_id"] = reloaded.ID
		result["wishlist_entry"] = reloaded
		result["wishlist_persisted"] = true
		if skillID == "cabinet.wishlist.mark_purchased" {
			result["purchase_sync_provenance"] = true
			result["inventory_quantity_sync_safe"] = true
			result["next_action"] = "Review Wishlist, Purchases, and Inventory evidence for the confirmed purchase sync."
		}
	case "cabinet.wishlist.soft_delete_entry":
		entryID := agentWishlistEntryID(params)
		if entryID == "" {
			return nil, "wishlist_entry_required", fmt.Errorf("wishlist entry required")
		}
		if err := svc.DeleteForProfile(ctx, profileID, entryID); err != nil {
			return nil, "wishlist_entry_persist_failed", err
		}
		deleted, err := svc.GetByIDForProfile(ctx, profileID, entryID)
		if err != nil {
			return nil, "wishlist_entry_required", err
		}
		result["operation"] = "wishlist.entry.soft_delete"
		result["status"] = "confirmed"
		result["wishlist_entry_id"] = entryID
		result["wishlist_entry"] = deleted
		result["wishlist_deleted"] = deleted.Deleted
		result["wishlist_persisted"] = true
		result["next_action"] = "Open deleted Wishlist entries to verify the hidden state before restoring or permanently deleting."
	case "cabinet.wishlist.restore_entry":
		entryID := agentWishlistEntryID(params)
		if entryID == "" {
			return nil, "wishlist_entry_required", fmt.Errorf("wishlist entry required")
		}
		if err := svc.RestoreForProfile(ctx, profileID, entryID); err != nil {
			return nil, "wishlist_entry_persist_failed", err
		}
		restored, err := svc.GetByIDForProfile(ctx, profileID, entryID)
		if err != nil {
			return nil, "wishlist_entry_required", err
		}
		result["operation"] = "wishlist.entry.restore"
		result["status"] = "confirmed"
		result["wishlist_entry_id"] = entryID
		result["wishlist_entry"] = restored
		result["wishlist_deleted"] = restored.Deleted
		result["wishlist_persisted"] = true
	}
	return result, "", nil
}

func filterAgentWishlistEntries(entries []wishlist.Entry, query string) []wishlist.Entry {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return entries
	}
	out := make([]wishlist.Entry, 0, len(entries))
	for _, entry := range entries {
		haystack := strings.ToLower(strings.Join([]string{
			entry.ID,
			entry.ItemID,
			entry.Priority,
			entry.Notes,
			entry.PurchaseURL,
			entry.PurchaseDate,
			entry.PurchaseCondition,
		}, " "))
		if strings.Contains(haystack, query) {
			out = append(out, entry)
		}
	}
	return out
}

func ensureAgentWishlistItem(ctx context.Context, conn *sql.DB, profileID string, params map[string]any) (string, bool, error) {
	if itemID := stringMapParam(params, "item_id"); itemID != "" {
		return itemID, false, nil
	}
	partNumber := stringMapParam(params, "part_number")
	title := stringMapParam(params, "title")
	if partNumber == "" || title == "" {
		return "", false, fmt.Errorf("wishlist item context required")
	}
	sourceURL := stringMapParam(params, "source_url")
	sourceURLs := []string{}
	if sourceURL != "" {
		sourceURLs = append(sourceURLs, sourceURL)
	}
	created, err := collection.NewRepository(conn).CreateItemForProfile(ctx, profileID, collection.Item{
		Brand:      firstNonEmptyString(stringMapParam(params, "brand"), "Unknown"),
		Category:   firstNonEmptyString(stringMapParam(params, "category"), "Wishlist"),
		PartNumber: partNumber,
		Title:      title,
		Status:     "wishlist",
		Priority:   stringMapParam(params, "priority"),
		Notes:      agentWishlistNotes(params),
		SourceURLs: sourceURLs,
	})
	if err != nil {
		return "", false, err
	}
	return created.ID, true, nil
}

func agentWishlistEntryFromParams(params map[string]any, base wishlist.Entry) wishlist.Entry {
	out := base
	if id := agentWishlistEntryID(params); id != "" {
		out.ID = id
	}
	if itemID := stringMapParam(params, "item_id"); itemID != "" {
		out.ItemID = itemID
	}
	if value, ok := params["target_price"]; ok && value != nil {
		out.TargetPrice = floatMapParam(params, "target_price")
	}
	if priority := stringMapParam(params, "priority"); priority != "" {
		out.Priority = priority
	}
	if _, ok := params["notes"]; ok {
		out.Notes = agentWishlistNotes(params)
	}
	if _, ok := params["highlight_hit"]; ok {
		out.HighlightHit = boolMapParam(params, "highlight_hit")
	}
	if _, ok := params["below_target_now"]; ok {
		out.BelowTargetNow = boolMapParam(params, "below_target_now")
	}
	if _, ok := params["owned"]; ok {
		out.Owned = boolMapParam(params, "owned")
	}
	if _, ok := params["delivered"]; ok {
		out.Delivered = boolMapParam(params, "delivered")
	}
	if value, ok := params["price_paid"]; ok && value != nil {
		out.PricePaid = floatMapParam(params, "price_paid")
	}
	if purchaseURL := stringMapParam(params, "purchase_url"); purchaseURL != "" {
		out.PurchaseURL = purchaseURL
	}
	if purchaseDate := stringMapParam(params, "purchase_date"); purchaseDate != "" {
		out.PurchaseDate = purchaseDate
	}
	if condition := stringMapParam(params, "purchase_condition"); condition != "" {
		out.PurchaseCondition = condition
	}
	if value, ok := params["quantity"]; ok && value != nil {
		out.Quantity = intMapParam(params, "quantity")
	}
	if value, ok := params["needed_quantity"]; ok && value != nil {
		out.NeededQuantity = intMapParam(params, "needed_quantity")
	}
	return out
}

func agentWishlistEntryID(params map[string]any) string {
	return firstNonEmptyString(stringMapParam(params, "wishlist_entry_id"), stringMapParam(params, "entry_id"))
}

func agentWishlistOperation(skillID string) string {
	switch skillID {
	case "cabinet.wishlist.mark_purchased":
		return "wishlist.entry.mark_purchased"
	case "cabinet.wishlist.update_entry":
		return "wishlist.entry.update"
	default:
		return "wishlist.entry.apply"
	}
}

func agentWishlistNotes(params map[string]any) string {
	notes := strings.TrimSpace(stringMapParam(params, "notes"))
	sourceURL := strings.TrimSpace(stringMapParam(params, "source_url"))
	if sourceURL == "" {
		return notes
	}
	if notes == "" {
		return "source_url=" + sourceURL
	}
	return notes + " source_url=" + sourceURL
}

const agentCollectionsWorkspaceKey = "collections.workspace.v1"

type agentCollectionsWorkspaceState struct {
	Collections      []string                        `json:"collections"`
	ActiveCollection string                          `json:"activeCollection"`
	Items            []agentCollectionsWorkspaceItem `json:"items"`
}

type agentCollectionsWorkspaceItem struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Detail         string `json:"detail"`
	CollectionName string `json:"collectionName,omitempty"`
}

func applyAgentCollectionsSkill(ctx context.Context, conn *sql.DB, skillID, profileID string, params map[string]any) (map[string]any, string, error) {
	if conn == nil {
		return nil, "collections_store_required", fmt.Errorf("database connection required")
	}
	state, err := loadAgentCollectionsWorkspace(ctx, conn, profileID)
	if err != nil {
		return nil, "collections_workspace_load_failed", err
	}
	result := map[string]any{"profile_id": profileID}
	switch skillID {
	case "cabinet.collections.search":
		filtered := filterAgentCollectionsWorkspace(state, stringMapParam(params, "query"))
		result["operation"] = "collections.search"
		result["read_only"] = true
		result["query"] = stringMapParam(params, "query")
		result["collections"] = filtered.Collections
		result["items"] = filtered.Items
		result["total"] = len(filtered.Collections)
		result["next_action"] = "Select a collection before changing metadata, assigning items, or deleting it."
	case "cabinet.collections.create":
		collectionName := agentCollectionName(params)
		if collectionName == "" {
			return nil, "collections_name_required", fmt.Errorf("collection name required")
		}
		state.Collections = ensureAgentCollectionName(state.Collections, "All Items")
		state.Collections = ensureAgentCollectionName(state.Collections, collectionName)
		state.ActiveCollection = collectionName
		if err := persistAgentCollectionsWorkspace(ctx, conn, profileID, state); err != nil {
			return nil, "collections_workspace_persist_failed", err
		}
		result["operation"] = "collections.create"
		result["status"] = "confirmed"
		result["collection_name"] = collectionName
		result["collection_persisted"] = true
		result["workspace"] = state
		result["next_action"] = "Open Collections to review the persisted workspace collection before assigning items."
	case "cabinet.collections.update_metadata":
		collectionName := agentCollectionName(params)
		if collectionName == "" {
			return nil, "collections_target_required", fmt.Errorf("collection target required")
		}
		if !agentCollectionExists(state.Collections, collectionName) {
			return nil, "collections_target_required", fmt.Errorf("collection target not found")
		}
		nextName := firstNonEmptyString(stringMapParam(params, "new_collection_name"), stringMapParam(params, "new_name"), collectionName)
		if strings.EqualFold(collectionName, "All Items") && !strings.EqualFold(nextName, "All Items") {
			return nil, "collections_all_items_protected", fmt.Errorf("All Items cannot be renamed")
		}
		state = renameAgentCollection(state, collectionName, nextName)
		state.ActiveCollection = nextName
		if err := persistAgentCollectionsWorkspace(ctx, conn, profileID, state); err != nil {
			return nil, "collections_workspace_persist_failed", err
		}
		result["operation"] = "collections.update_metadata"
		result["status"] = "confirmed"
		result["collection_name"] = nextName
		result["previous_collection_name"] = collectionName
		result["collection_persisted"] = true
		result["workspace"] = state
	case "cabinet.collections.assign_item", "cabinet.collection.assign_item":
		collectionName := agentCollectionName(params)
		itemID := stringMapParam(params, "item_id")
		if itemID == "" {
			return nil, "collections_item_required", fmt.Errorf("item required")
		}
		if collectionName == "" {
			return nil, "collections_target_required", fmt.Errorf("collection target required")
		}
		item, err := collection.NewRepository(conn).GetItemByID(ctx, itemID)
		if err != nil || strings.TrimSpace(item.ID) == "" || strings.TrimSpace(item.Status) == "deleted" || !agentCollectionItemBelongsToProfile(ctx, conn, profileID, itemID) {
			return nil, "collections_item_required", fmt.Errorf("collection assignment target not found")
		}
		state.Collections = ensureAgentCollectionName(state.Collections, "All Items")
		state.Collections = ensureAgentCollectionName(state.Collections, collectionName)
		state.ActiveCollection = collectionName
		state.Items = upsertAgentCollectionItem(state.Items, item, collectionName)
		if err := persistAgentCollectionsWorkspace(ctx, conn, profileID, state); err != nil {
			return nil, "collections_workspace_persist_failed", err
		}
		result["operation"] = "collections.item.assign"
		result["status"] = "confirmed"
		result["item_id"] = itemID
		result["collection_name"] = collectionName
		result["collection_persisted"] = true
		result["workspace"] = state
		result["next_action"] = "Open Collections to verify the item appears in the selected collection."
	case "cabinet.collections.soft_delete", "cabinet.collections.move_items_on_delete":
		collectionName := agentCollectionName(params)
		if collectionName == "" {
			return nil, "collections_target_required", fmt.Errorf("collection target required")
		}
		if strings.EqualFold(collectionName, "All Items") {
			return nil, "collections_all_items_protected", fmt.Errorf("All Items cannot be deleted")
		}
		if !agentCollectionExists(state.Collections, collectionName) {
			return nil, "collections_target_required", fmt.Errorf("collection target not found")
		}
		destination := stringMapParam(params, "destination_collection")
		assignedCount := countAgentCollectionItems(state.Items, collectionName)
		if skillID == "cabinet.collections.move_items_on_delete" && destination == "" {
			return nil, "collections_destination_required", fmt.Errorf("destination collection required")
		}
		if skillID == "cabinet.collections.soft_delete" && assignedCount > 0 && destination == "" && !boolMapParam(params, "remove_items") {
			return nil, "collections_delete_destination_required", fmt.Errorf("destination collection or remove_items required")
		}
		if destination != "" && strings.EqualFold(destination, collectionName) {
			return nil, "collections_destination_required", fmt.Errorf("destination collection must differ from source")
		}
		movedCount := 0
		if destination != "" {
			movedCount = assignedCount
		}
		removedCount := 0
		if destination == "" {
			removedCount = assignedCount
		}
		state = removeAgentCollection(state, collectionName, destination, boolMapParam(params, "remove_items"))
		if err := persistAgentCollectionsWorkspace(ctx, conn, profileID, state); err != nil {
			return nil, "collections_workspace_persist_failed", err
		}
		result["operation"] = agentCollectionsDeleteOperation(skillID)
		result["status"] = "confirmed"
		result["collection_name"] = collectionName
		result["destination_collection"] = destination
		result["moved_item_count"] = movedCount
		result["removed_item_count"] = removedCount
		result["collection_deleted"] = true
		result["collection_persisted"] = true
		result["workspace"] = state
	}
	return result, "", nil
}

func loadAgentCollectionsWorkspace(ctx context.Context, conn *sql.DB, profileID string) (agentCollectionsWorkspaceState, error) {
	state := agentCollectionsWorkspaceState{
		Collections:      []string{"All Items"},
		ActiveCollection: "All Items",
		Items:            []agentCollectionsWorkspaceItem{},
	}
	var raw string
	err := conn.QueryRowContext(ctx, `SELECT value FROM profile_settings WHERE profile_id = ? AND key = ?`, strings.TrimSpace(profileID), agentCollectionsWorkspaceKey).Scan(&raw)
	if err != nil {
		if err == sql.ErrNoRows {
			return state, nil
		}
		return state, err
	}
	if strings.TrimSpace(raw) == "" {
		return state, nil
	}
	if err := json.Unmarshal([]byte(raw), &state); err != nil {
		return agentCollectionsWorkspaceState{}, err
	}
	state.Collections = ensureAgentCollectionName(state.Collections, "All Items")
	if strings.TrimSpace(state.ActiveCollection) == "" {
		state.ActiveCollection = "All Items"
	}
	return state, nil
}

func persistAgentCollectionsWorkspace(ctx context.Context, conn *sql.DB, profileID string, state agentCollectionsWorkspaceState) error {
	state.Collections = ensureAgentCollectionName(state.Collections, "All Items")
	if strings.TrimSpace(state.ActiveCollection) == "" || !agentCollectionExists(state.Collections, state.ActiveCollection) {
		state.ActiveCollection = "All Items"
	}
	raw, err := json.Marshal(state)
	if err != nil {
		return err
	}
	_, err = conn.ExecContext(ctx, `
		INSERT INTO profile_settings(profile_id, key, value, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(profile_id, key) DO UPDATE SET value=excluded.value, updated_at=CURRENT_TIMESTAMP
	`, strings.TrimSpace(profileID), agentCollectionsWorkspaceKey, string(raw))
	return err
}

func filterAgentCollectionsWorkspace(state agentCollectionsWorkspaceState, query string) agentCollectionsWorkspaceState {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return state
	}
	out := agentCollectionsWorkspaceState{ActiveCollection: state.ActiveCollection}
	for _, name := range state.Collections {
		if strings.Contains(strings.ToLower(name), query) {
			out.Collections = append(out.Collections, name)
		}
	}
	for _, item := range state.Items {
		haystack := strings.ToLower(strings.Join([]string{item.ID, item.Name, item.Detail, item.CollectionName}, " "))
		if strings.Contains(haystack, query) {
			out.Items = append(out.Items, item)
			out.Collections = ensureAgentCollectionName(out.Collections, item.CollectionName)
		}
	}
	return out
}

func agentCollectionName(params map[string]any) string {
	return firstNonEmptyString(stringMapParam(params, "collection_name"), stringMapParam(params, "collection"))
}

func ensureAgentCollectionName(collections []string, name string) []string {
	name = strings.TrimSpace(name)
	if name == "" {
		return collections
	}
	for _, existing := range collections {
		if strings.EqualFold(strings.TrimSpace(existing), name) {
			return collections
		}
	}
	return append(collections, name)
}

func agentCollectionExists(collections []string, name string) bool {
	name = strings.TrimSpace(name)
	for _, existing := range collections {
		if strings.EqualFold(strings.TrimSpace(existing), name) {
			return true
		}
	}
	return false
}

func renameAgentCollection(state agentCollectionsWorkspaceState, oldName, newName string) agentCollectionsWorkspaceState {
	newName = strings.TrimSpace(newName)
	if newName == "" {
		newName = oldName
	}
	for i := range state.Collections {
		if strings.EqualFold(strings.TrimSpace(state.Collections[i]), oldName) {
			state.Collections[i] = newName
		}
	}
	state.Collections = ensureAgentCollectionName(state.Collections, newName)
	for i := range state.Items {
		if strings.EqualFold(strings.TrimSpace(state.Items[i].CollectionName), oldName) {
			state.Items[i].CollectionName = newName
		}
	}
	return state
}

func upsertAgentCollectionItem(items []agentCollectionsWorkspaceItem, item collection.Item, collectionName string) []agentCollectionsWorkspaceItem {
	detail := firstNonEmptyString(item.PartNumber, item.Category, "Assigned by Agent Skill")
	for i := range items {
		if strings.TrimSpace(items[i].ID) == item.ID {
			items[i].CollectionName = collectionName
			if strings.TrimSpace(items[i].Name) == "" {
				items[i].Name = item.Title
			}
			if strings.TrimSpace(items[i].Detail) == "" {
				items[i].Detail = detail
			}
			return items
		}
	}
	return append(items, agentCollectionsWorkspaceItem{
		ID:             item.ID,
		Name:           item.Title,
		Detail:         detail,
		CollectionName: collectionName,
	})
}

func agentCollectionItemBelongsToProfile(ctx context.Context, conn *sql.DB, profileID, itemID string) bool {
	var count int
	err := conn.QueryRowContext(ctx, `SELECT COUNT(1) FROM canonical_items WHERE id = ? AND profile_id = ?`, strings.TrimSpace(itemID), strings.TrimSpace(profileID)).Scan(&count)
	return err == nil && count == 1
}

func removeAgentCollection(state agentCollectionsWorkspaceState, collectionName, destination string, removeItems bool) agentCollectionsWorkspaceState {
	destination = strings.TrimSpace(destination)
	nextCollections := make([]string, 0, len(state.Collections))
	for _, existing := range state.Collections {
		if strings.EqualFold(strings.TrimSpace(existing), collectionName) {
			continue
		}
		nextCollections = ensureAgentCollectionName(nextCollections, existing)
	}
	if destination != "" {
		nextCollections = ensureAgentCollectionName(nextCollections, destination)
	}
	state.Collections = ensureAgentCollectionName(nextCollections, "All Items")
	for i := range state.Items {
		if !strings.EqualFold(strings.TrimSpace(state.Items[i].CollectionName), collectionName) {
			continue
		}
		switch {
		case destination != "":
			state.Items[i].CollectionName = destination
		case removeItems:
			state.Items[i].CollectionName = ""
		default:
			state.Items[i].CollectionName = "All Items"
		}
	}
	if strings.EqualFold(state.ActiveCollection, collectionName) {
		state.ActiveCollection = firstNonEmptyString(destination, "All Items")
	}
	return state
}

func countAgentCollectionItems(items []agentCollectionsWorkspaceItem, collectionName string) int {
	if strings.TrimSpace(collectionName) == "" {
		return 0
	}
	count := 0
	for _, item := range items {
		if strings.EqualFold(strings.TrimSpace(item.CollectionName), collectionName) {
			count++
		}
	}
	return count
}

func agentCollectionsDeleteOperation(skillID string) string {
	if skillID == "cabinet.collections.move_items_on_delete" {
		return "collections.items.move_on_delete"
	}
	return "collections.soft_delete"
}

func agentMarketWatchQuerySet(providerID, query string, params map[string]any) scanner.QuerySet {
	keywords := stringSliceMapParam(params, "keywords")
	if len(keywords) == 0 {
		keywords = strings.Fields(query)
	}
	name := firstNonEmptyString(
		stringMapParam(params, "watch_name"),
		stringMapParam(params, "name"),
		query,
	)
	return scanner.QuerySet{
		Name:          name,
		Keywords:      keywords,
		Exclusions:    stringSliceMapParam(params, "exclusions"),
		ProviderScope: []string{providerID},
		ItemsPerPage:  intMapParam(params, "items_per_page"),
		MaxPrice:      floatMapParam(params, "max_price"),
		Region:        stringMapParam(params, "region"),
		Condition:     stringMapParam(params, "condition"),
		ScheduleCron:  stringMapParam(params, "schedule_cron"),
		Enabled:       boolMapParam(params, "enabled"),
		RateLimitRPS:  intMapParam(params, "rate_limit_rps"),
		MaxRetryCount: intMapParam(params, "max_retry_count"),
	}
}

func agentMarketWatchUpdatedQuerySet(existing scanner.QuerySet, providerID string, params map[string]any) scanner.QuerySet {
	updated := existing
	if name := firstNonEmptyString(stringMapParam(params, "watch_name"), stringMapParam(params, "name")); name != "" {
		updated.Name = name
	}
	if query := firstNonEmptyString(stringMapParam(params, "watch_query"), stringMapParam(params, "query")); query != "" {
		updated.Keywords = strings.Fields(query)
	}
	if keywords := stringSliceMapParam(params, "keywords"); len(keywords) > 0 {
		updated.Keywords = keywords
	}
	if exclusions := stringSliceMapParam(params, "exclusions"); len(exclusions) > 0 {
		updated.Exclusions = exclusions
	}
	updated.ProviderScope = []string{providerID}
	if value := intMapParam(params, "items_per_page"); value > 0 {
		updated.ItemsPerPage = value
	}
	if value := floatMapParam(params, "max_price"); value > 0 {
		updated.MaxPrice = value
	}
	if value := stringMapParam(params, "region"); value != "" {
		updated.Region = value
	}
	if value := stringMapParam(params, "condition"); value != "" {
		updated.Condition = value
	}
	if value := stringMapParam(params, "schedule_cron"); value != "" {
		updated.ScheduleCron = value
	}
	if _, ok := params["enabled"]; ok {
		updated.Enabled = boolMapParam(params, "enabled")
	}
	if value := intMapParam(params, "rate_limit_rps"); value > 0 {
		updated.RateLimitRPS = value
	}
	if _, ok := params["max_retry_count"]; ok {
		value := intMapParam(params, "max_retry_count")
		updated.MaxRetryCount = value
	}
	return updated
}

func agentMarketWatchSearch(ctx context.Context, svc *scanner.Service, profileID, providerID, query string) ([]scanner.QuerySet, error) {
	watches, err := svc.ListQuerySetsByProfile(ctx, profileID)
	if err != nil {
		return nil, err
	}
	needle := strings.ToLower(strings.TrimSpace(query))
	providerID = strings.ToLower(strings.TrimSpace(providerID))
	var out []scanner.QuerySet
	for _, watch := range watches {
		if providerID != "" && !stringSliceContainsFold(watch.ProviderScope, providerID) {
			continue
		}
		if needle != "" && !agentMarketWatchMatchesQuery(watch, needle) {
			continue
		}
		out = append(out, watch)
	}
	return out, nil
}

func agentMarketWatchMatchesQuery(watch scanner.QuerySet, needle string) bool {
	if strings.Contains(strings.ToLower(watch.ID), needle) || strings.Contains(strings.ToLower(watch.Name), needle) {
		return true
	}
	for _, keyword := range watch.Keywords {
		if strings.Contains(strings.ToLower(keyword), needle) {
			return true
		}
	}
	return false
}

func receiveAgentPurchaseOrder(ctx context.Context, svc *commerce.Service, profileID, orderID string, params map[string]any) ([]map[string]any, error) {
	orders, err := svc.ListPurchaseOrdersByProfile(ctx, profileID, "all", orderID, 1, 100)
	if err != nil {
		return nil, err
	}
	var updates []map[string]any
	for _, order := range orders.Orders {
		if order.OrderID != orderID {
			continue
		}
		for _, line := range order.LineItems {
			if line.Status == "delivered" || line.Status == "reconciled" {
				continue
			}
			arrival, err := updateAgentPurchaseLine(ctx, svc, profileID, orderID, line.ExpectedArrivalID, "delivered", params)
			if err != nil {
				return nil, err
			}
			updates = append(updates, map[string]any{
				"item_id":             line.ItemID,
				"lifecycle_entry_id":  line.LifecycleEntryID,
				"expected_arrival_id": arrival.ID,
				"status":              arrival.Status,
			})
		}
	}
	if len(updates) == 0 {
		return nil, fmt.Errorf("matching unreceived purchase lines not found")
	}
	return updates, nil
}

func updateAgentPurchaseLine(ctx context.Context, svc *commerce.Service, profileID, orderID, lineItemID, status string, params map[string]any) (commerce.ExpectedArrival, error) {
	arrival, err := findAgentPurchaseArrival(ctx, svc, profileID, orderID, lineItemID)
	if err != nil {
		return commerce.ExpectedArrival{}, err
	}
	arrival.Status = status
	if deliveredOn := stringMapParam(params, "delivered_on"); deliveredOn != "" {
		arrival.DeliveredOn = deliveredOn
	}
	if status == "reconciled" {
		arrival.ReconciledInstanceID = firstNonEmptyString(
			stringMapParam(params, "instance_id"),
			stringMapParam(params, "target_instance_id"),
			stringMapParam(params, "item_id"),
		)
	}
	if note := stringMapParam(params, "notes"); note != "" {
		arrival.Notes = strings.TrimSpace(strings.TrimSpace(arrival.Notes) + " " + note)
	}
	if err := svc.UpdateArrivalForProfile(ctx, profileID, arrival); err != nil {
		return commerce.ExpectedArrival{}, err
	}
	return svc.GetArrivalByIDForProfile(ctx, profileID, arrival.ID)
}

func findAgentPurchaseArrival(ctx context.Context, svc *commerce.Service, profileID, orderID, lineItemID string) (commerce.ExpectedArrival, error) {
	orders, err := svc.ListPurchaseOrdersByProfile(ctx, profileID, "all", orderID, 1, 100)
	if err != nil {
		return commerce.ExpectedArrival{}, err
	}
	for _, order := range orders.Orders {
		if order.OrderID != orderID {
			continue
		}
		for _, line := range order.LineItems {
			if line.ExpectedArrivalID == lineItemID || line.LifecycleEntryID == lineItemID || line.ItemID == lineItemID {
				if line.ExpectedArrivalID == "" {
					return commerce.ExpectedArrival{}, fmt.Errorf("purchase line has no expected arrival")
				}
				return svc.GetArrivalByIDForProfile(ctx, profileID, line.ExpectedArrivalID)
			}
		}
	}
	return commerce.ExpectedArrival{}, fmt.Errorf("matching purchase line not found")
}

func ensureAgentPurchaseOrderItem(ctx context.Context, conn *sql.DB, profileID string, params map[string]any) (string, bool, error) {
	itemID := stringMapParam(params, "item_id")
	if itemID != "" {
		return itemID, false, nil
	}
	title := firstNonEmptyString(
		stringMapParam(params, "title"),
		stringMapParam(params, "item_title"),
		stringMapParam(params, "description"),
	)
	if title == "" {
		return "", false, fmt.Errorf("item context required")
	}
	itemID = uuid.NewString()
	partNumber := firstNonEmptyString(
		stringMapParam(params, "part_number"),
		strings.ToUpper("AGENT-PURCHASE-"+providerSettingSlug(firstNonEmptyString(stringMapParam(params, "order_id"), stringMapParam(params, "result_id"), itemID))),
	)
	sourceURL := stringMapParam(params, "source_url")
	sourceURLs := "[]"
	if sourceURL != "" {
		encoded, _ := json.Marshal([]string{sourceURL})
		sourceURLs = string(encoded)
	}
	_, err := conn.ExecContext(ctx, `
		INSERT INTO canonical_items(
			id, profile_id, brand, category, part_number, title, status, notes, source_urls_json, created_by, updated_by
		) VALUES (?, ?, ?, ?, ?, ?, 'active', ?, ?, 'agent.purchases.create_order', 'agent.purchases.create_order')
	`, itemID, strings.TrimSpace(profileID), firstNonEmptyString(stringMapParam(params, "brand"), stringMapParam(params, "purchase_source"), "Purchase"), firstNonEmptyString(stringMapParam(params, "category"), "Purchases"), partNumber, title, purchaseAgentSkillNotes(params), sourceURLs)
	if err != nil {
		return "", false, err
	}
	return itemID, true, nil
}

func purchaseAgentSkillNotes(params map[string]any) string {
	var notes []string
	if sourceURL := stringMapParam(params, "source_url"); sourceURL != "" {
		notes = append(notes, "source_url="+sourceURL)
	}
	if resultID := stringMapParam(params, "result_id"); resultID != "" {
		notes = append(notes, "result_id="+resultID)
	}
	if seller := stringMapParam(params, "seller"); seller != "" {
		notes = append(notes, "seller="+seller)
	}
	if tracking := stringMapParam(params, "tracking"); tracking != "" {
		notes = append(notes, "tracking="+tracking)
	}
	if note := stringMapParam(params, "notes"); note != "" {
		notes = append(notes, "notes="+note)
	}
	return strings.Join(notes, " ")
}

func agentSkillProviderHealthSnapshot(ctx context.Context, conn *sql.DB, providerID string) map[string]any {
	providerID = strings.TrimSpace(providerID)
	if providerID == "" {
		return map[string]any{
			"provider":    "",
			"status":      "unknown",
			"state":       "disabled",
			"category":    "not_checked",
			"label":       "Not checked yet",
			"last_error":  "",
			"next_action": "select_provider",
		}
	}
	if conn == nil {
		return providerHealthResponse(map[string]string{"provider": providerID, "status": "unknown", "message": ""})
	}
	var status, msg, updated string
	var retryAfterSeconds int
	err := conn.QueryRowContext(ctx, `
		SELECT status, message, updated_at, retry_after_seconds
		FROM provider_health WHERE provider = ?
	`, providerID).Scan(&status, &msg, &updated, &retryAfterSeconds)
	if err != nil {
		return providerHealthResponse(map[string]string{"provider": providerID, "status": "unknown", "message": ""})
	}
	return providerHealthResponse(map[string]string{
		"provider":            providerID,
		"status":              status,
		"message":             msg,
		"updated_at":          updated,
		"retry_after_seconds": fmt.Sprintf("%d", retryAfterSeconds),
	})
}

func applyAgentSettingsDataSkill(ctx context.Context, conn *sql.DB, skillID, profileID string, params map[string]any) (map[string]any, string, error) {
	result := map[string]any{
		"profile_id": profileID,
		"read_only":  false,
	}
	switch skillID {
	case "cabinet.settings.update_profile":
		return applyAgentSettingUpdate(ctx, conn, profileID, "settings.profile.update", result, params)
	case "cabinet.settings.update_account":
		return applyAgentSettingUpdate(ctx, conn, profileID, "settings.account.update", result, params)
	case "cabinet.settings.update_appearance":
		return applyAgentSettingUpdate(ctx, conn, profileID, "settings.appearance.update", result, params)
	case "cabinet.storage.show_status":
		result["operation"] = "storage.status.show"
		result["read_only"] = true
		result["storage_status"] = "available"
		result["backup_status"] = "not_configured"
		result["next_action"] = "Configure a backup target before relying on restore workflows."
	case "cabinet.storage.configure_backup":
		backupPath := stringMapParam(params, "backup_path")
		if backupPath == "" {
			backupPath = stringMapParam(params, "backup_target")
		}
		if backupPath == "" {
			return nil, "storage_backup_target_required", fmt.Errorf("backup target required")
		}
		if err := persistAgentProfileSettings(ctx, conn, profileID, map[string]string{"storage.backup_target": backupPath}); err != nil {
			return nil, "storage_backup_settings_persist_failed", err
		}
		result["operation"] = "storage.backup.configure"
		result["backup_target"] = backupPath
		result["settings_persisted"] = []string{"storage.backup_target"}
		result["status"] = "confirmed"
	case "cabinet.data.import_file":
		filePath := stringMapParam(params, "file_path")
		if filePath == "" {
			return nil, "data_import_file_required", fmt.Errorf("import file required")
		}
		result["operation"] = "data.import.file"
		result["file_path"] = filePath
		result["status"] = "confirmed"
		result["impact"] = "import_preview_confirmed"
	case "cabinet.data.export_bundle":
		result["operation"] = "data.export.bundle"
		result["read_only"] = true
		result["export_scope"] = stringMapParam(params, "export_scope")
		result["status"] = "ready"
	case "cabinet.data.restore_backup":
		backupPath := stringMapParam(params, "backup_path")
		if backupPath == "" {
			return nil, "data_backup_target_required", fmt.Errorf("backup target required")
		}
		result["operation"] = "data.backup.restore"
		result["backup_path"] = backupPath
		result["destructive_confirmation"] = true
		result["status"] = "confirmed"
	case "cabinet.maintenance.run_safe_check":
		result["operation"] = "maintenance.safe_check"
		result["read_only"] = true
		result["status"] = "healthy"
	}
	return result, "", nil
}

func applyAgentSettingUpdate(ctx context.Context, conn *sql.DB, profileID, operation string, result map[string]any, params map[string]any) (map[string]any, string, error) {
	settingKey := stringMapParam(params, "setting_key")
	if settingKey == "" {
		settingKey = stringMapParam(params, "setting_scope")
	}
	if settingKey == "" {
		return nil, "settings_target_required", fmt.Errorf("settings target required")
	}
	settingValue := stringMapParam(params, "setting_value")
	if settingValue == "" {
		return nil, "settings_value_required", fmt.Errorf("settings value required")
	}
	if err := persistAgentProfileSettings(ctx, conn, profileID, map[string]string{settingKey: settingValue}); err != nil {
		return nil, "settings_persist_failed", err
	}
	result["operation"] = operation
	result["setting_key"] = settingKey
	result["setting_scope"] = stringMapParam(params, "setting_scope")
	result["settings_persisted"] = []string{settingKey}
	result["status"] = "confirmed"
	return result, "", nil
}

func persistAgentProviderSettings(ctx context.Context, conn *sql.DB, profileID, providerID string, params map[string]any) ([]string, bool, error) {
	keys := providerSettingsKeys(providerID)
	settings := map[string]string{
		keys.EnabledKey: "true",
	}
	if setupStep := stringMapParam(params, "setup_step"); setupStep != "" {
		settings["integration."+providerSettingSlug(providerID)+".setup_step"] = setupStep
	}
	if baseURL := stringMapParam(params, "base_url"); baseURL != "" {
		settings[keys.BaseURLKey] = baseURL
	}
	if marketplace := stringMapParam(params, "marketplace"); marketplace != "" {
		settings[keys.MarketplaceKey] = marketplace
	}
	if itemsPerPage := stringMapParam(params, "items_per_page"); itemsPerPage != "" {
		settings[keys.ItemsPerPageKey] = itemsPerPage
	}
	if err := persistAgentProfileSettings(ctx, conn, profileID, settings); err != nil {
		return nil, false, err
	}
	secretPersisted := false
	if secret := firstAgentSecretParam(params); secret != "" {
		if err := profile.NewRepository(conn).PutSecret(ctx, profileID, keys.TokenKey, secret); err != nil {
			return nil, false, err
		}
		secretPersisted = true
	}
	persisted := make([]string, 0, len(settings))
	for key := range settings {
		persisted = append(persisted, key)
	}
	return persisted, secretPersisted, nil
}

func persistAgentProfileSettings(ctx context.Context, conn *sql.DB, profileID string, values map[string]string) error {
	if conn == nil {
		return fmt.Errorf("database connection required")
	}
	return profile.NewRepository(conn).PutSettings(ctx, profileID, values)
}

func firstAgentSecretParam(params map[string]any) string {
	for _, key := range []string{"provider_secret", "provider_token", "access_token", "bearer_token", "api_key"} {
		if value := stringMapParam(params, key); value != "" {
			return value
		}
	}
	return ""
}

func providerSettingSlug(providerID string) string {
	slug := strings.TrimSpace(strings.ToLower(providerID))
	slug = strings.ReplaceAll(slug, "-", "_")
	slug = strings.ReplaceAll(slug, ".", "_")
	if slug == "" {
		return "provider"
	}
	return slug
}

func applyAgentInboxSkill(ctx context.Context, chatSvc *chat.Service, profileID string, params map[string]any, status string) (map[string]any, string, error) {
	inboxID := stringMapParam(params, "notification_id")
	if inboxID == "" {
		inboxID = stringMapParam(params, "target_notification")
	}
	if inboxID == "" {
		return nil, "inbox_notification_target_required", fmt.Errorf("notification target required")
	}
	item, err := chatSvc.UpdateInboxItemStatus(ctx, profileID, inboxID, status)
	if err != nil {
		return nil, "inbox_notification_not_found", err
	}
	return map[string]any{"inbox_item": item}, "", nil
}

func resolveAgentSkillUserTarget(ctx context.Context, conn *sql.DB, profileID string, params map[string]any) (string, string, error) {
	target := stringMapParam(params, "target_user")
	email := stringMapParam(params, "target_email")
	if target == "" && email == "" {
		return "", "users_admin_target_required", fmt.Errorf("target user required")
	}
	users, err := listRuntimeUsers(ctx, conn, profileID)
	if err != nil {
		return "", "users_admin_target_lookup_failed", err
	}
	for _, user := range users {
		if target != "" && (user.ID == target || strings.EqualFold(user.Email, target) || strings.EqualFold(user.Username, target)) {
			return user.ID, "", nil
		}
		if email != "" && strings.EqualFold(user.Email, email) {
			return user.ID, "", nil
		}
	}
	return "", "users_admin_target_not_found", fmt.Errorf("target user not found")
}

func agentUserBlocker(err error) string {
	switch strings.TrimSpace(err.Error()) {
	case "protected_admin_required":
		return "users_admin_protected_owner_change_blocked"
	case "user_not_found":
		return "users_admin_target_not_found"
	case "invalid_invite":
		return "users_admin_target_email_required"
	default:
		return "users_admin_apply_failed"
	}
}

func containsSecretParameter(params map[string]any) bool {
	for key, value := range params {
		if value == nil {
			continue
		}
		normalized := strings.ToLower(strings.TrimSpace(key))
		if strings.Contains(normalized, "secret") || strings.Contains(normalized, "token") || strings.Contains(normalized, "api_key") || strings.Contains(normalized, "password") {
			return true
		}
	}
	return false
}

func stringMapParam(params map[string]any, key string) string {
	value, ok := params[key]
	if !ok || value == nil {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(text)
}

func stringSliceMapParam(params map[string]any, key string) []string {
	value, ok := params[key]
	if !ok || value == nil {
		return nil
	}
	switch typed := value.(type) {
	case []string:
		return compactStringSlice(typed)
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			text, ok := item.(string)
			if !ok {
				continue
			}
			out = append(out, text)
		}
		return compactStringSlice(out)
	case string:
		if strings.TrimSpace(typed) == "" {
			return nil
		}
		return compactStringSlice(strings.Split(typed, ","))
	default:
		return nil
	}
}

func compactStringSlice(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if text := strings.TrimSpace(value); text != "" {
			out = append(out, text)
		}
	}
	return out
}

func stringSliceContainsFold(values []string, want string) bool {
	want = strings.TrimSpace(strings.ToLower(want))
	for _, value := range values {
		if strings.TrimSpace(strings.ToLower(value)) == want {
			return true
		}
	}
	return false
}

func boolMapParam(params map[string]any, key string) bool {
	value, ok := params[key]
	if !ok || value == nil {
		return false
	}
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		parsed, _ := strconv.ParseBool(strings.TrimSpace(typed))
		return parsed
	default:
		return false
	}
}

func intMapParam(params map[string]any, key string) int {
	value, ok := params[key]
	if !ok || value == nil {
		return 0
	}
	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case string:
		parsed, _ := strconv.Atoi(strings.TrimSpace(typed))
		return parsed
	default:
		return 0
	}
}

func floatMapParam(params map[string]any, key string) float64 {
	value, ok := params[key]
	if !ok || value == nil {
		return 0
	}
	switch typed := value.(type) {
	case float64:
		return typed
	case float32:
		return float64(typed)
	case int:
		return float64(typed)
	case int64:
		return float64(typed)
	case string:
		parsed, _ := strconv.ParseFloat(strings.TrimSpace(typed), 64)
		return parsed
	default:
		return 0
	}
}
