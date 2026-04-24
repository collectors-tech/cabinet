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
	return s.CreateForProfile(ctx, "", in)
}

func (s *Service) CreateForProfile(ctx context.Context, profileID string, in Entry) (Entry, error) {
	if strings.TrimSpace(in.ItemID) == "" {
		return Entry{}, fmt.Errorf("item_id is required")
	}
	trimmedProfileID := strings.TrimSpace(profileID)
	in.Priority = normalizeWishlistPriority(in.Priority)
	in.ID = uuid.NewString()
	highlight := 0
	if in.HighlightHit {
		highlight = 1
	}
	belowTargetNow := 0
	if in.BelowTargetNow {
		belowTargetNow = 1
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO wishlist_entries(id, profile_id, item_id, target_price, priority, notes, highlight_hit, below_target_now)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, in.ID, trimmedProfileID, in.ItemID, in.TargetPrice, in.Priority, in.Notes, highlight, belowTargetNow)
	if err != nil {
		return Entry{}, fmt.Errorf("create wishlist entry: %w", err)
	}
	if err := s.syncItemWishlistState(ctx, trimmedProfileID, in.ItemID, "wishlist", in.Priority); err != nil {
		return Entry{}, err
	}
	return s.GetByIDForProfile(ctx, trimmedProfileID, in.ID)
}

func (s *Service) Update(ctx context.Context, in Entry) error {
	return s.UpdateForProfile(ctx, "", in)
}

func (s *Service) UpdateForProfile(ctx context.Context, profileID string, in Entry) error {
	if strings.TrimSpace(in.ID) == "" {
		return fmt.Errorf("id is required")
	}
	trimmedProfileID := strings.TrimSpace(profileID)
	in.Priority = normalizeWishlistPriority(in.Priority)
	itemID, err := s.itemIDForEntry(ctx, trimmedProfileID, in.ID)
	if err != nil {
		return err
	}
	highlight := 0
	if in.HighlightHit {
		highlight = 1
	}
	belowTargetNow := 0
	if in.BelowTargetNow {
		belowTargetNow = 1
	}
	_, err = s.db.ExecContext(ctx, `
		UPDATE wishlist_entries
		SET target_price = ?, priority = ?, notes = ?, highlight_hit = ?, below_target_now = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND (? = '' OR profile_id = ?)
	`, in.TargetPrice, in.Priority, in.Notes, highlight, belowTargetNow, in.ID, trimmedProfileID, trimmedProfileID)
	if err != nil {
		return err
	}
	return s.syncItemWishlistState(ctx, trimmedProfileID, itemID, "wishlist", in.Priority)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.DeleteForProfile(ctx, "", id)
}

func (s *Service) ConvertToOwnedForProfile(ctx context.Context, profileID, id string) error {
	trimmedProfileID := strings.TrimSpace(profileID)
	itemID, err := s.itemIDForEntry(ctx, trimmedProfileID, id)
	if err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `DELETE FROM wishlist_entries WHERE id = ? AND (? = '' OR profile_id = ?)`, strings.TrimSpace(id), trimmedProfileID, trimmedProfileID); err != nil {
		return err
	}
	return s.syncItemWishlistState(ctx, trimmedProfileID, itemID, "active", "")
}

