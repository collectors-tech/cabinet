package media

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
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
	"time"

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
	ID                  string   `json:"id"`
	Title               string   `json:"title"`
	Filename            string   `json:"filename"`
	UploadedAt          string   `json:"uploaded_at"`
	LinkageState        string   `json:"linkage_state"`
	AnalysisStatus      string   `json:"analysis_status"`
	Source              string   `json:"source"`
	ItemID              string   `json:"item_id,omitempty"`
	WishlistID          string   `json:"wishlist_id,omitempty"`
	ThumbnailURL        string   `json:"thumbnail_url,omitempty"`
	ThumbnailVariations []string `json:"thumbnail_variations,omitempty"`
	Notes               string   `json:"notes,omitempty"`
	DownloadFilename    string   `json:"download_filename"`
	StoredPath          string   `json:"-"`
}

type WorkspaceAssetMetadataUpdate struct {
	Title            string `json:"title"`
	Filename         string `json:"filename"`
	Source           string `json:"source"`
	DownloadFilename string `json:"download_filename"`
	Notes            string `json:"notes"`
}

type WorkspaceAttachment struct {
	ID        string `json:"id"`
	ProfileID string `json:"profile_id"`
	ThreadID  string `json:"thread_id"`
	Filename  string `json:"filename"`
	MimeType  string `json:"mime_type"`
	SizeBytes int64  `json:"size_bytes"`
	Path      string `json:"path"`
	CreatedAt string `json:"created_at"`
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

type LegacyMigrationPreflight struct {
	ProfileID string                  `json:"profile_id"`
	DryRun    bool                    `json:"dry_run"`
	Summary   LegacyMigrationSummary  `json:"summary"`
	Records   []LegacyMigrationRecord `json:"records"`
}

type LegacyMigrationSummary struct {
	Discovered      int `json:"discovered"`
	Pending         int `json:"pending"`
	AlreadyMigrated int `json:"already_migrated"`
	Duplicate       int `json:"duplicate"`
	Missing         int `json:"missing"`
	Orphan          int `json:"orphan"`
	Failed          int `json:"failed"`
}

type LegacyInventoryMigrationResult struct {
	Migrated        int `json:"migrated"`
	AlreadyMigrated int `json:"already_migrated"`
	Skipped         int `json:"skipped"`
	Failed          int `json:"failed"`
}

type LegacyChatAttachmentMigrationResult struct {
	Migrated        int `json:"migrated"`
	AlreadyMigrated int `json:"already_migrated"`
	Skipped         int `json:"skipped"`
	Failed          int `json:"failed"`
}

type LegacyMigrationRecord struct {
	ID             string `json:"id"`
	RecordType     string `json:"record_type"`
	ItemID         string `json:"item_id,omitempty"`
	ThreadID       string `json:"thread_id,omitempty"`
	Filename       string `json:"filename"`
	Classification string `json:"classification"`
	PathClass      string `json:"path_class"`
	RecoveryAction string `json:"recovery_action,omitempty"`
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

type AssetManifest struct {
	Version    int                    `json:"version"`
	AssetID    string                 `json:"asset_id"`
	Files      AssetManifestFiles     `json:"files"`
	Original   AssetManifestOriginal  `json:"original"`
	Renditions []AssetManifestVariant `json:"renditions"`
	Variations []AssetManifestVariant `json:"variations"`
	Owners     []AssetManifestOwner   `json:"owners"`
	Provenance map[string]string      `json:"provenance"`
	CreatedAt  string                 `json:"created_at"`
}

type AssetManifestFiles struct {
	OriginalDir   string `json:"original_dir"`
	RenditionsDir string `json:"renditions_dir"`
	VariationsDir string `json:"variations_dir"`
}

type AssetManifestOriginal struct {
	Filename     string `json:"filename"`
	RelativePath string `json:"relative_path"`
	ContentHash  string `json:"content_hash"`
	MIMEType     string `json:"mime_type"`
	ByteSize     int64  `json:"byte_size"`
	Width        int    `json:"width,omitempty"`
	Height       int    `json:"height,omitempty"`
	Immutable    bool   `json:"immutable"`
}

type AssetManifestVariant struct {
	Name             string `json:"name"`
	RelativePath     string `json:"relative_path"`
	Generator        string `json:"generator"`
	GeneratorVersion string `json:"generator_version"`
}

type AssetManifestOwner struct {
	Type string `json:"type"`
	ID   string `json:"id"`
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

	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		ext = ".jpg"
	}
	safeName := safeMediaFilename(filename)
	if filepath.Ext(safeName) == "" {
		safeName += ext
	}
	origPath, previewPath, thumbPath, err := s.createCanonicalAsset(ctx, rootMediaDir, photoID, safeName, filename, contentTypeForFilename(safeName), r, []AssetManifestOwner{{Type: "inventory_item", ID: itemID}}, map[string]string{"source": "inventory.photo.upload"})
	if err != nil {
		return Photo{}, err
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
	`, photoID, itemID, filename, mediaRootRelativePath(rootMediaDir, origPath), mediaRootRelativePath(rootMediaDir, previewPath), mediaRootRelativePath(rootMediaDir, thumbPath), isPrimary, nextOrder); err != nil {
		_ = os.RemoveAll(filepath.Dir(filepath.Dir(origPath)))
		return Photo{}, fmt.Errorf("insert photo record: %w", err)
	}

	return s.GetByID(ctx, photoID)
}

func (s *Service) SaveWorkspaceAttachment(ctx context.Context, profileID, threadID, filename, mimeType string, src io.Reader) (WorkspaceAttachment, error) {
	profileID = strings.TrimSpace(profileID)
	threadID = strings.TrimSpace(threadID)
	filename = strings.TrimSpace(filename)
	mimeType = strings.TrimSpace(mimeType)
	if profileID == "" || threadID == "" {
		return WorkspaceAttachment{}, fmt.Errorf("profile_id and thread_id are required")
	}
	if filename == "" {
		return WorkspaceAttachment{}, fmt.Errorf("filename is required")
	}
	var threadCount int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM chat_threads WHERE id = ? AND profile_id = ?`, threadID, profileID).Scan(&threadCount); err != nil {
		return WorkspaceAttachment{}, fmt.Errorf("check media upload thread: %w", err)
	}
	if threadCount == 0 {
		return WorkspaceAttachment{}, fmt.Errorf("media upload thread not found")
	}
	mediaRoot, err := s.resolveMediaDirForProfile(ctx, profileID)
	if err != nil {
		return WorkspaceAttachment{}, err
	}
	assetID := uuid.NewString()
	safeName := safeMediaFilename(filename)
	if filepath.Ext(safeName) == "" {
		safeName += ".bin"
	}
	if mimeType == "" {
		mimeType = contentTypeForFilename(safeName)
	}
	origPath, _, _, err := s.createCanonicalAsset(ctx, mediaRoot, assetID, safeName, filename, mimeType, src, []AssetManifestOwner{{Type: "chat_thread", ID: threadID}}, map[string]string{"source": "media.workspace.upload"})
	if err != nil {
		return WorkspaceAttachment{}, err
	}
	info, err := os.Stat(origPath)
	if err != nil {
		_ = os.RemoveAll(filepath.Dir(filepath.Dir(origPath)))
		return WorkspaceAttachment{}, fmt.Errorf("stat workspace asset original: %w", err)
	}
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO chat_attachments(id, profile_id, thread_id, filename, mime_type, size_bytes, stored_path)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, assetID, profileID, threadID, filename, mimeType, info.Size(), mediaRootRelativePath(mediaRoot, origPath)); err != nil {
		_ = os.RemoveAll(filepath.Dir(filepath.Dir(origPath)))
		return WorkspaceAttachment{}, fmt.Errorf("save workspace media attachment: %w", err)
	}
	return s.getWorkspaceAttachment(ctx, profileID, assetID)
}

