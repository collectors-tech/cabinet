package backup

import (
	"archive/zip"
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
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
	backupRun, err := svc.CreateBackup(context.Background())
	if err != nil {
		t.Fatalf("CreateBackup error: %v", err)
	}
	if backupRun.Path == "" || backupRun.FileName == "" || backupRun.SizeBytes == 0 || backupRun.IntegrityCheck != "ok" {
		t.Fatalf("expected backup run metadata, got %+v", backupRun)
	}
	if !strings.HasPrefix(backupRun.FileName, "cabinet-backup-") || !strings.HasSuffix(backupRun.FileName, ".zip") {
		t.Fatalf("expected timestamped zip filename, got %q", backupRun.FileName)
	}
	if backupRun.ArchiveFormat != "zip" || backupRun.DownloadURL == "" {
		t.Fatalf("expected zip archive metadata, got %+v", backupRun)
	}
	assertBackupArchiveContains(t, backupRun.Path, "manifest.json")
	assertBackupArchiveContains(t, backupRun.Path, "database/cabinet.db")
	list, err := svc.ListBackups()
	if err != nil {
		t.Fatalf("ListBackups error: %v", err)
	}
	if len(list) == 0 {
		t.Fatal("expected at least one backup")
	}
	if list[0].Path == "" || list[0].FileName == "" || list[0].SizeBytes == 0 {
		t.Fatalf("expected backup list metadata, got %+v", list[0])
	}
	if list[0].ArchiveFormat != "zip" || list[0].DownloadURL == "" {
		t.Fatalf("expected zip list metadata, got %+v", list[0])
	}

	conn2, _ := sql.Open("sqlite", "file:"+dbPath)
	if _, err := conn2.Exec(`DELETE FROM t;`); err != nil {
		t.Fatalf("mutate sqlite: %v", err)
	}
	_ = conn2.Close()

	restore, err := svc.RestoreBackup(backupRun.Path)
	if err != nil {
		t.Fatalf("RestoreBackup error: %v", err)
	}
	if restore.RestoredPath != backupRun.Path || restore.IntegrityCheck != "ok" || restore.RestoredAt == "" {
		t.Fatalf("expected restore metadata, got %+v", restore)
	}
	if !restore.PreRestoreBackupTaken || restore.PreRestoreBackup.Path == "" || restore.PreRestoreBackup.FileName == backupRun.FileName {
		t.Fatalf("expected distinct pre-restore backup metadata, got %+v", restore)
	}
	assertBackupArchiveContains(t, restore.PreRestoreBackup.Path, "database/cabinet.db")
	conn3, _ := sql.Open("sqlite", "file:"+dbPath)
	defer conn3.Close()
	var count int
	if err := conn3.QueryRow(`SELECT COUNT(1) FROM t`).Scan(&count); err != nil {
		t.Fatalf("count after restore: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 row after restore, got %d", count)
	}
	preRestoreDBPath, err := extractArchiveDatabase(restore.PreRestoreBackup.Path)
	if err != nil {
		t.Fatalf("extract pre-restore backup: %v", err)
	}
	defer os.Remove(preRestoreDBPath)
	preRestoreConn, err := sql.Open("sqlite", "file:"+preRestoreDBPath)
	if err != nil {
		t.Fatalf("open pre-restore backup: %v", err)
	}
	defer preRestoreConn.Close()
	var preRestoreCount int
	if err := preRestoreConn.QueryRow(`SELECT COUNT(1) FROM t`).Scan(&preRestoreCount); err != nil {
		t.Fatalf("count pre-restore backup: %v", err)
	}
	if preRestoreCount != 0 {
		t.Fatalf("expected pre-restore backup to preserve mutated state, got %d rows", preRestoreCount)
	}
}

