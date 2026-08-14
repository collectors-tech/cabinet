package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/collectors-tech/cabinet/internal/agentskills"
	"github.com/collectors-tech/cabinet/internal/ai"
	"github.com/collectors-tech/cabinet/internal/chat"
	"github.com/collectors-tech/cabinet/internal/profile"
	"github.com/collectors-tech/cabinet/internal/telegramcapture"
)

func TestTelegramAgentConversationUsesStableThreadAndSharedPlanner(t *testing.T) {
	a := newTestApp(t)
	profileID := createTelegramConversationProfile(t, a, "Telegram conversation")
	service := newTelegramAgentConversationService(
		a.db,
		profile.NewRepository(a.db),
		chat.NewService(a.db, t.TempDir()),
		func(id string) agentskills.Registry { return agentSkillRegistryForTest(id) },
		ai.NewAssistantProviderRegistry(&captureAssistantProvider{responseText: `{"decision":"select_skill","skill_id":"cabinet.inventory.search_items","parameters":{"query":"TG-READ-1"},"message":"I found the matching Cabinet items."}`}),
	)

	first, err := service.HandleText(context.Background(), profileID, telegramAgentTextRequest{
		SenderID: "44001", ChatID: "44001", ChatType: "private", MessageID: "101", Text: "find my item TG-READ-1",
	})
	if err != nil {
		t.Fatalf("HandleText(first) error = %v", err)
	}
	second, err := service.HandleText(context.Background(), profileID, telegramAgentTextRequest{
		SenderID: "44001", ChatID: "44001", ChatType: "private", MessageID: "102", Text: "show it again",
	})
	if err != nil {
		t.Fatalf("HandleText(second) error = %v", err)
	}
	firstThread := telegramResultThreadID(t, first)
	secondThread := telegramResultThreadID(t, second)
	if firstThread == "" || secondThread != firstThread {
		t.Fatalf("expected stable Telegram thread, first=%q second=%q", firstThread, secondThread)
	}
	reply := telegramResultReply(t, first)
	if !strings.Contains(reply.Text, "matching Cabinet items") || len(reply.ActionButtons) != 0 {
		t.Fatalf("expected natural read-only reply without mutation controls, got %+v", reply)
	}
	raw, _ := json.Marshal(first)
	for _, forbidden := range []string{"cabinet.inventory.search_items", "skill_id", "key=value"} {
		if strings.Contains(string(raw), forbidden) {
			t.Fatalf("Telegram response leaked planner grammar %q: %s", forbidden, raw)
		}
	}
}

func TestTelegramAgentConversationDeduplicatesConcurrentAndRestartDelivery(t *testing.T) {
	a := newTestApp(t)
	profileID := createTelegramConversationProfile(t, a, "Telegram delivery replay")
	newService := func() *telegramAgentConversationService {
		return newTelegramAgentConversationService(
			a.db,
			profile.NewRepository(a.db),
			chat.NewService(a.db, t.TempDir()),
			func(id string) agentskills.Registry { return agentSkillRegistryForTest(id) },
			ai.NewAssistantProviderRegistry(&captureAssistantProvider{responseText: `{"decision":"select_skill","skill_id":"cabinet.inventory.create_item","parameters":{"part_number":"TG-REPLAY-1","title":"Telegram Replay Item","category":"Slot Cars"},"message":"I prepared the item for review."}`}),
		)
	}
	service := newService()
	req := telegramAgentTextRequest{
		SenderID: "44009", ChatID: "44009", ChatType: "private", MessageID: "901",
		Text: "add TG-REPLAY-1 named Telegram Replay Item", SourceMetadata: map[string]any{"update_id": 9901},
	}
	var wg sync.WaitGroup
	results := make(chan map[string]any, 2)
	errs := make(chan error, 2)
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := service.HandleText(context.Background(), profileID, req)
			results <- result
			errs <- err
		}()
	}
	wg.Wait()
	close(results)
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent delivery error: %v", err)
		}
	}
	previewIDs := map[string]bool{}
	for result := range results {
		previewIDs[telegramUntypedPreviewID(result)] = true
	}
	if len(previewIDs) != 1 || previewIDs[""] {
		t.Fatalf("concurrent duplicate created different previews: %+v", previewIDs)
	}

	restarted := newService()
	replay, err := restarted.HandleText(context.Background(), profileID, req)
	if err != nil || replay["duplicate"] != true || !previewIDs[telegramUntypedPreviewID(replay)] {
		t.Fatalf("restart replay did not return durable result: result=%+v err=%v", replay, err)
	}
	var messages, previews, mappings, threads int
	for query, target := range map[string]*int{
		`SELECT COUNT(1) FROM chat_messages WHERE profile_id = ? AND role = 'user'`:                   &messages,
		`SELECT COUNT(1) FROM agent_skill_previews WHERE profile_id = ?`:                              &previews,
		`SELECT COUNT(1) FROM telegram_agent_threads WHERE profile_id = ? AND sender_id = '44009'`:    &mappings,
		`SELECT COUNT(1) FROM chat_threads WHERE profile_id = ? AND title = 'Telegram Cabinet Agent'`: &threads,
	} {
		if err := a.db.QueryRow(query, profileID).Scan(target); err != nil {
			t.Fatalf("count durable Telegram state: %v", err)
		}
	}
	if messages != 1 || previews != 1 || mappings != 1 || threads != 1 {
		t.Fatalf("duplicate/restart state messages=%d previews=%d mappings=%d threads=%d", messages, previews, mappings, threads)
	}
}

