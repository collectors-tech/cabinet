package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/collectors-tech/cabinet/internal/agentskills"
	"github.com/collectors-tech/cabinet/internal/backup"
	"github.com/collectors-tech/cabinet/internal/chat"
	"github.com/collectors-tech/cabinet/internal/profile"
	"github.com/google/uuid"
)

const agentSkillPreviewTTL = 15 * time.Minute

var durableAgentSkillPreviewMutationMu sync.Mutex

type durableAgentSkillPreview struct {
	ID              string
	ProfileID       string
	SkillID         string
	SourceSurface   string
	SourceChannel   string
	SourceThreadID  string
	SourceMessageID string
	AgentContext    map[string]any
	Parameters      map[string]any
	Target          map[string]any
	Result          map[string]any
	SecretRef       string
	Status          string
	ErrorCode       string
	CreatedAt       string
	ExpiresAt       string
	AppliedAt       string
	CancelledAt     string
}

type agentSkillPreviewLifecycleError struct {
	Code        string
	NextAction  string
	Recoverable bool
}

func (e *agentSkillPreviewLifecycleError) Error() string {
	return e.Code
}

func createDurableAgentSkillPreview(ctx context.Context, conn *sql.DB, req agentskills.PreviewRequest, preview agentskills.PreviewResponse) (durableAgentSkillPreview, error) {
	if conn == nil {
		return durableAgentSkillPreview{}, fmt.Errorf("agent skill preview store required")
	}
	id := "asp_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	parametersForRecord := nonNilAgentSkillPreviewMap(req.Parameters)
	agentContextForRecord := nonNilAgentSkillPreviewMap(req.AgentContext)
	secretRef := ""
	if containsSensitiveAgentSkillPreviewValue(req.Parameters) || containsSensitiveAgentSkillPreviewValue(req.AgentContext) {
		secretEnvelope, err := json.Marshal(map[string]any{
			"parameters":    nonNilAgentSkillPreviewMap(req.Parameters),
			"agent_context": nonNilAgentSkillPreviewMap(req.AgentContext),
		})
		if err != nil {
			return durableAgentSkillPreview{}, fmt.Errorf("encode pending Agent Skill preview secret: %w", err)
		}
		secretRef = "agent_skill_preview_token_" + strings.TrimPrefix(id, "asp_")
		if err := profile.NewRepository(conn).PutSecret(ctx, strings.TrimSpace(req.ProfileID), secretRef, string(secretEnvelope)); err != nil {
			return durableAgentSkillPreview{}, fmt.Errorf("store pending Agent Skill preview secret: %w", err)
		}
		parametersForRecord = redactedAgentSkillPreviewMap(req.Parameters)
		agentContextForRecord = redactedAgentSkillPreviewMap(req.AgentContext)
	}
	parametersJSON, err := json.Marshal(parametersForRecord)
	if err != nil {
		return durableAgentSkillPreview{}, fmt.Errorf("encode Agent Skill preview parameters: %w", err)
	}
	agentContextJSON, err := json.Marshal(agentContextForRecord)
	if err != nil {
		return durableAgentSkillPreview{}, fmt.Errorf("encode Agent Skill preview context: %w", err)
	}
	targetJSON, err := json.Marshal(nonNilAgentSkillPreviewMap(preview.Target))
	if err != nil {
		return durableAgentSkillPreview{}, fmt.Errorf("encode Agent Skill preview target: %w", err)
	}
	now := time.Now().UTC()
	record := durableAgentSkillPreview{
		ID:              id,
		ProfileID:       strings.TrimSpace(req.ProfileID),
		SkillID:         strings.TrimSpace(req.SkillID),
		SourceSurface:   strings.TrimSpace(req.SourceSurface),
		SourceChannel:   strings.TrimSpace(req.SourceChannel),
		SourceThreadID:  strings.TrimSpace(req.SourceThreadID),
		SourceMessageID: strings.TrimSpace(req.SourceMessageID),
		AgentContext:    nonNilAgentSkillPreviewMap(req.AgentContext),
		Parameters:      nonNilAgentSkillPreviewMap(req.Parameters),
		Target:          nonNilAgentSkillPreviewMap(preview.Target),
		Result:          map[string]any{},
		SecretRef:       secretRef,
		Status:          "previewed",
		CreatedAt:       now.Format(time.RFC3339Nano),
		ExpiresAt:       now.Add(agentSkillPreviewTTL).Format(time.RFC3339Nano),
	}
	if _, err := conn.ExecContext(ctx, `
		INSERT INTO agent_skill_previews(
			id, profile_id, skill_id, source_surface, source_channel, source_thread_id, source_message_id,
			agent_context_json, parameters_json, target_json, result_json, secret_ref, status, error_code, created_at, expires_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, '{}', ?, 'previewed', '', ?, ?)
	`, record.ID, record.ProfileID, record.SkillID, record.SourceSurface, record.SourceChannel,
		record.SourceThreadID, record.SourceMessageID, string(agentContextJSON), string(parametersJSON), string(targetJSON),
		record.SecretRef, record.CreatedAt, record.ExpiresAt); err != nil {
		if record.SecretRef != "" {
			_ = profile.NewRepository(conn).DeleteSecret(ctx, record.ProfileID, record.SecretRef)
		}
		return durableAgentSkillPreview{}, fmt.Errorf("persist Agent Skill preview: %w", err)
	}
	return record, nil
}

