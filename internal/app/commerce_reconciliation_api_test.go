package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestCommerceLifecyclePurchaseCreatesExpectedArrival(t *testing.T) {
	t.Parallel()

	a, profileID := newCommerceProfileApp(t)
	if _, err := a.db.Exec(`INSERT INTO canonical_items(id, profile_id, brand, category, part_number, title) VALUES ('i1', ?, 'AFX','Slot','P-442','Commerce Item')`, profileID); err != nil {
		t.Fatalf("seed item: %v", err)
	}

	create := doRequest(t, a, http.MethodPost, "/api/commerce/lifecycle", strings.NewReader(`{"item_id":"i1","state":"purchase","source":"ebay","external_ref":"order-442","quantity":2,"amount":55.5,"currency":"aud","notes":"bought two"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create lifecycle status=%d body=%s", create.Code, create.Body.String())
	}
	var created struct {
		Entry struct {
			ID                string `json:"id"`
			State             string `json:"state"`
			ExpectedArrivalID string `json:"expected_arrival_id"`
		} `json:"entry"`
		ExpectedArrival struct {
			ID               string  `json:"id"`
			LifecycleEntryID string  `json:"lifecycle_entry_id"`
			Status           string  `json:"status"`
			Quantity         int     `json:"quantity"`
			Amount           float64 `json:"amount"`
		} `json:"expected_arrival"`
	}
	if err := json.NewDecoder(create.Body).Decode(&created); err != nil {
		t.Fatalf("decode create lifecycle: %v", err)
	}
	if created.Entry.ID == "" || created.ExpectedArrival.ID == "" {
		t.Fatalf("expected lifecycle and arrival ids, got %+v", created)
	}
	if created.Entry.State != "purchase" {
		t.Fatalf("expected purchase state, got %+v", created.Entry)
	}
	if created.ExpectedArrival.Status != "expected" || created.ExpectedArrival.Quantity != 2 {
		t.Fatalf("expected auto-created arrival, got %+v", created.ExpectedArrival)
	}
	if created.ExpectedArrival.LifecycleEntryID != created.Entry.ID {
		t.Fatalf("expected arrival linked to entry, got entry=%s arrival=%+v", created.Entry.ID, created.ExpectedArrival)
	}

	listLifecycle := doRequest(t, a, http.MethodGet, "/api/commerce/lifecycle?item_id=i1", nil, nil)
	if listLifecycle.Code != http.StatusOK {
		t.Fatalf("list lifecycle status=%d body=%s", listLifecycle.Code, listLifecycle.Body.String())
	}
	var lifecyclePayload struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.NewDecoder(listLifecycle.Body).Decode(&lifecyclePayload); err != nil {
		t.Fatalf("decode lifecycle list: %v", err)
	}
	if len(lifecyclePayload.Items) != 1 {
		t.Fatalf("expected one lifecycle entry, got %d", len(lifecyclePayload.Items))
	}

	listArrivals := doRequest(t, a, http.MethodGet, "/api/commerce/arrivals?item_id=i1&status=expected", nil, nil)
	if listArrivals.Code != http.StatusOK {
		t.Fatalf("list arrivals status=%d body=%s", listArrivals.Code, listArrivals.Body.String())
	}
	var arrivalsPayload struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.NewDecoder(listArrivals.Body).Decode(&arrivalsPayload); err != nil {
		t.Fatalf("decode arrivals list: %v", err)
	}
	if len(arrivalsPayload.Items) != 1 {
		t.Fatalf("expected one expected arrival, got %d", len(arrivalsPayload.Items))
	}

	update := doRequest(t, a, http.MethodPut, "/api/commerce/arrivals", strings.NewReader(`{"id":"`+created.ExpectedArrival.ID+`","status":"reconciled","delivered_on":"2026-04-03","reconciled_instance_id":"inst-1","notes":"checked in"}`), map[string]string{"Content-Type": "application/json"})
	if update.Code != http.StatusOK {
		t.Fatalf("update arrival status=%d body=%s", update.Code, update.Body.String())
	}
	listReconciled := doRequest(t, a, http.MethodGet, "/api/commerce/arrivals?item_id=i1&status=reconciled", nil, nil)
	if listReconciled.Code != http.StatusOK {
		t.Fatalf("list reconciled arrivals status=%d body=%s", listReconciled.Code, listReconciled.Body.String())
	}
	var reconciledPayload struct {
		Items []struct {
			Status               string `json:"status"`
			DeliveredOn          string `json:"delivered_on"`
			ReconciledInstanceID string `json:"reconciled_instance_id"`
		} `json:"items"`
	}
	if err := json.NewDecoder(listReconciled.Body).Decode(&reconciledPayload); err != nil {
		t.Fatalf("decode reconciled arrivals: %v", err)
	}
	if len(reconciledPayload.Items) != 1 {
		t.Fatalf("expected one reconciled arrival, got %d", len(reconciledPayload.Items))
	}
	if reconciledPayload.Items[0].Status != "reconciled" || reconciledPayload.Items[0].DeliveredOn != "2026-04-03" || reconciledPayload.Items[0].ReconciledInstanceID != "inst-1" {
		t.Fatalf("unexpected reconciled payload %+v", reconciledPayload.Items[0])
	}
}

func TestCommerceLifecycleNonPurchasePersistsWithoutExpectedArrival(t *testing.T) {
	t.Parallel()

	a, profileID := newCommerceProfileApp(t)
	if _, err := a.db.Exec(`INSERT INTO canonical_items(id, profile_id, brand, category, part_number, title) VALUES ('i2', ?, 'Hot Wheels','Diecast','P-442-B','Wishlist Item')`, profileID); err != nil {
		t.Fatalf("seed item: %v", err)
	}

	create := doRequest(t, a, http.MethodPost, "/api/commerce/lifecycle", strings.NewReader(`{"item_id":"i2","state":"wishlist","source":"manual","external_ref":"wish-442","quantity":1,"amount":15,"currency":"aud","notes":"want one"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create lifecycle status=%d body=%s", create.Code, create.Body.String())
	}
	var created struct {
		Entry struct {
			ID                string `json:"id"`
			State             string `json:"state"`
			ExpectedArrivalID string `json:"expected_arrival_id"`
		} `json:"entry"`
	}
	if err := json.NewDecoder(create.Body).Decode(&created); err != nil {
		t.Fatalf("decode create lifecycle: %v", err)
	}
	if created.Entry.ID == "" {
		t.Fatalf("expected lifecycle id, got %+v", created)
	}
	if created.Entry.State != "wishlist" {
		t.Fatalf("expected wishlist state, got %+v", created.Entry)
	}
	if created.Entry.ExpectedArrivalID != "" {
		t.Fatalf("expected no arrival id for non-purchase lifecycle, got %+v", created.Entry)
	}

	listLifecycle := doRequest(t, a, http.MethodGet, "/api/commerce/lifecycle?item_id=i2", nil, nil)
	if listLifecycle.Code != http.StatusOK {
		t.Fatalf("list lifecycle status=%d body=%s", listLifecycle.Code, listLifecycle.Body.String())
	}
	var lifecyclePayload struct {
		Items []struct {
			ID                string `json:"id"`
			State             string `json:"state"`
			ExpectedArrivalID string `json:"expected_arrival_id"`
		} `json:"items"`
	}
	if err := json.NewDecoder(listLifecycle.Body).Decode(&lifecyclePayload); err != nil {
		t.Fatalf("decode lifecycle list: %v", err)
	}
	if len(lifecyclePayload.Items) != 1 {
		t.Fatalf("expected one lifecycle entry, got %d", len(lifecyclePayload.Items))
	}
	if lifecyclePayload.Items[0].State != "wishlist" || lifecyclePayload.Items[0].ExpectedArrivalID != "" {
		t.Fatalf("unexpected lifecycle payload %+v", lifecyclePayload.Items[0])
	}

	listArrivals := doRequest(t, a, http.MethodGet, "/api/commerce/arrivals?item_id=i2", nil, nil)
	if listArrivals.Code != http.StatusOK {
		t.Fatalf("list arrivals status=%d body=%s", listArrivals.Code, listArrivals.Body.String())
	}
	var arrivalsPayload struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.NewDecoder(listArrivals.Body).Decode(&arrivalsPayload); err != nil {
		t.Fatalf("decode arrivals list: %v", err)
	}
	if len(arrivalsPayload.Items) != 0 {
		t.Fatalf("expected no arrivals for non-purchase lifecycle, got %d", len(arrivalsPayload.Items))
	}
}

func TestCommerceArrivalLifecycleStatesPersistAndFilter(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name                 string
		itemID               string
		status               string
		deliveredOn          string
		reconciledInstanceID string
		notes                string
	}{
		{
			name:                 "delivered",
			itemID:               "arrival-state-delivered-item",
			status:               "delivered",
			deliveredOn:          "2026-06-14",
			reconciledInstanceID: "instance-delivered-001",
			notes:                "package arrived at cabinet review",
		},
		{
			name:   "cancelled",
			itemID: "arrival-state-cancelled-item",
			status: "cancelled",
			notes:  "seller cancelled before forwarding",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			a, profileID := newCommerceProfileApp(t)
			if _, err := a.db.Exec(`INSERT INTO canonical_items(id, profile_id, brand, category, part_number, title) VALUES (?, ?, 'AFX','Slot', ?, ?)`, tc.itemID, profileID, strings.ToUpper(tc.status)+"-1", "Purchase "+tc.status); err != nil {
				t.Fatalf("seed item: %v", err)
			}

			create := doRequest(t, a, http.MethodPost, "/api/commerce/lifecycle", strings.NewReader(`{"item_id":"`+tc.itemID+`","state":"purchase","source":"ebay","external_ref":"order-`+tc.status+`","quantity":1,"amount":12.5,"currency":"aud","notes":"state transition coverage"}`), map[string]string{"Content-Type": "application/json"})
			if create.Code != http.StatusCreated {
				t.Fatalf("create lifecycle status=%d body=%s", create.Code, create.Body.String())
			}
			var created struct {
				Entry struct {
					ID string `json:"id"`
				} `json:"entry"`
				ExpectedArrival struct {
					ID               string `json:"id"`
					LifecycleEntryID string `json:"lifecycle_entry_id"`
					Status           string `json:"status"`
				} `json:"expected_arrival"`
			}
			if err := json.NewDecoder(create.Body).Decode(&created); err != nil {
				t.Fatalf("decode create lifecycle: %v", err)
			}
			if created.ExpectedArrival.ID == "" || created.ExpectedArrival.Status != "expected" || created.ExpectedArrival.LifecycleEntryID != created.Entry.ID {
				t.Fatalf("expected purchase to create expected arrival, got %+v", created)
			}

			updateBody := `{"id":"` + created.ExpectedArrival.ID + `","status":"` + tc.status + `","delivered_on":"` + tc.deliveredOn + `","reconciled_instance_id":"` + tc.reconciledInstanceID + `","notes":"` + tc.notes + `"}`
			update := doRequest(t, a, http.MethodPut, "/api/commerce/arrivals", strings.NewReader(updateBody), map[string]string{"Content-Type": "application/json"})
			if update.Code != http.StatusOK {
				t.Fatalf("update arrival status=%d body=%s", update.Code, update.Body.String())
			}

			list := doRequest(t, a, http.MethodGet, "/api/commerce/arrivals?item_id="+tc.itemID+"&status="+tc.status, nil, nil)
			if list.Code != http.StatusOK {
				t.Fatalf("list %s arrivals status=%d body=%s", tc.status, list.Code, list.Body.String())
			}
			var payload struct {
				Items []struct {
					ID                   string `json:"id"`
					Status               string `json:"status"`
					DeliveredOn          string `json:"delivered_on"`
					ReconciledInstanceID string `json:"reconciled_instance_id"`
					Notes                string `json:"notes"`
				} `json:"items"`
			}
			if err := json.NewDecoder(list.Body).Decode(&payload); err != nil {
				t.Fatalf("decode %s arrivals list: %v", tc.status, err)
			}
			if len(payload.Items) != 1 {
				t.Fatalf("expected one %s arrival, got %+v", tc.status, payload.Items)
			}
			got := payload.Items[0]
			if got.ID != created.ExpectedArrival.ID || got.Status != tc.status || got.DeliveredOn != tc.deliveredOn || got.ReconciledInstanceID != tc.reconciledInstanceID || got.Notes != tc.notes {
				t.Fatalf("unexpected %s arrival payload %+v", tc.status, got)
			}
		})
	}
}

func newCommerceProfileApp(t *testing.T) (*App, string) {
	t.Helper()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Commerce"}`), map[string]string{"Content-Type": "application/json"})
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
	return a, profile.ID
}
