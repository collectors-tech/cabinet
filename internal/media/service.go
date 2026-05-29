package media

import (
	"archive/zip"
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
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
	DisplayOrder  int    `json:"display_order"`
	CreatedAt     string `json:"created_at"`
}

type Service struct {
	db       *sql.DB
	mediaDir string
}

func NewService(db *sql.DB, mediaDir string) *Service {
	return &Service{db: db, mediaDir: mediaDir}
}

type WorkspaceAsset struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	Filename         string `json:"filename"`
	UploadedAt       string `json:"uploaded_at"`
	LinkageState     string `json:"linkage_state"`
	AnalysisStatus   string `json:"analysis_status"`
	Source           string `json:"source"`
	ItemID           string `json:"item_id,omitempty"`
	WishlistID       string `json:"wishlist_id,omitempty"`
	ThumbnailURL     string `json:"thumbnail_url,omitempty"`
	DownloadFilename string `json:"download_filename"`
	StoredPath       string `json:"-"`
}

type WorkspaceSummary struct {
	Total           int `json:"total"`
	Unlinked        int `json:"unlinked"`
	LinkedInventory int `json:"linked_inventory"`
	LinkedWishlist  int `json:"linked_wishlist"`
	LinkedBoth      int `json:"linked_both"`
	ReadyForReview  int `json:"ready_for_review"`
}

type WorkspaceList struct {
	Assets  []WorkspaceAsset `json:"assets"`
	Summary WorkspaceSummary `json:"summary"`
	Filter  string           `json:"filter"`
}

type AssignmentPreview struct {
	AssetID               string `json:"asset_id"`
	TargetType            string `json:"target_type"`
	TargetID              string `json:"target_id"`
	CurrentLinkageState   string `json:"current_linkage_state"`
	ProjectedLinkageState string `json:"projected_linkage_state"`
	RequiresConfirmation  bool   `json:"requires_confirmation"`
	Allowed               bool   `json:"allowed"`
	BlockedReason         string `json:"blocked_reason"`
	AuditSummary          string `json:"audit_summary"`
}

type AssignmentResult struct {
	AssignmentPreview
	Applied bool `json:"applied"`
}

type DownloadPreview struct {
	AssetIDs  []string `json:"asset_ids"`
	Count     int      `json:"count"`
	Filenames []string `json:"filenames"`
	Allowed   bool     `json:"allowed"`
}

type DownloadBundle struct {
	Filename    string
	ContentType string
	Bytes       []byte
	AssetIDs    []string
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
	rootMediaDir, err := s.resolveMediaDirForItem(ctx, itemID)
	if err != nil {
		return Photo{}, err
	}
	itemDir := filepath.Join(rootMediaDir, itemID)
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
	var nextOrder int
	if err := s.db.QueryRowContext(ctx, `SELECT COALESCE(MAX(display_order), 0) + 1 FROM item_photos WHERE item_id = ?`, itemID).Scan(&nextOrder); err != nil {
		return Photo{}, fmt.Errorf("next photo order: %w", err)
	}

	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO item_photos (id, item_id, filename, original_path, preview_path, thumbnail_path, is_primary, display_order)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, photoID, itemID, filename, origPath, previewPath, thumbPath, isPrimary, nextOrder); err != nil {
		return Photo{}, fmt.Errorf("insert photo record: %w", err)
	}

	return s.GetByID(ctx, photoID)
}

