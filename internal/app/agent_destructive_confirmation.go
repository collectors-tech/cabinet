package app

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/collectors-tech/cabinet/internal/agentskills"
	"github.com/collectors-tech/cabinet/internal/backup"
	"github.com/collectors-tech/cabinet/internal/profile"
	"github.com/google/uuid"
)

const agentStrongConfirmationTTL = 5 * time.Minute

type agentStrongConfirmation struct {
	ID                string
	PreviewID         string
	ProfileID         string
	SkillID           string
	TargetFingerprint string
	TokenHash         string
	Status            string
	CreatedAt         string
	ExpiresAt         string
	UsedAt            string
}

type agentStrongConfirmationResponse struct {
	ConfirmationToken string         `json:"confirmation_token"`
	ExpiresAt         string         `json:"expires_at"`
	Action            string         `json:"action"`
	Target            map[string]any `json:"target"`
	Impact            []string       `json:"impact"`
	Recovery          string         `json:"recovery"`
}

func issueAgentStrongConfirmation(ctx context.Context, conn *sql.DB, profiles *profile.Repository, backupSvc *backup.Service, registry agentskills.Registry, requested agentskills.PreviewRequest) (agentStrongConfirmationResponse, error) {
	durableAgentSkillPreviewMutationMu.Lock()
	defer durableAgentSkillPreviewMutationMu.Unlock()

	record, err := getDurableAgentSkillPreview(ctx, conn, requested.ProfileID, requested.PreviewID)
	if err != nil {
		return agentStrongConfirmationResponse{}, err
	}
	if record.Status != "previewed" {
		return agentStrongConfirmationResponse{}, agentSkillPreviewStatusError(record.Status)
	}
	if previewExpired(record, time.Now().UTC()) {
		return agentStrongConfirmationResponse{}, agentSkillPreviewStatusError("expired")
	}
	skill, ok := registry.Resolve(record.SkillID)
	if !ok || skill.SafetyLevel != agentskills.SafetyDestructive || !skill.Permissions.Destructive {
		return agentStrongConfirmationResponse{}, &agentSkillPreviewLifecycleError{
			Code: "agent_skill_strong_confirmation_not_required", Recoverable: false,
			NextAction: "Use the ordinary preview Apply control for this non-destructive action.",
		}
	}
	storedRequest := durableAgentSkillPreviewRequest(record, false)
	storedRequest, err = hydrateAgentUsersTargetFromServer(ctx, conn, storedRequest)
	if err != nil {
		return agentStrongConfirmationResponse{}, err
	}
	authority, err := reviewAgentSkillAuthority(ctx, profiles, registry, storedRequest, "direct-api-strong-confirmation")
	if err != nil {
		return agentStrongConfirmationResponse{}, fmt.Errorf("review destructive confirmation authority: %w", err)
	}
	if !authority.PreviewAllowed {
		return agentStrongConfirmationResponse{}, &agentSkillPreviewLifecycleError{
			Code: firstNonEmptyString(authority.Blocker, "agent_skill_preview_authority_denied"), Recoverable: true, NextAction: authority.NextAction,
		}
	}
	currentPreview, err := registry.Preview(storedRequest)
	if err != nil {
		return agentStrongConfirmationResponse{}, &agentSkillPreviewLifecycleError{Code: "agent_skill_preview_skill_unavailable", Recoverable: true, NextAction: "Create a fresh preview."}
	}
	if !durableAgentSkillPreviewTargetMatches(currentPreview, record) {
		return agentStrongConfirmationResponse{}, &agentSkillPreviewLifecycleError{Code: "agent_skill_preview_target_changed", Recoverable: true, NextAction: "Create a fresh preview for the current target."}
	}
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return agentStrongConfirmationResponse{}, fmt.Errorf("create destructive confirmation token: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)
	tokenHash := sha256.Sum256([]byte(token))
	confirmationTarget, err := currentAgentStrongConfirmationTarget(ctx, conn, backupSvc, record)
	if err != nil {
		return agentStrongConfirmationResponse{}, err
	}
	fingerprint, err := agentStrongConfirmationTargetFingerprint(record.ProfileID, record.SkillID, confirmationTarget)
	if err != nil {
		return agentStrongConfirmationResponse{}, err
	}
	now := time.Now().UTC()
	expiresAt := now.Add(agentStrongConfirmationTTL)
	if previewExpiry, parseErr := time.Parse(time.RFC3339Nano, record.ExpiresAt); parseErr == nil && previewExpiry.Before(expiresAt) {
		expiresAt = previewExpiry
	}
	if _, err := conn.ExecContext(ctx, `UPDATE agent_skill_strong_confirmations SET status = 'superseded' WHERE preview_id = ? AND profile_id = ? AND status = 'confirmed'`, record.ID, record.ProfileID); err != nil {
		return agentStrongConfirmationResponse{}, fmt.Errorf("supersede destructive confirmation: %w", err)
	}
	confirmationID := "asc_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err := conn.ExecContext(ctx, `
		INSERT INTO agent_skill_strong_confirmations(id, preview_id, profile_id, skill_id, target_fingerprint, token_hash, status, created_at, expires_at)
		VALUES (?, ?, ?, ?, ?, ?, 'confirmed', ?, ?)
	`, confirmationID, record.ID, record.ProfileID, record.SkillID, fingerprint, hex.EncodeToString(tokenHash[:]), now.Format(time.RFC3339Nano), expiresAt.Format(time.RFC3339Nano)); err != nil {
		return agentStrongConfirmationResponse{}, fmt.Errorf("persist destructive confirmation: %w", err)
	}
	if err := appendAgentAuthorityDecisionAudit(ctx, profiles, authority, agentSkillAuthorityRequest(storedRequest), "strong_confirmation_issued"); err != nil {
		_, _ = conn.ExecContext(ctx, `DELETE FROM agent_skill_strong_confirmations WHERE id = ? AND status = 'confirmed'`, confirmationID)
		return agentStrongConfirmationResponse{}, fmt.Errorf("record destructive confirmation audit: %w", err)
	}
	action, impact, recovery := agentStrongConfirmationDescription(record.SkillID)
	return agentStrongConfirmationResponse{
		ConfirmationToken: token,
		ExpiresAt:         expiresAt.Format(time.RFC3339Nano),
		Action:            action,
		Target:            confirmationTarget,
		Impact:            impact,
		Recovery:          recovery,
	}, nil
}

func validateAgentStrongConfirmation(ctx context.Context, conn *sql.DB, backupSvc *backup.Service, record durableAgentSkillPreview, token string) (agentStrongConfirmation, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return agentStrongConfirmation{}, &agentSkillPreviewLifecycleError{Code: "strong_confirmation_required", Recoverable: true, NextAction: "Review the destructive impact and use the dedicated strong-confirmation control."}
	}
	var confirmation agentStrongConfirmation
	err := conn.QueryRowContext(ctx, `
		SELECT id, preview_id, profile_id, skill_id, target_fingerprint, token_hash, status, created_at, expires_at, COALESCE(used_at, '')
		FROM agent_skill_strong_confirmations
		WHERE preview_id = ? AND profile_id = ? AND status = 'confirmed'
		ORDER BY created_at DESC LIMIT 1
	`, record.ID, record.ProfileID).Scan(&confirmation.ID, &confirmation.PreviewID, &confirmation.ProfileID, &confirmation.SkillID, &confirmation.TargetFingerprint, &confirmation.TokenHash, &confirmation.Status, &confirmation.CreatedAt, &confirmation.ExpiresAt, &confirmation.UsedAt)
	if err == sql.ErrNoRows {
		return agentStrongConfirmation{}, &agentSkillPreviewLifecycleError{Code: "strong_confirmation_required", Recoverable: true, NextAction: "Create a fresh action-specific strong confirmation."}
	}
	if err != nil {
		return agentStrongConfirmation{}, fmt.Errorf("load destructive confirmation: %w", err)
	}
	if confirmation.SkillID != record.SkillID {
		return agentStrongConfirmation{}, &agentSkillPreviewLifecycleError{Code: "strong_confirmation_context_mismatch", Recoverable: true, NextAction: "Create a fresh confirmation for this exact action."}
	}
	if expiry, parseErr := time.Parse(time.RFC3339Nano, confirmation.ExpiresAt); parseErr != nil || !expiry.After(time.Now().UTC()) {
		_, _ = conn.ExecContext(ctx, `UPDATE agent_skill_strong_confirmations SET status = 'expired' WHERE id = ? AND status = 'confirmed'`, confirmation.ID)
		return agentStrongConfirmation{}, &agentSkillPreviewLifecycleError{Code: "strong_confirmation_expired", Recoverable: true, NextAction: "Review the current target and create a fresh confirmation."}
	}
	presentedHash := sha256.Sum256([]byte(token))
	storedHash, decodeErr := hex.DecodeString(confirmation.TokenHash)
	if decodeErr != nil || len(storedHash) != len(presentedHash) || subtle.ConstantTimeCompare(storedHash, presentedHash[:]) != 1 {
		return agentStrongConfirmation{}, &agentSkillPreviewLifecycleError{Code: "strong_confirmation_invalid", Recoverable: true, NextAction: "Use the confirmation issued for this exact preview."}
	}
	confirmationTarget, err := currentAgentStrongConfirmationTarget(ctx, conn, backupSvc, record)
	if err != nil {
		return agentStrongConfirmation{}, &agentSkillPreviewLifecycleError{Code: "strong_confirmation_target_unavailable", Recoverable: true, NextAction: "Review the current target and create a fresh preview."}
	}
	fingerprint, err := agentStrongConfirmationTargetFingerprint(record.ProfileID, record.SkillID, confirmationTarget)
	if err != nil {
		return agentStrongConfirmation{}, err
	}
	if subtle.ConstantTimeCompare([]byte(confirmation.TargetFingerprint), []byte(fingerprint)) != 1 {
		return agentStrongConfirmation{}, &agentSkillPreviewLifecycleError{Code: "strong_confirmation_target_changed", Recoverable: true, NextAction: "Review the changed target and create a fresh preview."}
	}
	return confirmation, nil
}

func claimAgentStrongConfirmation(ctx context.Context, conn *sql.DB, confirmation agentStrongConfirmation) error {
	result, err := conn.ExecContext(ctx, `UPDATE agent_skill_strong_confirmations SET status = 'used', used_at = ? WHERE id = ? AND status = 'confirmed'`, time.Now().UTC().Format(time.RFC3339Nano), confirmation.ID)
	if err != nil {
		return fmt.Errorf("consume destructive confirmation: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil || rows != 1 {
		return &agentSkillPreviewLifecycleError{Code: "strong_confirmation_already_used", Recoverable: true, NextAction: "Create a fresh preview and confirmation before retrying."}
	}
	return nil
}

func agentStrongConfirmationTargetFingerprint(profileID, skillID string, target map[string]any) (string, error) {
	payload, err := json.Marshal(map[string]any{"profile_id": profileID, "skill_id": skillID, "target": nonNilAgentSkillPreviewMap(target)})
	if err != nil {
		return "", fmt.Errorf("fingerprint destructive target: %w", err)
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:]), nil
}

func currentAgentStrongConfirmationTarget(ctx context.Context, conn *sql.DB, backupSvc *backup.Service, record durableAgentSkillPreview) (map[string]any, error) {
	switch record.SkillID {
	case "cabinet.data.restore_backup":
		if backupSvc == nil {
			return nil, fmt.Errorf("backup service unavailable")
		}
		validation, err := backupSvc.InspectBackup(stringMapParam(record.Parameters, "backup_path"), record.ProfileID)
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"backup_file": validation.FileName, "backup_size_bytes": validation.SizeBytes,
			"backup_created_at": validation.CreatedAt, "archive_format": validation.ArchiveFormat,
			"backup_format": validation.BackupFormat, "database_sha256": validation.DatabaseSHA256,
			"integrity_check": validation.IntegrityCheck, "lifecycle_schema_compatible": validation.LifecycleSchemaCompatible,
			"profile_included": validation.ProfileIncluded, "compatible": validation.Compatible,
		}, nil
	case "cabinet.users.remove_user":
		request, err := hydrateAgentUsersTargetFromServer(ctx, conn, durableAgentSkillPreviewRequest(record, false))
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"target_user": request.Parameters["target_user"], "target_email": request.Parameters["target_email"],
			"target_display_name": request.Parameters["target_display_name"], "target_role_current": request.Parameters["target_role_current"],
			"target_status_current": request.Parameters["target_status_current"], "target_updated_at": request.Parameters["target_updated_at"],
			"protected": request.Parameters["protected"],
		}, nil
	default:
		return record.Target, nil
	}
}

func agentStrongConfirmationDescription(skillID string) (string, []string, string) {
	switch strings.TrimSpace(skillID) {
	case "cabinet.users.remove_user":
		return "remove_user", []string{"Remove the selected user from this profile.", "Revoke that user's Cabinet access.", "This cannot remove or demote the last active owner."}, "Invite the user again if access must be restored."
	case "cabinet.data.restore_backup":
		return "restore_backup", []string{"Replace current Cabinet data with the selected compatible backup.", "Preserve a pre-restore recovery backup before replacement.", "Revalidate backup identity and integrity immediately before apply."}, "Use the pre-restore recovery backup if the restored state is not acceptable."
	default:
		return "destructive_action", []string{"Apply the reviewed destructive change to the exact selected target."}, "Create a fresh preview if the target or desired outcome changes."
	}
}
