package datamgmt

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Snapshot struct {
	SchemaVersion int            `json:"schema_version"`
	ExportedAt    string         `json:"exported_at"`
	Items         []SnapshotItem `json:"items"`
}

type SnapshotItem struct {
	Brand       string             `json:"brand"`
	Category    string             `json:"category"`
	PartNumber  string             `json:"part_number"`
	Title       string             `json:"title"`
	Make        string             `json:"make"`
	Model       string             `json:"model"`
	Year        string             `json:"year"`
	Scale       string             `json:"scale"`
	Series      string             `json:"series"`
	Description string             `json:"description"`
	Tags        []string           `json:"tags"`
	Barcodes    []string           `json:"barcodes"`
	Instances   []SnapshotInstance `json:"instances"`
}

type SnapshotInstance struct {
	Condition        string  `json:"condition"`
	Status           string  `json:"status"`
	Quantity         int     `json:"quantity"`
	StorageLocation  string  `json:"storage_location"`
	AcquisitionPrice float64 `json:"acquisition_price"`
	AcquisitionDate  string  `json:"acquisition_date"`
	Notes            string  `json:"notes"`
}

type DryRunSummary struct {
	TotalItems      int               `json:"total_items"`
	NewItems        int               `json:"new_items"`
	Conflicts       int               `json:"conflicts"`
	ConflictDetails []ConflictDetails `json:"conflict_details"`
}

type ConflictDetails struct {
	PartNumber string `json:"part_number"`
	ExistingID string `json:"existing_id"`
}

type ApplyOptions struct {
	DefaultAction string            `json:"default_action"`
	Overrides     map[string]string `json:"overrides"`
}

type ApplySummary struct {
	TotalItems int `json:"total_items"`
	Created    int `json:"created"`
	Merged     int `json:"merged"`
	Skipped    int `json:"skipped"`
	Failed     int `json:"failed"`
}

type CSVImportRequest struct {
	CSV     string            `json:"csv"`
	Mapping map[string]string `json:"mapping"`
}

type ReindexSummary struct {
	OK                 bool   `json:"ok"`
	Operation          string `json:"operation"`
	RebuiltSearchIndex bool   `json:"rebuilt_search_index"`
	CompletedAt        string `json:"completed_at"`
}

type RepairSummary struct {
	OK             bool   `json:"ok"`
	Operation      string `json:"operation"`
	IntegrityCheck string `json:"integrity_check"`
	CompletedAt    string `json:"completed_at"`
}

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) ExportSnapshot(ctx context.Context) (Snapshot, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, brand, category, part_number, title, make, model, year, scale, series, description, tags_json
		FROM canonical_items
		ORDER BY created_at ASC
	`)
	if err != nil {
		return Snapshot{}, fmt.Errorf("list items for export: %w", err)
	}
	defer rows.Close()

	out := Snapshot{
		SchemaVersion: 1,
		ExportedAt:    time.Now().UTC().Format(time.RFC3339),
		Items:         []SnapshotItem{},
	}

	for rows.Next() {
		var itemID string
		var item SnapshotItem
		var tagsRaw string
		if err := rows.Scan(&itemID, &item.Brand, &item.Category, &item.PartNumber, &item.Title, &item.Make, &item.Model, &item.Year, &item.Scale, &item.Series, &item.Description, &tagsRaw); err != nil {
			return Snapshot{}, fmt.Errorf("scan export item: %w", err)
		}
		if tagsRaw != "" {
			_ = json.Unmarshal([]byte(tagsRaw), &item.Tags)
		}
		barcodes, err := s.loadBarcodes(ctx, itemID)
		if err != nil {
			return Snapshot{}, err
		}
		item.Barcodes = barcodes
		instances, err := s.loadInstances(ctx, itemID)
		if err != nil {
			return Snapshot{}, err
		}
		item.Instances = instances
		out.Items = append(out.Items, item)
	}
	if err := rows.Err(); err != nil {
		return Snapshot{}, fmt.Errorf("iterate export items: %w", err)
	}
	return out, nil
}

func (s *Service) ExportItemsCSV(ctx context.Context) (string, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT brand, category, part_number, title, make, model, year, scale, series, description
		FROM canonical_items ORDER BY created_at ASC
	`)
	if err != nil {
		return "", fmt.Errorf("list items for csv: %w", err)
	}
	defer rows.Close()

	var b strings.Builder
	w := csv.NewWriter(&b)
	if err := w.Write([]string{"brand", "category", "part_number", "title", "make", "model", "year", "scale", "series", "description"}); err != nil {
		return "", fmt.Errorf("write csv header: %w", err)
	}
	for rows.Next() {
		record := make([]string, 10)
		if err := rows.Scan(&record[0], &record[1], &record[2], &record[3], &record[4], &record[5], &record[6], &record[7], &record[8], &record[9]); err != nil {
			return "", fmt.Errorf("scan csv row: %w", err)
		}
		if err := w.Write(record); err != nil {
			return "", fmt.Errorf("write csv row: %w", err)
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return "", fmt.Errorf("flush csv: %w", err)
	}
	return b.String(), nil
}

