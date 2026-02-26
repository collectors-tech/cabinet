package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestItemsQuickAddAllowsPartNumberAndTitleOnly(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	resp := doRequest(t, a, http.MethodPost, "/api/items", strings.NewReader(`{"part_number":"QA-001","title":"Quick Add"}`), map[string]string{"Content-Type": "application/json"})
	if resp.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d (%s)", resp.Code, resp.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &created); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if strings.TrimSpace(created["part_number"].(string)) != "QA-001" {
		t.Fatalf("unexpected part_number: %#v", created["part_number"])
	}
	if strings.TrimSpace(created["title"].(string)) != "Quick Add" {
		t.Fatalf("unexpected title: %#v", created["title"])
	}
	if strings.TrimSpace(created["brand"].(string)) == "" {
		t.Fatalf("expected default brand, got %#v", created["brand"])
	}
	if strings.TrimSpace(created["category"].(string)) == "" {
		t.Fatalf("expected default category, got %#v", created["category"])
	}
}

func TestUpdateItemEndpoint(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createResp := doRequest(t, a, http.MethodPost, "/api/items", strings.NewReader(`{"part_number":"UPD-001","title":"Before","brand":"AFX","category":"Cars"}`), map[string]string{"Content-Type": "application/json"})
	if createResp.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", createResp.Code)
	}
	var created map[string]any
	if err := json.Unmarshal(createResp.Body.Bytes(), &created); err != nil {
		t.Fatalf("unmarshal create response: %v", err)
	}
	itemID := created["id"].(string)

	updateResp := doRequest(t, a, http.MethodPut, "/api/items/"+itemID, strings.NewReader(`{"title":"After","brand":"Tyco"}`), map[string]string{"Content-Type": "application/json"})
	if updateResp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", updateResp.Code, updateResp.Body.String())
	}
	var updated map[string]any
	if err := json.Unmarshal(updateResp.Body.Bytes(), &updated); err != nil {
		t.Fatalf("unmarshal update response: %v", err)
	}
	if updated["title"] != "After" {
		t.Fatalf("expected title After, got %#v", updated["title"])
	}
	if updated["brand"] != "Tyco" {
		t.Fatalf("expected brand Tyco, got %#v", updated["brand"])
	}
}

func TestBulkEditItemsEndpoint(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	createFirst := doRequest(t, a, http.MethodPost, "/api/items", strings.NewReader(`{"part_number":"BULK-1","title":"One","brand":"AFX","category":"Cars"}`), map[string]string{"Content-Type": "application/json"})
	createSecond := doRequest(t, a, http.MethodPost, "/api/items", strings.NewReader(`{"part_number":"BULK-2","title":"Two","brand":"Tyco","category":"Cars"}`), map[string]string{"Content-Type": "application/json"})
	if createFirst.Code != http.StatusCreated || createSecond.Code != http.StatusCreated {
		t.Fatalf("seed items failed (%d, %d)", createFirst.Code, createSecond.Code)
	}

	var first map[string]any
	var second map[string]any
	_ = json.Unmarshal(createFirst.Body.Bytes(), &first)
	_ = json.Unmarshal(createSecond.Body.Bytes(), &second)

	payload := `{"item_ids":["` + first["id"].(string) + `","` + second["id"].(string) + `"],"changes":{"category":"Bulk Updated"}}`
	bulkResp := doRequest(t, a, http.MethodPost, "/api/items/bulk-edit", strings.NewReader(payload), map[string]string{"Content-Type": "application/json"})
	if bulkResp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", bulkResp.Code, bulkResp.Body.String())
	}

	listResp := doRequest(t, a, http.MethodGet, "/api/items", nil, nil)
	if listResp.Code != http.StatusOK {
		t.Fatalf("expected list 200, got %d", listResp.Code)
	}
	var listed struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(listResp.Body.Bytes(), &listed); err != nil {
		t.Fatalf("unmarshal list: %v", err)
	}
	for _, item := range listed.Items {
		if strings.HasPrefix(item["part_number"].(string), "BULK-") && item["category"] != "Bulk Updated" {
			t.Fatalf("expected bulk edited category for item %#v", item)
		}
	}
}

func TestListItemsReturnsEmptyArrayWhenNoItems(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	resp := doRequest(t, a, http.MethodGet, "/api/items", nil, nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", resp.Code, resp.Body.String())
	}

	var payload map[string]json.RawMessage
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	rawItems, ok := payload["items"]
	if !ok {
		t.Fatalf("expected items field, got %s", resp.Body.String())
	}
	if string(rawItems) == "null" {
		t.Fatalf("expected items to be empty array, got null: %s", resp.Body.String())
	}
}
