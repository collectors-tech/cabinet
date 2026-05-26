package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestDashboardEndpoint(t *testing.T) {
	t.Parallel()
	a := newTestApp(t)
	if _, err := a.db.Exec(`INSERT INTO canonical_items(id, brand, category, part_number, title) VALUES ('i1','AFX','Slot','P-1','AFX P-1')`); err != nil {
		t.Fatalf("seed item: %v", err)
	}
	if _, err := a.db.Exec(`INSERT INTO instances(id, item_id, condition, status, quantity, storage_location, acquisition_price, acquisition_date, notes) VALUES ('in1','i1','used','loose',1,'shelf',10,'','')`); err != nil {
		t.Fatalf("seed instance: %v", err)
	}
	resp := doRequest(t, a, http.MethodGet, "/api/dashboard", nil, nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("dashboard status=%d body=%s", resp.Code, resp.Body.String())
	}
}

func TestDashboardEndpointContractForUI(t *testing.T) {
	t.Parallel()
	a := newTestApp(t)

	resp := doRequest(t, a, http.MethodGet, "/api/dashboard", nil, nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("dashboard status=%d body=%s", resp.Code, resp.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	required := []string{"new_discoveries", "wishlist_hits", "price_drops", "low_stock_discoveries", "restocks", "recently_added", "total_items", "total_instances"}
	for _, key := range required {
		if _, ok := body[key]; !ok {
			t.Fatalf("missing required key %q in dashboard payload: %#v", key, body)
		}
	}
}

func TestDashboardEndpointScopesToActiveProfile(t *testing.T) {
	t.Parallel()
	a := newTestApp(t)

	createA := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Dashboard A"}`), map[string]string{"Content-Type": "application/json"})
	if createA.Code != http.StatusCreated {
		t.Fatalf("create profile a status=%d body=%s", createA.Code, createA.Body.String())
	}
	var profileA struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(createA.Body.Bytes(), &profileA); err != nil {
		t.Fatalf("decode profile a: %v", err)
	}
	createB := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Dashboard B"}`), map[string]string{"Content-Type": "application/json"})
	if createB.Code != http.StatusCreated {
		t.Fatalf("create profile b status=%d body=%s", createB.Code, createB.Body.String())
	}
	var profileB struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(createB.Body.Bytes(), &profileB); err != nil {
		t.Fatalf("decode profile b: %v", err)
	}

	if _, err := a.db.Exec(`
		INSERT INTO canonical_items(id, profile_id, brand, category, part_number, title, created_at)
		VALUES
			('dash-item-a', ?, 'AFX', 'Slot', 'DA-1', 'Dashboard A Item', '2026-05-01T10:00:00Z'),
			('dash-item-b', ?, 'AFX', 'Slot', 'DB-1', 'Dashboard B Item', '2026-05-02T10:00:00Z')
	`, profileA.ID, profileB.ID); err != nil {
		t.Fatalf("seed dashboard items: %v", err)
	}
	if _, err := a.db.Exec(`
		INSERT INTO instances(id, item_id, condition, status, quantity, storage_location, acquisition_price, acquisition_date, notes)
		VALUES
			('dash-inst-a','dash-item-a','used','loose',1,'shelf',10,'',''),
			('dash-inst-b','dash-item-b','used','loose',9,'case',30,'','')
	`); err != nil {
		t.Fatalf("seed dashboard instances: %v", err)
	}

	activate := doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+profileB.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if activate.Code != http.StatusOK {
		t.Fatalf("activate profile b status=%d body=%s", activate.Code, activate.Body.String())
	}

	resp := doRequest(t, a, http.MethodGet, "/api/dashboard", nil, nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("dashboard status=%d body=%s", resp.Code, resp.Body.String())
	}
	var body struct {
		TotalItems     int      `json:"total_items"`
		TotalInstances int      `json:"total_instances"`
		RecentlyAdded  []string `json:"recently_added"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode dashboard: %v", err)
	}
	if body.TotalItems != 1 || body.TotalInstances != 9 {
		t.Fatalf("expected active profile dashboard totals, got %+v", body)
	}
	if len(body.RecentlyAdded) != 1 || body.RecentlyAdded[0] != "Dashboard B Item" {
		t.Fatalf("expected active profile recent items, got %+v", body.RecentlyAdded)
	}
}