func TestTelegramAgentHTTPRoutesRejectForgedPairedIdentifiers(t *testing.T) {
	a := newTestApp(t)
	profileID := createTelegramConversationProfile(t, a, "Telegram forged HTTP")
	if err := profile.NewRepository(a.db).PutSettings(context.Background(), profileID, map[string]string{
		"telegram.catalog_capture.sender_id": "44999",
		"telegram.catalog_capture.chat_id":   "44999",
	}); err != nil {
		t.Fatalf("pair test peer: %v", err)
	}
	for _, tc := range []struct {
		path string
		body string
	}{
		{path: "/api/telegram/agent-text", body: `{"sender_id":"44999","chat_id":"44999","chat_type":"private","message_id":"9001","text":"show my inventory"}`},
		{path: "/api/telegram/agent-text-callbacks", body: `{"sender_id":"44999","chat_id":"44999","chat_type":"private","callback_query_id":"forged","callback_data":"asp_00000000000000000000000000000000:apply"}`},
		{path: "/api/telegram/webhook/catalog-captures", body: `{"update_id":9002,"message":{"message_id":2,"from":{"id":44999},"chat":{"id":44999,"type":"private"},"text":"Barcode 123"}}`},
		{path: "/api/telegram/catalog-capture-callbacks", body: `{"sender_id":"44999","chat_id":"44999","callback_data":"forged"}`},
	} {
		req := httptest.NewRequest(http.MethodPost, tc.path, strings.NewReader(tc.body))
		req.Host = "127.0.0.1:8080"
		req.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()
		a.srv.Handler.ServeHTTP(recorder, req)
		if recorder.Code != http.StatusForbidden || !strings.Contains(recorder.Body.String(), "telegram_connector_only") {
			t.Fatalf("untrusted route %s status=%d body=%s", tc.path, recorder.Code, recorder.Body.String())
		}
	}
}

func TestTelegramAgentConversationMutationUsesOpaqueSingleUsePreview(t *testing.T) {
	a := newTestApp(t)
	profileID := createTelegramConversationProfile(t, a, "Telegram mutation")
	service := newTelegramAgentConversationService(
		a.db,
		profile.NewRepository(a.db),
		chat.NewService(a.db, t.TempDir()),
		func(id string) agentskills.Registry { return agentSkillRegistryForTest(id) },
		ai.NewAssistantProviderRegistry(&captureAssistantProvider{responseText: `{"decision":"select_skill","skill_id":"cabinet.inventory.create_item","parameters":{"part_number":"TG-WRITE-1","title":"Telegram Preview Item","category":"Slot Cars"},"message":"I prepared the item for review."}`}),
	)

	result, err := service.HandleText(context.Background(), profileID, telegramAgentTextRequest{
		SenderID: "44002", ChatID: "44002", ChatType: "private", MessageID: "201", Text: "add TG-WRITE-1 named Telegram Preview Item",
	})
	if err != nil {
		t.Fatalf("HandleText(mutation) error = %v", err)
	}
	previewID := telegramResultPreviewID(t, result)
	if !strings.HasPrefix(previewID, "asp_") {
		t.Fatalf("expected opaque durable preview id, got %q", previewID)
	}
	reply := telegramResultReply(t, result)
	if len(reply.ActionButtons) != 2 || reply.ActionButtons[0].CallbackData != previewID+":apply" || reply.ActionButtons[1].CallbackData != previewID+":cancel" {
		t.Fatalf("expected opaque apply/cancel callbacks, got %+v", reply.ActionButtons)
	}
	assertTelegramItemCount(t, a, profileID, "TG-WRITE-1", 0)

	callback := telegramAgentTextCallbackRequest{
		SenderID: "44002", ChatID: "44002", ChatType: "private", MessageID: "202", CallbackQueryID: "cb-1", CallbackData: previewID + ":apply",
	}
	applied, err := service.HandleCallback(context.Background(), profileID, callback)
	if err != nil {
		t.Fatalf("HandleCallback(apply) error = %v", err)
	}
	if applied["confirmation_state"] != "applied" {
		t.Fatalf("expected applied terminal state, got %+v", applied)
	}
	assertTelegramItemCount(t, a, profileID, "TG-WRITE-1", 1)

	replayed, err := service.HandleCallback(context.Background(), profileID, callback)
	if err != nil {
		t.Fatalf("HandleCallback(replay) must be idempotent, error = %v", err)
	}
	if replayed["confirmation_state"] != "applied" || replayed["duplicate"] != true {
		t.Fatalf("expected duplicate callback to return applied terminal state, got %+v", replayed)
	}
	assertTelegramItemCount(t, a, profileID, "TG-WRITE-1", 1)

	wrongSender := callback
	wrongSender.SenderID = "99999"
	if _, err := service.HandleCallback(context.Background(), profileID, wrongSender); err == nil {
		t.Fatal("expected cross-sender callback to fail closed")
	}
}

