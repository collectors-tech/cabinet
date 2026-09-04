package app

import (
	"context"
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"testing"
)

func TestPurchaseInboxLoadsDurableCompanionCardsAndCommitsHandOff(t *testing.T) {
	t.Parallel()
	a := newTestApp(t)
	profileID := prepareCompanionAPIProfile(t, a)
	card := `{"order_id":"ORDER-CAPTURED","listing_id":"12345","transaction_id":"TX-CAPTURED","listing_title":"Durable purchase","purchased_identity":"Durable purchase","quantity":1,"item_price":"AU $12.00","item_url":"https://www.ebay.com/itm/12345"}`
	if _, err := a.db.ExecContext(context.Background(), `
		INSERT INTO companion_captures(id, profile_id, session_id, module_id, module_version, schema_version, provider_id,
			integration_instance_id, payload_type, source_url, captured_at, payload_hash, idempotency_key,
			redaction_summary_json, raw_payload_json, state, created_at, updated_at)
		VALUES ('capture-purchase','`+profileID+`','session','ebay-purchase-capture','1.0.0','1','ebay','instance','purchase_order','https://www.ebay.com/mye/myebay/purchase','2026-08-06T00:00:00Z','sha256:capture','capture-purchase','[]','{}','review','2026-08-06T00:00:00Z','2026-08-06T00:00:00Z');
		INSERT INTO companion_purchase_inbox(id, capture_id, profile_id, provider_id, order_key, item_key, card_json, state, first_seen, last_seen, created_at, updated_at)
		VALUES ('purchase-1','capture-purchase','`+profileID+`','ebay','ORDER-CAPTURED','TX-CAPTURED',?,'review','2026-08-06T00:00:00Z','2026-08-06T00:00:00Z','2026-08-06T00:00:00Z','2026-08-06T00:00:00Z')
	`, card); err != nil {
		t.Fatalf("seed durable companion purchase: %v", err)
	}

	reviews := doRequest(t, a, http.MethodPost, "/api/integrations/ebay/purchase-inbox/reviews", strings.NewReader(`{}`), map[string]string{"Content-Type": "application/json"})
	if reviews.Code != http.StatusOK || !strings.Contains(reviews.Body.String(), "Durable purchase") {
		t.Fatalf("durable reviews status=%d body=%s", reviews.Code, reviews.Body.String())
	}
	action := doRequest(t, a, http.MethodPost, "/api/integrations/ebay/purchase-inbox/actions", strings.NewReader(`{"action_id":"convert_to_inventory_item","target_key":"TX-CAPTURED","confirmed":true,"card":`+card+`}`), map[string]string{"Content-Type": "application/json"})
	if action.Code != http.StatusOK {
		t.Fatalf("durable hand-off status=%d body=%s", action.Code, action.Body.String())
	}
	var state, linkedItemID string
	if err := a.db.QueryRowContext(context.Background(), `SELECT state, linked_item_id FROM companion_purchase_inbox WHERE id = 'purchase-1'`).Scan(&state, &linkedItemID); err != nil {
		t.Fatalf("load durable hand-off: %v", err)
	}
	if state != "converted" || linkedItemID == "" {
		t.Fatalf("durable hand-off state=%q linked_item_id=%q", state, linkedItemID)
	}
}

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