func (s *Service) ListWorkspaceAssets(ctx context.Context, profileID, filter string) (WorkspaceList, error) {
	profileID = strings.TrimSpace(profileID)
	if profileID == "" {
		return WorkspaceList{}, fmt.Errorf("profile_id is required")
	}
	filter = strings.ToLower(strings.TrimSpace(filter))
	if filter == "" {
		filter = "all"
	}
	if filter != "all" && filter != "unlinked" {
		return WorkspaceList{}, fmt.Errorf("filter must be all or unlinked")
	}

	links, err := s.loadWorkspaceLinks(ctx, profileID)
	if err != nil {
		return WorkspaceList{}, err
	}

	assets := make([]WorkspaceAsset, 0)
	photoRows, err := s.db.QueryContext(ctx, `
		SELECT ip.id, ip.filename, ip.created_at, ip.original_path, ci.id, ci.part_number, ci.title
		FROM item_photos ip
		INNER JOIN canonical_items ci ON ci.id = ip.item_id
		WHERE ci.profile_id = ?
		ORDER BY ip.created_at DESC, ip.display_order ASC, ip.id ASC
	`, profileID)
	if err != nil {
		return WorkspaceList{}, fmt.Errorf("list inventory media assets: %w", err)
	}
	for photoRows.Next() {
		var assetID, filename, createdAt, originalPath, itemID, partNumber, title string
		if err := photoRows.Scan(&assetID, &filename, &createdAt, &originalPath, &itemID, &partNumber, &title); err != nil {
			photoRows.Close()
			return WorkspaceList{}, fmt.Errorf("scan inventory media asset: %w", err)
		}
		displayTitle := strings.TrimSpace(title)
		if displayTitle == "" {
			displayTitle = strings.TrimSpace(filename)
		}
		link := links[assetID]
		linkageState := linkageStateForAsset(true, link)
		assets = append(assets, WorkspaceAsset{
			ID:               assetID,
			Title:            displayTitle,
			Filename:         filename,
			UploadedAt:       createdAt,
			LinkageState:     linkageState,
			AnalysisStatus:   "not_analyzed",
			Source:           "Inventory photo",
			ItemID:           itemID,
			WishlistID:       link.WishlistID,
			ThumbnailURL:     "/api/items/" + itemID + "/photos/" + assetID + "/file?variant=thumbnail",
			DownloadFilename: friendlyMediaFilename(partNumber, displayTitle, filename, assetID),
			StoredPath:       originalPath,
		})
	}
	if err := photoRows.Err(); err != nil {
		photoRows.Close()
		return WorkspaceList{}, fmt.Errorf("iterate inventory media assets: %w", err)
	}
	photoRows.Close()

	attachmentRows, err := s.db.QueryContext(ctx, `
		SELECT id, filename, stored_path, created_at
		FROM chat_attachments
		WHERE profile_id = ?
		ORDER BY created_at DESC, id ASC
	`, profileID)
	if err != nil {
		return WorkspaceList{}, fmt.Errorf("list unlinked media assets: %w", err)
	}
	for attachmentRows.Next() {
		var assetID, filename, storedPath, createdAt string
		if err := attachmentRows.Scan(&assetID, &filename, &storedPath, &createdAt); err != nil {
			attachmentRows.Close()
			return WorkspaceList{}, fmt.Errorf("scan unlinked media asset: %w", err)
		}
		link := links[assetID]
		assets = append(assets, WorkspaceAsset{
			ID:               assetID,
			Title:            strings.TrimSpace(filename),
			Filename:         filename,
			UploadedAt:       createdAt,
			LinkageState:     linkageStateForAsset(false, link),
			AnalysisStatus:   "pending",
			Source:           "Chat attachment",
			ItemID:           link.ItemID,
			WishlistID:       link.WishlistID,
			DownloadFilename: friendlyMediaFilename("", filename, filename, assetID),
			StoredPath:       storedPath,
		})
	}
	if err := attachmentRows.Err(); err != nil {
		attachmentRows.Close()
		return WorkspaceList{}, fmt.Errorf("iterate unlinked media assets: %w", err)
	}
	attachmentRows.Close()

	allAssets := assets
	if filter == "unlinked" {
		assets = make([]WorkspaceAsset, 0)
		for _, asset := range allAssets {
			if asset.LinkageState == "unlinked" {
				assets = append(assets, asset)
			}
		}
	}
	return WorkspaceList{Assets: assets, Summary: summarizeWorkspaceAssets(allAssets), Filter: filter}, nil
}

