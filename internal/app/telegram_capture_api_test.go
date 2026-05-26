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
		!strings.Contains(capture.Body.String(), `"action_buttons"`) ||
		!strings.Contains(capture.Body.String(), `"callback_data":"cabinet:catalog_capture:confirm:`) ||
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

func TestTelegramCatalogCaptureAPIPreservesLookupEvidence(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Telegram Lookup API"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	settings := doRequest(t, a, http.MethodPut, "/api/profiles/"+p.ID+"/settings", strings.NewReader(`{
		"settings":{
			"telegram.catalog_capture.sender_id":"sender-lookup",
			"telegram.catalog_capture.chat_id":"chat-lookup"
		}
	}`), map[string]string{"Content-Type": "application/json"})
	if settings.Code != http.StatusOK {
		t.Fatalf("settings status=%d body=%s", settings.Code, settings.Body.String())
	}

	capture := doRequest(t, a, http.MethodPost, "/api/telegram/catalog-captures", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"sender_id":"sender-lookup",
		"chat_id":"chat-lookup",
		"barcode":"4904810900019",
		"draft":{
			"part_number":"TOMY-001",
			"title":"Lookup-backed Tomy release",
			"brand":"Tomy",
			"category":"Die-cast",
			"lookup_source":"barcode_local",
			"lookup_url":"/api/barcodes/4904810900019",
			"lookup_confidence":"high"
		}
	}`), map[string]string{"Content-Type": "application/json"})
	if capture.Code != http.StatusCreated {
		t.Fatalf("capture status=%d body=%s", capture.Code, capture.Body.String())
	}
	body := capture.Body.String()
	if !strings.Contains(body, `"lookup":{"confidence":"high","source":"barcode_local","url":"/api/barcodes/4904810900019"}`) ||
		!strings.Contains(body, `"part_number":"TOMY-001"`) ||
		!strings.Contains(body, `"title":"Lookup-backed Tomy release"`) {
		t.Fatalf("expected direct API capture to preserve lookup evidence in preview and audit metadata, body=%s", body)
	}
}

func TestTelegramCatalogCaptureCallbackAPIConfirmsAuthorizedPreview(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Telegram Callback"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	settings := doRequest(t, a, http.MethodPut, "/api/profiles/"+p.ID+"/settings", strings.NewReader(`{
		"settings":{
			"telegram.catalog_capture.sender_id":"sender-callback",
			"telegram.catalog_capture.chat_id":"chat-callback"
		}
	}`), map[string]string{"Content-Type": "application/json"})
	if settings.Code != http.StatusOK {
		t.Fatalf("settings status=%d body=%s", settings.Code, settings.Body.String())
	}
	capture := doRequest(t, a, http.MethodPost, "/api/telegram/catalog-captures", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"sender_id":"sender-callback",
		"chat_id":"chat-callback",
		"barcode":"9312345678902",
		"draft":{"title":"Callback API Draft","brand":"AFX","category":"Slot"}
	}`), map[string]string{"Content-Type": "application/json"})
	if capture.Code != http.StatusCreated {
		t.Fatalf("capture status=%d body=%s", capture.Code, capture.Body.String())
	}
	var captured struct {
		Preview struct {
			ID string `json:"id"`
		} `json:"preview"`
	}
	if err := json.NewDecoder(capture.Body).Decode(&captured); err != nil {
		t.Fatalf("decode capture: %v", err)
	}
	if captured.Preview.ID == "" {
		t.Fatalf("expected preview id from capture")
	}
	unauthorized := doRequest(t, a, http.MethodPost, "/api/telegram/catalog-capture-callbacks", strings.NewReader(`{
		"sender_id":"wrong-sender",
		"chat_id":"chat-callback",
		"callback_data":"cabinet:catalog_capture:confirm:`+captured.Preview.ID+`"
	}`), map[string]string{"Content-Type": "application/json"})
	if unauthorized.Code != http.StatusForbidden {
		t.Fatalf("unauthorized callback status=%d body=%s", unauthorized.Code, unauthorized.Body.String())
	}
	confirmed := doRequest(t, a, http.MethodPost, "/api/telegram/catalog-capture-callbacks", strings.NewReader(`{
		"sender_id":"sender-callback",
		"chat_id":"chat-callback",
		"callback_data":"cabinet:catalog_capture:confirm:`+captured.Preview.ID+`"
	}`), map[string]string{"Content-Type": "application/json"})
	if confirmed.Code != http.StatusOK {
		t.Fatalf("confirmed callback status=%d body=%s", confirmed.Code, confirmed.Body.String())
	}
	body := confirmed.Body.String()
	if !strings.Contains(body, `"confirmation_state":"confirmed"`) ||
		!strings.Contains(body, `"applied":true`) ||
		!strings.Contains(body, `"item_id":"`) {
		t.Fatalf("expected callback confirmation to apply preview with Telegram reply, body=%s", body)
	}
	items := doRequest(t, a, http.MethodGet, "/api/items?profile_id="+p.ID, nil, nil)
	if items.Code != http.StatusOK {
		t.Fatalf("items status=%d body=%s", items.Code, items.Body.String())
	}
	if !strings.Contains(items.Body.String(), "Callback API Draft") || !strings.Contains(items.Body.String(), "9312345678902") {
		t.Fatalf("confirmed callback should create catalog item, body=%s", items.Body.String())
	}
}

