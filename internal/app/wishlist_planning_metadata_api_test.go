package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestWishlistEndpointPersistsManualWatchStatusForActiveProfile(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Wishlist Metadata P1"}`), map[string]string{"Content-Type": "application/json"})
	if createProfile.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", createProfile.Code, createProfile.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createProfile.Body).Decode(&profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	if profile.ID == "" {
		t.Fatal("expected profile id")
	}

	setActive := doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+profile.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if setActive.Code != http.StatusOK {
		t.Fatalf("set active profile status=%d body=%s", setActive.Code, setActive.Body.String())
	}

	createItem := doRequest(t, a, http.MethodPost, "/api/items", strings.NewReader(`{"part_number":"WISH-META-001","title":"Wishlist Metadata Item","brand":"AFX","category":"Slot"}`), map[string]string{"Content-Type": "application/json"})
	if createItem.Code != http.StatusCreated {
		t.Fatalf("create item status=%d body=%s", createItem.Code, createItem.Body.String())
	}
	var item struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createItem.Body).Decode(&item); err != nil {
		t.Fatalf("decode item: %v", err)
	}
	if item.ID == "" {
		t.Fatal("expected item id")
	}

	createWish := doRequest(t, a, http.MethodPost, "/api/wishlist", strings.NewReader(`{"item_id":"`+item.ID+`","target_price":42,"priority":"high","notes":"manual watch state","below_target_now":true,"highlight_hit":true,"owned":true,"price_paid":37.5,"purchase_url":"https://example.test/receipt","purchase_date":"2026-04-27","purchase_condition":"boxed","quantity":2,"needed_quantity":5}`), map[string]string{"Content-Type": "application/json"})
	if createWish.Code != http.StatusCreated {
		t.Fatalf("create wishlist status=%d body=%s", createWish.Code, createWish.Body.String())
	}
	var created struct {
		ID             string `json:"id"`
		BelowTargetNow bool   `json:"below_target_now"`
		HighlightHit   bool   `json:"highlight_hit"`
		Owned          bool   `json:"owned"`
	}
	if err := json.NewDecoder(createWish.Body).Decode(&created); err != nil {
		t.Fatalf("decode created wishlist: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected wishlist id")
	}
	if !created.BelowTargetNow {
		t.Fatalf("expected created wishlist to persist below_target_now=true, got %+v", created)
	}
	if !created.HighlightHit {
		t.Fatalf("expected created wishlist to persist highlight_hit=true, got %+v", created)
	}
	if !created.Owned {
		t.Fatalf("expected created wishlist to persist owned=true, got %+v", created)
	}

	updateWish := doRequest(t, a, http.MethodPut, "/api/wishlist", strings.NewReader(`{"id":"`+created.ID+`","item_id":"`+item.ID+`","target_price":42,"priority":"medium","notes":"updated manual watch state","below_target_now":false,"highlight_hit":false,"owned":true,"price_paid":35.25,"purchase_url":"https://example.test/updated-receipt","purchase_date":"2026-04-28","purchase_condition":"loose","quantity":3,"needed_quantity":4}`), map[string]string{"Content-Type": "application/json"})
	if updateWish.Code != http.StatusOK {
		t.Fatalf("update wishlist status=%d body=%s", updateWish.Code, updateWish.Body.String())
	}

	listWish := doRequest(t, a, http.MethodGet, "/api/wishlist", nil, nil)
	if listWish.Code != http.StatusOK {
		t.Fatalf("list wishlist status=%d body=%s", listWish.Code, listWish.Body.String())
	}
	var listPayload struct {
		Items []struct {
			ID                string  `json:"id"`
			BelowTargetNow    bool    `json:"below_target_now"`
			HighlightHit      bool    `json:"highlight_hit"`
			Priority          string  `json:"priority"`
			Notes             string  `json:"notes"`
			Owned             bool    `json:"owned"`
			PricePaid         float64 `json:"price_paid"`
			PurchaseURL       string  `json:"purchase_url"`
			PurchaseDate      string  `json:"purchase_date"`
			PurchaseCondition string  `json:"purchase_condition"`
			Quantity          int     `json:"quantity"`
			NeededQuantity    int     `json:"needed_quantity"`
		} `json:"items"`
	}
	if err := json.NewDecoder(listWish.Body).Decode(&listPayload); err != nil {
		t.Fatalf("decode wishlist payload: %v", err)
	}
	if len(listPayload.Items) != 1 {
		t.Fatalf("expected one wishlist entry, got %+v", listPayload.Items)
	}
	if listPayload.Items[0].BelowTargetNow {
		t.Fatalf("expected updated wishlist below_target_now=false after round-trip, got %+v", listPayload.Items[0])
	}
	if listPayload.Items[0].HighlightHit {
		t.Fatalf("expected updated wishlist highlight_hit=false after round-trip, got %+v", listPayload.Items[0])
	}
	if listPayload.Items[0].Priority != "medium" {
		t.Fatalf("expected updated wishlist priority=medium, got %+v", listPayload.Items[0])
	}
	if listPayload.Items[0].Notes != "updated manual watch state" {
		t.Fatalf("expected updated wishlist notes to persist, got %+v", listPayload.Items[0])
	}
	if !listPayload.Items[0].Owned || listPayload.Items[0].PricePaid != 35.25 || listPayload.Items[0].PurchaseURL != "https://example.test/updated-receipt" || listPayload.Items[0].PurchaseDate != "2026-04-28" || listPayload.Items[0].PurchaseCondition != "loose" || listPayload.Items[0].Quantity != 3 || listPayload.Items[0].NeededQuantity != 4 {
		t.Fatalf("expected ownership metadata round-trip after update, got %+v", listPayload.Items[0])
	}

	partialUpdate := doRequest(t, a, http.MethodPut, "/api/wishlist", strings.NewReader(`{"id":"`+created.ID+`","priority":"high"}`), map[string]string{"Content-Type": "application/json"})
	if partialUpdate.Code != http.StatusOK {
		t.Fatalf("partial update wishlist status=%d body=%s", partialUpdate.Code, partialUpdate.Body.String())
	}

	listAfterPartialUpdate := doRequest(t, a, http.MethodGet, "/api/wishlist", nil, nil)
	if listAfterPartialUpdate.Code != http.StatusOK {
		t.Fatalf("list wishlist after partial update status=%d body=%s", listAfterPartialUpdate.Code, listAfterPartialUpdate.Body.String())
	}
	var partialPayload struct {
		Items []struct {
			ID                string  `json:"id"`
			TargetPrice       float64 `json:"target_price"`
			BelowTargetNow    bool    `json:"below_target_now"`
			HighlightHit      bool    `json:"highlight_hit"`
			Priority          string  `json:"priority"`
			Notes             string  `json:"notes"`
			Owned             bool    `json:"owned"`
			PricePaid         float64 `json:"price_paid"`
			PurchaseURL       string  `json:"purchase_url"`
			PurchaseDate      string  `json:"purchase_date"`
			PurchaseCondition string  `json:"purchase_condition"`
			Quantity          int     `json:"quantity"`
			NeededQuantity    int     `json:"needed_quantity"`
		} `json:"items"`
	}
	if err := json.NewDecoder(listAfterPartialUpdate.Body).Decode(&partialPayload); err != nil {
		t.Fatalf("decode partial wishlist payload: %v", err)
	}
	if len(partialPayload.Items) != 1 {
		t.Fatalf("expected one wishlist entry after partial update, got %+v", partialPayload.Items)
	}
	if partialPayload.Items[0].Priority != "high" {
		t.Fatalf("expected partial update to set priority=high, got %+v", partialPayload.Items[0])
	}
	if partialPayload.Items[0].TargetPrice != 42 {
		t.Fatalf("expected partial update to preserve target_price=42, got %+v", partialPayload.Items[0])
	}
	if partialPayload.Items[0].Notes != "updated manual watch state" {
		t.Fatalf("expected partial update to preserve notes, got %+v", partialPayload.Items[0])
	}
	if partialPayload.Items[0].BelowTargetNow || partialPayload.Items[0].HighlightHit {
		t.Fatalf("expected partial update to preserve false watch flags, got %+v", partialPayload.Items[0])
	}
	if !partialPayload.Items[0].Owned || partialPayload.Items[0].PricePaid != 35.25 || partialPayload.Items[0].PurchaseURL != "https://example.test/updated-receipt" || partialPayload.Items[0].PurchaseDate != "2026-04-28" || partialPayload.Items[0].PurchaseCondition != "loose" || partialPayload.Items[0].Quantity != 3 || partialPayload.Items[0].NeededQuantity != 4 {
		t.Fatalf("expected partial update to preserve ownership metadata, got %+v", partialPayload.Items[0])
	}
}
