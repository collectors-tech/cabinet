package app

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"
)

func TestChatAPIsThreadMessageAttachmentAndPreviewApply(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"P1"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}

	threadResp := doRequest(t, a, http.MethodPost, "/api/chat/threads", strings.NewReader(`{"profile_id":"`+p.ID+`","title":"Main Thread","metadata":{"provider":"openai","model":"gpt-4o-mini"}}`), map[string]string{"Content-Type": "application/json"})
	if threadResp.Code != http.StatusCreated {
		t.Fatalf("create thread status=%d body=%s", threadResp.Code, threadResp.Body.String())
	}
	var thread struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(threadResp.Body).Decode(&thread); err != nil {
		t.Fatalf("decode thread: %v", err)
	}
	if strings.TrimSpace(thread.ID) == "" {
		t.Fatalf("expected thread id")
	}

	threadsList := doRequest(t, a, http.MethodGet, "/api/chat/threads?profile_id="+p.ID, nil, nil)
	if threadsList.Code != http.StatusOK {
		t.Fatalf("list threads status=%d body=%s", threadsList.Code, threadsList.Body.String())
	}

	msgResp := doRequest(t, a, http.MethodPost, "/api/chat/messages", strings.NewReader(`{"profile_id":"`+p.ID+`","thread_id":"`+thread.ID+`","role":"user","content":"hello","context":{"route":{"pathname":"/inventory","search":"?tab=all"},"profile":{"id":"`+p.ID+`"},"selection":{"active_workspace_collection":"All Items"},"assistant":{"provider":"openai","model":"gpt-4o-mini"}}}`), map[string]string{"Content-Type": "application/json"})
	if msgResp.Code != http.StatusCreated {
		t.Fatalf("create message status=%d body=%s", msgResp.Code, msgResp.Body.String())
	}
	if !strings.Contains(msgResp.Body.String(), `"pathname":"/inventory"`) {
		t.Fatalf("expected message response to include route context, body=%s", msgResp.Body.String())
	}
	if !strings.Contains(msgResp.Body.String(), `"assistant_handoff"`) {
		t.Fatalf("expected response to include assistant handoff payload, body=%s", msgResp.Body.String())
	}

	msgList := doRequest(t, a, http.MethodGet, "/api/chat/messages?profile_id="+p.ID+"&thread_id="+thread.ID, nil, nil)
	if msgList.Code != http.StatusOK {
		t.Fatalf("list messages status=%d body=%s", msgList.Code, msgList.Body.String())
	}
	if !strings.Contains(msgList.Body.String(), `"active_workspace_collection":"All Items"`) {
		t.Fatalf("expected listed messages to retain selection context, body=%s", msgList.Body.String())
	}
	if !strings.Contains(msgList.Body.String(), `Assistant handoff queued in Inbox.`) {
		t.Fatalf("expected assistant thread to surface queued handoff state, body=%s", msgList.Body.String())
	}

	inboxList := doRequest(t, a, http.MethodGet, "/api/chat/inbox?profile_id="+p.ID, nil, nil)
	if inboxList.Code != http.StatusOK {
		t.Fatalf("list inbox status=%d body=%s", inboxList.Code, inboxList.Body.String())
	}
	if !strings.Contains(inboxList.Body.String(), `"status":"queued"`) || !strings.Contains(inboxList.Body.String(), `"thread_id":"`+thread.ID+`"`) {
		t.Fatalf("expected inbox item with queued assistant linkage, body=%s", inboxList.Body.String())
	}

	// Must reject attachment calls that do not include explicit multipart file input.
	badAttachment := doRequest(t, a, http.MethodPost, "/api/chat/attachments", strings.NewReader(`{"profile_id":"`+p.ID+`","thread_id":"`+thread.ID+`","path":"C:\\secret.txt"}`), map[string]string{"Content-Type": "application/json"})
	if badAttachment.Code != http.StatusBadRequest {
		t.Fatalf("bad attachment status=%d body=%s", badAttachment.Code, badAttachment.Body.String())
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("profile_id", p.ID); err != nil {
		t.Fatalf("write field profile_id: %v", err)
	}
	if err := writer.WriteField("thread_id", thread.ID); err != nil {
		t.Fatalf("write field thread_id: %v", err)
	}
	part, err := writer.CreateFormFile("file", "notes.txt")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write([]byte("sample attachment")); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}
	attachmentResp := doRequest(t, a, http.MethodPost, "/api/chat/attachments", &body, map[string]string{"Content-Type": writer.FormDataContentType()})
	if attachmentResp.Code != http.StatusCreated {
		t.Fatalf("attachment status=%d body=%s", attachmentResp.Code, attachmentResp.Body.String())
	}

	previewResp := doRequest(t, a, http.MethodPost, "/api/chat/actions/preview", strings.NewReader(`{"profile_id":"`+p.ID+`","thread_id":"`+thread.ID+`","action":"create_item_stub","payload":{"part_number":"CHAT-001","title":"Chat Created Item","brand":"AFX","category":"General"}}`), map[string]string{"Content-Type": "application/json"})
	if previewResp.Code != http.StatusOK {
		t.Fatalf("preview status=%d body=%s", previewResp.Code, previewResp.Body.String())
	}
	var preview struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(previewResp.Body).Decode(&preview); err != nil {
		t.Fatalf("decode preview: %v", err)
	}

	itemsBefore := doRequest(t, a, http.MethodGet, "/api/items?profile_id="+p.ID, nil, nil)
	if itemsBefore.Code != http.StatusOK {
		t.Fatalf("items before status=%d body=%s", itemsBefore.Code, itemsBefore.Body.String())
	}
	if strings.Contains(strings.ToLower(itemsBefore.Body.String()), "chat-001") {
		t.Fatalf("item should not exist before apply")
	}

	applyMissingConfirm := doRequest(t, a, http.MethodPost, "/api/chat/actions/apply", strings.NewReader(`{"profile_id":"`+p.ID+`","thread_id":"`+thread.ID+`","preview_id":"`+preview.ID+`","confirm":false}`), map[string]string{"Content-Type": "application/json"})
	if applyMissingConfirm.Code != http.StatusBadRequest {
		t.Fatalf("apply without confirm status=%d body=%s", applyMissingConfirm.Code, applyMissingConfirm.Body.String())
	}

	applyResp := doRequest(t, a, http.MethodPost, "/api/chat/actions/apply", strings.NewReader(`{"profile_id":"`+p.ID+`","thread_id":"`+thread.ID+`","preview_id":"`+preview.ID+`","confirm":true}`), map[string]string{"Content-Type": "application/json"})
	if applyResp.Code != http.StatusOK {
		t.Fatalf("apply status=%d body=%s", applyResp.Code, applyResp.Body.String())
	}

	itemsAfter := doRequest(t, a, http.MethodGet, "/api/items?profile_id="+p.ID, nil, nil)
	if itemsAfter.Code != http.StatusOK {
		t.Fatalf("items after status=%d body=%s", itemsAfter.Code, itemsAfter.Body.String())
	}
	if !strings.Contains(strings.ToLower(itemsAfter.Body.String()), "chat-001") {
		t.Fatalf("expected item created after apply, body=%s", itemsAfter.Body.String())
	}
}

