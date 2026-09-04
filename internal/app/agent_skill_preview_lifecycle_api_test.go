package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/collectors-tech/cabinet/internal/config"
	"github.com/collectors-tech/cabinet/internal/profile"
	"github.com/collectors-tech/cabinet/internal/update"
)

type agentSkillLifecyclePreviewPayload struct {
	PreviewID            string         `json:"preview_id"`
	SkillID              string         `json:"skill_id"`
	Status               string         `json:"preview_status"`
	ExpiresAt            string         `json:"expires_at"`
	MutationApplied      bool           `json:"mutation_applied"`
	ConfirmationRequired bool           `json:"confirmation_required"`
	SourceSurface        string         `json:"source_surface"`
	SourceChannel        string         `json:"source_channel"`
	SourceThreadID       string         `json:"source_thread_id"`
	SourceMessageID      string         `json:"source_message_id"`
	Target               map[string]any `json:"target"`
}

func TestGenericAgentSkillPreviewConfirmsByOpaqueIDExactlyOnce(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	profileID := createAgentSkillLifecycleProfile(t, a, "Durable Agent Confirm")
	preview := previewAgentSkillWishlistCreate(t, a, profileID, "WISH-2087-CONFIRM")

	if preview.PreviewID == "" || strings.Contains(preview.PreviewID, profileID) || strings.Contains(preview.PreviewID, "wishlist") {
		t.Fatalf("expected opaque server-owned preview id, got %q", preview.PreviewID)
	}
	if preview.Status != "previewed" || preview.ExpiresAt == "" || !preview.ConfirmationRequired || preview.MutationApplied {
		t.Fatalf("expected pending non-mutating durable preview, got %+v", preview)
	}
	if countAgentSkillLifecycleWishlistItems(t, a, profileID, "WISH-2087-CONFIRM") != 0 {
		t.Fatal("preview mutated Wishlist before confirmation")
	}

	apply := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+profileID+`",
		"preview_id":"`+preview.PreviewID+`",
		"confirm":true
	}`), map[string]string{"Content-Type": "application/json"})
	if apply.Code != http.StatusOK {
		t.Fatalf("apply status=%d body=%s", apply.Code, apply.Body.String())
	}
	var applied agentSkillLifecyclePreviewPayload
	if err := json.NewDecoder(apply.Body).Decode(&applied); err != nil {
		t.Fatalf("decode apply response: %v", err)
	}
	if applied.PreviewID != preview.PreviewID || applied.Status != "applied" || !applied.MutationApplied {
		t.Fatalf("expected exact preview to apply once, got %+v", applied)
	}
	if countAgentSkillLifecycleWishlistItems(t, a, profileID, "WISH-2087-CONFIRM") != 1 {
		t.Fatal("confirmed preview did not persist exactly one Wishlist item")
	}
}

func TestGenericAgentSkillPreviewCancelPreventsLaterApply(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	profileID := createAgentSkillLifecycleProfile(t, a, "Durable Agent Cancel")
	preview := previewAgentSkillWishlistCreate(t, a, profileID, "WISH-2087-CANCEL")

	cancel := doRequest(t, a, http.MethodPost, "/api/agent/skills/cancel", strings.NewReader(`{
		"profile_id":"`+profileID+`",
		"preview_id":"`+preview.PreviewID+`"
	}`), map[string]string{"Content-Type": "application/json"})
	if cancel.Code != http.StatusOK {
		t.Fatalf("cancel status=%d body=%s", cancel.Code, cancel.Body.String())
	}
	if !strings.Contains(cancel.Body.String(), `"preview_status":"cancelled"`) ||
		!strings.Contains(cancel.Body.String(), `"mutation_applied":false`) {
		t.Fatalf("expected explicit cancelled non-mutating response, body=%s", cancel.Body.String())
	}

	apply := confirmAgentSkillLifecyclePreview(t, a, profileID, preview.PreviewID)
	if apply.Code != http.StatusConflict ||
		!strings.Contains(apply.Body.String(), `"error":"agent_skill_preview_cancelled"`) ||
		!strings.Contains(apply.Body.String(), `"recoverable":true`) {
		t.Fatalf("cancelled preview apply status=%d body=%s", apply.Code, apply.Body.String())
	}
	if countAgentSkillLifecycleWishlistItems(t, a, profileID, "WISH-2087-CANCEL") != 0 {
		t.Fatal("cancelled preview mutated Wishlist")
	}
}

