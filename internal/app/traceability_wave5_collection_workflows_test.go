package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestWave5SearchFiltersAndSavedFiltersProfileScoped(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createP1 := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Wave5P1"}`), map[string]string{"Content-Type": "application/json"})
	createP2 := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Wave5P2"}`), map[string]string{"Content-Type": "application/json"})
	if createP1.Code != http.StatusCreated || createP2.Code != http.StatusCreated {
		t.Fatalf("create profile failed p1=%d p2=%d", createP1.Code, createP2.Code)
	}
	var p1 struct{ ID string `json:"id"` }
	var p2 struct{ ID string `json:"id"` }
	_ = json.NewDecoder(createP1.Body).Decode(&p1)
	_ = json.NewDecoder(createP2.Body).Decode(&p2)

	_ = doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+p1.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	createP1Item := doRequest(t, a, http.MethodPost, "/api/items", strings.NewReader(`{"part_number":"W5-SRCH-001","title":"AFX Search Item","brand":"AFX","category":"Slot","tags":["collector"]}`), map[string]string{"Content-Type": "application/json"})
	if createP1Item.Code != http.StatusCreated {
		t.Fatalf("create p1 item status=%d body=%s", createP1Item.Code, createP1Item.Body.String())
	}

	searchP1 := doRequest(t, a, http.MethodGet, "/api/search/items?q=AFX&brand=AFX&sort=part_number", nil, nil)
	if searchP1.Code != http.StatusOK {
		t.Fatalf("search p1 status=%d body=%s", searchP1.Code, searchP1.Body.String())
	}
	var p1Results struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.NewDecoder(searchP1.Body).Decode(&p1Results); err != nil {
		t.Fatalf("decode p1 search: %v", err)
	}
	if len(p1Results.Items) == 0 {
		t.Fatal("expected p1 filtered search results")
	}

	saveFilter := doRequest(
		t,
		a,
		http.MethodPost,
		"/api/profiles/"+p1.ID+"/saved-filters",
		strings.NewReader(`{"name":"AFX Only","query":{"q":"AFX","brand":"AFX","sort":"part_number"}}`),
		map[string]string{"Content-Type": "application/json"},
	)
	if saveFilter.Code != http.StatusCreated {
		t.Fatalf("save filter status=%d body=%s", saveFilter.Code, saveFilter.Body.String())
	}

	listP1Filters := doRequest(t, a, http.MethodGet, "/api/profiles/"+p1.ID+"/saved-filters", nil, nil)
	if listP1Filters.Code != http.StatusOK {
		t.Fatalf("list p1 filters status=%d body=%s", listP1Filters.Code, listP1Filters.Body.String())
	}
	var p1Filters struct {
		Saved []map[string]any `json:"saved_filters"`
	}
	if err := json.NewDecoder(listP1Filters.Body).Decode(&p1Filters); err != nil {
		t.Fatalf("decode p1 filters: %v", err)
	}
	if len(p1Filters.Saved) != 1 {
		t.Fatalf("expected one p1 saved filter, got %d", len(p1Filters.Saved))
	}

	_ = doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+p2.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	searchP2 := doRequest(t, a, http.MethodGet, "/api/search/items?brand=AFX", nil, nil)
	if searchP2.Code != http.StatusOK {
		t.Fatalf("search p2 status=%d body=%s", searchP2.Code, searchP2.Body.String())
	}
	var p2Results struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.NewDecoder(searchP2.Body).Decode(&p2Results); err != nil {
		t.Fatalf("decode p2 search: %v", err)
	}
	if len(p2Results.Items) != 0 {
		t.Fatalf("expected profile-scoped empty search for p2, got %+v", p2Results.Items)
	}
}

