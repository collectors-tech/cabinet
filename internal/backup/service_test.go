package backup

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestCreateAndRestoreBackup(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	dbPath := filepath.Join(base, "cabinet.db")
	conn, err := sql.Open("sqlite", "file:"+dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if _, err := conn.Exec(`CREATE TABLE t(id INTEGER PRIMARY KEY, v TEXT); INSERT INTO t(v) VALUES ('a');`); err != nil {
		t.Fatalf("seed sqlite: %v", err)
	}
	_ = conn.Close()

	svc := NewService(dbPath, filepath.Join(base, "backups"), 1)
	backupPath, err := svc.CreateBackup(context.Background())
	if err != nil {
		t.Fatalf("CreateBackup error: %v", err)
	}
	list, err := svc.ListBackups()
	if err != nil {
		t.Fatalf("ListBackups error: %v", err)
	}
	if len(list) == 0 {
		t.Fatal("expected at least one backup")
	}

	conn2, _ := sql.Open("sqlite", "file:"+dbPath)
	if _, err := conn2.Exec(`DELETE FROM t;`); err != nil {
		t.Fatalf("mutate sqlite: %v", err)
	}
	_ = conn2.Close()

	if err := svc.RestoreBackup(backupPath); err != nil {
		t.Fatalf("RestoreBackup error: %v", err)
	}
	conn3, _ := sql.Open("sqlite", "file:"+dbPath)
	defer conn3.Close()
	var count int
	if err := conn3.QueryRow(`SELECT COUNT(1) FROM t`).Scan(&count); err != nil {
		t.Fatalf("count after restore: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 row after restore, got %d", count)
	}
}

func TestStartDoesNotPanic(t *testing.T) {
	t.Parallel()
	svc := NewService(filepath.Join(t.TempDir(), "db.sqlite"), filepath.Join(t.TempDir(), "backups"), 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc.Start(ctx)
	time.Sleep(20 * time.Millisecond)
}
