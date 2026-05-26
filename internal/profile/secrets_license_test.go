package profile

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/collectors-tech/cabinet/internal/db"
)

func TestPerProfileSecretAndLicenseIsolation(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "cabinet.db")
	conn, err := db.OpenAndMigrate(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	repo := NewRepository(conn)
	_ = os.Setenv("CABINET_ALLOW_INSECURE_SECRET_FALLBACK", "1")
	p1, err := repo.Create(context.Background(), "A")
	if err != nil {
		t.Fatalf("Create A error = %v", err)
	}
	p2, err := repo.Create(context.Background(), "B")
	if err != nil {
		t.Fatalf("Create B error = %v", err)
	}

	if err := repo.PutSecret(context.Background(), p1.ID, "openai_api_key", "sk-a"); err != nil {
		t.Fatalf("PutSecret p1 error = %v", err)
	}
	if err := repo.PutSecret(context.Background(), p2.ID, "openai_api_key", "sk-b"); err != nil {
		t.Fatalf("PutSecret p2 error = %v", err)
	}

	s1, err := repo.GetSecret(context.Background(), p1.ID, "openai_api_key")
	if err != nil {
		t.Fatalf("GetSecret p1 error = %v", err)
	}
	s2, err := repo.GetSecret(context.Background(), p2.ID, "openai_api_key")
	if err != nil {
		t.Fatalf("GetSecret p2 error = %v", err)
	}
	if s1 == s2 {
		t.Fatal("expected secrets to remain isolated by profile")
	}
	if err := repo.DeleteSecret(context.Background(), p1.ID, "openai_api_key"); err != nil {
		t.Fatalf("DeleteSecret p1 error = %v", err)
	}
	if _, err := repo.GetSecret(context.Background(), p1.ID, "openai_api_key"); err == nil {
		t.Fatal("expected deleted p1 secret lookup to fail")
	}
	s2AfterDelete, err := repo.GetSecret(context.Background(), p2.ID, "openai_api_key")
	if err != nil {
		t.Fatalf("GetSecret p2 after p1 delete error = %v", err)
	}
	if s2AfterDelete != s2 {
		t.Fatal("deleting p1 secret must not affect p2 secret")
	}

	if err := repo.PutLicense(context.Background(), p1.ID, `{"tier":"pro"}`); err != nil {
		t.Fatalf("PutLicense p1 error = %v", err)
	}
	if err := repo.PutLicense(context.Background(), p2.ID, `{"tier":"free"}`); err != nil {
		t.Fatalf("PutLicense p2 error = %v", err)
	}
	l1, err := repo.GetLicense(context.Background(), p1.ID)
	if err != nil {
		t.Fatalf("GetLicense p1 error = %v", err)
	}
	l2, err := repo.GetLicense(context.Background(), p2.ID)
	if err != nil {
		t.Fatalf("GetLicense p2 error = %v", err)
	}
	if l1 == l2 {
		t.Fatal("expected license payloads to remain isolated by profile")
	}
}
