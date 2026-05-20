package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestTelegramCatalogCaptureAPIRequiresPersistedSenderAuthorization(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Telegram API"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}

	unauthorized := doRequest(t, a, http.MethodPost, "/api/telegram/catalog-captures", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"sender_id":"sender-1",
		"chat_id":"chat-1",
		"barcode":"9312345678901",
		"draft":{"title":"Unauthorized draft"}
	}`), map[string]string{"Content-Type": "application/json"})
	if unauthorized.Code != http.StatusForbidden {
		t.Fatalf("unauthorized status=%d body=%s", unauthorized.Code, unauthorized.Body.String())
	}

	settings := doRequest(t, a, http.MethodPut, "/api/profiles/"+p.ID+"/settings", strings.NewReader(`{
		"settings":{
			"telegram.catalog_capture.sender_id":"sender-1",
			"telegram.catalog_capture.chat_id":"chat-1"
		}
	}`), map[string]string{"Content-Type": "application/json"})
	if settings.Code != http.StatusOK {
		t.Fatalf("settings status=%d body=%s", settings.Code, settings.Body.String())
	}

	capture := doRequest(t, a, http.MethodPost, "/api/telegram/catalog-captures", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"sender_id":"sender-1",
		"chat_id":"chat-1",
		"message_id":"message-42",
		"text":"Box says limited release.",
		"barcode":"9312345678901",
		"grouping_hint":"same-item",
		"draft":{"title":"Limited Release Model","brand":"AFX","category":"Die-cast"},
		"media":[{
			"file_id":"telegram-file-photo-1",
			"filename":"front.jpg",
			"mime_type":"image/jpeg",
			"kind":"photo",
			"content_base64":"ZnJvbnQtaW1hZ2UtYnl0ZXM="
		}]
	}`), map[string]string{"Content-Type": "application/json"})
	if capture.Code != http.StatusCreated {
		t.Fatalf("capture status=%d body=%s", capture.Code, capture.Body.String())
	}
	if !strings.Contains(capture.Body.String(), `"source":"telegram_catalog_capture"`) ||
		!strings.Contains(capture.Body.String(), `"confirmation_state":"preview_required"`) ||
		!strings.Contains(capture.Body.String(), `"telegram_reply"`) ||
		!strings.Contains(capture.Body.String(), `"review_url":"/chats?profile_id=`) ||
		!strings.Contains(capture.Body.String(), `"confirm_in_cabinet"`) ||
		!strings.Contains(capture.Body.String(), `"source_message_id":"message-42"`) ||
		!strings.Contains(capture.Body.String(), `"filename":"front.jpg"`) {
		t.Fatalf("expected capture response to include audit metadata and Telegram confirmation handoff, body=%s", capture.Body.String())
	}

	items := doRequest(t, a, http.MethodGet, "/api/items?profile_id="+p.ID, nil, nil)
	if items.Code != http.StatusOK {
		t.Fatalf("items status=%d body=%s", items.Code, items.Body.String())
	}
	if strings.Contains(items.Body.String(), "Limited Release Model") || strings.Contains(items.Body.String(), "9312345678901") {
		t.Fatalf("telegram API capture must not create catalog item before confirmation, body=%s", items.Body.String())
	}
}
