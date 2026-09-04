package app

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/collectors-tech/cabinet/internal/profile"
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
			"content_base64":"`+base64.StdEncoding.EncodeToString(sampleJPEG(t))+`"
		}]
	}`), map[string]string{"Content-Type": "application/json"})
	if capture.Code != http.StatusCreated {
		t.Fatalf("capture status=%d body=%s", capture.Code, capture.Body.String())
	}
	if !strings.Contains(capture.Body.String(), `"source":"telegram_catalog_capture"`) ||
		!strings.Contains(capture.Body.String(), `"confirmation_state":"preview_required"`) ||
		!strings.Contains(capture.Body.String(), `"workflow_run"`) ||
		!strings.Contains(capture.Body.String(), `"capability_id":"catalog_add_from_photo"`) ||
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

	var capturePayload struct {
		Thread struct {
			ID string `json:"id"`
		} `json:"thread"`
		Preview struct {
			ID string `json:"id"`
		} `json:"preview"`
		WorkflowRun struct {
			ID                string           `json:"id"`
			CapabilityID      string           `json:"capability_id"`
			SourceChannel     string           `json:"source_channel"`
			SourceMessageID   string           `json:"source_message_id"`
			Status            string           `json:"status"`
			ConfirmationState string           `json:"confirmation_state"`
			ProviderTrace     map[string]any   `json:"provider_trace"`
			Result            map[string]any   `json:"result"`
			BulkItems         []map[string]any `json:"bulk_items"`
		} `json:"workflow_run"`
	}
	if err := json.NewDecoder(capture.Body).Decode(&capturePayload); err != nil {
		t.Fatalf("decode capture payload: %v", err)
	}
	if capturePayload.WorkflowRun.ID == "" || capturePayload.WorkflowRun.Status != "completed" || capturePayload.WorkflowRun.ConfirmationState != "pending" {
		t.Fatalf("expected completed pending-confirmation workflow run, got %+v", capturePayload.WorkflowRun)
	}
	if capturePayload.WorkflowRun.CapabilityID != "catalog_add_from_photo" || capturePayload.WorkflowRun.SourceChannel != "telegram" || capturePayload.WorkflowRun.SourceMessageID != "message-42" {
		t.Fatalf("expected Telegram capability/source workflow audit fields, got %+v", capturePayload.WorkflowRun)
	}
	if capturePayload.WorkflowRun.ProviderTrace["live_provider"] != false || capturePayload.WorkflowRun.ProviderTrace["mode"] != "governed_preview_before_apply" {
		t.Fatalf("expected non-live-provider preview trace, got %+v", capturePayload.WorkflowRun.ProviderTrace)
	}
	if capturePayload.WorkflowRun.Result["preview_id"] != capturePayload.Preview.ID || len(capturePayload.WorkflowRun.BulkItems) != 1 {
		t.Fatalf("expected preview-linked workflow result and media bulk item, got %+v", capturePayload.WorkflowRun)
	}

	runs := doRequest(t, a, http.MethodGet, "/api/chat/workflow-runs?profile_id="+p.ID+"&thread_id="+capturePayload.Thread.ID, nil, nil)
	if runs.Code != http.StatusOK {
		t.Fatalf("list workflow runs status=%d body=%s", runs.Code, runs.Body.String())
	}
	if !strings.Contains(runs.Body.String(), capturePayload.WorkflowRun.ID) ||
		!strings.Contains(runs.Body.String(), `"source_channel":"telegram"`) ||
		!strings.Contains(runs.Body.String(), `"preview_id":"`+capturePayload.Preview.ID+`"`) {
		t.Fatalf("expected Telegram intake workflow run to be queryable by thread, body=%s", runs.Body.String())
	}
}

func TestTelegramAgentTextRejectsLegacySkillGrammar(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Telegram Agent Text"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}

	unauthorized := doRequest(t, a, http.MethodPost, "/api/telegram/agent-text", strings.NewReader(`{
		"sender_id":"sender-agent",
		"chat_id":"chat-agent",
		"message_id":"agent-message-unauthorized",
		"text":"show me my inventory",
		"skill_id":"cabinet.inventory.search_items",
		"parameters":{"query":"AFX"}
	}`), map[string]string{"Content-Type": "application/json"})
	if unauthorized.Code != http.StatusForbidden {
		t.Fatalf("unauthorized status=%d body=%s", unauthorized.Code, unauthorized.Body.String())
	}

	settings := doRequest(t, a, http.MethodPut, "/api/profiles/"+p.ID+"/settings", strings.NewReader(`{
		"settings":{
			"telegram.catalog_capture.sender_id":"agent-peer",
			"telegram.catalog_capture.chat_id":"agent-peer"
		}
	}`), map[string]string{"Content-Type": "application/json"})
	if settings.Code != http.StatusOK {
		t.Fatalf("settings status=%d body=%s", settings.Code, settings.Body.String())
	}

	readOnly := doRequest(t, a, http.MethodPost, "/api/telegram/agent-text", strings.NewReader(`{
		"sender_id":"agent-peer",
		"chat_id":"agent-peer",
		"chat_type":"private",
		"message_id":"agent-message-read",
		"text":"show me my AFX inventory",
		"skill_id":"cabinet.inventory.search_items",
		"parameters":{"query":"AFX"},
		"source_metadata":{"update_id":1705}
	}`), map[string]string{"Content-Type": "application/json"})
	if readOnly.Code != http.StatusForbidden {
		t.Fatalf("legacy grammar status=%d body=%s", readOnly.Code, readOnly.Body.String())
	}
	if !strings.Contains(readOnly.Body.String(), `"telegram_agent_request_rejected"`) {
		t.Fatalf("expected raw skill grammar to fail closed, body=%s", readOnly.Body.String())
	}
	return

	mutating := doRequest(t, a, http.MethodPost, "/api/telegram/agent-text", strings.NewReader(`{
		"sender_id":"sender-agent",
		"chat_id":"chat-agent",
		"message_id":"agent-message-create",
		"text":"create an inventory item for AFX truck",
		"skill_id":"cabinet.inventory.create_item",
		"parameters":{"title":"AFX Telegram Truck","part_number":"TG-1705","brand":"AFX","category":"Slot Cars"},
		"media":[{
			"file_id":"telegram-agent-photo-1",
			"file_unique_id":"telegram-agent-unique-1",
			"file_size":4096,
			"filename":"telegram-agent-front.jpg",
			"mime_type":"image/jpeg",
			"kind":"photo"
		}]
	}`), map[string]string{"Content-Type": "application/json"})
	if mutating.Code != http.StatusCreated {
		t.Fatalf("mutating status=%d body=%s", mutating.Code, mutating.Body.String())
	}
	mutateBody := mutating.Body.String()
	if !strings.Contains(mutateBody, `"confirmation_required":true`) ||
		!strings.Contains(mutateBody, `"confirmation_state":"preview_required"`) ||
		!strings.Contains(mutateBody, `"source":"telegram_agent_text"`) ||
		!strings.Contains(mutateBody, `"review_url":"/chats?`) ||
		!strings.Contains(mutateBody, `preview_id`) ||
		!strings.Contains(mutateBody, `"file_id":"telegram-agent-photo-1"`) ||
		!strings.Contains(mutateBody, `"file_unique_id":"telegram-agent-unique-1"`) ||
		!strings.Contains(mutateBody, `"media_count":1`) ||
		!strings.Contains(mutateBody, `"mutation_applied":false`) {
		t.Fatalf("expected mutating Telegram Agent text to create reviewable preview without apply, body=%s", mutateBody)
	}
	var mutatePayload struct {
		Thread struct {
			ID string `json:"id"`
		} `json:"thread"`
		ActionPreview struct {
			ID      string         `json:"id"`
			Action  string         `json:"action"`
			Status  string         `json:"status"`
			Payload map[string]any `json:"payload"`
		} `json:"action_preview"`
		WorkflowRun struct {
			Result map[string]any `json:"result"`
		} `json:"workflow_run"`
		InboxItem struct {
			Metadata map[string]any `json:"metadata"`
		} `json:"inbox_item"`
		TelegramReply struct {
			ReviewURL string `json:"review_url"`
		} `json:"telegram_reply"`
	}
	if err := json.Unmarshal([]byte(mutateBody), &mutatePayload); err != nil {
		t.Fatalf("decode mutating Telegram Agent text: %v", err)
	}
	if mutatePayload.ActionPreview.ID == "" ||
		mutatePayload.ActionPreview.Action != "create_inventory_item" ||
		mutatePayload.ActionPreview.Status != "previewed" ||
		mutatePayload.ActionPreview.Payload["title"] != "AFX Telegram Truck" {
		t.Fatalf("expected durable action preview for Telegram Agent review handoff, got %+v", mutatePayload.ActionPreview)
	}
	if !strings.Contains(mutatePayload.TelegramReply.ReviewURL, "preview_id="+mutatePayload.ActionPreview.ID) ||
		mutatePayload.WorkflowRun.Result["preview_id"] != mutatePayload.ActionPreview.ID ||
		mutatePayload.InboxItem.Metadata["preview_id"] != mutatePayload.ActionPreview.ID {
		t.Fatalf("expected workflow, Inbox, and Telegram review URL to share preview id %q, workflow=%+v inbox=%+v reply=%+v", mutatePayload.ActionPreview.ID, mutatePayload.WorkflowRun.Result, mutatePayload.InboxItem.Metadata, mutatePayload.TelegramReply)
	}
	runs := doRequest(t, a, http.MethodGet, "/api/chat/workflow-runs?profile_id="+p.ID+"&thread_id="+mutatePayload.Thread.ID, nil, nil)
	if runs.Code != http.StatusOK {
		t.Fatalf("workflow runs status=%d body=%s", runs.Code, runs.Body.String())
	}
	if !strings.Contains(runs.Body.String(), mutatePayload.ActionPreview.ID) ||
		!strings.Contains(runs.Body.String(), `"source_channel":"telegram"`) ||
		!strings.Contains(runs.Body.String(), `"capability_id":"cabinet.inventory.create_item"`) ||
		!strings.Contains(runs.Body.String(), `"file_id":"telegram-agent-photo-1"`) {
		t.Fatalf("expected queryable Telegram Agent workflow proof with preview id, body=%s", runs.Body.String())
	}
	items := doRequest(t, a, http.MethodGet, "/api/items?profile_id="+p.ID, nil, nil)
	if items.Code != http.StatusOK {
		t.Fatalf("items status=%d body=%s", items.Code, items.Body.String())
	}
	if strings.Contains(items.Body.String(), "AFX Telegram Truck") || strings.Contains(items.Body.String(), "TG-1705") {
		t.Fatalf("telegram Agent text must not mutate inventory before confirmation, body=%s", items.Body.String())
	}
	wrongCallback := doRequest(t, a, http.MethodPost, "/api/telegram/agent-text-callbacks", strings.NewReader(`{
		"sender_id":"sender-agent",
		"chat_id":"wrong-chat-agent",
		"message_id":"agent-message-confirm-wrong",
		"thread_id":"`+mutatePayload.Thread.ID+`",
		"preview_id":"`+mutatePayload.ActionPreview.ID+`",
		"confirmation":"confirm"
	}`), map[string]string{"Content-Type": "application/json"})
	if wrongCallback.Code != http.StatusForbidden {
		t.Fatalf("wrong callback status=%d body=%s", wrongCallback.Code, wrongCallback.Body.String())
	}
	confirmed := doRequest(t, a, http.MethodPost, "/api/telegram/agent-text-callbacks", strings.NewReader(`{
		"sender_id":"sender-agent",
		"chat_id":"chat-agent",
		"message_id":"agent-message-confirm",
		"thread_id":"`+mutatePayload.Thread.ID+`",
		"preview_id":"`+mutatePayload.ActionPreview.ID+`",
		"confirmation":"confirm",
		"callback_data":"cabinet:agent_text:confirm:`+mutatePayload.ActionPreview.ID+`"
	}`), map[string]string{"Content-Type": "application/json"})
	if confirmed.Code != http.StatusOK {
		t.Fatalf("confirmed callback status=%d body=%s", confirmed.Code, confirmed.Body.String())
	}
	confirmedBody := confirmed.Body.String()
	if !strings.Contains(confirmedBody, `"confirmation_state":"confirmed"`) ||
		!strings.Contains(confirmedBody, `"mutation_applied":true`) ||
		!strings.Contains(confirmedBody, `"preview_id":"`+mutatePayload.ActionPreview.ID+`"`) ||
		!strings.Contains(confirmedBody, `"source_channel":"telegram"`) ||
		!strings.Contains(confirmedBody, `"workflow_id":"telegram-agent-text-callback"`) {
		t.Fatalf("expected Telegram Agent callback confirmation proof, body=%s", confirmedBody)
	}
	items = doRequest(t, a, http.MethodGet, "/api/items?profile_id="+p.ID, nil, nil)
	if items.Code != http.StatusOK {
		t.Fatalf("items after confirm status=%d body=%s", items.Code, items.Body.String())
	}
	if !strings.Contains(items.Body.String(), "AFX Telegram Truck") || !strings.Contains(items.Body.String(), "TG-1705") {
		t.Fatalf("confirmed Telegram Agent callback should apply inventory mutation, body=%s", items.Body.String())
	}
	runsAfterConfirm := doRequest(t, a, http.MethodGet, "/api/chat/workflow-runs?profile_id="+p.ID+"&thread_id="+mutatePayload.Thread.ID, nil, nil)
	if runsAfterConfirm.Code != http.StatusOK {
		t.Fatalf("workflow runs after confirm status=%d body=%s", runsAfterConfirm.Code, runsAfterConfirm.Body.String())
	}
	if !strings.Contains(runsAfterConfirm.Body.String(), `"workflow_id":"telegram-agent-text-callback"`) ||
		!strings.Contains(runsAfterConfirm.Body.String(), `"confirmation_state":"confirmed"`) ||
		!strings.Contains(runsAfterConfirm.Body.String(), `"source_message_id":"agent-message-confirm"`) {
		t.Fatalf("expected queryable Telegram Agent callback workflow proof, body=%s", runsAfterConfirm.Body.String())
	}
}