func TestChatInboxItemStatusLifecycle(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Inbox Status"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}

	threadResp := doRequest(t, a, http.MethodPost, "/api/chat/threads", strings.NewReader(`{"profile_id":"`+p.ID+`","title":"Inbox Thread","metadata":{"provider":"openai","model":"gpt-4o-mini"}}`), map[string]string{"Content-Type": "application/json"})
	if threadResp.Code != http.StatusCreated {
		t.Fatalf("create thread status=%d body=%s", threadResp.Code, threadResp.Body.String())
	}
	var thread struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(threadResp.Body).Decode(&thread); err != nil {
		t.Fatalf("decode thread: %v", err)
	}

	msgResp := doRequest(t, a, http.MethodPost, "/api/chat/messages", strings.NewReader(`{"profile_id":"`+p.ID+`","thread_id":"`+thread.ID+`","role":"user","content":"follow up on this item","context":{"assistant":{"provider":"openai","model":"gpt-4o-mini"}}}`), map[string]string{"Content-Type": "application/json"})
	if msgResp.Code != http.StatusCreated {
		t.Fatalf("create message status=%d body=%s", msgResp.Code, msgResp.Body.String())
	}
	var msgPayload struct {
		AssistantHandoff struct {
			InboxItem struct {
				ID     string `json:"id"`
				Status string `json:"status"`
			} `json:"inbox_item"`
		} `json:"assistant_handoff"`
	}
	if err := json.NewDecoder(msgResp.Body).Decode(&msgPayload); err != nil {
		t.Fatalf("decode message payload: %v", err)
	}
	if msgPayload.AssistantHandoff.InboxItem.ID == "" {
		t.Fatalf("expected assistant handoff inbox item id")
	}

	for _, status := range []string{"read", "unread", "archived"} {
		statusResp := doRequest(t, a, http.MethodPatch, "/api/chat/inbox/"+msgPayload.AssistantHandoff.InboxItem.ID, strings.NewReader(`{"profile_id":"`+p.ID+`","status":"`+status+`"}`), map[string]string{"Content-Type": "application/json"})
		if statusResp.Code != http.StatusOK {
			t.Fatalf("update inbox status %q status=%d body=%s", status, statusResp.Code, statusResp.Body.String())
		}
		if !strings.Contains(statusResp.Body.String(), `"status":"`+status+`"`) {
			t.Fatalf("expected status %q in response, body=%s", status, statusResp.Body.String())
		}
	}

	badStatus := doRequest(t, a, http.MethodPatch, "/api/chat/inbox/"+msgPayload.AssistantHandoff.InboxItem.ID, strings.NewReader(`{"profile_id":"`+p.ID+`","status":"deleted"}`), map[string]string{"Content-Type": "application/json"})
	if badStatus.Code != http.StatusBadRequest {
		t.Fatalf("bad status status=%d body=%s", badStatus.Code, badStatus.Body.String())
	}

	inboxList := doRequest(t, a, http.MethodGet, "/api/chat/inbox?profile_id="+p.ID, nil, nil)
	if inboxList.Code != http.StatusOK {
		t.Fatalf("list inbox status=%d body=%s", inboxList.Code, inboxList.Body.String())
	}
	if !strings.Contains(inboxList.Body.String(), `"status":"archived"`) {
		t.Fatalf("expected persisted archived status, body=%s", inboxList.Body.String())
	}
}