func TestTelegramAgentConversationRejectsGroupsCrossProfileExpiryAndAdminApproval(t *testing.T) {
	a := newTestApp(t)
	profileID := createTelegramConversationProfile(t, a, "Telegram boundaries")
	otherProfileID := createTelegramConversationProfile(t, a, "Other Telegram profile")
	readService := newTelegramAgentConversationService(
		a.db,
		profile.NewRepository(a.db),
		chat.NewService(a.db, t.TempDir()),
		func(id string) agentskills.Registry { return agentSkillRegistryForTest(id) },
		ai.NewAssistantProviderRegistry(&captureAssistantProvider{responseText: `{"decision":"select_skill","skill_id":"cabinet.inventory.search_items","parameters":{},"message":"Searching Cabinet."}`}),
	)
	if _, err := readService.HandleText(context.Background(), profileID, telegramAgentTextRequest{SenderID: "44003", ChatID: "-1001", ChatType: "group", MessageID: "301", Text: "search Cabinet"}); err == nil {
		t.Fatal("expected group request to fail closed")
	}
	if _, err := readService.HandleText(context.Background(), otherProfileID, telegramAgentTextRequest{ProfileID: profileID, SenderID: "44003", ChatID: "44003", ChatType: "private", MessageID: "302", Text: "search Cabinet"}); err == nil {
		t.Fatal("expected cross-profile request to fail closed")
	}
	maintenanceService := newTelegramAgentConversationService(
		a.db,
		profile.NewRepository(a.db),
		chat.NewService(a.db, t.TempDir()),
		func(id string) agentskills.Registry { return agentSkillRegistryForTest(id) },
		ai.NewAssistantProviderRegistry(&captureAssistantProvider{responseText: `{"decision":"select_skill","skill_id":"cabinet.maintenance.run_safe_check","parameters":{"maintenance_check":"status"},"message":"Cabinet maintenance status is healthy."}`}),
	)
	maintenance, err := maintenanceService.HandleText(context.Background(), profileID, telegramAgentTextRequest{SenderID: "44003", ChatID: "44003", ChatType: "private", MessageID: "maintenance-303", Text: "run a safe maintenance status check"})
	if err != nil {
		t.Fatalf("safe maintenance read should remain available on Telegram: %v", err)
	}
	if maintenance["confirmation_state"] == "in_app_approval_required" || len(telegramResultReply(t, maintenance).ActionButtons) != 0 {
		t.Fatalf("safe maintenance read was incorrectly treated as elevated admin mutation: %+v", maintenance)
	}

	mutationService := newTelegramAgentConversationService(
		a.db,
		profile.NewRepository(a.db),
		chat.NewService(a.db, t.TempDir()),
		func(id string) agentskills.Registry { return agentSkillRegistryForTest(id) },
		ai.NewAssistantProviderRegistry(&captureAssistantProvider{responseText: `{"decision":"select_skill","skill_id":"cabinet.inventory.create_item","parameters":{"part_number":"TG-EXPIRED-1","title":"Expired Telegram Item","category":"Slot Cars"},"message":"I prepared the item for review."}`}),
	)
	mutation, err := mutationService.HandleText(context.Background(), profileID, telegramAgentTextRequest{SenderID: "44003", ChatID: "44003", ChatType: "private", MessageID: "304", Text: "add TG-EXPIRED-1 named Expired Telegram Item"})
	if err != nil {
		t.Fatalf("prepare expiring mutation: %v", err)
	}
	previewID := telegramResultPreviewID(t, mutation)
	if _, err := mutationService.HandleCallback(context.Background(), otherProfileID, telegramAgentTextCallbackRequest{SenderID: "44003", ChatID: "44003", ChatType: "private", CallbackQueryID: "cross-profile-304", CallbackData: previewID + ":apply"}); err == nil {
		t.Fatal("expected cross-profile callback to fail closed")
	}
	if _, err := a.db.Exec(`UPDATE agent_skill_previews SET expires_at = '2000-01-01T00:00:00Z' WHERE id = ?`, previewID); err != nil {
		t.Fatalf("expire preview: %v", err)
	}
	if _, err := mutationService.HandleCallback(context.Background(), profileID, telegramAgentTextCallbackRequest{SenderID: "44003", ChatID: "44003", ChatType: "private", CallbackQueryID: "expired-304", CallbackData: previewID + ":apply"}); err == nil {
		t.Fatal("expected expired Telegram callback to fail closed")
	}
	assertTelegramItemCount(t, a, profileID, "TG-EXPIRED-1", 0)

	adminService := newTelegramAgentConversationService(
		a.db,
		profile.NewRepository(a.db),
		chat.NewService(a.db, t.TempDir()),
		func(id string) agentskills.Registry { return agentSkillRegistryForTest(id) },
		ai.NewAssistantProviderRegistry(&captureAssistantProvider{responseText: `{"decision":"select_skill","skill_id":"cabinet.users.update_role","parameters":{"user_id":"user-1","role":"admin"},"message":"I prepared the role update."}`}),
	)
	result, err := adminService.HandleText(context.Background(), profileID, telegramAgentTextRequest{SenderID: "44003", ChatID: "44003", ChatType: "private", MessageID: "303", Text: "make user-1 an admin"})
	if err != nil {
		t.Fatalf("admin request should return safe in-app guidance, error=%v", err)
	}
	if result["confirmation_state"] != "in_app_approval_required" || telegramResultPreviewID(t, result) != "" || len(telegramResultReply(t, result).ActionButtons) != 1 {
		t.Fatalf("expected admin request to require in-app approval without Telegram callback, got %+v", result)
	}
}