func TestWave5DiscoveryWishlistPricingDashboardContracts(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Wave5Runtime"}`), map[string]string{"Content-Type": "application/json"})
	if createProfile.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", createProfile.Code, createProfile.Body.String())
	}
	var profile struct{ ID string `json:"id"` }
	_ = json.NewDecoder(createProfile.Body).Decode(&profile)
	_ = doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+profile.ID+`"}`), map[string]string{"Content-Type": "application/json"})

	if _, err := a.db.Exec(`INSERT INTO canonical_items(id, profile_id, brand, category, part_number, title) VALUES ('w5-item-1', ?, 'AFX','Slot','W5-001','Wave5 Item')`, profile.ID); err != nil {
		t.Fatalf("seed item: %v", err)
	}
	if _, err := a.db.Exec(`INSERT INTO instances(id, item_id, condition, status, quantity, storage_location, acquisition_price, acquisition_date, notes) VALUES ('w5-inst-1','w5-item-1','new','sealed',1,'shelf',45.00,'2026-02-01','')`); err != nil {
		t.Fatalf("seed instance: %v", err)
	}
	if _, err := a.db.Exec(`INSERT INTO scanner_query_sets(id, profile_id, name, keywords_json, exclusions_json) VALUES ('w5-q1', ?, 'W5 Query', '["afx"]', '[]')`, profile.ID); err != nil {
		t.Fatalf("seed query set: %v", err)
	}
	if _, err := a.db.Exec(`INSERT INTO scanner_candidates(id, profile_id, query_set_id, listing_id, title, price, shipping, url, image, seller, first_seen, last_seen, status, source, stock_state, stock_count) VALUES ('w5-c1', ?, 'w5-q1', 'W5-L1', 'Wave5 Discovery', 30, 0, 'https://example/listing', '', 'seller', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'new', 'ebay', 'low_stock', 2)`, profile.ID); err != nil {
		t.Fatalf("seed candidate: %v", err)
	}
	if _, err := a.db.Exec(`INSERT INTO scanner_matches(candidate_id, item_id, state, confidence, needs_review, extracted_part_number, updated_at) VALUES ('w5-c1', '', 'not_in_collection', 0.5, 1, 'W5-EXT', CURRENT_TIMESTAMP)`); err != nil {
		t.Fatalf("seed match: %v", err)
	}

	discovery := doRequest(t, a, http.MethodGet, "/api/discovery/not-in-collection?price_max=50&q=wave5&date_from=2020-01-01", nil, nil)
	if discovery.Code != http.StatusOK {
		t.Fatalf("discovery list status=%d body=%s", discovery.Code, discovery.Body.String())
	}
	var discoveryPayload struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.NewDecoder(discovery.Body).Decode(&discoveryPayload); err != nil {
		t.Fatalf("decode discovery payload: %v", err)
	}
	if len(discoveryPayload.Items) == 0 {
		t.Fatal("expected discovery triage items")
	}

	for _, actionType := range []string{"ignore", "add_to_wishlist", "track_price", "create_item"} {
		action := doRequest(t, a, http.MethodPost, "/api/discovery/action", strings.NewReader(`{"candidate_id":"w5-c1","type":"`+actionType+`"}`), map[string]string{"Content-Type": "application/json"})
		if action.Code != http.StatusOK {
			t.Fatalf("discovery action %s status=%d body=%s", actionType, action.Code, action.Body.String())
		}
	}

	createWish := doRequest(t, a, http.MethodPost, "/api/wishlist", strings.NewReader(`{"item_id":"w5-item-1","target_price":25,"priority":"medium","notes":"watch wave5"}`), map[string]string{"Content-Type": "application/json"})
	if createWish.Code != http.StatusCreated {
		t.Fatalf("create wishlist status=%d body=%s", createWish.Code, createWish.Body.String())
	}
	track := doRequest(t, a, http.MethodPost, "/api/pricing/track", strings.NewReader(`{"item_id":"w5-item-1"}`), map[string]string{"Content-Type": "application/json"})
	if track.Code != http.StatusOK {
		t.Fatalf("track pricing status=%d body=%s", track.Code, track.Body.String())
	}
	snapshot := doRequest(t, a, http.MethodPost, "/api/pricing/snapshot/run", strings.NewReader(`{}`), map[string]string{"Content-Type": "application/json"})
	if snapshot.Code != http.StatusOK {
		t.Fatalf("snapshot status=%d body=%s", snapshot.Code, snapshot.Body.String())
	}
	export := doRequest(t, a, http.MethodGet, "/api/pricing/history/export?item_id=w5-item-1", nil, nil)
	if export.Code != http.StatusOK {
		t.Fatalf("pricing export status=%d body=%s", export.Code, export.Body.String())
	}
	dashboard := doRequest(t, a, http.MethodGet, "/api/dashboard", nil, nil)
	if dashboard.Code != http.StatusOK {
		t.Fatalf("dashboard status=%d body=%s", dashboard.Code, dashboard.Body.String())
	}
	var dashboardPayload map[string]any
	if err := json.NewDecoder(dashboard.Body).Decode(&dashboardPayload); err != nil {
		t.Fatalf("decode dashboard payload: %v", err)
	}
	for _, field := range []string{"new_discoveries", "wishlist_hits", "price_drops", "total_items", "total_instances"} {
		if _, ok := dashboardPayload[field]; !ok {
			t.Fatalf("dashboard payload missing %q: %+v", field, dashboardPayload)
		}
	}
}

func TestWave5MatchingClassificationAndPartExtractionContract(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	if _, err := a.db.Exec(`INSERT INTO canonical_items(id, brand, category, part_number, title) VALUES ('w5-mi-1','AFX','Slot','W5-PN-1','Wave5 Match Item')`); err != nil {
		t.Fatalf("seed canonical item: %v", err)
	}
	if _, err := a.db.Exec(`INSERT INTO scanner_query_sets(id, name, keywords_json, exclusions_json) VALUES ('w5-mq-1','W5 Match Query','["afx"]','[]')`); err != nil {
		t.Fatalf("seed query set: %v", err)
	}
	if _, err := a.db.Exec(`INSERT INTO scanner_candidates(id, query_set_id, listing_id, title, price, shipping, url, image, seller, first_seen, last_seen, status, source) VALUES ('w5-mc-1','w5-mq-1','W5-ML-1','AFX W5-PN-1 sealed',10,0,'http://x','','seller',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP,'new','ebay')`); err != nil {
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
		t.Fatalf("decode matching results: %v", err)
	}
	if len(payload.Results) == 0 {
		t.Fatal("expected matching results")
	}
	first := payload.Results[0]
	state, _ := first["state"].(string)
	switch state {
	case "matched", "suggested", "not_in_collection":
	default:
		t.Fatalf("unexpected matching state %q payload=%+v", state, first)
	}
	if strings.TrimSpace(first["part_number"].(string)) == "" {
		t.Fatalf("expected extracted part_number in matching result: %+v", first)
	}
}