func (s *Service) getWorkspaceAttachment(ctx context.Context, profileID, attachmentID string) (WorkspaceAttachment, error) {
	var a WorkspaceAttachment
	err := s.db.QueryRowContext(ctx, `
		SELECT id, profile_id, thread_id, filename, mime_type, size_bytes, stored_path, created_at
		FROM chat_attachments
		WHERE id = ? AND profile_id = ?
	`, strings.TrimSpace(attachmentID), strings.TrimSpace(profileID)).Scan(&a.ID, &a.ProfileID, &a.ThreadID, &a.Filename, &a.MimeType, &a.SizeBytes, &a.Path, &a.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return WorkspaceAttachment{}, fmt.Errorf("workspace attachment not found")
		}
		return WorkspaceAttachment{}, fmt.Errorf("get workspace attachment: %w", err)
	}
	mediaRoot, err := s.resolveMediaDirForProfile(ctx, a.ProfileID)
	if err != nil {
		return WorkspaceAttachment{}, err
	}
	a.Path = resolveMediaRootPath(mediaRoot, a.Path)
	return a, nil
}

func (s *Service) createCanonicalAsset(ctx context.Context, mediaRoot, assetID, safeName, originalFilename, mimeType string, src io.Reader, owners []AssetManifestOwner, provenance map[string]string) (string, string, string, error) {
	if err := ctx.Err(); err != nil {
		return "", "", "", err
	}
	assetsRoot := filepath.Join(mediaRoot, "assets")
	stagingRoot := filepath.Join(mediaRoot, ".staging")
	if err := os.MkdirAll(stagingRoot, 0o755); err != nil {
		return "", "", "", fmt.Errorf("create asset staging root: %w", err)
	}
	stagingDir := filepath.Join(stagingRoot, assetID)
	assetDir := filepath.Join(assetsRoot, assetID)
	_ = os.RemoveAll(stagingDir)
	if err := os.MkdirAll(filepath.Join(stagingDir, "original"), 0o755); err != nil {
		return "", "", "", fmt.Errorf("create asset original dir: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(stagingDir, "renditions"), 0o755); err != nil {
		_ = os.RemoveAll(stagingDir)
		return "", "", "", fmt.Errorf("create asset renditions dir: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(stagingDir, "variations"), 0o755); err != nil {
		_ = os.RemoveAll(stagingDir)
		return "", "", "", fmt.Errorf("create asset variations dir: %w", err)
	}

	origPath := filepath.Join(stagingDir, "original", safeName)
	origFile, err := os.Create(origPath)
	if err != nil {
		_ = os.RemoveAll(stagingDir)
		return "", "", "", fmt.Errorf("create original file: %w", err)
	}
	hasher := sha256.New()
	size, copyErr := io.Copy(io.MultiWriter(origFile, hasher), src)
	closeErr := origFile.Close()
	if copyErr != nil {
		_ = os.RemoveAll(stagingDir)
		return "", "", "", fmt.Errorf("save original file: %w", copyErr)
	}
	if closeErr != nil {
		_ = os.RemoveAll(stagingDir)
		return "", "", "", fmt.Errorf("close original file: %w", closeErr)
	}

	previewPath := filepath.Join(stagingDir, "renditions", "preview.jpg")
	thumbPath := filepath.Join(stagingDir, "renditions", "thumbnail.jpg")
	renditions := []AssetManifestVariant{}
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(mimeType)), "image/") {
		if err := generateScaledJPEG(origPath, previewPath, 1024); err != nil {
			_ = os.RemoveAll(stagingDir)
			return "", "", "", fmt.Errorf("generate preview: %w", err)
		}
		if err := generateScaledJPEG(origPath, thumbPath, 256); err != nil {
			_ = os.RemoveAll(stagingDir)
			return "", "", "", fmt.Errorf("generate thumbnail: %w", err)
		}
		renditions = []AssetManifestVariant{
			{Name: "preview", RelativePath: "renditions/preview.jpg", Generator: "cabinet.media.generateScaledJPEG", GeneratorVersion: "1"},
			{Name: "thumbnail", RelativePath: "renditions/thumbnail.jpg", Generator: "cabinet.media.generateScaledJPEG", GeneratorVersion: "1"},
		}
	}

	width, height := imageDimensions(origPath)
	manifest := AssetManifest{
		Version: 1,
		AssetID: assetID,
		Files: AssetManifestFiles{
			OriginalDir:   "original",
			RenditionsDir: "renditions",
			VariationsDir: "variations",
		},
		Original: AssetManifestOriginal{
			Filename:     originalFilename,
			RelativePath: filepath.ToSlash(filepath.Join("original", safeName)),
			ContentHash:  "sha256:" + hex.EncodeToString(hasher.Sum(nil)),
			MIMEType:     mimeType,
			ByteSize:     size,
			Width:        width,
			Height:       height,
			Immutable:    true,
		},
		Renditions: renditions,
		Variations: []AssetManifestVariant{},
		Owners:     owners,
		Provenance: provenance,
		CreatedAt:  time.Now().UTC().Format(time.RFC3339),
	}
	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		_ = os.RemoveAll(stagingDir)
		return "", "", "", fmt.Errorf("marshal asset manifest: %w", err)
	}
	if err := os.WriteFile(filepath.Join(stagingDir, "manifest.json"), manifestBytes, 0o644); err != nil {
		_ = os.RemoveAll(stagingDir)
		return "", "", "", fmt.Errorf("write asset manifest: %w", err)
	}
	_ = os.Chmod(origPath, 0o444)

	if err := os.MkdirAll(assetsRoot, 0o755); err != nil {
		_ = os.RemoveAll(stagingDir)
		return "", "", "", fmt.Errorf("create assets root: %w", err)
	}
	_ = os.RemoveAll(assetDir)
	if err := os.Rename(stagingDir, assetDir); err != nil {
		_ = os.RemoveAll(stagingDir)
		return "", "", "", fmt.Errorf("promote asset folder: %w", err)
	}
	return filepath.Join(assetDir, "original", safeName), filepath.Join(assetDir, "renditions", "preview.jpg"), filepath.Join(assetDir, "renditions", "thumbnail.jpg"), nil
}