func TestGenericAgentSkillPreviewExpiryPreventsStaleApply(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	profileID := createAgentSkillLifecycleProfile(t, a, "Durable Agent Expiry")
	preview := previewAgentSkillWishlistCreate(t, a, profileID, "WISH-2087-EXPIRED")

	if _, err := a.db.Exec(`UPDATE agent_skill_previews SET expires_at = '2000-01-01T00:00:00Z' WHERE id = ?`, preview.PreviewID); err != nil {
		t.Fatalf("expire durable preview fixture: %v", err)
	}
	apply := confirmAgentSkillLifecyclePreview(t, a, profileID, preview.PreviewID)
	if apply.Code != http.StatusConflict ||
		!strings.Contains(apply.Body.String(), `"error":"agent_skill_preview_expired"`) ||
		!strings.Contains(apply.Body.String(), `"recoverable":true`) {
		t.Fatalf("expired preview apply status=%d body=%s", apply.Code, apply.Body.String())
	}
	if countAgentSkillLifecycleWishlistItems(t, a, profileID, "WISH-2087-EXPIRED") != 0 {
		t.Fatal("expired preview mutated Wishlist")
	}
}

func TestGenericAgentSkillPreviewRejectsDuplicateConfirmation(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	profileID := createAgentSkillLifecycleProfile(t, a, "Durable Agent Replay")
	preview := previewAgentSkillWishlistCreate(t, a, profileID, "WISH-2087-REPLAY")

	first := confirmAgentSkillLifecyclePreview(t, a, profileID, preview.PreviewID)
	if first.Code != http.StatusOK {
		t.Fatalf("first apply status=%d body=%s", first.Code, first.Body.String())
	}
	second := confirmAgentSkillLifecyclePreview(t, a, profileID, preview.PreviewID)
	if second.Code != http.StatusConflict ||
		!strings.Contains(second.Body.String(), `"error":"agent_skill_preview_already_applied"`) {
		t.Fatalf("duplicate apply status=%d body=%s", second.Code, second.Body.String())
	}
	if countAgentSkillLifecycleWishlistItems(t, a, profileID, "WISH-2087-REPLAY") != 1 {
		t.Fatal("duplicate confirmation created more than one Wishlist item")
	}
}

func TestGenericAgentSkillPreviewConcurrentConfirmationMutatesAtMostOnce(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	profileID := createAgentSkillLifecycleProfile(t, a, "Durable Agent Concurrent Replay")
	preview := previewAgentSkillWishlistCreate(t, a, profileID, "WISH-2087-CONCURRENT")

	type outcome struct {
		status int
		body   string
	}
	start := make(chan struct{})
	outcomes := make(chan outcome, 2)
	var workers sync.WaitGroup
	for range 2 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			response := confirmAgentSkillLifecyclePreview(t, a, profileID, preview.PreviewID)
			outcomes <- outcome{status: response.Code, body: response.Body.String()}
		}()
	}
	close(start)
	workers.Wait()
	close(outcomes)

	successes := 0
	conflicts := 0
	for result := range outcomes {
		switch result.status {
		case http.StatusOK:
			successes++
		case http.StatusConflict:
			conflicts++
		default:
			t.Fatalf("unexpected concurrent apply status=%d body=%s", result.status, result.body)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("expected one success and one replay conflict, successes=%d conflicts=%d", successes, conflicts)
	}
	if countAgentSkillLifecycleWishlistItems(t, a, profileID, "WISH-2087-CONCURRENT") != 1 {
		t.Fatal("concurrent confirmations mutated Wishlist more than once")
	}
}

