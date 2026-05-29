package app

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMediaWorkspaceAssetsAPIScopesActiveProfileAndFiltersUnlinked(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	profileResp := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Media API"}`), map[string]string{"Content-Type": "application/json"})
	if profileResp.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", profileResp.Code, profileResp.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(profileResp.Body).Decode(&profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	activeResp := doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+profile.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if activeResp.Code != http.StatusOK {
		t.Fatalf("set active profile status=%d body=%s", activeResp.Code, activeResp.Body.String())
	}
	itemResp := doRequest(t, a, http.MethodPost, "/api/items", strings.NewReader(`{"part_number":"MEDIA-API-1","title":"Media API Item","brand":"AFX","category":"Slot"}`), map[string]string{"Content-Type": "application/json"})
	if itemResp.Code != http.StatusCreated {
		t.Fatalf("create item status=%d body=%s", itemResp.Code, itemResp.Body.String())
	}
	var item struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(itemResp.Body).Decode(&item); err != nil {
		t.Fatalf("decode item: %v", err)
	}
	body, contentType := buildMultipartPhoto(t, "front.jpg", sampleJPEG(t))
	photoResp := doRequest(t, a, http.MethodPost, "/api/items/"+item.ID+"/photos", body, map[string]string{"Content-Type": contentType})
	if photoResp.Code != http.StatusCreated {
		t.Fatalf("upload photo status=%d body=%s", photoResp.Code, photoResp.Body.String())
	}
	if _, err := a.db.Exec(`
		INSERT INTO chat_threads (id, profile_id, title) VALUES ('media-api-thread', ?, 'Media API');
		INSERT INTO chat_attachments (id, profile_id, thread_id, filename, mime_type, size_bytes, stored_path)
		VALUES ('media-api-attachment', ?, 'media-api-thread', 'loose-reference.jpg', 'image/jpeg', 123, '/tmp/loose-reference.jpg');
	`, profile.ID, profile.ID); err != nil {
		t.Fatalf("seed chat attachment: %v", err)
	}

	allResp := doRequest(t, a, http.MethodGet, "/api/media/assets", nil, nil)
	if allResp.Code != http.StatusOK {
		t.Fatalf("media assets status=%d body=%s", allResp.Code, allResp.Body.String())
	}
	var all struct {
		Assets []struct {
			ID           string `json:"id"`
			LinkageState string `json:"linkage_state"`
		} `json:"assets"`
		Summary struct {
			Total           int `json:"total"`
			Unlinked        int `json:"unlinked"`
			LinkedInventory int `json:"linked_inventory"`
		} `json:"summary"`
	}
	if err := json.NewDecoder(allResp.Body).Decode(&all); err != nil {
		t.Fatalf("decode media assets: %v", err)
	}
	if len(all.Assets) != 2 || all.Summary.Total != 2 || all.Summary.Unlinked != 1 || all.Summary.LinkedInventory != 1 {
		t.Fatalf("unexpected media assets response: %+v", all)
	}

	unlinkedResp := doRequest(t, a, http.MethodGet, "/api/media/assets?filter=unlinked", nil, nil)
	if unlinkedResp.Code != http.StatusOK {
		t.Fatalf("unlinked media status=%d body=%s", unlinkedResp.Code, unlinkedResp.Body.String())
	}
	var unlinked struct {
		Assets []struct {
			ID           string `json:"id"`
			LinkageState string `json:"linkage_state"`
		} `json:"assets"`
		Summary struct {
			Total int `json:"total"`
		} `json:"summary"`
	}
	if err := json.NewDecoder(unlinkedResp.Body).Decode(&unlinked); err != nil {
		t.Fatalf("decode unlinked media assets: %v", err)
	}
	if len(unlinked.Assets) != 1 || unlinked.Assets[0].ID != "media-api-attachment" || unlinked.Assets[0].LinkageState != "unlinked" || unlinked.Summary.Total != 2 {
		t.Fatalf("unexpected unlinked media response: %+v", unlinked)
	}
}