func getDurableAgentSkillPreview(ctx context.Context, conn *sql.DB, profileID, previewID string) (durableAgentSkillPreview, error) {
	if conn == nil {
		return durableAgentSkillPreview{}, fmt.Errorf("agent skill preview store required")
	}
	var record durableAgentSkillPreview
	var agentContextJSON, parametersJSON, targetJSON, resultJSON string
	err := conn.QueryRowContext(ctx, `
		SELECT id, profile_id, skill_id, source_surface, source_channel, source_thread_id, source_message_id,
			agent_context_json, parameters_json, target_json, result_json, secret_ref, status, error_code, created_at, expires_at,
			COALESCE(applied_at, ''), COALESCE(cancelled_at, '')
		FROM agent_skill_previews
		WHERE id = ? AND profile_id = ?
	`, strings.TrimSpace(previewID), strings.TrimSpace(profileID)).Scan(
		&record.ID, &record.ProfileID, &record.SkillID, &record.SourceSurface, &record.SourceChannel,
		&record.SourceThreadID, &record.SourceMessageID, &agentContextJSON, &parametersJSON, &targetJSON,
		&resultJSON, &record.SecretRef, &record.Status, &record.ErrorCode, &record.CreatedAt, &record.ExpiresAt,
		&record.AppliedAt, &record.CancelledAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return durableAgentSkillPreview{}, agentSkillPreviewStatusError("not_found")
		}
		return durableAgentSkillPreview{}, err
	}
	if err := decodeAgentSkillPreviewMap(agentContextJSON, &record.AgentContext); err != nil {
		return durableAgentSkillPreview{}, err
	}
	if err := decodeAgentSkillPreviewMap(parametersJSON, &record.Parameters); err != nil {
		return durableAgentSkillPreview{}, err
	}
	if err := decodeAgentSkillPreviewMap(targetJSON, &record.Target); err != nil {
		return durableAgentSkillPreview{}, err
	}
	if err := decodeAgentSkillPreviewMap(resultJSON, &record.Result); err != nil {
		return durableAgentSkillPreview{}, err
	}
	if record.SecretRef != "" {
		secretEnvelope, err := profile.NewRepository(conn).GetSecret(ctx, record.ProfileID, record.SecretRef)
		if err != nil {
			return durableAgentSkillPreview{}, &agentSkillPreviewLifecycleError{
				Code:        "agent_skill_preview_secret_unavailable",
				Recoverable: true,
				NextAction:  "Create a fresh preview and provide the write-only credential again.",
			}
		}
		var secretPayload struct {
			Parameters   map[string]any `json:"parameters"`
			AgentContext map[string]any `json:"agent_context"`
		}
		if err := json.Unmarshal([]byte(secretEnvelope), &secretPayload); err != nil {
			return durableAgentSkillPreview{}, fmt.Errorf("decode pending Agent Skill preview secret: %w", err)
		}
		record.Parameters = nonNilAgentSkillPreviewMap(secretPayload.Parameters)
		record.AgentContext = nonNilAgentSkillPreviewMap(secretPayload.AgentContext)
	}
	return record, nil
}