func TestTelegramAgentTextRejectsConfirmedLegacyMutationGrammar(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Telegram Agent Read Only"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	repo := profile.NewRepository(a.db)
	if err := repo.PutSettings(context.Background(), p.ID, map[string]string{
		"telegram.catalog_capture.sender_id": "read-only-agent",
		"telegram.catalog_capture.chat_id":   "read-only-agent",
	}); err != nil {
		t.Fatalf("authorize telegram sender: %v", err)
	}
	if _, err := repo.PutAgentAuthorityPolicy(context.Background(), p.ID, profile.AgentAuthorityPolicy{
		Mode: profile.AgentAuthorityModeReadOnly,
	}); err != nil {
		t.Fatalf("set read-only authority policy: %v", err)
	}

	blocked := doRequest(t, a, http.MethodPost, "/api/telegram/agent-text", strings.NewReader(`{
		"sender_id":"read-only-agent",
		"chat_id":"read-only-agent",
		"chat_type":"private",
		"message_id":"agent-read-only-create",
		"text":"create an inventory item even though the profile is read only",
		"skill_id":"cabinet.inventory.create_item",
		"confirm":true,
		"parameters":{"title":"Blocked Telegram Authority Item","part_number":"TG-RO-1932"}
	}`), map[string]string{"Content-Type": "application/json"})
	if blocked.Code != http.StatusForbidden {
		t.Fatalf("blocked telegram authority status=%d body=%s", blocked.Code, blocked.Body.String())
	}
	if !strings.Contains(blocked.Body.String(), `"telegram_agent_request_rejected"`) {
		t.Fatalf("expected confirmed legacy mutation grammar to fail closed, body=%s", blocked.Body.String())
	}
	var itemCount int
	if err := a.db.QueryRow(`SELECT COUNT(1) FROM canonical_items WHERE profile_id = ? AND part_number = 'TG-RO-1932'`, p.ID).Scan(&itemCount); err != nil {
		t.Fatalf("count blocked telegram item: %v", err)
	}
	if itemCount != 0 {
		t.Fatalf("read-only Telegram authority must not create inventory item, got %d", itemCount)
	}
}