func TestGenericAgentSkillPreviewRejectsCrossProfileReplay(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	ownerID := createAgentSkillLifecycleProfile(t, a, "Durable Agent Owner")
	otherID := createAgentSkillLifecycleProfile(t, a, "Durable Agent Other")
	preview := previewAgentSkillWishlistCreate(t, a, ownerID, "WISH-2087-PROFILE")

	crossProfile := confirmAgentSkillLifecyclePreview(t, a, otherID, preview.PreviewID)
	if crossProfile.Code != http.StatusNotFound ||
		!strings.Contains(crossProfile.Body.String(), `"error":"agent_skill_preview_not_found"`) {
		t.Fatalf("cross-profile apply status=%d body=%s", crossProfile.Code, crossProfile.Body.String())
	}
	if countAgentSkillLifecycleWishlistItems(t, a, ownerID, "WISH-2087-PROFILE") != 0 ||
		countAgentSkillLifecycleWishlistItems(t, a, otherID, "WISH-2087-PROFILE") != 0 {
		t.Fatal("cross-profile replay mutated Wishlist")
	}

	ownerApply := confirmAgentSkillLifecyclePreview(t, a, ownerID, preview.PreviewID)
	if ownerApply.Code != http.StatusOK {
		t.Fatalf("owner apply status=%d body=%s", ownerApply.Code, ownerApply.Body.String())
	}
}

func TestGenericAgentSkillPreviewBindsProvenanceAndRejectsSecrets(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	profileID := createAgentSkillLifecycleProfile(t, a, "Durable Agent Provenance")
	preview := previewAgentSkillWishlistCreate(t, a, profileID, "WISH-2087-PROVENANCE")

	var skillID, sourceSurface, sourceChannel, sourceThreadID, sourceMessageID, targetJSON, status string
	if err := a.db.QueryRow(`
		SELECT skill_id, source_surface, source_channel, source_thread_id, source_message_id, target_json, status
		FROM agent_skill_previews WHERE id = ? AND profile_id = ?
	`, preview.PreviewID, profileID).Scan(&skillID, &sourceSurface, &sourceChannel, &sourceThreadID, &sourceMessageID, &targetJSON, &status); err != nil {
		t.Fatalf("read durable preview provenance: %v", err)
	}
	if skillID != "cabinet.wishlist.create_entry" || sourceSurface != "chats.main" || sourceChannel != "in-app" ||
		sourceThreadID != "thread-2087" || sourceMessageID != "message-2087" || status != "previewed" {
		t.Fatalf("durable preview provenance mismatch: skill=%q surface=%q channel=%q thread=%q message=%q status=%q", skillID, sourceSurface, sourceChannel, sourceThreadID, sourceMessageID, status)
	}
	if !strings.Contains(targetJSON, "WISH-2087-PROVENANCE") || strings.Contains(targetJSON, "private-note-2087") {
		t.Fatalf("expected bounded target provenance without unrestricted notes, target=%s", targetJSON)
	}
	if notes, ok := preview.Target["notes"].(string); ok && strings.Contains(strings.TrimSpace(notes), "private-note-2087") {
		t.Fatalf("preview response leaked unrestricted notes: %+v", preview.Target)
	}

	secret := "provider-secret-2087-do-not-echo"
	if _, err := profile.NewRepository(a.db).PutAgentAuthorityPolicy(context.Background(), profileID, profile.AgentAuthorityPolicy{
		Mode:                  profile.AgentAuthorityModeApprovedExternalActions,
		ExternalWriteApproved: true,
	}); err != nil {
		t.Fatalf("approve external Agent action for secret-backed preview: %v", err)
	}
	secretPreview := doRequest(t, a, http.MethodPost, "/api/agent/skills/preview", strings.NewReader(`{
		"profile_id":"`+profileID+`",
		"skill_id":"cabinet.integrations.configure_provider",
		"source_surface":"settings.integrations",
		"source_channel":"in-app",
		"source_thread_id":"thread-secret-2087",
		"source_message_id":"message-secret-2087",
		"parameters":{"provider_id":"openai","setup_payload":"api-key","api_key":"`+secret+`"}
	}`), map[string]string{"Content-Type": "application/json"})
	if secretPreview.Code != http.StatusOK || strings.Contains(secretPreview.Body.String(), secret) {
		t.Fatalf("secret preview status=%d body=%s", secretPreview.Code, secretPreview.Body.String())
	}
	var secretPayload agentSkillLifecyclePreviewPayload
	if err := json.NewDecoder(secretPreview.Body).Decode(&secretPayload); err != nil || secretPayload.PreviewID == "" {
		t.Fatalf("decode secret-backed preview: err=%v payload=%+v", err, secretPayload)
	}
	var parametersJSON, secretRef string
	if err := a.db.QueryRow(`SELECT parameters_json, secret_ref FROM agent_skill_previews WHERE id = ?`, secretPayload.PreviewID).Scan(&parametersJSON, &secretRef); err != nil {
		t.Fatalf("read secret-backed preview record: %v", err)
	}
	if strings.Contains(parametersJSON, secret) || secretRef == "" {
		t.Fatalf("durable preview row must contain redacted parameters plus an opaque secret reference: parameters=%s ref=%q", parametersJSON, secretRef)
	}
	pendingSecret, err := profile.NewRepository(a.db).GetSecret(context.Background(), profileID, secretRef)
	if err != nil {
		t.Fatalf("read pending preview secret through storage boundary: %v", err)
	}
	if !strings.Contains(pendingSecret, secret) {
		t.Fatalf("pending preview secret must preserve the approved value for apply")
	}
	secretApply := confirmAgentSkillLifecyclePreview(t, a, profileID, secretPayload.PreviewID)
	if secretApply.Code != http.StatusOK || strings.Contains(secretApply.Body.String(), secret) {
		t.Fatalf("secret-backed apply status=%d body=%s", secretApply.Code, secretApply.Body.String())
	}
	if _, err := profile.NewRepository(a.db).GetSecret(context.Background(), profileID, secretRef); err == nil {
		t.Fatal("pending preview secret must be deleted after apply")
	}
}

