package backup

import (
	"archive/zip"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type Service struct {
	dbPath     string
	backupDir  string
	dataDir    string
	interval   time.Duration
	mu         sync.Mutex
	lastBackup string
}

type BackupInfo struct {
	Path          string `json:"path"`
	FileName      string `json:"file_name"`
	SizeBytes     int64  `json:"size_bytes"`
	CreatedAt     string `json:"created_at"`
	ArchiveFormat string `json:"archive_format"`
	DownloadURL   string `json:"download_url"`
}

type BackupRunResult struct {
	BackupInfo
	IntegrityCheck string `json:"integrity_check"`
}

type RestoreResult struct {
	RestoredPath          string     `json:"restored_path"`
	RestoredAt            string     `json:"restored_at"`
	IntegrityCheck        string     `json:"integrity_check"`
	PreRestoreBackup      BackupInfo `json:"pre_restore_backup"`
	PreRestoreBackupTaken bool       `json:"pre_restore_backup_taken"`
}

func NewService(dbPath, backupDir string, intervalMinutes int) *Service {
	return NewServiceWithDataDir(dbPath, backupDir, intervalMinutes, "")
}

func NewServiceWithDataDir(dbPath, backupDir string, intervalMinutes int, dataDir string) *Service {
	if intervalMinutes <= 0 {
		intervalMinutes = 60
	}
	return &Service{
		dbPath:    dbPath,
		backupDir: backupDir,
		dataDir:   strings.TrimSpace(dataDir),
		interval:  time.Duration(intervalMinutes) * time.Minute,
	}
}

func (s *Service) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_, _ = s.CreateBackup(context.Background())
			}
		}
	}()
}

func (s *Service) CreateBackup(ctx context.Context) (BackupRunResult, error) {
	_ = ctx
	if err := os.MkdirAll(s.backupDir, 0o755); err != nil {
		return BackupRunResult{}, fmt.Errorf("create backup dir: %w", err)
	}
	createdAt := time.Now().UTC()
	name := "cabinet-backup-" + createdAt.Format("2006-01-02-150405.000000000") + ".zip"
	target := filepath.Join(s.backupDir, name)
	if err := createArchive(s.dbPath, target, createdAt, s.dataDir); err != nil {
		return BackupRunResult{}, fmt.Errorf("create backup archive: %w", err)
	}
	ok, err := validateSQLiteFile(target)
	if err != nil {
		return BackupRunResult{}, err
	}
	if ok != "ok" {
		return BackupRunResult{}, fmt.Errorf("backup integrity check failed: %s", ok)
	}
	s.mu.Lock()
	s.lastBackup = target
	s.mu.Unlock()
	info, err := describeBackup(target)
	if err != nil {
		return BackupRunResult{}, err
	}
	return BackupRunResult{BackupInfo: info, IntegrityCheck: ok}, nil
}

func (s *Service) ListBackups() ([]BackupInfo, error) {
	entries, err := os.ReadDir(s.backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []BackupInfo{}, nil
		}
		return nil, fmt.Errorf("list backups: %w", err)
	}
	out := []BackupInfo{}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if ext != ".zip" && ext != ".db" {
			continue
		}
		info, err := describeBackup(filepath.Join(s.backupDir, e.Name()))
		if err != nil {
			return nil, err
		}
		out = append(out, info)
	}
	return out, nil
}

func (s *Service) ResolveBackupPath(nameOrPath string) (string, error) {
	name := strings.TrimSpace(nameOrPath)
	if name == "" {
		return "", fmt.Errorf("backup file name required")
	}
	base := filepath.Base(name)
	if base != name && filepath.Clean(name) != filepath.Clean(filepath.Join(s.backupDir, base)) {
		return "", fmt.Errorf("backup path must resolve inside backup dir")
	}
	target := filepath.Join(s.backupDir, base)
	if _, err := os.Stat(target); err != nil {
		return "", fmt.Errorf("stat backup: %w", err)
	}
	return target, nil
}

