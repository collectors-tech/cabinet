package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/collectors-tech/cabinet/internal/profile"
)

func TestTelegramLocalConnectorValidatesResolvesWebhookAndPairsPrivateChat(t *testing.T) {
	const token = "123456:test-runtime-secret"
	var mu sync.Mutex
	methods := []string{}
	pairingCode := ""
	webhookDeleted := false
	updatesDelivered := false
	telegram := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method := strings.TrimPrefix(r.URL.Path, "/bot"+token+"/")
		mu.Lock()
		methods = append(methods, method)
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		switch method {
		case "getMe":
			_, _ = w.Write([]byte(`{"ok":true,"result":{"id":2085,"is_bot":true,"first_name":"Cabinet","username":"cabinet_fixture_bot"}}`))
		case "getWebhookInfo":
			if webhookDeleted {
				_, _ = w.Write([]byte(`{"ok":true,"result":{"url":"","pending_update_count":0}}`))
			} else {
				_, _ = w.Write([]byte(`{"ok":true,"result":{"url":"https://old.example/telegram","pending_update_count":2}}`))
			}
		case "deleteWebhook":
			webhookDeleted = true
			_, _ = w.Write([]byte(`{"ok":true,"result":true}`))
		case "getUpdates":
			if pairingCode == "" || updatesDelivered {
				_, _ = w.Write([]byte(`{"ok":true,"result":[]}`))
				return
			}
			updatesDelivered = true
			_, _ = w.Write([]byte(`{"ok":true,"result":[{"update_id":7001,"message":{"message_id":91,"from":{"id":4444},"chat":{"id":4444,"type":"private"},"text":"/start ` + pairingCode + `"}}]}`))
		case "sendMessage":
			_, _ = w.Write([]byte(`{"ok":true,"result":{"message_id":92}}`))
		default:
			t.Fatalf("unexpected Telegram Bot API method %q", method)
		}
	}))
	defer telegram.Close()
	t.Setenv("CABINET_E2E_MODE", "1")
	t.Setenv("CABINET_TELEGRAM_TEST_API_BASE_URL", telegram.URL)

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Telegram Local Connector"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var created struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&created); err != nil {
		t.Fatalf("decode profile: %v", err)
	}

	testConnection := doRequest(t, a, http.MethodPost, "/api/telegram/connection/test", strings.NewReader(`{"profile_id":"`+created.ID+`","bot_token":"`+token+`"}`), map[string]string{"Content-Type": "application/json"})
	if testConnection.Code != http.StatusConflict {
		t.Fatalf("connection test status=%d body=%s", testConnection.Code, testConnection.Body.String())
	}
	body := testConnection.Body.String()
	for _, want := range []string{`"code":"TELEGRAM_WEBHOOK_CONFLICT"`, `"username":"cabinet_fixture_bot"`, `"credential_returned":false`, `"next_action":"resolve_webhook_conflict"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("connection test missing %s: %s", want, body)
		}
	}
	if strings.Contains(body, token) {
		t.Fatalf("connection test leaked token: %s", body)
	}

	secretRead := doRequest(t, a, http.MethodGet, "/api/profiles/"+created.ID+"/secrets?key=telegram_bot_token", nil, nil)
	if secretRead.Code != http.StatusMethodNotAllowed {
		t.Fatalf("Telegram token must be write-only, status=%d body=%s", secretRead.Code, secretRead.Body.String())
	}

	resolve := doRequest(t, a, http.MethodPost, "/api/telegram/connection/resolve-webhook", strings.NewReader(`{"profile_id":"`+created.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if resolve.Code != http.StatusOK || !strings.Contains(resolve.Body.String(), `"webhook_conflict":false`) {
		t.Fatalf("resolve webhook status=%d body=%s", resolve.Code, resolve.Body.String())
	}

	pairing := doRequest(t, a, http.MethodPost, "/api/telegram/pairing-codes", strings.NewReader(`{"profile_id":"`+created.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if pairing.Code != http.StatusCreated {
		t.Fatalf("create pairing status=%d body=%s", pairing.Code, pairing.Body.String())
	}
	var pairingPayload struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(pairing.Body).Decode(&pairingPayload); err != nil {
		t.Fatalf("decode pairing code: %v", err)
	}
	pairingCode = pairingPayload.Code
	if !strings.HasPrefix(pairingCode, "CAB-") {
		t.Fatalf("unexpected pairing code: %q", pairingCode)
	}

	if _, err := a.telegramConnector.PollOnce(context.Background(), created.ID); err != nil {
		t.Fatalf("PollOnce(pairing) error = %v", err)
	}
	status := doRequest(t, a, http.MethodGet, "/api/telegram/connection/status?profile_id="+created.ID, nil, nil)
	if status.Code != http.StatusOK {
		t.Fatalf("status code=%d body=%s", status.Code, status.Body.String())
	}
	statusBody := status.Body.String()
	for _, want := range []string{`"sender_id":"4444"`, `"chat_id":"4444"`, `"paired":true`, `"transport":"long_polling"`, `"public_listener":false`, `"offset":7002`, `"credential_returned":false`} {
		if !strings.Contains(statusBody, want) {
			t.Fatalf("connector status missing %s: %s", want, statusBody)
		}
	}
	if strings.Contains(statusBody, token) {
		t.Fatalf("connector status leaked token: %s", statusBody)
	}
	disconnect := doRequest(t, a, http.MethodPost, "/api/telegram/connection/disconnect", strings.NewReader(`{"profile_id":"`+created.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if disconnect.Code != http.StatusOK {
		t.Fatalf("disconnect status=%d body=%s", disconnect.Code, disconnect.Body.String())
	}
	disconnectBody := disconnect.Body.String()
	for _, want := range []string{`"bot_token_present":false`, `"paired":false`, `"offset":0`, `"state":"disabled"`} {
		if !strings.Contains(disconnectBody, want) {
			t.Fatalf("disconnect status missing %s: %s", want, disconnectBody)
		}
	}

	mu.Lock()
	defer mu.Unlock()
	for _, want := range []string{"getMe", "getWebhookInfo", "deleteWebhook", "getUpdates", "sendMessage"} {
		if !containsTelegramMethod(methods, want) {
			t.Fatalf("Telegram Bot API method %q not called: %v", want, methods)
		}
	}
}

func TestTelegramLocalConnectorReportsRevokedTokenWithoutPersistingOrReturningIt(t *testing.T) {
	const token = "123456:revoked-runtime-secret"
	telegram := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"ok":false,"error_code":401,"description":"Unauthorized ` + token + `"}`))
	}))
	defer telegram.Close()
	t.Setenv("CABINET_E2E_MODE", "1")
	t.Setenv("CABINET_TELEGRAM_TEST_API_BASE_URL", telegram.URL)

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Telegram Revoked Token"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var created struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&created); err != nil {
		t.Fatalf("decode profile: %v", err)
	}

	response := doRequest(t, a, http.MethodPost, "/api/telegram/connection/test", strings.NewReader(`{"profile_id":"`+created.ID+`","bot_token":"`+token+`"}`), map[string]string{"Content-Type": "application/json"})
	if response.Code != http.StatusBadRequest {
		t.Fatalf("revoked token status=%d body=%s", response.Code, response.Body.String())
	}
	body := response.Body.String()
	if !strings.Contains(body, `"code":"TELEGRAM_BOT_TOKEN_INVALID"`) || !strings.Contains(body, `"next_action":"review_bot_token"`) {
		t.Fatalf("revoked token response is not actionable: %s", body)
	}
	if strings.Contains(body, token) {
		t.Fatalf("revoked token response leaked token: %s", body)
	}

	status := doRequest(t, a, http.MethodGet, "/api/telegram/connection/status?profile_id="+created.ID, nil, nil)
	if status.Code != http.StatusOK || !strings.Contains(status.Body.String(), `"bot_token_present":false`) {
		t.Fatalf("revoked token was persisted: status=%d body=%s", status.Code, status.Body.String())
	}
}

func TestTelegramLocalConnectorRotationToDifferentBotClearsPairingAndOffset(t *testing.T) {
	const oldToken = "100:old-bot-secret"
	const newToken = "200:new-bot-secret"
	telegram := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		botID := 100
		username := "old_cabinet_bot"
		if strings.HasPrefix(r.URL.Path, "/bot"+newToken+"/") {
			botID = 200
			username = "new_cabinet_bot"
		} else if !strings.HasPrefix(r.URL.Path, "/bot"+oldToken+"/") {
			t.Fatalf("unexpected token-bound Telegram path %q", r.URL.Path)
		}
		switch {
		case strings.HasSuffix(r.URL.Path, "/getMe"):
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "result": map[string]any{"id": botID, "is_bot": true, "username": username}})
		case strings.HasSuffix(r.URL.Path, "/getWebhookInfo"):
			_, _ = w.Write([]byte(`{"ok":true,"result":{"url":"","pending_update_count":0}}`))
		default:
			t.Fatalf("unexpected Telegram method path %q", r.URL.Path)
		}
	}))
	defer telegram.Close()
	t.Setenv("CABINET_E2E_MODE", "1")
	t.Setenv("CABINET_TELEGRAM_TEST_API_BASE_URL", telegram.URL)

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Telegram Rotation"}`), map[string]string{"Content-Type": "application/json"})
	var created struct {
		ID string `json:"id"`
	}
	if create.Code != http.StatusCreated || json.NewDecoder(create.Body).Decode(&created) != nil {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	response := doRequest(t, a, http.MethodPost, "/api/telegram/connection/test", strings.NewReader(`{"profile_id":"`+created.ID+`","bot_token":"`+oldToken+`"}`), map[string]string{"Content-Type": "application/json"})
	if response.Code != http.StatusOK {
		t.Fatalf("initial token test status=%d body=%s", response.Code, response.Body.String())
	}
	if _, err := a.db.ExecContext(context.Background(), `INSERT INTO profile_settings(profile_id, key, value) VALUES(?, 'telegram.catalog_capture.sender_id', '4444'), (?, 'telegram.catalog_capture.chat_id', '4444') ON CONFLICT(profile_id, key) DO UPDATE SET value=excluded.value`, created.ID, created.ID); err != nil {
		t.Fatalf("persist pairing mapping: %v", err)
	}
	if _, err := a.db.ExecContext(context.Background(), `INSERT INTO telegram_connector_state(profile_id, update_offset) VALUES(?, 7002)`, created.ID); err != nil {
		t.Fatalf("seed connector offset: %v", err)
	}

	rotate := doRequest(t, a, http.MethodPost, "/api/telegram/connection/test", strings.NewReader(`{"profile_id":"`+created.ID+`","bot_token":"`+newToken+`"}`), map[string]string{"Content-Type": "application/json"})
	if rotate.Code != http.StatusOK {
		t.Fatalf("rotated token status=%d body=%s", rotate.Code, rotate.Body.String())
	}
	status := doRequest(t, a, http.MethodGet, "/api/telegram/connection/status?profile_id="+created.ID, nil, nil)
	if status.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", status.Code, status.Body.String())
	}
	body := status.Body.String()
	for _, want := range []string{`"bot_username":"new_cabinet_bot"`, `"paired":false`, `"offset":0`, `"next_action":"create_pairing_code"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("rotation status missing %s: %s", want, body)
		}
	}
}

func TestTelegramLocalConnectorDispatchesNaturalAgentPreviewAndCallbackOnce(t *testing.T) {
	const token = "2086:natural-agent-fixture"
	var mu sync.Mutex
	previewID := ""
	methods := map[string]int{}
	telegram := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method := strings.TrimPrefix(r.URL.Path, "/bot"+token+"/")
		mu.Lock()
		methods[method]++
		currentPreview := previewID
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		switch method {
		case "getMe":
			_, _ = w.Write([]byte(`{"ok":true,"result":{"id":2086,"is_bot":true,"username":"cabinet_agent_fixture_bot"}}`))
		case "getWebhookInfo":
			_, _ = w.Write([]byte(`{"ok":true,"result":{"url":"","pending_update_count":0}}`))
		case "getUpdates":
			var body struct {
				Offset int64 `json:"offset"`
			}
			_ = json.NewDecoder(r.Body).Decode(&body)
			switch {
			case body.Offset <= 8101:
				_, _ = w.Write([]byte(`{"ok":true,"result":[{"update_id":8101,"message":{"message_id":501,"from":{"id":4444},"chat":{"id":4444,"type":"private"},"text":"add TG-E2E-2086 to inventory"}}]}`))
			case body.Offset == 8102 && currentPreview != "":
				_, _ = w.Write([]byte(`{"ok":true,"result":[{"update_id":8102,"callback_query":{"id":"callback-8102","from":{"id":4444},"message":{"message_id":502,"chat":{"id":4444,"type":"private"}},"data":"` + currentPreview + `:apply"}}]}`))
			default:
				_, _ = w.Write([]byte(`{"ok":true,"result":[]}`))
			}
		case "sendMessage", "answerCallbackQuery", "editMessageText":
			_, _ = w.Write([]byte(`{"ok":true,"result":true}`))
		default:
			t.Fatalf("unexpected Telegram Bot API method %q", method)
		}
	}))
	defer telegram.Close()
	t.Setenv("CABINET_E2E_MODE", "1")
	t.Setenv("CABINET_TELEGRAM_TEST_API_BASE_URL", telegram.URL)

	a := newTestApp(t)
	profileID := createTelegramConversationProfile(t, a, "Telegram natural connector")
	connected := doRequest(t, a, http.MethodPost, "/api/telegram/connection/test", strings.NewReader(`{"profile_id":"`+profileID+`","bot_token":"`+token+`"}`), map[string]string{"Content-Type": "application/json"})
	if connected.Code != http.StatusOK {
		t.Fatalf("connect fixture bot status=%d body=%s", connected.Code, connected.Body.String())
	}
	if err := profile.NewRepository(a.db).PutSettings(context.Background(), profileID, map[string]string{
		"telegram.catalog_capture.sender_id": "4444",
		"telegram.catalog_capture.chat_id":   "4444",
		"telegram.polling.paused":            "false",
		"integration.telegram.enabled":       "true",
	}); err != nil {
		t.Fatalf("pair fixture peer: %v", err)
	}
	if _, err := a.telegramConnector.PollOnce(context.Background(), profileID); err != nil {
		t.Fatalf("poll natural Agent text: %v", err)
	}
	var createdPreviewID string
	if err := a.db.QueryRow(`SELECT id FROM agent_skill_previews WHERE profile_id = ? AND source_channel = 'telegram'`, profileID).Scan(&createdPreviewID); err != nil {
		t.Fatalf("natural text did not create Telegram durable preview: %v", err)
	}
	mu.Lock()
	previewID = createdPreviewID
	mu.Unlock()
	if !strings.HasPrefix(previewID, "asp_") {
		t.Fatalf("natural connector preview is not opaque: %q", previewID)
	}
	if _, err := a.telegramConnector.PollOnce(context.Background(), profileID); err != nil {
		t.Fatalf("poll Agent callback: %v", err)
	}
	assertTelegramItemCount(t, a, profileID, "TG-E2E-2086", 1)
	if _, err := a.telegramConnector.PollOnce(context.Background(), profileID); err != nil {
		t.Fatalf("poll duplicate boundary: %v", err)
	}
	assertTelegramItemCount(t, a, profileID, "TG-E2E-2086", 1)

	mu.Lock()
	defer mu.Unlock()
	if methods["sendMessage"] != 1 || methods["answerCallbackQuery"] != 1 || methods["editMessageText"] != 1 {
		t.Fatalf("expected one natural reply and one callback ack/edit, methods=%+v", methods)
	}
}

func containsTelegramMethod(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