func TestTelegramAgentTextUnauthorizedSenderCreatesSetupInboxEvent(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Telegram Agent Auth Inbox"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}

	unauthorized := doRequest(t, a, http.MethodPost, "/api/telegram/agent-text", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"sender_id":"sender-agent-unapproved",
		"chat_id":"chat-agent-unapproved",
		"message_id":"agent-message-unauthorized-inbox",
		"text":"show me my inventory",
		"skill_id":"cabinet.inventory.search_items",
		"parameters":{"query":"AFX"}
	}`), map[string]string{"Content-Type": "application/json"})
	if unauthorized.Code != http.StatusForbidden {
		t.Fatalf("unauthorized status=%d body=%s", unauthorized.Code, unauthorized.Body.String())
	}

	inbox := doRequest(t, a, http.MethodGet, "/api/chat/inbox?profile_id="+p.ID, nil, nil)
	if inbox.Code != http.StatusOK {
		t.Fatalf("inbox status=%d body=%s", inbox.Code, inbox.Body.String())
	}
	body := inbox.Body.String()
	for _, want := range []string{
		`"source":"provider_workflow"`,
		`"provider_id":"telegram"`,
		`"workflow_action_id":"telegram.agent_text"`,
		`"required_action_code":"authorize_sender_chat"`,
		`"target_route":"/integrations"`,
		`"source_message_id":"agent-message-unauthorized-inbox"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected unauthorized Telegram Agent text to create setup Inbox evidence %q, body=%s", want, body)
		}
	}
}