func (s *Service) RestoreBackup(backupPath string) (RestoreResult, error) {
	ok, err := validateSQLiteFile(backupPath)
	if err != nil {
		return RestoreResult{}, err
	}
	if ok != "ok" {
		return RestoreResult{}, fmt.Errorf("backup integrity check failed: %s", ok)
	}
	sourcePath := backupPath
	cleanup := func() {}
	if strings.EqualFold(filepath.Ext(backupPath), ".zip") {
		extracted, err := extractArchiveDatabase(backupPath)
		if err != nil {
			return RestoreResult{}, err
		}
		sourcePath = extracted
		cleanup = func() { _ = os.Remove(extracted) }
	}
	defer cleanup()
	preRestoreBackup, err := s.CreateBackup(context.Background())
	if err != nil {
		return RestoreResult{}, fmt.Errorf("create pre-restore backup: %w", err)
	}
	if err := copyFile(sourcePath, s.dbPath); err != nil {
		return RestoreResult{}, err
	}
	if strings.EqualFold(filepath.Ext(backupPath), ".zip") && strings.TrimSpace(s.dataDir) != "" {
		if err := restoreArchiveMedia(backupPath, s.dataDir); err != nil {
			return RestoreResult{}, err
		}
	}
	return RestoreResult{
		RestoredPath:          backupPath,
		RestoredAt:            time.Now().UTC().Format(time.RFC3339),
		IntegrityCheck:        ok,
		PreRestoreBackup:      preRestoreBackup.BackupInfo,
		PreRestoreBackupTaken: true,
	}, nil
}

func describeBackup(path string) (BackupInfo, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return BackupInfo{}, fmt.Errorf("stat backup: %w", err)
	}
	archiveFormat := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
	if archiveFormat == "" {
		archiveFormat = "unknown"
	}
	fileName := filepath.Base(path)
	return BackupInfo{
		Path:          path,
		FileName:      fileName,
		SizeBytes:     stat.Size(),
		CreatedAt:     stat.ModTime().UTC().Format(time.RFC3339),
		ArchiveFormat: archiveFormat,
		DownloadURL:   "/api/backup/download?file_name=" + fileName,
	}, nil
}

func validateSQLiteFile(path string) (string, error) {
	if strings.EqualFold(filepath.Ext(path), ".zip") {
		extracted, err := extractArchiveDatabase(path)
		if err != nil {
			return "", err
		}
		defer os.Remove(extracted)
		return validateSQLiteFile(extracted)
	}
	conn, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		return "", fmt.Errorf("open backup sqlite: %w", err)
	}
	defer conn.Close()
	var result string
	if err := conn.QueryRow(`PRAGMA integrity_check`).Scan(&result); err != nil {
		return "", fmt.Errorf("run integrity_check: %w", err)
	}
	return result, nil
}

func createArchive(dbPath, target string, createdAt time.Time, dataDir string) error {
	tmp := target + ".tmp"
	if err := os.RemoveAll(tmp); err != nil {
		return err
	}
	out, err := os.Create(tmp)
	if err != nil {
		return err
	}
	zw := zip.NewWriter(out)
	manifest := map[string]string{
		"format":     "cabinet-backup-zip-v1",
		"created_at": createdAt.Format(time.RFC3339),
		"database":   "database/cabinet.db",
	}
	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		_ = zw.Close()
		_ = out.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := addBytesToZip(zw, "manifest.json", manifestBytes); err != nil {
		_ = zw.Close()
		_ = out.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := addFileToZip(zw, dbPath, "database/cabinet.db"); err != nil {
		_ = zw.Close()
		_ = out.Close()
		_ = os.Remove(tmp)
		return err
	}
	if strings.TrimSpace(dataDir) != "" {
		if err := addMediaDirsToZip(zw, dataDir); err != nil {
			_ = zw.Close()
			_ = out.Close()
			_ = os.Remove(tmp)
			return err
		}
	}
	if err := zw.Close(); err != nil {
		_ = out.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := out.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, target)
}

func addBytesToZip(zw *zip.Writer, name string, data []byte) error {
	w, err := zw.Create(name)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func addFileToZip(zw *zip.Writer, src, name string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	w, err := zw.Create(name)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, in)
	return err
}

func addMediaDirsToZip(zw *zip.Writer, dataDir string) error {
	for _, root := range mediaArchiveRoots(dataDir) {
		if _, err := os.Stat(root); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
		if err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				return nil
			}
			rel, err := filepath.Rel(dataDir, path)
			if err != nil {
				return err
			}
			return addFileToZip(zw, path, filepath.ToSlash(rel))
		}); err != nil {
			return err
		}
	}
	return nil
}

