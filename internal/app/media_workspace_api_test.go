package app

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
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

func TestMediaWorkspaceCreateAssetPersistsUnlinkedUploadMetadata(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	profileResp := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Media Upload"}`), map[string]string{"Content-Type": "application/json"})
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

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("title", "Loose chassis reference"); err != nil {
		t.Fatalf("write title field: %v", err)
	}
	if err := writer.WriteField("source", "Bench intake"); err != nil {
		t.Fatalf("write source field: %v", err)
	}
	if err := writer.WriteField("notes", "Rear axle detail"); err != nil {
		t.Fatalf("write notes field: %v", err)
	}
	part, err := writer.CreateFormFile("file", "loose-chassis.jpg")
	if err != nil {
		t.Fatalf("create media form file: %v", err)
	}
	if _, err := part.Write(sampleJPEG(t)); err != nil {
		t.Fatalf("write media form file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	createResp := doRequest(t, a, http.MethodPost, "/api/media/assets", &body, map[string]string{"Content-Type": writer.FormDataContentType()})
	if createResp.Code != http.StatusCreated {
		t.Fatalf("create media asset status=%d body=%s", createResp.Code, createResp.Body.String())
	}
	var created struct {
		AssetID  string `json:"asset_id"`
		Filename string `json:"filename"`
		Title    string `json:"title"`
		Source   string `json:"source"`
		Notes    string `json:"notes"`
	}
	if err := json.NewDecoder(createResp.Body).Decode(&created); err != nil {
		t.Fatalf("decode created media asset: %v", err)
	}
	if created.AssetID == "" || created.Filename != "loose-chassis.jpg" || created.Title != "Loose chassis reference" || created.Source != "Bench intake" || created.Notes != "Rear axle detail" {
		t.Fatalf("unexpected created media asset response: %+v", created)
	}
	var storedPath, uploadThreadID string
	if err := a.db.QueryRow(`SELECT stored_path, thread_id FROM chat_attachments WHERE profile_id = ? AND id = ?`, profile.ID, created.AssetID).Scan(&storedPath, &uploadThreadID); err != nil {
		t.Fatalf("query stored media path: %v", err)
	}
	assetDir := filepath.Join(a.cfg.DataDir, "profiles", profile.ID, "media", "assets", created.AssetID)
	if storedPath != filepath.Join(assetDir, "original", "loose-chassis.jpg") {
		t.Fatalf("expected canonical media workspace stored path, got %s", storedPath)
	}
	if _, err := os.Stat(filepath.Join(assetDir, "original")); err != nil {
		t.Fatalf("expected canonical original dir: %v", err)
	}
	if _, err := os.Stat(filepath.Join(assetDir, "renditions")); err != nil {
		t.Fatalf("expected canonical renditions dir: %v", err)
	}
	if _, err := os.Stat(filepath.Join(assetDir, "variations")); err != nil {
		t.Fatalf("expected canonical variations dir: %v", err)
	}
	rawManifest, err := os.ReadFile(filepath.Join(assetDir, "manifest.json"))
	if err != nil {
		t.Fatalf("read canonical media manifest: %v", err)
	}
	var manifest struct {
		Version  int    `json:"version"`
		AssetID  string `json:"asset_id"`
		Original struct {
			Filename     string `json:"filename"`
			RelativePath string `json:"relative_path"`
			ContentHash  string `json:"content_hash"`
			MIMEType     string `json:"mime_type"`
			Immutable    bool   `json:"immutable"`
		} `json:"original"`
		Owners []struct {
			Type string `json:"type"`
			ID   string `json:"id"`
		} `json:"owners"`
	}
	if err := json.Unmarshal(rawManifest, &manifest); err != nil {
		t.Fatalf("decode canonical media manifest: %v", err)
	}
	if manifest.Version != 1 || manifest.AssetID != created.AssetID || manifest.Original.Filename != "loose-chassis.jpg" || manifest.Original.RelativePath != "original/loose-chassis.jpg" || manifest.Original.MIMEType != "image/jpeg" || !manifest.Original.Immutable || !strings.HasPrefix(manifest.Original.ContentHash, "sha256:") {
		t.Fatalf("unexpected canonical media manifest: %+v", manifest)
	}
	if len(manifest.Owners) != 1 || manifest.Owners[0].Type != "chat_thread" || manifest.Owners[0].ID != uploadThreadID {
		t.Fatalf("unexpected canonical media manifest owners: %+v", manifest.Owners)
	}

	assetsResp := doRequest(t, a, http.MethodGet, "/api/media/assets?filter=unlinked", nil, nil)
	if assetsResp.Code != http.StatusOK {
		t.Fatalf("media assets status=%d body=%s", assetsResp.Code, assetsResp.Body.String())
	}
	var listed struct {
		Assets []struct {
			ID           string `json:"id"`
			Filename     string `json:"filename"`
			LinkageState string `json:"linkage_state"`
		} `json:"assets"`
		Summary struct {
			Total    int `json:"total"`
			Unlinked int `json:"unlinked"`
		} `json:"summary"`
	}
	if err := json.NewDecoder(assetsResp.Body).Decode(&listed); err != nil {
		t.Fatalf("decode listed media assets: %v", err)
	}
	if len(listed.Assets) != 1 || listed.Assets[0].ID != created.AssetID || listed.Assets[0].Filename != "loose-chassis.jpg" || listed.Assets[0].LinkageState != "unlinked" {
		t.Fatalf("created media asset not listed as unlinked: %+v", listed)
	}
	if listed.Summary.Total != 1 || listed.Summary.Unlinked != 1 {
		t.Fatalf("unexpected created media summary: %+v", listed.Summary)
	}

	var metadataJSON string
	if err := a.db.QueryRow(`SELECT context_json FROM chat_messages WHERE profile_id = ? AND content = 'Media asset added from Media workspace.'`, profile.ID).Scan(&metadataJSON); err != nil {
		t.Fatalf("query saved media metadata message: %v", err)
	}
	if !strings.Contains(metadataJSON, `"title":"Loose chassis reference"`) || !strings.Contains(metadataJSON, `"notes":"Rear axle detail"`) {
		t.Fatalf("metadata was not persisted in chat message context: %s", metadataJSON)
	}

	var badBody bytes.Buffer
	badWriter := multipart.NewWriter(&badBody)
	badPart, err := badWriter.CreateFormFile("file", "not-media.txt")
	if err != nil {
		t.Fatalf("create bad media form file: %v", err)
	}
	if _, err := badPart.Write([]byte("not image")); err != nil {
		t.Fatalf("write bad media form file: %v", err)
	}
	if err := badWriter.Close(); err != nil {
		t.Fatalf("close bad multipart writer: %v", err)
	}
	badResp := doRequest(t, a, http.MethodPost, "/api/media/assets", &badBody, map[string]string{"Content-Type": badWriter.FormDataContentType()})
	if badResp.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("unsupported media status=%d body=%s", badResp.Code, badResp.Body.String())
	}
}

func TestMediaWorkspaceUpdateAssetMetadataPersistsEditedFields(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)
	profileResp := doRequest(t, a, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Media Edit"}`), map[string]string{"Content-Type": "application/json"})
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
	if _, err := a.db.Exec(`
		INSERT INTO chat_threads (id, profile_id, title) VALUES ('media-edit-thread', ?, 'Media Uploads');
		INSERT INTO chat_attachments (id, profile_id, thread_id, filename, mime_type, size_bytes, stored_path)
		VALUES ('media-edit-attachment', ?, 'media-edit-thread', 'slot-car-front.jpg', 'image/jpeg', 123, '/tmp/slot-car-front.jpg');
	`, profile.ID, profile.ID); err != nil {
		t.Fatalf("seed editable media asset: %v", err)
	}

	updateResp := doRequest(t, a, http.MethodPatch, "/api/media/assets/media-edit-attachment/metadata", strings.NewReader(`{
		"title":"AFX Mustang hero angle",
		"filename":"slot-car-hero.jpg",
		"source":"Bench edit",
		"download_filename":"afx-mustang-hero-angle.jpg",
		"notes":"Updated crop and metadata"
	}`), map[string]string{"Content-Type": "application/json"})
	if updateResp.Code != http.StatusOK {
		t.Fatalf("update media metadata status=%d body=%s", updateResp.Code, updateResp.Body.String())
	}
	var updated struct {
		ID               string   `json:"id"`
		Title            string   `json:"title"`
		Filename         string   `json:"filename"`
		Source           string   `json:"source"`
		DownloadFilename string   `json:"download_filename"`
		Notes            string   `json:"notes"`
		Variations       []string `json:"thumbnail_variations"`
	}
	if err := json.NewDecoder(updateResp.Body).Decode(&updated); err != nil {
		t.Fatalf("decode updated media metadata: %v", err)
	}
	if updated.ID != "media-edit-attachment" || updated.Title != "AFX Mustang hero angle" || updated.Filename != "slot-car-hero.jpg" || updated.Source != "Bench edit" || updated.DownloadFilename != "afx-mustang-hero-angle.jpg" || updated.Notes != "Updated crop and metadata" || len(updated.Variations) == 0 {
		t.Fatalf("unexpected updated media metadata: %+v", updated)
	}

	assetsResp := doRequest(t, a, http.MethodGet, "/api/media/assets", nil, nil)
	if assetsResp.Code != http.StatusOK {
		t.Fatalf("media assets status=%d body=%s", assetsResp.Code, assetsResp.Body.String())
	}
	var listed struct {
		Assets []struct {
			ID               string `json:"id"`
			Title            string `json:"title"`
			Filename         string `json:"filename"`
			Source           string `json:"source"`
			DownloadFilename string `json:"download_filename"`
			Notes            string `json:"notes"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(assetsResp.Body).Decode(&listed); err != nil {
		t.Fatalf("decode listed media assets: %v", err)
	}
	if len(listed.Assets) != 1 || listed.Assets[0].Title != "AFX Mustang hero angle" || listed.Assets[0].Filename != "slot-car-hero.jpg" || listed.Assets[0].Source != "Bench edit" || listed.Assets[0].DownloadFilename != "afx-mustang-hero-angle.jpg" || listed.Assets[0].Notes != "Updated crop and metadata" {
		t.Fatalf("updated media metadata not reflected in list: %+v", listed)
	}
}

func TestMediaWorkspaceAssignmentAPIPersistsConfirmedLinks(t *testing.T) {
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
	if !assign.Allowed || !assign.RequiresConfirmation || assign.ProjectedLinkageState != "linked_wishlist" || assign.BlockedReason != "" {
		t.Fatalf("unexpected assignment preview: %+v", assign)
	}

	applyResp := doRequest(t, a, http.MethodPost, "/api/media/assignments", strings.NewReader(`{"asset_id":"media-preview-attachment","target_type":"wishlist","target_id":"media-preview-wish"}`), map[string]string{"Content-Type": "application/json"})
	if applyResp.Code != http.StatusOK {
		t.Fatalf("assignment apply status=%d body=%s", applyResp.Code, applyResp.Body.String())
	}
	var applied struct {
		Applied             bool   `json:"applied"`
		CurrentLinkageState string `json:"current_linkage_state"`
		AuditSummary        string `json:"audit_summary"`
	}
	if err := json.NewDecoder(applyResp.Body).Decode(&applied); err != nil {
		t.Fatalf("decode assignment apply: %v", err)
	}
	if !applied.Applied || applied.CurrentLinkageState != "linked_wishlist" || applied.AuditSummary == "" {
		t.Fatalf("unexpected assignment apply response: %+v", applied)
	}

	assetsResp := doRequest(t, a, http.MethodGet, "/api/media/assets", nil, nil)
	if assetsResp.Code != http.StatusOK {
		t.Fatalf("media assets after assignment status=%d body=%s", assetsResp.Code, assetsResp.Body.String())
	}
	var assetsAfter struct {
		Assets []struct {
			ID           string `json:"id"`
			LinkageState string `json:"linkage_state"`
			WishlistID   string `json:"wishlist_id"`
		} `json:"assets"`
		Summary struct {
			Unlinked       int `json:"unlinked"`
			LinkedWishlist int `json:"linked_wishlist"`
		} `json:"summary"`
	}
	if err := json.NewDecoder(assetsResp.Body).Decode(&assetsAfter); err != nil {
		t.Fatalf("decode media assets after assignment: %v", err)
	}
	if len(assetsAfter.Assets) != 1 || assetsAfter.Assets[0].ID != "media-preview-attachment" || assetsAfter.Assets[0].LinkageState != "linked_wishlist" || assetsAfter.Assets[0].WishlistID != "media-preview-wish" {
		t.Fatalf("assignment did not update media asset linkage: %+v", assetsAfter)
	}
	if assetsAfter.Summary.Unlinked != 0 || assetsAfter.Summary.LinkedWishlist != 1 {
		t.Fatalf("assignment did not update media summary: %+v", assetsAfter.Summary)
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
