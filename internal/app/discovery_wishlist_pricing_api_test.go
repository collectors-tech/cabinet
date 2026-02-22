package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestDiscoveryPanelActionsAndReset(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	if _, err := a.db.Exec(`INSERT INTO scanner_query_sets(id, name, keywords_json, exclusions_json) VALUES ('q1','Q','["afx"]','[]')`); err != nil {
		t.Fatalf("seed query set: %v", err)
	}
	if _, err := a.db.Exec(`INSERT INTO scanner_candidates(id, query_set_id, listing_id, title, price, shipping, url, image, seller, first_seen, last_seen, status, source) VALUES ('c1','q1','L1','AFX P-9',12,0,'http://x/1','','s',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP,'new','ebay')`); err != nil {
		t.Fatalf("seed candidate: %v", err)
	}
	if _, err := a.db.Exec(`INSERT INTO scanner_matches(candidate_id, item_id, state, confidence, needs_review, extracted_part_number, updated_at) VALUES ('c1','','not_in_collection',0,1,'P-9',CURRENT_TIMESTAMP)`); err != nil {
		t.Fatalf("seed match: %v", err)
	}
	list := doRequest(t, a, http.MethodGet, "/api/discovery/not-in-collection?price_max=20&q=afx", nil, nil)
	if list.Code != http.StatusOK {
		t.Fatalf("list not-in-collection status=%d body=%s", list.Code, list.Body.String())
	}
	act := doRequest(t, a, http.MethodPost, "/api/discovery/action", strings.NewReader(`{"candidate_id":"c1","type":"ignore"}`), map[string]string{"Content-Type": "application/json"})
	if act.Code != http.StatusOK {
		t.Fatalf("apply action status=%d body=%s", act.Code, act.Body.String())
	}
	reset := doRequest(t, a, http.MethodPost, "/api/settings/reset-ignore-rules", strings.NewReader(`{}`), map[string]string{"Content-Type": "application/json"})
	if reset.Code != http.StatusOK {
		t.Fatalf("reset ignore status=%d body=%s", reset.Code, reset.Body.String())
	}
}

func TestWishlistAndPricingEndpoints(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"P1"}`), map[string]string{"Content-Type": "application/json"})
	if createProfile.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", createProfile.Code, createProfile.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createProfile.Body).Decode(&profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	setActive := doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+profile.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if setActive.Code != http.StatusOK {
		t.Fatalf("set active profile status=%d body=%s", setActive.Code, setActive.Body.String())
	}
	if _, err := a.db.Exec(`INSERT INTO canonical_items(id, brand, category, part_number, title) VALUES ('i1','AFX','Slot','P-1','AFX P-1')`); err != nil {
		t.Fatalf("seed item: %v", err)
	}
	if _, err := a.db.Exec(`INSERT INTO scanner_query_sets(id, name, keywords_json, exclusions_json) VALUES ('q1','Q','["afx"]','[]')`); err != nil {
		t.Fatalf("seed query set: %v", err)
	}
	if _, err := a.db.Exec(`INSERT INTO scanner_candidates(id, query_set_id, listing_id, title, price, shipping, url, image, seller, first_seen, last_seen, status, source) VALUES ('c1','q1','L1','AFX P-1',15,0,'http://x/1','','s',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP,'new','ebay')`); err != nil {
		t.Fatalf("seed candidate: %v", err)
	}
	createWish := doRequest(t, a, http.MethodPost, "/api/wishlist", strings.NewReader(`{"item_id":"i1","target_price":20,"priority":"high","notes":"watch"}`), map[string]string{"Content-Type": "application/json"})
	if createWish.Code != http.StatusCreated {
		t.Fatalf("create wishlist status=%d body=%s", createWish.Code, createWish.Body.String())
	}
	var created map[string]any
	if err := json.NewDecoder(createWish.Body).Decode(&created); err != nil {
		t.Fatalf("decode wishlist create: %v", err)
	}
	wid, _ := created["id"].(string)
	if wid == "" {
		t.Fatal("expected wishlist id")
	}
	listWish := doRequest(t, a, http.MethodGet, "/api/wishlist", nil, nil)
	if listWish.Code != http.StatusOK {
		t.Fatalf("list wishlist status=%d body=%s", listWish.Code, listWish.Body.String())
	}
	hits := doRequest(t, a, http.MethodGet, "/api/wishlist/hits?item_id=i1", nil, nil)
	if hits.Code != http.StatusOK {
		t.Fatalf("wishlist hits status=%d body=%s", hits.Code, hits.Body.String())
	}
	track := doRequest(t, a, http.MethodPost, "/api/pricing/track", strings.NewReader(`{"item_id":"i1"}`), map[string]string{"Content-Type": "application/json"})
	if track.Code != http.StatusOK {
		t.Fatalf("track price status=%d body=%s", track.Code, track.Body.String())
	}
	run := doRequest(t, a, http.MethodPost, "/api/pricing/snapshot/run", strings.NewReader(`{}`), map[string]string{"Content-Type": "application/json"})
	if run.Code != http.StatusOK {
		t.Fatalf("run snapshot status=%d body=%s", run.Code, run.Body.String())
	}
	history := doRequest(t, a, http.MethodGet, "/api/pricing/history?item_id=i1", nil, nil)
	if history.Code != http.StatusOK {
		t.Fatalf("history status=%d body=%s", history.Code, history.Body.String())
	}
	graph := doRequest(t, a, http.MethodGet, "/api/pricing/graph?item_id=i1", nil, nil)
	if graph.Code != http.StatusOK {
		t.Fatalf("graph status=%d body=%s", graph.Code, graph.Body.String())
	}
	bySource := doRequest(t, a, http.MethodGet, "/api/pricing/by-source?item_id=i1", nil, nil)
	if bySource.Code != http.StatusOK {
		t.Fatalf("by-source status=%d body=%s", bySource.Code, bySource.Body.String())
	}
	stats := doRequest(t, a, http.MethodGet, "/api/pricing/stats?item_id=i1", nil, nil)
	if stats.Code != http.StatusOK {
		t.Fatalf("stats status=%d body=%s", stats.Code, stats.Body.String())
	}
	trend := doRequest(t, a, http.MethodGet, "/api/pricing/trend?item_id=i1", nil, nil)
	if trend.Code != http.StatusOK {
		t.Fatalf("trend status=%d body=%s", trend.Code, trend.Body.String())
	}
	export := doRequest(t, a, http.MethodGet, "/api/pricing/history/export?item_id=i1", nil, nil)
	if export.Code != http.StatusOK {
		t.Fatalf("export status=%d body=%s", export.Code, export.Body.String())
	}
	delWish := doRequest(t, a, http.MethodDelete, "/api/wishlist?id="+wid, nil, nil)
	if delWish.Code != http.StatusNoContent {
		t.Fatalf("delete wishlist status=%d body=%s", delWish.Code, delWish.Body.String())
	}
}
