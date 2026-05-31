package discovery

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
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
	a.Payload = s.enrichDiscoveryActionPayload(ctx, a.CandidateID, a.Payload)
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
			var profileID, listingURL, seller, stockSignal, sourceProvider, querySetID, queryName, providerScopeJSON string
			var observedPrice float64
			scanErr := s.db.QueryRowContext(ctx, `
				SELECT c.profile_id, c.url, c.seller, c.stock_state, c.price, c.source, c.query_set_id, COALESCE(q.name, ''), COALESCE(q.provider_scope_json, '[]')
				FROM scanner_candidates c
				LEFT JOIN scanner_query_sets q ON q.id = c.query_set_id
				WHERE c.id = ?
			`, a.CandidateID).Scan(&profileID, &listingURL, &seller, &stockSignal, &observedPrice, &sourceProvider, &querySetID, &queryName, &providerScopeJSON)
			if scanErr != nil {
				_ = s.db.QueryRowContext(ctx, `
					SELECT url, seller, stock_state, price, source, query_set_id
					FROM scanner_candidates
					WHERE id = ?
				`, a.CandidateID).Scan(&listingURL, &seller, &stockSignal, &observedPrice, &sourceProvider, &querySetID)
			}
			profileID = strings.TrimSpace(profileID)
			if profileID == "" {
				_ = s.db.QueryRowContext(ctx, `SELECT profile_id FROM canonical_items WHERE id = ?`, itemID).Scan(&profileID)
				profileID = strings.TrimSpace(profileID)
			}
			metadata := buildDiscoveryMetadataNote(listingURL, seller, stockSignal, observedPrice, sourceProvider, querySetID, queryName, decodeStringArray(providerScopeJSON))
			var existingID, existingNotes string
			if err := s.db.QueryRowContext(ctx, `SELECT id, notes FROM wishlist_entries WHERE item_id = ? AND (? = '' OR profile_id = ?)`, itemID, profileID, profileID).Scan(&existingID, &existingNotes); err == nil {
				mergedNotes := mergeDiscoveryMetadataNotes(existingNotes, metadata)
				_, _ = s.db.ExecContext(ctx, `
					UPDATE wishlist_entries
					SET notes = ?, highlight_hit = 1, updated_at = CURRENT_TIMESTAMP
					WHERE id = ?
				`, mergedNotes, existingID)
				_, _ = s.db.ExecContext(ctx, `
					UPDATE canonical_items
					SET status = 'wishlist', updated_at = CURRENT_TIMESTAMP, updated_by = 'discovery.service'
					WHERE id = ? AND (? = '' OR profile_id = ?)
				`, itemID, profileID, profileID)
				break
			}
			_, _ = s.db.ExecContext(ctx, `
				INSERT INTO wishlist_entries(id, profile_id, item_id, target_price, priority, notes, highlight_hit)
				VALUES (?, ?, ?, ?, 'medium', ?, 1)
				ON CONFLICT(item_id) DO UPDATE SET notes = excluded.notes, highlight_hit = 1, updated_at = CURRENT_TIMESTAMP
			`, uuid.NewString(), profileID, itemID, 0.0, metadata)
			_, _ = s.db.ExecContext(ctx, `
				UPDATE canonical_items
				SET status = 'wishlist', priority = 'medium', updated_at = CURRENT_TIMESTAMP, updated_by = 'discovery.service'
				WHERE id = ? AND (? = '' OR profile_id = ?)
			`, itemID, profileID, profileID)
		}
	case ActionCreateItem:
		var title, partNumber, profileID, listingURL, seller, stockSignal, sourceProvider, querySetID, queryName, providerScopeJSON string
		var observedPrice float64
		if err := s.db.QueryRowContext(ctx, `
			SELECT c.title, m.extracted_part_number, c.profile_id, c.url, c.seller, c.stock_state, c.price, c.source, c.query_set_id, COALESCE(q.name, ''), COALESCE(q.provider_scope_json, '[]')
			FROM scanner_candidates c
			JOIN scanner_matches m ON m.candidate_id = c.id
			LEFT JOIN scanner_query_sets q ON q.id = c.query_set_id
			WHERE c.id = ?
		`, a.CandidateID).Scan(&title, &partNumber, &profileID, &listingURL, &seller, &stockSignal, &observedPrice, &sourceProvider, &querySetID, &queryName, &providerScopeJSON); err == nil {
			if strings.TrimSpace(partNumber) == "" {
				partNumber = "AUTO-" + strings.ToUpper(uuid.NewString()[:8])
			}
			metadata := buildDiscoveryMetadataNote(listingURL, seller, stockSignal, observedPrice, sourceProvider, querySetID, queryName, decodeStringArray(providerScopeJSON))
			sourceURLs, _ := json.Marshal([]string{strings.TrimSpace(listingURL)})
			_, _ = s.db.ExecContext(ctx, `
				INSERT INTO canonical_items(id, profile_id, brand, category, part_number, title, status, notes, source_urls_json, created_by, updated_by)
				VALUES (?, ?, ?, ?, ?, ?, 'active', ?, ?, 'discovery.service', 'discovery.service')
			`, uuid.NewString(), strings.TrimSpace(profileID), "Unknown", "Unknown", partNumber, title, metadata, string(sourceURLs))
		}
	}
	return nil
}

