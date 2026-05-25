package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestCompanionModuleRegistryExposesPassiveModules(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	resp := doRequest(t, a, http.MethodGet, "/api/companion/modules", nil, nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("registry status=%d body=%s", resp.Code, resp.Body.String())
	}
	var payload struct {
		Modules []struct {
			ID          string   `json:"id"`
			Site        string   `json:"site"`
			Actions     []string `json:"actions"`
			PassiveOnly bool     `json:"passive_only"`
		} `json:"modules"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode registry: %v", err)
	}
	if len(payload.Modules) != 1 {
		t.Fatalf("expected one default module, got %+v", payload.Modules)
	}
	module := payload.Modules[0]
	if module.ID != "ebay-purchase-capture" || module.Site != "ebay" || !module.PassiveOnly || len(module.Actions) == 0 {
		t.Fatalf("unexpected companion module %+v", module)
	}
}

func TestCompanionPayloadAPIRequiresProfileScopedAuth(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	body := `{"profile_id":"profile-api","module_id":"ebay-purchase-capture","url":"https://www.ebay.com/itm/123","payload_type":"purchase_order","passive":true,"confidence_score":0.82}`
	resp := doRequest(t, a, http.MethodPost, "/api/companion/payloads", strings.NewReader(body), map[string]string{"Content-Type": "application/json"})
	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("expected missing auth to fail, status=%d body=%s", resp.Code, resp.Body.String())
	}
}

func TestCompanionPayloadAPIAcceptsOnlyPassiveCaptures(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	headers := map[string]string{"Content-Type": "application/json", "Authorization": "Bearer companion:profile-api"}
	writeBody := `{"profile_id":"profile-api","module_id":"ebay-purchase-capture","url":"https://www.ebay.com/itm/123","payload_type":"purchase_order","passive":false,"attempted_write":true,"confidence_score":0.82}`
	writeResp := doRequest(t, a, http.MethodPost, "/api/companion/payloads", strings.NewReader(writeBody), headers)
	if writeResp.Code != http.StatusBadRequest {
		t.Fatalf("expected write attempt to fail, status=%d body=%s", writeResp.Code, writeResp.Body.String())
	}

	passiveBody := `{"profile_id":"profile-api","module_id":"ebay-purchase-capture","url":"https://www.ebay.com/itm/123","payload_type":"purchase_order","passive":true,"confidence_score":0.82,"data":{"order_id":"123"}}`
	passiveResp := doRequest(t, a, http.MethodPost, "/api/companion/payloads", strings.NewReader(passiveBody), headers)
	if passiveResp.Code != http.StatusAccepted {
		t.Fatalf("expected passive payload accepted, status=%d body=%s", passiveResp.Code, passiveResp.Body.String())
	}
	var accepted struct {
		Accepted    bool     `json:"accepted"`
		SyncMode    string   `json:"sync_mode"`
		RemoteWrite bool     `json:"remote_write"`
		AuditTrail  []string `json:"audit_trail"`
	}
	if err := json.NewDecoder(passiveResp.Body).Decode(&accepted); err != nil {
		t.Fatalf("decode accepted payload: %v", err)
	}
	if !accepted.Accepted || accepted.SyncMode != "passive_capture" || accepted.RemoteWrite {
		t.Fatalf("unexpected accepted payload %+v", accepted)
	}
}