func (s *Service) PreviewAssignment(ctx context.Context, profileID, assetID, targetType, targetID string) (AssignmentPreview, error) {
	profileID = strings.TrimSpace(profileID)
	assetID = strings.TrimSpace(assetID)
	targetType = strings.ToLower(strings.TrimSpace(targetType))
	targetID = strings.TrimSpace(targetID)
	if profileID == "" || assetID == "" || targetType == "" || targetID == "" {
		return AssignmentPreview{}, fmt.Errorf("profile_id, asset_id, target_type, and target_id are required")
	}
	if targetType != "inventory" && targetType != "wishlist" {
		return AssignmentPreview{}, fmt.Errorf("target_type must be inventory or wishlist")
	}
	list, err := s.ListWorkspaceAssets(ctx, profileID, "all")
	if err != nil {
		return AssignmentPreview{}, err
	}
	var asset WorkspaceAsset
	found := false
	for _, item := range list.Assets {
		if item.ID == assetID {
			asset = item
			found = true
			break
		}
	}
	if !found {
		return AssignmentPreview{}, fmt.Errorf("media asset not found")
	}
	if err := s.assertTargetExists(ctx, profileID, targetType, targetID); err != nil {
		return AssignmentPreview{}, err
	}
	projected := projectedLinkageState(asset.LinkageState, targetType)
	return AssignmentPreview{
		AssetID:               assetID,
		TargetType:            targetType,
		TargetID:              targetID,
		CurrentLinkageState:   asset.LinkageState,
		ProjectedLinkageState: projected,
		RequiresConfirmation:  true,
		Allowed:               true,
		AuditSummary:          "Will preserve media asset " + assetID + " provenance while linking to " + targetType + " target " + targetID + ".",
	}, nil
}

