package app

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"slices"
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
	if strings.Contains(msgResp.Body.String(), `"assistant_handoff"`) {
		t.Fatalf("normal hello must not create assistant handoff payload, body=%s", msgResp.Body.String())
	}
	if !strings.Contains(msgResp.Body.String(), `"assistant_response"`) || !strings.Contains(msgResp.Body.String(), `"mode":"direct"`) {
		t.Fatalf("expected direct assistant response payload for normal hello, body=%s", msgResp.Body.String())
	}

	msgList := doRequest(t, a, http.MethodGet, "/api/chat/messages?profile_id="+p.ID+"&thread_id="+thread.ID, nil, nil)
	if msgList.Code != http.StatusOK {
		t.Fatalf("list messages status=%d body=%s", msgList.Code, msgList.Body.String())
	}
	if !strings.Contains(msgList.Body.String(), `"active_workspace_collection":"All Items"`) {
		t.Fatalf("expected listed messages to retain selection context, body=%s", msgList.Body.String())
	}
	if strings.Contains(msgList.Body.String(), `Assistant handoff queued in Inbox.`) {
		t.Fatalf("normal hello must not surface queued handoff state, body=%s", msgList.Body.String())
	}
	if !strings.Contains(msgList.Body.String(), `I can help with Cabinet inventory, media, integrations, purchases, settings, and guided actions from this chat.`) {
		t.Fatalf("expected assistant thread to surface direct normal-text response, body=%s", msgList.Body.String())
	}

	inboxList := doRequest(t, a, http.MethodGet, "/api/chat/inbox?profile_id="+p.ID, nil, nil)
	if inboxList.Code != http.StatusOK {
		t.Fatalf("list inbox status=%d body=%s", inboxList.Code, inboxList.Body.String())
	}
	if strings.Contains(inboxList.Body.String(), `"source":"assistant_handoff"`) || strings.Contains(inboxList.Body.String(), `"thread_id":"`+thread.ID+`"`) {
		t.Fatalf("normal hello must not create default assistant Inbox linkage, body=%s", inboxList.Body.String())
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
	var attachment struct {
		ID        string `json:"id"`
		ProfileID string `json:"profile_id"`
		ThreadID  string `json:"thread_id"`
		Filename  string `json:"filename"`
		MimeType  string `json:"mime_type"`
		SizeBytes int64  `json:"size_bytes"`
		Path      string `json:"path"`
	}
	if err := json.NewDecoder(attachmentResp.Body).Decode(&attachment); err != nil {
		t.Fatalf("decode attachment: %v", err)
	}
	if attachment.ID == "" || attachment.ProfileID != p.ID || attachment.ThreadID != thread.ID || attachment.Filename != "notes.txt" || attachment.SizeBytes != int64(len("sample attachment")) {
		t.Fatalf("expected scoped attachment metadata, got %+v", attachment)
	}

	msgWithAttachment := doRequest(t, a, http.MethodPost, "/api/chat/messages", strings.NewReader(`{"profile_id":"`+p.ID+`","thread_id":"`+thread.ID+`","role":"user","content":"use this explicit attachment","attachment_ids":["`+attachment.ID+`"],"context":{"route":{"pathname":"/chats/"},"profile":{"id":"`+p.ID+`"},"assistant":{"provider":"openai","model":"gpt-4o-mini"}}}`), map[string]string{"Content-Type": "application/json"})
	if msgWithAttachment.Code != http.StatusCreated {
		t.Fatalf("message with attachment status=%d body=%s", msgWithAttachment.Code, msgWithAttachment.Body.String())
	}
	if !strings.Contains(msgWithAttachment.Body.String(), `"attachments_json"`) ||
		!strings.Contains(msgWithAttachment.Body.String(), attachment.ID) ||
		!strings.Contains(msgWithAttachment.Body.String(), `explicit_user_upload`) ||
		!strings.Contains(msgWithAttachment.Body.String(), `in_app_chat`) {
		t.Fatalf("expected message to retain scoped attachment context, body=%s", msgWithAttachment.Body.String())
	}

	otherThreadResp := doRequest(t, a, http.MethodPost, "/api/chat/threads", strings.NewReader(`{"profile_id":"`+p.ID+`","title":"Other Thread"}`), map[string]string{"Content-Type": "application/json"})
	if otherThreadResp.Code != http.StatusCreated {
		t.Fatalf("create other thread status=%d body=%s", otherThreadResp.Code, otherThreadResp.Body.String())
	}
	var otherThread struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(otherThreadResp.Body).Decode(&otherThread); err != nil {
		t.Fatalf("decode other thread: %v", err)
	}
	crossThreadAttachment := doRequest(t, a, http.MethodPost, "/api/chat/messages", strings.NewReader(`{"profile_id":"`+p.ID+`","thread_id":"`+otherThread.ID+`","role":"user","content":"wrong thread attachment","attachment_ids":["`+attachment.ID+`"],"context":{"route":{"pathname":"/chats/"},"profile":{"id":"`+p.ID+`"},"assistant":{"provider":"openai","model":"gpt-4o-mini"}}}`), map[string]string{"Content-Type": "application/json"})
	if crossThreadAttachment.Code != http.StatusBadRequest {
		t.Fatalf("cross-thread attachment status=%d body=%s", crossThreadAttachment.Code, crossThreadAttachment.Body.String())
	}

	var badRecordCount int
	if err := a.db.QueryRow(`SELECT COUNT(1) FROM chat_attachments WHERE filename = 'C:\\secret.txt' OR stored_path LIKE '%secret.txt%'`).Scan(&badRecordCount); err != nil {
		t.Fatalf("count bad attachment records: %v", err)
	}
	if badRecordCount != 0 {
		t.Fatalf("unsupported local-path attachment request created %d records", badRecordCount)
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

	cancelPreviewResp := doRequest(t, a, http.MethodPost, "/api/chat/actions/preview", strings.NewReader(`{"profile_id":"`+p.ID+`","thread_id":"`+thread.ID+`","action":"update_inventory_item","payload":{"item_id":"cancel-target","part_number":"CHAT-CANCEL-001","title":"Canceled Chat Update"}}`), map[string]string{"Content-Type": "application/json"})
	if cancelPreviewResp.Code != http.StatusOK {
		t.Fatalf("cancel preview status=%d body=%s", cancelPreviewResp.Code, cancelPreviewResp.Body.String())
	}
	var cancelPreview struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(cancelPreviewResp.Body).Decode(&cancelPreview); err != nil {
		t.Fatalf("decode cancel preview: %v", err)
	}
	cancelResp := doRequest(t, a, http.MethodPost, "/api/chat/actions/cancel", strings.NewReader(`{"profile_id":"`+p.ID+`","thread_id":"`+thread.ID+`","preview_id":"`+cancelPreview.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if cancelResp.Code != http.StatusOK {
		t.Fatalf("cancel status=%d body=%s", cancelResp.Code, cancelResp.Body.String())
	}
	if !strings.Contains(cancelResp.Body.String(), `"applied":false`) || !strings.Contains(cancelResp.Body.String(), `"preview_id":"`+cancelPreview.ID+`"`) {
		t.Fatalf("expected cancel result with applied=false and preview id, body=%s", cancelResp.Body.String())
	}
	itemsAfterCancel := doRequest(t, a, http.MethodGet, "/api/items?profile_id="+p.ID, nil, nil)
	if itemsAfterCancel.Code != http.StatusOK {
		t.Fatalf("items after cancel status=%d body=%s", itemsAfterCancel.Code, itemsAfterCancel.Body.String())
	}
	if strings.Contains(strings.ToLower(itemsAfterCancel.Body.String()), "chat-cancel-001") {
		t.Fatalf("canceled preview should not mutate inventory, body=%s", itemsAfterCancel.Body.String())
	}
	messagesAfterCancel := doRequest(t, a, http.MethodGet, "/api/chat/messages?profile_id="+p.ID+"&thread_id="+thread.ID, nil, nil)
	if messagesAfterCancel.Code != http.StatusOK {
		t.Fatalf("messages after cancel status=%d body=%s", messagesAfterCancel.Code, messagesAfterCancel.Body.String())
	}
	if !strings.Contains(messagesAfterCancel.Body.String(), "Canceled update_inventory_item") || !strings.Contains(messagesAfterCancel.Body.String(), "no mutation applied") {
		t.Fatalf("expected canceled assistant history outcome, body=%s", messagesAfterCancel.Body.String())
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

func TestNotificationHistoryAPIPromotesLocalHistoryToDurableInbox(t *testing.T) {
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

	body := `{"profile_id":"` + p.ID + `","records":[{"local_history_id":"local-toast-1","level":"warning","title":"Settings save warning","summary":"Banner warning preserved for review.","source_label":"Settings Banner","category":"system","created_at":"2026-06-22T10:00:00Z"}]}`
	record := doRequest(t, a, http.MethodPost, "/api/chat/inbox", strings.NewReader(body), map[string]string{"Content-Type": "application/json"})
	if record.Code != http.StatusCreated {
		t.Fatalf("record history status=%d body=%s", record.Code, record.Body.String())
	}
	if !strings.Contains(record.Body.String(), `"source":"notification_history"`) || !strings.Contains(record.Body.String(), `"local_history_id":"local-toast-1"`) {
		t.Fatalf("expected durable notification history item, body=%s", record.Body.String())
	}

	duplicate := doRequest(t, a, http.MethodPost, "/api/chat/inbox", strings.NewReader(body), map[string]string{"Content-Type": "application/json"})
	if duplicate.Code != http.StatusCreated {
		t.Fatalf("duplicate history status=%d body=%s", duplicate.Code, duplicate.Body.String())
	}
	list := doRequest(t, a, http.MethodGet, "/api/chat/inbox?profile_id="+p.ID, nil, nil)
	if list.Code != http.StatusOK {
		t.Fatalf("list inbox status=%d body=%s", list.Code, list.Body.String())
	}
	if count := strings.Count(list.Body.String(), "local-toast-1"); count != 1 {
		t.Fatalf("expected one deduped durable history item, count=%d body=%s", count, list.Body.String())
	}
	if !strings.Contains(list.Body.String(), `"source_label":"Settings Banner"`) || !strings.Contains(list.Body.String(), `"captured_at":"2026-06-22T10:00:00Z"`) {
		t.Fatalf("expected source and captured-at metadata in durable history item, body=%s", list.Body.String())
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
			ID               string   `json:"id"`
			Group            string   `json:"group"`
			Mode             string   `json:"mode"`
			PermissionState  string   `json:"permission_state"`
			Requires         []string `json:"requires"`
			ProviderRequires []string `json:"provider_requires"`
			InputSchema      string   `json:"input_schema"`
			PreviewShape     string   `json:"preview_shape"`
			ApplyBehavior    string   `json:"apply_behavior"`
			AuditBehavior    string   `json:"audit_behavior"`
			ResultLink       string   `json:"result_link"`
			Unavailable      bool     `json:"unavailable"`
		} `json:"capabilities"`
		GuidedWorkflows []struct {
			ID              string   `json:"id"`
			RequiredContext []string `json:"required_context"`
			UITargets       []string `json:"ui_targets"`
		} `json:"guided_workflows"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode capabilities: %v", err)
	}
	if payload.ProfileID != p.ID || payload.Route != "/inventory" {
		t.Fatalf("expected profile and route context, got %+v", payload)
	}

	seen := map[string]struct {
		mode             string
		permission       string
		unavailable      bool
		providerRequires []string
		inputSchema      string
		previewShape     string
		applyBehavior    string
		auditBehavior    string
		resultLink       string
	}{}
	for _, capability := range payload.Capabilities {
		seen[capability.ID] = struct {
			mode             string
			permission       string
			unavailable      bool
			providerRequires []string
			inputSchema      string
			previewShape     string
			applyBehavior    string
			auditBehavior    string
			resultLink       string
		}{mode: capability.Mode, permission: capability.PermissionState, unavailable: capability.Unavailable, providerRequires: capability.ProviderRequires, inputSchema: capability.InputSchema, previewShape: capability.PreviewShape, applyBehavior: capability.ApplyBehavior, auditBehavior: capability.AuditBehavior, resultLink: capability.ResultLink}
		if capability.Group == "" || len(capability.Requires) == 0 {
			t.Fatalf("capability must expose group and context requirements: %+v", capability)
		}
		if capability.InputSchema == "" || capability.PreviewShape == "" || capability.ApplyBehavior == "" || capability.AuditBehavior == "" || capability.ResultLink == "" {
			t.Fatalf("capability must expose schema, preview, apply, audit and result contract: %+v", capability)
		}
	}
	if got := seen["inventory.item.create"]; got.mode != "confirm-required" || got.permission != "available" || got.unavailable {
		t.Fatalf("inventory create must be available but confirm-required, got %+v", got)
	}
	if got := seen["collections.item.assign"]; got.mode != "confirm-required" || got.permission != "available" || got.unavailable {
		t.Fatalf("collections assignment must expose confirm-required assignment boundary, got %+v", got)
	}
	if got := seen["integrations.provider.run"]; got.mode != "unavailable" || got.permission != "setup-needed" || !got.unavailable {
		t.Fatalf("provider runs must be setup-needed/unavailable until connected, got %+v", got)
	}
	analyze := seen["image_analyze"]
	if analyze.mode != "unavailable" || analyze.permission != "setup-needed" || !analyze.unavailable || analyze.previewShape != "image_analysis_preview_with_sources" || analyze.applyBehavior != "preview_only_no_mutation" || !slices.Contains(analyze.providerRequires, "openai") || !slices.Contains(analyze.providerRequires, "provider_test_passed") || !slices.Contains(analyze.providerRequires, "media_read_access") || analyze.resultLink != "/media" {
		t.Fatalf("image_analyze must expose setup-needed preview-only media analysis contract, got %+v", analyze)
	}
	process := seen["image_process"]
	if process.mode != "unavailable" || process.permission != "setup-needed" || !process.unavailable || process.previewShape != "image_process_variant_preview" || process.applyBehavior != "requires_explicit_confirmation" || !slices.Contains(process.providerRequires, "openai") || !slices.Contains(process.providerRequires, "provider_test_passed") || !slices.Contains(process.providerRequires, "media_write_access") || process.resultLink != "/media" {
		t.Fatalf("image_process must expose setup-needed confirmation-gated media variant contract, got %+v", process)
	}
	content := seen["content_generate"]
	if content.mode != "unavailable" || content.permission != "setup-needed" || !content.unavailable || content.previewShape != "catalog_content_draft_preview" || content.applyBehavior != "preview_only_no_mutation" || !slices.Contains(content.providerRequires, "openai") {
		t.Fatalf("content_generate must expose preview-only setup-needed OpenAI contract, got %+v", content)
	}
	listing := seen["listing_draft_generate"]
	if listing.mode != "unavailable" || listing.permission != "setup-needed" || !listing.unavailable || listing.previewShape != "listing_draft_preview_with_sources" || listing.applyBehavior != "requires_explicit_confirmation" || !slices.Contains(listing.providerRequires, "provider_test_passed") {
		t.Fatalf("listing_draft_generate must expose confirmation-gated setup-needed OpenAI contract, got %+v", listing)
	}
	var inventoryWorkflow, wishlistWorkflow bool
	for _, workflow := range payload.GuidedWorkflows {
		switch workflow.ID {
		case "inventory.item.update":
			inventoryWorkflow = slices.Contains(workflow.RequiredContext, "target_inventory_item") &&
				slices.Contains(workflow.UITargets, "inventory.item.editor.title")
		case "wishlist.entry.create":
			wishlistWorkflow = true
		}
	}
	if !inventoryWorkflow || !wishlistWorkflow {
		t.Fatalf("expected capabilities payload to list inventory and non-inventory guided workflows, got %+v", payload.GuidedWorkflows)
	}

	missingProfile := doRequest(t, a, http.MethodGet, "/api/chat/capabilities", nil, nil)
	if missingProfile.Code != http.StatusBadRequest {
		t.Fatalf("missing profile status=%d body=%s", missingProfile.Code, missingProfile.Body.String())
	}
}

func TestChatActionPreviewEndpointUsesCapabilityRegistry(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Capability Preview API"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}

	threadResp := doRequest(t, a, http.MethodPost, "/api/chat/threads", strings.NewReader(`{"profile_id":"`+p.ID+`","title":"Capability Preview Thread"}`), map[string]string{"Content-Type": "application/json"})
	if threadResp.Code != http.StatusCreated {
		t.Fatalf("create thread status=%d body=%s", threadResp.Code, threadResp.Body.String())
	}
	var thread struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(threadResp.Body).Decode(&thread); err != nil {
		t.Fatalf("decode thread: %v", err)
	}

	preview := doRequest(t, a, http.MethodPost, "/api/chat/actions/preview", strings.NewReader(`{"profile_id":"`+p.ID+`","thread_id":"`+thread.ID+`","capability_id":"wishlist.entry.create","payload":{"part_number":"CAP-WISH-001","title":"Capability Wishlist","priority":"high"}}`), map[string]string{"Content-Type": "application/json"})
	if preview.Code != http.StatusOK {
		t.Fatalf("capability preview status=%d body=%s", preview.Code, preview.Body.String())
	}
	if !strings.Contains(preview.Body.String(), `"capability_id":"wishlist.entry.create"`) || !strings.Contains(preview.Body.String(), `"action":"create_wishlist_entry"`) {
		t.Fatalf("expected capability-backed wishlist action preview, body=%s", preview.Body.String())
	}

	unavailable := doRequest(t, a, http.MethodPost, "/api/chat/actions/preview", strings.NewReader(`{"profile_id":"`+p.ID+`","thread_id":"`+thread.ID+`","capability_id":"integrations.provider.run","payload":{"provider":"openai"}}`), map[string]string{"Content-Type": "application/json"})
	if unavailable.Code != http.StatusBadRequest {
		t.Fatalf("unavailable capability status=%d body=%s", unavailable.Code, unavailable.Body.String())
	}
	if !strings.Contains(unavailable.Body.String(), "integrations.provider.run") || !strings.Contains(unavailable.Body.String(), "setup needed") {
		t.Fatalf("expected deterministic unavailable capability guidance, body=%s", unavailable.Body.String())
	}
}

func TestCabinetAgentAppControlCapabilitiesAndOpenItemTitleMutation(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Agent Control Profile"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	if activate := doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+p.ID+`"}`), map[string]string{"Content-Type": "application/json"}); activate.Code != http.StatusOK {
		t.Fatalf("activate profile status=%d body=%s", activate.Code, activate.Body.String())
	}

	itemResp := doRequest(t, a, http.MethodPost, "/api/items", strings.NewReader(`{"part_number":"AGENT-001","title":"Original Agent Title","brand":"AFX","category":"Slot Cars"}`), map[string]string{"Content-Type": "application/json"})
	if itemResp.Code != http.StatusCreated {
		t.Fatalf("create item status=%d body=%s", itemResp.Code, itemResp.Body.String())
	}
	var item struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(itemResp.Body).Decode(&item); err != nil {
		t.Fatalf("decode item: %v", err)
	}

	threadResp := doRequest(t, a, http.MethodPost, "/api/chat/threads", strings.NewReader(`{"profile_id":"`+p.ID+`","title":"Agent Control Thread"}`), map[string]string{"Content-Type": "application/json"})
	if threadResp.Code != http.StatusCreated {
		t.Fatalf("create thread status=%d body=%s", threadResp.Code, threadResp.Body.String())
	}
	var thread struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(threadResp.Body).Decode(&thread); err != nil {
		t.Fatalf("decode thread: %v", err)
	}

	capabilities := doRequest(t, a, http.MethodGet, "/api/chat/capabilities?profile_id="+p.ID+"&route=/inventory/"+item.ID, nil, nil)
	if capabilities.Code != http.StatusOK {
		t.Fatalf("capabilities status=%d body=%s", capabilities.Code, capabilities.Body.String())
	}
	body := capabilities.Body.String()
	if !strings.Contains(body, `"id":"navigate.open_surface"`) || !strings.Contains(body, `"input_schema":"agent.navigate.open_surface.v1"`) || !strings.Contains(body, `"result_link":"/media"`) {
		t.Fatalf("expected navigate.open_surface app-control capability with Media target, body=%s", body)
	}
	for _, expectedTarget := range []string{`"id":"dashboard"`, `"route":"/inventory"`, `"route":"/wishlist"`, `"route":"/collections"`, `"route":"/media"`, `"route":"/discoveries"`, `"route":"/scanner"`, `"route":"/purchases"`, `"route":"/integrations"`, `"route":"/chats"`, `"route":"/inbox"`, `"route":"/settings/profile"`, `"route":"/settings/account"`, `"route":"/settings/appearance"`, `"route":"/settings/storage"`} {
		if !strings.Contains(body, expectedTarget) {
			t.Fatalf("expected navigate.open_surface allowlist target %s, body=%s", expectedTarget, body)
		}
	}
	if !strings.Contains(body, `"id":"update_open_item_title"`) || !strings.Contains(body, `"mode":"confirm-required"`) || !strings.Contains(body, `"input_schema":"agent.update_open_item_title.v1"`) {
		t.Fatalf("expected confirm-required open item title capability, body=%s", body)
	}

	preview := doRequest(t, a, http.MethodPost, "/api/chat/actions/preview", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"thread_id":"`+thread.ID+`",
		"action":"update_open_item_title",
		"payload":{"item_id":"`+item.ID+`","title":"Agent Updated Title","source_route":"/inventory/`+item.ID+`"}
	}`), map[string]string{"Content-Type": "application/json"})
	if preview.Code != http.StatusOK {
		t.Fatalf("preview status=%d body=%s", preview.Code, preview.Body.String())
	}
	var previewObj struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(preview.Body).Decode(&previewObj); err != nil {
		t.Fatalf("decode preview: %v", err)
	}

	withoutConfirm := doRequest(t, a, http.MethodPost, "/api/chat/actions/apply", strings.NewReader(`{"profile_id":"`+p.ID+`","thread_id":"`+thread.ID+`","preview_id":"`+previewObj.ID+`","confirm":false}`), map[string]string{"Content-Type": "application/json"})
	if withoutConfirm.Code != http.StatusBadRequest {
		t.Fatalf("apply without confirmation status=%d body=%s", withoutConfirm.Code, withoutConfirm.Body.String())
	}

	apply := doRequest(t, a, http.MethodPost, "/api/chat/actions/apply", strings.NewReader(`{"profile_id":"`+p.ID+`","thread_id":"`+thread.ID+`","preview_id":"`+previewObj.ID+`","confirm":true}`), map[string]string{"Content-Type": "application/json"})
	if apply.Code != http.StatusOK {
		t.Fatalf("apply status=%d body=%s", apply.Code, apply.Body.String())
	}
	if !strings.Contains(apply.Body.String(), `"action":"update_open_item_title"`) || !strings.Contains(apply.Body.String(), `"item_id":"`+item.ID+`"`) || !strings.Contains(apply.Body.String(), `"title":"Agent Updated Title"`) {
		t.Fatalf("expected app-control update result evidence, body=%s", apply.Body.String())
	}

	items := doRequest(t, a, http.MethodGet, "/api/items?profile_id="+p.ID, nil, nil)
	if items.Code != http.StatusOK {
		t.Fatalf("list items status=%d body=%s", items.Code, items.Body.String())
	}
	if !strings.Contains(items.Body.String(), `"title":"Agent Updated Title"`) || strings.Contains(items.Body.String(), `"title":"Original Agent Title"`) {
		t.Fatalf("expected persisted title update for active profile item, body=%s", items.Body.String())
	}

	messages := doRequest(t, a, http.MethodGet, "/api/chat/messages?profile_id="+p.ID+"&thread_id="+thread.ID, nil, nil)
	if messages.Code != http.StatusOK {
		t.Fatalf("list messages status=%d body=%s", messages.Code, messages.Body.String())
	}
	if !strings.Contains(messages.Body.String(), "Applied update_open_item_title") || !strings.Contains(messages.Body.String(), `"mutation_applied":true`) || !strings.Contains(messages.Body.String(), `"confirmation":"confirmed"`) {
		t.Fatalf("expected confirmed app-control audit message, body=%s", messages.Body.String())
	}
}

func TestChatMessageAppControlPlannerDispatchesDeterministicActions(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Chat Planner Profile"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	if activate := doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+p.ID+`"}`), map[string]string{"Content-Type": "application/json"}); activate.Code != http.StatusOK {
		t.Fatalf("activate profile status=%d body=%s", activate.Code, activate.Body.String())
	}

	threadResp := doRequest(t, a, http.MethodPost, "/api/chat/threads", strings.NewReader(`{"profile_id":"`+p.ID+`","title":"Planner Thread"}`), map[string]string{"Content-Type": "application/json"})
	if threadResp.Code != http.StatusCreated {
		t.Fatalf("create thread status=%d body=%s", threadResp.Code, threadResp.Body.String())
	}
	var thread struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(threadResp.Body).Decode(&thread); err != nil {
		t.Fatalf("decode thread: %v", err)
	}

	routeResp := doRequest(t, a, http.MethodPost, "/api/chat/messages", strings.NewReader(`{"profile_id":"`+p.ID+`","thread_id":"`+thread.ID+`","role":"user","content":"open media","context":{"route":{"pathname":"/chats"},"assistant":{"provider":"openai","model":"gpt-4o-mini"}}}`), map[string]string{"Content-Type": "application/json"})
	if routeResp.Code != http.StatusCreated {
		t.Fatalf("route app-control status=%d body=%s", routeResp.Code, routeResp.Body.String())
	}
	routeBody := routeResp.Body.String()
	if strings.Contains(routeBody, `"assistant_handoff"`) {
		t.Fatalf("handled app-control route must not create default inbox handoff, body=%s", routeBody)
	}
	if !strings.Contains(routeBody, `"capability_id":"navigate.open_surface"`) || !strings.Contains(routeBody, `"route":"/media"`) || !strings.Contains(routeBody, `"confirmation_state":"not_required"`) {
		t.Fatalf("expected navigate.open_surface app-control route result, body=%s", routeBody)
	}

	integrationsResp := doRequest(t, a, http.MethodPost, "/api/chat/messages", strings.NewReader(`{"profile_id":"`+p.ID+`","thread_id":"`+thread.ID+`","role":"user","content":"go to integrations","context":{"route":{"pathname":"/chats"},"assistant":{"provider":"openai","model":"gpt-4o-mini"}}}`), map[string]string{"Content-Type": "application/json"})
	if integrationsResp.Code != http.StatusCreated {
		t.Fatalf("integrations app-control status=%d body=%s", integrationsResp.Code, integrationsResp.Body.String())
	}
	integrationsBody := integrationsResp.Body.String()
	if strings.Contains(integrationsBody, `"assistant_handoff"`) {
		t.Fatalf("handled integrations route must not create default inbox handoff, body=%s", integrationsBody)
	}
	if !strings.Contains(integrationsBody, `"capability_id":"navigate.open_surface"`) || !strings.Contains(integrationsBody, `"route":"/integrations"`) || !strings.Contains(integrationsBody, `"confirmation_state":"not_required"`) {
		t.Fatalf("expected integrations open-surface route result, body=%s", integrationsBody)
	}

	unsafeResp := doRequest(t, a, http.MethodPost, "/api/chat/messages", strings.NewReader(`{"profile_id":"`+p.ID+`","thread_id":"`+thread.ID+`","role":"user","content":"open admin console","context":{"route":{"pathname":"/chats"},"assistant":{"provider":"openai","model":"gpt-4o-mini"}}}`), map[string]string{"Content-Type": "application/json"})
	if unsafeResp.Code != http.StatusCreated {
		t.Fatalf("unsafe app-control status=%d body=%s", unsafeResp.Code, unsafeResp.Body.String())
	}
	unsafeBody := unsafeResp.Body.String()
	if strings.Contains(unsafeBody, `"assistant_handoff"`) {
		t.Fatalf("rejected unsafe route must not create default inbox handoff, body=%s", unsafeBody)
	}
	if !strings.Contains(unsafeBody, `"capability_id":"navigate.open_surface"`) || !strings.Contains(unsafeBody, `"code":"unknown_surface"`) || strings.Contains(unsafeBody, `"route":"/admin`) {
		t.Fatalf("expected unknown surface rejection without unsafe route, body=%s", unsafeBody)
	}

	createResp := doRequest(t, a, http.MethodPost, "/api/chat/messages", strings.NewReader(`{"profile_id":"`+p.ID+`","thread_id":"`+thread.ID+`","role":"user","content":"create an inventory item AFX-001 Test Car","context":{"route":{"pathname":"/chats"},"assistant":{"provider":"openai","model":"gpt-4o-mini"}}}`), map[string]string{"Content-Type": "application/json"})
	if createResp.Code != http.StatusCreated {
		t.Fatalf("create preview app-control status=%d body=%s", createResp.Code, createResp.Body.String())
	}
	createBody := createResp.Body.String()
	if strings.Contains(createBody, `"assistant_handoff"`) {
		t.Fatalf("handled app-control create must not create default inbox handoff, body=%s", createBody)
	}
	if !strings.Contains(createBody, `"capability_id":"inventory.item.create"`) || !strings.Contains(createBody, `"action":"create_inventory_item"`) || !strings.Contains(createBody, `"confirmation_state":"pending"`) || !strings.Contains(createBody, `"part_number":"AFX-001"`) {
		t.Fatalf("expected create_inventory_item preview result, body=%s", createBody)
	}
	itemsBeforeConfirm := doRequest(t, a, http.MethodGet, "/api/items?profile_id="+p.ID, nil, nil)
	if itemsBeforeConfirm.Code != http.StatusOK {
		t.Fatalf("items status=%d body=%s", itemsBeforeConfirm.Code, itemsBeforeConfirm.Body.String())
	}
	if strings.Contains(itemsBeforeConfirm.Body.String(), "AFX-001") {
		t.Fatalf("planner preview must not mutate inventory before confirmation, body=%s", itemsBeforeConfirm.Body.String())
	}

	itemResp := doRequest(t, a, http.MethodPost, "/api/items", strings.NewReader(`{"part_number":"REN-001","title":"Original Planner Title","brand":"AFX","category":"Slot Cars"}`), map[string]string{"Content-Type": "application/json"})
	if itemResp.Code != http.StatusCreated {
		t.Fatalf("create item status=%d body=%s", itemResp.Code, itemResp.Body.String())
	}
	var item struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(itemResp.Body).Decode(&item); err != nil {
		t.Fatalf("decode item: %v", err)
	}
	renameResp := doRequest(t, a, http.MethodPost, "/api/chat/messages", strings.NewReader(`{"profile_id":"`+p.ID+`","thread_id":"`+thread.ID+`","role":"user","content":"rename this item to Planner Updated Title","context":{"route":{"pathname":"/inventory/`+item.ID+`"},"assistant":{"provider":"openai","model":"gpt-4o-mini"}}}`), map[string]string{"Content-Type": "application/json"})
	if renameResp.Code != http.StatusCreated {
		t.Fatalf("rename preview app-control status=%d body=%s", renameResp.Code, renameResp.Body.String())
	}
	renameBody := renameResp.Body.String()
	if !strings.Contains(renameBody, `"capability_id":"update_open_item_title"`) || !strings.Contains(renameBody, `"action":"update_open_item_title"`) || !strings.Contains(renameBody, `"item_id":"`+item.ID+`"`) || !strings.Contains(renameBody, `"title":"Planner Updated Title"`) {
		t.Fatalf("expected open item rename preview result, body=%s", renameBody)
	}
	itemsAfterRenamePreview := doRequest(t, a, http.MethodGet, "/api/items?profile_id="+p.ID, nil, nil)
	if itemsAfterRenamePreview.Code != http.StatusOK {
		t.Fatalf("items after rename preview status=%d body=%s", itemsAfterRenamePreview.Code, itemsAfterRenamePreview.Body.String())
	}
	if strings.Contains(itemsAfterRenamePreview.Body.String(), "Planner Updated Title") {
		t.Fatalf("rename preview must not mutate inventory before confirmation, body=%s", itemsAfterRenamePreview.Body.String())
	}

	runs := doRequest(t, a, http.MethodGet, "/api/chat/workflow-runs?profile_id="+p.ID+"&thread_id="+thread.ID, nil, nil)
	if runs.Code != http.StatusOK {
		t.Fatalf("workflow runs status=%d body=%s", runs.Code, runs.Body.String())
	}
	if !strings.Contains(runs.Body.String(), `"workflow_id":"chat.app_control.dispatch"`) || !strings.Contains(runs.Body.String(), `"capability_id":"navigate.open_surface"`) || !strings.Contains(runs.Body.String(), `"capability_id":"inventory.item.create"`) || !strings.Contains(runs.Body.String(), `"capability_id":"update_open_item_title"`) || !strings.Contains(runs.Body.String(), `"status":"failed"`) || !strings.Contains(runs.Body.String(), `"code":"unknown_surface"`) {
		t.Fatalf("expected durable app-control workflow audit runs, body=%s", runs.Body.String())
	}
}

func TestAssistantContentListingGenerationRunsStayPreviewFirst(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Generated Content Profile"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	activate := doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+p.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if activate.Code != http.StatusOK {
		t.Fatalf("activate profile status=%d body=%s", activate.Code, activate.Body.String())
	}

	itemResp := doRequest(t, a, http.MethodPost, "/api/items", strings.NewReader(`{"part_number":"GEN-001","title":"Original catalog title","brand":"AFX","category":"Cars","item_type":"Slot Cars","condition":"10+ - New, in packaging"}`), map[string]string{"Content-Type": "application/json"})
	if itemResp.Code != http.StatusCreated {
		t.Fatalf("create item status=%d body=%s", itemResp.Code, itemResp.Body.String())
	}
	var item struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(itemResp.Body).Decode(&item); err != nil {
		t.Fatalf("decode item: %v", err)
	}

	threadResp := doRequest(t, a, http.MethodPost, "/api/chat/threads", strings.NewReader(`{"profile_id":"`+p.ID+`","title":"Generated Content Thread"}`), map[string]string{"Content-Type": "application/json"})
	if threadResp.Code != http.StatusCreated {
		t.Fatalf("create thread status=%d body=%s", threadResp.Code, threadResp.Body.String())
	}
	var thread struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(threadResp.Body).Decode(&thread); err != nil {
		t.Fatalf("decode thread: %v", err)
	}

	contentRun := doRequest(t, a, http.MethodPost, "/api/chat/workflow-runs", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"workflow_id":"openai-content-generate",
		"capability_id":"content_generate",
		"source_channel":"in_app_chat",
		"source_thread_id":"`+thread.ID+`",
		"confirmation_state":"not_required",
		"input":{"item_id":"`+item.ID+`","fields":["description","condition_notes"]},
		"provider_trace":{"provider":"openai","setup_needed":"provider_test_required"}
	}`), map[string]string{"Content-Type": "application/json"})
	if contentRun.Code != http.StatusCreated {
		t.Fatalf("create content workflow run status=%d body=%s", contentRun.Code, contentRun.Body.String())
	}
	var content struct {
		ID                string         `json:"id"`
		CapabilityID      string         `json:"capability_id"`
		ConfirmationState string         `json:"confirmation_state"`
		ProviderTrace     map[string]any `json:"provider_trace"`
	}
	if err := json.NewDecoder(contentRun.Body).Decode(&content); err != nil {
		t.Fatalf("decode content run: %v", err)
	}
	if content.CapabilityID != "content_generate" || content.ConfirmationState != "not_required" || content.ProviderTrace["setup_needed"] != "provider_test_required" {
		t.Fatalf("content_generate must remain setup-needed preview-only metadata, got %+v", content)
	}

	contentDone := doRequest(t, a, http.MethodPatch, "/api/chat/workflow-runs/"+content.ID, strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"status":"completed",
		"confirmation_state":"not_required",
		"result":{"preview_id":"content-preview-1","item_id":"`+item.ID+`","description":"Generated sales copy","condition_notes":"Generated condition notes"}
	}`), map[string]string{"Content-Type": "application/json"})
	if contentDone.Code != http.StatusOK {
		t.Fatalf("complete content run status=%d body=%s", contentDone.Code, contentDone.Body.String())
	}
	if !strings.Contains(contentDone.Body.String(), `"preview_id":"content-preview-1"`) || !strings.Contains(contentDone.Body.String(), `"confirmation_state":"not_required"`) {
		t.Fatalf("content run must return preview result without pending apply state, body=%s", contentDone.Body.String())
	}

	itemsAfterContent := doRequest(t, a, http.MethodGet, "/api/items", nil, nil)
	if itemsAfterContent.Code != http.StatusOK {
		t.Fatalf("list items after content run status=%d body=%s", itemsAfterContent.Code, itemsAfterContent.Body.String())
	}
	if !strings.Contains(itemsAfterContent.Body.String(), `"title":"Original catalog title"`) || strings.Contains(itemsAfterContent.Body.String(), "Generated sales copy") {
		t.Fatalf("content_generate result must not silently mutate inventory item, body=%s", itemsAfterContent.Body.String())
	}

	listingRun := doRequest(t, a, http.MethodPost, "/api/chat/workflow-runs", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"workflow_id":"openai-listing-draft-generate",
		"capability_id":"listing_draft_generate",
		"source_channel":"in_app_chat",
		"source_thread_id":"`+thread.ID+`",
		"confirmation_state":"required",
		"input":{"item_id":"`+item.ID+`","target_marketplace":"ebay"},
		"provider_trace":{"provider":"openai","setup_needed":"provider_test_required"}
	}`), map[string]string{"Content-Type": "application/json"})
	if listingRun.Code != http.StatusCreated {
		t.Fatalf("create listing workflow run status=%d body=%s", listingRun.Code, listingRun.Body.String())
	}
	var listing struct {
		ID                string `json:"id"`
		CapabilityID      string `json:"capability_id"`
		ConfirmationState string `json:"confirmation_state"`
	}
	if err := json.NewDecoder(listingRun.Body).Decode(&listing); err != nil {
		t.Fatalf("decode listing run: %v", err)
	}
	if listing.CapabilityID != "listing_draft_generate" || listing.ConfirmationState != "required" {
		t.Fatalf("listing_draft_generate must start confirmation-gated, got %+v", listing)
	}

	listingDone := doRequest(t, a, http.MethodPatch, "/api/chat/workflow-runs/"+listing.ID, strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"status":"completed",
		"confirmation_state":"pending",
		"result":{"preview_id":"listing-preview-1","marketplace":"ebay","title":"Generated listing title","sources":[{"type":"item","id":"`+item.ID+`"}]}
	}`), map[string]string{"Content-Type": "application/json"})
	if listingDone.Code != http.StatusOK {
		t.Fatalf("complete listing run status=%d body=%s", listingDone.Code, listingDone.Body.String())
	}
	if !strings.Contains(listingDone.Body.String(), `"confirmation_state":"pending"`) || !strings.Contains(listingDone.Body.String(), `"sources"`) {
		t.Fatalf("listing draft must preserve source-attributed pending confirmation preview, body=%s", listingDone.Body.String())
	}
}

