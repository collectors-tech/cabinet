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

func TestEbaySellerOperationExecuteAllowsReadOnlySyncOnly(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	body := `{"operation":"orders","capability":"read_only","action":"sync","reference_id":"orders-local-sync"}`
	resp := doRequest(t, a, http.MethodPost, "/api/providers/ebay/seller-operations/execute", strings.NewReader(body), map[string]string{"Content-Type": "application/json"})
	if resp.Code != http.StatusOK {
		t.Fatalf("execute status=%d body=%s", resp.Code, resp.Body.String())
	}

	var payload struct {
		Provider  string `json:"provider"`
		Mode      string `json:"mode"`
		Execution struct {
			Operation   string `json:"operation"`
			Action      string `json:"action"`
			Allowed     bool   `json:"allowed"`
			RemoteWrite bool   `json:"remote_write"`
			Executed    bool   `json:"executed"`
			LocalOnly   bool   `json:"local_only"`
			Status      string `json:"status"`
			Blocker     string `json:"blocker"`
		} `json:"execution"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode execute payload: %v", err)
	}
	if payload.Provider != "ebay" || payload.Mode != "seller_operation_execute" {
		t.Fatalf("unexpected provider/mode: %+v", payload)
	}
	if payload.Execution.Operation != "sold_orders" || payload.Execution.Action != "sync" {
		t.Fatalf("seller execute did not normalize requested action: %+v", payload.Execution)
	}
	if !payload.Execution.Allowed || !payload.Execution.Executed || !payload.Execution.LocalOnly || payload.Execution.RemoteWrite {
		t.Fatalf("read-only seller sync should execute locally only, got %+v", payload.Execution)
	}
	if payload.Execution.Status != "read_only_sync_ready" || payload.Execution.Blocker != "" {
		t.Fatalf("expected local sync ready status with no blocker, got %+v", payload.Execution)
	}
}

func TestEbaySellerOperationExecuteRejectsRemoteWriteWithoutAdapter(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	body := `{"operation":"fulfilment","capability":"confirmed_api","action":"ship","reference_id":"order-1","confirmed":true}`
	resp := doRequest(t, a, http.MethodPost, "/api/providers/ebay/seller-operations/execute", strings.NewReader(body), map[string]string{"Content-Type": "application/json"})
	if resp.Code != http.StatusConflict {
		t.Fatalf("expected confirmed remote write execution to be blocked, status=%d body=%s", resp.Code, resp.Body.String())
	}

	var payload struct {
		Execution struct {
			Allowed     bool   `json:"allowed"`
			RemoteWrite bool   `json:"remote_write"`
			Executed    bool   `json:"executed"`
			Status      string `json:"status"`
			Blocker     string `json:"blocker"`
		} `json:"execution"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode blocked execute payload: %v", err)
	}
	if payload.Execution.Allowed || payload.Execution.Executed || !payload.Execution.RemoteWrite {
		t.Fatalf("remote write execution must remain blocked without adapter, got %+v", payload.Execution)
	}
	if payload.Execution.Status != "blocked" || payload.Execution.Blocker != "ebay_seller_remote_write_execution_not_configured" {
		t.Fatalf("expected remote write adapter blocker, got %+v", payload.Execution)
	}
}
