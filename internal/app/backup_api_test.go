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

func TestBackupRunAndRestoreEndpoints(t *testing.T) {
	t.Setenv("CABINET_SEED_SAMPLE_DATA", "0")

	a := newTestApp(t)
	if _, err := a.db.Exec(`INSERT INTO canonical_items (id, brand, category, part_number, title) VALUES ('item-a','AFX','Slot Car','P-1','Car')`); err != nil {
		t.Fatalf("seed item: %v", err)
	}

	runResp := doRequest(t, a, http.MethodPost, "/api/backup/run", nil, nil)
	if runResp.Code != http.StatusOK {
		t.Fatalf("backup run status = %d body=%s", runResp.Code, runResp.Body.String())
	}
	var runPayload struct {
		Backup struct {
			Path           string `json:"path"`
			FileName       string `json:"file_name"`
			SizeBytes      int64  `json:"size_bytes"`
			CreatedAt      string `json:"created_at"`
			ArchiveFormat  string `json:"archive_format"`
			DownloadURL    string `json:"download_url"`
			IntegrityCheck string `json:"integrity_check"`
		} `json:"backup"`
	}
	if err := json.NewDecoder(runResp.Body).Decode(&runPayload); err != nil {
		t.Fatalf("decode backup run response: %v", err)
	}
	if strings.TrimSpace(runPayload.Backup.Path) == "" || runPayload.Backup.FileName == "" || runPayload.Backup.SizeBytes == 0 || runPayload.Backup.IntegrityCheck != "ok" {
		t.Fatalf("expected backup metadata in response, got %+v", runPayload.Backup)
	}
	if !strings.HasPrefix(runPayload.Backup.FileName, "cabinet-backup-") || !strings.HasSuffix(runPayload.Backup.FileName, ".zip") || runPayload.Backup.ArchiveFormat != "zip" || runPayload.Backup.DownloadURL == "" {
		t.Fatalf("expected timestamped zip backup metadata, got %+v", runPayload.Backup)
	}

	downloadResp := doRequest(t, a, http.MethodGet, "/api/backup/download?file_name="+runPayload.Backup.FileName, nil, nil)
	if downloadResp.Code != http.StatusOK {
		t.Fatalf("backup download status = %d body=%s", downloadResp.Code, downloadResp.Body.String())
	}
	if ct := downloadResp.Header().Get("Content-Type"); !strings.Contains(ct, "application/zip") {
		t.Fatalf("expected zip content type, got %q", ct)
	}
	if cd := downloadResp.Header().Get("Content-Disposition"); !strings.Contains(cd, runPayload.Backup.FileName) {
		t.Fatalf("expected attachment filename in content disposition, got %q", cd)
	}

	listResp := doRequest(t, a, http.MethodGet, "/api/backup/list", nil, nil)
	if listResp.Code != http.StatusOK {
		t.Fatalf("backup list status = %d body=%s", listResp.Code, listResp.Body.String())
	}
	var listPayload struct {
		Backups []struct {
			Path      string `json:"path"`
			FileName  string `json:"file_name"`
			SizeBytes int64  `json:"size_bytes"`
			CreatedAt string `json:"created_at"`
		} `json:"backups"`
	}
	if err := json.NewDecoder(listResp.Body).Decode(&listPayload); err != nil {
		t.Fatalf("decode backup list response: %v", err)
	}
	if len(listPayload.Backups) != 1 || listPayload.Backups[0].Path != runPayload.Backup.Path || listPayload.Backups[0].SizeBytes == 0 {
		t.Fatalf("expected listed backup metadata, got %+v", listPayload.Backups)
	}

	if _, err := a.db.Exec(`INSERT INTO canonical_items (id, brand, category, part_number, title) VALUES ('item-b','AFX','Slot Car','P-2','Car')`); err != nil {
		t.Fatalf("seed second item: %v", err)
	}

	unconfirmedBody := bytes.NewBufferString(`{"backup_path":"` + strings.ReplaceAll(runPayload.Backup.Path, `\`, `\\`) + `"}`)
	unconfirmedResp := doRequest(t, a, http.MethodPost, "/api/backup/restore", unconfirmedBody, map[string]string{"Content-Type": "application/json"})
	if unconfirmedResp.Code != http.StatusBadRequest {
		t.Fatalf("backup restore without confirmation status = %d body=%s", unconfirmedResp.Code, unconfirmedResp.Body.String())
	}
	if !strings.Contains(unconfirmedResp.Body.String(), "restore_confirmation_required") {
		t.Fatalf("expected confirmation error, got %s", unconfirmedResp.Body.String())
	}

	restoreBody := bytes.NewBufferString(`{"backup_path":"` + strings.ReplaceAll(runPayload.Backup.Path, `\`, `\\`) + `","confirm_restore":true}`)
	restoreResp := doRequest(t, a, http.MethodPost, "/api/backup/restore", restoreBody, map[string]string{"Content-Type": "application/json"})
	if restoreResp.Code != http.StatusOK {
		t.Fatalf("backup restore status = %d body=%s", restoreResp.Code, restoreResp.Body.String())
	}
	var restorePayload struct {
		Restore struct {
			RestoredPath          string `json:"restored_path"`
			RestoredAt            string `json:"restored_at"`
			IntegrityCheck        string `json:"integrity_check"`
			PreRestoreBackupTaken bool   `json:"pre_restore_backup_taken"`
			PreRestoreBackup      struct {
				FileName      string `json:"file_name"`
				ArchiveFormat string `json:"archive_format"`
				DownloadURL   string `json:"download_url"`
			} `json:"pre_restore_backup"`
		} `json:"restore"`
	}
	if err := json.NewDecoder(restoreResp.Body).Decode(&restorePayload); err != nil {
		t.Fatalf("decode backup restore response: %v", err)
	}
	if restorePayload.Restore.RestoredPath != runPayload.Backup.Path || restorePayload.Restore.IntegrityCheck != "ok" || restorePayload.Restore.RestoredAt == "" {
		t.Fatalf("expected restore metadata, got %+v", restorePayload.Restore)
	}
	if !restorePayload.Restore.PreRestoreBackupTaken || restorePayload.Restore.PreRestoreBackup.FileName == "" || restorePayload.Restore.PreRestoreBackup.FileName == runPayload.Backup.FileName || restorePayload.Restore.PreRestoreBackup.ArchiveFormat != "zip" || restorePayload.Restore.PreRestoreBackup.DownloadURL == "" {
		t.Fatalf("expected distinct pre-restore backup metadata, got %+v", restorePayload.Restore)
	}

	itemsResp := doRequest(t, a, http.MethodGet, "/api/items", nil, nil)
	if itemsResp.Code != http.StatusOK {
		t.Fatalf("items status = %d body=%s", itemsResp.Code, itemsResp.Body.String())
	}
	var itemsPayload struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.NewDecoder(itemsResp.Body).Decode(&itemsPayload); err != nil {
		t.Fatalf("decode items response: %v", err)
	}
	if len(itemsPayload.Items) != 1 {
		t.Fatalf("expected 1 item after restore, got %d", len(itemsPayload.Items))
	}
}

func TestBackupRunAndRestorePreservesCanonicalMediaAssets(t *testing.T) {
	t.Setenv("CABINET_SEED_SAMPLE_DATA", "0")

	a := newTestApp(t)
	if _, err := a.db.Exec(`INSERT INTO canonical_items (id, brand, category, part_number, title) VALUES ('item-media','AFX','Slot Car','MEDIA-BACKUP','Media Backup')`); err != nil {
		t.Fatalf("seed item: %v", err)
	}
	body, contentType := buildMultipartPhoto(t, "front.jpg", sampleJPEG(t))
	uploadResp := doRequest(t, a, http.MethodPost, "/api/items/item-media/photos", body, map[string]string{"Content-Type": contentType})
	if uploadResp.Code != http.StatusCreated {
		t.Fatalf("upload photo status=%d body=%s", uploadResp.Code, uploadResp.Body.String())
	}
	var uploaded struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(uploadResp.Body).Decode(&uploaded); err != nil {
		t.Fatalf("decode upload: %v", err)
	}

	runResp := doRequest(t, a, http.MethodPost, "/api/backup/run", nil, nil)
	if runResp.Code != http.StatusOK {
		t.Fatalf("backup run status=%d body=%s", runResp.Code, runResp.Body.String())
	}
	var runPayload struct {
		Backup struct {
			Path string `json:"path"`
		} `json:"backup"`
	}
	if err := json.NewDecoder(runResp.Body).Decode(&runPayload); err != nil {
		t.Fatalf("decode backup run: %v", err)
	}
	manifestEntry := filepath.ToSlash(filepath.Join("media", "assets", uploaded.ID, "manifest.json"))
	originalEntry := filepath.ToSlash(filepath.Join("media", "assets", uploaded.ID, "original", "front.jpg"))
	manifestBefore := archiveEntryBytes(t, runPayload.Backup.Path, manifestEntry)
	originalBefore := archiveEntryBytes(t, runPayload.Backup.Path, originalEntry)
	if len(manifestBefore) == 0 || !strings.Contains(string(manifestBefore), `"asset_id": "`+uploaded.ID+`"`) || len(originalBefore) == 0 {
		t.Fatalf("backup archive did not preserve canonical media asset manifest/original")
	}

	assetDir := filepath.Join(a.cfg.DataDir, "media", "assets", uploaded.ID)
	if err := os.RemoveAll(assetDir); err != nil {
		t.Fatalf("remove media asset before restore: %v", err)
	}
	restoreBody := bytes.NewBufferString(`{"backup_path":"` + strings.ReplaceAll(runPayload.Backup.Path, `\`, `\\`) + `","confirm_restore":true}`)
	restoreResp := doRequest(t, a, http.MethodPost, "/api/backup/restore", restoreBody, map[string]string{"Content-Type": "application/json"})
	if restoreResp.Code != http.StatusOK {
		t.Fatalf("backup restore status=%d body=%s", restoreResp.Code, restoreResp.Body.String())
	}
	restoredManifest, err := os.ReadFile(filepath.Join(assetDir, "manifest.json"))
	if err != nil {
		t.Fatalf("read restored manifest: %v", err)
	}
	restoredOriginal, err := os.ReadFile(filepath.Join(assetDir, "original", "front.jpg"))
	if err != nil {
		t.Fatalf("read restored original: %v", err)
	}
	if !bytes.Equal(restoredManifest, manifestBefore) || !bytes.Equal(restoredOriginal, originalBefore) {
		t.Fatalf("restore did not round-trip canonical media manifest/original bytes")
	}
}

func archiveEntryBytes(t *testing.T, path, name string) []byte {
	t.Helper()
	zr, err := zip.OpenReader(path)
	if err != nil {
		t.Fatalf("open backup archive: %v", err)
	}
	defer zr.Close()
	for _, file := range zr.File {
		if file.Name != name {
			continue
		}
		rc, err := file.Open()
		if err != nil {
			t.Fatalf("open archive entry %s: %v", name, err)
		}
		data, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			t.Fatalf("read archive entry %s: %v", name, err)
		}
		return data
	}
	t.Fatalf("backup archive missing %s", name)
	return nil
}