func claimDurableAgentSkillPreview(ctx context.Context, conn *sql.DB, profileID, previewID string) (durableAgentSkillPreview, error) {
	record, err := getDurableAgentSkillPreview(ctx, conn, profileID, previewID)
	if err != nil {
		return durableAgentSkillPreview{}, err
	}
	now := time.Now().UTC()
	if previewExpired(record, now) {
		_, _ = conn.ExecContext(ctx, `
			UPDATE agent_skill_previews SET status = 'expired', error_code = 'agent_skill_preview_expired', secret_ref = ''
			WHERE id = ? AND profile_id = ? AND status = 'previewed'
		`, record.ID, record.ProfileID)
		deleteDurableAgentSkillPreviewSecret(ctx, conn, record)
		return durableAgentSkillPreview{}, agentSkillPreviewStatusError("expired")
	}
	if record.Status != "previewed" {
		return durableAgentSkillPreview{}, agentSkillPreviewStatusError(record.Status)
	}
	result, err := conn.ExecContext(ctx, `
		UPDATE agent_skill_previews SET status = 'applying', secret_ref = ''
		WHERE id = ? AND profile_id = ? AND status = 'previewed' AND expires_at > ?
	`, record.ID, record.ProfileID, now.Format(time.RFC3339Nano))
	if err != nil {
		return durableAgentSkillPreview{}, fmt.Errorf("claim Agent Skill preview: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return durableAgentSkillPreview{}, fmt.Errorf("inspect Agent Skill preview claim: %w", err)
	}
	if rows != 1 {
		latest, loadErr := getDurableAgentSkillPreview(ctx, conn, profileID, previewID)
		if loadErr != nil {
			return durableAgentSkillPreview{}, loadErr
		}
		if previewExpired(latest, now) {
			return durableAgentSkillPreview{}, agentSkillPreviewStatusError("expired")
		}
		return durableAgentSkillPreview{}, agentSkillPreviewStatusError(latest.Status)
	}
	deleteDurableAgentSkillPreviewSecret(ctx, conn, record)
	record.SecretRef = ""
	record.Status = "applying"
	return record, nil
}

func cancelDurableAgentSkillPreview(ctx context.Context, conn *sql.DB, profileID, previewID string) (durableAgentSkillPreview, error) {
	record, err := getDurableAgentSkillPreview(ctx, conn, profileID, previewID)
	if err != nil {
		return durableAgentSkillPreview{}, err
	}
	now := time.Now().UTC()
	if previewExpired(record, now) {
		_, _ = conn.ExecContext(ctx, `
			UPDATE agent_skill_previews SET status = 'expired', error_code = 'agent_skill_preview_expired', secret_ref = ''
			WHERE id = ? AND profile_id = ? AND status = 'previewed'
		`, record.ID, record.ProfileID)
		deleteDurableAgentSkillPreviewSecret(ctx, conn, record)
		return durableAgentSkillPreview{}, agentSkillPreviewStatusError("expired")
	}
	if record.Status != "previewed" {
		return durableAgentSkillPreview{}, agentSkillPreviewStatusError(record.Status)
	}
	result, err := conn.ExecContext(ctx, `
		UPDATE agent_skill_previews
		SET status = 'cancelled', cancelled_at = ?, error_code = '', secret_ref = ''
		WHERE id = ? AND profile_id = ? AND status = 'previewed' AND expires_at > ?
	`, now.Format(time.RFC3339Nano), record.ID, record.ProfileID, now.Format(time.RFC3339Nano))
	if err != nil {
		return durableAgentSkillPreview{}, fmt.Errorf("cancel Agent Skill preview: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return durableAgentSkillPreview{}, fmt.Errorf("inspect Agent Skill preview cancellation: %w", err)
	}
	if rows != 1 {
		latest, loadErr := getDurableAgentSkillPreview(ctx, conn, profileID, previewID)
		if loadErr != nil {
			return durableAgentSkillPreview{}, loadErr
		}
		return durableAgentSkillPreview{}, agentSkillPreviewStatusError(latest.Status)
	}
	record.Status = "cancelled"
	record.CancelledAt = now.Format(time.RFC3339Nano)
	deleteDurableAgentSkillPreviewSecret(ctx, conn, record)
	record.SecretRef = ""
	return record, nil
}

func completeDurableAgentSkillPreview(ctx context.Context, conn *sql.DB, record durableAgentSkillPreview, result map[string]any) error {
	boundedResult := boundedAgentSkillPreviewResult(result)
	resultJSON, err := json.Marshal(boundedResult)
	if err != nil {
		return fmt.Errorf("encode Agent Skill preview result: %w", err)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	updated, err := conn.ExecContext(ctx, `
		UPDATE agent_skill_previews
		SET status = 'applied', result_json = ?, error_code = '', applied_at = ?, secret_ref = ''
		WHERE id = ? AND profile_id = ? AND status = 'applying'
	`, string(resultJSON), now, record.ID, record.ProfileID)
	if err != nil {
		return fmt.Errorf("complete Agent Skill preview: %w", err)
	}
	rows, err := updated.RowsAffected()
	if err != nil {
		return fmt.Errorf("inspect Agent Skill preview completion: %w", err)
	}
	if rows != 1 {
		return fmt.Errorf("Agent Skill preview completion lost ownership")
	}
	deleteDurableAgentSkillPreviewSecret(ctx, conn, record)
	return nil
}

func failDurableAgentSkillPreview(ctx context.Context, conn *sql.DB, record durableAgentSkillPreview, blocker string) {
	blocker = strings.TrimSpace(blocker)
	if blocker == "" {
		blocker = "agent_skill_preview_apply_failed"
	}
	_, _ = conn.ExecContext(ctx, `
		UPDATE agent_skill_previews SET status = 'failed', error_code = ?, secret_ref = ''
		WHERE id = ? AND profile_id = ? AND status = 'applying'
	`, blocker, record.ID, record.ProfileID)
	deleteDurableAgentSkillPreviewSecret(ctx, conn, record)
}

func deleteDurableAgentSkillPreviewSecret(ctx context.Context, conn *sql.DB, record durableAgentSkillPreview) {
	if conn == nil || strings.TrimSpace(record.SecretRef) == "" {
		return
	}
	_ = profile.NewRepository(conn).DeleteSecret(ctx, record.ProfileID, record.SecretRef)
}

func cleanupExpiredDurableAgentSkillPreviews(ctx context.Context, conn *sql.DB, now string) error {
	durableAgentSkillPreviewMutationMu.Lock()
	defer durableAgentSkillPreviewMutationMu.Unlock()

	if conn == nil {
		return fmt.Errorf("agent skill preview store required")
	}
	now = strings.TrimSpace(now)
	if _, err := time.Parse(time.RFC3339Nano, now); err != nil {
		return fmt.Errorf("invalid Agent Skill preview cleanup time: %w", err)
	}
	rows, err := conn.QueryContext(ctx, `
		SELECT id, profile_id, secret_ref
		FROM agent_skill_previews
		WHERE status = 'previewed' AND expires_at <= ?
	`, now)
	if err != nil {
		return fmt.Errorf("list expired Agent Skill previews: %w", err)
	}
	type expiredPreview struct {
		ID        string
		ProfileID string
		SecretRef string
	}
	expired := []expiredPreview{}
	for rows.Next() {
		var candidate expiredPreview
		if err := rows.Scan(&candidate.ID, &candidate.ProfileID, &candidate.SecretRef); err != nil {
			_ = rows.Close()
			return fmt.Errorf("scan expired Agent Skill preview: %w", err)
		}
		expired = append(expired, candidate)
	}
	if err := rows.Close(); err != nil {
		return fmt.Errorf("close expired Agent Skill preview rows: %w", err)
	}
	for _, candidate := range expired {
		updated, err := conn.ExecContext(ctx, `
			UPDATE agent_skill_previews
			SET status = 'expired', error_code = 'agent_skill_preview_expired', secret_ref = ''
			WHERE id = ? AND profile_id = ? AND status = 'previewed' AND expires_at <= ?
		`, candidate.ID, candidate.ProfileID, now)
		if err != nil {
			return fmt.Errorf("expire Agent Skill preview: %w", err)
		}
		count, err := updated.RowsAffected()
		if err != nil {
			return fmt.Errorf("inspect expired Agent Skill preview update: %w", err)
		}
		if count == 1 {
			deleteDurableAgentSkillPreviewSecret(ctx, conn, durableAgentSkillPreview{
				ID: candidate.ID, ProfileID: candidate.ProfileID, SecretRef: candidate.SecretRef,
			})
		}
	}
	if _, err := conn.ExecContext(ctx, `
		UPDATE agent_skill_strong_confirmations
		SET status = 'expired'
		WHERE status = 'confirmed'
		  AND (expires_at <= ? OR preview_id IN (
			SELECT id FROM agent_skill_previews WHERE status = 'expired'
		  ))
	`, now); err != nil {
		return fmt.Errorf("expire Agent Skill strong confirmations: %w", err)
	}
	return nil
}

func runDurableAgentSkillPreviewCleanup(conn *sql.DB, stop <-chan struct{}, done chan<- struct{}) {
	defer close(done)
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			_ = cleanupExpiredDurableAgentSkillPreviews(context.Background(), conn, time.Now().UTC().Format(time.RFC3339Nano))
		case <-stop:
			return
		}
	}
}

func applyDurableAgentSkillPreview(ctx context.Context, conn *sql.DB, chatSvc *chat.Service, profiles *profile.Repository, backupSvc *backup.Service, registry agentskills.Registry, requested agentskills.PreviewRequest) (agentskills.PreviewResponse, error) {
	durableAgentSkillPreviewMutationMu.Lock()
	defer durableAgentSkillPreviewMutationMu.Unlock()

	record, err := getDurableAgentSkillPreview(ctx, conn, requested.ProfileID, requested.PreviewID)
	if err != nil {
		return agentskills.PreviewResponse{}, err
	}
	if !durableAgentSkillPreviewContextMatches(requested, record) {
		return agentskills.PreviewResponse{}, &agentSkillPreviewLifecycleError{
			Code:        "agent_skill_preview_context_mismatch",
			Recoverable: true,
			NextAction:  "Confirm this preview from the profile, thread, surface, and channel where Cabinet created it.",
		}
	}
	if record.Status != "previewed" {
		return agentskills.PreviewResponse{}, agentSkillPreviewStatusError(record.Status)
	}
	if previewExpired(record, time.Now().UTC()) {
		return agentskills.PreviewResponse{}, agentSkillPreviewStatusError("expired")
	}
	skill, skillAvailable := registry.Resolve(record.SkillID)
	if !skillAvailable {
		return agentskills.PreviewResponse{}, &agentSkillPreviewLifecycleError{
			Code: "agent_skill_preview_skill_unavailable", Recoverable: true,
			NextAction: "Create a fresh preview after the Agent Skill becomes available.",
		}
	}
	isDestructive := skill.SafetyLevel == agentskills.SafetyDestructive || skill.Permissions.Destructive
	var strongConfirmation agentStrongConfirmation
	if isDestructive {
		strongConfirmation, err = validateAgentStrongConfirmation(ctx, conn, backupSvc, record, requested.StrongConfirmationToken)
		if err != nil {
			return agentskills.PreviewResponse{}, err
		}
	}
	storedRequest := durableAgentSkillPreviewRequest(record, true)
	storedRequest, err = hydrateAgentUsersTargetFromServer(ctx, conn, storedRequest)
	if err != nil {
		return agentskills.PreviewResponse{}, err
	}
	if clarification, ok := agentSkillContextClarification(registry, storedRequest); ok {
		return agentskills.PreviewResponse{}, lifecycleValidationError(clarification)
	}
	if clarification, ok := agentSkillInboxNotificationContextClarification(ctx, conn, storedRequest); ok {
		return agentskills.PreviewResponse{}, lifecycleValidationError(clarification)
	}
	authority, err := reviewAgentSkillAuthorityWithStrongConfirmation(ctx, profiles, registry, storedRequest, "direct-api-preview-confirm", isDestructive)
	if err != nil {
		return agentskills.PreviewResponse{}, fmt.Errorf("review Agent Skill preview authority: %w", err)
	}
	if !authority.ApplyAllowed {
		return agentskills.PreviewResponse{}, &agentSkillPreviewLifecycleError{
			Code:        firstNonEmptyString(authority.Blocker, "agent_skill_preview_authority_denied"),
			Recoverable: true,
			NextAction:  authority.NextAction,
		}
	}
	preview, err := registry.Preview(storedRequest)
	if err != nil {
		return agentskills.PreviewResponse{}, &agentSkillPreviewLifecycleError{
			Code:        "agent_skill_preview_skill_unavailable",
			Recoverable: true,
			NextAction:  "Create a fresh preview after the Agent Skill becomes available.",
		}
	}
	if preview.Blocker != "" && preview.Blocker != "confirmation_required" {
		return agentskills.PreviewResponse{}, &agentSkillPreviewLifecycleError{Code: preview.Blocker, Recoverable: true, NextAction: preview.NextAction}
	}
	if !durableAgentSkillPreviewTargetMatches(preview, record) {
		return agentskills.PreviewResponse{}, &agentSkillPreviewLifecycleError{
			Code:        "agent_skill_preview_target_changed",
			Recoverable: true,
			NextAction:  "Create a fresh preview for the current target before confirming.",
		}
	}
	if isDestructive {
		if err := claimAgentStrongConfirmation(ctx, conn, strongConfirmation); err != nil {
			return agentskills.PreviewResponse{}, err
		}
	}
	claimed, err := claimDurableAgentSkillPreview(ctx, conn, record.ProfileID, record.ID)
	if err != nil {
		return agentskills.PreviewResponse{}, err
	}
	var result map[string]any
	var blocker string
	if storedRequest.SkillID == "cabinet.data.restore_backup" {
		result, blocker, err = applyAgentDataRestoreBackupWithService(backupSvc, storedRequest.ProfileID, storedRequest.Parameters)
	} else {
		result, blocker, err = applyAgentSkill(ctx, conn, chatSvc, storedRequest.SkillID, storedRequest.ProfileID, storedRequest.Parameters)
	}
	if err != nil {
		failDurableAgentSkillPreview(ctx, conn, claimed, blocker)
		return agentskills.PreviewResponse{}, &agentSkillPreviewLifecycleError{
			Code:        firstNonEmptyString(strings.TrimSpace(blocker), "agent_skill_preview_apply_failed"),
			Recoverable: true,
			NextAction:  "Review the current target and create a fresh preview before retrying.",
		}
	}
	if storedRequest.SkillID == "cabinet.data.restore_backup" {
		if err := restoreDurableAgentSkillReceiptAfterDataRestore(ctx, conn, claimed, strongConfirmation); err != nil {
			return agentskills.PreviewResponse{}, err
		}
	}
	boundedResult := boundedAgentSkillPreviewResult(result)
	if err := completeDurableAgentSkillPreview(ctx, conn, claimed, boundedResult); err != nil {
		return agentskills.PreviewResponse{}, err
	}
	if err := appendAgentAuthorityDecisionAudit(ctx, profiles, authority, agentSkillAuthorityRequest(storedRequest), "applied"); err != nil {
		return agentskills.PreviewResponse{}, fmt.Errorf("record Agent Skill preview authority outcome: %w", err)
	}
	claimed.Status = "applied"
	preview = bindDurableAgentSkillPreviewResponse(preview, claimed)
	preview.Allowed = true
	preview.PreviewOnly = false
	preview.MutationApplied = true
	preview.Blocker = ""
	preview.NextAction = ""
	preview.Target = boundedResult
	if _, err := recordDirectAgentSkillWorkflowRun(ctx, chatSvc, "agent-skill-durable-preview-apply", storedRequest, authority, preview, boundedResult); err != nil {
		return agentskills.PreviewResponse{}, fmt.Errorf("record Agent Skill durable apply timeline: %w", err)
	}
	return preview, nil
}

func restoreDurableAgentSkillReceiptAfterDataRestore(ctx context.Context, conn *sql.DB, record durableAgentSkillPreview, confirmation agentStrongConfirmation) error {
	parameters := record.Parameters
	agentContext := record.AgentContext
	if containsSensitiveAgentSkillPreviewValue(parameters) {
		parameters = redactedAgentSkillPreviewMap(parameters)
	}
	if containsSensitiveAgentSkillPreviewValue(agentContext) {
		agentContext = redactedAgentSkillPreviewMap(agentContext)
	}
	parametersJSON, err := json.Marshal(nonNilAgentSkillPreviewMap(parameters))
	if err != nil {
		return fmt.Errorf("encode restored Agent Skill receipt parameters: %w", err)
	}
	agentContextJSON, err := json.Marshal(nonNilAgentSkillPreviewMap(agentContext))
	if err != nil {
		return fmt.Errorf("encode restored Agent Skill receipt context: %w", err)
	}
	targetJSON, err := json.Marshal(nonNilAgentSkillPreviewMap(record.Target))
	if err != nil {
		return fmt.Errorf("encode restored Agent Skill receipt target: %w", err)
	}
	if _, err := conn.ExecContext(ctx, `
		INSERT INTO agent_skill_previews(
			id, profile_id, skill_id, source_surface, source_channel, source_thread_id, source_message_id,
			agent_context_json, parameters_json, target_json, result_json, secret_ref, status, error_code, created_at, expires_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, '{}', '', 'applying', '', ?, ?)
		ON CONFLICT(id) DO UPDATE SET status = 'applying', error_code = '', secret_ref = ''
	`, record.ID, record.ProfileID, record.SkillID, record.SourceSurface, record.SourceChannel, record.SourceThreadID,
		record.SourceMessageID, string(agentContextJSON), string(parametersJSON), string(targetJSON), record.CreatedAt, record.ExpiresAt); err != nil {
		return fmt.Errorf("restore Agent Skill preview receipt after data restore: %w", err)
	}
	if confirmation.ID != "" {
		if _, err := conn.ExecContext(ctx, `
			INSERT INTO agent_skill_strong_confirmations(id, preview_id, profile_id, skill_id, target_fingerprint, token_hash, status, created_at, expires_at, used_at)
			VALUES (?, ?, ?, ?, ?, ?, 'used', ?, ?, ?)
			ON CONFLICT(id) DO UPDATE SET status = 'used', used_at = excluded.used_at
		`, confirmation.ID, confirmation.PreviewID, confirmation.ProfileID, confirmation.SkillID, confirmation.TargetFingerprint,
			confirmation.TokenHash, confirmation.CreatedAt, confirmation.ExpiresAt, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
			return fmt.Errorf("restore strong-confirmation receipt after data restore: %w", err)
		}
	}
	return nil
}

func readDurableAgentSkillPreviewResponse(ctx context.Context, conn *sql.DB, registry agentskills.Registry, profileID, previewID string) (agentskills.PreviewResponse, error) {
	durableAgentSkillPreviewMutationMu.Lock()
	defer durableAgentSkillPreviewMutationMu.Unlock()

	record, err := getDurableAgentSkillPreview(ctx, conn, profileID, previewID)
	if err != nil {
		return agentskills.PreviewResponse{}, err
	}
	if record.Status == "previewed" && previewExpired(record, time.Now().UTC()) {
		_, _ = conn.ExecContext(ctx, `
			UPDATE agent_skill_previews
			SET status = 'expired', error_code = 'agent_skill_preview_expired', secret_ref = ''
			WHERE id = ? AND profile_id = ? AND status = 'previewed'
		`, record.ID, record.ProfileID)
		deleteDurableAgentSkillPreviewSecret(ctx, conn, record)
		record.Status = "expired"
		record.SecretRef = ""
	}
	storedRequest := durableAgentSkillPreviewRequest(record, false)
	preview, err := registry.Preview(storedRequest)
	if err != nil {
		preview = agentskills.PreviewResponse{
			SkillID:              record.SkillID,
			PreviewOnly:          true,
			ConfirmationRequired: true,
		}
	}
	preview = bindDurableAgentSkillPreviewResponse(preview, record)
	preview.Target = record.Target
	preview.MutationApplied = record.Status == "applied"
	preview.PreviewOnly = record.Status != "applied"
	if record.Status == "applied" && len(record.Result) > 0 {
		preview.Allowed = true
		preview.Target = record.Result
		preview.Blocker = ""
		preview.NextAction = ""
	} else if record.Status != "previewed" {
		statusErr, _ := agentSkillPreviewStatusError(record.Status).(*agentSkillPreviewLifecycleError)
		if statusErr != nil {
			preview.Allowed = false
			preview.Blocker = statusErr.Code
			preview.NextAction = statusErr.NextAction
		}
	}
	return preview, nil
}

func cancelDurableAgentSkillPreviewResponse(ctx context.Context, conn *sql.DB, chatSvc *chat.Service, registry agentskills.Registry, requested agentskills.PreviewRequest) (agentskills.PreviewResponse, error) {
	durableAgentSkillPreviewMutationMu.Lock()
	defer durableAgentSkillPreviewMutationMu.Unlock()

	record, err := getDurableAgentSkillPreview(ctx, conn, requested.ProfileID, requested.PreviewID)
	if err != nil {
		return agentskills.PreviewResponse{}, err
	}
	if !durableAgentSkillPreviewContextMatches(requested, record) {
		return agentskills.PreviewResponse{}, &agentSkillPreviewLifecycleError{
			Code:        "agent_skill_preview_context_mismatch",
			Recoverable: true,
			NextAction:  "Cancel this preview from the profile, thread, surface, and channel where Cabinet created it.",
		}
	}
	record, err = cancelDurableAgentSkillPreview(ctx, conn, record.ProfileID, record.ID)
	if err != nil {
		return agentskills.PreviewResponse{}, err
	}
	_, _ = conn.ExecContext(ctx, `UPDATE agent_skill_strong_confirmations SET status = 'cancelled' WHERE preview_id = ? AND profile_id = ? AND status = 'confirmed'`, record.ID, record.ProfileID)
	storedRequest := durableAgentSkillPreviewRequest(record, false)
	preview, err := registry.Preview(storedRequest)
	if err != nil {
		preview = agentskills.PreviewResponse{
			SkillID:              record.SkillID,
			PreviewOnly:          true,
			MutationApplied:      false,
			ConfirmationRequired: true,
			Target:               record.Target,
		}
	}
	preview = bindDurableAgentSkillPreviewResponse(preview, record)
	preview.Allowed = false
	preview.PreviewOnly = true
	preview.MutationApplied = false
	preview.Blocker = "agent_skill_preview_cancelled"
	preview.NextAction = "Create a new preview if you still want Cabinet to apply this change."
	authority := agentskills.AgentAuthorityReview{
		ProfileID:            record.ProfileID,
		EntryPoint:           "direct-api-preview-cancel",
		SkillID:              record.SkillID,
		Decision:             "cancelled",
		ConfirmationRequired: true,
	}
	if _, err := recordDirectAgentSkillWorkflowRun(ctx, chatSvc, "agent-skill-durable-preview-cancel", storedRequest, authority, preview, record.Target); err != nil {
		return agentskills.PreviewResponse{}, fmt.Errorf("record Agent Skill durable cancel timeline: %w", err)
	}
	return preview, nil
}

func lifecycleValidationError(clarification map[string]any) error {
	code := strings.TrimSpace(fmt.Sprint(clarification["error"]))
	if code == "" || code == "<nil>" {
		code = strings.TrimSpace(fmt.Sprint(clarification["blocker"]))
	}
	if code == "" || code == "<nil>" {
		code = "agent_skill_preview_target_unavailable"
	}
	return &agentSkillPreviewLifecycleError{
		Code:        code,
		Recoverable: true,
		NextAction:  strings.TrimSpace(fmt.Sprint(clarification["next_action"])),
	}
}

func writeAgentSkillPreviewLifecycleError(w http.ResponseWriter, err error) {
	lifecycleErr, ok := err.(*agentSkillPreviewLifecycleError)
	if !ok {
		log.Printf("Agent Skill preview lifecycle failed: %T: %v", err, err)
		http.Error(w, `{"error":"agent_skill_preview_lifecycle_failed"}`, http.StatusInternalServerError)
		return
	}
	status := http.StatusConflict
	if lifecycleErr.Code == "agent_skill_preview_not_found" {
		status = http.StatusNotFound
	}
	if lifecycleErr.Code == "agent_skill_preview_sensitive_parameters_not_allowed" {
		status = http.StatusBadRequest
	}
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error":       lifecycleErr.Code,
		"recoverable": lifecycleErr.Recoverable,
		"next_action": lifecycleErr.NextAction,
	})
}

