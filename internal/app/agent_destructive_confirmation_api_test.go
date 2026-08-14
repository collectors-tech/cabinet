package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func previewAndConfirmDestructiveAgentSkill(t *testing.T, a *App, profileID, skillID string, parameters map[string]any, headers map[string]string) (agentSkillLifecyclePreviewPayload, agentStrongConfirmationPayload) {
	t.Helper()
	body, err := json.Marshal(map[string]any{"profile_id": profileID, "skill_id": skillID, "parameters": parameters})
	if err != nil {
		t.Fatalf("marshal destructive preview: %v", err)
	}
	previewResponse := doRequest(t, a, http.MethodPost, "/api/agent/skills/preview", strings.NewReader(string(body)), headers)
	if previewResponse.Code != http.StatusOK {
		t.Fatalf("destructive preview status=%d body=%s", previewResponse.Code, previewResponse.Body.String())
	}
	var preview agentSkillLifecyclePreviewPayload
	if err := json.NewDecoder(previewResponse.Body).Decode(&preview); err != nil || preview.PreviewID == "" {
		t.Fatalf("decode destructive preview: err=%v payload=%+v", err, preview)
	}
	confirmationResponse := doRequest(t, a, http.MethodPost, "/api/agent/skills/confirm-destructive", strings.NewReader(`{"profile_id":"`+profileID+`","preview_id":"`+preview.PreviewID+`"}`), headers)
	if confirmationResponse.Code != http.StatusOK {
		t.Fatalf("issue destructive confirmation status=%d body=%s", confirmationResponse.Code, confirmationResponse.Body.String())
	}
	var confirmation agentStrongConfirmationPayload
	if err := json.NewDecoder(confirmationResponse.Body).Decode(&confirmation); err != nil || confirmation.ConfirmationToken == "" {
		t.Fatalf("decode destructive confirmation: err=%v payload=%+v", err, confirmation)
	}
	return preview, confirmation
}

type agentStrongConfirmationPayload struct {
	ConfirmationToken string         `json:"confirmation_token"`
	ExpiresAt         string         `json:"expires_at"`
	Action            string         `json:"action"`
	Target            map[string]any `json:"target"`
	Impact            []string       `json:"impact"`
}