func (s *Service) ApplyAssignment(ctx context.Context, profileID, assetID, targetType, targetID string) (AssignmentResult, error) {
	preview, err := s.PreviewAssignment(ctx, profileID, assetID, targetType, targetID)
	if err != nil {
		return AssignmentResult{}, err
	}
	assetType, err := s.assetType(ctx, strings.TrimSpace(profileID), strings.TrimSpace(assetID))
	if err != nil {
		return AssignmentResult{}, err
	}
	linkID := uuid.NewString()
	auditSummary := "Preserved media asset " + preview.AssetID + " provenance while linking to " + preview.TargetType + " target " + preview.TargetID + "."
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO media_asset_links(id, profile_id, asset_id, asset_type, target_type, target_id, source, audit_summary)
		VALUES (?, ?, ?, ?, ?, ?, 'media.workspace', ?)
		ON CONFLICT(profile_id, asset_id, target_type, target_id) DO UPDATE SET
			source = excluded.source,
			audit_summary = excluded.audit_summary,
			updated_at = CURRENT_TIMESTAMP
	`, linkID, strings.TrimSpace(profileID), preview.AssetID, assetType, preview.TargetType, preview.TargetID, auditSummary)
	if err != nil {
		return AssignmentResult{}, fmt.Errorf("apply media assignment: %w", err)
	}
	preview.CurrentLinkageState = preview.ProjectedLinkageState
	preview.BlockedReason = ""
	preview.AuditSummary = auditSummary
	return AssignmentResult{AssignmentPreview: preview, Applied: true}, nil
}

func (s *Service) PreviewDownload(ctx context.Context, profileID string, assetIDs []string, filter string) (DownloadPreview, error) {
	assets, err := s.resolveDownloadAssets(ctx, profileID, assetIDs, filter)
	if err != nil {
		return DownloadPreview{}, err
	}
	out := DownloadPreview{Allowed: true}
	for _, asset := range assets {
		out.AssetIDs = append(out.AssetIDs, asset.ID)
		out.Filenames = append(out.Filenames, asset.DownloadFilename)
	}
	out.Count = len(out.AssetIDs)
	return out, nil
}

func (s *Service) BuildDownload(ctx context.Context, profileID string, assetIDs []string, filter string) (DownloadBundle, error) {
	assets, err := s.resolveDownloadAssets(ctx, profileID, assetIDs, filter)
	if err != nil {
		return DownloadBundle{}, err
	}
	if len(assets) == 0 {
		return DownloadBundle{}, fmt.Errorf("no media assets selected for download")
	}
	if len(assets) == 1 {
		asset := assets[0]
		data, err := os.ReadFile(asset.StoredPath)
		if err != nil {
			return DownloadBundle{}, fmt.Errorf("read media asset %s: %w", asset.ID, err)
		}
		return DownloadBundle{
			Filename:    asset.DownloadFilename,
			ContentType: contentTypeForFilename(asset.Filename),
			Bytes:       data,
			AssetIDs:    []string{asset.ID},
		}, nil
	}

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	usedNames := map[string]int{}
	bundleAssetIDs := make([]string, 0, len(assets))
	for _, asset := range assets {
		data, err := os.ReadFile(asset.StoredPath)
		if err != nil {
			_ = zw.Close()
			return DownloadBundle{}, fmt.Errorf("read media asset %s: %w", asset.ID, err)
		}
		name := uniqueDownloadName(asset.DownloadFilename, usedNames)
		entry, err := zw.Create(name)
		if err != nil {
			_ = zw.Close()
			return DownloadBundle{}, fmt.Errorf("create media archive entry: %w", err)
		}
		if _, err := entry.Write(data); err != nil {
			_ = zw.Close()
			return DownloadBundle{}, fmt.Errorf("write media archive entry: %w", err)
		}
		bundleAssetIDs = append(bundleAssetIDs, asset.ID)
	}
	if err := zw.Close(); err != nil {
		return DownloadBundle{}, fmt.Errorf("close media archive: %w", err)
	}
	return DownloadBundle{
		Filename:    "cabinet-media-download.zip",
		ContentType: "application/zip",
		Bytes:       buf.Bytes(),
		AssetIDs:    bundleAssetIDs,
	}, nil
}

func (s *Service) resolveDownloadAssets(ctx context.Context, profileID string, assetIDs []string, filter string) ([]WorkspaceAsset, error) {
	list, err := s.ListWorkspaceAssets(ctx, profileID, filter)
	if err != nil {
		return nil, err
	}
	requested := make(map[string]bool, len(assetIDs))
	for _, id := range assetIDs {
		id = strings.TrimSpace(id)
		if id != "" {
			requested[id] = true
		}
	}
	out := make([]WorkspaceAsset, 0, len(list.Assets))
	for _, asset := range list.Assets {
		if len(requested) > 0 && !requested[asset.ID] {
			continue
		}
		if strings.TrimSpace(asset.StoredPath) == "" {
			return nil, fmt.Errorf("media asset %s has no stored path", asset.ID)
		}
		out = append(out, asset)
	}
	if len(requested) > 0 && len(out) != len(requested) {
		return nil, fmt.Errorf("one or more media assets were not found in current scope")
	}
	return out, nil
}

func (s *Service) assertTargetExists(ctx context.Context, profileID, targetType, targetID string) error {
	var count int
	var err error
	switch targetType {
	case "inventory":
		err = s.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM canonical_items WHERE id = ? AND profile_id = ?`, targetID, profileID).Scan(&count)
	case "wishlist":
		err = s.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM wishlist_entries WHERE id = ? AND profile_id = ?`, targetID, profileID).Scan(&count)
	default:
		return fmt.Errorf("target_type must be inventory or wishlist")
	}
	if err != nil {
		return fmt.Errorf("check assignment target: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("assignment target not found")
	}
	return nil
}

type workspaceLinkState struct {
	ItemID     string
	WishlistID string
}

func (s *Service) loadWorkspaceLinks(ctx context.Context, profileID string) (map[string]workspaceLinkState, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT asset_id, target_type, target_id
		FROM media_asset_links
		WHERE profile_id = ?
		ORDER BY created_at ASC, id ASC
	`, profileID)
	if err != nil {
		return nil, fmt.Errorf("list media asset links: %w", err)
	}
	defer rows.Close()

	out := map[string]workspaceLinkState{}
	for rows.Next() {
		var assetID, targetType, targetID string
		if err := rows.Scan(&assetID, &targetType, &targetID); err != nil {
			return nil, fmt.Errorf("scan media asset link: %w", err)
		}
		state := out[assetID]
		switch targetType {
		case "inventory":
			if state.ItemID == "" {
				state.ItemID = targetID
			}
		case "wishlist":
			if state.WishlistID == "" {
				state.WishlistID = targetID
			}
		}
		out[assetID] = state
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate media asset links: %w", err)
	}
	return out, nil
}