func imageDimensions(path string) (int, int) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0
	}
	defer f.Close()
	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		return 0, 0
	}
	return cfg.Width, cfg.Height
}

func safeMediaFilename(filename string) string {
	filename = filepath.Base(strings.TrimSpace(filename))
	if filename == "." || filename == string(filepath.Separator) || filename == "" {
		filename = "original"
	}
	safe := strings.Map(func(r rune) rune {
		if r == '/' || r == '\\' || r == ':' || r == '*' || r == '?' || r == '"' || r == '<' || r == '>' || r == '|' || r < 32 {
			return '_'
		}
		return r
	}, filename)
	safe = strings.TrimSpace(safe)
	if safe == "" {
		return "original"
	}
	return safe
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
	metadata, err := s.loadWorkspaceMetadata(ctx, profileID)
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
		originalPath = s.resolveStoredPhotoPath(ctx, itemID, originalPath)
		asset := WorkspaceAsset{
			ID:                  assetID,
			Title:               displayTitle,
			Filename:            filename,
			UploadedAt:          createdAt,
			LinkageState:        linkageState,
			AnalysisStatus:      "not_analyzed",
			Source:              "Inventory photo",
			ItemID:              itemID,
			WishlistID:          link.WishlistID,
			ThumbnailURL:        "/api/items/" + itemID + "/photos/" + assetID + "/file?variant=thumbnail",
			ThumbnailVariations: []string{"Original", "Preview", "Thumbnail"},
			DownloadFilename:    friendlyMediaFilename(partNumber, displayTitle, filename, assetID),
			StoredPath:          originalPath,
		}
		applyWorkspaceMetadata(&asset, metadata[assetID])
		assets = append(assets, asset)
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
		mediaRoot, err := s.resolveMediaDirForProfile(ctx, profileID)
		if err != nil {
			attachmentRows.Close()
			return WorkspaceList{}, err
		}
		storedPath = resolveMediaRootPath(mediaRoot, storedPath)
		asset := WorkspaceAsset{
			ID:                  assetID,
			Title:               strings.TrimSpace(filename),
			Filename:            filename,
			UploadedAt:          createdAt,
			LinkageState:        linkageStateForAsset(false, link),
			AnalysisStatus:      "pending",
			Source:              "Chat attachment",
			ItemID:              link.ItemID,
			WishlistID:          link.WishlistID,
			ThumbnailVariations: []string{"Original", "Thumbnail", "Review crop"},
			DownloadFilename:    friendlyMediaFilename("", filename, filename, assetID),
			StoredPath:          storedPath,
		}
		applyWorkspaceMetadata(&asset, metadata[assetID])
		assets = append(assets, asset)
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

func (s *Service) PreflightLegacyMediaMigration(ctx context.Context, profileID string) (LegacyMigrationPreflight, error) {
	profileID = strings.TrimSpace(profileID)
	if profileID == "" {
		return LegacyMigrationPreflight{}, fmt.Errorf("profile_id is required")
	}
	mediaRoot, err := s.resolveMediaDirForProfile(ctx, profileID)
	if err != nil {
		return LegacyMigrationPreflight{}, err
	}

	report := LegacyMigrationPreflight{ProfileID: profileID, DryRun: true}
	referenced := make(map[string]bool)
	type preflightRecord struct {
		record  LegacyMigrationRecord
		pathKey string
	}
	records := make([]preflightRecord, 0)

	photoRows, err := s.db.QueryContext(ctx, `
		SELECT ip.id, ip.item_id, ip.filename, ip.original_path, ci.id
		FROM item_photos ip
		INNER JOIN canonical_items ci ON ci.id = ip.item_id
		WHERE ci.profile_id = ?
		ORDER BY ip.id ASC
	`, profileID)
	if err != nil {
		return LegacyMigrationPreflight{}, fmt.Errorf("preflight inventory photos: %w", err)
	}
	for photoRows.Next() {
		var record LegacyMigrationRecord
		var originalPath, itemIDCheck string
		if err := photoRows.Scan(&record.ID, &record.ItemID, &record.Filename, &originalPath, &itemIDCheck); err != nil {
			photoRows.Close()
			return LegacyMigrationPreflight{}, fmt.Errorf("scan inventory photo preflight: %w", err)
		}
		record.RecordType = "inventory_photo"
		resolvedPath := resolveMediaRootPath(mediaRoot, originalPath)
		markReferencedPath(referenced, resolvedPath)
		classifyLegacyMigrationPath(mediaRoot, resolvedPath, &record)
		records = append(records, preflightRecord{record: record, pathKey: canonicalPathKey(resolvedPath)})
	}
	if err := photoRows.Err(); err != nil {
		photoRows.Close()
		return LegacyMigrationPreflight{}, fmt.Errorf("iterate inventory photo preflight: %w", err)
	}
	photoRows.Close()

	attachmentRows, err := s.db.QueryContext(ctx, `
		SELECT id, thread_id, filename, stored_path
		FROM chat_attachments
		WHERE profile_id = ?
		ORDER BY id ASC
	`, profileID)
	if err != nil {
		return LegacyMigrationPreflight{}, fmt.Errorf("preflight chat attachments: %w", err)
	}
	for attachmentRows.Next() {
		var record LegacyMigrationRecord
		var storedPath string
		if err := attachmentRows.Scan(&record.ID, &record.ThreadID, &record.Filename, &storedPath); err != nil {
			attachmentRows.Close()
			return LegacyMigrationPreflight{}, fmt.Errorf("scan chat attachment preflight: %w", err)
		}
		record.RecordType = "chat_attachment"
		resolvedPath := resolveMediaRootPath(mediaRoot, storedPath)
		markReferencedPath(referenced, resolvedPath)
		classifyLegacyMigrationPath(mediaRoot, resolvedPath, &record)
		records = append(records, preflightRecord{record: record, pathKey: canonicalPathKey(resolvedPath)})
	}
	if err := attachmentRows.Err(); err != nil {
		attachmentRows.Close()
		return LegacyMigrationPreflight{}, fmt.Errorf("iterate chat attachment preflight: %w", err)
	}
	attachmentRows.Close()

	duplicateCounts := make(map[string]int)
	for _, entry := range records {
		if entry.pathKey == "" || entry.record.Classification != "pending" {
			continue
		}
		duplicateCounts[entry.pathKey]++
	}
	for _, entry := range records {
		if duplicateCounts[entry.pathKey] > 1 {
			entry.record.Classification = "duplicate"
			entry.record.RecoveryAction = "Resolve duplicate source references before applying migration."
		}
		report.addLegacyMigrationRecord(entry.record)
	}

	if err := report.addLegacyMediaOrphans(mediaRoot, referenced); err != nil {
		return LegacyMigrationPreflight{}, err
	}
	return report, nil
}

func (s *Service) ApplyLegacyInventoryPhotoMigration(ctx context.Context, profileID string) (LegacyInventoryMigrationResult, error) {
	profileID = strings.TrimSpace(profileID)
	if profileID == "" {
		return LegacyInventoryMigrationResult{}, fmt.Errorf("profile_id is required")
	}
	mediaRoot, err := s.resolveMediaDirForProfile(ctx, profileID)
	if err != nil {
		return LegacyInventoryMigrationResult{}, err
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT ip.id, ip.item_id, ip.filename, ip.original_path, ip.preview_path, ip.thumbnail_path
		FROM item_photos ip
		INNER JOIN canonical_items ci ON ci.id = ip.item_id
		WHERE ci.profile_id = ?
		ORDER BY ip.id ASC
	`, profileID)
	if err != nil {
		return LegacyInventoryMigrationResult{}, fmt.Errorf("list inventory photos for migration: %w", err)
	}
	defer rows.Close()

	type photoMigrationRow struct {
		id            string
		itemID        string
		filename      string
		originalPath  string
		previewPath   string
		thumbnailPath string
	}
	pending := make([]photoMigrationRow, 0)
	resumable := make([]photoMigrationRow, 0)
	result := LegacyInventoryMigrationResult{}
	for rows.Next() {
		var row photoMigrationRow
		if err := rows.Scan(&row.id, &row.itemID, &row.filename, &row.originalPath, &row.previewPath, &row.thumbnailPath); err != nil {
			return LegacyInventoryMigrationResult{}, fmt.Errorf("scan inventory photo migration row: %w", err)
		}
		resolvedOriginal := resolveMediaRootPath(mediaRoot, row.originalPath)
		record := LegacyMigrationRecord{ID: row.id, RecordType: "inventory_photo", ItemID: row.itemID, Filename: row.filename}
		classifyLegacyMigrationPath(mediaRoot, resolvedOriginal, &record)
		switch record.Classification {
		case "pending":
			pending = append(pending, row)
		case "already_migrated":
			result.AlreadyMigrated++
		case "missing", "failed":
			ok, err := promotedCanonicalAssetExists(mediaRoot, row.id)
			if err != nil {
				return LegacyInventoryMigrationResult{}, err
			}
			if ok {
				resumable = append(resumable, row)
			} else {
				result.Skipped++
			}
		default:
			result.Skipped++
		}
	}
	if err := rows.Err(); err != nil {
		return LegacyInventoryMigrationResult{}, fmt.Errorf("iterate inventory photo migration rows: %w", err)
	}

	for _, row := range pending {
		if err := s.migrateLegacyInventoryPhoto(ctx, mediaRoot, row.id, row.itemID, row.filename, row.originalPath); err != nil {
			result.Failed++
			return result, err
		}
		result.Migrated++
	}
	for _, row := range resumable {
		if err := s.resumeLegacyInventoryPhotoMigration(ctx, mediaRoot, row.id, row.itemID); err != nil {
			result.Failed++
			return result, err
		}
		result.Migrated++
	}
	return result, nil
}

func (s *Service) ApplyLegacyChatAttachmentMigration(ctx context.Context, profileID string) (LegacyChatAttachmentMigrationResult, error) {
	profileID = strings.TrimSpace(profileID)
	if profileID == "" {
		return LegacyChatAttachmentMigrationResult{}, fmt.Errorf("profile_id is required")
	}
	mediaRoot, err := s.resolveMediaDirForProfile(ctx, profileID)
	if err != nil {
		return LegacyChatAttachmentMigrationResult{}, err
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, thread_id, filename, mime_type, stored_path
		FROM chat_attachments
		WHERE profile_id = ?
		ORDER BY id ASC
	`, profileID)
	if err != nil {
		return LegacyChatAttachmentMigrationResult{}, fmt.Errorf("list chat attachments for migration: %w", err)
	}
	defer rows.Close()

	type attachmentMigrationRow struct {
		id         string
		threadID   string
		filename   string
		mimeType   string
		storedPath string
	}
	pending := make([]attachmentMigrationRow, 0)
	resumable := make([]attachmentMigrationRow, 0)
	result := LegacyChatAttachmentMigrationResult{}
	for rows.Next() {
		var row attachmentMigrationRow
		if err := rows.Scan(&row.id, &row.threadID, &row.filename, &row.mimeType, &row.storedPath); err != nil {
			return LegacyChatAttachmentMigrationResult{}, fmt.Errorf("scan chat attachment migration row: %w", err)
		}
		resolvedPath := resolveMediaRootPath(mediaRoot, row.storedPath)
		record := LegacyMigrationRecord{ID: row.id, RecordType: "chat_attachment", ThreadID: row.threadID, Filename: row.filename}
		classifyLegacyMigrationPath(mediaRoot, resolvedPath, &record)
		switch record.Classification {
		case "pending":
			pending = append(pending, row)
		case "already_migrated":
			result.AlreadyMigrated++
		case "missing", "failed":
			ok, err := promotedCanonicalAssetExists(mediaRoot, row.id)
			if err != nil {
				return LegacyChatAttachmentMigrationResult{}, err
			}
			if ok {
				resumable = append(resumable, row)
			} else {
				result.Skipped++
			}
		default:
			result.Skipped++
		}
	}
	if err := rows.Err(); err != nil {
		return LegacyChatAttachmentMigrationResult{}, fmt.Errorf("iterate chat attachment migration rows: %w", err)
	}

	for _, row := range pending {
		if err := s.migrateLegacyChatAttachment(ctx, mediaRoot, row.id, row.threadID, row.filename, row.mimeType, row.storedPath); err != nil {
			result.Failed++
			return result, err
		}
		result.Migrated++
	}
	for _, row := range resumable {
		if err := s.resumeLegacyChatAttachmentMigration(ctx, mediaRoot, row.id); err != nil {
			result.Failed++
			return result, err
		}
		result.Migrated++
	}
	return result, nil
}

func (s *Service) UpdateWorkspaceAssetMetadata(ctx context.Context, profileID, assetID string, update WorkspaceAssetMetadataUpdate) (WorkspaceAsset, error) {
	profileID = strings.TrimSpace(profileID)
	assetID = strings.TrimSpace(assetID)
	if profileID == "" || assetID == "" {
		return WorkspaceAsset{}, fmt.Errorf("profile_id and asset_id are required")
	}
	update.Title = strings.TrimSpace(update.Title)
	update.Filename = strings.TrimSpace(update.Filename)
	update.Source = strings.TrimSpace(update.Source)
	update.DownloadFilename = strings.TrimSpace(update.DownloadFilename)
	update.Notes = strings.TrimSpace(update.Notes)
	if update.Title == "" || update.Filename == "" || update.Source == "" || update.DownloadFilename == "" {
		return WorkspaceAsset{}, fmt.Errorf("title, filename, source, and download_filename are required")
	}

	assetType, err := s.assetType(ctx, profileID, assetID)
	if err != nil {
		return WorkspaceAsset{}, err
	}
	switch assetType {
	case "chat_attachment":
		if _, err := s.db.ExecContext(ctx, `
			UPDATE chat_attachments
			SET filename = ?
			WHERE id = ? AND profile_id = ?
		`, update.Filename, assetID, profileID); err != nil {
			return WorkspaceAsset{}, fmt.Errorf("update chat attachment filename: %w", err)
		}
	case "item_photo":
		if _, err := s.db.ExecContext(ctx, `
			UPDATE item_photos
			SET filename = ?
			WHERE id = ? AND item_id IN (SELECT id FROM canonical_items WHERE profile_id = ?)
		`, update.Filename, assetID, profileID); err != nil {
			return WorkspaceAsset{}, fmt.Errorf("update inventory photo filename: %w", err)
		}
	default:
		return WorkspaceAsset{}, fmt.Errorf("unsupported media asset type")
	}

	threadID, err := s.ensureWorkspaceMetadataThread(ctx, profileID)
	if err != nil {
		return WorkspaceAsset{}, err
	}
	contextJSON, err := json.Marshal(map[string]any{
		"source":            "media.workspace",
		"metadata_flow":     "edit-media-dialog",
		"asset_id":          assetID,
		"title":             update.Title,
		"origin":            update.Source,
		"filename":          update.Filename,
		"download_filename": update.DownloadFilename,
		"notes":             update.Notes,
	})
	if err != nil {
		return WorkspaceAsset{}, fmt.Errorf("marshal media metadata: %w", err)
	}
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO chat_messages(id, profile_id, thread_id, role, content, attachments_json, context_json)
		VALUES (?, ?, ?, 'user', 'Media asset metadata updated from Media workspace.', '[]', ?)
	`, uuid.NewString(), profileID, threadID, string(contextJSON)); err != nil {
		return WorkspaceAsset{}, fmt.Errorf("persist media metadata update: %w", err)
	}

	list, err := s.ListWorkspaceAssets(ctx, profileID, "all")
	if err != nil {
		return WorkspaceAsset{}, err
	}
	for _, asset := range list.Assets {
		if asset.ID == assetID {
			return asset, nil
		}
	}
	return WorkspaceAsset{}, fmt.Errorf("updated media asset not found")
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

type workspaceMetadata struct {
	Title            string `json:"title"`
	Source           string `json:"origin"`
	DownloadFilename string `json:"download_filename"`
	Notes            string `json:"notes"`
}

func (s *Service) loadWorkspaceMetadata(ctx context.Context, profileID string) (map[string]workspaceMetadata, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT context_json
		FROM chat_messages
		WHERE profile_id = ?
		  AND content IN (
			'Media asset added from Media workspace.',
			'Media asset metadata updated from Media workspace.'
		  )
		ORDER BY created_at ASC, rowid ASC
	`, profileID)
	if err != nil {
		return nil, fmt.Errorf("list media metadata messages: %w", err)
	}
	defer rows.Close()

	out := map[string]workspaceMetadata{}
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, fmt.Errorf("scan media metadata: %w", err)
		}
		var payload struct {
			AssetID          string `json:"asset_id"`
			Title            string `json:"title"`
			Source           string `json:"origin"`
			DownloadFilename string `json:"download_filename"`
			Notes            string `json:"notes"`
		}
		if err := json.Unmarshal([]byte(raw), &payload); err != nil {
			continue
		}
		assetID := strings.TrimSpace(payload.AssetID)
		if assetID == "" {
			continue
		}
		out[assetID] = workspaceMetadata{
			Title:            strings.TrimSpace(payload.Title),
			Source:           strings.TrimSpace(payload.Source),
			DownloadFilename: strings.TrimSpace(payload.DownloadFilename),
			Notes:            strings.TrimSpace(payload.Notes),
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate media metadata: %w", err)
	}
	return out, nil
}

