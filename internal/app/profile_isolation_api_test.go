package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestCollectionDataIsIsolatedByActiveProfile(t *testing.T) {
	a := newTestApp(t)

	createP1 := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"P1"}`), map[string]string{"Content-Type": "application/json"})
	if createP1.Code != http.StatusCreated {
		t.Fatalf("create p1 status=%d body=%s", createP1.Code, createP1.Body.String())
	}
	createP2 := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"P2"}`), map[string]string{"Content-Type": "application/json"})
	if createP2.Code != http.StatusCreated {
		t.Fatalf("create p2 status=%d body=%s", createP2.Code, createP2.Body.String())
	}

	var p1 struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createP1.Body).Decode(&p1); err != nil {
		t.Fatalf("decode p1: %v", err)
	}
	var p2 struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createP2.Body).Decode(&p2); err != nil {
		t.Fatalf("decode p2: %v", err)
	}

	setP1 := doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+p1.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if setP1.Code != http.StatusOK {
		t.Fatalf("set active p1 status=%d body=%s", setP1.Code, setP1.Body.String())
	}
	createItem := doRequest(t, a, http.MethodPost, "/api/items", strings.NewReader(`{"part_number":"P1-ONLY-001","title":"P1 Item","brand":"AFX","category":"Slot"}`), map[string]string{"Content-Type": "application/json"})
	if createItem.Code != http.StatusCreated {
		t.Fatalf("create item status=%d body=%s", createItem.Code, createItem.Body.String())
	}

	setP2 := doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+p2.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if setP2.Code != http.StatusOK {
		t.Fatalf("set active p2 status=%d body=%s", setP2.Code, setP2.Body.String())
	}
	listP2 := doRequest(t, a, http.MethodGet, "/api/items", nil, nil)
	if listP2.Code != http.StatusOK {
		t.Fatalf("list p2 items status=%d body=%s", listP2.Code, listP2.Body.String())
	}
	var payload struct {
		Items []struct {
			PartNumber string `json:"part_number"`
		} `json:"items"`
	}
	if err := json.NewDecoder(listP2.Body).Decode(&payload); err != nil {
		t.Fatalf("decode p2 items: %v", err)
	}
	if len(payload.Items) != 0 {
		t.Fatalf("expected no items for p2 profile, got %d (first=%s)", len(payload.Items), payload.Items[0].PartNumber)
	}
}

func TestSearchDataIsIsolatedByActiveProfile(t *testing.T) {
	a := newTestApp(t)

	createP1 := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"P1"}`), map[string]string{"Content-Type": "application/json"})
	createP2 := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"P2"}`), map[string]string{"Content-Type": "application/json"})
	if createP1.Code != http.StatusCreated || createP2.Code != http.StatusCreated {
		t.Fatalf("create profiles failed p1=%d p2=%d", createP1.Code, createP2.Code)
	}

	var p1 struct {
		ID string `json:"id"`
	}
	var p2 struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createP1.Body).Decode(&p1); err != nil {
		t.Fatalf("decode p1: %v", err)
	}
	if err := json.NewDecoder(createP2.Body).Decode(&p2); err != nil {
		t.Fatalf("decode p2: %v", err)
	}

	_ = doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+p1.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	createItem := doRequest(t, a, http.MethodPost, "/api/items", strings.NewReader(`{"part_number":"SEARCH-P1-001","title":"Search Item P1","brand":"AFX","category":"Slot"}`), map[string]string{"Content-Type": "application/json"})
	if createItem.Code != http.StatusCreated {
		t.Fatalf("create item status=%d body=%s", createItem.Code, createItem.Body.String())
	}
	_ = doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+p2.ID+`"}`), map[string]string{"Content-Type": "application/json"})

	searchP2 := doRequest(t, a, http.MethodGet, "/api/search/items?brand=AFX", nil, nil)
	if searchP2.Code != http.StatusOK {
		t.Fatalf("search p2 status=%d body=%s", searchP2.Code, searchP2.Body.String())
	}
	var payload struct {
		Items []struct {
			PartNumber string `json:"part_number"`
		} `json:"items"`
	}
	if err := json.NewDecoder(searchP2.Body).Decode(&payload); err != nil {
		t.Fatalf("decode search payload: %v", err)
	}
	if len(payload.Items) != 0 {
		t.Fatalf("expected no search items for p2 profile, got %d (first=%s)", len(payload.Items), payload.Items[0].PartNumber)
	}
}