func (s *Service) ParseCSVToSnapshot(req CSVImportRequest) (Snapshot, error) {
	mapping := req.Mapping
	if mapping == nil {
		mapping = map[string]string{}
	}
	reader := csv.NewReader(strings.NewReader(req.CSV))
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err != nil {
		if err == io.EOF {
			return Snapshot{SchemaVersion: 1, ExportedAt: time.Now().UTC().Format(time.RFC3339), Items: []SnapshotItem{}}, nil
		}
		return Snapshot{}, fmt.Errorf("read csv header: %w", err)
	}
	index := map[string]int{}
	for i, h := range header {
		index[strings.TrimSpace(strings.ToLower(h))] = i
	}
	resolve := func(field, fallback string) (int, error) {
		key := strings.TrimSpace(strings.ToLower(mapping[field]))
		if key == "" {
			key = fallback
		}
		i, ok := index[key]
		if !ok {
			return 0, fmt.Errorf("missing mapped csv column for %q (%s)", field, key)
		}
		return i, nil
	}
	resolveOptional := func(field, fallback string) int {
		key := strings.TrimSpace(strings.ToLower(mapping[field]))
		if key == "" {
			key = fallback
		}
		i, ok := index[key]
		if !ok {
			return -1
		}
		return i
	}

	brandIdx, err := resolve("brand", "brand")
	if err != nil {
		return Snapshot{}, err
	}
	categoryIdx, err := resolve("category", "category")
	if err != nil {
		return Snapshot{}, err
	}
	partIdx, err := resolve("part_number", "part_number")
	if err != nil {
		return Snapshot{}, err
	}
	titleIdx, err := resolve("title", "title")
	if err != nil {
		return Snapshot{}, err
	}
	makeIdx := resolveOptional("make", "make")
	modelIdx := resolveOptional("model", "model")
	yearIdx := resolveOptional("year", "year")
	scaleIdx := resolveOptional("scale", "scale")
	seriesIdx := resolveOptional("series", "series")
	descIdx := resolveOptional("description", "description")

	snap := Snapshot{
		SchemaVersion: 1,
		ExportedAt:    time.Now().UTC().Format(time.RFC3339),
		Items:         []SnapshotItem{},
	}
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return Snapshot{}, fmt.Errorf("read csv record: %w", err)
		}
		get := func(i int) string {
			if i >= 0 && i < len(record) {
				return strings.TrimSpace(record[i])
			}
			return ""
		}
		item := SnapshotItem{
			Brand:       get(brandIdx),
			Category:    get(categoryIdx),
			PartNumber:  get(partIdx),
			Title:       get(titleIdx),
			Make:        get(makeIdx),
			Model:       get(modelIdx),
			Year:        get(yearIdx),
			Scale:       get(scaleIdx),
			Series:      get(seriesIdx),
			Description: get(descIdx),
			Tags:        []string{},
			Barcodes:    []string{},
			Instances:   []SnapshotInstance{},
		}
		if item.Brand == "" || item.Category == "" || item.PartNumber == "" || item.Title == "" {
			continue
		}
		snap.Items = append(snap.Items, item)
	}
	return snap, nil
}

