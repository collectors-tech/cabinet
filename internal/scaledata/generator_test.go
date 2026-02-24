package scaledata

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/collectors-tech/cabinet/internal/db"
)

func openTestDB(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	path := filepath.Join(root, "scale.db")
	conn, err := db.OpenAndMigrate(context.Background(), path)
	if err != nil {
		t.Fatalf("open and migrate db: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return path
}

func TestDatasetProfilesExposeDeterministicCounts(t *testing.T) {
	tests := []struct {
		name     DatasetProfile
		items    int
		instances int
		photos   int
		barcodes int
		discovery int
		wishlist int
	}{
		{name: DatasetS0Empty, items: 0, instances: 0, photos: 0, barcodes: 0, discovery: 0, wishlist: 0},
		{name: DatasetS1Starter, items: 100, instances: 200, photos: 300, barcodes: 150, discovery: 50, wishlist: 0},
		{name: DatasetS2Growth, items: 5000, instances: 15000, photos: 20000, barcodes: 8000, discovery: 2000, wishlist: 1000},
		{name: DatasetS3Stress, items: 25000, instances: 80000, photos: 150000, barcodes: 40000, discovery: 10000, wishlist: 5000},
	}
	for _, tt := range tests {
		profile, ok := DatasetProfileDefinition(tt.name)
		if !ok {
			t.Fatalf("profile not found: %s", tt.name)
		}
		if profile.Items != tt.items || profile.Instances != tt.instances || profile.Photos != tt.photos || profile.Barcodes != tt.barcodes || profile.DiscoveryCandidates != tt.discovery || profile.WishlistEntries != tt.wishlist {
			t.Fatalf("profile mismatch for %s: %+v", tt.name, profile)
		}
	}
}

func TestGenerateReplaceAppendAndSnapshot(t *testing.T) {
	dbPath := openTestDB(t)
	g := NewGenerator(dbPath)
	ctx := context.Background()

	snapshotPath := filepath.Join(t.TempDir(), "s1-snapshot.json")
	first, err := g.Generate(ctx, GenerateOptions{
		ProfileID:        "p-scale",
		DatasetProfile:   DatasetS1Starter,
		Seed:             42,
		Mode:             ModeReplace,
		IncludePricing:   true,
		IncludeDiscovery: true,
		SnapshotPath:     snapshotPath,
	})
	if err != nil {
		t.Fatalf("replace generate failed: %v", err)
	}
	if first.Items != 100 || first.Instances != 200 || first.Photos != 300 || first.Barcodes != 150 || first.DiscoveryCandidates != 50 {
		t.Fatalf("unexpected first summary: %+v", first)
	}
	raw, err := os.ReadFile(snapshotPath)
	if err != nil {
		t.Fatalf("read snapshot: %v", err)
	}
	var snap map[string]any
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatalf("decode snapshot: %v", err)
	}
	if snap["profile_id"] != "p-scale" {
		t.Fatalf("snapshot profile mismatch: %#v", snap)
	}

	second, err := g.Generate(ctx, GenerateOptions{
		ProfileID:        "p-scale",
		DatasetProfile:   DatasetS1Starter,
		Seed:             42,
		Mode:             ModeAppend,
		IncludePricing:   false,
		IncludeDiscovery: true,
	})
	if err != nil {
		t.Fatalf("append generate failed: %v", err)
	}
	if second.Items != 200 {
		t.Fatalf("expected append growth to 200 items, got %+v", second)
	}
}

func TestGenerateProfileIsolation(t *testing.T) {
	dbPath := openTestDB(t)
	g := NewGenerator(dbPath)
	ctx := context.Background()

	if _, err := g.Generate(ctx, GenerateOptions{
		ProfileID:      "p-a",
		DatasetProfile: DatasetS1Starter,
		Seed:           11,
		Mode:           ModeReplace,
	}); err != nil {
		t.Fatalf("seed profile a: %v", err)
	}
	if _, err := g.Generate(ctx, GenerateOptions{
		ProfileID:      "p-b",
		DatasetProfile: DatasetS0Empty,
		Seed:           12,
		Mode:           ModeReplace,
	}); err != nil {
		t.Fatalf("seed profile b: %v", err)
	}

	countsA, err := g.CountsForProfile(ctx, "p-a")
	if err != nil {
		t.Fatalf("count profile a: %v", err)
	}
	countsB, err := g.CountsForProfile(ctx, "p-b")
	if err != nil {
		t.Fatalf("count profile b: %v", err)
	}
	if countsA.Items != 100 {
		t.Fatalf("expected p-a items 100, got %+v", countsA)
	}
	if countsB.Items != 0 {
		t.Fatalf("expected p-b items 0, got %+v", countsB)
	}
}