func TestChatAPIsValidateErrorsAndProfileIsolation(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createP1 := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"P1"}`), map[string]string{"Content-Type": "application/json"})
	if createP1.Code != http.StatusCreated {
		t.Fatalf("create p1 status=%d body=%s", createP1.Code, createP1.Body.String())
	}
	createP2 := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"P2"}`), map[string]string{"Content-Type": "application/json"})
	if createP2.Code != http.StatusCreated {
		t.Fatalf("create p2 status=%d body=%s", createP2.Code, createP2.Body.String())
	}
	var p1 struct {
		ID string `json:"id"`
	}
	var p2 struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(createP1.Body).Decode(&p1)
	_ = json.NewDecoder(createP2.Body).Decode(&p2)

	createThread := doRequest(t, a, http.MethodPost, "/api/chat/threads", strings.NewReader(`{"profile_id":"`+p1.ID+`","title":"T1"}`), map[string]string{"Content-Type": "application/json"})
	if createThread.Code != http.StatusCreated {
		t.Fatalf("create thread status=%d body=%s", createThread.Code, createThread.Body.String())
	}
	var thread struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(createThread.Body).Decode(&thread)

	badRole := doRequest(t, a, http.MethodPost, "/api/chat/messages", strings.NewReader(`{"profile_id":"`+p1.ID+`","thread_id":"`+thread.ID+`","role":"invalid","content":"hello"}`), map[string]string{"Content-Type": "application/json"})
	if badRole.Code != http.StatusBadRequest {
		t.Fatalf("bad role status=%d body=%s", badRole.Code, badRole.Body.String())
	}

	preview := doRequest(t, a, http.MethodPost, "/api/chat/actions/preview", strings.NewReader(`{"profile_id":"`+p1.ID+`","thread_id":"`+thread.ID+`","action":"create_item_stub","payload":{"part_number":"ISO-1","title":"Isolated"}}`), map[string]string{"Content-Type": "application/json"})
	if preview.Code != http.StatusOK {
		t.Fatalf("preview status=%d body=%s", preview.Code, preview.Body.String())
	}
	var previewObj struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(preview.Body).Decode(&previewObj)

	// Different profile cannot apply another profile's preview.
	crossProfileApply := doRequest(t, a, http.MethodPost, "/api/chat/actions/apply", strings.NewReader(`{"profile_id":"`+p2.ID+`","thread_id":"`+thread.ID+`","preview_id":"`+previewObj.ID+`","confirm":true}`), map[string]string{"Content-Type": "application/json"})
	if crossProfileApply.Code != http.StatusBadRequest {
		t.Fatalf("cross profile apply status=%d body=%s", crossProfileApply.Code, crossProfileApply.Body.String())
	}
}

