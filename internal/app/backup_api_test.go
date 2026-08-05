package app

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/collectors-tech/cabinet/internal/media"
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

func TestBackupRestoreRelocatesCompanionCapturesAndCanonicalMedia(t *testing.T) {
	t.Setenv("CABINET_SEED_SAMPLE_DATA", "0")
	sourceApp := newTestApp(t)
	profileID := prepareCompanionAPIProfile(t, sourceApp)
	if _, err := sourceApp.db.ExecContext(context.Background(), `
		INSERT INTO companion_captures(id, profile_id, session_id, module_id, module_version, schema_version, provider_id,
			integration_instance_id, payload_type, source_url, captured_at, page_complete, payload_hash, idempotency_key,
			redaction_summary_json, raw_payload_json, state, checkpoint_json, created_at, updated_at)
		VALUES ('companion-backup-capture',?,'session','ebay-purchase-capture','1.0.0','1','ebay','instance','purchase_order',
			'https://www.ebay.com/mye/myebay/purchase','2026-08-06T00:00:00Z',1,'sha256:capture','companion-backup-capture',
			'["no_cookies","no_raw_page","no_tokens"]','{"redacted":true}','review','{"records_committed":1}',
			'2026-08-06T00:00:00Z','2026-08-06T00:00:00Z');
		INSERT INTO companion_purchase_inbox(id, capture_id, profile_id, provider_id, order_key, item_key, card_json, state,
			first_seen, last_seen, created_at, updated_at)
		VALUES ('companion-backup-purchase','companion-backup-capture',?,'ebay','ORDER-BACKUP','TX-BACKUP',
			'{"transaction_id":"TX-BACKUP","listing_title":"Backup purchase","quantity":1,"item_price":"AU $5.00"}',
			'review','2026-08-06T00:00:00Z','2026-08-06T00:00:00Z','2026-08-06T00:00:00Z','2026-08-06T00:00:00Z')
	`, profileID, profileID); err != nil {
		t.Fatalf("seed companion backup records: %v", err)
	}
	imageBytes := sampleJPEG(t)
	digest := sha256.Sum256(imageBytes)
	asset, err := media.NewService(sourceApp.db, filepath.Join(sourceApp.cfg.DataDir, "media")).SaveCompanionAsset(context.Background(), media.CompanionAssetInput{
		ProfileID: profileID, CaptureID: "companion-backup-capture", FieldName: "cards[0].image_url", Filename: "backup.jpg",
		IdempotencyKey: "companion-backup-media",
		MIMEType:       "image/jpeg", ContentHash: hex.EncodeToString(digest[:]), SourceURL: "https://www.ebay.com/itm/backup",
		Provenance: map[string]string{"module_id": "ebay-purchase-capture", "provider_id": "ebay"},
	}, bytes.NewReader(imageBytes))
	if err != nil {
		t.Fatalf("save companion backup asset: %v", err)
	}
	profileExport := doRequest(t, sourceApp, http.MethodGet, "/api/data/export/json", nil, nil)
	if profileExport.Code != http.StatusOK {
		t.Fatalf("companion-safe profile export status=%d body=%s", profileExport.Code, profileExport.Body.String())
	}
	if strings.Contains(profileExport.Body.String(), "companion-backup-capture") || strings.Contains(profileExport.Body.String(), `"redacted":true`) || strings.Contains(profileExport.Body.String(), "credential_verifier") {
		t.Fatalf("profile export leaked companion queue, raw payload or credential metadata: %s", profileExport.Body.String())
	}

	runResp := doRequest(t, sourceApp, http.MethodPost, "/api/backup/run", nil, nil)
	if runResp.Code != http.StatusOK {
		t.Fatalf("companion backup run status=%d body=%s", runResp.Code, runResp.Body.String())
	}
	var runPayload struct {
		Backup struct {
			Path string `json:"path"`
		} `json:"backup"`
	}
	if err := json.NewDecoder(runResp.Body).Decode(&runPayload); err != nil {
		t.Fatalf("decode companion backup: %v", err)
	}
	archiveEntryBytes(t, runPayload.Backup.Path, filepath.ToSlash(filepath.Join("media", "assets", asset.ID, "manifest.json")))
	archiveEntryBytes(t, runPayload.Backup.Path, filepath.ToSlash(filepath.Join("media", "assets", asset.ID, "original", "backup.jpg")))

	targetApp := newTestApp(t)
	targetBackupPath := filepath.Join(targetApp.cfg.DataDir, "backups", filepath.Base(runPayload.Backup.Path))
	if err := os.MkdirAll(filepath.Dir(targetBackupPath), 0o755); err != nil {
		t.Fatalf("create companion relocated backup dir: %v", err)
	}
	copyFileForTest(t, runPayload.Backup.Path, targetBackupPath)
	restoreBody := bytes.NewBufferString(`{"backup_path":"` + strings.ReplaceAll(targetBackupPath, `\`, `\\`) + `","confirm_restore":true}`)
	restoreResp := doRequest(t, targetApp, http.MethodPost, "/api/backup/restore", restoreBody, map[string]string{"Content-Type": "application/json"})
	if restoreResp.Code != http.StatusOK {
		t.Fatalf("companion relocated restore status=%d body=%s", restoreResp.Code, restoreResp.Body.String())
	}
	var captures, purchases, assets, links int
	for _, check := range []struct {
		query string
		dest  *int
	}{
		{`SELECT COUNT(*) FROM companion_captures WHERE id = 'companion-backup-capture'`, &captures},
		{`SELECT COUNT(*) FROM companion_purchase_inbox WHERE id = 'companion-backup-purchase'`, &purchases},
		{`SELECT COUNT(*) FROM companion_media_assets WHERE id = '` + asset.ID + `'`, &assets},
		{`SELECT COUNT(*) FROM companion_media_links WHERE asset_id = '` + asset.ID + `'`, &links},
	} {
		if err := targetApp.db.QueryRow(check.query).Scan(check.dest); err != nil {
			t.Fatalf("verify restored companion record: %v", err)
		}
	}
	if captures != 1 || purchases != 1 || assets != 1 || links != 1 {
		t.Fatalf("restored companion counts captures=%d purchases=%d assets=%d links=%d", captures, purchases, assets, links)
	}
	restoredOriginal, err := os.ReadFile(filepath.Join(targetApp.cfg.DataDir, "media", "assets", asset.ID, "original", "backup.jpg"))
	if err != nil || !bytes.Equal(restoredOriginal, imageBytes) {
		t.Fatalf("relocated companion original mismatch: %v", err)
	}
	if strings.Contains(filepath.Join("media", "assets", asset.ID, "original", "backup.jpg"), sourceApp.cfg.DataDir) {
		t.Fatal("companion media record retained the source data path")
	}
}

func TestBackupRestoreRelocatesMigratedLegacyMediaAssets(t *testing.T) {
	t.Setenv("CABINET_SEED_SAMPLE_DATA", "0")

	sourceApp := newTestApp(t)
	createProfile := doRequest(t, sourceApp, http.MethodPost, "/api/profiles", strings.NewReader(`{"name":"Migration Backup Source"}`), map[string]string{"Content-Type": "application/json"})
	if createProfile.Code != http.StatusCreated {
		t.Fatalf("create source profile status=%d body=%s", createProfile.Code, createProfile.Body.String())
	}
	var sourceProfile struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(createProfile.Body).Decode(&sourceProfile); err != nil {
		t.Fatalf("decode source profile: %v", err)
	}
	activateProfile := doRequest(t, sourceApp, http.MethodPut, "/api/profiles/active", strings.NewReader(`{"profile_id":"`+sourceProfile.ID+`"}`), map[string]string{"Content-Type": "application/json"})
	if activateProfile.Code != http.StatusOK {
		t.Fatalf("activate source profile status=%d body=%s", activateProfile.Code, activateProfile.Body.String())
	}
	sourceMediaRoot := filepath.Join(sourceApp.cfg.DataDir, "media")
	if _, err := sourceApp.db.Exec(`
		INSERT INTO profile_settings(profile_id, key, value)
		VALUES (?, 'storage.media_dir', ?)
		ON CONFLICT(profile_id, key) DO UPDATE SET value = excluded.value
	`, sourceProfile.ID, sourceMediaRoot); err != nil {
		t.Fatalf("set source media root: %v", err)
	}
	legacyDir := filepath.Join(sourceMediaRoot, "legacy-item")
	if err := os.MkdirAll(legacyDir, 0o755); err != nil {
		t.Fatalf("create legacy media dir: %v", err)
	}
	legacyOriginal := filepath.Join(legacyDir, "front_orig.jpg")
	legacyBytes := sampleJPEG(t)
	if err := os.WriteFile(legacyOriginal, legacyBytes, 0o644); err != nil {
		t.Fatalf("write legacy media: %v", err)
	}
	if _, err := sourceApp.db.Exec(`
		INSERT INTO canonical_items (id, profile_id, brand, category, part_number, title)
		VALUES ('legacy-item', ?,'AFX','Slot Car','MIG-BACKUP','Migrated Backup');
		INSERT INTO item_photos (id, item_id, filename, original_path, preview_path, thumbnail_path, is_primary, display_order)
		VALUES ('legacy-photo','legacy-item','front.jpg','legacy-item/front_orig.jpg','','',1,1);
	`, sourceProfile.ID); err != nil {
		t.Fatalf("seed legacy media row: %v", err)
	}

	mediaSvc := media.NewService(sourceApp.db, sourceMediaRoot)
	evidence, err := mediaSvc.ApplyLegacyMediaMigration(t.Context(), sourceProfile.ID)
	if err != nil {
		t.Fatalf("ApplyLegacyMediaMigration() error: %v", err)
	}
	if evidence.Summary.Migrated != 1 || evidence.Summary.Failed != 0 || evidence.Summary.Skipped != 0 {
		t.Fatalf("unexpected migration evidence before backup: %+v", evidence.Summary)
	}

	runResp := doRequest(t, sourceApp, http.MethodPost, "/api/backup/run", nil, nil)
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

	targetApp := newTestApp(t)
	targetBackupPath := filepath.Join(targetApp.cfg.DataDir, "backups", filepath.Base(runPayload.Backup.Path))
	if err := os.MkdirAll(filepath.Dir(targetBackupPath), 0o755); err != nil {
		t.Fatalf("create relocated backup dir: %v", err)
	}
	copyFileForTest(t, runPayload.Backup.Path, targetBackupPath)
	restoreBody := bytes.NewBufferString(`{"backup_path":"` + strings.ReplaceAll(targetBackupPath, `\`, `\\`) + `","confirm_restore":true}`)
	restoreResp := doRequest(t, targetApp, http.MethodPost, "/api/backup/restore", restoreBody, map[string]string{"Content-Type": "application/json"})
	if restoreResp.Code != http.StatusOK {
		t.Fatalf("relocated backup restore status=%d body=%s", restoreResp.Code, restoreResp.Body.String())
	}

	var originalPath, previewPath, thumbnailPath string
	if err := targetApp.db.QueryRow(`
		SELECT original_path, preview_path, thumbnail_path
		FROM item_photos
		WHERE id = 'legacy-photo'
	`).Scan(&originalPath, &previewPath, &thumbnailPath); err != nil {
		t.Fatalf("read relocated migrated photo row: %v", err)
	}
	expectedOriginal := filepath.ToSlash(filepath.Join("assets", "legacy-photo", "original", "front.jpg"))
	expectedPreview := filepath.ToSlash(filepath.Join("assets", "legacy-photo", "renditions", "preview.jpg"))
	expectedThumbnail := filepath.ToSlash(filepath.Join("assets", "legacy-photo", "renditions", "thumbnail.jpg"))
	if originalPath != expectedOriginal || previewPath != expectedPreview || thumbnailPath != expectedThumbnail {
		t.Fatalf("restored migrated paths should stay media-root-relative, got original=%q preview=%q thumbnail=%q", originalPath, previewPath, thumbnailPath)
	}
	restoredOriginal := filepath.Join(targetApp.cfg.DataDir, "media", filepath.FromSlash(originalPath))
	restoredBytes, err := os.ReadFile(restoredOriginal)
	if err != nil {
		t.Fatalf("read relocated restored original: %v", err)
	}
	if !bytes.Equal(restoredBytes, legacyBytes) {
		t.Fatalf("relocated restore did not preserve migrated original bytes")
	}
	if !strings.HasPrefix(restoredOriginal, filepath.Join(targetApp.cfg.DataDir, "media")) || strings.Contains(originalPath, sourceApp.cfg.DataDir) {
		t.Fatalf("restored path did not relocate cleanly: db=%q resolved=%q sourceData=%q", originalPath, restoredOriginal, sourceApp.cfg.DataDir)
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

func copyFileForTest(t *testing.T, src, dst string) {
	t.Helper()
	in, err := os.Open(src)
	if err != nil {
		t.Fatalf("open source file for copy: %v", err)
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		t.Fatalf("create destination file for copy: %v", err)
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		t.Fatalf("copy file: %v", err)
	}
	if err := out.Close(); err != nil {
		t.Fatalf("close destination file: %v", err)
	}
}
