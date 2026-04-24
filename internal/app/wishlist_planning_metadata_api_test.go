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

	createWish := doRequest(t, a, http.MethodPost, "/api/wishlist", strings.NewReader(`{"item_id":"`+item.ID+`","target_price":42,"priority":"high","notes":"manual watch state","below_target_now":true,"highlight_hit":true}`), map[string]string{"Content-Type": "application/json"})
	if createWish.Code != http.StatusCreated {
		t.Fatalf("create wishlist status=%d body=%s", createWish.Code, createWish.Body.String())
	}
	var created struct {
		ID             string `json:"id"`
		BelowTargetNow bool   `json:"below_target_now"`
		HighlightHit   bool   `json:"highlight_hit"`
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

	updateWish := doRequest(t, a, http.MethodPut, "/api/wishlist", strings.NewReader(`{"id":"`+created.ID+`","item_id":"`+item.ID+`","target_price":42,"priority":"medium","notes":"updated manual watch state","below_target_now":false,"highlight_hit":false}`), map[string]string{"Content-Type": "application/json"})
	if updateWish.Code != http.StatusOK {
		t.Fatalf("update wishlist status=%d body=%s", updateWish.Code, updateWish.Body.String())
	}

	listWish := doRequest(t, a, http.MethodGet, "/api/wishlist", nil, nil)
	if listWish.Code != http.StatusOK {
		t.Fatalf("list wishlist status=%d body=%s", listWish.Code, listWish.Body.String())
	}
	var listPayload struct {
		Items []struct {
			ID             string `json:"id"`
			BelowTargetNow bool   `json:"below_target_now"`
			HighlightHit   bool   `json:"highlight_hit"`
			Priority       string `json:"priority"`
			Notes          string `json:"notes"`
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
}