func applyWorkspaceMetadata(asset *WorkspaceAsset, metadata workspaceMetadata) {
	if metadata.Title != "" {
		asset.Title = metadata.Title
	}
	if metadata.Source != "" {
		asset.Source = metadata.Source
	}
	if metadata.DownloadFilename != "" {
		asset.DownloadFilename = metadata.DownloadFilename
	}
	if metadata.Notes != "" {
		asset.Notes = metadata.Notes
	}
}

func (s *Service) ensureWorkspaceMetadataThread(ctx context.Context, profileID string) (string, error) {
	var threadID string
	err := s.db.QueryRowContext(ctx, `
		SELECT id
		FROM chat_threads
		WHERE profile_id = ? AND title = 'Media Uploads'
		ORDER BY created_at ASC, id ASC
		LIMIT 1
	`, profileID).Scan(&threadID)
	if err == nil {
		return threadID, nil
	}
	if err != sql.ErrNoRows {
		return "", fmt.Errorf("find media metadata thread: %w", err)
	}
	threadID = uuid.NewString()
	metadataJSON, err := json.Marshal(map[string]any{"source": "media.workspace"})
	if err != nil {
		return "", fmt.Errorf("marshal media thread metadata: %w", err)
	}
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO chat_threads(id, profile_id, title, metadata_json)
		VALUES (?, ?, 'Media Uploads', ?)
	`, threadID, profileID, string(metadataJSON)); err != nil {
		return "", fmt.Errorf("create media metadata thread: %w", err)
	}
	return threadID, nil
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

func (s *Service) resolveMediaDirForProfile(ctx context.Context, profileID string) (string, error) {
	profileID = strings.TrimSpace(profileID)
	if profileID == "" {
		return "", fmt.Errorf("profile_id is required")
	}

	var configuredMediaDir sql.NullString
	err := s.db.QueryRowContext(
		ctx,
		`
		SELECT ps.value
		FROM profiles p
		LEFT JOIN profile_settings ps
			ON ps.profile_id = p.id
			AND ps.key = 'storage.media_dir'
		WHERE p.id = ?
		`,
		profileID,
	).Scan(&configuredMediaDir)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("profile not found")
		}
		return "", fmt.Errorf("resolve profile media dir: %w", err)
	}

	mediaDir := strings.TrimSpace(configuredMediaDir.String)
	if mediaDir != "" {
		return mediaDir, nil
	}

	return s.mediaDir, nil
}

func (s *Service) resolvePhotoPaths(ctx context.Context, p *Photo) error {
	mediaRoot, err := s.resolveMediaDirForItem(ctx, p.ItemID)
	if err != nil {
		return err
	}
	p.OriginalPath = resolveMediaRootPath(mediaRoot, p.OriginalPath)
	p.PreviewPath = resolveMediaRootPath(mediaRoot, p.PreviewPath)
	p.ThumbnailPath = resolveMediaRootPath(mediaRoot, p.ThumbnailPath)
	return nil
}

func (s *Service) resolveStoredPhotoPath(ctx context.Context, itemID, storedPath string) string {
	mediaRoot, err := s.resolveMediaDirForItem(ctx, itemID)
	if err != nil {
		return storedPath
	}
	return resolveMediaRootPath(mediaRoot, storedPath)
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
		if err := s.resolvePhotoPaths(ctx, &p); err != nil {
			return nil, err
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
	if err := s.resolvePhotoPaths(ctx, &p); err != nil {
		return Photo{}, err
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
	if err := s.cleanupDeletedPhotoFiles(ctx, itemID, p); err != nil {
		return err
	}

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

func (s *Service) cleanupDeletedPhotoFiles(ctx context.Context, itemID string, p Photo) error {
	mediaRoot, err := s.resolveMediaDirForItem(ctx, itemID)
	if err != nil {
		return err
	}
	assetDir := filepath.Join(mediaRoot, "assets", p.ID)
	if pathWithinDir(mediaRoot, assetDir) {
		if _, err := os.Stat(filepath.Join(assetDir, "manifest.json")); err == nil {
			references, err := s.countAssetReferences(ctx, p.ID)
			if err != nil {
				return err
			}
			if references > 0 {
				return nil
			}
			if err := os.RemoveAll(assetDir); err != nil {
				return fmt.Errorf("remove orphan asset folder: %w", err)
			}
			return nil
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("stat asset manifest: %w", err)
		}
	}

	_ = os.Remove(p.OriginalPath)
	_ = os.Remove(p.PreviewPath)
	_ = os.Remove(p.ThumbnailPath)
	return nil
}

func (s *Service) countAssetReferences(ctx context.Context, assetID string) (int, error) {
	var photoRefs, attachmentRefs, linkRefs int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM item_photos WHERE id = ?`, assetID).Scan(&photoRefs); err != nil {
		return 0, fmt.Errorf("count item photo asset references: %w", err)
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM chat_attachments WHERE id = ?`, assetID).Scan(&attachmentRefs); err != nil {
		return 0, fmt.Errorf("count chat attachment asset references: %w", err)
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM media_asset_links WHERE asset_id = ?`, assetID).Scan(&linkRefs); err != nil {
		return 0, fmt.Errorf("count media asset link references: %w", err)
	}
	return photoRefs + attachmentRefs + linkRefs, nil
}

