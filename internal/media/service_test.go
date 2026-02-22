package media

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"path/filepath"
	"testing"

	"github.com/collectors-tech/cabinet/internal/db"
)

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
