package media

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/collectors-tech/cabinet/internal/db"
)

func TestSaveCompanionAssetUsesCanonicalStorageAndContentAddressedLinks(t *testing.T) {
	t.Parallel()
	base := t.TempDir()
	conn, err := db.OpenAndMigrate(context.Background(), filepath.Join(base, "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	mediaRoot := filepath.Join(base, "relocatable-media")
	if _, err := conn.Exec(`
		INSERT INTO profiles(id, name) VALUES ('profile-1', 'One');
		INSERT INTO profile_settings(profile_id, key, value) VALUES ('profile-1', 'storage.media_dir', ?);
		INSERT INTO companion_captures(id, profile_id, session_id, module_id, module_version, schema_version, provider_id,
			integration_instance_id, payload_type, source_url, captured_at, page_complete, payload_hash, idempotency_key,
			redaction_summary_json, raw_payload_json, state, created_at, updated_at)
		VALUES
			('capture-1','profile-1','session-1','frontline-search-results','1.0.0','1','frontline','frontline-1','search_results','https://frontlinehobbies.com.au/search','2026-08-06T00:00:00Z',1,'sha256:capture','capture-1','[]','{}','completed','2026-08-06T00:00:00Z','2026-08-06T00:00:00Z'),
			('capture-2','profile-1','session-1','frontline-search-results','1.0.0','1','frontline','frontline-1','item_detail','https://frontlinehobbies.com.au/item','2026-08-06T00:00:00Z',1,'sha256:capture2','capture-2','[]','{}','completed','2026-08-06T00:00:00Z','2026-08-06T00:00:00Z');
	`, mediaRoot); err != nil {
		t.Fatalf("seed companion media records: %v", err)
	}

	body := sampleJPEG(t)
	digest := sha256.Sum256(body)
	hash := hex.EncodeToString(digest[:])
	svc := NewService(conn, filepath.Join(base, "default-media"))
	input := CompanionAssetInput{
		ProfileID: "profile-1", CaptureID: "capture-1", FieldName: "items[0].image_url", Filename: "product.jpeg",
		IdempotencyKey: "media-capture-1",
		MIMEType:       "image/jpeg", ContentHash: hash, SourceURL: "https://frontlinehobbies.com.au/item",
		Provenance: map[string]string{"provider_id": "frontline", "module_id": "frontline-search-results"},
	}
	created, err := svc.SaveCompanionAsset(context.Background(), input, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("SaveCompanionAsset() error = %v", err)
	}
	if !created.Committed || created.Deduplicated || created.ID == "" || created.ContentHash != hash || created.ByteSize != int64(len(body)) {
		t.Fatalf("unexpected created asset: %+v", created)
	}
	assetDir := filepath.Join(mediaRoot, "assets", created.ID)
	manifestBytes, err := os.ReadFile(filepath.Join(assetDir, "manifest.json"))
	if err != nil {
		t.Fatalf("read canonical manifest: %v", err)
	}
	var manifest AssetManifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatalf("decode canonical manifest: %v", err)
	}
	if manifest.Original.ContentHash != "sha256:"+hash || manifest.Owners[0].Type != "companion_capture" || manifest.Owners[0].ID != "capture-1" {
		t.Fatalf("unexpected companion manifest: %+v", manifest)
	}
	if _, err := os.Stat(filepath.Join(assetDir, filepath.FromSlash(manifest.Original.RelativePath))); err != nil {
		t.Fatalf("canonical original missing: %v", err)
	}

	input.CaptureID = "capture-2"
	input.FieldName = "image_url"
	input.IdempotencyKey = "media-capture-2"
	replayed, err := svc.SaveCompanionAsset(context.Background(), input, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("SaveCompanionAsset(replay) error = %v", err)
	}
	if !replayed.Committed || !replayed.Deduplicated || replayed.ID != created.ID {
		t.Fatalf("unexpected deduplicated asset: %+v", replayed)
	}
	var assetCount, linkCount int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM companion_media_assets WHERE profile_id = 'profile-1'`).Scan(&assetCount); err != nil {
		t.Fatalf("count companion assets: %v", err)
	}
	if err := conn.QueryRow(`SELECT COUNT(*) FROM companion_media_links WHERE asset_id = ?`, created.ID).Scan(&linkCount); err != nil {
		t.Fatalf("count companion links: %v", err)
	}
	if assetCount != 1 || linkCount != 2 {
		t.Fatalf("content-addressed persistence counts assets=%d links=%d", assetCount, linkCount)
	}
}

func TestSaveCompanionAssetRejectsFalseImageBeforeWriting(t *testing.T) {
	t.Parallel()
	base := t.TempDir()
	conn, err := db.OpenAndMigrate(context.Background(), filepath.Join(base, "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if _, err := conn.Exec(`
		INSERT INTO profiles(id, name) VALUES ('profile-1', 'One');
		INSERT INTO companion_captures(id, profile_id, session_id, module_id, module_version, schema_version, provider_id,
			integration_instance_id, payload_type, source_url, captured_at, payload_hash, idempotency_key,
			redaction_summary_json, raw_payload_json, state, created_at, updated_at)
		VALUES ('capture-1','profile-1','session-1','module','1','1','provider','instance','item_detail','https://example.test/item','2026-08-06T00:00:00Z','sha256:capture','capture-1','[]','{}','completed','2026-08-06T00:00:00Z','2026-08-06T00:00:00Z')
	`); err != nil {
		t.Fatalf("seed capture: %v", err)
	}
	body := []byte("<html><title>Sign in</title></html>")
	digest := sha256.Sum256(body)
	_, err = NewService(conn, filepath.Join(base, "media")).SaveCompanionAsset(context.Background(), CompanionAssetInput{
		ProfileID: "profile-1", CaptureID: "capture-1", FieldName: "image_url", Filename: "challenge.jpg",
		IdempotencyKey: "false-image",
		MIMEType:       "image/jpeg", ContentHash: hex.EncodeToString(digest[:]),
	}, bytes.NewReader(body))
	if err == nil {
		t.Fatal("SaveCompanionAsset() accepted HTML as an image")
	}
	if _, statErr := os.Stat(filepath.Join(base, "media", "assets")); !os.IsNotExist(statErr) {
		t.Fatalf("false image created canonical storage: %v", statErr)
	}
}