func pathWithinDir(root, candidate string) bool {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return false
	}
	candidateAbs, err := filepath.Abs(candidate)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(rootAbs, candidateAbs)
	if err != nil {
		return false
	}
	return rel != "." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && rel != ".."
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
	originalMode := os.FileMode(0)
	if info, err := os.Stat(path); err == nil {
		originalMode = info.Mode().Perm()
		if originalMode&0o200 == 0 {
			_ = os.Chmod(path, originalMode|0o200)
		}
	}
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
	if originalMode != 0 && originalMode&0o200 == 0 {
		_ = os.Chmod(path, originalMode)
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

func (r *LegacyMigrationPreflight) addLegacyMigrationRecord(record LegacyMigrationRecord) {
	r.Records = append(r.Records, record)
	r.Summary.Discovered++
	switch record.Classification {
	case "pending":
		r.Summary.Pending++
	case "already_migrated":
		r.Summary.AlreadyMigrated++
	case "duplicate":
		r.Summary.Duplicate++
	case "missing":
		r.Summary.Missing++
	case "orphan":
		r.Summary.Orphan++
	case "failed":
		r.Summary.Failed++
	}
}

func (r *LegacyMigrationPreflight) addLegacyMediaOrphans(mediaRoot string, referenced map[string]bool) error {
	if strings.TrimSpace(mediaRoot) == "" {
		return nil
	}
	if _, err := os.Stat(mediaRoot); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("stat media root for orphan audit: %w", err)
	}
	return filepath.WalkDir(mediaRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("walk media orphan audit: %w", err)
		}
		if path == mediaRoot {
			return nil
		}
		name := d.Name()
		if d.IsDir() && (name == "assets" || name == ".staging") {
			return filepath.SkipDir
		}
		if d.IsDir() {
			return nil
		}
		if referenced[canonicalPathKey(path)] {
			return nil
		}
		r.addLegacyMigrationRecord(LegacyMigrationRecord{
			ID:             "orphan:" + filepath.ToSlash(mediaRootRelativePath(mediaRoot, path)),
			RecordType:     "orphan_file",
			Filename:       filepath.Base(path),
			Classification: "orphan",
			PathClass:      "legacy_media",
			RecoveryAction: "Review manually; migration will not delete orphan files.",
		})
		return nil
	})
}