func (s *Service) assetType(ctx context.Context, profileID, assetID string) (string, error) {
	var count int
	if err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(1)
		FROM item_photos ip
		INNER JOIN canonical_items ci ON ci.id = ip.item_id
		WHERE ip.id = ? AND ci.profile_id = ?
	`, assetID, profileID).Scan(&count); err != nil {
		return "", fmt.Errorf("check item photo asset: %w", err)
	}
	if count > 0 {
		return "item_photo", nil
	}
	if err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(1)
		FROM chat_attachments
		WHERE id = ? AND profile_id = ?
	`, assetID, profileID).Scan(&count); err != nil {
		return "", fmt.Errorf("check chat attachment asset: %w", err)
	}
	if count > 0 {
		return "chat_attachment", nil
	}
	return "", fmt.Errorf("media asset not found")
}

func (s *Service) resolveMediaDirForItem(ctx context.Context, itemID string) (string, error) {
	itemID = strings.TrimSpace(itemID)
	if itemID == "" {
		return "", fmt.Errorf("item_id is required")
	}

	var configuredMediaDir sql.NullString
	err := s.db.QueryRowContext(
		ctx,
		`
		SELECT ps.value
		FROM canonical_items ci
		LEFT JOIN profile_settings ps
			ON ps.profile_id = ci.profile_id
			AND ps.key = 'storage.media_dir'
		WHERE ci.id = ?
		`,
		itemID,
	).Scan(&configuredMediaDir)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("item not found")
		}
		return "", fmt.Errorf("resolve item media dir: %w", err)
	}

	mediaDir := strings.TrimSpace(configuredMediaDir.String)
	if mediaDir != "" {
		return mediaDir, nil
	}

	return s.mediaDir, nil
}

func (s *Service) ListByItem(ctx context.Context, itemID string) ([]Photo, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, item_id, filename, original_path, preview_path, thumbnail_path, is_primary, display_order, created_at
		FROM item_photos
		WHERE item_id = ?
		ORDER BY CASE WHEN display_order > 0 THEN 0 ELSE 1 END ASC, display_order ASC, created_at ASC
	`, itemID)
	if err != nil {
		return nil, fmt.Errorf("list photos: %w", err)
	}
	defer rows.Close()

	var out []Photo
	for rows.Next() {
		var p Photo
		var isPrimary int
		if err := rows.Scan(&p.ID, &p.ItemID, &p.Filename, &p.OriginalPath, &p.PreviewPath, &p.ThumbnailPath, &isPrimary, &p.DisplayOrder, &p.CreatedAt); err != nil {
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
		SELECT id, item_id, filename, original_path, preview_path, thumbnail_path, is_primary, display_order, created_at
		FROM item_photos WHERE id = ?
	`, photoID).Scan(&p.ID, &p.ItemID, &p.Filename, &p.OriginalPath, &p.PreviewPath, &p.ThumbnailPath, &isPrimary, &p.DisplayOrder, &p.CreatedAt)
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

