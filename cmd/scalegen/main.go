package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/collectors-tech/cabinet/internal/config"
	"github.com/collectors-tech/cabinet/internal/scaledata"
)

func main() {
	cfg := config.Load()
	dbPath := flag.String("db", cfg.DBPath, "SQLite database path")
	profileID := flag.String("profile-id", "default", "Profile ID for generation")
	dataset := flag.String("dataset-profile", "S1", "Dataset profile: S0|S1|S2|S3")
	seed := flag.Int64("seed", 1, "Deterministic seed")
	mode := flag.String("mode", string(scaledata.ModeReplace), "Generation mode: replace|append")
	dateSpanMonths := flag.Int("date-span-months", 0, "Override pricing span in months")
	includePricing := flag.Bool("include-pricing", true, "Include pricing snapshots")
	includeDiscovery := flag.Bool("include-discovery", true, "Include discovery candidates")
	snapshotPath := flag.String("snapshot", "", "Optional JSON snapshot export path")
	flag.Parse()

	gen := scaledata.NewGenerator(*dbPath)
	summary, err := gen.Generate(context.Background(), scaledata.GenerateOptions{
		ProfileID:        *profileID,
		DatasetProfile:   scaledata.DatasetProfile(*dataset),
		Seed:             *seed,
		Mode:             scaledata.GenerationMode(*mode),
		DateSpanMonths:   *dateSpanMonths,
		IncludePricing:   *includePricing,
		IncludeDiscovery: *includeDiscovery,
		SnapshotPath:     *snapshotPath,
	})
	if err != nil {
		log.Fatalf("scale generation failed: %v", err)
	}

	fmt.Fprintf(os.Stdout, "Generated %s for profile %s\n", summary.DatasetProfile, summary.ProfileID)
	fmt.Fprintf(os.Stdout, "Items=%d Instances=%d Photos=%d Barcodes=%d Discovery=%d Wishlist=%d PriceSnapshots=%d\n",
		summary.Items, summary.Instances, summary.Photos, summary.Barcodes, summary.DiscoveryCandidates, summary.WishlistEntries, summary.PriceSnapshots,
	)
	if *snapshotPath != "" {
		fmt.Fprintf(os.Stdout, "Snapshot=%s\n", *snapshotPath)
	}
}
