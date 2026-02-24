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
	StockCount   int     `json:"stock_count"`
}

type TrendPoint struct {
	Date   string  `json:"date"`
	Latest float64 `json:"latest"`
}

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) TrackItem(ctx context.Context, itemID string) error {
	return s.TrackItemForProfile(ctx, "", itemID)
}

func (s *Service) TrackItemForProfile(ctx context.Context, profileID, itemID string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO tracked_items(item_id, profile_id, created_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(item_id) DO UPDATE SET profile_id = excluded.profile_id
	`, itemID, strings.TrimSpace(profileID))
	return err
}

func (s *Service) RunDailySnapshot(ctx context.Context) error {
	return s.RunDailySnapshotForProfile(ctx, "")
}

func (s *Service) RunDailySnapshotForProfile(ctx context.Context, profileID string) error {
	rows, err := s.db.QueryContext(ctx, `SELECT item_id FROM tracked_items WHERE (? = '' OR profile_id = ?)`, strings.TrimSpace(profileID), strings.TrimSpace(profileID))
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
		if err := s.snapshotItemForProfile(ctx, strings.TrimSpace(profileID), itemID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) snapshotItem(ctx context.Context, itemID string) error {
	return s.snapshotItemForProfile(ctx, "", itemID)
}

func (s *Service) snapshotItemForProfile(ctx context.Context, profileID, itemID string) error {
	var part string
	if err := s.db.QueryRowContext(ctx, `SELECT part_number FROM canonical_items WHERE id = ? AND (? = '' OR profile_id = ?)`, itemID, strings.TrimSpace(profileID), strings.TrimSpace(profileID)).Scan(&part); err != nil {
		return err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT source, price, stock_count
		FROM scanner_candidates
		WHERE LOWER(title) LIKE '%' || LOWER(?) || '%' AND (? = '' OR profile_id = ?)
		ORDER BY source ASC, last_seen ASC
	`, part, strings.TrimSpace(profileID), strings.TrimSpace(profileID))
	if err != nil {
		return err
	}
	defer rows.Close()
	bySource := map[string][]float64{}
	stockBySource := map[string]int{}
	var latest float64
	for rows.Next() {
		var source string
		var price float64
		var stockCount int
		if err := rows.Scan(&source, &price, &stockCount); err != nil {
			return err
		}
		if strings.TrimSpace(source) == "" {
			source = "unknown"
		}
		bySource[source] = append(bySource[source], price)
		stockBySource[source] = stockCount
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
			INSERT INTO price_snapshots(id, item_id, snapshot_date, source, min_price, median_price, latest_price, stock_count, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		`, uuid.NewString(), itemID, date, source, min, median, latest, stockBySource[source])
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) History(ctx context.Context, itemID string) ([]Snapshot, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT snapshot_date, source, min_price, median_price, latest_price, stock_count
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
		if err := rows.Scan(&snap.SnapshotDate, &snap.Source, &snap.MinPrice, &snap.MedianPrice, &snap.LatestPrice, &snap.StockCount); err != nil {
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
	if err := w.Write([]string{"snapshot_date", "min_price", "median_price", "latest_price", "stock_count", "source"}); err != nil {
		return "", err
	}
	for _, snap := range h {
		if err := w.Write([]string{
			snap.SnapshotDate,
			fmt.Sprintf("%.2f", snap.MinPrice),
			fmt.Sprintf("%.2f", snap.MedianPrice),
			fmt.Sprintf("%.2f", snap.LatestPrice),
			fmt.Sprintf("%d", snap.StockCount),
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

func (s *Service) Trend(ctx context.Context, itemID string) ([]TrendPoint, error) {
	history, err := s.History(ctx, itemID)
	if err != nil {
		return nil, err
	}
	byDate := map[string]TrendPoint{}
	for _, snap := range history {
		p := byDate[snap.SnapshotDate]
		p.Date = snap.SnapshotDate
		if snap.LatestPrice > p.Latest {
			p.Latest = snap.LatestPrice
		}
		byDate[snap.SnapshotDate] = p
	}
	var dates []string
	for d := range byDate {
		dates = append(dates, d)
	}
	sort.Strings(dates)
	out := make([]TrendPoint, 0, len(dates))
	for _, d := range dates {
		out = append(out, byDate[d])
	}
	return out, nil
}
