package media

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/collectors-tech/cabinet/internal/db"
)

func TestListWorkspaceAssetsScopesInventoryAndUnlinkedMediaByProfile(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	dbPath := filepath.Join(base, "cabinet.db")
	conn, err := db.OpenAndMigrate(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	if _, err := conn.Exec(`
		INSERT INTO profiles (id, name) VALUES ('profile-1','One'), ('profile-2','Two');
		INSERT INTO canonical_items (id, profile_id, brand, category, part_number, title) VALUES
			('item-1','profile-1','AFX','Slot Car','AFX-1','Mustang Front'),
			('item-2','profile-2','AFX','Slot Car','AFX-2','Other Profile');
		INSERT INTO chat_threads (id, profile_id, title) VALUES ('thread-1','profile-1','Media'), ('thread-2','profile-2','Media');
		INSERT INTO chat_attachments (id, profile_id, thread_id, filename, mime_type, size_bytes, stored_path) VALUES
			('attach-1','profile-1','thread-1','loose-reference.jpg','image/jpeg',123,'/tmp/loose-reference.jpg'),
			('attach-2','profile-2','thread-2','other-profile.jpg','image/jpeg',123,'/tmp/other-profile.jpg');
	`); err != nil {
		t.Fatalf("seed workspace data: %v", err)
	}

	svc := NewService(conn, filepath.Join(base, "media"))
	photo, err := svc.Upload(context.Background(), "item-1", "front.jpg", bytes.NewReader(sampleJPEG(t)))
	if err != nil {
		t.Fatalf("Upload() profile one photo error = %v", err)
	}
	if _, err := svc.Upload(context.Background(), "item-2", "other.jpg", bytes.NewReader(sampleJPEG(t))); err != nil {
		t.Fatalf("Upload() profile two photo error = %v", err)
	}

	list, err := svc.ListWorkspaceAssets(context.Background(), "profile-1", "all")
	if err != nil {
		t.Fatalf("ListWorkspaceAssets() error = %v", err)
	}
	if len(list.Assets) != 2 || list.Summary.Total != 2 || list.Summary.Unlinked != 1 || list.Summary.LinkedInventory != 1 {
		t.Fatalf("unexpected scoped summary/list: %+v", list)
	}
	if !containsAssetState(list.Assets, photo.ID, "linked_inventory") || !containsAssetState(list.Assets, "attach-1", "unlinked") {
		t.Fatalf("expected profile one inventory and unlinked assets, got %+v", list.Assets)
	}
	if containsAssetState(list.Assets, "attach-2", "unlinked") {
		t.Fatalf("profile two asset leaked into profile one list: %+v", list.Assets)
	}

	unlinked, err := svc.ListWorkspaceAssets(context.Background(), "profile-1", "unlinked")
	if err != nil {
		t.Fatalf("ListWorkspaceAssets(unlinked) error = %v", err)
	}
	if len(unlinked.Assets) != 1 || unlinked.Assets[0].ID != "attach-1" || unlinked.Summary.Total != 2 {
		t.Fatalf("unexpected unlinked filter response: %+v", unlinked)
	}
}

func TestPreviewAssignmentAndDownloadExposeConfirmableMediaWorkspaceActions(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	dbPath := filepath.Join(base, "cabinet.db")
	conn, err := db.OpenAndMigrate(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	if _, err := conn.Exec(`
		INSERT INTO profiles (id, name) VALUES ('profile-1','One');
		INSERT INTO canonical_items (id, profile_id, brand, category, part_number, title) VALUES
			('item-1','profile-1','AFX','Slot Car','AFX-1','Mustang Front');
		INSERT INTO wishlist_entries (id, profile_id, item_id) VALUES ('wish-1','profile-1','item-1');
		INSERT INTO chat_threads (id, profile_id, title) VALUES ('thread-1','profile-1','Media');
		INSERT INTO chat_attachments (id, profile_id, thread_id, filename, mime_type, size_bytes, stored_path) VALUES
			('attach-1','profile-1','thread-1','loose-reference.jpg','image/jpeg',123,'/tmp/loose-reference.jpg');
	`); err != nil {
		t.Fatalf("seed preview data: %v", err)
	}

	svc := NewService(conn, filepath.Join(base, "media"))
	preview, err := svc.PreviewAssignment(context.Background(), "profile-1", "attach-1", "wishlist", "wish-1")
	if err != nil {
		t.Fatalf("PreviewAssignment() error = %v", err)
	}
	if !preview.Allowed || !preview.RequiresConfirmation || preview.CurrentLinkageState != "unlinked" || preview.ProjectedLinkageState != "linked_wishlist" {
		t.Fatalf("assignment preview should allow confirmed wishlist linkage, got %+v", preview)
	}
	if preview.BlockedReason != "" || preview.AuditSummary == "" {
		t.Fatalf("assignment preview must explain audit summary without blocker, got %+v", preview)
	}

	download, err := svc.PreviewDownload(context.Background(), "profile-1", []string{"attach-1"}, "all")
	if err != nil {
		t.Fatalf("PreviewDownload() error = %v", err)
	}
	if !download.Allowed || download.Count != 1 || download.AssetIDs[0] != "attach-1" || download.Filenames[0] != "loose-reference-jpg-attach-1.jpg" {
		t.Fatalf("unexpected download preview: %+v", download)
	}
}

func TestApplyAssignmentPersistsMediaLinksWithoutDuplicatingAssets(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	dbPath := filepath.Join(base, "cabinet.db")
	conn, err := db.OpenAndMigrate(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	if _, err := conn.Exec(`
		INSERT INTO profiles (id, name) VALUES ('profile-1','One'), ('profile-2','Two');
		INSERT INTO canonical_items (id, profile_id, brand, category, part_number, title) VALUES
			('item-1','profile-1','AFX','Slot Car','AFX-1','Mustang Front'),
			('item-2','profile-2','AFX','Slot Car','AFX-2','Other Profile');
		INSERT INTO wishlist_entries (id, profile_id, item_id) VALUES ('wish-1','profile-1','item-1'), ('wish-2','profile-2','item-2');
		INSERT INTO chat_threads (id, profile_id, title) VALUES ('thread-1','profile-1','Media'), ('thread-2','profile-2','Media');
		INSERT INTO chat_attachments (id, profile_id, thread_id, filename, mime_type, size_bytes, stored_path) VALUES
			('attach-1','profile-1','thread-1','loose-reference.jpg','image/jpeg',123,'/tmp/loose-reference.jpg'),
			('attach-2','profile-2','thread-2','other-profile.jpg','image/jpeg',123,'/tmp/other-profile.jpg');
	`); err != nil {
		t.Fatalf("seed assignment data: %v", err)
	}

	svc := NewService(conn, filepath.Join(base, "media"))
	result, err := svc.ApplyAssignment(context.Background(), "profile-1", "attach-1", "wishlist", "wish-1")
	if err != nil {
		t.Fatalf("ApplyAssignment() error = %v", err)
	}
	if !result.Applied || !result.Allowed || result.CurrentLinkageState != "linked_wishlist" || result.ProjectedLinkageState != "linked_wishlist" {
		t.Fatalf("unexpected assignment result: %+v", result)
	}

	list, err := svc.ListWorkspaceAssets(context.Background(), "profile-1", "all")
	if err != nil {
		t.Fatalf("ListWorkspaceAssets() after assignment error = %v", err)
	}
	if !containsAssetTarget(list.Assets, "attach-1", "linked_wishlist", "", "wish-1") {
		t.Fatalf("expected assigned attachment to show wishlist linkage, got %+v", list.Assets)
	}
	if list.Summary.Unlinked != 0 || list.Summary.LinkedWishlist != 1 {
		t.Fatalf("expected assignment to update summary counts, got %+v", list.Summary)
	}

	var linkCount, attachmentCount int
	if err := conn.QueryRow(`SELECT COUNT(1) FROM media_asset_links WHERE profile_id = 'profile-1' AND asset_id = 'attach-1' AND target_type = 'wishlist' AND target_id = 'wish-1'`).Scan(&linkCount); err != nil {
		t.Fatalf("count media links: %v", err)
	}
	if err := conn.QueryRow(`SELECT COUNT(1) FROM chat_attachments WHERE id = 'attach-1'`).Scan(&attachmentCount); err != nil {
		t.Fatalf("count chat attachments: %v", err)
	}
	if linkCount != 1 || attachmentCount != 1 {
		t.Fatalf("expected one link and original attachment preserved, links=%d attachments=%d", linkCount, attachmentCount)
	}

	if _, err := svc.ApplyAssignment(context.Background(), "profile-1", "attach-2", "wishlist", "wish-1"); err == nil {
		t.Fatal("ApplyAssignment() should reject another profile's asset")
	}
	if _, err := svc.ApplyAssignment(context.Background(), "profile-1", "attach-1", "wishlist", "wish-2"); err == nil {
		t.Fatal("ApplyAssignment() should reject another profile's target")
	}
}

func TestBuildDownloadReturnsScopedSingleAssetPayload(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	dbPath := filepath.Join(base, "cabinet.db")
	conn, err := db.OpenAndMigrate(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	attachmentPath := filepath.Join(base, "loose-reference.jpg")
	attachmentBytes := []byte("cabinet loose reference bytes")
	if err := os.WriteFile(attachmentPath, attachmentBytes, 0o644); err != nil {
		t.Fatalf("write attachment fixture: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO profiles (id, name) VALUES ('profile-1','One'), ('profile-2','Two');
		INSERT INTO chat_threads (id, profile_id, title) VALUES ('thread-1','profile-1','Media'), ('thread-2','profile-2','Other');
	`); err != nil {
		t.Fatalf("seed profile/thread data: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO chat_attachments (id, profile_id, thread_id, filename, mime_type, size_bytes, stored_path) VALUES
			('attach-1','profile-1','thread-1','loose-reference.jpg','image/jpeg',123,?),
			('attach-2','profile-2','thread-2','other-reference.jpg','image/jpeg',123,?)
	`, attachmentPath, filepath.Join(base, "other-reference.jpg")); err != nil {
		t.Fatalf("seed attachment data: %v", err)
	}

	svc := NewService(conn, filepath.Join(base, "media"))
	bundle, err := svc.BuildDownload(context.Background(), "profile-1", []string{"attach-1"}, "all")
	if err != nil {
		t.Fatalf("BuildDownload() error = %v", err)
	}
	if bundle.Filename != "loose-reference-jpg-attach-1.jpg" || bundle.ContentType != "image/jpeg" || !bytes.Equal(bundle.Bytes, attachmentBytes) {
		t.Fatalf("unexpected single download bundle: %+v bytes=%q", bundle, string(bundle.Bytes))
	}
	if _, err := svc.BuildDownload(context.Background(), "profile-1", []string{"attach-2"}, "all"); err == nil {
		t.Fatal("BuildDownload() should reject another profile's asset")
	}
}

func TestBuildDownloadZipsMultipleSelectedAssets(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	dbPath := filepath.Join(base, "cabinet.db")
	conn, err := db.OpenAndMigrate(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	attachmentPath := filepath.Join(base, "loose-reference.jpg")
	attachmentBytes := []byte("cabinet attachment bytes")
	if err := os.WriteFile(attachmentPath, attachmentBytes, 0o644); err != nil {
		t.Fatalf("write attachment fixture: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO profiles (id, name) VALUES ('profile-1','One');
		INSERT INTO canonical_items (id, profile_id, brand, category, part_number, title) VALUES
			('item-1','profile-1','AFX','Slot Car','AFX-1','Mustang Front');
		INSERT INTO chat_threads (id, profile_id, title) VALUES ('thread-1','profile-1','Media');
	`); err != nil {
		t.Fatalf("seed profile/item/thread data: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO chat_attachments (id, profile_id, thread_id, filename, mime_type, size_bytes, stored_path) VALUES
			('attach-1','profile-1','thread-1','loose-reference.jpg','image/jpeg',123,?)
	`, attachmentPath); err != nil {
		t.Fatalf("seed workspace attachment: %v", err)
	}

	svc := NewService(conn, filepath.Join(base, "media"))
	photo, err := svc.Upload(context.Background(), "item-1", "front.jpg", bytes.NewReader(sampleJPEG(t)))
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}

	bundle, err := svc.BuildDownload(context.Background(), "profile-1", []string{photo.ID, "attach-1"}, "all")
	if err != nil {
		t.Fatalf("BuildDownload() error = %v", err)
	}
	if bundle.Filename != "cabinet-media-download.zip" || bundle.ContentType != "application/zip" || len(bundle.AssetIDs) != 2 {
		t.Fatalf("unexpected zip bundle metadata: %+v", bundle)
	}
	zr, err := zip.NewReader(bytes.NewReader(bundle.Bytes), int64(len(bundle.Bytes)))
	if err != nil {
		t.Fatalf("open zip bundle: %v", err)
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
	if _, ok := entries["afx-1-mustang-front-"+photo.ID[:8]+".jpg"]; !ok {
		t.Fatalf("inventory photo friendly filename missing from zip: %v", entries)
	}
	if !bytes.Equal(entries["loose-reference-jpg-attach-1.jpg"], attachmentBytes) {
		t.Fatalf("attachment bytes missing from zip: %v", entries)
	}
}

func TestUploadSetPrimaryDeleteAndRebuild(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	dbPath := filepath.Join(base, "cabinet.db")
	conn, err := db.OpenAndMigrate(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	if _, err := conn.Exec(`INSERT INTO canonical_items (id, brand, category, part_number, title) VALUES ('item-1','AFX','Slot Car','P-1','Car')`); err != nil {
		t.Fatalf("seed item: %v", err)
	}

	svc := NewService(conn, filepath.Join(base, "media"))
	img := sampleJPEG(t)

	p1, err := svc.Upload(context.Background(), "item-1", "a.jpg", bytes.NewReader(img))
	if err != nil {
		t.Fatalf("Upload() p1 error = %v", err)
	}
	if !p1.IsPrimary {
		t.Fatal("first upload should be primary")
	}

	p2, err := svc.Upload(context.Background(), "item-1", "b.jpg", bytes.NewReader(img))
	if err != nil {
		t.Fatalf("Upload() p2 error = %v", err)
	}
	if p2.IsPrimary {
		t.Fatal("second upload should not be primary by default")
	}

	if err := svc.SetPrimary(context.Background(), "item-1", p2.ID); err != nil {
		t.Fatalf("SetPrimary() error = %v", err)
	}
	list, err := svc.ListByItem(context.Background(), "item-1")
	if err != nil {
		t.Fatalf("ListByItem() error = %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 photos, got %d", len(list))
	}

	if err := svc.RebuildThumbnails(context.Background(), "item-1"); err != nil {
		t.Fatalf("RebuildThumbnails() error = %v", err)
	}

	if err := svc.Delete(context.Background(), "item-1", p2.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	listAfterDelete, err := svc.ListByItem(context.Background(), "item-1")
	if err != nil {
		t.Fatalf("ListByItem() after delete error = %v", err)
	}
	if len(listAfterDelete) != 1 {
		t.Fatalf("expected 1 photo after delete, got %d", len(listAfterDelete))
	}
	if !listAfterDelete[0].IsPrimary {
		t.Fatal("remaining photo should be primary")
	}
}

func TestDeletePreservesSharedCanonicalAssetAndCleansOrphan(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	dbPath := filepath.Join(base, "cabinet.db")
	conn, err := db.OpenAndMigrate(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	if _, err := conn.Exec(`
		INSERT INTO profiles (id, name) VALUES ('profile-1','One');
		INSERT INTO canonical_items (id, profile_id, brand, category, part_number, title)
		VALUES ('item-1','profile-1','AFX','Slot Car','P-1','Car');
	`); err != nil {
		t.Fatalf("seed item: %v", err)
	}

	svc := NewService(conn, filepath.Join(base, "media"))
	img := sampleJPEG(t)

	shared, err := svc.Upload(context.Background(), "item-1", "shared.jpg", bytes.NewReader(img))
	if err != nil {
		t.Fatalf("Upload() shared error = %v", err)
	}
	sharedAssetDir := filepath.Dir(filepath.Dir(shared.OriginalPath))
	if _, err := conn.Exec(`
		INSERT INTO media_asset_links(id, profile_id, asset_id, asset_type, target_type, target_id, source)
		VALUES ('link-1','profile-1',?,'item_photo','wishlist','wish-1','test')
	`, shared.ID); err != nil {
		t.Fatalf("seed shared link: %v", err)
	}

	orphan, err := svc.Upload(context.Background(), "item-1", "orphan.jpg", bytes.NewReader(img))
	if err != nil {
		t.Fatalf("Upload() orphan error = %v", err)
	}
	orphanAssetDir := filepath.Dir(filepath.Dir(orphan.OriginalPath))

	if err := svc.Delete(context.Background(), "item-1", shared.ID); err != nil {
		t.Fatalf("Delete() shared error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(sharedAssetDir, "manifest.json")); err != nil {
		t.Fatalf("shared canonical asset folder should survive while linked: %v", err)
	}
	if _, err := os.Stat(shared.OriginalPath); err != nil {
		t.Fatalf("shared original should survive while linked: %v", err)
	}

	if err := svc.Delete(context.Background(), "item-1", orphan.ID); err != nil {
		t.Fatalf("Delete() orphan error = %v", err)
	}
	if _, err := os.Stat(orphanAssetDir); !os.IsNotExist(err) {
		t.Fatalf("orphan canonical asset folder should be removed, stat err=%v", err)
	}
}

func TestCreateCanonicalAssetCleansStagingAfterInterruptedRead(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	mediaRoot := filepath.Join(base, "media")
	svc := NewService(nil, mediaRoot)
	readErr := errors.New("interrupted upload")

	_, _, _, err := svc.createCanonicalAsset(
		context.Background(),
		mediaRoot,
		"asset-interrupted",
		"broken.bin",
		"broken.bin",
		"application/octet-stream",
		&interruptingReader{data: []byte("partial original bytes"), err: readErr},
		[]AssetManifestOwner{{Type: "chat_thread", ID: "thread-1"}},
		map[string]string{"source": "test"},
	)
	if !errors.Is(err, readErr) {
		t.Fatalf("createCanonicalAsset() error = %v, want %v", err, readErr)
	}
	if _, err := os.Stat(filepath.Join(mediaRoot, "assets", "asset-interrupted")); !os.IsNotExist(err) {
		t.Fatalf("interrupted ingestion should not leave visible asset folder, stat err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(mediaRoot, ".staging", "asset-interrupted")); !os.IsNotExist(err) {
		t.Fatalf("interrupted ingestion should remove staging folder, stat err=%v", err)
	}
}

func TestPreflightLegacyMediaMigrationReportsMixedLegacyNewAndOrphanStore(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	mediaRoot := filepath.Join(base, "media")
	legacyItemDir := filepath.Join(mediaRoot, "item-legacy")
	if err := os.MkdirAll(legacyItemDir, 0o755); err != nil {
		t.Fatalf("create legacy item dir: %v", err)
	}
	legacyOriginal := filepath.Join(legacyItemDir, "photo-legacy_orig.jpg")
	legacyOriginalStored := filepath.ToSlash(filepath.Join("item-legacy", "photo-legacy_orig.jpg"))
	if err := os.WriteFile(legacyOriginal, sampleJPEG(t), 0o644); err != nil {
		t.Fatalf("write legacy original: %v", err)
	}
	orphanPath := filepath.Join(legacyItemDir, "orphan_orig.jpg")
	if err := os.WriteFile(orphanPath, sampleJPEG(t), 0o644); err != nil {
		t.Fatalf("write orphan legacy file: %v", err)
	}
	legacyAttachmentRoot := filepath.Join(mediaRoot, "chat-attachments")
	if err := os.MkdirAll(legacyAttachmentRoot, 0o755); err != nil {
		t.Fatalf("create legacy attachment dir: %v", err)
	}
	legacyAttachmentPath := filepath.Join(legacyAttachmentRoot, "attach-legacy-note.txt")
	legacyAttachmentStored := filepath.ToSlash(filepath.Join("chat-attachments", "attach-legacy-note.txt"))
	if err := os.WriteFile(legacyAttachmentPath, []byte("legacy note"), 0o644); err != nil {
		t.Fatalf("write legacy attachment: %v", err)
	}
	missingOriginalStored := filepath.ToSlash(filepath.Join("item-legacy", "missing", "missing_orig.jpg"))

	dbPath := filepath.Join(base, "cabinet.db")
	conn, err := db.OpenAndMigrate(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	if _, err := conn.Exec(`
		INSERT INTO profiles (id, name) VALUES ('profile-1','One');
		INSERT INTO canonical_items (id, profile_id, brand, category, part_number, title) VALUES
			('item-legacy','profile-1','AFX','Slot Car','LEG-1','Legacy Car'),
			('item-canonical','profile-1','AFX','Slot Car','CAN-1','Canonical Car');
		INSERT INTO chat_threads (id, profile_id, title) VALUES ('thread-1','profile-1','Legacy attachments');
	`); err != nil {
		t.Fatalf("seed legacy media records: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO profile_settings(profile_id, key, value) VALUES ('profile-1','storage.media_dir', ?)`, mediaRoot); err != nil {
		t.Fatalf("seed media root setting: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO item_photos (id, item_id, filename, original_path, preview_path, thumbnail_path, is_primary, display_order) VALUES
			('photo-legacy','item-legacy','legacy.jpg', ?, '', '', 1, 1),
			('photo-missing','item-legacy','missing.jpg', ?, '', '', 0, 2)
	`, legacyOriginalStored, missingOriginalStored); err != nil {
		t.Fatalf("seed legacy photo rows: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO chat_attachments (id, profile_id, thread_id, filename, mime_type, size_bytes, stored_path)
		VALUES ('attach-legacy','profile-1','thread-1','attach-legacy-note.txt','text/plain',11, ?)
	`, legacyAttachmentStored); err != nil {
		t.Fatalf("seed legacy attachment row: %v", err)
	}

	svc := NewService(conn, mediaRoot)
	resolvedMediaRoot, err := svc.resolveMediaDirForProfile(context.Background(), "profile-1")
	if err != nil {
		t.Fatalf("resolve media root: %v", err)
	}
	if resolvedMediaRoot != mediaRoot {
		t.Fatalf("resolved media root = %q, want %q", resolvedMediaRoot, mediaRoot)
	}
	if _, err := svc.Upload(context.Background(), "item-canonical", "canonical.jpg", bytes.NewReader(sampleJPEG(t))); err != nil {
		t.Fatalf("seed canonical upload: %v", err)
	}

	report, err := svc.PreflightLegacyMediaMigration(context.Background(), "profile-1")
	if err != nil {
		t.Fatalf("PreflightLegacyMediaMigration() error = %v", err)
	}
	if !report.DryRun || report.ProfileID != "profile-1" {
		t.Fatalf("preflight should be profile-scoped dry run, got %+v", report)
	}
	if report.Summary.Discovered != 5 || report.Summary.Pending != 2 || report.Summary.AlreadyMigrated != 1 || report.Summary.Missing != 1 || report.Summary.Orphan != 1 {
		t.Fatalf("unexpected migration summary: %+v records=%+v", report.Summary, report.Records)
	}
	if !containsLegacyMigrationRecord(report.Records, "photo-legacy", "inventory_photo", "pending", "legacy_media") {
		t.Fatalf("legacy inventory photo not classified as pending: %+v", report.Records)
	}
	if !containsLegacyMigrationRecord(report.Records, "attach-legacy", "chat_attachment", "pending", "legacy_media") {
		t.Fatalf("legacy chat attachment not classified as pending legacy media: %+v", report.Records)
	}
	if !containsLegacyMigrationRecord(report.Records, "photo-missing", "inventory_photo", "missing", "legacy_media") {
		t.Fatalf("missing inventory photo not reported with record id: %+v", report.Records)
	}
	if !containsLegacyMigrationRecord(report.Records, "", "inventory_photo", "already_migrated", "canonical_asset") {
		t.Fatalf("canonical inventory photo not classified as already migrated: %+v", report.Records)
	}
	if !containsLegacyMigrationRecord(report.Records, "", "orphan_file", "orphan", "legacy_media") {
		t.Fatalf("orphan legacy file not reported without deletion: %+v", report.Records)
	}

	var originalPath string
	if err := conn.QueryRow(`SELECT original_path FROM item_photos WHERE id = 'photo-legacy'`).Scan(&originalPath); err != nil {
		t.Fatalf("read legacy original path after dry run: %v", err)
	}
	if originalPath != legacyOriginalStored {
		t.Fatalf("preflight mutated legacy path: got %q want %q", originalPath, legacyOriginalStored)
	}
	if _, err := os.Stat(orphanPath); err != nil {
		t.Fatalf("preflight should not delete orphan file: %v", err)
	}
}

func TestApplyLegacyInventoryPhotoMigrationPreservesOrderAndHashIdempotently(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	mediaRoot := filepath.Join(base, "media")
	legacyItemDir := filepath.Join(mediaRoot, "item-legacy")
	if err := os.MkdirAll(legacyItemDir, 0o755); err != nil {
		t.Fatalf("create legacy item dir: %v", err)
	}
	legacyBytes := sampleJPEG(t)
	legacyOriginal := filepath.Join(legacyItemDir, "front_orig.jpg")
	legacyOriginalStored := filepath.ToSlash(filepath.Join("item-legacy", "front_orig.jpg"))
	if err := os.WriteFile(legacyOriginal, legacyBytes, 0o644); err != nil {
		t.Fatalf("write legacy original: %v", err)
	}

	dbPath := filepath.Join(base, "cabinet.db")
	conn, err := db.OpenAndMigrate(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	if _, err := conn.Exec(`
		INSERT INTO profiles (id, name) VALUES ('profile-1','One');
		INSERT INTO canonical_items (id, profile_id, brand, category, part_number, title)
		VALUES ('item-legacy','profile-1','AFX','Slot Car','LEG-APPLY','Legacy Apply Car');
	`); err != nil {
		t.Fatalf("seed profile/item: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO profile_settings(profile_id, key, value) VALUES ('profile-1','storage.media_dir', ?)`, mediaRoot); err != nil {
		t.Fatalf("seed media root setting: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO item_photos (id, item_id, filename, original_path, preview_path, thumbnail_path, is_primary, display_order)
		VALUES ('photo-legacy','item-legacy','front.jpg', ?, '', '', 1, 7)
	`, legacyOriginalStored); err != nil {
		t.Fatalf("seed legacy photo row: %v", err)
	}

	svc := NewService(conn, mediaRoot)
	result, err := svc.ApplyLegacyInventoryPhotoMigration(context.Background(), "profile-1")
	if err != nil {
		t.Fatalf("ApplyLegacyInventoryPhotoMigration() error = %v", err)
	}
	if result.Migrated != 1 || result.AlreadyMigrated != 0 || result.Skipped != 0 || result.Failed != 0 {
		t.Fatalf("unexpected apply result: %+v", result)
	}

	var originalPath, previewPath, thumbnailPath string
	var isPrimary, displayOrder int
	if err := conn.QueryRow(`
		SELECT original_path, preview_path, thumbnail_path, is_primary, display_order
		FROM item_photos WHERE id = 'photo-legacy'
	`).Scan(&originalPath, &previewPath, &thumbnailPath, &isPrimary, &displayOrder); err != nil {
		t.Fatalf("read migrated photo row: %v", err)
	}
	if originalPath != filepath.ToSlash(filepath.Join("assets", "photo-legacy", "original", "front.jpg")) ||
		previewPath != filepath.ToSlash(filepath.Join("assets", "photo-legacy", "renditions", "preview.jpg")) ||
		thumbnailPath != filepath.ToSlash(filepath.Join("assets", "photo-legacy", "renditions", "thumbnail.jpg")) {
		t.Fatalf("unexpected migrated paths: original=%q preview=%q thumbnail=%q", originalPath, previewPath, thumbnailPath)
	}
	if isPrimary != 1 || displayOrder != 7 {
		t.Fatalf("migration changed primary/order: isPrimary=%d displayOrder=%d", isPrimary, displayOrder)
	}
	if _, err := os.Stat(legacyOriginal); err != nil {
		t.Fatalf("legacy source should be retained after verified migration: %v", err)
	}

	manifest := readAssetManifest(t, filepath.Join(mediaRoot, "assets", "photo-legacy", "manifest.json"))
	sourceHash := sha256.Sum256(legacyBytes)
	if manifest.Original.ContentHash != "sha256:"+hex.EncodeToString(sourceHash[:]) {
		t.Fatalf("manifest hash = %q, want source hash", manifest.Original.ContentHash)
	}
	if manifest.Provenance["legacy_source_hash"] != manifest.Original.ContentHash || manifest.Provenance["source"] != "legacy.inventory_photo.migration" {
		t.Fatalf("manifest provenance missing hash/source evidence: %+v", manifest.Provenance)
	}

	second, err := svc.ApplyLegacyInventoryPhotoMigration(context.Background(), "profile-1")
	if err != nil {
		t.Fatalf("ApplyLegacyInventoryPhotoMigration() second run error = %v", err)
	}
	if second.Migrated != 0 || second.AlreadyMigrated != 1 || second.Skipped != 0 || second.Failed != 0 {
		t.Fatalf("second run should be idempotent, got %+v", second)
	}
	entries, err := os.ReadDir(filepath.Join(mediaRoot, "assets"))
	if err != nil {
		t.Fatalf("read asset root: %v", err)
	}
	if len(entries) != 1 || entries[0].Name() != "photo-legacy" {
		t.Fatalf("migration should create exactly one stable asset folder, got %+v", entries)
	}
}

func TestApplyLegacyChatAttachmentMigrationPreservesLinksAndMetadataIdempotently(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	mediaRoot := filepath.Join(base, "media")
	legacyAttachmentRoot := filepath.Join(mediaRoot, "chat-attachments")
	if err := os.MkdirAll(legacyAttachmentRoot, 0o755); err != nil {
		t.Fatalf("create legacy attachment dir: %v", err)
	}
	legacyBytes := []byte("legacy chat attachment bytes")
	legacyAttachment := filepath.Join(legacyAttachmentRoot, "reference note.txt")
	legacyAttachmentStored := filepath.ToSlash(filepath.Join("chat-attachments", "reference note.txt"))
	if err := os.WriteFile(legacyAttachment, legacyBytes, 0o644); err != nil {
		t.Fatalf("write legacy attachment: %v", err)
	}

	dbPath := filepath.Join(base, "cabinet.db")
	conn, err := db.OpenAndMigrate(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	if _, err := conn.Exec(`
		INSERT INTO profiles (id, name) VALUES ('profile-1','One');
		INSERT INTO chat_threads (id, profile_id, title) VALUES ('thread-legacy','profile-1','Legacy attachments');
	`); err != nil {
		t.Fatalf("seed profile/thread: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO profile_settings(profile_id, key, value) VALUES ('profile-1','storage.media_dir', ?)`, mediaRoot); err != nil {
		t.Fatalf("seed media root setting: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO chat_attachments (id, profile_id, thread_id, filename, mime_type, size_bytes, stored_path)
		VALUES ('attach-legacy','profile-1','thread-legacy','Reference Note.txt','text/plain', ?, ?);
		INSERT INTO chat_messages(id, profile_id, thread_id, role, content, attachments_json, context_json)
		VALUES ('message-legacy','profile-1','thread-legacy','user','please use this attachment',
			'[{"id":"attach-legacy","filename":"Reference Note.txt","mime_type":"text/plain","size_bytes":28}]',
			'{"source":"test"}');
	`, len(legacyBytes), legacyAttachmentStored); err != nil {
		t.Fatalf("seed legacy chat attachment: %v", err)
	}

	svc := NewService(conn, mediaRoot)
	result, err := svc.ApplyLegacyChatAttachmentMigration(context.Background(), "profile-1")
	if err != nil {
		t.Fatalf("ApplyLegacyChatAttachmentMigration() error = %v", err)
	}
	if result.Migrated != 1 || result.AlreadyMigrated != 0 || result.Skipped != 0 || result.Failed != 0 {
		t.Fatalf("unexpected chat apply result: %+v", result)
	}

	var storedPath, filename, mimeType, messageAttachments string
	var sizeBytes int
	if err := conn.QueryRow(`
		SELECT ca.stored_path, ca.filename, ca.mime_type, ca.size_bytes, cm.attachments_json
		FROM chat_attachments ca
		INNER JOIN chat_messages cm ON cm.thread_id = ca.thread_id
		WHERE ca.id = 'attach-legacy'
	`).Scan(&storedPath, &filename, &mimeType, &sizeBytes, &messageAttachments); err != nil {
		t.Fatalf("read migrated chat attachment row: %v", err)
	}
	if storedPath != filepath.ToSlash(filepath.Join("assets", "attach-legacy", "original", "Reference Note.txt")) {
		t.Fatalf("unexpected migrated chat attachment path: %q", storedPath)
	}
	if filename != "Reference Note.txt" || mimeType != "text/plain" || sizeBytes != len(legacyBytes) {
		t.Fatalf("migration changed attachment metadata: filename=%q mime=%q size=%d", filename, mimeType, sizeBytes)
	}
	if messageAttachments != `[{"id":"attach-legacy","filename":"Reference Note.txt","mime_type":"text/plain","size_bytes":28}]` {
		t.Fatalf("migration changed message attachment links: %s", messageAttachments)
	}
	if _, err := os.Stat(legacyAttachment); err != nil {
		t.Fatalf("legacy source should be retained after verified chat migration: %v", err)
	}

	manifest := readAssetManifest(t, filepath.Join(mediaRoot, "assets", "attach-legacy", "manifest.json"))
	sourceHash := sha256.Sum256(legacyBytes)
	if manifest.Original.Filename != "Reference Note.txt" || manifest.Original.MIMEType != "text/plain" || manifest.Original.ContentHash != "sha256:"+hex.EncodeToString(sourceHash[:]) {
		t.Fatalf("manifest original metadata mismatch: %+v", manifest.Original)
	}
	if len(manifest.Owners) != 1 || manifest.Owners[0].Type != "chat_thread" || manifest.Owners[0].ID != "thread-legacy" {
		t.Fatalf("manifest owners should preserve chat thread linkage: %+v", manifest.Owners)
	}
	if manifest.Provenance["legacy_source_hash"] != manifest.Original.ContentHash || manifest.Provenance["source"] != "legacy.chat_attachment.migration" {
		t.Fatalf("manifest provenance missing hash/source evidence: %+v", manifest.Provenance)
	}

	second, err := svc.ApplyLegacyChatAttachmentMigration(context.Background(), "profile-1")
	if err != nil {
		t.Fatalf("ApplyLegacyChatAttachmentMigration() second run error = %v", err)
	}
	if second.Migrated != 0 || second.AlreadyMigrated != 1 || second.Skipped != 0 || second.Failed != 0 {
		t.Fatalf("second run should be idempotent, got %+v", second)
	}
	entries, err := os.ReadDir(filepath.Join(mediaRoot, "assets"))
	if err != nil {
		t.Fatalf("read asset root: %v", err)
	}
	if len(entries) != 1 || entries[0].Name() != "attach-legacy" {
		t.Fatalf("migration should create exactly one stable chat asset folder, got %+v", entries)
	}
}

func TestReorderPersistsListOrder(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	dbPath := filepath.Join(base, "cabinet.db")
	conn, err := db.OpenAndMigrate(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	if _, err := conn.Exec(`INSERT INTO canonical_items (id, brand, category, part_number, title) VALUES ('item-1','AFX','Slot Car','P-1','Car')`); err != nil {
		t.Fatalf("seed item: %v", err)
	}

	svc := NewService(conn, filepath.Join(base, "media"))
	img := sampleJPEG(t)

	p1, err := svc.Upload(context.Background(), "item-1", "a.jpg", bytes.NewReader(img))
	if err != nil {
		t.Fatalf("Upload() p1 error = %v", err)
	}
	p2, err := svc.Upload(context.Background(), "item-1", "b.jpg", bytes.NewReader(img))
	if err != nil {
		t.Fatalf("Upload() p2 error = %v", err)
	}
	p3, err := svc.Upload(context.Background(), "item-1", "c.jpg", bytes.NewReader(img))
	if err != nil {
		t.Fatalf("Upload() p3 error = %v", err)
	}

	if err := svc.Reorder(context.Background(), "item-1", []string{p3.ID, p1.ID, p2.ID}); err != nil {
		t.Fatalf("Reorder() error = %v", err)
	}
	ordered, err := svc.ListByItem(context.Background(), "item-1")
	if err != nil {
		t.Fatalf("ListByItem() error = %v", err)
	}
	if len(ordered) != 3 {
		t.Fatalf("expected 3 photos, got %d", len(ordered))
	}
	if ordered[0].ID != p3.ID || ordered[1].ID != p1.ID || ordered[2].ID != p2.ID {
		t.Fatalf("unexpected order: %s %s %s", ordered[0].ID, ordered[1].ID, ordered[2].ID)
	}
}

func sampleJPEG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 4), G: uint8(y * 4), B: 100, A: 255})
		}
	}
	var b bytes.Buffer
	if err := jpeg.Encode(&b, img, &jpeg.Options{Quality: 85}); err != nil {
		t.Fatalf("encode sample jpeg: %v", err)
	}
	return b.Bytes()
}

type interruptingReader struct {
	data []byte
	err  error
	read bool
}

func (r *interruptingReader) Read(p []byte) (int, error) {
	if r.read {
		return 0, r.err
	}
	r.read = true
	return copy(p, r.data), nil
}

func containsAssetState(assets []WorkspaceAsset, id, state string) bool {
	for _, asset := range assets {
		if asset.ID == id && asset.LinkageState == state {
			return true
		}
	}
	return false
}

func containsAssetTarget(assets []WorkspaceAsset, id, state, itemID, wishlistID string) bool {
	for _, asset := range assets {
		if asset.ID == id && asset.LinkageState == state && asset.ItemID == itemID && asset.WishlistID == wishlistID {
			return true
		}
	}
	return false
}

func containsLegacyMigrationRecord(records []LegacyMigrationRecord, id, recordType, classification, pathClass string) bool {
	for _, record := range records {
		if id != "" && record.ID != id {
			continue
		}
		if record.RecordType == recordType && record.Classification == classification && record.PathClass == pathClass {
			return true
		}
	}
	return false
}

func readAssetManifest(t *testing.T, path string) AssetManifest {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read asset manifest: %v", err)
	}
	var manifest AssetManifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatalf("decode asset manifest: %v", err)
	}
	return manifest
}
