package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestWishlistPurchasedDeliveredSyncsCommerceAndInventoryState(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Wishlist Purchase Delivery"}`), map[string]string{"Content-Type": "application/json"})
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

	createItem := doRequest(t, a, http.MethodPost, "/api/items", strings.NewReader(`{"part_number":"WISH-PURCHASE-001","title":"Wishlist Purchase Delivery Item","brand":"AFX","category":"Slot Cars","status":"wishlist"}`), map[string]string{"Content-Type": "application/json"})
	if createItem.Code != http.StatusCreated {
		t.Fatalf("create item status=%d body=%s", createItem.Code, createItem.Body.String())
	}
	var item struct {
		ID       string `json:"id"`
		Category string `json:"category"`
	}
	if err := json.NewDecoder(createItem.Body).Decode(&item); err != nil {
		t.Fatalf("decode item: %v", err)
	}
	if item.ID == "" || item.Category != "Slot Cars" {
		t.Fatalf("expected created wishlist item with category, got %+v", item)
	}

	createWish := doRequest(t, a, http.MethodPost, "/api/wishlist", strings.NewReader(`{"item_id":"`+item.ID+`","target_price":42,"priority":"high","notes":"wishlist provenance","owned":true,"price_paid":37.5,"purchase_url":"https://example.test/order","purchase_date":"2026-04-27","purchase_condition":"sealed","quantity":2,"needed_quantity":2}`), map[string]string{"Content-Type": "application/json"})
	if createWish.Code != http.StatusCreated {
		t.Fatalf("create purchased wishlist status=%d body=%s", createWish.Code, createWish.Body.String())
	}
	var createdWish struct {
		ID        string `json:"id"`
		Owned     bool   `json:"owned"`
		Delivered bool   `json:"delivered"`
	}
	if err := json.NewDecoder(createWish.Body).Decode(&createdWish); err != nil {
		t.Fatalf("decode purchased wishlist: %v", err)
	}
	if createdWish.ID == "" || !createdWish.Owned || createdWish.Delivered {
		t.Fatalf("expected purchased but undelivered wishlist response, got %+v", createdWish)
	}

	listLifecycle := doRequest(t, a, http.MethodGet, "/api/commerce/lifecycle?item_id="+item.ID, nil, nil)
	if listLifecycle.Code != http.StatusOK {
		t.Fatalf("list purchase lifecycle status=%d body=%s", listLifecycle.Code, listLifecycle.Body.String())
	}
	var lifecyclePayload struct {
		Items []struct {
			State             string  `json:"state"`
			Source            string  `json:"source"`
			ExternalRef       string  `json:"external_ref"`
			Quantity          int     `json:"quantity"`
			Amount            float64 `json:"amount"`
			ExpectedArrivalID string  `json:"expected_arrival_id"`
		} `json:"items"`
	}
	if err := json.NewDecoder(listLifecycle.Body).Decode(&lifecyclePayload); err != nil {
		t.Fatalf("decode lifecycle list: %v", err)
	}
	if len(lifecyclePayload.Items) != 1 {
		t.Fatalf("expected one wishlist purchase lifecycle entry, got %+v", lifecyclePayload.Items)
	}
	if lifecyclePayload.Items[0].State != "purchase" || lifecyclePayload.Items[0].Source != "wishlist" || lifecyclePayload.Items[0].ExternalRef != createdWish.ID || lifecyclePayload.Items[0].Quantity != 2 || lifecyclePayload.Items[0].Amount != 37.5 || lifecyclePayload.Items[0].ExpectedArrivalID == "" {
		t.Fatalf("unexpected purchase lifecycle entry: %+v", lifecyclePayload.Items[0])
	}

	deliverWish := doRequest(t, a, http.MethodPut, "/api/wishlist", strings.NewReader(`{"id":"`+createdWish.ID+`","delivered":true,"purchase_condition":"sealed","quantity":2}`), map[string]string{"Content-Type": "application/json"})
	if deliverWish.Code != http.StatusOK {
		t.Fatalf("deliver wishlist status=%d body=%s", deliverWish.Code, deliverWish.Body.String())
	}

	listWish := doRequest(t, a, http.MethodGet, "/api/wishlist", nil, nil)
	if listWish.Code != http.StatusOK {
		t.Fatalf("list wishlist status=%d body=%s", listWish.Code, listWish.Body.String())
	}
	var wishPayload struct {
		Items []struct {
			Owned     bool `json:"owned"`
			Delivered bool `json:"delivered"`
		} `json:"items"`
	}
	if err := json.NewDecoder(listWish.Body).Decode(&wishPayload); err != nil {
		t.Fatalf("decode wishlist after delivery: %v", err)
	}
	if len(wishPayload.Items) != 1 || !wishPayload.Items[0].Owned || !wishPayload.Items[0].Delivered {
		t.Fatalf("expected delivery to imply purchased and persist delivered state, got %+v", wishPayload.Items)
	}

	listDeliveredArrivals := doRequest(t, a, http.MethodGet, "/api/commerce/arrivals?item_id="+item.ID+"&status=delivered", nil, nil)
	if listDeliveredArrivals.Code != http.StatusOK {
		t.Fatalf("list delivered arrivals status=%d body=%s", listDeliveredArrivals.Code, listDeliveredArrivals.Body.String())
	}
	var arrivalsPayload struct {
		Items []struct {
			Status               string `json:"status"`
			ReconciledInstanceID string `json:"reconciled_instance_id"`
		} `json:"items"`
	}
	if err := json.NewDecoder(listDeliveredArrivals.Body).Decode(&arrivalsPayload); err != nil {
		t.Fatalf("decode delivered arrivals: %v", err)
	}
	if len(arrivalsPayload.Items) != 1 || arrivalsPayload.Items[0].Status != "delivered" || arrivalsPayload.Items[0].ReconciledInstanceID == "" {
		t.Fatalf("expected delivered arrival linked to inventory instance, got %+v", arrivalsPayload.Items)
	}

	var itemStatus, itemCategory string
	if err := a.db.QueryRow(`SELECT status, category FROM canonical_items WHERE id = ? AND profile_id = ?`, item.ID, profile.ID).Scan(&itemStatus, &itemCategory); err != nil {
		t.Fatalf("load delivered item: %v", err)
	}
	if itemStatus != "active" || itemCategory != "Slot Cars" {
		t.Fatalf("expected delivered item active with category carried through, got status=%q category=%q", itemStatus, itemCategory)
	}

	var instanceCount int
	if err := a.db.QueryRow(`SELECT COUNT(1) FROM instances WHERE item_id = ? AND status = 'sealed' AND quantity = 2 AND acquisition_price = 37.5 AND acquisition_date = '2026-04-27'`, item.ID).Scan(&instanceCount); err != nil {
		t.Fatalf("count delivered inventory instances: %v", err)
	}
	if instanceCount != 1 {
		t.Fatalf("expected one delivered inventory instance, got %d", instanceCount)
	}
}
