package app

import (
	"net/http"
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