func TestInspectBackupRequiresLifecycleSchemaAndExactProfile(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	dbPath := filepath.Join(base, "cabinet.db")
	conn, err := sql.Open("sqlite", "file:"+dbPath)
	if err != nil {
		t.Fatalf("open compatibility sqlite: %v", err)
	}
	if _, err := conn.Exec(`CREATE TABLE profiles(id TEXT PRIMARY KEY); INSERT INTO profiles(id) VALUES ('profile-a');`); err != nil {
		t.Fatalf("seed profile-only sqlite: %v", err)
	}
	_ = conn.Close()

	svc := NewService(dbPath, filepath.Join(base, "backups"), 1)
	legacy, err := svc.CreateBackup(context.Background())
	if err != nil {
		t.Fatalf("create legacy-shape backup: %v", err)
	}
	if _, err := svc.InspectBackup(legacy.Path, "profile-a"); err == nil || !strings.Contains(err.Error(), "receipt preservation") {
		t.Fatalf("backup without lifecycle schema must be incompatible, err=%v", err)
	}

	conn, err = sql.Open("sqlite", "file:"+dbPath)
	if err != nil {
		t.Fatalf("reopen compatibility sqlite: %v", err)
	}
	if _, err := conn.Exec(`CREATE TABLE agent_skill_previews(id TEXT PRIMARY KEY); CREATE TABLE agent_skill_strong_confirmations(id TEXT PRIMARY KEY);`); err != nil {
		t.Fatalf("seed lifecycle schema: %v", err)
	}
	_ = conn.Close()
	compatible, err := svc.CreateBackup(context.Background())
	if err != nil {
		t.Fatalf("create compatible backup: %v", err)
	}
	validation, err := svc.InspectBackup(compatible.Path, "profile-a")
	if err != nil {
		t.Fatalf("inspect compatible backup: %v", err)
	}
	if !validation.Compatible || !validation.LifecycleSchemaCompatible || !validation.ProfileIncluded {
		t.Fatalf("expected exact-profile compatible backup, got %+v", validation)
	}
	if _, err := svc.InspectBackup(compatible.Path, "profile-b"); err == nil || !strings.Contains(err.Error(), "active profile") {
		t.Fatalf("backup missing required profile must fail closed, err=%v", err)
	}
}

func TestRestoreCopyRejectsActiveDatabaseAlias(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "cabinet.db")
	original := []byte("active workspace database")
	if err := os.WriteFile(dbPath, original, 0o644); err != nil {
		t.Fatalf("seed active database: %v", err)
	}

	if err := copyFile(dbPath, dbPath); err == nil {
		t.Fatal("expected alias copy to fail")
	}
	got, err := os.ReadFile(dbPath)
	if err != nil {
		t.Fatalf("read active database: %v", err)
	}
	if string(got) != string(original) {
		t.Fatalf("active database was changed after failed alias restore copy: %q", got)
	}
}

func TestRestoreRejectsIncompleteArchiveWithoutChangingActiveDatabase(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	dbPath := filepath.Join(base, "cabinet.db")
	conn, err := sql.Open("sqlite", "file:"+dbPath)
	if err != nil {
		t.Fatalf("open active sqlite: %v", err)
	}
	if _, err := conn.Exec(`CREATE TABLE t(id INTEGER PRIMARY KEY, v TEXT); INSERT INTO t(v) VALUES ('active');`); err != nil {
		t.Fatalf("seed active sqlite: %v", err)
	}
	_ = conn.Close()

	badArchive := filepath.Join(base, "incomplete-backup.zip")
	out, err := os.Create(badArchive)
	if err != nil {
		t.Fatalf("create incomplete archive: %v", err)
	}
	zw := zip.NewWriter(out)
	if err := addBytesToZip(zw, "manifest.json", []byte(`{"format":"cabinet-backup-zip-v1"}`)); err != nil {
		t.Fatalf("write incomplete archive manifest: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close incomplete archive writer: %v", err)
	}
	if err := out.Close(); err != nil {
		t.Fatalf("close incomplete archive file: %v", err)
	}

	backupDir := filepath.Join(base, "backups")
	svc := NewService(dbPath, backupDir, 1)
	if _, err := svc.RestoreBackup(badArchive); err == nil || !strings.Contains(err.Error(), "missing database/cabinet.db") {
		t.Fatalf("expected incomplete archive restore to fail before replacement, got %v", err)
	}

	activeConn, err := sql.Open("sqlite", "file:"+dbPath)
	if err != nil {
		t.Fatalf("open active sqlite after failed restore: %v", err)
	}
	defer activeConn.Close()
	var value string
	if err := activeConn.QueryRow(`SELECT v FROM t WHERE id = 1`).Scan(&value); err != nil {
		t.Fatalf("read active sqlite after failed restore: %v", err)
	}
	if value != "active" {
		t.Fatalf("active database was changed after failed restore: %q", value)
	}
	entries, err := os.ReadDir(backupDir)
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("read backup dir after failed restore: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected no pre-restore backup for rejected archive, got %d entries", len(entries))
	}
}

func assertBackupArchiveContains(t *testing.T, path, name string) {
	t.Helper()
	zr, err := zip.OpenReader(path)
	if err != nil {
		t.Fatalf("open backup zip: %v", err)
	}
	defer zr.Close()
	for _, f := range zr.File {
		if f.Name == name {
			return
		}
	}
	t.Fatalf("backup zip missing %s", name)
}

func TestStartDoesNotPanic(t *testing.T) {
	t.Parallel()
	svc := NewService(filepath.Join(t.TempDir(), "db.sqlite"), filepath.Join(t.TempDir(), "backups"), 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc.Start(ctx)
	time.Sleep(20 * time.Millisecond)
}