func createTelegramConversationProfile(t *testing.T, a *App, name string) string {
	t.Helper()
	resp := doRequest(t, a, "POST", "/api/profiles", strings.NewReader(`{"name":"`+name+`"}`), map[string]string{"Content-Type": "application/json"})
	if resp.Code != 201 {
		t.Fatalf("create profile status=%d body=%s", resp.Code, resp.Body.String())
	}
	var created struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	return created.ID
}

func agentSkillRegistryForTest(profileID string) agentskills.Registry {
	return agentskills.NewProfileRegistry(profileID, nil, nil)
}

func telegramResultThreadID(t *testing.T, result map[string]any) string {
	t.Helper()
	thread, ok := result["thread"].(chat.Thread)
	if !ok {
		t.Fatalf("missing typed thread result: %+v", result)
	}
	return thread.ID
}

func telegramResultPreviewID(t *testing.T, result map[string]any) string {
	t.Helper()
	value, _ := result["preview_id"].(string)
	return value
}

func telegramUntypedPreviewID(result map[string]any) string {
	value, _ := result["preview_id"].(string)
	return value
}

func telegramResultReply(t *testing.T, result map[string]any) telegramcapture.TelegramReply {
	t.Helper()
	reply, ok := result["telegram_reply"].(telegramcapture.TelegramReply)
	if !ok {
		t.Fatalf("missing typed Telegram reply: %+v", result)
	}
	return reply
}

func assertTelegramItemCount(t *testing.T, a *App, profileID, partNumber string, want int) {
	t.Helper()
	var count int
	if err := a.db.QueryRow(`SELECT COUNT(1) FROM canonical_items WHERE profile_id = ? AND part_number = ? AND COALESCE(deleted_at, '') = ''`, profileID, partNumber).Scan(&count); err != nil {
		t.Fatalf("count items: %v", err)
	}
	if count != want {
		t.Fatalf("item count for %s = %d, want %d", partNumber, count, want)
	}
}