func (s *Service) Rotate(ctx context.Context, itemID, photoID, direction string) (Photo, error) {
	itemID = strings.TrimSpace(itemID)
	photoID = strings.TrimSpace(photoID)
	direction = strings.ToLower(strings.TrimSpace(direction))
	if itemID == "" {
		return Photo{}, fmt.Errorf("item_id is required")
	}
	if photoID == "" {
		return Photo{}, fmt.Errorf("photo_id is required")
	}
	if direction != "left" && direction != "right" {
		return Photo{}, fmt.Errorf("direction must be left or right")
	}

	p, err := s.GetByID(ctx, photoID)
	if err != nil {
		return Photo{}, err
	}
	if p.ItemID != itemID {
		return Photo{}, fmt.Errorf("photo does not belong to item")
	}

	srcFile, err := os.Open(p.OriginalPath)
	if err != nil {
		return Photo{}, fmt.Errorf("open original image: %w", err)
	}
	src, format, err := image.Decode(srcFile)
	closeErr := srcFile.Close()
	if err != nil {
		return Photo{}, fmt.Errorf("decode original image: %w", err)
	}
	if closeErr != nil {
		return Photo{}, fmt.Errorf("close original image: %w", closeErr)
	}

	rotated := rotateImage(src, direction)
	if err := encodeOriginalImage(p.OriginalPath, format, rotated); err != nil {
		return Photo{}, err
	}
	if err := generateScaledJPEG(p.OriginalPath, p.PreviewPath, 1024); err != nil {
		return Photo{}, fmt.Errorf("rebuild rotated preview: %w", err)
	}
	if err := generateScaledJPEG(p.OriginalPath, p.ThumbnailPath, 256); err != nil {
		return Photo{}, fmt.Errorf("rebuild rotated thumbnail: %w", err)
	}

	return s.GetByID(ctx, photoID)
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

func (s *Service) RebuildAllThumbnails(ctx context.Context) (int, int, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT DISTINCT item_id
		FROM item_photos
		WHERE TRIM(COALESCE(item_id, '')) <> ''
		ORDER BY item_id ASC
	`)
	if err != nil {
		return 0, 0, fmt.Errorf("list photo items: %w", err)
	}
	defer rows.Close()

	itemIDs := make([]string, 0)
	for rows.Next() {
		var itemID string
		if err := rows.Scan(&itemID); err != nil {
			return 0, 0, fmt.Errorf("scan photo item: %w", err)
		}
		itemID = strings.TrimSpace(itemID)
		if itemID == "" {
			continue
		}
		itemIDs = append(itemIDs, itemID)
	}
	if err := rows.Err(); err != nil {
		return 0, 0, fmt.Errorf("iterate photo items: %w", err)
	}

	rebuiltItems := 0
	rebuiltPhotos := 0
	for _, itemID := range itemIDs {
		photos, err := s.ListByItem(ctx, itemID)
		if err != nil {
			return rebuiltItems, rebuiltPhotos, err
		}
		if len(photos) == 0 {
			continue
		}
		if err := s.RebuildThumbnails(ctx, itemID); err != nil {
			return rebuiltItems, rebuiltPhotos, err
		}
		rebuiltItems += 1
		rebuiltPhotos += len(photos)
	}

	return rebuiltItems, rebuiltPhotos, nil
}

func (s *Service) ResolveVariantPath(ctx context.Context, itemID, photoID, variant string) (string, error) {
	p, err := s.GetByID(ctx, photoID)
	if err != nil {
		return "", err
	}
	if p.ItemID != itemID {
		return "", fmt.Errorf("photo does not belong to item")
	}
	switch strings.ToLower(strings.TrimSpace(variant)) {
	case "", "original":
		return p.OriginalPath, nil
	case "preview":
		return p.PreviewPath, nil
	case "thumbnail", "thumb":
		return p.ThumbnailPath, nil
	default:
		return "", fmt.Errorf("invalid photo variant")
	}
}

func (s *Service) Reorder(ctx context.Context, itemID string, orderedIDs []string) error {
	itemID = strings.TrimSpace(itemID)
	if itemID == "" {
		return fmt.Errorf("item_id is required")
	}
	if len(orderedIDs) == 0 {
		return fmt.Errorf("photo_ids are required")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	for idx, id := range orderedIDs {
		photoID := strings.TrimSpace(id)
		if photoID == "" {
			return fmt.Errorf("photo_id is required")
		}
		res, err := tx.ExecContext(ctx, `UPDATE item_photos SET display_order = ? WHERE id = ? AND item_id = ?`, idx+1, photoID, itemID)
		if err != nil {
			return fmt.Errorf("update display order: %w", err)
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			return fmt.Errorf("photo not found")
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit reorder: %w", err)
	}
	return nil
}

func encodeOriginalImage(path, format string, img image.Image) error {
	tmpPath := path + ".rotate-tmp"
	outFile, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("create rotated original: %w", err)
	}

	switch strings.ToLower(strings.TrimSpace(format)) {
	case "png":
		err = png.Encode(outFile, img)
	default:
		err = jpeg.Encode(outFile, img, &jpeg.Options{Quality: 90})
	}
	closeErr := outFile.Close()
	if err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("encode rotated original: %w", err)
	}
	if closeErr != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("close rotated original: %w", closeErr)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("replace rotated original: %w", err)
	}
	return nil
}

func rotateImage(src image.Image, direction string) *image.RGBA {
	b := src.Bounds()
	w := b.Dx()
	h := b.Dy()
	if w <= 0 || h <= 0 {
		return image.NewRGBA(image.Rect(0, 0, 1, 1))
	}
	dst := image.NewRGBA(image.Rect(0, 0, h, w))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := src.At(b.Min.X+x, b.Min.Y+y)
			if direction == "left" {
				dst.Set(y, w-1-x, c)
			} else {
				dst.Set(h-1-y, x, c)
			}
		}
	}
	return dst
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

func summarizeWorkspaceAssets(assets []WorkspaceAsset) WorkspaceSummary {
	var summary WorkspaceSummary
	for _, asset := range assets {
		summary.Total++
		switch asset.LinkageState {
		case "unlinked":
			summary.Unlinked++
		case "linked_inventory":
			summary.LinkedInventory++
		case "linked_wishlist":
			summary.LinkedWishlist++
		case "linked_both":
			summary.LinkedBoth++
		}
		if asset.AnalysisStatus == "ready" || asset.AnalysisStatus == "pending" {
			summary.ReadyForReview++
		}
	}
	return summary
}

func projectedLinkageState(current, targetType string) string {
	switch targetType {
	case "inventory":
		if current == "linked_wishlist" || current == "linked_both" {
			return "linked_both"
		}
		return "linked_inventory"
	case "wishlist":
		if current == "linked_inventory" || current == "linked_both" {
			return "linked_both"
		}
		return "linked_wishlist"
	default:
		return current
	}
}

func linkageStateForAsset(hasInventoryLink bool, link workspaceLinkState) string {
	inventoryLinked := hasInventoryLink || strings.TrimSpace(link.ItemID) != ""
	wishlistLinked := strings.TrimSpace(link.WishlistID) != ""
	switch {
	case inventoryLinked && wishlistLinked:
		return "linked_both"
	case inventoryLinked:
		return "linked_inventory"
	case wishlistLinked:
		return "linked_wishlist"
	default:
		return "unlinked"
	}
}

var mediaFilenameUnsafe = regexp.MustCompile(`[^a-z0-9]+`)

func friendlyMediaFilename(partNumber, title, filename, assetID string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		ext = ".jpg"
	}
	baseParts := []string{strings.TrimSpace(partNumber), strings.TrimSpace(title)}
	base := strings.TrimSpace(strings.Join(baseParts, " "))
	if base == "" {
		base = strings.TrimSuffix(strings.TrimSpace(filename), filepath.Ext(filename))
	}
	base = strings.ToLower(base)
	base = mediaFilenameUnsafe.ReplaceAllString(base, "-")
	base = strings.Trim(base, "-")
	if base == "" {
		base = "media"
	}
	token := strings.TrimSpace(assetID)
	if len(token) > 8 {
		token = token[:8]
	}
	if token != "" {
		base += "-" + strings.ToLower(token)
	}
	return base + ext
}

func contentTypeForFilename(filename string) string {
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}

func uniqueDownloadName(filename string, used map[string]int) string {
	name := strings.TrimSpace(filename)
	if name == "" {
		name = "media"
	}
	used[name]++
	if used[name] == 1 {
		return name
	}
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	return fmt.Sprintf("%s-%d%s", base, used[name], ext)
}
