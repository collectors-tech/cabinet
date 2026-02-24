package scaledata

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/collectors-tech/cabinet/internal/db"
)

type DatasetProfile string

const (
	DatasetS0Empty   DatasetProfile = "S0"
	DatasetS1Starter DatasetProfile = "S1"
	DatasetS2Growth  DatasetProfile = "S2"
	DatasetS3Stress  DatasetProfile = "S3"
)

type GenerationMode string

const (
	ModeReplace GenerationMode = "replace"
	ModeAppend  GenerationMode = "append"
)

type ProfileDefinition struct {
	Name                DatasetProfile
	Items               int
	Instances           int
	Photos              int
	Barcodes            int
	DiscoveryCandidates int
	WishlistEntries     int
	DefaultMonths       int
}

type GenerateOptions struct {
	ProfileID        string
	DatasetProfile   DatasetProfile
	Seed             int64
	Mode             GenerationMode
	DateSpanMonths   int
	IncludePricing   bool
	IncludeDiscovery bool
	SnapshotPath     string
}

type Summary struct {
	ProfileID          string `json:"profile_id"`
	DatasetProfile     string `json:"dataset_profile"`
	Mode               string `json:"mode"`
	Seed               int64  `json:"seed"`
	Items              int    `json:"items"`
	Instances          int    `json:"instances"`
	Photos             int    `json:"photos"`
	Barcodes           int    `json:"barcodes"`
	DiscoveryCandidates int   `json:"discovery_candidates"`
	WishlistEntries    int    `json:"wishlist_entries"`
	PriceSnapshots     int    `json:"price_snapshots"`
}

type Generator struct {
	dbPath string
}

func NewGenerator(dbPath string) *Generator {
	return &Generator{dbPath: dbPath}
}

func DatasetProfileDefinition(name DatasetProfile) (ProfileDefinition, bool) {
	defs := map[DatasetProfile]ProfileDefinition{
		DatasetS0Empty: {
			Name:                DatasetS0Empty,
			Items:               0,
			Instances:           0,
			Photos:              0,
			Barcodes:            0,
			DiscoveryCandidates: 0,
			WishlistEntries:     0,
			DefaultMonths:       0,
		},
		DatasetS1Starter: {
			Name:                DatasetS1Starter,
			Items:               100,
			Instances:           200,
			Photos:              300,
			Barcodes:            150,
			DiscoveryCandidates: 50,
			WishlistEntries:     0,
			DefaultMonths:       0,
		},
		DatasetS2Growth: {
			Name:                DatasetS2Growth,
			Items:               5000,
			Instances:           15000,
			Photos:              20000,
			Barcodes:            8000,
			DiscoveryCandidates: 2000,
			WishlistEntries:     1000,
			DefaultMonths:       12,
		},
		DatasetS3Stress: {
			Name:                DatasetS3Stress,
			Items:               25000,
			Instances:           80000,
			Photos:              150000,
			Barcodes:            40000,
			DiscoveryCandidates: 10000,
			WishlistEntries:     5000,
			DefaultMonths:       24,
		},
	}
	out, ok := defs[name]
	return out, ok
}

