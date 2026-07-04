package app

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/collectors-tech/cabinet/internal/chat"
	"github.com/collectors-tech/cabinet/internal/profile"
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
		return applyAgentMarketWatchSkill(ctx, conn, skillID, params)
	case "cabinet.purchases.search_orders",
		"cabinet.purchases.create_order",
		"cabinet.purchases.add_line_item",
		"cabinet.purchases.receive_order",
		"cabinet.purchases.receive_line_item",
		"cabinet.purchases.reconcile_item",
		"cabinet.purchases.review_purchase":
		return applyAgentPurchasesSkill(skillID, profileID, params)
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

func applyAgentMarketWatchSkill(ctx context.Context, conn *sql.DB, skillID string, params map[string]any) (map[string]any, string, error) {
	providerID := stringMapParam(params, "provider_id")
	if providerID == "" {
		providerID = stringMapParam(params, "provider_name")
	}
	if providerID == "" {
		return nil, "market_watch_provider_required", fmt.Errorf("provider required")
	}
	result := map[string]any{
		"provider_id":              providerID,
		"external_write_claimed":   false,
		"provenance_preserved":     false,
		"provider_health":          agentSkillProviderHealthSnapshot(ctx, conn, providerID),
		"live_provider_dispatched": false,
	}
	switch skillID {
	case "cabinet.market_watch.search_watches":
		result["operation"] = "market_watch.watch.search"
		result["read_only"] = true
		result["query"] = stringMapParam(params, "query")
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
		result["operation"] = "market_watch.watch.create"
		result["status"] = "confirmed_preview"
		result["watch_query"] = query
		result["next_action"] = "Persist the saved watch through the Market Watch workflow once provider setup is healthy."
	case "cabinet.market_watch.update_saved_watch", "cabinet.market_watch.run_watch":
		watchID := stringMapParam(params, "watch_id")
		if watchID == "" {
			return nil, "market_watch_watch_required", fmt.Errorf("watch required")
		}
		if skillID == "cabinet.market_watch.update_saved_watch" {
			result["operation"] = "market_watch.watch.update"
			result["status"] = "confirmed_preview"
		} else {
			result["operation"] = "market_watch.watch.run"
			result["status"] = "confirmed_provider_ready_check"
		}
		result["watch_id"] = watchID
		result["query"] = stringMapParam(params, "query")
		result["next_action"] = "Review provider-health evidence before treating any live result sync as complete."
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
		result["status"] = "confirmed_preview"
		result["next_action"] = "Review the destination preview before applying Wishlist, Purchases, or Inventory state."
	}
	return result, "", nil
}

func applyAgentPurchasesSkill(skillID, profileID string, params map[string]any) (map[string]any, string, error) {
	result := map[string]any{
		"profile_id":           profileID,
		"provenance_preserved": false,
	}
	switch skillID {
	case "cabinet.purchases.search_orders":
		result["operation"] = "purchases.orders.search"
		result["read_only"] = true
		result["query"] = stringMapParam(params, "query")
		result["next_action"] = "Select an order or line item before applying purchase state."
	case "cabinet.purchases.review_purchase":
		result["operation"] = "purchases.order.review"
		result["read_only"] = true
		result["order_id"] = stringMapParam(params, "order_id")
		result["review_status"] = stringMapParam(params, "review_status")
	case "cabinet.purchases.create_order":
		source := stringMapParam(params, "purchase_source")
		if source == "" {
			source = stringMapParam(params, "source")
		}
		if source == "" {
			return nil, "purchases_source_required", fmt.Errorf("purchase source required")
		}
		result["operation"] = "purchases.order.create"
		result["purchase_source"] = source
		result["status"] = "confirmed_preview"
		result["provenance_preserved"] = true
	case "cabinet.purchases.add_line_item":
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
		result["operation"] = "purchases.order.add_line_item"
		result["order_id"] = orderID
		result["item_id"] = itemID
		result["source"] = stringMapParam(params, "source")
		result["result_id"] = stringMapParam(params, "result_id")
		result["status"] = "confirmed_preview"
		result["provenance_preserved"] = true
	case "cabinet.purchases.receive_order":
		orderID := stringMapParam(params, "order_id")
		if orderID == "" {
			return nil, "purchases_order_required", fmt.Errorf("order required")
		}
		result["operation"] = "purchases.order.receive"
		result["order_id"] = orderID
		result["status"] = "confirmed_preview"
	case "cabinet.purchases.receive_line_item":
		orderID := stringMapParam(params, "order_id")
		if orderID == "" {
			return nil, "purchases_order_required", fmt.Errorf("order required")
		}
		lineItemID := stringMapParam(params, "line_item_id")
		if lineItemID == "" {
			return nil, "purchases_line_item_required", fmt.Errorf("line item required")
		}
		result["operation"] = "purchases.line_item.receive"
		result["order_id"] = orderID
		result["line_item_id"] = lineItemID
		result["status"] = "confirmed_preview"
	case "cabinet.purchases.reconcile_item":
		orderID := stringMapParam(params, "order_id")
		itemID := stringMapParam(params, "item_id")
		if orderID == "" {
			return nil, "purchases_order_required", fmt.Errorf("order required")
		}
		if itemID == "" {
			return nil, "purchases_item_required", fmt.Errorf("item required")
		}
		result["operation"] = "purchases.item.reconcile"
		result["order_id"] = orderID
		result["item_id"] = itemID
		result["status"] = "confirmed_preview"
		result["provenance_preserved"] = true
	}
	return result, "", nil
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