func durableAgentSkillPreviewRequest(record durableAgentSkillPreview, confirmed bool) agentskills.PreviewRequest {
	return agentskills.PreviewRequest{
		PreviewID:       record.ID,
		SkillID:         record.SkillID,
		ProfileID:       record.ProfileID,
		Confirm:         confirmed,
		SourceSurface:   record.SourceSurface,
		SourceChannel:   record.SourceChannel,
		SourceThreadID:  record.SourceThreadID,
		SourceMessageID: record.SourceMessageID,
		AgentContext:    record.AgentContext,
		Parameters:      record.Parameters,
	}
}

func bindDurableAgentSkillPreviewResponse(preview agentskills.PreviewResponse, record durableAgentSkillPreview) agentskills.PreviewResponse {
	preview.PreviewID = record.ID
	preview.PreviewStatus = record.Status
	preview.ExpiresAt = record.ExpiresAt
	preview.SourceSurface = record.SourceSurface
	preview.SourceChannel = record.SourceChannel
	preview.SourceThreadID = record.SourceThreadID
	preview.SourceMessageID = record.SourceMessageID
	if preview.SafetyLevel == agentskills.SafetyDestructive {
		preview.StrongConfirmationRequired = true
		preview.StrongConfirmationEndpoint = "/api/agent/skills/confirm-destructive"
	}
	return preview
}