func (g *Generator) Generate(ctx context.Context, opts GenerateOptions) (Summary, error) {
	if strings.TrimSpace(opts.ProfileID) == "" {
		return Summary{}, fmt.Errorf("profile_id_required")
	}
	if opts.Mode == "" {
		opts.Mode = ModeReplace
	}
	if opts.Mode != ModeReplace && opts.Mode != ModeAppend {
		return Summary{}, fmt.Errorf("unsupported_mode")
	}
	def, ok := DatasetProfileDefinition(opts.DatasetProfile)
	if !ok {
		return Summary{}, fmt.Errorf("unsupported_dataset_profile")
	}
	if opts.DateSpanMonths <= 0 {
		opts.DateSpanMonths = def.DefaultMonths
	}
	if opts.Seed == 0 {
		opts.Seed = 1
	}

	conn, err := db.OpenAndMigrate(ctx, g.dbPath)
	if err != nil {
		return Summary{}, err
	}
	defer conn.Close()

	if err := ensureProfile(ctx, conn, opts.ProfileID); err != nil {
		return Summary{}, err
	}
	if opts.Mode == ModeReplace {
		if err := clearProfileData(ctx, conn, opts.ProfileID); err != nil {
			return Summary{}, err
		}
	}

	rng := rand.New(rand.NewSource(stableSeed(opts.ProfileID, opts.Seed)))
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return Summary{}, fmt.Errorf("begin tx: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	offset, err := countByProfile(ctx, tx, "canonical_items", opts.ProfileID)
	if err != nil {
		return Summary{}, err
	}
	instanceOffset, err := countRowsTx(ctx, tx, `SELECT COUNT(1) FROM instances i JOIN canonical_items c ON c.id = i.item_id WHERE c.profile_id = ?`, opts.ProfileID)
	if err != nil {
		return Summary{}, err
	}
	photoOffset, err := countRowsTx(ctx, tx, `SELECT COUNT(1) FROM item_photos p JOIN canonical_items c ON c.id = p.item_id WHERE c.profile_id = ?`, opts.ProfileID)
	if err != nil {
		return Summary{}, err
	}
	barcodeOffset, err := countRowsTx(ctx, tx, `SELECT COUNT(1) FROM item_barcodes b JOIN canonical_items c ON c.id = b.item_id WHERE c.profile_id = ?`, opts.ProfileID)
	if err != nil {
		return Summary{}, err
	}
	discoveryOffset, err := countRowsTx(ctx, tx, `SELECT COUNT(1) FROM scanner_candidates WHERE profile_id = ?`, opts.ProfileID)
	if err != nil {
		return Summary{}, err
	}
	wishlistOffset, err := countRowsTx(ctx, tx, `SELECT COUNT(1) FROM wishlist_entries WHERE profile_id = ?`, opts.ProfileID)
	if err != nil {
		return Summary{}, err
	}

	itemIDs, err := insertItems(ctx, tx, opts.ProfileID, def.Items, offset, rng)
	if err != nil {
		return Summary{}, err
	}
	if err := insertInstances(ctx, tx, itemIDs, def.Instances, instanceOffset, rng); err != nil {
		return Summary{}, err
	}
	if err := insertPhotos(ctx, tx, itemIDs, def.Photos, photoOffset, rng); err != nil {
		return Summary{}, err
	}
	if err := insertBarcodes(ctx, tx, itemIDs, def.Barcodes, barcodeOffset, rng); err != nil {
		return Summary{}, err
	}

	querySetIDs := []string{}
	if opts.IncludeDiscovery {
		querySetIDs, err = insertDiscovery(ctx, tx, opts.ProfileID, itemIDs, def.DiscoveryCandidates, discoveryOffset, rng)
		if err != nil {
			return Summary{}, err
		}
	}
	_ = querySetIDs

	if err := insertWishlist(ctx, tx, opts.ProfileID, itemIDs, def.WishlistEntries, wishlistOffset, rng); err != nil {
		return Summary{}, err
	}

	if opts.IncludePricing && opts.DateSpanMonths > 0 {
		if err := insertPricingSnapshots(ctx, tx, opts.ProfileID, itemIDs, def.WishlistEntries, opts.DateSpanMonths, offset, rng); err != nil {
			return Summary{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return Summary{}, fmt.Errorf("commit tx: %w", err)
	}
	committed = true

	summary, err := g.CountsForProfile(ctx, opts.ProfileID)
	if err != nil {
		return Summary{}, err
	}
	summary.ProfileID = opts.ProfileID
	summary.DatasetProfile = string(opts.DatasetProfile)
	summary.Mode = string(opts.Mode)
	summary.Seed = opts.Seed

	if strings.TrimSpace(opts.SnapshotPath) != "" {
		if err := exportSnapshot(ctx, conn, opts.SnapshotPath, summary); err != nil {
			return Summary{}, err
		}
	}

	return summary, nil
}

func (g *Generator) CountsForProfile(ctx context.Context, profileID string) (Summary, error) {
	conn, err := db.OpenAndMigrate(ctx, g.dbPath)
	if err != nil {
		return Summary{}, err
	}
	defer conn.Close()
	return countsForProfileOnConn(ctx, conn, profileID)
}

func countsForProfileOnConn(ctx context.Context, conn *sql.DB, profileID string) (Summary, error) {
	s := Summary{ProfileID: profileID}
	var err error
	if s.Items, err = countRows(ctx, conn, `SELECT COUNT(1) FROM canonical_items WHERE profile_id = ?`, profileID); err != nil {
		return Summary{}, err
	}
	if s.Instances, err = countRows(ctx, conn, `SELECT COUNT(1) FROM instances i JOIN canonical_items c ON c.id = i.item_id WHERE c.profile_id = ?`, profileID); err != nil {
		return Summary{}, err
	}
	if s.Photos, err = countRows(ctx, conn, `SELECT COUNT(1) FROM item_photos p JOIN canonical_items c ON c.id = p.item_id WHERE c.profile_id = ?`, profileID); err != nil {
		return Summary{}, err
	}
	if s.Barcodes, err = countRows(ctx, conn, `SELECT COUNT(1) FROM item_barcodes b JOIN canonical_items c ON c.id = b.item_id WHERE c.profile_id = ?`, profileID); err != nil {
		return Summary{}, err
	}
	if s.DiscoveryCandidates, err = countRows(ctx, conn, `SELECT COUNT(1) FROM scanner_candidates WHERE profile_id = ?`, profileID); err != nil {
		return Summary{}, err
	}
	if s.WishlistEntries, err = countRows(ctx, conn, `SELECT COUNT(1) FROM wishlist_entries WHERE profile_id = ?`, profileID); err != nil {
		return Summary{}, err
	}
	if s.PriceSnapshots, err = countRows(ctx, conn, `SELECT COUNT(1) FROM price_snapshots ps JOIN canonical_items c ON c.id = ps.item_id WHERE c.profile_id = ?`, profileID); err != nil {
		return Summary{}, err
	}
	return s, nil
}

func countByProfile(ctx context.Context, tx *sql.Tx, table, profileID string) (int, error) {
	var count int
	if err := tx.QueryRowContext(ctx, fmt.Sprintf(`SELECT COUNT(1) FROM %s WHERE profile_id = ?`, table), profileID).Scan(&count); err != nil {
		return 0, fmt.Errorf("count %s: %w", table, err)
	}
	return count, nil
}

func ensureProfile(ctx context.Context, conn *sql.DB, profileID string) error {
	_, err := conn.ExecContext(ctx, `INSERT OR IGNORE INTO profiles(id, name) VALUES (?, ?)`, profileID, "Scale "+profileID)
	if err != nil {
		return fmt.Errorf("ensure profile: %w", err)
	}
	return nil
}

func clearProfileData(ctx context.Context, conn *sql.DB, profileID string) error {
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin clear tx: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	queries := []struct {
		sql  string
		args []any
	}{
		{sql: `DELETE FROM wishlist_entries WHERE profile_id = ?`, args: []any{profileID}},
		{sql: `DELETE FROM tracked_items WHERE profile_id = ?`, args: []any{profileID}},
		{sql: `DELETE FROM scanner_query_sets WHERE profile_id = ?`, args: []any{profileID}},
		{sql: `DELETE FROM scanner_candidates WHERE profile_id = ?`, args: []any{profileID}},
		{sql: `DELETE FROM canonical_items WHERE profile_id = ?`, args: []any{profileID}},
	}
	for _, q := range queries {
		if _, err := tx.ExecContext(ctx, q.sql, q.args...); err != nil {
			return fmt.Errorf("clear profile data: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit clear tx: %w", err)
	}
	committed = true
	return nil
}

func countRows(ctx context.Context, conn *sql.DB, q string, args ...any) (int, error) {
	var out int
	if err := conn.QueryRowContext(ctx, q, args...).Scan(&out); err != nil {
		return 0, fmt.Errorf("count rows: %w", err)
	}
	return out, nil
}

func countRowsTx(ctx context.Context, tx *sql.Tx, q string, args ...any) (int, error) {
	var out int
	if err := tx.QueryRowContext(ctx, q, args...).Scan(&out); err != nil {
		return 0, fmt.Errorf("count rows tx: %w", err)
	}
	return out, nil
}

func insertItems(ctx context.Context, tx *sql.Tx, profileID string, count, offset int, rng *rand.Rand) ([]string, error) {
	if count <= 0 {
		return nil, nil
	}
	brands := buildNames("Brand", 60)
	categories := buildNames("Category", 40)
	scales := []string{"1:64", "1:43", "HO", "1:24", "1:18"}
	series := buildNames("Series", 80)
	tags := buildNames("tag", 300)
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO canonical_items(id, profile_id, brand, category, part_number, title, make, model, year, scale, series, description, tags_json, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`)
	if err != nil {
		return nil, fmt.Errorf("prepare item insert: %w", err)
	}
	defer stmt.Close()

	itemIDs := make([]string, 0, count)
	for i := 0; i < count; i++ {
		idx := offset + i + 1
		id := fmt.Sprintf("scale-item-%s-%d", profileID, idx)
		partNumber := fmt.Sprintf("PN-%s-%06d", strings.ToUpper(profileID), idx)
		tagSlice := []string{tags[idx%len(tags)], tags[(idx*7)%len(tags)], tags[(idx*13)%len(tags)]}
		tagJSON, _ := json.Marshal(tagSlice)
		if _, err := stmt.ExecContext(
			ctx,
			id,
			profileID,
			brands[idx%len(brands)],
			categories[idx%len(categories)],
			partNumber,
			fmt.Sprintf("%s %s %d", brands[idx%len(brands)], categories[idx%len(categories)], idx),
			fmt.Sprintf("Make-%02d", idx%15),
			fmt.Sprintf("Model-%03d", idx%200),
			fmt.Sprintf("%d", 1960+(idx%65)),
			scales[idx%len(scales)],
			series[idx%len(series)],
			fmt.Sprintf("Generated item %d for scalability profile %s", idx, profileID),
			string(tagJSON),
		); err != nil {
			return nil, fmt.Errorf("insert item %d: %w", idx, err)
		}
		itemIDs = append(itemIDs, id)
	}
	return itemIDs, nil
}

func insertInstances(ctx context.Context, tx *sql.Tx, itemIDs []string, count, offset int, rng *rand.Rand) error {
	if count <= 0 || len(itemIDs) == 0 {
		return nil
	}
	statuses := []string{"sealed", "blister", "loose", "custom", "on_track"}
	conditions := []string{"mint", "excellent", "good", "fair"}
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO instances(id, item_id, condition, status, quantity, storage_location, acquisition_price, acquisition_date, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`)
	if err != nil {
		return fmt.Errorf("prepare instance insert: %w", err)
	}
	defer stmt.Close()
	for i := 0; i < count; i++ {
		itemID := itemIDs[i%len(itemIDs)]
		status := statuses[i%len(statuses)]
		idx := offset + i + 1
		if _, err := stmt.ExecContext(
			ctx,
			fmt.Sprintf("scale-instance-%07d", idx),
			itemID,
			conditions[(i*3)%len(conditions)],
			status,
			1+(i%4),
			fmt.Sprintf("Shelf-%d", i%120),
			math.Round((15+rng.Float64()*120)*100)/100,
			time.Now().AddDate(0, 0, -i%720).Format("2006-01-02"),
			fmt.Sprintf("instance note %d", i+1),
		); err != nil {
			return fmt.Errorf("insert instance %d: %w", i+1, err)
		}
	}
	return nil
}

func insertPhotos(ctx context.Context, tx *sql.Tx, itemIDs []string, count, offset int, _ *rand.Rand) error {
	if count <= 0 || len(itemIDs) == 0 {
		return nil
	}
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO item_photos(id, item_id, filename, original_path, preview_path, thumbnail_path, is_primary, display_order, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`)
	if err != nil {
		return fmt.Errorf("prepare photo insert: %w", err)
	}
	defer stmt.Close()
	for i := 0; i < count; i++ {
		itemID := itemIDs[i%len(itemIDs)]
		idx := offset + i + 1
		if _, err := stmt.ExecContext(
			ctx,
			fmt.Sprintf("scale-photo-%08d", idx),
			itemID,
			fmt.Sprintf("photo-%08d.jpg", idx),
			fmt.Sprintf("media/original/photo-%08d.jpg", idx),
			fmt.Sprintf("media/preview/photo-%08d.jpg", idx),
			fmt.Sprintf("media/thumb/photo-%08d.jpg", idx),
			boolToInt(i%10 == 0),
			i%5,
		); err != nil {
			return fmt.Errorf("insert photo %d: %w", i+1, err)
		}
	}
	return nil
}

func insertBarcodes(ctx context.Context, tx *sql.Tx, itemIDs []string, count, offset int, rng *rand.Rand) error {
	if count <= 0 || len(itemIDs) == 0 {
		return nil
	}
	stmt, err := tx.PrepareContext(ctx, `INSERT INTO item_barcodes(id, item_id, barcode, created_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP)`)
	if err != nil {
		return fmt.Errorf("prepare barcode insert: %w", err)
	}
	defer stmt.Close()
	for i := 0; i < count; i++ {
		itemID := itemIDs[i%len(itemIDs)]
		idx := offset + i + 1
		barcode := fmt.Sprintf("%013d", 1000000000000+int64(idx*17)+int64(rng.Intn(9)))
		if _, err := stmt.ExecContext(ctx, fmt.Sprintf("scale-barcode-%08d", idx), itemID, barcode); err != nil {
			return fmt.Errorf("insert barcode %d: %w", i+1, err)
		}
	}
	return nil
}

func insertDiscovery(ctx context.Context, tx *sql.Tx, profileID string, itemIDs []string, candidates, offset int, rng *rand.Rand) ([]string, error) {
	if candidates <= 0 {
		return nil, nil
	}
	querySetID := fmt.Sprintf("scale-queryset-%s", profileID)
	keywordsJSON := `["slot cars","collectibles"]`
	if _, err := tx.ExecContext(ctx, `
		INSERT OR IGNORE INTO scanner_query_sets(
			id, profile_id, name, keywords_json, exclusions_json, max_price, region, condition_filter, schedule_cron, enabled, rate_limit_rps, max_retry_count, created_at, updated_at
		) VALUES (?, ?, ?, ?, '[]', 250, 'AU', '', '', 1, 2, 2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, querySetID, profileID, "Scale Data Query", keywordsJSON); err != nil {
		return nil, fmt.Errorf("insert query set: %w", err)
	}

	candidateStmt, err := tx.PrepareContext(ctx, `
		INSERT INTO scanner_candidates(
			id, profile_id, query_set_id, listing_id, title, price, shipping, url, image, seller, first_seen, last_seen, status, source
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'new', ?)
	`)
	if err != nil {
		return nil, fmt.Errorf("prepare candidate insert: %w", err)
	}
	defer candidateStmt.Close()
	matchStmt, err := tx.PrepareContext(ctx, `
		INSERT INTO scanner_matches(candidate_id, item_id, state, confidence, needs_review, extracted_part_number, updated_at)
		VALUES (?, '', 'not_in_collection', ?, 1, ?, CURRENT_TIMESTAMP)
	`)
	if err != nil {
		return nil, fmt.Errorf("prepare match insert: %w", err)
	}
	defer matchStmt.Close()
	sources := []string{"ebay", "bonza", "frontline", "andrews"}
	for i := 0; i < candidates; i++ {
		idx := offset + i + 1
		candidateID := fmt.Sprintf("scale-candidate-%s-%07d", profileID, idx)
		part := fmt.Sprintf("PN-%s-%06d", strings.ToUpper(profileID), idx+len(itemIDs))
		title := fmt.Sprintf("Discovery %d %s", idx, part)
		if _, err := candidateStmt.ExecContext(
			ctx,
			candidateID,
			profileID,
			querySetID,
			fmt.Sprintf("LIST-%s-%08d", strings.ToUpper(profileID), idx),
			title,
			math.Round((10+rng.Float64()*140)*100)/100,
			math.Round((2+rng.Float64()*20)*100)/100,
			fmt.Sprintf("https://example.test/listing/%s/%d", profileID, idx),
			fmt.Sprintf("https://example.test/images/%s/%d.jpg", profileID, idx),
			fmt.Sprintf("seller-%03d", idx%500),
			sources[i%len(sources)],
		); err != nil {
			return nil, fmt.Errorf("insert candidate %d: %w", i+1, err)
		}
		if _, err := matchStmt.ExecContext(ctx, candidateID, 0.5+rng.Float64()*0.49, part); err != nil {
			return nil, fmt.Errorf("insert match %d: %w", i+1, err)
		}
	}
	return []string{querySetID}, nil
}

func insertWishlist(ctx context.Context, tx *sql.Tx, profileID string, itemIDs []string, count, offset int, _ *rand.Rand) error {
	if count <= 0 || len(itemIDs) == 0 {
		return nil
	}
	stmt, err := tx.PrepareContext(ctx, `
		INSERT OR IGNORE INTO wishlist_entries(id, profile_id, item_id, target_price, priority, notes, highlight_hit, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`)
	if err != nil {
		return fmt.Errorf("prepare wishlist insert: %w", err)
	}
	defer stmt.Close()
	for i := 0; i < count && i < len(itemIDs); i++ {
		priority := "normal"
		if i%10 == 0 {
			priority = "high"
		}
		idx := offset + i + 1
		if _, err := stmt.ExecContext(
			ctx,
			fmt.Sprintf("scale-wishlist-%07d", idx),
			profileID,
			itemIDs[i],
			float64(15+(i%200)),
			priority,
			fmt.Sprintf("wishlist note %d", i+1),
		); err != nil {
			return fmt.Errorf("insert wishlist %d: %w", i+1, err)
		}
	}
	return nil
}

func insertPricingSnapshots(ctx context.Context, tx *sql.Tx, profileID string, itemIDs []string, wishlistCount, months, offset int, rng *rand.Rand) error {
	if months <= 0 || len(itemIDs) == 0 {
		return nil
	}
	tracked := wishlistCount
	if tracked <= 0 {
		tracked = min(100, len(itemIDs))
	}
	if tracked > len(itemIDs) {
		tracked = len(itemIDs)
	}
	if tracked <= 0 {
		return nil
	}
	trackStmt, err := tx.PrepareContext(ctx, `INSERT OR IGNORE INTO tracked_items(item_id, profile_id, created_at) VALUES (?, ?, CURRENT_TIMESTAMP)`)
	if err != nil {
		return fmt.Errorf("prepare tracked insert: %w", err)
	}
	defer trackStmt.Close()
	snapshotStmt, err := tx.PrepareContext(ctx, `
		INSERT INTO price_snapshots(id, item_id, snapshot_date, source, min_price, median_price, latest_price, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`)
	if err != nil {
		return fmt.Errorf("prepare snapshot insert: %w", err)
	}
	defer snapshotStmt.Close()
	days := months * 30
	if days <= 0 {
		days = 1
	}
	sources := []string{"ebay", "bonza", "frontline", "andrews", "mrtoys"}
	for i := 0; i < tracked; i++ {
		itemID := itemIDs[i]
		if _, err := trackStmt.ExecContext(ctx, itemID, profileID); err != nil {
			return fmt.Errorf("insert tracked item: %w", err)
		}
		base := 20 + rng.Float64()*140
		for d := 0; d < days; d++ {
			day := time.Now().AddDate(0, 0, -days+d).Format("2006-01-02")
			wave := math.Sin(float64(d)/7.0) * 3
			latest := math.Round((base+wave+rng.Float64()*2)*100) / 100
			median := math.Round((latest+(rng.Float64()*2-1))*100) / 100
			minPrice := math.Round((math.Min(latest, median)-rng.Float64())*100) / 100
			if minPrice < 0.01 {
				minPrice = 0.01
			}
			id := fmt.Sprintf("scale-price-%s-%d-%d-%d", profileID, offset+i+1, d+1, i%len(sources))
			if _, err := snapshotStmt.ExecContext(ctx, id, itemID, day, sources[(i+d)%len(sources)], minPrice, median, latest); err != nil {
				return fmt.Errorf("insert price snapshot: %w", err)
			}
		}
	}
	return nil
}

func exportSnapshot(ctx context.Context, conn *sql.DB, path string, summary Summary) error {
	screens := map[string]any{
		"dashboard": map[string]any{
			"new_discoveries": summary.DiscoveryCandidates,
			"wishlist_hits":   summary.WishlistEntries,
			"price_snapshots": summary.PriceSnapshots,
		},
		"collection": map[string]any{
			"items":     summary.Items,
			"instances": summary.Instances,
			"photos":    summary.Photos,
			"barcodes":  summary.Barcodes,
		},
	}
	payload := map[string]any{
		"profile_id":       summary.ProfileID,
		"dataset_profile":  summary.DatasetProfile,
		"mode":             summary.Mode,
		"seed":             summary.Seed,
		"generated_at":     time.Now().UTC().Format(time.RFC3339),
		"summary":          summary,
		"screen_fixtures":  screens,
		"per_screen_counts": map[string]any{
			"discoveries": summary.DiscoveryCandidates,
			"pricing":     summary.PriceSnapshots,
			"settings":    1,
		},
	}
	out, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal snapshot: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir snapshot dir: %w", err)
	}
	if err := os.WriteFile(path, out, 0o644); err != nil {
		return fmt.Errorf("write snapshot: %w", err)
	}
	_ = ctx
	_ = conn
	return nil
}

func buildNames(prefix string, count int) []string {
	out := make([]string, 0, count)
	for i := 0; i < count; i++ {
		out = append(out, fmt.Sprintf("%s-%02d", prefix, i+1))
	}
	return out
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func stableSeed(profileID string, seed int64) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(profileID))
	return int64(h.Sum64()) ^ seed
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
