package app

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/jpeg"
	"net/http"
	"os"
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

func TestPhotoUploadCreatesCanonicalAssetFolderAndManifest(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	if _, err := a.db.Exec(`INSERT INTO canonical_items (id, brand, category, part_number, title) VALUES ('item-canonical','AFX','Slot Car','P-CANON','Canonical Car')`); err != nil {
		t.Fatalf("seed item: %v", err)
	}

	body, contentType := buildMultipartPhoto(t, "front:angle.jpg", sampleJPEG(t))
	resp := doRequest(t, a, http.MethodPost, "/api/items/item-canonical/photos", body, map[string]string{"Content-Type": contentType})
	if resp.Code != http.StatusCreated {
		t.Fatalf("upload status = %d body=%s", resp.Code, resp.Body.String())
	}

	var created struct {
		ID            string `json:"id"`
		OriginalPath  string `json:"original_path"`
		PreviewPath   string `json:"preview_path"`
		ThumbnailPath string `json:"thumbnail_path"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("decode photo: %v", err)
	}

	assetDir := filepath.Join(a.cfg.DataDir, "media", "assets", created.ID)
	if created.OriginalPath != filepath.Join(assetDir, "original", "front_angle.jpg") {
		t.Fatalf("expected canonical original path under asset folder, got %s", created.OriginalPath)
	}
	if created.PreviewPath != filepath.Join(assetDir, "renditions", "preview.jpg") {
		t.Fatalf("expected deterministic preview path, got %s", created.PreviewPath)
	}
	if created.ThumbnailPath != filepath.Join(assetDir, "renditions", "thumbnail.jpg") {
		t.Fatalf("expected deterministic thumbnail path, got %s", created.ThumbnailPath)
	}
	for _, dir := range []string{"original", "renditions", "variations"} {
		info, err := os.Stat(filepath.Join(assetDir, dir))
		if err != nil || !info.IsDir() {
			t.Fatalf("expected asset %s directory, info=%v err=%v", dir, info, err)
		}
	}

	rawManifest, err := os.ReadFile(filepath.Join(assetDir, "manifest.json"))
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	var manifest struct {
		Version  int    `json:"version"`
		AssetID  string `json:"asset_id"`
		Original struct {
			Filename     string `json:"filename"`
			RelativePath string `json:"relative_path"`
			ContentHash  string `json:"content_hash"`
			MIMEType     string `json:"mime_type"`
			ByteSize     int64  `json:"byte_size"`
			Immutable    bool   `json:"immutable"`
		} `json:"original"`
		Renditions []struct {
			Name         string `json:"name"`
			RelativePath string `json:"relative_path"`
		} `json:"renditions"`
		Owners []struct {
			Type string `json:"type"`
			ID   string `json:"id"`
		} `json:"owners"`
	}
	if err := json.Unmarshal(rawManifest, &manifest); err != nil {
		t.Fatalf("decode manifest: %v", err)
	}
	if manifest.Version != 1 || manifest.AssetID != created.ID || manifest.Original.Filename != "front:angle.jpg" || manifest.Original.RelativePath != "original/front_angle.jpg" || manifest.Original.MIMEType != "image/jpeg" || manifest.Original.ByteSize == 0 || !manifest.Original.Immutable || !strings.HasPrefix(manifest.Original.ContentHash, "sha256:") {
		t.Fatalf("unexpected manifest original metadata: %+v", manifest)
	}
	if len(manifest.Renditions) != 2 || manifest.Renditions[0].Name != "preview" || manifest.Renditions[0].RelativePath != "renditions/preview.jpg" || manifest.Renditions[1].Name != "thumbnail" || manifest.Renditions[1].RelativePath != "renditions/thumbnail.jpg" {
		t.Fatalf("unexpected manifest renditions: %+v", manifest.Renditions)
	}
	if len(manifest.Owners) != 1 || manifest.Owners[0].Type != "inventory_item" || manifest.Owners[0].ID != "item-canonical" {
		t.Fatalf("unexpected manifest owners: %+v", manifest.Owners)
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

func TestPhotosRotateEndpoint_UpdatesStoredOrientation(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	if _, err := a.db.Exec(`INSERT INTO canonical_items (id, brand, category, part_number, title) VALUES ('item-1','AFX','Slot Car','P-1','Car')`); err != nil {
		t.Fatalf("seed item: %v", err)
	}

	body, contentType := buildMultipartPhoto(t, "wide.jpg", rectangularJPEG(t, 80, 32))
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

	rotateRight := doRequest(t, a, http.MethodPut, "/api/items/item-1/photos/"+created.ID+"/rotate", strings.NewReader(`{"direction":"right"}`), map[string]string{"Content-Type": "application/json"})
	if rotateRight.Code != http.StatusOK {
		t.Fatalf("rotate right status = %d body=%s", rotateRight.Code, rotateRight.Body.String())
	}
	originalRight := doRequest(t, a, http.MethodGet, "/api/items/item-1/photos/"+created.ID+"/file?variant=original", nil, nil)
	if originalRight.Code != http.StatusOK {
		t.Fatalf("rotated original status = %d body=%s", originalRight.Code, originalRight.Body.String())
	}
	width, height := jpegDimensions(t, originalRight.Body.Bytes())
	if width != 32 || height != 80 {
		t.Fatalf("expected right-rotated dimensions 32x80, got %dx%d", width, height)
	}

	rotateLeft := doRequest(t, a, http.MethodPut, "/api/items/item-1/photos/"+created.ID+"/rotate", strings.NewReader(`{"direction":"left"}`), map[string]string{"Content-Type": "application/json"})
	if rotateLeft.Code != http.StatusOK {
		t.Fatalf("rotate left status = %d body=%s", rotateLeft.Code, rotateLeft.Body.String())
	}
	originalLeft := doRequest(t, a, http.MethodGet, "/api/items/item-1/photos/"+created.ID+"/file?variant=original", nil, nil)
	if originalLeft.Code != http.StatusOK {
		t.Fatalf("restored original status = %d body=%s", originalLeft.Code, originalLeft.Body.String())
	}
	width, height = jpegDimensions(t, originalLeft.Body.Bytes())
	if width != 80 || height != 32 {
		t.Fatalf("expected left-rotated dimensions 80x32, got %dx%d", width, height)
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

func rectangularJPEG(t *testing.T, width, height int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: uint8((x * 255) / width), G: uint8((y * 255) / height), B: 120, A: 255})
		}
	}
	var b bytes.Buffer
	if err := jpeg.Encode(&b, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("encode rectangular jpeg: %v", err)
	}
	return b.Bytes()
}

func jpegDimensions(t *testing.T, data []byte) (int, int) {
	t.Helper()
	cfg, err := jpeg.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode jpeg config: %v", err)
	}
	return cfg.Width, cfg.Height
}