func durableAgentSkillPreviewContextMatches(req agentskills.PreviewRequest, record durableAgentSkillPreview) bool {
	checks := [][2]string{
		{req.SkillID, record.SkillID},
		{req.SourceSurface, record.SourceSurface},
		{req.SourceChannel, record.SourceChannel},
		{req.SourceThreadID, record.SourceThreadID},
		{req.SourceMessageID, record.SourceMessageID},
	}
	for _, check := range checks {
		if strings.TrimSpace(check[0]) != "" && strings.TrimSpace(check[0]) != strings.TrimSpace(check[1]) {
			return false
		}
	}
	return true
}

func durableAgentSkillPreviewTargetMatches(preview agentskills.PreviewResponse, record durableAgentSkillPreview) bool {
	want, err := json.Marshal(nonNilAgentSkillPreviewMap(record.Target))
	if err != nil {
		return false
	}
	got, err := json.Marshal(nonNilAgentSkillPreviewMap(preview.Target))
	return err == nil && string(got) == string(want)
}

func previewExpired(record durableAgentSkillPreview, now time.Time) bool {
	expiresAt, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(record.ExpiresAt))
	return err != nil || !expiresAt.After(now)
}

func agentSkillPreviewStatusError(status string) error {
	switch strings.TrimSpace(status) {
	case "not_found":
		return &agentSkillPreviewLifecycleError{Code: "agent_skill_preview_not_found"}
	case "cancelled":
		return &agentSkillPreviewLifecycleError{Code: "agent_skill_preview_cancelled", Recoverable: true, NextAction: "Create a new preview if you still want Cabinet to apply this change."}
	case "expired":
		return &agentSkillPreviewLifecycleError{Code: "agent_skill_preview_expired", Recoverable: true, NextAction: "Create a fresh preview and review the current target before confirming."}
	case "applied":
		return &agentSkillPreviewLifecycleError{Code: "agent_skill_preview_already_applied"}
	case "applying":
		return &agentSkillPreviewLifecycleError{Code: "agent_skill_preview_apply_in_progress", Recoverable: true, NextAction: "Refresh the Agent thread to inspect the final action state."}
	case "failed":
		return &agentSkillPreviewLifecycleError{Code: "agent_skill_preview_failed", Recoverable: true, NextAction: "Review the target and create a fresh preview before retrying."}
	default:
		return &agentSkillPreviewLifecycleError{Code: "agent_skill_preview_unavailable", Recoverable: true, NextAction: "Create a fresh preview before retrying."}
	}
}

