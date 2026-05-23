package app

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"testing"
)

func TestPurchaseInboxReviewsAPIPreparesReviewRecordsWithoutInventoryMutation(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Purchase Inbox API"}`), map[string]string{"Content-Type": "application/json"})
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

	reviews := doRequest(t, a, http.MethodPost, "/api/integrations/ebay/purchase-inbox/reviews", strings.NewReader(`{
		"cards":[{
			"order_id":"20-14595-70928",
			"listing_id":"316046161178",
			"variation_id":"615514115326",
			"transaction_id":"10080684936020",
			"listing_title":"Accompanying Flute listing",
			"purchased_identity":"Accompanying Flute TWM 142 (142/167)",
			"quantity":4,
			"item_price":"AU $2.40",
			"seller_username":"seller-one",
			"order_total":"AU $8.10",
			"currency":"AUD",
			"aspects":{"Card":"Accompanying Flute TWM 142 (142/167)","Quantity":"4"}
		},{
			"order_id":"20-14595-70928",
			"listing_id":"316046161179",
			"listing_title":"Mystery purchase",
			"seller_username":"seller-one"
		}]
	}`), map[string]string{"Content-Type": "application/json"})
	if reviews.Code != http.StatusOK {
		t.Fatalf("reviews status=%d body=%s", reviews.Code, reviews.Body.String())
	}

	var payload struct {
		ProfileID string `json:"profile_id"`
		Source    string `json:"source"`
		Reviews   []struct {
			Status string `json:"status"`
			Order  struct {
				OrderID string `json:"order_id"`
			} `json:"order"`
			Items []struct {
				Status           string   `json:"status"`
				MissingFields    []string `json:"missing_fields"`
				SuggestedActions []struct {
					ID                   string `json:"id"`
					RequiresConfirmation bool   `json:"requires_confirmation"`
				} `json:"suggested_actions"`
			} `json:"items"`
		} `json:"reviews"`
	}
	if err := json.NewDecoder(reviews.Body).Decode(&payload); err != nil {
		t.Fatalf("decode reviews: %v body=%s", err, reviews.Body.String())
	}
	if payload.ProfileID != profile.ID || payload.Source != "ebay_purchase_capture" {
		t.Fatalf("expected active profile/source in response, got %+v", payload)
	}
	if len(payload.Reviews) != 1 {
		t.Fatalf("expected one grouped order review, got %+v", payload.Reviews)
	}
	if payload.Reviews[0].Order.OrderID != "20-14595-70928" || payload.Reviews[0].Status != "needs_review" {
		t.Fatalf("expected grouped order with needs_review status, got %+v", payload.Reviews[0])
	}
	if len(payload.Reviews[0].Items) != 2 {
		t.Fatalf("expected two item reviews, got %+v", payload.Reviews[0].Items)
	}
	ready := payload.Reviews[0].Items[0]
	if ready.Status != "ready_to_link_or_convert" || len(ready.SuggestedActions) != 2 {
		t.Fatalf("expected ready item with link/convert actions, got %+v", ready)
	}
	for _, action := range ready.SuggestedActions {
		if !action.RequiresConfirmation {
			t.Fatalf("mutating action must require confirmation, got %+v", action)
		}
	}
	incomplete := payload.Reviews[0].Items[1]
	if incomplete.Status != "needs_review" || !slices.Contains(incomplete.MissingFields, "quantity") || !slices.Contains(incomplete.MissingFields, "item_price") {
		t.Fatalf("expected incomplete item missing field review, got %+v", incomplete)
	}
	if len(incomplete.SuggestedActions) != 1 || incomplete.SuggestedActions[0].ID != "complete_purchase_item_fields" || incomplete.SuggestedActions[0].RequiresConfirmation {
		t.Fatalf("expected non-mutating complete-fields action, got %+v", incomplete.SuggestedActions)
	}

	items := doRequest(t, a, http.MethodGet, "/api/items?profile_id="+profile.ID, nil, nil)
	if items.Code != http.StatusOK {
		t.Fatalf("items status=%d body=%s", items.Code, items.Body.String())
	}
	if strings.Contains(items.Body.String(), "Accompanying Flute") || strings.Contains(items.Body.String(), "Mystery purchase") {
		t.Fatalf("Purchase Inbox review API must not create inventory items, body=%s", items.Body.String())
	}
}

