package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestWishlistDeleteRestoreAndPermanentDeleteEndpoint(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createProfile := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Wishlist Delete Restore"}`), map[string]string{"Content-Type": "application/json"})
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

	createItem := doRequest(t, a, http.MethodPost, "/api/items", strings.NewReader(`{"part_number":"WISH-DEL-001","title":"Wishlist Delete Restore Item","brand":"AFX","category":"Slot Cars"}`), map[string]string{"Content-Type": "application/json"})
	if createItem.Code != http.StatusCreated {
		t.Fatalf("create item status=%d body=%s", createItem.Code, createItem.Body.String())
	}
	var item struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createItem.Body).Decode(&item); err != nil {
		t.Fatalf("decode item: %v", err)
	}
	createWish := doRequest(t, a, http.MethodPost, "/api/wishlist", strings.NewReader(`{"item_id":"`+item.ID+`","target_price":42,"priority":"high","notes":"delete restore"}`), map[string]string{"Content-Type": "application/json"})
	if createWish.Code != http.StatusCreated {
		t.Fatalf("create wishlist status=%d body=%s", createWish.Code, createWish.Body.String())
	}
	var wish struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createWish.Body).Decode(&wish); err != nil {
		t.Fatalf("decode wishlist: %v", err)
	}

	softDelete := doRequest(t, a, http.MethodDelete, "/api/wishlist?id="+wish.ID, nil, nil)
	if softDelete.Code != http.StatusNoContent {
		t.Fatalf("soft delete wishlist status=%d body=%s", softDelete.Code, softDelete.Body.String())
	}
	activeList := doRequest(t, a, http.MethodGet, "/api/wishlist", nil, nil)
	if activeList.Code != http.StatusOK {
		t.Fatalf("active wishlist status=%d body=%s", activeList.Code, activeList.Body.String())
	}
	assertWishlistIDs(t, activeList.Body.String(), nil)
	deletedList := doRequest(t, a, http.MethodGet, "/api/wishlist?deleted=true", nil, nil)
	if deletedList.Code != http.StatusOK {
		t.Fatalf("deleted wishlist status=%d body=%s", deletedList.Code, deletedList.Body.String())
	}
	assertWishlistIDs(t, deletedList.Body.String(), []string{wish.ID})

	restore := doRequest(t, a, http.MethodPost, "/api/wishlist/"+wish.ID+"/restore", nil, nil)
	if restore.Code != http.StatusOK {
		t.Fatalf("restore wishlist status=%d body=%s", restore.Code, restore.Body.String())
	}
	restoredList := doRequest(t, a, http.MethodGet, "/api/wishlist", nil, nil)
	if restoredList.Code != http.StatusOK {
		t.Fatalf("restored wishlist status=%d body=%s", restoredList.Code, restoredList.Body.String())
	}
	assertWishlistIDs(t, restoredList.Body.String(), []string{wish.ID})

	if _, err := a.db.Exec(`INSERT INTO instances(id, item_id, condition, status, quantity, notes) VALUES ('wishlist-delete-restore-instance', ?, 'loose', 'loose', 1, 'preserve inventory')`, item.ID); err != nil {
		t.Fatalf("seed inventory instance: %v", err)
	}
	if resp := doRequest(t, a, http.MethodDelete, "/api/wishlist?id="+wish.ID, nil, nil); resp.Code != http.StatusNoContent {
		t.Fatalf("soft delete before permanent status=%d body=%s", resp.Code, resp.Body.String())
	}
	permanentDelete := doRequest(t, a, http.MethodDelete, "/api/wishlist?id="+wish.ID+"&permanent=true", nil, nil)
	if permanentDelete.Code != http.StatusNoContent {
		t.Fatalf("permanent delete wishlist status=%d body=%s", permanentDelete.Code, permanentDelete.Body.String())
	}
	deletedAfterPermanent := doRequest(t, a, http.MethodGet, "/api/wishlist?deleted=true", nil, nil)
	if deletedAfterPermanent.Code != http.StatusOK {
		t.Fatalf("deleted wishlist after permanent status=%d body=%s", deletedAfterPermanent.Code, deletedAfterPermanent.Body.String())
	}
	assertWishlistIDs(t, deletedAfterPermanent.Body.String(), nil)
	var instanceCount int
	if err := a.db.QueryRow(`SELECT COUNT(1) FROM instances WHERE id = 'wishlist-delete-restore-instance'`).Scan(&instanceCount); err != nil {
		t.Fatalf("count inventory instance: %v", err)
	}
	if instanceCount != 1 {
		t.Fatalf("expected inventory instance to survive permanent wishlist delete, got %d", instanceCount)
	}
}

func assertWishlistIDs(t *testing.T, body string, want []string) {
	t.Helper()
	var payload struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	if err := json.NewDecoder(strings.NewReader(body)).Decode(&payload); err != nil {
		t.Fatalf("decode wishlist payload: %v", err)
	}
	if len(payload.Items) != len(want) {
		t.Fatalf("expected wishlist ids %v, got %+v", want, payload.Items)
	}
	for index, id := range want {
		if payload.Items[index].ID != id {
			t.Fatalf("expected wishlist id %q at index %d, got %+v", id, index, payload.Items)
		}
	}
}
