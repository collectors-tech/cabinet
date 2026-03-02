package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestInventoryDeleteLifecycleTransitionsAndRestore(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/items", strings.NewReader(`{"part_number":"DEL-001","title":"Delete Lifecycle Item"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", create.Code, create.Body.String())
	}

	var created map[string]any
	if err := json.Unmarshal(create.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create: %v", err)
	}
	itemID, _ := created["id"].(string)
	if itemID == "" {
		t.Fatalf("missing item id in create payload: %v", created)
	}

	firstDelete := doRequest(t, a, http.MethodDelete, "/api/items/"+itemID, nil, nil)
	if firstDelete.Code != http.StatusOK {
		t.Fatalf("first delete status=%d body=%s", firstDelete.Code, firstDelete.Body.String())
	}

	deletedList := doRequest(t, a, http.MethodGet, "/api/items?status=deleted", nil, nil)
	if deletedList.Code != http.StatusOK {
		t.Fatalf("deleted list status=%d body=%s", deletedList.Code, deletedList.Body.String())
	}
	var deletedPayload struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(deletedList.Body.Bytes(), &deletedPayload); err != nil {
		t.Fatalf("decode deleted list: %v", err)
	}
	if len(deletedPayload.Items) != 1 {
		t.Fatalf("expected 1 deleted item, got %d (%s)", len(deletedPayload.Items), deletedList.Body.String())
	}

	secondDelete := doRequest(t, a, http.MethodDelete, "/api/items/"+itemID, nil, nil)
	if secondDelete.Code != http.StatusOK {
		t.Fatalf("second delete status=%d body=%s", secondDelete.Code, secondDelete.Body.String())
	}

	recycleList := doRequest(t, a, http.MethodGet, "/api/items?status=recycle", nil, nil)
	if recycleList.Code != http.StatusOK {
		t.Fatalf("recycle list status=%d body=%s", recycleList.Code, recycleList.Body.String())
	}
	var recyclePayload struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(recycleList.Body.Bytes(), &recyclePayload); err != nil {
		t.Fatalf("decode recycle list: %v", err)
	}
	if len(recyclePayload.Items) != 1 {
		t.Fatalf("expected 1 recycle item, got %d (%s)", len(recyclePayload.Items), recycleList.Body.String())
	}

	restore := doRequest(t, a, http.MethodPost, "/api/items/"+itemID+"/restore", nil, nil)
	if restore.Code != http.StatusOK {
		t.Fatalf("restore status=%d body=%s", restore.Code, restore.Body.String())
	}

	activeList := doRequest(t, a, http.MethodGet, "/api/items", nil, nil)
	if activeList.Code != http.StatusOK {
		t.Fatalf("active list status=%d body=%s", activeList.Code, activeList.Body.String())
	}
	var activePayload struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(activeList.Body.Bytes(), &activePayload); err != nil {
		t.Fatalf("decode active list: %v", err)
	}
	if len(activePayload.Items) != 1 {
		t.Fatalf("expected restored active item, got %d (%s)", len(activePayload.Items), activeList.Body.String())
	}
}

func TestInventoryDeleteLifecyclePermanentDeleteBlockedByDependencies(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/items", strings.NewReader(`{"part_number":"DEL-DEP-001","title":"Delete Dependency Item"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", create.Code, create.Body.String())
	}

	var created map[string]any
	if err := json.Unmarshal(create.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create: %v", err)
	}
	itemID, _ := created["id"].(string)
	if itemID == "" {
		t.Fatalf("missing item id in create payload: %v", created)
	}

	instanceCreate := doRequest(t, a, http.MethodPost, "/api/items/"+itemID+"/instances", strings.NewReader(`{"status":"sealed","quantity":1}`), map[string]string{"Content-Type": "application/json"})
	if instanceCreate.Code != http.StatusCreated {
		t.Fatalf("create instance status=%d body=%s", instanceCreate.Code, instanceCreate.Body.String())
	}

	if resp := doRequest(t, a, http.MethodDelete, "/api/items/"+itemID, nil, nil); resp.Code != http.StatusOK {
		t.Fatalf("first delete status=%d body=%s", resp.Code, resp.Body.String())
	}
	if resp := doRequest(t, a, http.MethodDelete, "/api/items/"+itemID, nil, nil); resp.Code != http.StatusOK {
		t.Fatalf("second delete status=%d body=%s", resp.Code, resp.Body.String())
	}

	blocked := doRequest(t, a, http.MethodDelete, "/api/items/"+itemID, nil, nil)
	if blocked.Code != http.StatusConflict {
		t.Fatalf("expected 409 on permanent delete with dependencies, got %d body=%s", blocked.Code, blocked.Body.String())
	}
	var blockedPayload map[string]any
	if err := json.Unmarshal(blocked.Body.Bytes(), &blockedPayload); err != nil {
		t.Fatalf("decode blocked payload: %v", err)
	}
	if blockedPayload["error"] != "item_has_dependencies" {
		t.Fatalf("expected error item_has_dependencies, got %v", blockedPayload["error"])
	}
}

func TestInventoryDeleteLifecyclePermanentDeleteRemovesRecycledItemWithoutDependencies(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/items", strings.NewReader(`{"part_number":"DEL-FINAL-001","title":"Delete Final Item"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", create.Code, create.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(create.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create: %v", err)
	}
	itemID, _ := created["id"].(string)
	if itemID == "" {
		t.Fatalf("missing item id in create payload: %v", created)
	}

	if resp := doRequest(t, a, http.MethodDelete, "/api/items/"+itemID, nil, nil); resp.Code != http.StatusOK {
		t.Fatalf("first delete status=%d body=%s", resp.Code, resp.Body.String())
	}
	if resp := doRequest(t, a, http.MethodDelete, "/api/items/"+itemID, nil, nil); resp.Code != http.StatusOK {
		t.Fatalf("second delete status=%d body=%s", resp.Code, resp.Body.String())
	}
	finalDelete := doRequest(t, a, http.MethodDelete, "/api/items/"+itemID, nil, nil)
	if finalDelete.Code != http.StatusNoContent {
		t.Fatalf("expected 204 on permanent delete without dependencies, got %d body=%s", finalDelete.Code, finalDelete.Body.String())
	}

	recycleList := doRequest(t, a, http.MethodGet, "/api/items?status=recycle", nil, nil)
	if recycleList.Code != http.StatusOK {
		t.Fatalf("recycle list status=%d body=%s", recycleList.Code, recycleList.Body.String())
	}
	var payload struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(recycleList.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode recycle list: %v", err)
	}
	if len(payload.Items) != 0 {
		t.Fatalf("expected item to be permanently removed from recycle list, got %d", len(payload.Items))
	}
}
