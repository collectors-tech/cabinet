package wishlist

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type Entry struct {
	ID             string  `json:"id"`
	ItemID         string  `json:"item_id"`
	TargetPrice    float64 `json:"target_price"`
	Priority       string  `json:"priority"`
	Notes          string  `json:"notes"`
	HighlightHit   bool    `json:"highlight_hit"`
	BelowTargetNow bool    `json:"below_target_now"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

type Hit struct {
	CandidateID string  `json:"candidate_id"`
	ItemID      string  `json:"item_id"`
	Title       string  `json:"title"`
	Price       float64 `json:"price"`
	URL         string  `json:"url"`
	LastSeen    string  `json:"last_seen"`
}

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) Create(ctx context.Context, in Entry) (Entry, error) {
	if strings.TrimSpace(in.ItemID) == "" {
		return Entry{}, fmt.Errorf("item_id is required")
	}
	if strings.TrimSpace(in.Priority) == "" {
		in.Priority = "normal"
	}
	in.ID = uuid.NewString()
	highlight := 0
	if in.HighlightHit {
		highlight = 1
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO wishlist_entries(id, item_id, target_price, priority, notes, highlight_hit)
		VALUES (?, ?, ?, ?, ?, ?)
	`, in.ID, in.ItemID, in.TargetPrice, in.Priority, in.Notes, highlight)
	if err != nil {
		return Entry{}, fmt.Errorf("create wishlist entry: %w", err)
	}
	return s.GetByID(ctx, in.ID)
}

func (s *Service) Update(ctx context.Context, in Entry) error {
	if strings.TrimSpace(in.ID) == "" {
		return fmt.Errorf("id is required")
	}
	highlight := 0
	if in.HighlightHit {
		highlight = 1
	}
	_, err := s.db.ExecContext(ctx, `
		UPDATE wishlist_entries
		SET target_price = ?, priority = ?, notes = ?, highlight_hit = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, in.TargetPrice, in.Priority, in.Notes, highlight, in.ID)
	return err
}

func (s *Service) Delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM wishlist_entries WHERE id = ?`, id)
	return err
}

func (s *Service) GetByID(ctx context.Context, id string) (Entry, error) {
	var e Entry
	var highlight int
	err := s.db.QueryRowContext(ctx, `
		SELECT id, item_id, target_price, priority, notes, highlight_hit, created_at, updated_at
		FROM wishlist_entries WHERE id = ?
	`, id).Scan(&e.ID, &e.ItemID, &e.TargetPrice, &e.Priority, &e.Notes, &highlight, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return Entry{}, fmt.Errorf("wishlist entry not found")
		}
		return Entry{}, err
	}
	e.HighlightHit = highlight == 1
	e.BelowTargetNow = s.isBelowTarget(ctx, e.ItemID, e.TargetPrice)
	return e, nil
}

func (s *Service) List(ctx context.Context) ([]Entry, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id FROM wishlist_entries ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Entry
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		e, err := s.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *Service) Hits(ctx context.Context, itemID string) ([]Hit, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT c.id, w.item_id, c.title, c.price, c.url, c.last_seen
		FROM wishlist_entries w
		JOIN canonical_items i ON i.id = w.item_id
		JOIN scanner_candidates c ON LOWER(c.title) LIKE '%' || LOWER(i.part_number) || '%'
		WHERE w.highlight_hit = 1 AND (? = '' OR w.item_id = ?)
		ORDER BY c.last_seen DESC
	`, itemID, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Hit
	for rows.Next() {
		var h Hit
		if err := rows.Scan(&h.CandidateID, &h.ItemID, &h.Title, &h.Price, &h.URL, &h.LastSeen); err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

func (s *Service) isBelowTarget(ctx context.Context, itemID string, target float64) bool {
	if target <= 0 {
		return false
	}
	var part string
	if err := s.db.QueryRowContext(ctx, `SELECT part_number FROM canonical_items WHERE id = ?`, itemID).Scan(&part); err != nil {
		return false
	}
	var minPrice float64
	err := s.db.QueryRowContext(ctx, `
		SELECT COALESCE(MIN(price), 0)
		FROM scanner_candidates
		WHERE LOWER(title) LIKE '%' || LOWER(?) || '%'
	`, part).Scan(&minPrice)
	if err != nil {
		return false
	}
	return minPrice > 0 && minPrice <= target
}
