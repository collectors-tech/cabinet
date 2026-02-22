package media

import (
	"context"
	"database/sql"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type Photo struct {
	ID            string `json:"id"`
	ItemID        string `json:"item_id"`
	Filename      string `json:"filename"`
	OriginalPath  string `json:"original_path"`
	PreviewPath   string `json:"preview_path"`
	ThumbnailPath string `json:"thumbnail_path"`
	IsPrimary     bool   `json:"is_primary"`
	CreatedAt     string `json:"created_at"`
}

type Service struct {
	db       *sql.DB
	mediaDir string
}

func NewService(db *sql.DB, mediaDir string) *Service {
	return &Service{db: db, mediaDir: mediaDir}
}

func (s *Service) Upload(ctx context.Context, itemID, filename string, r io.Reader) (Photo, error) {
	itemID = strings.TrimSpace(itemID)
	if itemID == "" {
		return Photo{}, fmt.Errorf("item_id is required")
	}
	if strings.TrimSpace(filename) == "" {
		filename = "upload.jpg"
	}

	photoID := uuid.NewString()
	itemDir := filepath.Join(s.mediaDir, itemID)
	if err := os.MkdirAll(itemDir, 0o755); err != nil {
		return Photo{}, fmt.Errorf("create item media dir: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		ext = ".jpg"
	}
	origPath := filepath.Join(itemDir, photoID+"_orig"+ext)
	previewPath := filepath.Join(itemDir, photoID+"_preview.jpg")
	thumbPath := filepath.Join(itemDir, photoID+"_thumb.jpg")

	origFile, err := os.Create(origPath)
	if err != nil {
		return Photo{}, fmt.Errorf("create original file: %w", err)
	}
	if _, err := io.Copy(origFile, r); err != nil {
		origFile.Close()
		return Photo{}, fmt.Errorf("save original file: %w", err)
	}
	if err := origFile.Close(); err != nil {
		return Photo{}, fmt.Errorf("close original file: %w", err)
	}

	if err := generateScaledJPEG(origPath, previewPath, 1024); err != nil {
		return Photo{}, fmt.Errorf("generate preview: %w", err)
	}
	if err := generateScaledJPEG(origPath, thumbPath, 256); err != nil {
		return Photo{}, fmt.Errorf("generate thumbnail: %w", err)
	}

	var existingCount int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM item_photos WHERE item_id = ?`, itemID).Scan(&existingCount); err != nil {
		return Photo{}, fmt.Errorf("count existing photos: %w", err)
	}
	isPrimary := 0
	if existingCount == 0 {
		isPrimary = 1
	}

	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO item_photos (id, item_id, filename, original_path, preview_path, thumbnail_path, is_primary)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, photoID, itemID, filename, origPath, previewPath, thumbPath, isPrimary); err != nil {
		return Photo{}, fmt.Errorf("insert photo record: %w", err)
	}

	return s.GetByID(ctx, photoID)
}

func (s *Service) ListByItem(ctx context.Context, itemID string) ([]Photo, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, item_id, filename, original_path, preview_path, thumbnail_path, is_primary, created_at
		FROM item_photos WHERE item_id = ? ORDER BY created_at ASC
	`, itemID)
	if err != nil {
		return nil, fmt.Errorf("list photos: %w", err)
	}
	defer rows.Close()

	var out []Photo
	for rows.Next() {
		var p Photo
		var isPrimary int
		if err := rows.Scan(&p.ID, &p.ItemID, &p.Filename, &p.OriginalPath, &p.PreviewPath, &p.ThumbnailPath, &isPrimary, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan photo: %w", err)
		}
		p.IsPrimary = isPrimary == 1
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate photos: %w", err)
	}
	return out, nil
}

func (s *Service) GetByID(ctx context.Context, photoID string) (Photo, error) {
	var p Photo
	var isPrimary int
	err := s.db.QueryRowContext(ctx, `
		SELECT id, item_id, filename, original_path, preview_path, thumbnail_path, is_primary, created_at
		FROM item_photos WHERE id = ?
	`, photoID).Scan(&p.ID, &p.ItemID, &p.Filename, &p.OriginalPath, &p.PreviewPath, &p.ThumbnailPath, &isPrimary, &p.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return Photo{}, fmt.Errorf("photo not found")
		}
		return Photo{}, fmt.Errorf("get photo: %w", err)
	}
	p.IsPrimary = isPrimary == 1
	return p, nil
}

