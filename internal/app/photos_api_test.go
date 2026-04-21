package app

import (
	"encoding/json"
	"net/http"
	"path/filepath"
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

func TestPhotosReorderEndpoint(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	if _, err := a.db.Exec(`INSERT INTO canonical_items (id, brand, category, part_number, title) VALUES ('item-1','AFX','Slot Car','P-1','Car')`); err != nil {
		t.Fatalf("seed item: %v", err)
	}

	upload := func(filename string) string {
		body, contentType := buildMultipartPhoto(t, filename, sampleJPEG(t))
		resp := doRequest(t, a, http.MethodPost, "/api/items/item-1/photos", body, map[string]string{"Content-Type": contentType})
		if resp.Code != http.StatusCreated {
			t.Fatalf("upload %s status = %d body=%s", filename, resp.Code, resp.Body.String())
		}
		var created struct {
			ID string `json:"id"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
			t.Fatalf("decode upload response: %v", err)
		}
		return created.ID
	}

	p1 := upload("a.jpg")
	p2 := upload("b.jpg")
	p3 := upload("c.jpg")

	reorderPayload := `{"photo_ids":["` + p3 + `","` + p1 + `","` + p2 + `"]}`
	reorderResp := doRequest(t, a, http.MethodPost, "/api/items/item-1/photos/reorder", strings.NewReader(reorderPayload), map[string]string{"Content-Type": "application/json"})
	if reorderResp.Code != http.StatusOK {
		t.Fatalf("reorder status = %d body=%s", reorderResp.Code, reorderResp.Body.String())
	}

	listResp := doRequest(t, a, http.MethodGet, "/api/items/item-1/photos", nil, nil)
	if listResp.Code != http.StatusOK {
		t.Fatalf("list status = %d body=%s", listResp.Code, listResp.Body.String())
	}
	var listed struct {
		Photos []struct {
			ID string `json:"id"`
		} `json:"photos"`
	}
	if err := json.NewDecoder(listResp.Body).Decode(&listed); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(listed.Photos) != 3 {
		t.Fatalf("expected 3 photos, got %d", len(listed.Photos))
	}
	if listed.Photos[0].ID != p3 || listed.Photos[1].ID != p1 || listed.Photos[2].ID != p2 {
		t.Fatalf("unexpected reordered list: %#v", listed.Photos)
	}
}

func TestPhotosUpload_UsesActiveProfileMediaDirectory(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	createProfile := func(name string) string {
		resp := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"`+name+`"}`), map[string]string{"Content-Type": "application/json"})
		if resp.Code != http.StatusCreated {
			t.Fatalf("create profile %s status=%d body=%s", name, resp.Code, resp.Body.String())
		}
		var created struct {
			ID string `json:"id"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
			t.Fatalf("decode profile %s: %v", name, err)
		}
		return created.ID
	}

	profileOneID := createProfile("Photos Profile One")
	profileTwoID := createProfile("Photos Profile Two")

	profileOneMediaDir := filepath.Join(a.cfg.DataDir, "profiles", profileOneID, "media-custom")
	profileTwoMediaDir := filepath.Join(a.cfg.DataDir, "profiles", profileTwoID, "media-custom")

	for profileID, mediaDir := range map[string]string{
		profileOneID: profileOneMediaDir,
		profileTwoID: profileTwoMediaDir,
	} {
		payloadBytes, err := json.Marshal(map[string]any{
			"settings": map[string]string{
				"storage.media_dir": mediaDir,
			},
		})
		if err != nil {
			t.Fatalf("marshal settings for %s: %v", profileID, err)
		}
		saveSettings := doRequest(
			t,
			a,
			http.MethodPut,
			"/api/profiles/"+profileID+"/settings",
			strings.NewReader(string(payloadBytes)),
			map[string]string{"Content-Type": "application/json"},
		)
		if saveSettings.Code != http.StatusOK {
			t.Fatalf("save profile settings %s status=%d body=%s", profileID, saveSettings.Code, saveSettings.Body.String())
		}
	}

	setActiveProfile := func(profileID string) {
		resp := doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+profileID+`"}`), map[string]string{"Content-Type": "application/json"})
		if resp.Code != http.StatusOK {
			t.Fatalf("set active profile %s status=%d body=%s", profileID, resp.Code, resp.Body.String())
		}
	}

	createItem := func(partNumber string) string {
		resp := doRequest(t, a, http.MethodPost, "/api/items", strings.NewReader(`{"part_number":"`+partNumber+`","title":"`+partNumber+`"}`), map[string]string{"Content-Type": "application/json"})
		if resp.Code != http.StatusCreated {
			t.Fatalf("create item %s status=%d body=%s", partNumber, resp.Code, resp.Body.String())
		}
		var created struct {
			ID string `json:"id"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
			t.Fatalf("decode item %s: %v", partNumber, err)
		}
		return created.ID
	}

	uploadPhoto := func(itemID string) struct {
		OriginalPath  string `json:"original_path"`
		PreviewPath   string `json:"preview_path"`
		ThumbnailPath string `json:"thumbnail_path"`
	} {
		body, contentType := buildMultipartPhoto(t, "photo.jpg", sampleJPEG(t))
		resp := doRequest(t, a, http.MethodPost, "/api/items/"+itemID+"/photos", body, map[string]string{"Content-Type": contentType})
		if resp.Code != http.StatusCreated {
			t.Fatalf("upload photo for %s status=%d body=%s", itemID, resp.Code, resp.Body.String())
		}
		var created struct {
			OriginalPath  string `json:"original_path"`
			PreviewPath   string `json:"preview_path"`
			ThumbnailPath string `json:"thumbnail_path"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
			t.Fatalf("decode photo for %s: %v", itemID, err)
		}
		return created
	}

	setActiveProfile(profileOneID)
	itemOneID := createItem("PHOTO-P1")
	photoOne := uploadPhoto(itemOneID)

	setActiveProfile(profileTwoID)
	itemTwoID := createItem("PHOTO-P2")
	photoTwo := uploadPhoto(itemTwoID)

	if !strings.HasPrefix(photoOne.OriginalPath, profileOneMediaDir) {
		t.Fatalf("expected profile one original path in %s, got %s", profileOneMediaDir, photoOne.OriginalPath)
	}
	if !strings.HasPrefix(photoOne.PreviewPath, profileOneMediaDir) {
		t.Fatalf("expected profile one preview path in %s, got %s", profileOneMediaDir, photoOne.PreviewPath)
	}
	if !strings.HasPrefix(photoOne.ThumbnailPath, profileOneMediaDir) {
		t.Fatalf("expected profile one thumbnail path in %s, got %s", profileOneMediaDir, photoOne.ThumbnailPath)
	}
	if !strings.HasPrefix(photoTwo.OriginalPath, profileTwoMediaDir) {
		t.Fatalf("expected profile two original path in %s, got %s", profileTwoMediaDir, photoTwo.OriginalPath)
	}
	if !strings.HasPrefix(photoTwo.PreviewPath, profileTwoMediaDir) {
		t.Fatalf("expected profile two preview path in %s, got %s", profileTwoMediaDir, photoTwo.PreviewPath)
	}
	if !strings.HasPrefix(photoTwo.ThumbnailPath, profileTwoMediaDir) {
		t.Fatalf("expected profile two thumbnail path in %s, got %s", profileTwoMediaDir, photoTwo.ThumbnailPath)
	}
}
