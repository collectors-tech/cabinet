package app

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/collectors-tech/cabinet/internal/chat"
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
		return applyAgentIntegrationSkill(skillID, params)
	case "cabinet.settings.update_profile",
		"cabinet.settings.update_account",
		"cabinet.settings.update_appearance",
		"cabinet.storage.show_status",
		"cabinet.storage.configure_backup",
		"cabinet.data.import_file",
		"cabinet.data.export_bundle",
		"cabinet.data.restore_backup",
		"cabinet.maintenance.run_safe_check":
		return applyAgentSettingsDataSkill(skillID, profileID, params)
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

func applyAgentIntegrationSkill(skillID string, params map[string]any) (map[string]any, string, error) {
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
		result["connection_status"] = "setup_needed"
		result["next_action"] = "Review non-secret setup requirements before running a live provider health check."
	case "cabinet.integrations.explain_required_setup":
		result["operation"] = "integrations.provider.explain_setup"
		result["read_only"] = true
		result["setup_required"] = true
		result["next_action"] = "Open Integrations settings for provider-specific credential and permission setup."
	case "cabinet.integrations.configure_provider":
		result["operation"] = "integrations.provider.configure"
		result["status"] = "confirmed"
		result["setup_step"] = stringMapParam(params, "setup_step")
		result["next_action"] = "Persist provider credentials through the provider settings surface before live health validation."
	case "cabinet.integrations.repair_provider":
		result["operation"] = "integrations.provider.repair"
		result["status"] = "confirmed"
		result["next_action"] = "Run a provider health check after reviewing repaired setup steps."
	case "cabinet.integrations.disable_provider":
		result["operation"] = "integrations.provider.disable"
		result["status"] = "confirmed"
		result["next_action"] = "Confirm provider disabled state in Integrations before routing provider-backed workflows."
	}
	return result, "", nil
}

func applyAgentSettingsDataSkill(skillID, profileID string, params map[string]any) (map[string]any, string, error) {
	result := map[string]any{
		"profile_id": profileID,
		"read_only":  false,
	}
	switch skillID {
	case "cabinet.settings.update_profile":
		return applyAgentSettingUpdate("settings.profile.update", result, params)
	case "cabinet.settings.update_account":
		return applyAgentSettingUpdate("settings.account.update", result, params)
	case "cabinet.settings.update_appearance":
		return applyAgentSettingUpdate("settings.appearance.update", result, params)
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
		result["operation"] = "storage.backup.configure"
		result["backup_target"] = backupPath
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

func applyAgentSettingUpdate(operation string, result map[string]any, params map[string]any) (map[string]any, string, error) {
	settingKey := stringMapParam(params, "setting_key")
	if settingKey == "" {
		settingKey = stringMapParam(params, "setting_scope")
	}
	if settingKey == "" {
		return nil, "settings_target_required", fmt.Errorf("settings target required")
	}
	result["operation"] = operation
	result["setting_key"] = settingKey
	result["setting_scope"] = stringMapParam(params, "setting_scope")
	result["status"] = "confirmed"
	return result, "", nil
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