func TestAssistantWorkflowRunsPersistLifecycleAndBulkResults(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Workflow Profile"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	threadResp := doRequest(t, a, http.MethodPost, "/api/chat/threads", strings.NewReader(`{"profile_id":"`+p.ID+`","title":"Workflow Thread"}`), map[string]string{"Content-Type": "application/json"})
	if threadResp.Code != http.StatusCreated {
		t.Fatalf("create thread status=%d body=%s", threadResp.Code, threadResp.Body.String())
	}
	var thread struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(threadResp.Body).Decode(&thread); err != nil {
		t.Fatalf("decode thread: %v", err)
	}

	createRun := doRequest(t, a, http.MethodPost, "/api/chat/workflow-runs", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"workflow_id":"catalog-draft-from-text",
		"capability_id":"catalog_add_from_text",
		"source_channel":"telegram",
		"source_thread_id":"`+thread.ID+`",
		"source_message_id":"tg-message-1",
		"confirmation_state":"required",
		"input":{"text":"AFX slot car notes"},
		"provider_trace":{"provider":"openai","method":"api_key","model":"gpt-4o-mini"}
	}`), map[string]string{"Content-Type": "application/json"})
	if createRun.Code != http.StatusCreated {
		t.Fatalf("create workflow run status=%d body=%s", createRun.Code, createRun.Body.String())
	}
	var run struct {
		ID                string         `json:"id"`
		ProfileID         string         `json:"profile_id"`
		CapabilityID      string         `json:"capability_id"`
		SourceChannel     string         `json:"source_channel"`
		SourceThreadID    string         `json:"source_thread_id"`
		SourceMessageID   string         `json:"source_message_id"`
		Status            string         `json:"status"`
		ConfirmationState string         `json:"confirmation_state"`
		ProviderTrace     map[string]any `json:"provider_trace"`
		Input             map[string]any `json:"input"`
	}
	if err := json.NewDecoder(createRun.Body).Decode(&run); err != nil {
		t.Fatalf("decode run: %v", err)
	}
	if run.ID == "" || run.ProfileID != p.ID || run.Status != "queued" || run.ConfirmationState != "required" {
		t.Fatalf("unexpected created workflow run: %+v", run)
	}
	if run.CapabilityID != "catalog_add_from_text" || run.SourceChannel != "telegram" || run.SourceThreadID != thread.ID || run.SourceMessageID != "tg-message-1" {
		t.Fatalf("expected source/capability metadata to persist, got %+v", run)
	}
	if run.ProviderTrace["provider"] != "openai" || run.Input["text"] != "AFX slot car notes" {
		t.Fatalf("expected provider trace and input payload, got %+v", run)
	}

	running := doRequest(t, a, http.MethodPatch, "/api/chat/workflow-runs/"+run.ID, strings.NewReader(`{"profile_id":"`+p.ID+`","status":"running","provider_trace":{"provider":"openai","request_id":"req_123","model":"gpt-4o-mini"}}`), map[string]string{"Content-Type": "application/json"})
	if running.Code != http.StatusOK {
		t.Fatalf("running status=%d body=%s", running.Code, running.Body.String())
	}
	if !strings.Contains(running.Body.String(), `"started_at"`) || !strings.Contains(running.Body.String(), `"request_id":"req_123"`) {
		t.Fatalf("expected running update to persist started_at/provider trace, body=%s", running.Body.String())
	}

	completed := doRequest(t, a, http.MethodPatch, "/api/chat/workflow-runs/"+run.ID, strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"status":"completed",
		"confirmation_state":"pending",
		"result":{"preview_id":"preview-1","summary":"Draft ready"},
		"bulk_items":[
			{"item_key":"photo-1","status":"completed","result_id":"draft-1"},
			{"item_key":"photo-2","status":"failed","error_code":"low_confidence"}
		]
	}`), map[string]string{"Content-Type": "application/json"})
	if completed.Code != http.StatusOK {
		t.Fatalf("completed status=%d body=%s", completed.Code, completed.Body.String())
	}
	if !strings.Contains(completed.Body.String(), `"status":"completed"`) || !strings.Contains(completed.Body.String(), `"confirmation_state":"pending"`) {
		t.Fatalf("expected completed pending-confirmation state, body=%s", completed.Body.String())
	}
	if !strings.Contains(completed.Body.String(), `"item_key":"photo-2"`) || !strings.Contains(completed.Body.String(), `"error_code":"low_confidence"`) {
		t.Fatalf("expected per-item bulk failure result, body=%s", completed.Body.String())
	}

	list := doRequest(t, a, http.MethodGet, "/api/chat/workflow-runs?profile_id="+p.ID+"&thread_id="+thread.ID, nil, nil)
	if list.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", list.Code, list.Body.String())
	}
	if !strings.Contains(list.Body.String(), run.ID) || !strings.Contains(list.Body.String(), `"preview_id":"preview-1"`) {
		t.Fatalf("expected listed run with result payload, body=%s", list.Body.String())
	}

	failed := doRequest(t, a, http.MethodPost, "/api/chat/workflow-runs", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"workflow_id":"provider-test",
		"capability_id":"provider_test",
		"provider_trace":{"provider":"openai","setup_needed":"api_key_missing"}
	}`), map[string]string{"Content-Type": "application/json"})
	if failed.Code != http.StatusCreated {
		t.Fatalf("create failed-run candidate status=%d body=%s", failed.Code, failed.Body.String())
	}
	var failedRun struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(failed.Body).Decode(&failedRun); err != nil {
		t.Fatalf("decode failed run: %v", err)
	}
	failedUpdate := doRequest(t, a, http.MethodPatch, "/api/chat/workflow-runs/"+failedRun.ID, strings.NewReader(`{"profile_id":"`+p.ID+`","status":"failed","error":{"code":"setup_needed","message":"OpenAI API key is not connected","retry":"connect_provider"}}`), map[string]string{"Content-Type": "application/json"})
	if failedUpdate.Code != http.StatusOK {
		t.Fatalf("failed update status=%d body=%s", failedUpdate.Code, failedUpdate.Body.String())
	}
	if !strings.Contains(failedUpdate.Body.String(), `"status":"failed"`) || !strings.Contains(failedUpdate.Body.String(), `"retry":"connect_provider"`) {
		t.Fatalf("expected failure payload and retry guidance, body=%s", failedUpdate.Body.String())
	}

	badStatus := doRequest(t, a, http.MethodPatch, "/api/chat/workflow-runs/"+failedRun.ID, strings.NewReader(`{"profile_id":"`+p.ID+`","status":"mystery"}`), map[string]string{"Content-Type": "application/json"})
	if badStatus.Code != http.StatusBadRequest {
		t.Fatalf("bad status=%d body=%s", badStatus.Code, badStatus.Body.String())
	}
}

func TestAssistantImageCapabilityRunsPreserveOriginalAndAuditLinks(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Image Capability Profile"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	threadResp := doRequest(t, a, http.MethodPost, "/api/chat/threads", strings.NewReader(`{"profile_id":"`+p.ID+`","title":"Image Capability Thread"}`), map[string]string{"Content-Type": "application/json"})
	if threadResp.Code != http.StatusCreated {
		t.Fatalf("create thread status=%d body=%s", threadResp.Code, threadResp.Body.String())
	}
	var thread struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(threadResp.Body).Decode(&thread); err != nil {
		t.Fatalf("decode thread: %v", err)
	}

	analyzeRun := doRequest(t, a, http.MethodPost, "/api/chat/workflow-runs", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"workflow_id":"openai-image-analyze",
		"capability_id":"image_analyze",
		"source_channel":"in_app_chat",
		"source_thread_id":"`+thread.ID+`",
		"confirmation_state":"not_required",
		"input":{"media_id":"media-original-1","analysis_goal":"identify visible item details"},
		"provider_trace":{"provider":"openai","setup_needed":"provider_test_required","media_access":"read"}
	}`), map[string]string{"Content-Type": "application/json"})
	if analyzeRun.Code != http.StatusCreated {
		t.Fatalf("create image analyze run status=%d body=%s", analyzeRun.Code, analyzeRun.Body.String())
	}
	var analyze struct {
		ID                string         `json:"id"`
		CapabilityID      string         `json:"capability_id"`
		ConfirmationState string         `json:"confirmation_state"`
		ProviderTrace     map[string]any `json:"provider_trace"`
	}
	if err := json.NewDecoder(analyzeRun.Body).Decode(&analyze); err != nil {
		t.Fatalf("decode analyze run: %v", err)
	}
	if analyze.CapabilityID != "image_analyze" || analyze.ConfirmationState != "not_required" || analyze.ProviderTrace["media_access"] != "read" {
		t.Fatalf("image_analyze must start as setup-needed preview-only read run, got %+v", analyze)
	}
	analyzeDone := doRequest(t, a, http.MethodPatch, "/api/chat/workflow-runs/"+analyze.ID, strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"status":"completed",
		"confirmation_state":"not_required",
		"result":{"preview_id":"image-analysis-preview-1","media_id":"media-original-1","findings":[{"field":"title","value":"AFX slot car","confidence":0.74}],"source_links":[{"type":"media","id":"media-original-1","href":"/media"}]}
	}`), map[string]string{"Content-Type": "application/json"})
	if analyzeDone.Code != http.StatusOK {
		t.Fatalf("complete image analyze run status=%d body=%s", analyzeDone.Code, analyzeDone.Body.String())
	}
	if !strings.Contains(analyzeDone.Body.String(), `"confirmation_state":"not_required"`) || !strings.Contains(analyzeDone.Body.String(), `"source_links"`) || strings.Contains(analyzeDone.Body.String(), `"variant_media_id"`) {
		t.Fatalf("image_analyze must return source-linked preview findings without processed variant mutation, body=%s", analyzeDone.Body.String())
	}

	processRun := doRequest(t, a, http.MethodPost, "/api/chat/workflow-runs", strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"workflow_id":"openai-image-process",
		"capability_id":"image_process",
		"source_channel":"in_app_chat",
		"source_thread_id":"`+thread.ID+`",
		"confirmation_state":"required",
		"input":{"media_id":"media-original-1","operation":"background_cleanup","preserve_original":true},
		"provider_trace":{"provider":"openai","setup_needed":"provider_test_required","media_access":"read_write"}
	}`), map[string]string{"Content-Type": "application/json"})
	if processRun.Code != http.StatusCreated {
		t.Fatalf("create image process run status=%d body=%s", processRun.Code, processRun.Body.String())
	}
	var process struct {
		ID                string         `json:"id"`
		CapabilityID      string         `json:"capability_id"`
		ConfirmationState string         `json:"confirmation_state"`
		ProviderTrace     map[string]any `json:"provider_trace"`
	}
	if err := json.NewDecoder(processRun.Body).Decode(&process); err != nil {
		t.Fatalf("decode process run: %v", err)
	}
	if process.CapabilityID != "image_process" || process.ConfirmationState != "required" || process.ProviderTrace["setup_needed"] != "provider_test_required" || process.ProviderTrace["media_access"] != "read_write" {
		t.Fatalf("image_process must start confirmation-gated, got %+v", process)
	}
	processDone := doRequest(t, a, http.MethodPatch, "/api/chat/workflow-runs/"+process.ID, strings.NewReader(`{
		"profile_id":"`+p.ID+`",
		"status":"completed",
		"confirmation_state":"pending",
		"result":{"preview_id":"image-process-preview-1","original_media_id":"media-original-1","variant_media_id":"media-variant-1","provenance":{"operation":"background_cleanup","source_media_id":"media-original-1"},"result_links":[{"type":"media","id":"media-variant-1","href":"/media"}]}
	}`), map[string]string{"Content-Type": "application/json"})
	if processDone.Code != http.StatusOK {
		t.Fatalf("complete image process run status=%d body=%s", processDone.Code, processDone.Body.String())
	}
	if !strings.Contains(processDone.Body.String(), `"confirmation_state":"pending"`) || !strings.Contains(processDone.Body.String(), `"original_media_id":"media-original-1"`) || !strings.Contains(processDone.Body.String(), `"variant_media_id":"media-variant-1"`) || !strings.Contains(processDone.Body.String(), `"source_media_id":"media-original-1"`) {
		t.Fatalf("image_process must preserve original media provenance and return pending-confirmation variant links, body=%s", processDone.Body.String())
	}
}
