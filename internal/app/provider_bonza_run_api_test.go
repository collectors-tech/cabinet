package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBonzaRunAggregatesPagesAndEnrichesWatchedStock(t *testing.T) {
	t.Parallel()

	requests := map[string]int{}
	bonza := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests[r.URL.Path]++
		switch {
		case strings.HasPrefix(r.URL.Path, "/wp-json/wc/store/v1/products"):
			page := r.URL.Query().Get("page")
			w.Header().Set("Content-Type", "application/json")
			if page == "2" {
				_, _ = w.Write([]byte(`[
{"id":1001,"name":"AFX Camaro Variant","permalink":"` + r.Host + `/product/1001","is_in_stock":true,"low_stock_remaining":null},
{"id":1003,"name":"AFX Mustang","permalink":"` + r.Host + `/product/1003","is_in_stock":false,"low_stock_remaining":null}
]`))
				return
			}
			w.Header().Set("X-WP-TotalPages", "2")
			_, _ = w.Write([]byte(`[
{"id":1001,"name":"AFX Camaro","permalink":"` + r.Host + `/product/1001","is_in_stock":true,"low_stock_remaining":null},
{"id":1002,"name":"AFX Thunderjet","permalink":"` + r.Host + `/product/1002","is_in_stock":null,"low_stock_remaining":null}
]`))
		case strings.HasPrefix(r.URL.Path, "/product/1002"):
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(`<html><body><span>Only 2 left in stock</span></body></html>`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer bonza.Close()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"BonzaProfile"}`), map[string]string{"Content-Type": "application/json"})
	if createProfile.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", createProfile.Code, createProfile.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createProfile.Body).Decode(&profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	activate := doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+profile.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if activate.Code != http.StatusOK {
		t.Fatalf("activate profile status=%d body=%s", activate.Code, activate.Body.String())
	}

	settingsBody := fmt.Sprintf(`{"settings":{"integration.bonzaslotcars.base_url":"%s","integration.bonzaslotcars.items_per_page":"36"}}`, bonza.URL)
	saveSettings := doRequest(t, a, http.MethodPut, "/api/profiles/"+profile.ID+"/settings", strings.NewReader(settingsBody), map[string]string{"Content-Type": "application/json"})
	if saveSettings.Code != http.StatusOK {
		t.Fatalf("save settings status=%d body=%s", saveSettings.Code, saveSettings.Body.String())
	}

	createQuery := doRequest(t, a, http.MethodPost, "/api/scanner/query-sets", strings.NewReader(`{"name":"AFX","keywords":["AFX"],"provider_scope":["bonzaslotcars"],"enabled":true}`), map[string]string{"Content-Type": "application/json"})
	if createQuery.Code != http.StatusCreated {
		t.Fatalf("create query set status=%d body=%s", createQuery.Code, createQuery.Body.String())
	}
	var qs struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createQuery.Body).Decode(&qs); err != nil {
		t.Fatalf("decode query set: %v", err)
	}

	run := doRequest(t, a, http.MethodPost, "/api/providers/bonza/run", strings.NewReader(`{"query_set_id":"`+qs.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if run.Code != http.StatusOK {
		t.Fatalf("bonza run status=%d body=%s", run.Code, run.Body.String())
	}

	var payload struct {
		QuerySetID       string                   `json:"query_set_id"`
		PageCount        int                      `json:"page_count"`
		ObservedPageSize int                      `json:"observed_page_size"`
		Candidates       []map[string]any         `json:"candidates"`
		RunSummary       map[string]any           `json:"run_summary"`
	}
	if err := json.NewDecoder(run.Body).Decode(&payload); err != nil {
		t.Fatalf("decode bonza payload: %v", err)
	}
	if payload.QuerySetID != qs.ID {
		t.Fatalf("query_set_id mismatch: got=%s want=%s", payload.QuerySetID, qs.ID)
	}
	if payload.PageCount != 2 {
		t.Fatalf("expected page_count=2, got=%d", payload.PageCount)
	}
	if payload.ObservedPageSize == 0 {
		t.Fatalf("expected observed_page_size > 0, got=%d", payload.ObservedPageSize)
	}
	if len(payload.Candidates) != 3 {
		t.Fatalf("expected 3 deduplicated candidates, got=%d payload=%+v", len(payload.Candidates), payload.Candidates)
	}

	var enriched bool
	for _, candidate := range payload.Candidates {
		listingID, _ := candidate["listing_id"].(string)
		if listingID == "bonza-1002" {
			state, _ := candidate["stock_state"].(string)
			if state == "low_stock" {
				enriched = true
			}
		}
	}
	if !enriched {
		t.Fatalf("expected fallback detail stock enrichment for bonza-1002, payload=%+v", payload.Candidates)
	}
}