func TestGenericAgentSkillPreviewRechecksAuthorityAtApply(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	profileID := createAgentSkillLifecycleProfile(t, a, "Durable Agent Authority Recheck")
	preview := previewAgentSkillWishlistCreate(t, a, profileID, "WISH-2087-AUTHORITY")
	if _, err := profile.NewRepository(a.db).PutAgentAuthorityPolicy(context.Background(), profileID, profile.AgentAuthorityPolicy{
		Mode: profile.AgentAuthorityModeReadOnly,
	}); err != nil {
		t.Fatalf("set read-only authority policy: %v", err)
	}

	apply := confirmAgentSkillLifecyclePreview(t, a, profileID, preview.PreviewID)
	if apply.Code != http.StatusConflict || !strings.Contains(apply.Body.String(), `"error":"agent_authority_read_only"`) {
		t.Fatalf("authority recheck status=%d body=%s", apply.Code, apply.Body.String())
	}
	if countAgentSkillLifecycleWishlistItems(t, a, profileID, "WISH-2087-AUTHORITY") != 0 {
		t.Fatal("authority-blocked preview mutated Wishlist")
	}
	var status string
	if err := a.db.QueryRow(`SELECT status FROM agent_skill_previews WHERE id = ?`, preview.PreviewID).Scan(&status); err != nil {
		t.Fatalf("read authority-blocked preview status: %v", err)
	}
	if status != "previewed" {
		t.Fatalf("authority denial should leave preview pending for policy recovery, got %q", status)
	}
}