func mediaArchiveRoots(dataDir string) []string {
	roots := []string{filepath.Join(dataDir, "media")}
	profilesDir := filepath.Join(dataDir, "profiles")
	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		return roots
	}
	for _, entry := range entries {
		if entry.IsDir() {
			roots = append(roots, filepath.Join(profilesDir, entry.Name(), "media"))
		}
	}
	return roots
}

func extractArchiveDatabase(path string) (string, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return "", fmt.Errorf("open backup archive: %w", err)
	}
	defer zr.Close()
	for _, f := range zr.File {
		if f.Name != "database/cabinet.db" {
			continue
		}
		in, err := f.Open()
		if err != nil {
			return "", fmt.Errorf("open archive database: %w", err)
		}
		tmp, err := os.CreateTemp("", "cabinet-backup-*.db")
		if err != nil {
			_ = in.Close()
			return "", err
		}
		if _, err := io.Copy(tmp, in); err != nil {
			_ = in.Close()
			_ = tmp.Close()
			_ = os.Remove(tmp.Name())
			return "", err
		}
		_ = in.Close()
		if err := tmp.Close(); err != nil {
			_ = os.Remove(tmp.Name())
			return "", err
		}
		return tmp.Name(), nil
	}
	return "", fmt.Errorf("backup archive missing database/cabinet.db")
}

func restoreArchiveMedia(path, dataDir string) error {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return fmt.Errorf("open backup archive: %w", err)
	}
	defer zr.Close()
	for _, file := range zr.File {
		if !isMediaArchiveEntry(file.Name) {
			continue
		}
		cleanName := filepath.Clean(filepath.FromSlash(file.Name))
		target := filepath.Join(dataDir, cleanName)
		if !pathWithin(dataDir, target) {
			return fmt.Errorf("backup media path escapes data dir: %s", file.Name)
		}
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		in, err := file.Open()
		if err != nil {
			return fmt.Errorf("open archive media file: %w", err)
		}
		out, err := os.Create(target)
		if err != nil {
			_ = in.Close()
			return err
		}
		_, copyErr := io.Copy(out, in)
		closeInErr := in.Close()
		closeOutErr := out.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeInErr != nil {
			return closeInErr
		}
		if closeOutErr != nil {
			return closeOutErr
		}
	}
	return nil
}

func isMediaArchiveEntry(name string) bool {
	name = filepath.ToSlash(strings.TrimSpace(name))
	return strings.HasPrefix(name, "media/") || (strings.HasPrefix(name, "profiles/") && strings.Contains(name, "/media/"))
}

func pathWithin(root, target string) bool {
	rel, err := filepath.Rel(root, target)
	return err == nil && rel != "." && !strings.HasPrefix(rel, "..")
}

func copyFile(src, dst string) error {
	if samePath, err := sameFilePath(src, dst); err != nil {
		return err
	} else if samePath {
		return fmt.Errorf("restore source and active database destination must differ")
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	dstDir := filepath.Dir(dst)
	tmp, err := os.CreateTemp(dstDir, ".cabinet-restore-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	cleanupTmp := true
	defer func() {
		if cleanupTmp {
			_ = os.Remove(tmpPath)
		}
	}()

	if _, err := io.Copy(tmp, in); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	if _, err := os.Stat(dst); err != nil && !os.IsNotExist(err) {
		return err
	}

	if err := overwriteFile(tmpPath, dst); err != nil {
		return err
	}
	return nil
}

func overwriteFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

func sameFilePath(a, b string) (bool, error) {
	aInfo, aErr := os.Stat(a)
	bInfo, bErr := os.Stat(b)
	if aErr == nil && bErr == nil {
		return os.SameFile(aInfo, bInfo), nil
	}
	if aErr != nil && !os.IsNotExist(aErr) {
		return false, aErr
	}
	if bErr != nil && !os.IsNotExist(bErr) {
		return false, bErr
	}
	aAbs, err := filepath.Abs(a)
	if err != nil {
		return false, err
	}
	bAbs, err := filepath.Abs(b)
	if err != nil {
		return false, err
	}
	return strings.EqualFold(filepath.Clean(aAbs), filepath.Clean(bAbs)), nil
}