func containsSensitiveAgentSkillPreviewValue(value any) bool {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			if isSensitiveAgentSkillPreviewKey(key) {
				return true
			}
			if containsSensitiveAgentSkillPreviewValue(child) {
				return true
			}
		}
	case []any:
		for _, child := range typed {
			if containsSensitiveAgentSkillPreviewValue(child) {
				return true
			}
		}
	}
	return false
}

func redactedAgentSkillPreviewMap(value map[string]any) map[string]any {
	redacted, _ := redactAgentSkillPreviewValue(nonNilAgentSkillPreviewMap(value)).(map[string]any)
	return nonNilAgentSkillPreviewMap(redacted)
}

func redactAgentSkillPreviewValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(typed))
		for key, child := range typed {
			if isSensitiveAgentSkillPreviewKey(key) {
				out[key] = "[redacted]"
				continue
			}
			out[key] = redactAgentSkillPreviewValue(child)
		}
		return out
	case []any:
		out := make([]any, len(typed))
		for index, child := range typed {
			out[index] = redactAgentSkillPreviewValue(child)
		}
		return out
	default:
		return value
	}
}

func isSensitiveAgentSkillPreviewKey(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	return strings.Contains(normalized, "secret") || strings.Contains(normalized, "token") ||
		strings.Contains(normalized, "api_key") || strings.Contains(normalized, "apikey") ||
		strings.Contains(normalized, "password") || strings.Contains(normalized, "authorization") ||
		strings.Contains(normalized, "credential")
}