func TestTelegramAgentTextAuthorizedRouteResolvesSetupInboxEvent(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Telegram Agent Auth Inbox Resolution"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}

	unauthorized := doRequest(t, a, http.MethodPost, "/api/telegram/agent-text", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"sender_id":"sender-agent-resolution",
		"chat_id":"chat-agent-resolution",
		"message_id":"agent-message-unauthorized-resolution",
		"text":"show me my inventory",
		"skill_id":"cabinet.inventory.search_items",
		"parameters":{"query":"AFX"}
	}`), map[string]string{"Content-Type": "application/json"})
	if unauthorized.Code != http.StatusForbidden {
		t.Fatalf("unauthorized status=%d body=%s", unauthorized.Code, unauthorized.Body.String())
	}

	settings := doRequest(t, a, http.MethodPut, "/api/profiles/"+p.ID+"/settings", strings.NewReader(`{
		"settings":{
			"telegram.catalog_capture.sender_id":"agent-resolution",
			"telegram.catalog_capture.chat_id":"agent-resolution"
		}
	}`), map[string]string{"Content-Type": "application/json"})
	if settings.Code != http.StatusOK {
		t.Fatalf("settings status=%d body=%s", settings.Code, settings.Body.String())
	}

	authorized := doRequest(t, a, http.MethodPost, "/api/telegram/agent-text", strings.NewReader(`{
		"sender_id":"agent-resolution",
		"chat_id":"agent-resolution",
		"chat_type":"private",
		"message_id":"agent-message-authorized-resolution",
		"text":"show me my AFX inventory"
	}`), map[string]string{"Content-Type": "application/json"})
	if authorized.Code != http.StatusCreated {
		t.Fatalf("authorized status=%d body=%s", authorized.Code, authorized.Body.String())
	}

	inbox := doRequest(t, a, http.MethodGet, "/api/chat/inbox?profile_id="+p.ID, nil, nil)
	if inbox.Code != http.StatusOK {
		t.Fatalf("inbox status=%d body=%s", inbox.Code, inbox.Body.String())
	}
	body := inbox.Body.String()
	for _, want := range []string{
		`"source":"provider_workflow"`,
		`"status":"read"`,
		`"workflow_action_id":"telegram.agent_text"`,
		`"required_action_code":"authorize_sender_chat"`,
		`"resolution":"agent_text_authorized"`,
		`"workflow_result":"routed"`,
		`"sender_authorized":true`,
		`"resolved_by_message":"agent-message-authorized-resolution"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected authorized Telegram Agent text to resolve setup Inbox evidence %q, body=%s", want, body)
		}
	}
	if strings.Contains(body, "bot_token") {
		t.Fatalf("resolved Telegram Agent text Inbox metadata must not expose token material, body=%s", body)
	}
}

