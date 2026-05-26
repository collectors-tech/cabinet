package dashboard

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
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
	NewDiscoveries      int             `json:"new_discoveries"`
	WishlistHits        int             `json:"wishlist_hits"`
	PriceDrops          int             `json:"price_drops"`
	LowStockDiscoveries int             `json:"low_stock_discoveries"`
	Restocks            int             `json:"restocks"`
	RecentlyAdded       []string        `json:"recently_added"`
	TotalItems          int             `json:"total_items"`
	TotalInstances      int             `json:"total_instances"`
	EstimatedValue      float64         `json:"estimated_value"`
	Collection          CollectionStats `json:"collection"`
	Cards               []Card          `json:"cards"`
}

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) Summary(ctx context.Context, profileID string) (Summary, error) {
	profileID = strings.TrimSpace(profileID)
	out := Summary{}
	if err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(1)
		FROM scanner_matches m
		JOIN scanner_candidates c ON c.id = m.candidate_id
		WHERE m.state = 'not_in_collection' AND (? = '' OR c.profile_id = ?)
	`, profileID, profileID).Scan(&out.NewDiscoveries); err != nil {
		return Summary{}, fmt.Errorf("discoveries count: %w", err)
	}
	if err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(1)
		FROM wishlist_entries w
		JOIN canonical_items i ON i.id = w.item_id
		JOIN scanner_candidates c ON LOWER(c.title) LIKE '%' || LOWER(i.part_number) || '%'
		WHERE w.highlight_hit = 1 AND (? = '' OR w.profile_id = ?)
	`, profileID, profileID).Scan(&out.WishlistHits); err != nil {
		return Summary{}, fmt.Errorf("wishlist hits count: %w", err)
	}
	if err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(1)
		FROM (
			SELECT p.item_id, p.source, p.snapshot_date, p.latest_price,
				LAG(p.latest_price) OVER (PARTITION BY p.item_id, p.source ORDER BY p.snapshot_date) AS prev_price
			FROM price_snapshots p
			JOIN canonical_items i ON i.id = p.item_id
			WHERE ? = '' OR i.profile_id = ?
		) t
		WHERE prev_price IS NOT NULL AND latest_price < prev_price
	`, profileID, profileID).Scan(&out.PriceDrops); err != nil {
		return Summary{}, fmt.Errorf("price drops count: %w", err)
	}
	if err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(1)
		FROM scanner_candidates c
		JOIN scanner_matches m ON m.candidate_id = c.id
		WHERE m.state = 'not_in_collection' AND (? = '' OR c.profile_id = ?) AND (
			LOWER(c.stock_state) = 'low_stock' OR
			(c.stock_count > 0 AND c.stock_count <= 3)
		)
	`, profileID, profileID).Scan(&out.LowStockDiscoveries); err != nil {
		return Summary{}, fmt.Errorf("low stock discoveries count: %w", err)
	}
	if err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(1)
		FROM (
			SELECT p.item_id, p.source, p.snapshot_date, p.stock_count,
				LAG(p.stock_count) OVER (PARTITION BY p.item_id, p.source ORDER BY p.snapshot_date) AS prev_stock
			FROM price_snapshots p
			JOIN canonical_items i ON i.id = p.item_id
			WHERE ? = '' OR i.profile_id = ?
		) t
		WHERE prev_stock = 0 AND stock_count > 0
	`, profileID, profileID).Scan(&out.Restocks); err != nil {
		return Summary{}, fmt.Errorf("restock count: %w", err)
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM canonical_items WHERE ? = '' OR profile_id = ?`, profileID, profileID).Scan(&out.Collection.TotalItems); err != nil {
		return Summary{}, fmt.Errorf("items count: %w", err)
	}
	if err := s.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(inst.quantity),0), COALESCE(SUM(inst.quantity*inst.acquisition_price),0)
		FROM instances inst
		JOIN canonical_items i ON i.id = inst.item_id
		WHERE ? = '' OR i.profile_id = ?
	`, profileID, profileID).Scan(&out.Collection.TotalInstances, &out.Collection.EstimatedValue); err != nil {
		return Summary{}, fmt.Errorf("instances stats: %w", err)
	}
	rows, err := s.db.QueryContext(ctx, `SELECT title FROM canonical_items WHERE ? = '' OR profile_id = ? ORDER BY created_at DESC LIMIT 5`, profileID, profileID)
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
		{Title: "Low Stock", Value: out.LowStockDiscoveries, Link: "/discoveries"},
		{Title: "Wishlist Hits", Value: out.WishlistHits, Link: "/wishlist"},
		{Title: "Price Drops", Value: out.PriceDrops, Link: "/pricing"},
		{Title: "Restocks", Value: out.Restocks, Link: "/pricing"},
		{Title: "Recently Added", Value: len(out.RecentlyAdded), Link: "/collections"},
	}
	return out, nil
}