func classifyLegacyMigrationPath(mediaRoot, resolvedPath string, record *LegacyMigrationRecord) {
	switch {
	case isCanonicalMediaPath(mediaRoot, resolvedPath):
		record.PathClass = "canonical_asset"
		if err := validateCanonicalMigrationPath(mediaRoot, resolvedPath); err != nil {
			record.Classification = "failed"
			record.RecoveryAction = "Repair canonical asset manifest/original before migration."
		} else {
			record.Classification = "already_migrated"
		}
	case strings.TrimSpace(resolvedPath) == "":
		record.Classification = "missing"
		record.PathClass = "empty"
		record.RecoveryAction = "Restore or relink the missing source before migration."
	default:
		record.PathClass = legacyMigrationPathClass(mediaRoot, resolvedPath)
		if info, err := os.Stat(resolvedPath); err == nil {
			if err := validateReadableLegacyMigrationSource(resolvedPath, info); err != nil {
				record.Classification = "failed"
				record.RecoveryAction = "Resolve file access error and retry preflight."
			} else {
				record.Classification = "pending"
			}
		} else if os.IsNotExist(err) {
			record.Classification = "missing"
			record.RecoveryAction = "Restore or relink the missing source before migration."
		} else {
			record.Classification = "failed"
			record.RecoveryAction = "Resolve file access error and retry preflight."
		}
	}
}