func TestWishlistDataIsIsolatedByActiveProfile(t *testing.T) {
	a := newTestApp(t)

	createP1 := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"P1"}`), map[string]string{"Content-Type": "application/json"})
	createP2 := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"P2"}`), map[string]string{"Content-Type": "application/json"})
	if createP1.Code != http.StatusCreated || createP2.Code != http.StatusCreated {
		t.Fatalf("create profiles failed p1=%d p2=%d", createP1.Code, createP2.Code)
	}
	var p1 struct {
		ID string `json:"id"`
	}
	var p2 struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createP1.Body).Decode(&p1); err != nil {
		t.Fatalf("decode p1: %v", err)
	}
	if err := json.NewDecoder(createP2.Body).Decode(&p2); err != nil {
		t.Fatalf("decode p2: %v", err)
	}

	_ = doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+p1.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	createItem := doRequest(t, a, http.MethodPost, "/api/items", strings.NewReader(`{"part_number":"WISH-P1-001","title":"Wish Item P1","brand":"AFX","category":"Slot"}`), map[string]string{"Content-Type": "application/json"})
	if createItem.Code != http.StatusCreated {
		t.Fatalf("create item status=%d body=%s", createItem.Code, createItem.Body.String())
	}
	var item struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createItem.Body).Decode(&item); err != nil {
		t.Fatalf("decode created item: %v", err)
	}
	createWishlist := doRequest(t, a, http.MethodPost, "/api/wishlist", strings.NewReader(`{"item_id":"`+item.ID+`","target_price":20}`), map[string]string{"Content-Type": "application/json"})
	if createWishlist.Code != http.StatusCreated {
		t.Fatalf("create wishlist status=%d body=%s", createWishlist.Code, createWishlist.Body.String())
	}

	_ = doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+p2.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	listP2 := doRequest(t, a, http.MethodGet, "/api/wishlist", nil, nil)
	if listP2.Code != http.StatusOK {
		t.Fatalf("list wishlist p2 status=%d body=%s", listP2.Code, listP2.Body.String())
	}
	var payload struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.NewDecoder(listP2.Body).Decode(&payload); err != nil {
		t.Fatalf("decode wishlist payload: %v", err)
	}
	if payload.Items == nil {
		t.Fatal("expected empty wishlist items array for isolated profile, got null")
	}
	if len(payload.Items) != 0 {
		t.Fatalf("expected no wishlist items for p2 profile, got %d", len(payload.Items))
	}
}

func TestScannerQuerySetsAreIsolatedByActiveProfile(t *testing.T) {
	a := newTestApp(t)

	createP1 := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"P1"}`), map[string]string{"Content-Type": "application/json"})
	createP2 := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"P2"}`), map[string]string{"Content-Type": "application/json"})
	if createP1.Code != http.StatusCreated || createP2.Code != http.StatusCreated {
		t.Fatalf("create profiles failed p1=%d p2=%d", createP1.Code, createP2.Code)
	}
	var p1 struct {
		ID string `json:"id"`
	}
	var p2 struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createP1.Body).Decode(&p1); err != nil {
		t.Fatalf("decode p1: %v", err)
	}
	if err := json.NewDecoder(createP2.Body).Decode(&p2); err != nil {
		t.Fatalf("decode p2: %v", err)
	}

	_ = doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+p1.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	createQS := doRequest(t, a, http.MethodPost, "/api/scanner/query-sets", strings.NewReader(`{"name":"P1 Set","keywords":["afx"],"enabled":true}`), map[string]string{"Content-Type": "application/json"})
	if createQS.Code != http.StatusCreated {
		t.Fatalf("create query set status=%d body=%s", createQS.Code, createQS.Body.String())
	}

	_ = doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+p2.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	listP2 := doRequest(t, a, http.MethodGet, "/api/scanner/query-sets", nil, nil)
	if listP2.Code != http.StatusOK {
		t.Fatalf("list query sets p2 status=%d body=%s", listP2.Code, listP2.Body.String())
	}
	var payload struct {
		QuerySets []map[string]any `json:"query_sets"`
	}
	if err := json.NewDecoder(listP2.Body).Decode(&payload); err != nil {
		t.Fatalf("decode query sets payload: %v", err)
	}
	if len(payload.QuerySets) != 0 {
		t.Fatalf("expected no query sets for p2 profile, got %d", len(payload.QuerySets))
	}
}

func TestPricingEndpointsRejectItemsFromOtherProfiles(t *testing.T) {
	a := newTestApp(t)

	createP1 := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"P1"}`), map[string]string{"Content-Type": "application/json"})
	createP2 := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"P2"}`), map[string]string{"Content-Type": "application/json"})
	if createP1.Code != http.StatusCreated || createP2.Code != http.StatusCreated {
		t.Fatalf("create profiles failed p1=%d p2=%d", createP1.Code, createP2.Code)
	}
	var p1 struct {
		ID string `json:"id"`
	}
	var p2 struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createP1.Body).Decode(&p1); err != nil {
		t.Fatalf("decode p1: %v", err)
	}
	if err := json.NewDecoder(createP2.Body).Decode(&p2); err != nil {
		t.Fatalf("decode p2: %v", err)
	}

	_ = doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+p1.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	createItem := doRequest(t, a, http.MethodPost, "/api/items", strings.NewReader(`{"part_number":"PRICE-P1-001","title":"Price Item P1","brand":"AFX","category":"Slot"}`), map[string]string{"Content-Type": "application/json"})
	if createItem.Code != http.StatusCreated {
		t.Fatalf("create item status=%d body=%s", createItem.Code, createItem.Body.String())
	}
	var item struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createItem.Body).Decode(&item); err != nil {
		t.Fatalf("decode created item: %v", err)
	}

	_ = doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+p2.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	track := doRequest(t, a, http.MethodPost, "/api/pricing/track", strings.NewReader(`{"item_id":"`+item.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if track.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request for cross-profile track, got %d body=%s", track.Code, track.Body.String())
	}
	history := doRequest(t, a, http.MethodGet, "/api/pricing/history?item_id="+item.ID, nil, nil)
	if history.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request for cross-profile history, got %d body=%s", history.Code, history.Body.String())
	}
}