func (s *Service) enrichDiscoveryActionPayload(ctx context.Context, candidateID string, payload map[string]any) map[string]any {
	enriched := make(map[string]any, len(payload)+4)
	for key, value := range payload {
		enriched[key] = value
	}
	var sourceProvider, querySetID, queryName, providerScopeJSON string
	if err := s.db.QueryRowContext(ctx, `
		SELECT c.source, c.query_set_id, COALESCE(q.name, ''), COALESCE(q.provider_scope_json, '[]')
		FROM scanner_candidates c
		LEFT JOIN scanner_query_sets q ON q.id = c.query_set_id
		WHERE c.id = ?
	`, candidateID).Scan(&sourceProvider, &querySetID, &queryName, &providerScopeJSON); err != nil {
		return enriched
	}
	enriched["source_provider"] = strings.TrimSpace(sourceProvider)
	enriched["query_set_id"] = strings.TrimSpace(querySetID)
	enriched["query_name"] = strings.TrimSpace(queryName)
	enriched["provider_scope"] = decodeStringArray(providerScopeJSON)
	return enriched
}

func buildDiscoveryMetadataNote(listingURL, seller, stockSignal string, observedPrice float64, sourceProvider string, querySetID string, queryName string, providerScope []string) string {
	payload := map[string]any{
		"listing_url":     strings.TrimSpace(listingURL),
		"seller":          strings.TrimSpace(seller),
		"stock_signal":    strings.TrimSpace(stockSignal),
		"observed_price":  math.Round(observedPrice*100) / 100,
		"source_provider": strings.TrimSpace(sourceProvider),
		"query_set_id":    strings.TrimSpace(querySetID),
		"query_name":      strings.TrimSpace(queryName),
		"provider_scope":  providerScope,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "[discovery_metadata]{}"
	}
	return "[discovery_metadata]" + string(raw)
}

func decodeStringArray(raw string) []string {
	var values []string
	if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &values); err != nil {
		return []string{}
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			out = append(out, value)
		}
	}
	return out
}

func mergeDiscoveryMetadataNotes(existing, metadata string) string {
	existing = strings.TrimSpace(existing)
	metadata = strings.TrimSpace(metadata)
	if existing == "" {
		return metadata
	}
	const marker = "[discovery_metadata]"
	if idx := strings.Index(existing, marker); idx >= 0 {
		base := strings.TrimSpace(existing[:idx])
		if base == "" {
			return metadata
		}
		return base + "\n" + metadata
	}
	return existing + "\n" + metadata
}

func (s *Service) ResetIgnored(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM ignored_candidates`)
	return err
}
