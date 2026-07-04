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
