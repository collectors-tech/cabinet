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
	if got := seen["collections.item.assign"]; got.mode != "preview-only" || got.permission != "preview-only" || got.unavailable {
		t.Fatalf("collections assignment must expose preview-only boundary, got %+v", got)
	}
	if got := seen["integrations.provider.run"]; got.mode != "unavailable" || got.permission != "setup-needed" || !got.unavailable {
		t.Fatalf("provider runs must be setup-needed/unavailable until connected, got %+v", got)
	}
	content := seen["content_generate"]
	if content.mode != "unavailable" || content.permission != "setup-needed" || !content.unavailable || content.previewShape != "catalog_content_draft_preview" || content.applyBehavior != "preview_only_no_mutation" || !slices.Contains(content.providerRequires, "openai") {
		t.Fatalf("content_generate must expose preview-only setup-needed OpenAI contract, got %+v", content)
	}
	listing := seen["listing_draft_generate"]
	if listing.mode != "unavailable" || listing.permission != "setup-needed" || !listing.unavailable || listing.previewShape != "listing_draft_preview_with_sources" || listing.applyBehavior != "requires_explicit_confirmation" || !slices.Contains(listing.providerRequires, "provider_test_passed") {
		t.Fatalf("listing_draft_generate must expose confirmation-gated setup-needed OpenAI contract, got %+v", listing)
	}

	missingProfile := doRequest(t, a, http.MethodGet, "/api/chat/capabilities", nil, nil)
	if missingProfile.Code != http.StatusBadRequest {
		t.Fatalf("missing profile status=%d body=%s", missingProfile.Code, missingProfile.Body.String())
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

	itemResp := doRequest(t, a, http.MethodPost, "/api/items", strings.NewReader(`{"part_number":"GEN-001","title":"Original catalog title","brand":"AFX","category":"Cars","condition":"Used"}`), map[string]string{"Content-Type": "application/json"})
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
