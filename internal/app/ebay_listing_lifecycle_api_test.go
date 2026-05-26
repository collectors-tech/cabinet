package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestEbayListingLifecyclePreviewExposesLocalDraftOnly(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	body := `{"command":"draft","capability":"draft_only","item_id":"item-1","title":"Listing title"}`
	resp := doRequest(t, a, http.MethodPost, "/api/providers/ebay/listing-lifecycle/preview", strings.NewReader(body), map[string]string{"Content-Type": "application/json"})
	if resp.Code != http.StatusOK {
		t.Fatalf("preview status=%d body=%s", resp.Code, resp.Body.String())
	}

	var payload struct {
		Provider string `json:"provider"`
		Mode     string `json:"mode"`
		Preview  struct {
			Command     string `json:"command"`
			Capability  string `json:"capability"`
			Allowed     bool   `json:"allowed"`
			LocalOnly   bool   `json:"local_only"`
			RemoteWrite bool   `json:"remote_write"`
			Blocker     string `json:"blocker"`
		} `json:"preview"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode preview payload: %v", err)
	}
	if payload.Provider != "ebay" || payload.Mode != "listing_lifecycle_preview" {
		t.Fatalf("unexpected provider/mode: %+v", payload)
	}
	if payload.Preview.Command != "draft" || payload.Preview.Capability != "draft_only" {
		t.Fatalf("draft preview did not normalize contract fields: %+v", payload.Preview)
	}
	if !payload.Preview.Allowed || !payload.Preview.LocalOnly || payload.Preview.RemoteWrite || payload.Preview.Blocker != "" {
		t.Fatalf("draft preview should be allowed local-only without remote write: %+v", payload.Preview)
	}
}

func TestEbayListingLifecyclePreviewRequiresConfirmationBeforeRemoteWrite(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	body := `{"command":"publish","capability":"confirmed_api","draft_id":"draft-1"}`
	resp := doRequest(t, a, http.MethodPost, "/api/providers/ebay/listing-lifecycle/preview", strings.NewReader(body), map[string]string{"Content-Type": "application/json"})
	if resp.Code != http.StatusOK {
		t.Fatalf("preview status=%d body=%s", resp.Code, resp.Body.String())
	}
	var payload struct {
		Preview struct {
			Allowed              bool   `json:"allowed"`
			RemoteWrite          bool   `json:"remote_write"`
			ConfirmationRequired bool   `json:"confirmation_required"`
			Blocker              string `json:"blocker"`
		} `json:"preview"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode unconfirmed preview: %v", err)
	}
	if payload.Preview.Allowed || !payload.Preview.RemoteWrite || !payload.Preview.ConfirmationRequired || payload.Preview.Blocker != "ebay_listing_lifecycle_confirmation_required" {
		t.Fatalf("publish preview must require confirmation before remote write: %+v", payload.Preview)
	}
}

func TestEbayListingLifecycleExecuteAllowsLocalDraftAndBlocksRemoteAdapter(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	draft := doRequest(t, a, http.MethodPost, "/api/providers/ebay/listing-lifecycle/execute", strings.NewReader(`{"command":"draft","capability":"draft_only","item_id":"item-1","title":"Listing title"}`), map[string]string{"Content-Type": "application/json"})
	if draft.Code != http.StatusOK {
		t.Fatalf("draft execute status=%d body=%s", draft.Code, draft.Body.String())
	}
	var draftPayload struct {
		Execution struct {
			Command     string `json:"command"`
			Allowed     bool   `json:"allowed"`
			LocalOnly   bool   `json:"local_only"`
			RemoteWrite bool   `json:"remote_write"`
			Executed    bool   `json:"executed"`
			Status      string `json:"status"`
			Response    struct {
				Provider string `json:"provider"`
				DraftID  string `json:"draft_id"`
				Status   string `json:"status"`
			} `json:"response"`
		} `json:"execution"`
	}
	if err := json.NewDecoder(draft.Body).Decode(&draftPayload); err != nil {
		t.Fatalf("decode draft execute: %v", err)
	}
	if draftPayload.Execution.Command != "draft" || !draftPayload.Execution.Allowed || !draftPayload.Execution.Executed || !draftPayload.Execution.LocalOnly || draftPayload.Execution.RemoteWrite {
		t.Fatalf("draft execute should complete as local-only: %+v", draftPayload.Execution)
	}
	if draftPayload.Execution.Status != "local_draft_ready" || draftPayload.Execution.Response.Provider != "cabinet" || draftPayload.Execution.Response.DraftID == "" {
		t.Fatalf("draft execute should return local draft response: %+v", draftPayload.Execution)
	}

	remote := doRequest(t, a, http.MethodPost, "/api/providers/ebay/listing-lifecycle/execute", strings.NewReader(`{"command":"publish","capability":"confirmed_api","draft_id":"draft-1","confirmed":true}`), map[string]string{"Content-Type": "application/json"})
	if remote.Code != http.StatusConflict {
		t.Fatalf("expected remote execution to be blocked without adapter, status=%d body=%s", remote.Code, remote.Body.String())
	}
	var remotePayload struct {
		Execution struct {
			Allowed     bool   `json:"allowed"`
			RemoteWrite bool   `json:"remote_write"`
			Executed    bool   `json:"executed"`
			Status      string `json:"status"`
			Blocker     string `json:"blocker"`
		} `json:"execution"`
	}
	if err := json.NewDecoder(remote.Body).Decode(&remotePayload); err != nil {
		t.Fatalf("decode remote execute: %v", err)
	}
	if remotePayload.Execution.Allowed || remotePayload.Execution.Executed || !remotePayload.Execution.RemoteWrite {
		t.Fatalf("remote write should remain blocked without adapter: %+v", remotePayload.Execution)
	}
	if remotePayload.Execution.Status != "blocked" || remotePayload.Execution.Blocker != "ebay_listing_lifecycle_adapter_required" {
		t.Fatalf("expected adapter blocker, got %+v", remotePayload.Execution)
	}
}

func TestEbayListingLifecyclePreviewRejectsMissingCommand(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	resp := doRequest(t, a, http.MethodPost, "/api/providers/ebay/listing-lifecycle/preview", strings.NewReader(`{"command":"unknown"}`), map[string]string{"Content-Type": "application/json"})
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected missing command rejection, status=%d body=%s", resp.Code, resp.Body.String())
	}
}
