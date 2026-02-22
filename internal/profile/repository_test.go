package profile

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/collectors-tech/cabinet/internal/db"
)

func TestRepositoryCreateAndList(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "cabinet.db")
	conn, err := db.OpenAndMigrate(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	repo := NewRepository(conn)
	if _, err := repo.Create(context.Background(), " "); err == nil {
		t.Fatal("expected error for blank name")
	}

	created, err := repo.Create(context.Background(), "Default")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.ID == "" || created.CreatedAt == "" {
		t.Fatalf("unexpected created profile: %+v", created)
	}

	all, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(all))
	}
	if all[0].Name != "Default" {
		t.Fatalf("expected profile name Default, got %q", all[0].Name)
	}

	if err := repo.SetActiveProfile(context.Background(), created.ID); err != nil {
		t.Fatalf("SetActiveProfile() error = %v", err)
	}
	active, err := repo.GetActiveProfile(context.Background())
	if err != nil {
		t.Fatalf("GetActiveProfile() error = %v", err)
	}
	if active.ID != created.ID {
		t.Fatalf("expected active profile %q, got %q", created.ID, active.ID)
	}

	if err := repo.PutSettings(context.Background(), created.ID, map[string]string{
		"theme":            "dark",
		"scanner_schedule": "0 */6 * * *",
	}); err != nil {
		t.Fatalf("PutSettings() error = %v", err)
	}
	settings, err := repo.GetSettings(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("GetSettings() error = %v", err)
	}
	if settings["theme"] != "dark" {
		t.Fatalf("unexpected theme setting: %q", settings["theme"])
	}
}
