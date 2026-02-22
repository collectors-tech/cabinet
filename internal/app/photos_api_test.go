package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestPhotosFileEndpoint_ServesVariant(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	if _, err := a.db.Exec(`INSERT INTO canonical_items (id, brand, category, part_number, title) VALUES ('item-1','AFX','Slot Car','P-1','Car')`); err != nil {
		t.Fatalf("seed item: %v", err)
	}

	body, contentType := buildMultipartPhoto(t, "photo.jpg", sampleJPEG(t))
	resp := doRequest(t, a, http.MethodPost, "/api/items/item-1/photos", body, map[string]string{"Content-Type": contentType})
	if resp.Code != http.StatusCreated {
		t.Fatalf("upload status = %d body=%s", resp.Code, resp.Body.String())
	}

	var created struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("decode upload response: %v", err)
	}
	if strings.TrimSpace(created.ID) == "" {
		t.Fatal("expected uploaded photo id")
	}

	fileResp := doRequest(t, a, http.MethodGet, "/api/items/item-1/photos/"+created.ID+"/file?variant=preview", nil, nil)
	if fileResp.Code != http.StatusOK {
		t.Fatalf("file status = %d body=%s", fileResp.Code, fileResp.Body.String())
	}
	if ct := fileResp.Header().Get("Content-Type"); !strings.HasPrefix(ct, "image/") {
		t.Fatalf("expected image content type, got %q", ct)
	}
	if fileResp.Body.Len() == 0 {
		t.Fatal("expected non-empty image response body")
	}
}

func TestPhotosFileEndpoint_InvalidVariant(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	if _, err := a.db.Exec(`INSERT INTO canonical_items (id, brand, category, part_number, title) VALUES ('item-1','AFX','Slot Car','P-1','Car')`); err != nil {
		t.Fatalf("seed item: %v", err)
	}

	body, contentType := buildMultipartPhoto(t, "photo.jpg", sampleJPEG(t))
	resp := doRequest(t, a, http.MethodPost, "/api/items/item-1/photos", body, map[string]string{"Content-Type": contentType})
	if resp.Code != http.StatusCreated {
		t.Fatalf("upload status = %d body=%s", resp.Code, resp.Body.String())
	}
	var created struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("decode upload response: %v", err)
	}

	fileResp := doRequest(t, a, http.MethodGet, "/api/items/item-1/photos/"+created.ID+"/file?variant=bad", nil, nil)
	if fileResp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid variant, got %d", fileResp.Code)
	}
}