func boundedAgentSkillPreviewResult(result map[string]any) map[string]any {
	bounded := map[string]any{}
	for key, value := range result {
		normalized := strings.ToLower(strings.TrimSpace(key))
		if normalized == "" || containsSensitiveAgentSkillPreviewValue(map[string]any{normalized: value}) {
			continue
		}
		if isBoundedAgentSkillPreviewResultKey(normalized) {
			switch value.(type) {
			case string, bool, float64, float32, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, nil:
				bounded[key] = value
			}
		}
	}
	return bounded
}

func isBoundedAgentSkillPreviewResultKey(key string) bool {
	for _, exact := range []string{
		"operation", "profile_id", "item_id", "wishlist_id", "wishlist_entry_id", "collection_id", "collection_name",
		"media_id", "discovery_id", "result_id", "watch_id", "order_id", "line_item_id", "provider_id", "user_id",
		"removed_user_id", "setting_key", "part_number", "title", "status", "route", "next_action", "mutation_applied",
		"destructive_confirmation", "restore_drill_verified", "profile_isolated", "profile_scope", "integrity_check",
		"selected_backup_redacted", "raw_payload_redacted", "restored_manifest_sha256", "restored_manifest_bytes",
		"pre_restore_backup_taken", "restore_recovery_available",
	} {
		if key == exact {
			return true
		}
	}
	for _, suffix := range []string{"_persisted", "_applied", "_created", "_updated", "_deleted", "_restored", "_linked", "_detached", "_preserved", "_count"} {
		if strings.HasSuffix(key, suffix) {
			return true
		}
	}
	return false
}

func nonNilAgentSkillPreviewMap(value map[string]any) map[string]any {
	if value == nil {
		return map[string]any{}
	}
	return value
}

func decodeAgentSkillPreviewMap(raw string, target *map[string]any) error {
	*target = map[string]any{}
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	if err := json.Unmarshal([]byte(raw), target); err != nil {
		return fmt.Errorf("decode Agent Skill preview record: %w", err)
	}
	return nil
}