func TestGenericAgentSkillPreviewRechecksTargetBeforeMutation(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	profileID := createAgentSkillLifecycleProfile(t, a, "Durable Agent Target Recheck")
	create := doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+profileID+`",
		"skill_id":"cabinet.inventory.create_item",
		"confirm":true,
		"parameters":{"part_number":"INV-2087-TARGET","title":"Target before deletion","brand":"AFX","category":"Slot Cars"}
	}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusOK {
		t.Fatalf("create target status=%d body=%s", create.Code, create.Body.String())
	}
	var created struct {
		Target struct {
			ItemID string `json:"item_id"`
		} `json:"target"`
	}
	if err := json.NewDecoder(create.Body).Decode(&created); err != nil || created.Target.ItemID == "" {
		t.Fatalf("decode created target: err=%v payload=%+v", err, created)
	}

	previewResponse := doRequest(t, a, http.MethodPost, "/api/agent/skills/preview", strings.NewReader(`{
		"profile_id":"`+profileID+`",
		"skill_id":"cabinet.inventory.update_item",
		"source_surface":"inventory.detail",
		"source_channel":"in-app",
		"source_thread_id":"thread-target-2087",
		"source_message_id":"message-target-2087",
		"parameters":{"item_id":"`+created.Target.ItemID+`","title":"Target after update"}
	}`), map[string]string{"Content-Type": "application/json"})
	if previewResponse.Code != http.StatusOK {
		t.Fatalf("target preview status=%d body=%s", previewResponse.Code, previewResponse.Body.String())
	}
	var preview agentSkillLifecyclePreviewPayload
	if err := json.NewDecoder(previewResponse.Body).Decode(&preview); err != nil || preview.PreviewID == "" {
		t.Fatalf("decode target preview: err=%v payload=%+v", err, preview)
	}
	if _, err := a.db.Exec(`DELETE FROM canonical_items WHERE id = ? AND profile_id = ?`, created.Target.ItemID, profileID); err != nil {
		t.Fatalf("remove target fixture after preview: %v", err)
	}

	apply := confirmAgentSkillLifecyclePreview(t, a, profileID, preview.PreviewID)
	if apply.Code != http.StatusConflict ||
		!strings.Contains(apply.Body.String(), `"error":"inventory_item_required"`) ||
		!strings.Contains(apply.Body.String(), `"recoverable":true`) {
		t.Fatalf("stale target apply status=%d body=%s", apply.Code, apply.Body.String())
	}
	var status string
	if err := a.db.QueryRow(`SELECT status FROM agent_skill_previews WHERE id = ?`, preview.PreviewID).Scan(&status); err != nil {
		t.Fatalf("read stale-target preview status: %v", err)
	}
	if status != "failed" {
		t.Fatalf("stale target preview should become single-use failed state, got %q", status)
	}
}

func TestGenericAgentSkillPreviewExpiryCleanupRemovesPendingSecret(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	profileID := createAgentSkillLifecycleProfile(t, a, "Durable Agent Expiry Secret Cleanup")
	secret := "provider-secret-2087-expiry-cleanup"
	previewResponse := doRequest(t, a, http.MethodPost, "/api/agent/skills/preview", strings.NewReader(`{
		"profile_id":"`+profileID+`",
		"skill_id":"cabinet.integrations.configure_provider",
		"source_surface":"settings.integrations",
		"source_channel":"in-app",
		"parameters":{"provider_id":"openai","setup_payload":"api-key","api_key":"`+secret+`"}
	}`), map[string]string{"Content-Type": "application/json"})
	if previewResponse.Code != http.StatusOK || strings.Contains(previewResponse.Body.String(), secret) {
		t.Fatalf("secret expiry preview status=%d body=%s", previewResponse.Code, previewResponse.Body.String())
	}
	var preview agentSkillLifecyclePreviewPayload
	if err := json.NewDecoder(previewResponse.Body).Decode(&preview); err != nil || preview.PreviewID == "" {
		t.Fatalf("decode secret expiry preview: err=%v payload=%+v", err, preview)
	}
	var secretRef string
	if err := a.db.QueryRow(`SELECT secret_ref FROM agent_skill_previews WHERE id = ?`, preview.PreviewID).Scan(&secretRef); err != nil || secretRef == "" {
		t.Fatalf("read pending secret reference: err=%v ref=%q", err, secretRef)
	}
	if _, err := a.db.Exec(`UPDATE agent_skill_previews SET expires_at = '2000-01-01T00:00:00Z' WHERE id = ?`, preview.PreviewID); err != nil {
		t.Fatalf("expire secret-backed preview: %v", err)
	}
	if err := cleanupExpiredDurableAgentSkillPreviews(context.Background(), a.db, "2026-08-12T00:00:00Z"); err != nil {
		t.Fatalf("cleanup expired durable previews: %v", err)
	}
	var status, storedSecretRef string
	if err := a.db.QueryRow(`SELECT status, secret_ref FROM agent_skill_previews WHERE id = ?`, preview.PreviewID).Scan(&status, &storedSecretRef); err != nil {
		t.Fatalf("read cleaned preview: %v", err)
	}
	if status != "expired" || storedSecretRef != "" {
		t.Fatalf("expired preview cleanup mismatch: status=%q secret_ref=%q", status, storedSecretRef)
	}
	if _, err := profile.NewRepository(a.db).GetSecret(context.Background(), profileID, secretRef); err == nil {
		t.Fatal("expired pending secret must be removed")
	}
}

func TestGenericAgentSkillPreviewPersistsAcrossRestartAndCanBeRetrieved(t *testing.T) {
	base := t.TempDir()
	cfg := config.Config{
		Addr:           "127.0.0.1:0",
		DataDir:        base,
		DBPath:         filepath.Join(base, "cabinet.db"),
		UpdateChannel:  update.ChannelStable,
		WebAuthnRPID:   "127.0.0.1",
		WebAuthnOrigin: "http://127.0.0.1:8080",
		WebAuthnName:   "Cabinet Durable Preview Test",
		BackupInterval: 60,
	}
	a, err := New(cfg)
	if err != nil {
		t.Fatalf("create first app: %v", err)
	}
	profileID := createAgentSkillLifecycleProfile(t, a, "Durable Agent Restart")
	preview := previewAgentSkillWishlistCreate(t, a, profileID, "WISH-2087-RESTART")
	if err := a.close(); err != nil {
		t.Fatalf("close first app: %v", err)
	}

	a, err = New(cfg)
	if err != nil {
		t.Fatalf("reopen app: %v", err)
	}
	t.Cleanup(func() { _ = a.close() })
	get := doRequest(t, a, http.MethodGet, "/api/agent/skills/preview?profile_id="+profileID+"&preview_id="+preview.PreviewID, nil, nil)
	if get.Code != http.StatusOK {
		t.Fatalf("get persisted preview status=%d body=%s", get.Code, get.Body.String())
	}
	var persisted agentSkillLifecyclePreviewPayload
	if err := json.NewDecoder(get.Body).Decode(&persisted); err != nil {
		t.Fatalf("decode persisted preview: %v", err)
	}
	if persisted.PreviewID != preview.PreviewID || persisted.Status != "previewed" || persisted.SourceThreadID != "thread-2087" {
		t.Fatalf("persisted preview mismatch: %+v", persisted)
	}
	apply := confirmAgentSkillLifecyclePreview(t, a, profileID, preview.PreviewID)
	if apply.Code != http.StatusOK {
		t.Fatalf("apply persisted preview status=%d body=%s", apply.Code, apply.Body.String())
	}
	if countAgentSkillLifecycleWishlistItems(t, a, profileID, "WISH-2087-RESTART") != 1 {
		t.Fatal("persisted preview did not apply exactly once after restart")
	}
}

func createAgentSkillLifecycleProfile(t *testing.T, a *App, name string) string {
	t.Helper()
	resp := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"`+name+`"}`), map[string]string{"Content-Type": "application/json"})
	if resp.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", resp.Code, resp.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	return profile.ID
}

