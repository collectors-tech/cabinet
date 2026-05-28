package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestCommerceLandedCostPlanAPI(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	body := `{
		"items": [
			{"id":"card-b","purchase_cents":30000,"domestic_shipping_cents":1500,"tax_cents":3000,"weight_grams":300},
			{"id":"card-a","purchase_cents":10000,"domestic_shipping_cents":500,"tax_cents":1000,"weight_grams":100}
		],
		"components": [
			{"id":"intl","label":"International shipping","amount_cents":8000,"allocation_method":"weight","provenance":"forwarder-shipment:SHIP-1"},
			{"id":"handling","label":"Handling","amount_cents":1200,"allocation_method":"equal","provenance":"forwarder-invoice:INV-1"}
		],
		"consolidation": {
			"shipment_fee_cents":2500,
			"destination_limit_cents":60000,
			"warning_buffer_cents":1500
		}
	}`

	resp := doRequest(t, a, http.MethodPost, "/api/commerce/landed-cost/plan", strings.NewReader(body), map[string]string{"Content-Type": "application/json"})
	if resp.Code != http.StatusOK {
		t.Fatalf("landed-cost plan status=%d body=%s", resp.Code, resp.Body.String())
	}

	var payload struct {
		Mode       string `json:"mode"`
		Mutable    bool   `json:"mutable"`
		Allocation struct {
			TotalDirectCents int64 `json:"total_direct_cents"`
			TotalSharedCents int64 `json:"total_shared_cents"`
			TotalLandedCents int64 `json:"total_landed_cents"`
			Items            []struct {
				ItemID             string `json:"item_id"`
				LandedCostCents    int64  `json:"landed_cost_cents"`
				AllocatedCostCents int64  `json:"allocated_cost_cents"`
			} `json:"items"`
		} `json:"allocation"`
		Consolidation struct {
			ItemIDs             []string `json:"item_ids"`
			EstimatedTotalCents int64    `json:"estimated_total_cents"`
			ThresholdState      string   `json:"threshold_state"`
			Mutable             bool     `json:"mutable"`
		} `json:"consolidation"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode landed-cost plan payload: %v", err)
	}
	if payload.Mode != "landed_cost_plan" || payload.Mutable {
		t.Fatalf("expected non-mutating landed cost plan mode, got %+v", payload)
	}
	if payload.Allocation.TotalDirectCents != 46000 || payload.Allocation.TotalSharedCents != 9200 || payload.Allocation.TotalLandedCents != 55200 {
		t.Fatalf("unexpected allocation totals: %+v", payload.Allocation)
	}
	if len(payload.Allocation.Items) != 2 || payload.Allocation.Items[0].ItemID != "card-b" || payload.Allocation.Items[0].LandedCostCents != 41100 {
		t.Fatalf("allocation items should preserve explainable per-item landed costs, got %+v", payload.Allocation.Items)
	}
	if got := payload.Consolidation.ItemIDs; len(got) != 2 || got[0] != "card-a" || got[1] != "card-b" {
		t.Fatalf("expected deterministic sorted consolidation item ids, got %#v", got)
	}
	if payload.Consolidation.EstimatedTotalCents != 57700 || payload.Consolidation.ThresholdState != "under_limit" || payload.Consolidation.Mutable {
		t.Fatalf("unexpected non-mutating consolidation plan: %+v", payload.Consolidation)
	}
}

func TestCommerceLandedCostPlanAPIRejectsInvalidManualShare(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	body := `{
		"items": [{"id":"card-a","purchase_cents":1000}],
		"components": [{"id":"manual","amount_cents":100,"allocation_method":"manual","manual_shares":{"missing":1}}]
	}`

	resp := doRequest(t, a, http.MethodPost, "/api/commerce/landed-cost/plan", strings.NewReader(body), map[string]string{"Content-Type": "application/json"})
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid manual share rejection, status=%d body=%s", resp.Code, resp.Body.String())
	}
}