func TestPurchaseInboxActionsRequireConfirmationAndApplyToActiveProfile(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Purchase Inbox Actions"}`), map[string]string{"Content-Type": "application/json"})
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

	card := `"card":{"order_id":"20-14595-70928","listing_id":"316046161178","transaction_id":"10080684936020","listing_title":"Accompanying Flute listing","purchased_identity":"Accompanying Flute TWM 142 (142/167)","quantity":4,"item_price":"AU $2.40","item_url":"https://www.ebay.com.au/itm/316046161178"}`
	unconfirmed := doRequest(t, a, http.MethodPost, "/api/integrations/ebay/purchase-inbox/actions", strings.NewReader(`{"action_id":"convert_to_inventory_item","target_key":"10080684936020","confirmed":false,`+card+`}`), map[string]string{"Content-Type": "application/json"})
	if unconfirmed.Code != http.StatusConflict || !strings.Contains(unconfirmed.Body.String(), "purchase_inbox_action_requires_confirmation") {
		t.Fatalf("expected confirmation conflict, status=%d body=%s", unconfirmed.Code, unconfirmed.Body.String())
	}

	convert := doRequest(t, a, http.MethodPost, "/api/integrations/ebay/purchase-inbox/actions", strings.NewReader(`{"action_id":"convert_to_inventory_item","target_key":"10080684936020","confirmed":true,`+card+`}`), map[string]string{"Content-Type": "application/json"})
	if convert.Code != http.StatusOK {
		t.Fatalf("convert status=%d body=%s", convert.Code, convert.Body.String())
	}
	var converted struct {
		ProfileID   string `json:"profile_id"`
		CreatedItem struct {
			ID         string   `json:"id"`
			Title      string   `json:"title"`
			PartNumber string   `json:"part_number"`
			SourceURLs []string `json:"source_urls"`
		} `json:"created_item"`
	}
	if err := json.NewDecoder(convert.Body).Decode(&converted); err != nil {
		t.Fatalf("decode converted: %v body=%s", err, convert.Body.String())
	}
	if converted.ProfileID != profile.ID || converted.CreatedItem.ID == "" || converted.CreatedItem.PartNumber != "316046161178" {
		t.Fatalf("expected active-profile converted inventory item, got %+v", converted)
	}
	if converted.CreatedItem.Title != "Accompanying Flute TWM 142 (142/167)" || !slices.Contains(converted.CreatedItem.SourceURLs, "https://www.ebay.com.au/itm/316046161178") {
		t.Fatalf("expected purchase identity and source URL, got %+v", converted.CreatedItem)
	}

	createExisting := doRequest(t, a, http.MethodPost, "/api/items", strings.NewReader(`{"part_number":"CAB-PUR-EXISTING","title":"Existing Purchase Target","brand":"AFX","category":"Slot Cars"}`), map[string]string{"Content-Type": "application/json"})
	if createExisting.Code != http.StatusCreated {
		t.Fatalf("create existing item status=%d body=%s", createExisting.Code, createExisting.Body.String())
	}
	var existing struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createExisting.Body).Decode(&existing); err != nil {
		t.Fatalf("decode existing: %v", err)
	}
	linkCard := `"card":{"order_id":"20-14595-70929","listing_id":"316046161180","transaction_id":"10080684936021","listing_title":"Existing listing","purchased_identity":"Existing Purchase Target","quantity":1,"item_price":"AU $9.00","item_url":"https://www.ebay.com.au/itm/316046161180"}`
	link := doRequest(t, a, http.MethodPost, "/api/integrations/ebay/purchase-inbox/actions", strings.NewReader(`{"action_id":"link_existing_inventory_item","target_key":"10080684936021","existing_item_id":"`+existing.ID+`","confirmed":true,`+linkCard+`}`), map[string]string{"Content-Type": "application/json"})
	if link.Code != http.StatusOK || !strings.Contains(link.Body.String(), "316046161180") {
		t.Fatalf("link status=%d body=%s", link.Code, link.Body.String())
	}
}
