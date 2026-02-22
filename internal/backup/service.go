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

func (s *Service) CreateBackup(ctx context.Context) (string, error) {
	_ = ctx
	if err := os.MkdirAll(s.backupDir, 0o755); err != nil {
		return "", fmt.Errorf("create backup dir: %w", err)
	}
	name := "cabinet-backup-" + time.Now().UTC().Format("20060102-150405") + ".db"
	target := filepath.Join(s.backupDir, name)
	if err := copyFile(s.dbPath, target); err != nil {
		return "", fmt.Errorf("copy db backup: %w", err)
	}
	ok, err := validateSQLiteFile(target)
	if err != nil {
		return "", err
	}
	if ok != "ok" {
		return "", fmt.Errorf("backup integrity check failed: %s", ok)
	}
	s.mu.Lock()
	s.lastBackup = target
	s.mu.Unlock()
	return target, nil
}

func (s *Service) ListBackups() ([]string, error) {
	entries, err := os.ReadDir(s.backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("list backups: %w", err)
	}
	out := []string{}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		out = append(out, filepath.Join(s.backupDir, e.Name()))
	}
	return out, nil
}

func (s *Service) RestoreBackup(backupPath string) error {
	ok, err := validateSQLiteFile(backupPath)
	if err != nil {
		return err
	}
	if ok != "ok" {
		return fmt.Errorf("backup integrity check failed: %s", ok)
	}
	return copyFile(backupPath, s.dbPath)
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
