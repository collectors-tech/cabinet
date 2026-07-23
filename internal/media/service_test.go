package media

import (
	"archive/zip"
	"bytes"
	"context"
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