func (s *Service) SetPrimary(ctx context.Context, itemID, photoID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `UPDATE item_photos SET is_primary = 0 WHERE item_id = ?`, itemID); err != nil {
		return fmt.Errorf("clear primary: %w", err)
	}
	res, err := tx.ExecContext(ctx, `UPDATE item_photos SET is_primary = 1 WHERE id = ? AND item_id = ?`, photoID, itemID)
	if err != nil {
		return fmt.Errorf("set primary: %w", err)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("photo not found")
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

func (s *Service) Delete(ctx context.Context, itemID, photoID string) error {
	p, err := s.GetByID(ctx, photoID)
	if err != nil {
		return err
	}
	if p.ItemID != itemID {
		return fmt.Errorf("photo does not belong to item")
	}

	if _, err := s.db.ExecContext(ctx, `DELETE FROM item_photos WHERE id = ?`, photoID); err != nil {
		return fmt.Errorf("delete photo record: %w", err)
	}
	_ = os.Remove(p.OriginalPath)
	_ = os.Remove(p.PreviewPath)
	_ = os.Remove(p.ThumbnailPath)

	photos, err := s.ListByItem(ctx, itemID)
	if err == nil && len(photos) > 0 {
		hasPrimary := false
		for _, ph := range photos {
			if ph.IsPrimary {
				hasPrimary = true
				break
			}
		}
		if !hasPrimary {
			_ = s.SetPrimary(ctx, itemID, photos[0].ID)
		}
	}
	return nil
}

func (s *Service) RebuildThumbnails(ctx context.Context, itemID string) error {
	photos, err := s.ListByItem(ctx, itemID)
	if err != nil {
		return err
	}
	for _, p := range photos {
		if err := generateScaledJPEG(p.OriginalPath, p.PreviewPath, 1024); err != nil {
			return fmt.Errorf("rebuild preview %s: %w", p.ID, err)
		}
		if err := generateScaledJPEG(p.OriginalPath, p.ThumbnailPath, 256); err != nil {
			return fmt.Errorf("rebuild thumbnail %s: %w", p.ID, err)
		}
	}
	return nil
}

func generateScaledJPEG(inputPath, outputPath string, maxSize int) error {
	f, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("open input image: %w", err)
	}
	defer f.Close()

	src, _, err := image.Decode(f)
	if err != nil {
		return fmt.Errorf("decode image: %w", err)
	}

	dst := scaleToFit(src, maxSize)
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create output image: %w", err)
	}
	defer outFile.Close()

	if err := jpeg.Encode(outFile, dst, &jpeg.Options{Quality: 85}); err != nil {
		return fmt.Errorf("encode jpeg: %w", err)
	}
	return nil
}

func scaleToFit(src image.Image, maxSize int) *image.RGBA {
	b := src.Bounds()
	w := b.Dx()
	h := b.Dy()
	if w <= 0 || h <= 0 {
		return image.NewRGBA(image.Rect(0, 0, 1, 1))
	}
	scale := 1.0
	if w > maxSize || h > maxSize {
		scale = math.Min(float64(maxSize)/float64(w), float64(maxSize)/float64(h))
	}
	nw := int(math.Max(1, math.Round(float64(w)*scale)))
	nh := int(math.Max(1, math.Round(float64(h)*scale)))
	dst := image.NewRGBA(image.Rect(0, 0, nw, nh))
	for y := 0; y < nh; y++ {
		srcY := int(float64(y) / float64(nh) * float64(h))
		if srcY >= h {
			srcY = h - 1
		}
		for x := 0; x < nw; x++ {
			srcX := int(float64(x) / float64(nw) * float64(w))
			if srcX >= w {
				srcX = w - 1
			}
			dst.Set(x, y, src.At(b.Min.X+srcX, b.Min.Y+srcY))
		}
	}
	return dst
}
