package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestPokemonPriceAlertsRejectsMissingSetID(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	resp := doRequest(t, a, http.MethodGet, "/api/integrations/pokemon/price-alerts", nil, nil)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for missing set_id, got %d body=%s", resp.Code, resp.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["error"] != "missing_set_id" {
		t.Fatalf("expected error missing_set_id, got %#v", payload["error"])
	}
}

func TestPokemonPriceAlertsReturnsMultiSourceStatsAndThresholdAlerts(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	create := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/items",
		strings.NewReader(`{"part_number":"PKM-PRICE-001","title":"Charizard Base","tags":["set:base-set","variant:holo","language:en"],"grading_status":"graded"}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if create.Code != http.StatusCreated {
		t.Fatalf("create item status=%d body=%s", create.Code, create.Body.String())
	}
	var item map[string]any
	if err := json.Unmarshal(create.Body.Bytes(), &item); err != nil {
		t.Fatalf("decode create payload: %v", err)
	}
	itemID, _ := item["id"].(string)
	if strings.TrimSpace(itemID) == "" {
		t.Fatalf("expected item id in payload: %#v", item)
	}

	conn := a.db
	_, err := conn.Exec(`
		INSERT INTO price_snapshots (id, item_id, snapshot_date, source, min_price, median_price, latest_price, stock_count, created_at)
		VALUES
		  ('snap-ebay-prev', ?, '2026-03-01', 'ebay', 95, 100, 100, 2, CURRENT_TIMESTAMP),
		  ('snap-ebay-latest', ?, '2026-03-02', 'ebay', 65, 70, 70, 1, CURRENT_TIMESTAMP),
		  ('snap-amazon-prev', ?, '2026-03-01', 'amazon', 108, 110, 110, 4, CURRENT_TIMESTAMP),
		  ('snap-amazon-latest', ?, '2026-03-02', 'amazon', 100, 105, 105, 3, CURRENT_TIMESTAMP)
	`, itemID, itemID, itemID, itemID)
	if err != nil {
		t.Fatalf("seed price snapshots: %v", err)
	}

	resp := doRequest(t, a, http.MethodGet, "/api/integrations/pokemon/price-alerts?set_id=base-set&drop_threshold_pct=20", nil, nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", resp.Code, resp.Body.String())
	}

	var payload map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload["set_id"] != "base-set" {
		t.Fatalf("expected set_id base-set, got %#v", payload["set_id"])
	}
	sources, ok := payload["sources"].([]any)
	if !ok || len(sources) < 2 {
		t.Fatalf("expected >=2 sources, got %#v", payload["sources"])
	}
	alerts, ok := payload["alerts"].([]any)
	if !ok || len(alerts) == 0 {
		t.Fatalf("expected threshold alerts, got %#v", payload["alerts"])
	}
	firstAlert, ok := alerts[0].(map[string]any)
	if !ok {
		t.Fatalf("expected alert object, got %#v", alerts[0])
	}
	if firstAlert["item_id"] != itemID {
		t.Fatalf("expected alert item_id=%s, got %#v", itemID, firstAlert["item_id"])
	}
	if _, ok := firstAlert["change_pct"].(float64); !ok {
		t.Fatalf("expected numeric change_pct, got %#v", firstAlert["change_pct"])
	}
	if firstAlert["threshold_pct"] != float64(20) {
		t.Fatalf("expected threshold_pct=20, got %#v", firstAlert["threshold_pct"])
	}
}
