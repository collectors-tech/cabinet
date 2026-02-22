package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestMatchingRunAndResultsEndpoints(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	if _, err := a.db.Exec(`INSERT INTO canonical_items (id, brand, category, part_number, title) VALUES ('i1','AFX','Slot','P-1','AFX P-1')`); err != nil {
		t.Fatalf("seed item: %v", err)
	}
	if _, err := a.db.Exec(`INSERT INTO scanner_query_sets (id, name, keywords_json, exclusions_json) VALUES ('q1','q','["afx"]','[]')`); err != nil {
		t.Fatalf("seed query set: %v", err)
	}
	if _, err := a.db.Exec(`INSERT INTO scanner_candidates (id, query_set_id, listing_id, title, price, shipping, url, image, seller, first_seen, last_seen, status, source) VALUES ('c1','q1','L1','AFX P-1',10,0,'http://x','','s',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP,'new','ebay')`); err != nil {
		t.Fatalf("seed candidate: %v", err)
	}
	run := doRequest(t, a, http.MethodPost, "/api/matching/run", strings.NewReader(`{}`), map[string]string{"Content-Type": "application/json"})
	if run.Code != http.StatusOK {
		t.Fatalf("matching run status=%d body=%s", run.Code, run.Body.String())
	}
	results := doRequest(t, a, http.MethodGet, "/api/matching/results", nil, nil)
	if results.Code != http.StatusOK {
		t.Fatalf("matching results status=%d body=%s", results.Code, results.Body.String())
	}
	var payload struct {
		Results []map[string]any `json:"results"`
	}
	if err := json.NewDecoder(results.Body).Decode(&payload); err != nil {
		t.Fatalf("decode results payload: %v", err)
	}
	if len(payload.Results) == 0 {
		t.Fatal("expected persisted matching results")
	}
}
