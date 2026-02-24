package discovery

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type Filter struct {
	Query    string
	PriceMax float64
	DateFrom string
}

type Item struct {
	CandidateID string  `json:"candidate_id"`
	Title       string  `json:"title"`
	Price       float64 `json:"price"`
	URL         string  `json:"url"`
	LastSeen    string  `json:"last_seen"`
	StockState  string  `json:"stock_state"`
	StockCount  int     `json:"stock_count"`
}

type ActionType string

const (
	ActionIgnore      ActionType = "ignore"
	ActionAddWishlist ActionType = "add_to_wishlist"
	ActionTrackPrice  ActionType = "track_price"
	ActionCreateItem  ActionType = "create_item"
)

type Action struct {
	CandidateID string         `json:"candidate_id"`
	Type        ActionType     `json:"type"`
	Payload     map[string]any `json:"payload"`
}

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) ListNotInCollection(ctx context.Context, f Filter) ([]Item, error) {
	q := `
		SELECT c.id, c.title, c.price, c.url, c.last_seen, c.stock_state, c.stock_count
		FROM scanner_candidates c
		JOIN scanner_matches m ON m.candidate_id = c.id
		LEFT JOIN ignored_candidates i ON i.candidate_id = c.id
		WHERE m.state = 'not_in_collection' AND i.candidate_id IS NULL
	`
	args := []any{}
	if strings.TrimSpace(f.Query) != "" {
		q += ` AND LOWER(c.title) LIKE ?`
		args = append(args, "%"+strings.ToLower(strings.TrimSpace(f.Query))+"%")
	}
	if f.PriceMax > 0 {
		q += ` AND c.price <= ?`
		args = append(args, f.PriceMax)
	}
	if strings.TrimSpace(f.DateFrom) != "" {
		q += ` AND c.last_seen >= ?`
		args = append(args, strings.TrimSpace(f.DateFrom))
	}
	q += ` ORDER BY c.last_seen DESC`
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list not_in_collection: %w", err)
	}
	defer rows.Close()
	var out []Item
	for rows.Next() {
		var it Item
		if err := rows.Scan(&it.CandidateID, &it.Title, &it.Price, &it.URL, &it.LastSeen, &it.StockState, &it.StockCount); err != nil {
			return nil, fmt.Errorf("scan not_in_collection row: %w", err)
		}
		out = append(out, it)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate not_in_collection rows: %w", err)
	}
	return out, nil
}

func (s *Service) ApplyAction(ctx context.Context, a Action) error {
	if strings.TrimSpace(a.CandidateID) == "" {
		return fmt.Errorf("candidate_id is required")
	}
	if a.Payload == nil {
		a.Payload = map[string]any{}
	}
	raw, _ := json.Marshal(a.Payload)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO discovery_actions(id, candidate_id, action_type, payload_json)
		VALUES (?, ?, ?, ?)
	`, uuid.NewString(), a.CandidateID, string(a.Type), string(raw))
	if err != nil {
		return fmt.Errorf("insert discovery action: %w", err)
	}

	switch a.Type {
	case ActionIgnore:
		_, err = s.db.ExecContext(ctx, `
			INSERT INTO ignored_candidates(candidate_id, ignored_at)
			VALUES (?, CURRENT_TIMESTAMP)
			ON CONFLICT(candidate_id) DO UPDATE SET ignored_at = CURRENT_TIMESTAMP
		`, a.CandidateID)
		if err != nil {
			return fmt.Errorf("ignore candidate: %w", err)
		}
	case ActionTrackPrice:
		var itemID string
		if err := s.db.QueryRowContext(ctx, `SELECT item_id FROM scanner_matches WHERE candidate_id = ?`, a.CandidateID).Scan(&itemID); err == nil && strings.TrimSpace(itemID) != "" {
			_, _ = s.db.ExecContext(ctx, `INSERT INTO tracked_items(item_id, created_at) VALUES (?, CURRENT_TIMESTAMP) ON CONFLICT(item_id) DO NOTHING`, itemID)
		}
	case ActionAddWishlist:
		var itemID string
		if err := s.db.QueryRowContext(ctx, `SELECT item_id FROM scanner_matches WHERE candidate_id = ?`, a.CandidateID).Scan(&itemID); err == nil && strings.TrimSpace(itemID) != "" {
			_, _ = s.db.ExecContext(ctx, `
				INSERT INTO wishlist_entries(id, item_id, target_price, priority, notes, highlight_hit)
				VALUES (?, ?, ?, 'normal', '', 1)
				ON CONFLICT(item_id) DO NOTHING
			`, uuid.NewString(), itemID, 0.0)
		}
	case ActionCreateItem:
		var title, partNumber string
		if err := s.db.QueryRowContext(ctx, `
			SELECT c.title, m.extracted_part_number
			FROM scanner_candidates c
			JOIN scanner_matches m ON m.candidate_id = c.id
			WHERE c.id = ?
		`, a.CandidateID).Scan(&title, &partNumber); err == nil {
			if strings.TrimSpace(partNumber) == "" {
				partNumber = "AUTO-" + strings.ToUpper(uuid.NewString()[:8])
			}
			_, _ = s.db.ExecContext(ctx, `
				INSERT INTO canonical_items(id, brand, category, part_number, title)
				VALUES (?, ?, ?, ?, ?)
			`, uuid.NewString(), "Unknown", "Unknown", partNumber, title)
		}
	}
	return nil
}

func (s *Service) ResetIgnored(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM ignored_candidates`)
	return err
}
