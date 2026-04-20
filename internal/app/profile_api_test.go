package app

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/jpeg"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"
)

func uploadTestPhoto(t *testing.T, a *App, itemID string) {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			img.Set(x, y, color.RGBA{R: 200, G: 20, B: 40, A: 255})
		}
	}

	var imageBytes bytes.Buffer
	if err := jpeg.Encode(&imageBytes, img, nil); err != nil {
		t.Fatalf("encode test image: %v", err)
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "test.jpg")
	if err != nil {
		t.Fatalf("create multipart file: %v", err)
	}
	if _, err := part.Write(imageBytes.Bytes()); err != nil {
		t.Fatalf("write multipart file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	resp := doRequest(t, a, http.MethodPost, "/api/items/"+itemID+"/photos", &body, map[string]string{"Content-Type": writer.FormDataContentType()})
	if resp.Code != http.StatusCreated {
		t.Fatalf("upload photo status=%d body=%s", resp.Code, resp.Body.String())
	}
}

func TestProfileStorageSecretAndLicenseEndpoints(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	create := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"P1"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", create.Code, create.Body.String())
	}
	var p struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&p); err != nil {
		t.Fatalf("decode profile: %v", err)
	}

	storage := doRequest(t, a, http.MethodGet, "/api/profiles/"+p.ID+"/storage", nil, nil)
	if storage.Code != http.StatusOK {
		t.Fatalf("storage status=%d body=%s", storage.Code, storage.Body.String())
	}
	var st map[string]string
	if err := json.NewDecoder(storage.Body).Decode(&st); err != nil {
		t.Fatalf("decode storage: %v", err)
	}
	if strings.TrimSpace(st["db_path"]) == "" || strings.TrimSpace(st["media_dir"]) == "" {
		t.Fatalf("expected non-empty storage paths: %+v", st)
	}

	putSecret := doRequest(t, a, http.MethodPut, "/api/profiles/"+p.ID+"/secrets", strings.NewReader(`{"key":"openai_api_key","value":"sk-test"}`), map[string]string{"Content-Type": "application/json"})
	if putSecret.Code != http.StatusOK {
		t.Fatalf("put secret status=%d body=%s", putSecret.Code, putSecret.Body.String())
	}
	getSecret := doRequest(t, a, http.MethodGet, "/api/profiles/"+p.ID+"/secrets?key=openai_api_key", nil, nil)
	if getSecret.Code != http.StatusOK {
		t.Fatalf("get secret status=%d body=%s", getSecret.Code, getSecret.Body.String())
	}

	putLicense := doRequest(t, a, http.MethodPut, "/api/profiles/"+p.ID+"/license", strings.NewReader(`{"license_json":"{\"tier\":\"pro\"}"}`), map[string]string{"Content-Type": "application/json"})
	if putLicense.Code != http.StatusOK {
		t.Fatalf("put license status=%d body=%s", putLicense.Code, putLicense.Body.String())
	}
	getLicense := doRequest(t, a, http.MethodGet, "/api/profiles/"+p.ID+"/license", nil, nil)
	if getLicense.Code != http.StatusOK {
		t.Fatalf("get license status=%d body=%s", getLicense.Code, getLicense.Body.String())
	}
}

func TestStorageMaintenanceEndpoints(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	create := doRequest(t, a, http.MethodPost, "/api/items", strings.NewReader(`{"part_number":"STORAGE-001","title":"Storage Item"}`), map[string]string{"Content-Type": "application/json"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create item status=%d body=%s", create.Code, create.Body.String())
	}
	var item struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(create.Body).Decode(&item); err != nil {
		t.Fatalf("decode item: %v", err)
	}

	uploadTestPhoto(t, a, item.ID)

	reindex := doRequest(t, a, http.MethodPost, "/api/data/reindex", strings.NewReader(`{}`), map[string]string{"Content-Type": "application/json"})
	if reindex.Code != http.StatusOK {
		t.Fatalf("reindex status=%d body=%s", reindex.Code, reindex.Body.String())
	}

	rebuild := doRequest(t, a, http.MethodPost, "/api/data/rebuild-thumbnails", strings.NewReader(`{}`), map[string]string{"Content-Type": "application/json"})
	if rebuild.Code != http.StatusOK {
		t.Fatalf("rebuild thumbnails status=%d body=%s", rebuild.Code, rebuild.Body.String())
	}
	var rebuildPayload map[string]any
	if err := json.NewDecoder(rebuild.Body).Decode(&rebuildPayload); err != nil {
		t.Fatalf("decode rebuild response: %v", err)
	}
	if rebuildPayload["ok"] != true {
		t.Fatalf("expected ok rebuild response, got %+v", rebuildPayload)
	}
	if rebuildPayload["rebuilt_items"].(float64) < 1 || rebuildPayload["rebuilt_photos"].(float64) < 1 {
		t.Fatalf("expected rebuilt item/photo counts, got %+v", rebuildPayload)
	}
}
