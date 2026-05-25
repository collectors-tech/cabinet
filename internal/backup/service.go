package backup

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type Service struct {
	dbPath     string
	backupDir  string
	interval   time.Duration
	mu         sync.Mutex
	lastBackup string
}

type BackupInfo struct {
	Path      string `json:"path"`
	FileName  string `json:"file_name"`
	SizeBytes int64  `json:"size_bytes"`
	CreatedAt string `json:"created_at"`
}

type BackupRunResult struct {
	BackupInfo
	IntegrityCheck string `json:"integrity_check"`
}

type RestoreResult struct {
	RestoredPath   string `json:"restored_path"`
	RestoredAt     string `json:"restored_at"`
	IntegrityCheck string `json:"integrity_check"`
}

func NewService(dbPath, backupDir string, intervalMinutes int) *Service {
	if intervalMinutes <= 0 {
		intervalMinutes = 60
	}
	return &Service{
		dbPath:    dbPath,
		backupDir: backupDir,
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
	name := "cabinet-backup-" + time.Now().UTC().Format("20060102-150405") + ".db"
	target := filepath.Join(s.backupDir, name)
	if err := copyFile(s.dbPath, target); err != nil {
		return BackupRunResult{}, fmt.Errorf("copy db backup: %w", err)
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
		info, err := describeBackup(filepath.Join(s.backupDir, e.Name()))
		if err != nil {
			return nil, err
		}
		out = append(out, info)
	}
	return out, nil
}

func (s *Service) RestoreBackup(backupPath string) (RestoreResult, error) {
	ok, err := validateSQLiteFile(backupPath)
	if err != nil {
		return RestoreResult{}, err
	}
	if ok != "ok" {
		return RestoreResult{}, fmt.Errorf("backup integrity check failed: %s", ok)
	}
	if err := copyFile(backupPath, s.dbPath); err != nil {
		return RestoreResult{}, err
	}
	return RestoreResult{
		RestoredPath:   backupPath,
		RestoredAt:     time.Now().UTC().Format(time.RFC3339),
		IntegrityCheck: ok,
	}, nil
}

func describeBackup(path string) (BackupInfo, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return BackupInfo{}, fmt.Errorf("stat backup: %w", err)
	}
	return BackupInfo{
		Path:      path,
		FileName:  filepath.Base(path),
		SizeBytes: stat.Size(),
		CreatedAt: stat.ModTime().UTC().Format(time.RFC3339),
	}, nil
}

func validateSQLiteFile(path string) (string, error) {
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

func copyFile(src, dst string) error {
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
