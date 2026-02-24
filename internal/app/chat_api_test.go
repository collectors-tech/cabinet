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

	threadResp := doRequest(t, a, http.MethodPost, "/api/chat/threads", strings.NewReader(`{"profile_id":"`+p.ID+`","title":"Main Thread"}`), map[string]string{"Content-Type": "application/json"})
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

	msgResp := doRequest(t, a, http.MethodPost, "/api/chat/messages", strings.NewReader(`{"profile_id":"`+p.ID+`","thread_id":"`+thread.ID+`","role":"user","content":"hello"}`), map[string]string{"Content-Type": "application/json"})
	if msgResp.Code != http.StatusCreated {
		t.Fatalf("create message status=%d body=%s", msgResp.Code, msgResp.Body.String())
	}

	msgList := doRequest(t, a, http.MethodGet, "/api/chat/messages?profile_id="+p.ID+"&thread_id="+thread.ID, nil, nil)
	if msgList.Code != http.StatusOK {
		t.Fatalf("list messages status=%d body=%s", msgList.Code, msgList.Body.String())
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