func validateReadableLegacyMigrationSource(path string, info os.FileInfo) error {
	if info == nil {
		return fmt.Errorf("legacy source metadata is unavailable")
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("legacy source is not a regular file")
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	return f.Close()
}

func validateCanonicalMigrationPath(mediaRoot, resolvedPath string) error {
	assetID, ok := canonicalAssetIDForPath(mediaRoot, resolvedPath)
	if !ok {
		return fmt.Errorf("canonical asset id not found")
	}
	_, err := readPromotedCanonicalManifest(mediaRoot, assetID)
	return err
}

func canonicalAssetIDForPath(mediaRoot, path string) (string, bool) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", false
	}
	if strings.HasPrefix(filepath.ToSlash(path), "assets/") {
		parts := strings.Split(filepath.ToSlash(path), "/")
		if len(parts) >= 2 && strings.TrimSpace(parts[1]) != "" {
			return parts[1], true
		}
		return "", false
	}
	assetRoot := filepath.Join(mediaRoot, "assets")
	if !pathWithinDir(assetRoot, path) {
		return "", false
	}
	rel, err := filepath.Rel(assetRoot, path)
	if err != nil {
		return "", false
	}
	parts := strings.Split(filepath.ToSlash(rel), "/")
	if len(parts) == 0 || strings.TrimSpace(parts[0]) == "" || parts[0] == "." || strings.HasPrefix(parts[0], "..") {
		return "", false
	}
	return parts[0], true
}

func (s *Service) migrateLegacyInventoryPhoto(ctx context.Context, mediaRoot, photoID, itemID, filename, storedOriginalPath string) error {
	sourcePath := resolveMediaRootPath(mediaRoot, storedOriginalPath)
	sourceHash, err := fileSHA256(sourcePath)
	if err != nil {
		return fmt.Errorf("hash legacy inventory photo %s: %w", photoID, err)
	}
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("open legacy inventory photo %s: %w", photoID, err)
	}
	defer sourceFile.Close()

	safeName := safeMediaFilename(filename)
	if filepath.Ext(safeName) == "" {
		safeName += filepath.Ext(sourcePath)
	}
	if filepath.Ext(safeName) == "" {
		safeName += ".jpg"
	}
	origPath, previewPath, thumbPath, err := s.createCanonicalAsset(
		ctx,
		mediaRoot,
		photoID,
		safeName,
		filename,
		contentTypeForFilename(safeName),
		sourceFile,
		[]AssetManifestOwner{{Type: "inventory_item", ID: itemID}},
		map[string]string{
			"source":              "legacy.inventory_photo.migration",
			"legacy_source_class": legacyMigrationPathClass(mediaRoot, sourcePath),
			"legacy_source_hash":  sourceHash,
		},
	)
	if err != nil {
		return fmt.Errorf("create canonical asset for legacy inventory photo %s: %w", photoID, err)
	}

	manifest, err := readAssetManifestFile(filepath.Join(mediaRoot, "assets", photoID, "manifest.json"))
	if err != nil {
		_ = os.RemoveAll(filepath.Join(mediaRoot, "assets", photoID))
		return fmt.Errorf("read migrated manifest for %s: %w", photoID, err)
	}
	if manifest.Original.ContentHash != sourceHash {
		_ = os.RemoveAll(filepath.Join(mediaRoot, "assets", photoID))
		return fmt.Errorf("verify migrated original hash for %s: got %s want %s", photoID, manifest.Original.ContentHash, sourceHash)
	}

	if _, err := s.db.ExecContext(ctx, `
		UPDATE item_photos
		SET original_path = ?, preview_path = ?, thumbnail_path = ?
		WHERE id = ? AND item_id = ?
	`, mediaRootRelativePath(mediaRoot, origPath), mediaRootRelativePath(mediaRoot, previewPath), mediaRootRelativePath(mediaRoot, thumbPath), photoID, itemID); err != nil {
		_ = os.RemoveAll(filepath.Join(mediaRoot, "assets", photoID))
		return fmt.Errorf("update migrated inventory photo %s: %w", photoID, err)
	}
	return nil
}

func (s *Service) migrateLegacyChatAttachment(ctx context.Context, mediaRoot, attachmentID, threadID, filename, mimeType, storedPath string) error {
	sourcePath := resolveMediaRootPath(mediaRoot, storedPath)
	sourceHash, err := fileSHA256(sourcePath)
	if err != nil {
		return fmt.Errorf("hash legacy chat attachment %s: %w", attachmentID, err)
	}
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("open legacy chat attachment %s: %w", attachmentID, err)
	}
	defer sourceFile.Close()

	safeName := safeMediaFilename(filename)
	if filepath.Ext(safeName) == "" {
		safeName += filepath.Ext(sourcePath)
	}
	if strings.TrimSpace(mimeType) == "" {
		mimeType = contentTypeForFilename(safeName)
	}
	origPath, _, _, err := s.createCanonicalAsset(
		ctx,
		mediaRoot,
		attachmentID,
		safeName,
		filename,
		mimeType,
		sourceFile,
		[]AssetManifestOwner{{Type: "chat_thread", ID: threadID}},
		map[string]string{
			"source":              "legacy.chat_attachment.migration",
			"legacy_source_class": legacyMigrationPathClass(mediaRoot, sourcePath),
			"legacy_source_hash":  sourceHash,
		},
	)
	if err != nil {
		return fmt.Errorf("create canonical asset for legacy chat attachment %s: %w", attachmentID, err)
	}

	manifest, err := readAssetManifestFile(filepath.Join(mediaRoot, "assets", attachmentID, "manifest.json"))
	if err != nil {
		_ = os.RemoveAll(filepath.Join(mediaRoot, "assets", attachmentID))
		return fmt.Errorf("read migrated chat attachment manifest for %s: %w", attachmentID, err)
	}
	if manifest.Original.ContentHash != sourceHash {
		_ = os.RemoveAll(filepath.Join(mediaRoot, "assets", attachmentID))
		return fmt.Errorf("verify migrated chat attachment hash for %s: got %s want %s", attachmentID, manifest.Original.ContentHash, sourceHash)
	}

	if _, err := s.db.ExecContext(ctx, `
		UPDATE chat_attachments
		SET stored_path = ?
		WHERE id = ?
	`, mediaRootRelativePath(mediaRoot, origPath), attachmentID); err != nil {
		_ = os.RemoveAll(filepath.Join(mediaRoot, "assets", attachmentID))
		return fmt.Errorf("update migrated chat attachment %s: %w", attachmentID, err)
	}
	return nil
}