func TestTelegramExternalIntakeProofRequiresAuthorizedProviderEvidence(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Telegram Proof API"}`), map[string]string{"Content-Type": "application/json"})
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
			"telegram.catalog_capture.sender_id":"proof-sender",
			"telegram.catalog_capture.chat_id":"proof-chat"
		}
	}`), map[string]string{"Content-Type": "application/json"})
	if settings.Code != http.StatusOK {
		t.Fatalf("settings status=%d body=%s", settings.Code, settings.Body.String())
	}

	capture := doRequest(t, a, http.MethodPost, "/api/telegram/catalog-captures", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"sender_id":"proof-sender",
		"chat_id":"proof-chat",
		"message_id":"proof-message-1",
		"text":"Please draft from this proof packet.",
		"barcode":"9312345678999",
		"draft":{"part_number":"PROOF-001","title":"Proof Packet Draft","brand":"AFX","category":"Slot Cars"}
	}`), map[string]string{"Content-Type": "application/json"})
	if capture.Code != http.StatusCreated {
		t.Fatalf("capture status=%d body=%s", capture.Code, capture.Body.String())
	}
	var capturePayload struct {
		Thread struct {
			ID string `json:"id"`
		} `json:"thread"`
		Preview struct {
			ID string `json:"id"`
		} `json:"preview"`
	}
	if err := json.NewDecoder(capture.Body).Decode(&capturePayload); err != nil {
		t.Fatalf("decode capture payload: %v", err)
	}

	missingProof := doRequest(t, a, http.MethodPost, "/api/telegram/external-intake-proofs", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"sender_id":"proof-sender",
		"chat_id":"proof-chat",
		"source_thread_id":"`+capturePayload.Thread.ID+`",
		"source_message_id":"proof-message-1",
		"capability_id":"catalog_add_from_text",
		"preview_id":"`+capturePayload.Preview.ID+`",
		"confirmation_state":"pending",
		"provider_trace":{"provider":"openai","live_provider":true}
	}`), map[string]string{"Content-Type": "application/json"})
	if missingProof.Code != http.StatusBadRequest {
		t.Fatalf("missing proof status=%d body=%s", missingProof.Code, missingProof.Body.String())
	}
	if !strings.Contains(missingProof.Body.String(), `"error":"invalid_external_intake_proof"`) ||
		!strings.Contains(missingProof.Body.String(), `"missing_fields"`) {
		t.Fatalf("expected invalid proof response with missing fields, body=%s", missingProof.Body.String())
	}

	proof := doRequest(t, a, http.MethodPost, "/api/telegram/external-intake-proofs", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"sender_id":"proof-sender",
		"chat_id":"proof-chat",
		"source_thread_id":"`+capturePayload.Thread.ID+`",
		"source_message_id":"proof-message-1",
		"capability_id":"catalog_add_from_text",
		"preview_id":"`+capturePayload.Preview.ID+`",
		"confirmation_state":"pending",
		"proof_approved":true,
		"provider_trace":{
			"provider":"openai",
			"live_provider":true,
			"request_id":"req_proof_123",
			"result_id":"result_proof_123",
			"model":"gpt-4o-mini",
			"credential_returned":false
		}
	}`), map[string]string{"Content-Type": "application/json"})
	if proof.Code != http.StatusCreated {
		t.Fatalf("proof status=%d body=%s", proof.Code, proof.Body.String())
	}
	var proofPayload struct {
		WorkflowRun struct {
			ID                string         `json:"id"`
			WorkflowID        string         `json:"workflow_id"`
			CapabilityID      string         `json:"capability_id"`
			SourceChannel     string         `json:"source_channel"`
			SourceThreadID    string         `json:"source_thread_id"`
			SourceMessageID   string         `json:"source_message_id"`
			Status            string         `json:"status"`
			ConfirmationState string         `json:"confirmation_state"`
			ProviderTrace     map[string]any `json:"provider_trace"`
			Result            map[string]any `json:"result"`
		} `json:"workflow_run"`
	}
	if err := json.NewDecoder(proof.Body).Decode(&proofPayload); err != nil {
		t.Fatalf("decode proof payload: %v", err)
	}
	run := proofPayload.WorkflowRun
	if run.ID == "" || run.WorkflowID != "telegram-openai-external-intake-proof" || run.Status != "completed" {
		t.Fatalf("expected completed proof workflow run, got %+v", run)
	}
	if run.CapabilityID != "catalog_add_from_text" || run.SourceChannel != "telegram" || run.SourceThreadID != capturePayload.Thread.ID || run.SourceMessageID != "proof-message-1" {
		t.Fatalf("expected source/capability proof metadata, got %+v", run)
	}
	if run.ProviderTrace["provider"] != "openai" || run.ProviderTrace["live_provider"] != true || run.ProviderTrace["credential_returned"] != false || run.ProviderTrace["request_id"] == "" {
		t.Fatalf("expected non-secret OpenAI provider proof trace, got %+v", run.ProviderTrace)
	}
	if _, leaked := run.ProviderTrace["api_key"]; leaked {
		t.Fatalf("proof trace must not return secret material, got %+v", run.ProviderTrace)
	}
	if run.Result["preview_id"] != capturePayload.Preview.ID || run.Result["proof_packet"] != "authorized_telegram_openai_external_intake" {
		t.Fatalf("expected preview-linked proof result, got %+v", run.Result)
	}

	list := doRequest(t, a, http.MethodGet, "/api/chat/workflow-runs?profile_id="+p.ID+"&thread_id="+capturePayload.Thread.ID, nil, nil)
	if list.Code != http.StatusOK {
		t.Fatalf("list workflow runs status=%d body=%s", list.Code, list.Body.String())
	}
	if !strings.Contains(list.Body.String(), run.ID) ||
		!strings.Contains(list.Body.String(), `"proof_packet":"authorized_telegram_openai_external_intake"`) ||
		strings.Contains(list.Body.String(), "sk-") {
		t.Fatalf("expected queryable non-secret proof run, body=%s", list.Body.String())
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
