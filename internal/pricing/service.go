package pricing

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Snapshot struct {
	SnapshotDate string  `json:"snapshot_date"`
	Source       string  `json:"source"`
	MinPrice     float64 `json:"min_price"`
	MedianPrice  float64 `json:"median_price"`
	LatestPrice  float64 `json:"latest_price"`
}

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) TrackItem(ctx context.Context, itemID string) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO tracked_items(item_id, created_at) VALUES (?, CURRENT_TIMESTAMP) ON CONFLICT(item_id) DO NOTHING`, itemID)
	return err
}

func (s *Service) RunDailySnapshot(ctx context.Context) error {
	rows, err := s.db.QueryContext(ctx, `SELECT item_id FROM tracked_items`)
	if err != nil {
		return err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, itemID := range ids {
		if err := s.snapshotItem(ctx, itemID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) snapshotItem(ctx context.Context, itemID string) error {
	var part string
	if err := s.db.QueryRowContext(ctx, `SELECT part_number FROM canonical_items WHERE id = ?`, itemID).Scan(&part); err != nil {
		return err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT source, price
		FROM scanner_candidates
		WHERE LOWER(title) LIKE '%' || LOWER(?) || '%'
	`, part)
	if err != nil {
		return err
	}
	defer rows.Close()
	bySource := map[string][]float64{}
	var latest float64
	for rows.Next() {
		var source string
		var price float64
		if err := rows.Scan(&source, &price); err != nil {
			return err
		}
		if strings.TrimSpace(source) == "" {
			source = "unknown"
		}
		bySource[source] = append(bySource[source], price)
		latest = price
	}
	if err := rows.Err(); err != nil {
		return err
	}
	date := time.Now().UTC().Format("2006-01-02")
	for source, prices := range bySource {
		if len(prices) == 0 {
			continue
		}
		sort.Float64s(prices)
		min := prices[0]
		median := prices[len(prices)/2]
		_, err := s.db.ExecContext(ctx, `
			INSERT INTO price_snapshots(id, item_id, snapshot_date, source, min_price, median_price, latest_price, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		`, uuid.NewString(), itemID, date, source, min, median, latest)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) History(ctx context.Context, itemID string) ([]Snapshot, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT snapshot_date, source, min_price, median_price, latest_price
		FROM price_snapshots
		WHERE item_id = ?
		ORDER BY snapshot_date ASC, source ASC
	`, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Snapshot
	for rows.Next() {
		var snap Snapshot
		if err := rows.Scan(&snap.SnapshotDate, &snap.Source, &snap.MinPrice, &snap.MedianPrice, &snap.LatestPrice); err != nil {
			return nil, err
		}
		out = append(out, snap)
	}
	return out, rows.Err()
}

func (s *Service) BySource(ctx context.Context, itemID string) (map[string][]Snapshot, error) {
	h, err := s.History(ctx, itemID)
	if err != nil {
		return nil, err
	}
	out := map[string][]Snapshot{}
	for _, snap := range h {
		out[snap.Source] = append(out[snap.Source], snap)
	}
	return out, nil
}

func (s *Service) ExportCSV(ctx context.Context, itemID string) (string, error) {
	h, err := s.History(ctx, itemID)
	if err != nil {
		return "", err
	}
	b := &strings.Builder{}
	w := csv.NewWriter(b)
	if err := w.Write([]string{"snapshot_date", "min_price", "median_price", "latest_price", "source"}); err != nil {
		return "", err
	}
	for _, snap := range h {
		if err := w.Write([]string{
			snap.SnapshotDate,
			fmt.Sprintf("%.2f", snap.MinPrice),
			fmt.Sprintf("%.2f", snap.MedianPrice),
			fmt.Sprintf("%.2f", snap.LatestPrice),
			snap.Source,
		}); err != nil {
			return "", err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return "", err
	}
	return b.String(), nil
}
