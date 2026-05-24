package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestEbaySellerOperationPreviewBlocksUnverifiedWrite(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	body := "{\"operation\":\"messages\",\"capability\":\"read_only\",\"action\":\"reply\",\"confirmed\":true,\"reference_id\":\"msg-1\"}"
	resp := doRequest(t, a, http.MethodPost, "/api/providers/ebay/seller-operations/preview", strings.NewReader(body), map[string]string{"Content-Type": "application/json"})
	if resp.Code != http.StatusOK {
		t.Fatalf("preview status=%d body=%s", resp.Code, resp.Body.String())
	}

	var payload struct {
		Provider string `json:"provider"`
		Mode     string `json:"mode"`
		Preview  struct {
			Operation      string `json:"operation"`
			Action         string `json:"action"`
			ReadAvailable  bool   `json:"read_available"`
			WriteAvailable bool   `json:"write_available"`
			Allowed        bool   `json:"allowed"`
			RemoteWrite    bool   `json:"remote_write"`
			Blocker        string `json:"blocker"`
		} `json:"preview"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode preview payload: %v", err)
	}
	if payload.Provider != "ebay" || payload.Mode != "seller_operation_preview" {
		t.Fatalf("unexpected provider/mode: %+v", payload)
	}
	if payload.Preview.Operation != "messages" || payload.Preview.Action != "reply" {
		t.Fatalf("seller preview did not normalize requested action: %+v", payload.Preview)
	}
	if !payload.Preview.ReadAvailable || payload.Preview.WriteAvailable || payload.Preview.Allowed || payload.Preview.RemoteWrite {
		t.Fatalf("read-only seller reply must stay blocked from remote write: %+v", payload.Preview)
	}
	if payload.Preview.Blocker != "ebay_write_capability_not_verified" {
		t.Fatalf("expected explicit write blocker, got %+v", payload.Preview)
	}
}

func TestEbaySellerOperationPreviewRequiresConfirmationBeforeRemoteWrite(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	resp := doRequest(t, a, http.MethodPost, "/api/providers/ebay/seller-operations/preview", strings.NewReader("{\"operation\":\"fulfilment\",\"capability\":\"confirmed_api\",\"action\":\"ship\",\"reference_id\":\"order-1\"}"), map[string]string{"Content-Type": "application/json"})
	if resp.Code != http.StatusOK {
		t.Fatalf("preview status=%d body=%s", resp.Code, resp.Body.String())
	}
	var unconfirmed struct {
		Preview struct {
			Allowed              bool   `json:"allowed"`
			RemoteWrite          bool   `json:"remote_write"`
			ConfirmationRequired bool   `json:"confirmation_required"`
			Blocker              string `json:"blocker"`
		} `json:"preview"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&unconfirmed); err != nil {
		t.Fatalf("decode unconfirmed preview: %v", err)
	}
	if unconfirmed.Preview.Allowed || unconfirmed.Preview.RemoteWrite || !unconfirmed.Preview.ConfirmationRequired || unconfirmed.Preview.Blocker != "ebay_seller_action_confirmation_required" {
		t.Fatalf("confirmed API write must be blocked until confirmation: %+v", unconfirmed.Preview)
	}

	confirmed := doRequest(t, a, http.MethodPost, "/api/providers/ebay/seller-operations/preview", strings.NewReader("{\"operation\":\"fulfilment\",\"capability\":\"confirmed_api\",\"action\":\"ship\",\"reference_id\":\"order-1\",\"confirmed\":true}"), map[string]string{"Content-Type": "application/json"})
	if confirmed.Code != http.StatusOK {
		t.Fatalf("confirmed preview status=%d body=%s", confirmed.Code, confirmed.Body.String())
	}
	var confirmedPayload struct {
		Preview struct {
			Operation   string `json:"operation"`
			Action      string `json:"action"`
			Allowed     bool   `json:"allowed"`
			RemoteWrite bool   `json:"remote_write"`
			Blocker     string `json:"blocker"`
		} `json:"preview"`
	}
	if err := json.NewDecoder(confirmed.Body).Decode(&confirmedPayload); err != nil {
		t.Fatalf("decode confirmed preview: %v", err)
	}
	if confirmedPayload.Preview.Operation != "fulfillment" || confirmedPayload.Preview.Action != "fulfill" || !confirmedPayload.Preview.Allowed || !confirmedPayload.Preview.RemoteWrite || confirmedPayload.Preview.Blocker != "" {
		t.Fatalf("confirmed seller write preview did not allow expected remote write: %+v", confirmedPayload.Preview)
	}
}

func TestEbaySellerOperationPreviewRejectsMissingOperationOrAction(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	resp := doRequest(t, a, http.MethodPost, "/api/providers/ebay/seller-operations/preview", strings.NewReader("{\"operation\":\"unknown\",\"action\":\"reply\"}"), map[string]string{"Content-Type": "application/json"})
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid seller preview to be rejected, status=%d body=%s", resp.Code, resp.Body.String())
	}
}