func (s *Service) DeleteForProfile(ctx context.Context, profileID, id string) error {
	trimmedProfileID := strings.TrimSpace(profileID)
	itemID, _ := s.itemIDForEntry(ctx, trimmedProfileID, id)
	_, err := s.db.ExecContext(ctx, `DELETE FROM wishlist_entries WHERE id = ? AND (? = '' OR profile_id = ?)`, id, trimmedProfileID, trimmedProfileID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(itemID) == "" {
		return nil
	}
	return s.syncItemWishlistState(ctx, trimmedProfileID, itemID, "active", "")
}

func (s *Service) GetByID(ctx context.Context, id string) (Entry, error) {
	return s.GetByIDForProfile(ctx, "", id)
}

func (s *Service) GetByIDForProfile(ctx context.Context, profileID, id string) (Entry, error) {
	var e Entry
	var highlight int
	var belowTargetNow int
	err := s.db.QueryRowContext(ctx, `
		SELECT id, item_id, target_price, priority, notes, highlight_hit, below_target_now, created_at, updated_at
		FROM wishlist_entries WHERE id = ? AND (? = '' OR profile_id = ?)
	`, id, strings.TrimSpace(profileID), strings.TrimSpace(profileID)).Scan(&e.ID, &e.ItemID, &e.TargetPrice, &e.Priority, &e.Notes, &highlight, &belowTargetNow, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return Entry{}, fmt.Errorf("wishlist entry not found")
		}
		return Entry{}, err
	}
	e.HighlightHit = highlight == 1
	e.BelowTargetNow = belowTargetNow == 1 || s.isBelowTarget(ctx, e.ItemID, e.TargetPrice)
	return e, nil
}

func (s *Service) List(ctx context.Context) ([]Entry, error) {
	return s.ListByProfile(ctx, "")
}

func (s *Service) ListByProfile(ctx context.Context, profileID string) ([]Entry, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id FROM wishlist_entries WHERE (? = '' OR profile_id = ?) ORDER BY created_at ASC`, strings.TrimSpace(profileID), strings.TrimSpace(profileID))
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
		e, err := s.GetByIDForProfile(ctx, strings.TrimSpace(profileID), id)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *Service) Hits(ctx context.Context, itemID string) ([]Hit, error) {
	return s.HitsByProfile(ctx, "", itemID)
}

func (s *Service) HitsByProfile(ctx context.Context, profileID, itemID string) ([]Hit, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT c.id, w.item_id, c.title, c.price, c.url, c.last_seen
		FROM wishlist_entries w
		JOIN canonical_items i ON i.id = w.item_id
		JOIN scanner_candidates c ON LOWER(c.title) LIKE '%' || LOWER(i.part_number) || '%'
		WHERE w.highlight_hit = 1 AND (? = '' OR w.profile_id = ?) AND (? = '' OR w.item_id = ?)
		ORDER BY c.last_seen DESC
	`, strings.TrimSpace(profileID), strings.TrimSpace(profileID), itemID, itemID)
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

func normalizeWishlistPriority(raw string) string {
	priority := strings.ToLower(strings.TrimSpace(raw))
	if priority == "" || priority == "normal" {
		return "medium"
	}
	return priority
}

func (s *Service) itemIDForEntry(ctx context.Context, profileID, entryID string) (string, error) {
	var itemID string
	err := s.db.QueryRowContext(ctx, `
		SELECT item_id
		FROM wishlist_entries
		WHERE id = ? AND (? = '' OR profile_id = ?)
	`, entryID, strings.TrimSpace(profileID), strings.TrimSpace(profileID)).Scan(&itemID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("wishlist entry not found")
		}
		return "", err
	}
	return strings.TrimSpace(itemID), nil
}

func (s *Service) syncItemWishlistState(ctx context.Context, profileID, itemID, status, priority string) error {
	if strings.TrimSpace(itemID) == "" {
		return nil
	}
	_, err := s.db.ExecContext(ctx, `
		UPDATE canonical_items
		SET status = ?,
			priority = CASE WHEN ? = '' THEN priority ELSE ? END,
			updated_at = CURRENT_TIMESTAMP,
			updated_by = 'wishlist.service'
		WHERE id = ? AND (? = '' OR profile_id = ?)
	`, strings.TrimSpace(status), strings.TrimSpace(priority), strings.TrimSpace(priority), strings.TrimSpace(itemID), strings.TrimSpace(profileID), strings.TrimSpace(profileID))
	if err != nil {
		return fmt.Errorf("sync wishlist item status: %w", err)
	}
	return nil
}