func TestChatCapabilitiesDiscoveryExposesGovernedRegistry(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Capability Profile"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}

	resp := doRequest(t, a, http.MethodGet, "/api/chat/capabilities?profile_id="+p.ID+"&route=/inventory", nil, nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("capabilities status=%d body=%s", resp.Code, resp.Body.String())
	}
	var payload struct {
		ProfileID    string `json:"profile_id"`
		Route        string `json:"route"`
		Capabilities []struct {
			ID              string   `json:"id"`
			Group           string   `json:"group"`
			Mode            string   `json:"mode"`
			PermissionState string   `json:"permission_state"`
			Requires        []string `json:"requires"`
			Unavailable     bool     `json:"unavailable"`
		} `json:"capabilities"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode capabilities: %v", err)
	}
	if payload.ProfileID != p.ID || payload.Route != "/inventory" {
		t.Fatalf("expected profile and route context, got %+v", payload)
	}

	seen := map[string]struct {
		mode        string
		permission  string
		unavailable bool
	}{}
	for _, capability := range payload.Capabilities {
		seen[capability.ID] = struct {
			mode        string
			permission  string
			unavailable bool
		}{mode: capability.Mode, permission: capability.PermissionState, unavailable: capability.Unavailable}
		if capability.Group == "" || len(capability.Requires) == 0 {
			t.Fatalf("capability must expose group and context requirements: %+v", capability)
		}
	}
	if got := seen["inventory.item.create"]; got.mode != "confirm-required" || got.permission != "available" || got.unavailable {
		t.Fatalf("inventory create must be available but confirm-required, got %+v", got)
	}
	if got := seen["collections.item.assign"]; got.mode != "preview-only" || got.permission != "preview-only" || got.unavailable {
		t.Fatalf("collections assignment must expose preview-only boundary, got %+v", got)
	}
	if got := seen["integrations.provider.run"]; got.mode != "unavailable" || got.permission != "setup-needed" || !got.unavailable {
		t.Fatalf("provider runs must be setup-needed/unavailable until connected, got %+v", got)
	}

	missingProfile := doRequest(t, a, http.MethodGet, "/api/chat/capabilities", nil, nil)
	if missingProfile.Code != http.StatusBadRequest {
		t.Fatalf("missing profile status=%d body=%s", missingProfile.Code, missingProfile.Body.String())
	}
}