func TestMediaWorkspacePreviewAPIsAreExplicitlyNonMutating(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	profileResp := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Media Preview"}`), map[string]string{"Content-Type": "application/json"})
	if profileResp.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", profileResp.Code, profileResp.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(profileResp.Body).Decode(&profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	activeResp := doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+profile.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if activeResp.Code != http.StatusOK {
		t.Fatalf("set active profile status=%d body=%s", activeResp.Code, activeResp.Body.String())
	}
	itemResp := doRequest(t, a, http.MethodPost, "/api/items", strings.NewReader(`{"part_number":"MEDIA-PREV-1","title":"Media Preview Item","brand":"AFX","category":"Slot"}`), map[string]string{"Content-Type": "application/json"})
	if itemResp.Code != http.StatusCreated {
		t.Fatalf("create item status=%d body=%s", itemResp.Code, itemResp.Body.String())
	}
	var item struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(itemResp.Body).Decode(&item); err != nil {
		t.Fatalf("decode item: %v", err)
	}
	if _, err := a.db.Exec(`
		INSERT INTO wishlist_entries (id, profile_id, item_id) VALUES ('media-preview-wish', ?, ?);
		INSERT INTO chat_threads (id, profile_id, title) VALUES ('media-preview-thread', ?, 'Media Preview');
		INSERT INTO chat_attachments (id, profile_id, thread_id, filename, mime_type, size_bytes, stored_path)
		VALUES ('media-preview-attachment', ?, 'media-preview-thread', 'wishlist-reference.jpg', 'image/jpeg', 123, '/tmp/wishlist-reference.jpg');
	`, profile.ID, item.ID, profile.ID, profile.ID); err != nil {
		t.Fatalf("seed preview data: %v", err)
	}

	assignResp := doRequest(t, a, http.MethodPost, "/api/media/assignments/preview", strings.NewReader(`{"asset_id":"media-preview-attachment","target_type":"wishlist","target_id":"media-preview-wish"}`), map[string]string{"Content-Type": "application/json"})
	if assignResp.Code != http.StatusOK {
		t.Fatalf("assignment preview status=%d body=%s", assignResp.Code, assignResp.Body.String())
	}
	var assign struct {
		Allowed               bool   `json:"allowed"`
		RequiresConfirmation  bool   `json:"requires_confirmation"`
		ProjectedLinkageState string `json:"projected_linkage_state"`
		BlockedReason         string `json:"blocked_reason"`
	}
	if err := json.NewDecoder(assignResp.Body).Decode(&assign); err != nil {
		t.Fatalf("decode assignment preview: %v", err)
	}
	if assign.Allowed || !assign.RequiresConfirmation || assign.ProjectedLinkageState != "linked_wishlist" || assign.BlockedReason == "" {
		t.Fatalf("unexpected assignment preview: %+v", assign)
	}

	downloadResp := doRequest(t, a, http.MethodPost, "/api/media/downloads/preview", strings.NewReader(`{"asset_ids":["media-preview-attachment"],"filter":"all"}`), map[string]string{"Content-Type": "application/json"})
	if downloadResp.Code != http.StatusOK {
		t.Fatalf("download preview status=%d body=%s", downloadResp.Code, downloadResp.Body.String())
	}
	var download struct {
		Allowed   bool     `json:"allowed"`
		Count     int      `json:"count"`
		Filenames []string `json:"filenames"`
	}
	if err := json.NewDecoder(downloadResp.Body).Decode(&download); err != nil {
		t.Fatalf("decode download preview: %v", err)
	}
	if !download.Allowed || download.Count != 1 || download.Filenames[0] != "wishlist-reference-jpg-media-pr.jpg" {
		t.Fatalf("unexpected download preview: %+v", download)
	}
}