func applyStrongConfirmedAgentSkill(t *testing.T, a *App, profileID, skillID string, parameters map[string]any, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	previewBody, err := json.Marshal(map[string]any{"profile_id": profileID, "skill_id": skillID, "parameters": parameters})
	if err != nil {
		t.Fatalf("marshal destructive preview request: %v", err)
	}
	previewResponse := doRequest(t, a, http.MethodPost, "/api/agent/skills/preview", strings.NewReader(string(previewBody)), headers)
	if previewResponse.Code != http.StatusOK {
		t.Fatalf("destructive preview status=%d body=%s", previewResponse.Code, previewResponse.Body.String())
	}
	var preview agentSkillLifecyclePreviewPayload
	if err := json.NewDecoder(previewResponse.Body).Decode(&preview); err != nil || preview.PreviewID == "" {
		t.Fatalf("decode destructive preview: err=%v payload=%+v", err, preview)
	}
	confirmationResponse := doRequest(t, a, http.MethodPost, "/api/agent/skills/confirm-destructive", strings.NewReader(`{"profile_id":"`+profileID+`","preview_id":"`+preview.PreviewID+`"}`), headers)
	if confirmationResponse.Code != http.StatusOK {
		t.Fatalf("issue destructive confirmation status=%d body=%s", confirmationResponse.Code, confirmationResponse.Body.String())
	}
	var confirmation agentStrongConfirmationPayload
	if err := json.NewDecoder(confirmationResponse.Body).Decode(&confirmation); err != nil || confirmation.ConfirmationToken == "" {
		t.Fatalf("decode destructive confirmation: err=%v payload=%+v", err, confirmation)
	}
	return doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{"profile_id":"`+profileID+`","preview_id":"`+preview.PreviewID+`","confirm":true,"strong_confirmation_token":"`+confirmation.ConfirmationToken+`"}`), headers)
}

func TestDestructiveAgentSkillRequiresSingleUseServerConfirmation(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	profileID := createAgentSkillLifecycleProfile(t, a, "Destructive Agent Confirmation")
	backupRun, err := a.backupSvc.CreateBackup(context.Background())
	if err != nil {
		t.Fatalf("create destructive confirmation backup: %v", err)
	}
	previewRequest := `{
		"profile_id":"` + profileID + `",
		"skill_id":"cabinet.data.restore_backup",
		"source_surface":"chats.main",
		"source_channel":"in-app",
		"parameters":{"backup_path":"` + strings.ReplaceAll(backupRun.Path, `\`, `\\`) + `"}
	}`
	previewResponse := doRequest(t, a, http.MethodPost, "/api/agent/skills/preview", strings.NewReader(previewRequest), map[string]string{"Content-Type": "application/json"})
	if previewResponse.Code != http.StatusOK {
		t.Fatalf("destructive preview status=%d body=%s", previewResponse.Code, previewResponse.Body.String())
	}
	var preview agentSkillLifecyclePreviewPayload
	if err := json.NewDecoder(previewResponse.Body).Decode(&preview); err != nil {
		t.Fatalf("decode destructive preview: %v", err)
	}
	if preview.PreviewID == "" || preview.Status != "previewed" || preview.MutationApplied {
		t.Fatalf("expected durable non-mutating destructive preview, got %+v", preview)
	}

	ordinary := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+profileID+`",
		"preview_id":"`+preview.PreviewID+`",
		"confirm":true
	}`), map[string]string{"Content-Type": "application/json"})
	if ordinary.Code != http.StatusConflict || !strings.Contains(ordinary.Body.String(), `"error":"strong_confirmation_required"`) {
		t.Fatalf("ordinary confirmation must not authorize destructive apply, status=%d body=%s", ordinary.Code, ordinary.Body.String())
	}

	confirmResponse := doRequest(t, a, http.MethodPost, "/api/agent/skills/confirm-destructive", strings.NewReader(`{
		"profile_id":"`+profileID+`",
		"preview_id":"`+preview.PreviewID+`"
	}`), map[string]string{"Content-Type": "application/json"})
	if confirmResponse.Code != http.StatusOK {
		t.Fatalf("strong confirmation status=%d body=%s", confirmResponse.Code, confirmResponse.Body.String())
	}
	var confirmation agentStrongConfirmationPayload
	if err := json.NewDecoder(confirmResponse.Body).Decode(&confirmation); err != nil {
		t.Fatalf("decode strong confirmation: %v", err)
	}
	if confirmation.ConfirmationToken == "" || confirmation.ExpiresAt == "" || confirmation.Action != "restore_backup" || len(confirmation.Impact) == 0 {
		t.Fatalf("expected action-specific impact-bound confirmation, got %+v", confirmation)
	}
	if confirmation.Target["backup_file"] == nil || confirmation.Target["database_sha256"] == nil || confirmation.Target["compatible"] != true {
		t.Fatalf("expected selected restore target, got %+v", confirmation.Target)
	}
	var confirmationAudit string
	if err := a.db.QueryRow(`SELECT after_json FROM audit_events WHERE entity_id = ? AND entity_type = 'profile_agent_authority_decision' AND after_json LIKE '%strong_confirmation_issued%' ORDER BY created_at DESC LIMIT 1`, profileID).Scan(&confirmationAudit); err != nil {
		t.Fatalf("read strong confirmation issuance audit: %v", err)
	}
	if strings.Contains(confirmationAudit, confirmation.ConfirmationToken) || !strings.Contains(confirmationAudit, `"outcome":"strong_confirmation_issued"`) {
		t.Fatalf("strong confirmation audit must be explicit and token-free: %s", confirmationAudit)
	}

	apply := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+profileID+`",
		"preview_id":"`+preview.PreviewID+`",
		"confirm":true,
		"strong_confirmation_token":"`+confirmation.ConfirmationToken+`"
	}`), map[string]string{"Content-Type": "application/json"})
	if apply.Code != http.StatusOK || !strings.Contains(apply.Body.String(), `"mutation_applied":true`) {
		t.Fatalf("strong-confirmed apply status=%d body=%s", apply.Code, apply.Body.String())
	}

	replay := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+profileID+`",
		"preview_id":"`+preview.PreviewID+`",
		"confirm":true,
		"strong_confirmation_token":"`+confirmation.ConfirmationToken+`"
	}`), map[string]string{"Content-Type": "application/json"})
	if replay.Code != http.StatusConflict || !strings.Contains(replay.Body.String(), `"error":"agent_skill_preview_already_applied"`) {
		t.Fatalf("destructive replay must fail closed, status=%d body=%s", replay.Code, replay.Body.String())
	}
}

func TestDestructiveAgentSkillRejectsDirectBooleanConfirmation(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	profileID := createAgentSkillLifecycleProfile(t, a, "No Destructive Boolean Bypass")
	response := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+profileID+`",
		"skill_id":"cabinet.data.restore_backup",
		"confirm":true,
		"parameters":{"backup_path":"backups/bypass.zip"}
	}`), map[string]string{"Content-Type": "application/json"})
	if response.Code != http.StatusConflict || !strings.Contains(response.Body.String(), `"error":"agent_destructive_preview_required"`) {
		t.Fatalf("direct destructive boolean confirmation bypassed durable review, status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestDestructiveAgentSkillConfirmationFailsClosedAcrossLifecycleBoundaries(t *testing.T) {
	a := newTestApp(t)
	profileID := createAgentSkillLifecycleProfile(t, a, "Destructive Lifecycle Boundaries")
	headers := agentAdminTestHeaders(t, a, profileID)
	target, err := createRuntimeUser(context.Background(), a.db, profileID, "Lifecycle", "Target", "lifecycle-target", "lifecycle.target@example.test", "+61 400 000 208", "view")
	if err != nil {
		t.Fatalf("create destructive lifecycle target: %v", err)
	}

	preview, firstConfirmation := previewAndConfirmDestructiveAgentSkill(t, a, profileID, "cabinet.users.remove_user", map[string]any{"target_user": target.ID}, headers)
	_, secondConfirmation := previewAndConfirmExistingDestructiveAgentSkill(t, a, profileID, preview.PreviewID, headers)
	if firstConfirmation.ConfirmationToken == secondConfirmation.ConfirmationToken {
		t.Fatal("fresh strong confirmation reused its bearer token")
	}
	firstHash := sha256.Sum256([]byte(firstConfirmation.ConfirmationToken))
	var storedHash, status string
	if err := a.db.QueryRow(`SELECT token_hash, status FROM agent_skill_strong_confirmations WHERE preview_id = ? ORDER BY created_at ASC LIMIT 1`, preview.PreviewID).Scan(&storedHash, &status); err != nil {
		t.Fatalf("read persisted confirmation hash: %v", err)
	}
	if storedHash == firstConfirmation.ConfirmationToken || storedHash != hex.EncodeToString(firstHash[:]) || status != "superseded" {
		t.Fatalf("confirmation persistence was not hashed/superseded: hash=%q status=%q", storedHash, status)
	}

	stale := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{"profile_id":"`+profileID+`","preview_id":"`+preview.PreviewID+`","confirm":true,"strong_confirmation_token":"`+firstConfirmation.ConfirmationToken+`"}`), headers)
	if stale.Code != http.StatusConflict || !strings.Contains(stale.Body.String(), `"error":"strong_confirmation_invalid"`) {
		t.Fatalf("superseded confirmation must fail closed: status=%d body=%s", stale.Code, stale.Body.String())
	}

	if _, err := updateRuntimeUser(context.Background(), a.db, profileID, target.ID, target.FirstName, target.LastName, target.Username, target.Email, target.PhoneNumber, "active", "admin"); err != nil {
		t.Fatalf("change destructive target after confirmation: %v", err)
	}
	changed := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{"profile_id":"`+profileID+`","preview_id":"`+preview.PreviewID+`","confirm":true,"strong_confirmation_token":"`+secondConfirmation.ConfirmationToken+`"}`), headers)
	if changed.Code != http.StatusConflict || !strings.Contains(changed.Body.String(), `"error":"strong_confirmation_target_changed"`) {
		t.Fatalf("changed target must require a fresh confirmation: status=%d body=%s", changed.Code, changed.Body.String())
	}

	if _, err := a.db.Exec(`UPDATE agent_skill_strong_confirmations SET expires_at = '2000-01-01T00:00:00Z' WHERE preview_id = ? AND status = 'confirmed'`, preview.PreviewID); err != nil {
		t.Fatalf("expire destructive confirmation: %v", err)
	}
	expired := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{"profile_id":"`+profileID+`","preview_id":"`+preview.PreviewID+`","confirm":true,"strong_confirmation_token":"`+secondConfirmation.ConfirmationToken+`"}`), headers)
	if expired.Code != http.StatusConflict || !strings.Contains(expired.Body.String(), `"error":"strong_confirmation_expired"`) {
		t.Fatalf("expired confirmation must fail closed: status=%d body=%s", expired.Code, expired.Body.String())
	}

	cancelPreview, _ := previewAndConfirmDestructiveAgentSkill(t, a, profileID, "cabinet.users.remove_user", map[string]any{"target_user": target.ID}, headers)
	cancelled := doRequest(t, a, http.MethodPost, "/api/agent/skills/cancel", strings.NewReader(`{"profile_id":"`+profileID+`","preview_id":"`+cancelPreview.PreviewID+`"}`), headers)
	if cancelled.Code != http.StatusOK {
		t.Fatalf("cancel destructive preview status=%d body=%s", cancelled.Code, cancelled.Body.String())
	}
	if err := a.db.QueryRow(`SELECT status FROM agent_skill_strong_confirmations WHERE preview_id = ? ORDER BY created_at DESC LIMIT 1`, cancelPreview.PreviewID).Scan(&status); err != nil || status != "cancelled" {
		t.Fatalf("cancel did not revoke strong confirmation: status=%q err=%v", status, err)
	}

	otherProfileID := createAgentSkillLifecycleProfile(t, a, "Destructive Cross Profile")
	otherHeaders := agentAdminTestHeaders(t, a, otherProfileID)
	crossProfile := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{"profile_id":"`+otherProfileID+`","preview_id":"`+preview.PreviewID+`","confirm":true,"strong_confirmation_token":"`+secondConfirmation.ConfirmationToken+`"}`), otherHeaders)
	if crossProfile.Code != http.StatusNotFound || !strings.Contains(crossProfile.Body.String(), `"error":"agent_skill_preview_not_found"`) {
		t.Fatalf("cross-profile confirmation must fail closed: status=%d body=%s", crossProfile.Code, crossProfile.Body.String())
	}
}

func previewAndConfirmExistingDestructiveAgentSkill(t *testing.T, a *App, profileID, previewID string, headers map[string]string) (agentSkillLifecyclePreviewPayload, agentStrongConfirmationPayload) {
	t.Helper()
	var preview agentSkillLifecyclePreviewPayload
	previewResponse := doRequest(t, a, http.MethodGet, "/api/agent/skills/preview?profile_id="+profileID+"&preview_id="+previewID, nil, headers)
	if previewResponse.Code != http.StatusOK {
		t.Fatalf("read destructive preview status=%d body=%s", previewResponse.Code, previewResponse.Body.String())
	}
	if err := json.NewDecoder(previewResponse.Body).Decode(&preview); err != nil {
		t.Fatalf("decode existing destructive preview: %v", err)
	}
	confirmationResponse := doRequest(t, a, http.MethodPost, "/api/agent/skills/confirm-destructive", strings.NewReader(`{"profile_id":"`+profileID+`","preview_id":"`+previewID+`"}`), headers)
	if confirmationResponse.Code != http.StatusOK {
		t.Fatalf("reissue destructive confirmation status=%d body=%s", confirmationResponse.Code, confirmationResponse.Body.String())
	}
	var confirmation agentStrongConfirmationPayload
	if err := json.NewDecoder(confirmationResponse.Body).Decode(&confirmation); err != nil {
		t.Fatalf("decode reissued destructive confirmation: %v", err)
	}
	return preview, confirmation
}
