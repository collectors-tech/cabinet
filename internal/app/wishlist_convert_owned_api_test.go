package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestWishlistConvertOwnedEndpointMovesItemToInventoryForActiveProfile(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	p1 := createWishlistConvertProfile(t, a, "Wishlist P1")
	p2 := createWishlistConvertProfile(t, a, "Wishlist P2")
	activateWishlistConvertProfile(t, a, p1.ID)

	createItem := doRequest(t, a, http.MethodPost, "/api/items", strings.NewReader(`{"part_number":"WISH-CONVERT-001","title":"Wishlist Convert Item","brand":"AFX","category":"Slot"}`), map[string]string{"Content-Type": "application/json"})
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

	createWish := doRequest(t, a, http.MethodPost, "/api/wishlist", strings.NewReader(`{"item_id":"`+item.ID+`","target_price":42,"priority":"high","notes":"ready to buy"}`), map[string]string{"Content-Type": "application/json"})
	if createWish.Code != http.StatusCreated {
		t.Fatalf("create wishlist status=%d body=%s", createWish.Code, createWish.Body.String())
	}
	var wish struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createWish.Body).Decode(&wish); err != nil {
		t.Fatalf("decode wishlist: %v", err)
	}
	if wish.ID == "" {
		t.Fatal("expected wishlist id")
	}

	activateWishlistConvertProfile(t, a, p2.ID)
	otherProfileConvert := doRequest(t, a, http.MethodPost, "/api/wishlist/convert-owned", strings.NewReader(`{"id":"`+wish.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if otherProfileConvert.Code != http.StatusBadRequest {
		t.Fatalf("expected cross-profile conversion rejection, status=%d body=%s", otherProfileConvert.Code, otherProfileConvert.Body.String())
	}

	activateWishlistConvertProfile(t, a, p1.ID)
	convert := doRequest(t, a, http.MethodPost, "/api/wishlist/convert-owned", strings.NewReader(`{"id":"`+wish.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if convert.Code != http.StatusOK {
		t.Fatalf("convert wishlist status=%d body=%s", convert.Code, convert.Body.String())
	}

	wishlistItems := doRequest(t, a, http.MethodGet, "/api/items?status=wishlist", nil, nil)
	if wishlistItems.Code != http.StatusOK {
		t.Fatalf("wishlist items status=%d body=%s", wishlistItems.Code, wishlistItems.Body.String())
	}
	var wishlistPayload struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	if err := json.NewDecoder(wishlistItems.Body).Decode(&wishlistPayload); err != nil {
		t.Fatalf("decode wishlist items: %v", err)
	}
	if len(wishlistPayload.Items) != 0 {
		t.Fatalf("expected converted item removed from wishlist, got %+v", wishlistPayload.Items)
	}

	inventoryItems := doRequest(t, a, http.MethodGet, "/api/items", nil, nil)
	if inventoryItems.Code != http.StatusOK {
		t.Fatalf("inventory items status=%d body=%s", inventoryItems.Code, inventoryItems.Body.String())
	}
	var inventoryPayload struct {
		Items []struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"items"`
	}
	if err := json.NewDecoder(inventoryItems.Body).Decode(&inventoryPayload); err != nil {
		t.Fatalf("decode inventory items: %v", err)
	}
	if len(inventoryPayload.Items) != 1 || inventoryPayload.Items[0].ID != item.ID || inventoryPayload.Items[0].Status != "active" {
		t.Fatalf("expected converted item in active inventory, got %+v", inventoryPayload.Items)
	}
}

func createWishlistConvertProfile(t *testing.T, a *App, name string) struct {
	ID string `json:"id"`
} {
	t.Helper()
	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"`+name+`"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	if profile.ID == "" {
		t.Fatal("expected profile id")
	}
	return profile
}

func activateWishlistConvertProfile(t *testing.T, a *App, profileID string) {
	t.Helper()
	setActive := doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+profileID+`"}`), map[string]string{"Content-Type": "application/json"})
	if setActive.Code != http.StatusOK {
		t.Fatalf("set active profile status=%d body=%s", setActive.Code, setActive.Body.String())
	}
	var active struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(setActive.Body).Decode(&active); err != nil {
		t.Fatalf("decode active profile: %v", err)
	}
	if active.ID != profileID {
		t.Fatalf("expected active profile %q, got %q", profileID, active.ID)
	}
}