func TestTelegramCatalogCaptureWebhookAPIResolvesProfileAuthorization(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Telegram Webhook"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}

	settings := doRequest(t, a, http.MethodPut, "/api/profiles/"+p.ID+"/settings", strings.NewReader(`{
		"settings":{
			"telegram.catalog_capture.sender_id":"12345",
			"telegram.catalog_capture.chat_id":"-5235769556"
		}
	}`), map[string]string{"Content-Type": "application/json"})
	if settings.Code != http.StatusOK {
		t.Fatalf("settings status=%d body=%s", settings.Code, settings.Body.String())
	}

	unauthorized := doRequest(t, a, http.MethodPost, "/api/telegram/webhook/catalog-captures", strings.NewReader(`{
		"update_id": 7001,
		"message": {
			"message_id": 42,
			"from": {"id": 99999},
			"chat": {"id": -5235769556},
			"text": "Please draft barcode 4904810900016"
		}
	}`), map[string]string{"Content-Type": "application/json"})
	if unauthorized.Code != http.StatusForbidden {
		t.Fatalf("unauthorized webhook status=%d body=%s", unauthorized.Code, unauthorized.Body.String())
	}

	capture := doRequest(t, a, http.MethodPost, "/api/telegram/webhook/catalog-captures", strings.NewReader(`{
		"update_id": 7002,
		"message": {
			"message_id": 43,
			"date": 1716170000,
			"from": {"id": 12345},
			"chat": {"id": -5235769556},
			"media_group_id": "album-1",
			"caption": "Front image and barcode 4904810900016",
			"photo": [
				{"file_id":"small-photo","file_unique_id":"small","width":90,"height":90,"file_size":1024},
				{"file_id":"large-photo","file_unique_id":"large","width":1280,"height":720,"file_size":2048}
			]
		}
	}`), map[string]string{"Content-Type": "application/json"})
	if capture.Code != http.StatusCreated {
		t.Fatalf("capture status=%d body=%s", capture.Code, capture.Body.String())
	}
	body := capture.Body.String()
	if !strings.Contains(body, `"profile_id":"`+p.ID+`"`) ||
		!strings.Contains(body, `"source":"telegram_catalog_capture"`) ||
		!strings.Contains(body, `"source_message_id":"43"`) ||
		!strings.Contains(body, `"barcode":"4904810900016"`) ||
		!strings.Contains(body, `"filename":"large.jpg"`) ||
		!strings.Contains(body, `"payload_type":"caption+photo"`) ||
		!strings.Contains(body, `"telegram_reply"`) {
		t.Fatalf("expected webhook capture response to include normalized source/profile/media metadata, body=%s", body)
	}

	items := doRequest(t, a, http.MethodGet, "/api/items?profile_id="+p.ID, nil, nil)
	if items.Code != http.StatusOK {
		t.Fatalf("items status=%d body=%s", items.Code, items.Body.String())
	}
	if strings.Contains(items.Body.String(), "Barcode 4904810900016") || strings.Contains(items.Body.String(), "4904810900016") {
		t.Fatalf("telegram webhook capture must not create catalog item before confirmation, body=%s", items.Body.String())
	}
}