func previewAgentSkillWishlistCreate(t *testing.T, a *App, profileID, partNumber string) agentSkillLifecyclePreviewPayload {
	t.Helper()
	resp := doRequest(t, a, http.MethodPost, "/api/agent/skills/preview", strings.NewReader(`{
		"profile_id":"`+profileID+`",
		"skill_id":"cabinet.wishlist.create_entry",
		"source_surface":"chats.main",
		"source_channel":"in-app",
		"source_thread_id":"thread-2087",
		"source_message_id":"message-2087",
		"parameters":{
			"part_number":"`+partNumber+`",
			"title":"Durable Agent Wishlist Item",
			"brand":"AFX",
			"category":"Slot Cars",
			"notes":"private-note-2087"
		}
	}`), map[string]string{"Content-Type": "application/json"})
	if resp.Code != http.StatusOK {
		t.Fatalf("preview status=%d body=%s", resp.Code, resp.Body.String())
	}
	var preview agentSkillLifecyclePreviewPayload
	if err := json.NewDecoder(resp.Body).Decode(&preview); err != nil {
		t.Fatalf("decode preview: %v", err)
	}
	if preview.PreviewID == "" {
		t.Fatalf("expected durable preview id, body=%s", resp.Body.String())
	}
	return preview
}

func confirmAgentSkillLifecyclePreview(t *testing.T, a *App, profileID, previewID string) *httptest.ResponseRecorder {
	t.Helper()
	return doRequest(t, a, http.MethodPost, "/api/agent/skills/apply", strings.NewReader(`{
		"profile_id":"`+profileID+`",
		"preview_id":"`+previewID+`",
		"confirm":true
	}`), map[string]string{"Content-Type": "application/json"})
}

func countAgentSkillLifecycleWishlistItems(t *testing.T, a *App, profileID, partNumber string) int {
	t.Helper()
	var count int
	if err := a.db.QueryRow(`SELECT COUNT(1) FROM canonical_items WHERE profile_id = ? AND part_number = ?`, profileID, partNumber).Scan(&count); err != nil {
		t.Fatalf("count Wishlist item: %v", err)
	}
	return count
}