func (s *Service) DryRunImport(ctx context.Context, snap Snapshot) (DryRunSummary, error) {
	if snap.SchemaVersion != 1 {
		return DryRunSummary{}, fmt.Errorf("unsupported schema version: %d", snap.SchemaVersion)
	}

	sum := DryRunSummary{
		TotalItems:      len(snap.Items),
		ConflictDetails: []ConflictDetails{},
	}

	for _, item := range snap.Items {
		partNumber := strings.TrimSpace(item.PartNumber)
		existingID, err := s.findItemIDByPartNumber(ctx, partNumber)
		if err != nil {
			return DryRunSummary{}, err
		}
		if existingID == "" {
			sum.NewItems++
			continue
		}
		sum.Conflicts++
		sum.ConflictDetails = append(sum.ConflictDetails, ConflictDetails{
			PartNumber: partNumber,
			ExistingID: existingID,
		})
	}
	return sum, nil
}

func (s *Service) ApplyImport(ctx context.Context, snap Snapshot, opts ApplyOptions) (ApplySummary, error) {
	if snap.SchemaVersion != 1 {
		return ApplySummary{TotalItems: len(snap.Items), Failed: len(snap.Items)}, fmt.Errorf("unsupported schema version: %d", snap.SchemaVersion)
	}
	sum := ApplySummary{TotalItems: len(snap.Items)}
	if opts.DefaultAction == "" {
		opts.DefaultAction = "merge"
	}
	if opts.Overrides == nil {
		opts.Overrides = map[string]string{}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		sum = failedApplySummary(sum)
		return sum, fmt.Errorf("begin import tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	for _, item := range snap.Items {
		partNumber := strings.TrimSpace(item.PartNumber)
		existingID, err := s.findItemIDByPartNumberTx(ctx, tx, partNumber)
		if err != nil {
			sum = failedApplySummary(sum)
			return sum, err
		}

		action := opts.DefaultAction
		if override := strings.TrimSpace(opts.Overrides[partNumber]); override != "" {
			action = override
		}
		action = strings.ToLower(strings.TrimSpace(action))
		if action != "merge" && action != "create" && action != "skip" {
			sum = failedApplySummary(sum)
			return sum, fmt.Errorf("invalid action %q for part_number %q", action, partNumber)
		}

		targetItemID := existingID
		if existingID == "" || action == "create" {
			targetItemID, err = s.insertItemTx(ctx, tx, item, action == "create")
			if err != nil {
				sum = failedApplySummary(sum)
				return sum, err
			}
			existingID = targetItemID
			sum.Created++
		} else if action == "skip" {
			sum.Skipped++
			continue
		} else {
			sum.Merged++
		}

		if err := s.mergeChildrenTx(ctx, tx, targetItemID, item); err != nil {
			sum = failedApplySummary(sum)
			return sum, err
		}
	}

	if err := tx.Commit(); err != nil {
		sum = failedApplySummary(sum)
		return sum, fmt.Errorf("commit import tx: %w", err)
	}
	return sum, nil
}

func failedApplySummary(sum ApplySummary) ApplySummary {
	sum.Created = 0
	sum.Merged = 0
	sum.Skipped = 0
	sum.Failed = sum.TotalItems
	return sum
}

func (s *Service) Reindex(ctx context.Context) (ReindexSummary, error) {
	if _, err := s.db.ExecContext(ctx, `INSERT INTO canonical_items_fts(canonical_items_fts) VALUES('rebuild')`); err != nil {
		return ReindexSummary{}, fmt.Errorf("reindex fts: %w", err)
	}
	return ReindexSummary{
		OK:                 true,
		Operation:          "reindex_search",
		RebuiltSearchIndex: true,
		CompletedAt:        time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (s *Service) Repair(ctx context.Context) (RepairSummary, error) {
	var result string
	if err := s.db.QueryRowContext(ctx, `PRAGMA integrity_check`).Scan(&result); err != nil {
		return RepairSummary{}, fmt.Errorf("integrity check: %w", err)
	}
	return RepairSummary{
		OK:             strings.EqualFold(strings.TrimSpace(result), "ok"),
		Operation:      "integrity_check",
		IntegrityCheck: result,
		CompletedAt:    time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (s *Service) loadBarcodes(ctx context.Context, itemID string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT barcode FROM item_barcodes WHERE item_id = ? ORDER BY created_at ASC`, itemID)
	if err != nil {
		return nil, fmt.Errorf("load barcodes: %w", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var b string
		if err := rows.Scan(&b); err != nil {
			return nil, fmt.Errorf("scan barcode: %w", err)
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (s *Service) loadInstances(ctx context.Context, itemID string) ([]SnapshotInstance, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT condition, status, quantity, storage_location, acquisition_price, acquisition_date, notes
		FROM instances WHERE item_id = ? ORDER BY created_at ASC
	`, itemID)
	if err != nil {
		return nil, fmt.Errorf("load instances: %w", err)
	}
	defer rows.Close()
	var out []SnapshotInstance
	for rows.Next() {
		var in SnapshotInstance
		if err := rows.Scan(&in.Condition, &in.Status, &in.Quantity, &in.StorageLocation, &in.AcquisitionPrice, &in.AcquisitionDate, &in.Notes); err != nil {
			return nil, fmt.Errorf("scan instance: %w", err)
		}
		out = append(out, in)
	}
	return out, rows.Err()
}

func (s *Service) findItemIDByPartNumber(ctx context.Context, partNumber string) (string, error) {
	var id string
	err := s.db.QueryRowContext(ctx, `SELECT id FROM canonical_items WHERE part_number = ?`, partNumber).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("find item by part number: %w", err)
	}
	return id, nil
}

func (s *Service) findItemIDByPartNumberTx(ctx context.Context, tx *sql.Tx, partNumber string) (string, error) {
	var id string
	err := tx.QueryRowContext(ctx, `SELECT id FROM canonical_items WHERE part_number = ?`, partNumber).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("find item by part number tx: %w", err)
	}
	return id, nil
}

func (s *Service) insertItemTx(ctx context.Context, tx *sql.Tx, item SnapshotItem, forceNewPartNumber bool) (string, error) {
	itemID := uuid.NewString()
	partNumber := strings.TrimSpace(item.PartNumber)
	if partNumber == "" {
		partNumber = "IMPORTED-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	if forceNewPartNumber {
		partNumber = partNumber + "-IMP-" + itemID[:8]
	}
	tagsRaw, _ := json.Marshal(item.Tags)
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO canonical_items (id, brand, category, part_number, title, make, model, year, scale, series, description, tags_json)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, itemID, item.Brand, item.Category, partNumber, item.Title, item.Make, item.Model, item.Year, item.Scale, item.Series, item.Description, string(tagsRaw)); err != nil {
		return "", fmt.Errorf("insert import item: %w", err)
	}
	return itemID, nil
}

func (s *Service) mergeChildrenTx(ctx context.Context, tx *sql.Tx, itemID string, item SnapshotItem) error {
	for _, b := range item.Barcodes {
		b = strings.TrimSpace(b)
		if b == "" {
			continue
		}
		var count int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(1) FROM item_barcodes WHERE item_id = ? AND barcode = ?`, itemID, b).Scan(&count); err != nil {
			return fmt.Errorf("check barcode duplicate: %w", err)
		}
		if count == 0 {
			if _, err := tx.ExecContext(ctx, `INSERT INTO item_barcodes (id, item_id, barcode) VALUES (?, ?, ?)`, uuid.NewString(), itemID, b); err != nil {
				return fmt.Errorf("insert import barcode: %w", err)
			}
		}
	}
	for _, in := range item.Instances {
		qty := in.Quantity
		if qty <= 0 {
			qty = 1
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO instances (id, item_id, condition, status, quantity, storage_location, acquisition_price, acquisition_date, notes)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, uuid.NewString(), itemID, in.Condition, in.Status, qty, in.StorageLocation, in.AcquisitionPrice, in.AcquisitionDate, in.Notes); err != nil {
			return fmt.Errorf("insert import instance: %w", err)
		}
	}
	return nil
}