func TestMediaWorkspaceDownloadAPIReturnsScopedZipPayload(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	profileResp := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Media Download"}`), map[string]string{"Content-Type": "application/json"})
	if profileResp.Code != http.StatusCreated {
		t.Fatalf("create profile status=%d body=%s", profileResp.Code, profileResp.Body.String())
	}
	var profile struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(profileResp.Body).Decode(&profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	activeResp := doRequest(t, a, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+profile.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if activeResp.Code != http.StatusOK {
		t.Fatalf("set active profile status=%d body=%s", activeResp.Code, activeResp.Body.String())
	}
	itemResp := doRequest(t, a, http.MethodPost, "/api/items", strings.NewReader(`{"part_number":"MEDIA-DL-1","title":"Media Download Item","brand":"AFX","category":"Slot"}`), map[string]string{"Content-Type": "application/json"})
	if itemResp.Code != http.StatusCreated {
		t.Fatalf("create item status=%d body=%s", itemResp.Code, itemResp.Body.String())
	}
	var item struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(itemResp.Body).Decode(&item); err != nil {
		t.Fatalf("decode item: %v", err)
	}
	body, contentType := buildMultipartPhoto(t, "front.jpg", sampleJPEG(t))
	photoResp := doRequest(t, a, http.MethodPost, "/api/items/"+item.ID+"/photos", body, map[string]string{"Content-Type": contentType})
	if photoResp.Code != http.StatusCreated {
		t.Fatalf("upload photo status=%d body=%s", photoResp.Code, photoResp.Body.String())
	}
	var photo struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(photoResp.Body).Decode(&photo); err != nil {
		t.Fatalf("decode photo: %v", err)
	}
	attachmentPath := filepath.Join(t.TempDir(), "loose-reference.jpg")
	attachmentBytes := []byte("api attachment bytes")
	if err := os.WriteFile(attachmentPath, attachmentBytes, 0o644); err != nil {
		t.Fatalf("write attachment fixture: %v", err)
	}
	if _, err := a.db.Exec(`INSERT INTO chat_threads (id, profile_id, title) VALUES ('media-download-thread', ?, 'Media Download')`, profile.ID); err != nil {
		t.Fatalf("seed chat thread: %v", err)
	}
	if _, err := a.db.Exec(`
		INSERT INTO chat_attachments (id, profile_id, thread_id, filename, mime_type, size_bytes, stored_path)
		VALUES ('media-download-attachment', ?, 'media-download-thread', 'loose-reference.jpg', 'image/jpeg', 123, ?)
	`, profile.ID, attachmentPath); err != nil {
		t.Fatalf("seed chat attachment: %v", err)
	}

	downloadResp := doRequest(t, a, http.MethodPost, "/api/media/downloads", strings.NewReader(`{"asset_ids":["`+photo.ID+`","media-download-attachment"],"filter":"all"}`), map[string]string{"Content-Type": "application/json"})
	if downloadResp.Code != http.StatusOK {
		assetsResp := doRequest(t, a, http.MethodGet, "/api/media/assets", nil, nil)
		t.Fatalf("download status=%d body=%s assets=%s", downloadResp.Code, downloadResp.Body.String(), assetsResp.Body.String())
	}
	if got := downloadResp.Header().Get("Content-Type"); !strings.Contains(got, "application/zip") {
		t.Fatalf("download content type = %q", got)
	}
	if got := downloadResp.Header().Get("Content-Disposition"); !strings.Contains(got, "cabinet-media-download.zip") {
		t.Fatalf("download disposition = %q", got)
	}
	if got := downloadResp.Header().Get("X-Cabinet-Media-Asset-Count"); got != "2" {
		t.Fatalf("asset count header = %q", got)
	}
	zr, err := zip.NewReader(bytes.NewReader(downloadResp.Body.Bytes()), int64(downloadResp.Body.Len()))
	if err != nil {
		t.Fatalf("open zip response: %v", err)
	}
	entries := map[string][]byte{}
	for _, file := range zr.File {
		rc, err := file.Open()
		if err != nil {
			t.Fatalf("open zip entry: %v", err)
		}
		data, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			t.Fatalf("read zip entry: %v", err)
		}
		entries[file.Name] = data
	}
	if _, ok := entries["media-dl-1-media-download-item-"+photo.ID[:8]+".jpg"]; !ok {
		t.Fatalf("inventory photo missing from zip: %v", entries)
	}
	if !bytes.Equal(entries["loose-reference-jpg-media-do.jpg"], attachmentBytes) {
		t.Fatalf("attachment missing from zip: %v", entries)
	}
}
