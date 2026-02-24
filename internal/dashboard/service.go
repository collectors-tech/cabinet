package dashboard

import (
	"context"
	"database/sql"
	"fmt"
)

type Card struct {
	Title string `json:"title"`
	Value int    `json:"value"`
	Link  string `json:"link"`
}

type CollectionStats struct {
	TotalItems     int     `json:"total_items"`
	TotalInstances int     `json:"total_instances"`
	EstimatedValue float64 `json:"estimated_value"`
}

type Summary struct {
	NewDiscoveries int             `json:"new_discoveries"`
	WishlistHits   int             `json:"wishlist_hits"`
	PriceDrops     int             `json:"price_drops"`
	RecentlyAdded  []string        `json:"recently_added"`
	TotalItems     int             `json:"total_items"`
	TotalInstances int             `json:"total_instances"`
	EstimatedValue float64         `json:"estimated_value"`
	Collection     CollectionStats `json:"collection"`
	Cards          []Card          `json:"cards"`
}

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) Summary(ctx context.Context) (Summary, error) {
	out := Summary{}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM scanner_matches WHERE state = 'not_in_collection'`).Scan(&out.NewDiscoveries); err != nil {
		return Summary{}, fmt.Errorf("discoveries count: %w", err)
	}
	if err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(1)
		FROM wishlist_entries w
		JOIN canonical_items i ON i.id = w.item_id
		JOIN scanner_candidates c ON LOWER(c.title) LIKE '%' || LOWER(i.part_number) || '%'
		WHERE w.highlight_hit = 1
	`).Scan(&out.WishlistHits); err != nil {
		return Summary{}, fmt.Errorf("wishlist hits count: %w", err)
	}
	if err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(1)
		FROM (
			SELECT item_id, source, snapshot_date, latest_price,
				LAG(latest_price) OVER (PARTITION BY item_id, source ORDER BY snapshot_date) AS prev_price
			FROM price_snapshots
		) t
		WHERE prev_price IS NOT NULL AND latest_price < prev_price
	`).Scan(&out.PriceDrops); err != nil {
		return Summary{}, fmt.Errorf("price drops count: %w", err)
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM canonical_items`).Scan(&out.Collection.TotalItems); err != nil {
		return Summary{}, fmt.Errorf("items count: %w", err)
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(quantity),0), COALESCE(SUM(quantity*acquisition_price),0) FROM instances`).Scan(&out.Collection.TotalInstances, &out.Collection.EstimatedValue); err != nil {
		return Summary{}, fmt.Errorf("instances stats: %w", err)
	}
	rows, err := s.db.QueryContext(ctx, `SELECT title FROM canonical_items ORDER BY created_at DESC LIMIT 5`)
	if err != nil {
		return Summary{}, fmt.Errorf("recently added query: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			return Summary{}, fmt.Errorf("scan recently added: %w", err)
		}
		out.RecentlyAdded = append(out.RecentlyAdded, t)
	}
	if err := rows.Err(); err != nil {
		return Summary{}, fmt.Errorf("iterate recently added: %w", err)
	}
	out.TotalItems = out.Collection.TotalItems
	out.TotalInstances = out.Collection.TotalInstances
	out.EstimatedValue = out.Collection.EstimatedValue
	out.Cards = []Card{
		{Title: "New Discoveries", Value: out.NewDiscoveries, Link: "/discoveries"},
		{Title: "Wishlist Hits", Value: out.WishlistHits, Link: "/wishlist"},
		{Title: "Price Drops", Value: out.PriceDrops, Link: "/pricing"},
		{Title: "Recently Added", Value: len(out.RecentlyAdded), Link: "/collection"},
	}
	return out, nil
}