func (s *Service) resumeLegacyInventoryPhotoMigration(ctx context.Context, mediaRoot, photoID, itemID string) error {
	manifest, err := readPromotedCanonicalManifest(mediaRoot, photoID)
	if err != nil {
		return fmt.Errorf("resume legacy inventory photo %s: %w", photoID, err)
	}
	previewPath, err := manifestVariantPath(manifest, "preview")
	if err != nil {
		return fmt.Errorf("resume legacy inventory photo %s: %w", photoID, err)
	}
	thumbnailPath, err := manifestVariantPath(manifest, "thumbnail")
	if err != nil {
		return fmt.Errorf("resume legacy inventory photo %s: %w", photoID, err)
	}
	assetDir := filepath.Join(mediaRoot, "assets", photoID)
	if _, err := s.db.ExecContext(ctx, `
		UPDATE item_photos
		SET original_path = ?, preview_path = ?, thumbnail_path = ?
		WHERE id = ? AND item_id = ?
	`, mediaRootRelativePath(mediaRoot, filepath.Join(assetDir, filepath.FromSlash(manifest.Original.RelativePath))), mediaRootRelativePath(mediaRoot, filepath.Join(assetDir, filepath.FromSlash(previewPath))), mediaRootRelativePath(mediaRoot, filepath.Join(assetDir, filepath.FromSlash(thumbnailPath))), photoID, itemID); err != nil {
		return fmt.Errorf("update resumed inventory photo %s: %w", photoID, err)
	}
	return nil
}

func (s *Service) resumeLegacyChatAttachmentMigration(ctx context.Context, mediaRoot, attachmentID string) error {
	manifest, err := readPromotedCanonicalManifest(mediaRoot, attachmentID)
	if err != nil {
		return fmt.Errorf("resume legacy chat attachment %s: %w", attachmentID, err)
	}
	assetDir := filepath.Join(mediaRoot, "assets", attachmentID)
	if _, err := s.db.ExecContext(ctx, `
		UPDATE chat_attachments
		SET stored_path = ?
		WHERE id = ?
	`, mediaRootRelativePath(mediaRoot, filepath.Join(assetDir, filepath.FromSlash(manifest.Original.RelativePath))), attachmentID); err != nil {
		return fmt.Errorf("update resumed chat attachment %s: %w", attachmentID, err)
	}
	return nil
}

func promotedCanonicalAssetExists(mediaRoot, assetID string) (bool, error) {
	_, err := readPromotedCanonicalManifest(mediaRoot, assetID)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func readPromotedCanonicalManifest(mediaRoot, assetID string) (AssetManifest, error) {
	assetDir := filepath.Join(mediaRoot, "assets", assetID)
	manifest, err := readAssetManifestFile(filepath.Join(assetDir, "manifest.json"))
	if err != nil {
		return AssetManifest{}, err
	}
	if manifest.AssetID != assetID {
		return AssetManifest{}, fmt.Errorf("canonical asset manifest id mismatch: got %q want %q", manifest.AssetID, assetID)
	}
	if strings.TrimSpace(manifest.Original.RelativePath) == "" {
		return AssetManifest{}, fmt.Errorf("canonical asset manifest missing original path")
	}
	originalPath := filepath.Join(assetDir, filepath.FromSlash(manifest.Original.RelativePath))
	if _, err := os.Stat(originalPath); err != nil {
		return AssetManifest{}, err
	}
	return manifest, nil
}

func manifestVariantPath(manifest AssetManifest, name string) (string, error) {
	for _, variant := range manifest.Renditions {
		if variant.Name == name && strings.TrimSpace(variant.RelativePath) != "" {
			return variant.RelativePath, nil
		}
	}
	return "", fmt.Errorf("canonical asset manifest missing %s rendition", name)
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return "", err
	}
	return "sha256:" + hex.EncodeToString(hasher.Sum(nil)), nil
}

func readAssetManifestFile(path string) (AssetManifest, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return AssetManifest{}, err
	}
	var manifest AssetManifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return AssetManifest{}, err
	}
	return manifest, nil
}

func legacyMigrationPathClass(mediaRoot, path string) string {
	if pathWithinDir(mediaRoot, path) {
		return "legacy_media"
	}
	return "legacy_external"
}

func isCanonicalMediaPath(mediaRoot, path string) bool {
	path = strings.TrimSpace(path)
	if path == "" {
		return false
	}
	assetRoot := filepath.Join(mediaRoot, "assets")
	if strings.HasPrefix(filepath.ToSlash(path), "assets/") {
		return true
	}
	return pathWithinDir(assetRoot, path)
}

func markReferencedPath(referenced map[string]bool, path string) {
	key := canonicalPathKey(path)
	if key != "" {
		referenced[key] = true
	}
}

func canonicalPathKey(path string) string {
	path = strings.TrimSpace(path)
	if path == "" || strings.Contains(path, "://") {
		return ""
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return filepath.Clean(path)
	}
	return filepath.Clean(abs)
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

func mediaRootRelativePath(mediaRoot, path string) string {
	path = strings.TrimSpace(path)
	if path == "" || strings.Contains(path, "://") {
		return path
	}
	if filepath.IsAbs(path) {
		if rel, err := filepath.Rel(mediaRoot, path); err == nil && rel != "." && !strings.HasPrefix(rel, "..") {
			return filepath.ToSlash(rel)
		}
	}
	return filepath.ToSlash(path)
}

func resolveMediaRootPath(mediaRoot, storedPath string) string {
	storedPath = strings.TrimSpace(storedPath)
	if storedPath == "" || strings.Contains(storedPath, "://") || filepath.IsAbs(storedPath) {
		return storedPath
	}
	return filepath.Join(mediaRoot, filepath.FromSlash(storedPath))
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