func TestTelegramCatalogCaptureWebhookAPIUsesLocalBarcodeLookup(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Telegram Webhook Lookup"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}

	settings := doRequest(t, a, http.MethodPut, "/api/profiles/"+p.ID+"/settings", strings.NewReader(`{
		"settings":{
			"telegram.catalog_capture.sender_id":"12345",
			"telegram.catalog_capture.chat_id":"-5235769556"
		}
	}`), map[string]string{"Content-Type": "application/json"})
	if settings.Code != http.StatusOK {
		t.Fatalf("settings status=%d body=%s", settings.Code, settings.Body.String())
	}
	if _, err := a.db.Exec(`INSERT INTO canonical_items(id, profile_id, brand, category, part_number, title) VALUES (?,?,?,?,?,?)`, "tg-lookup-item", p.ID, "Tomy", "Die-cast", "TOMY-LOOKUP-001", "Lookup Matched Tomy Release"); err != nil {
		t.Fatalf("insert local lookup item: %v", err)
	}
	if _, err := a.db.Exec(`INSERT INTO item_barcodes(id, item_id, barcode) VALUES (?,?,?)`, "tg-lookup-barcode", "tg-lookup-item", "4904810900999"); err != nil {
		t.Fatalf("insert local lookup barcode: %v", err)
	}

	capture := doRequest(t, a, http.MethodPost, "/api/telegram/webhook/catalog-captures", strings.NewReader(`{
		"update_id": 7010,
		"message": {
			"message_id": 50,
			"from": {"id": 12345},
			"chat": {"id": -5235769556},
			"text": "Please draft barcode 4904810900999"
		}
	}`), map[string]string{"Content-Type": "application/json"})
	if capture.Code != http.StatusCreated {
		t.Fatalf("capture status=%d body=%s", capture.Code, capture.Body.String())
	}
	body := capture.Body.String()
	if !strings.Contains(body, `"part_number":"TOMY-LOOKUP-001"`) ||
		!strings.Contains(body, `"title":"Lookup Matched Tomy Release"`) ||
		!strings.Contains(body, `"brand":"Tomy"`) ||
		!strings.Contains(body, `"category":"Die-cast"`) ||
		!strings.Contains(body, `"lookup":{"confidence":"high","source":"barcode_local","url":"/api/barcodes/4904810900999"}`) {
		t.Fatalf("expected webhook barcode capture to use local lookup-backed draft evidence, body=%s", body)
	}

	items := doRequest(t, a, http.MethodGet, "/api/items?profile_id="+p.ID, nil, nil)
	if items.Code != http.StatusOK {
		t.Fatalf("items status=%d body=%s", items.Code, items.Body.String())
	}
	if strings.Count(items.Body.String(), "Lookup Matched Tomy Release") != 1 {
		t.Fatalf("telegram webhook lookup capture must not create a duplicate catalog item before confirmation, body=%s", items.Body.String())
	}
}

func TestTelegramCatalogCaptureWebhookAPIRequestsFollowUpForAmbiguousText(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Telegram Follow Up"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}

	settings := doRequest(t, a, http.MethodPut, "/api/profiles/"+p.ID+"/settings", strings.NewReader(`{
		"settings":{
			"telegram.catalog_capture.sender_id":"12345",
			"telegram.catalog_capture.chat_id":"-5235769556"
		}
	}`), map[string]string{"Content-Type": "application/json"})
	if settings.Code != http.StatusOK {
		t.Fatalf("settings status=%d body=%s", settings.Code, settings.Body.String())
	}

	followUp := doRequest(t, a, http.MethodPost, "/api/telegram/webhook/catalog-captures", strings.NewReader(`{
		"update_id": 7003,
		"message": {
			"message_id": 44,
			"from": {"id": 12345},
			"chat": {"id": -5235769556},
			"text": "This is the blue boxed one from the bench"
		}
	}`), map[string]string{"Content-Type": "application/json"})
	if followUp.Code != http.StatusUnprocessableEntity {
		t.Fatalf("follow-up status=%d body=%s", followUp.Code, followUp.Body.String())
	}
	body := followUp.Body.String()
	if !strings.Contains(body, `"error":"telegram_capture_needs_follow_up"`) ||
		!strings.Contains(body, `"confirmation_state":"follow_up_required"`) ||
		!strings.Contains(body, `"action_buttons"`) ||
		!strings.Contains(body, `"reply_with_barcode"`) ||
		!strings.Contains(body, `"barcode_or_part_number"`) {
		t.Fatalf("expected Telegram-visible follow-up response for ambiguous text, body=%s", body)
	}

	items := doRequest(t, a, http.MethodGet, "/api/items?profile_id="+p.ID, nil, nil)
	if items.Code != http.StatusOK {
		t.Fatalf("items status=%d body=%s", items.Code, items.Body.String())
	}
	if strings.Contains(items.Body.String(), "blue boxed") {
		t.Fatalf("ambiguous Telegram follow-up must not create catalog item, body=%s", items.Body.String())
	}
}
