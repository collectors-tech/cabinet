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
	CandidateID      string  `json:"candidate_id"`
	SourceProvider   string  `json:"source_provider"`
	QuerySetID       string  `json:"query_set_id"`
	QueryName        string  `json:"query_name"`
	ListingID        string  `json:"listing_id"`
	Title            string  `json:"title"`
	Price            float64 `json:"price"`
	ObservedCurrency string  `json:"observed_currency"`
	URL              string  `json:"url"`
	Seller           string  `json:"seller"`
	FirstSeen        string  `json:"first_seen"`
	LastSeen         string  `json:"last_seen"`
	Status           string  `json:"status"`
	Confidence       float64 `json:"confidence"`
	NeedsReview      bool    `json:"needs_review"`
	ReviewerNotes    string  `json:"reviewer_notes"`
	SourceResultURL  string  `json:"source_result_url"`
	ExtractedPart    string  `json:"extracted_part_number"`
	StockState       string  `json:"stock_state"`
	StockCount       int     `json:"stock_count"`
}

type ActionType string

const (
	ActionIgnore      ActionType = "ignore"
	ActionAddWishlist ActionType = "add_to_wishlist"
	ActionTrackPrice  ActionType = "track_price"
	ActionCreateItem  ActionType = "create_item"
	ActionReview      ActionType = "review"
	ActionArchive     ActionType = "archive"
)

type Action struct {
	CandidateID string         `json:"candidate_id"`
	Type        ActionType     `json:"type"`
	Payload     map[string]any `json:"payload"`
}

type ActionResult struct {
	OK          bool           `json:"ok"`
	Action      ActionType     `json:"action"`
	CandidateID string         `json:"candidate_id"`
	Audit       map[string]any `json:"audit"`
}

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) ListNotInCollection(ctx context.Context, f Filter) ([]Item, error) {
	q := `
		SELECT c.id, c.source, c.query_set_id, COALESCE(q.name, ''), c.listing_id,
			c.title, c.price, c.observed_currency, c.url, c.seller, c.first_seen, c.last_seen,
			c.status, m.confidence, m.needs_review, c.reviewer_notes,
			COALESCE(NULLIF(c.source_result_url, ''), c.url), m.extracted_part_number,
			c.stock_state, c.stock_count
		FROM scanner_candidates c
		JOIN scanner_matches m ON m.candidate_id = c.id
		LEFT JOIN scanner_query_sets q ON q.id = c.query_set_id
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
		var needsReview int
		if err := rows.Scan(
			&it.CandidateID, &it.SourceProvider, &it.QuerySetID, &it.QueryName, &it.ListingID,
			&it.Title, &it.Price, &it.ObservedCurrency, &it.URL, &it.Seller, &it.FirstSeen, &it.LastSeen,
			&it.Status, &it.Confidence, &needsReview, &it.ReviewerNotes, &it.SourceResultURL,
			&it.ExtractedPart, &it.StockState, &it.StockCount,
		); err != nil {
			return nil, fmt.Errorf("scan not_in_collection row: %w", err)
		}
		it.NeedsReview = needsReview != 0
		out = append(out, it)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate not_in_collection rows: %w", err)
	}
	return out, nil
}

func (s *Service) ApplyAction(ctx context.Context, a Action) error {
	_, err := s.ApplyActionWithResult(ctx, a)
	return err
}

func (s *Service) ApplyActionWithResult(ctx context.Context, a Action) (ActionResult, error) {
	if strings.TrimSpace(a.CandidateID) == "" {
		return ActionResult{}, fmt.Errorf("candidate_id is required")
	}
	if a.Payload == nil {
		a.Payload = map[string]any{}
	}
	a.Payload = s.enrichDiscoveryActionPayload(ctx, a.CandidateID, a.Payload)
	result := ActionResult{
		OK:          true,
		Action:      a.Type,
		CandidateID: strings.TrimSpace(a.CandidateID),
		Audit:       a.Payload,
	}
	reviewerNotes := payloadString(a.Payload, "reviewer_notes")
	if reviewerNotes == "" {
		reviewerNotes = payloadString(a.Payload, "notes")
	}
	raw, _ := json.Marshal(a.Payload)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO discovery_actions(id, candidate_id, action_type, payload_json)
		VALUES (?, ?, ?, ?)
	`, uuid.NewString(), a.CandidateID, string(a.Type), string(raw))
	if err != nil {
		return ActionResult{}, fmt.Errorf("insert discovery action: %w", err)
	}

	switch a.Type {
	case ActionReview:
		if err := s.updateCandidateTriage(ctx, a.CandidateID, "reviewing", reviewerNotes); err != nil {
			return ActionResult{}, err
		}
	case ActionIgnore:
		if err := s.updateCandidateTriage(ctx, a.CandidateID, "ignored", reviewerNotes); err != nil {
			return ActionResult{}, err
		}
		_, err = s.db.ExecContext(ctx, `
			INSERT INTO ignored_candidates(candidate_id, ignored_at)
			VALUES (?, CURRENT_TIMESTAMP)
			ON CONFLICT(candidate_id) DO UPDATE SET ignored_at = CURRENT_TIMESTAMP
		`, a.CandidateID)
		if err != nil {
			return ActionResult{}, fmt.Errorf("ignore candidate: %w", err)
		}
	case ActionArchive:
		if err := s.updateCandidateTriage(ctx, a.CandidateID, "archived", reviewerNotes); err != nil {
			return ActionResult{}, err
		}
		_, err = s.db.ExecContext(ctx, `
			INSERT INTO ignored_candidates(candidate_id, ignored_at)
			VALUES (?, CURRENT_TIMESTAMP)
			ON CONFLICT(candidate_id) DO UPDATE SET ignored_at = CURRENT_TIMESTAMP
		`, a.CandidateID)
		if err != nil {
			return ActionResult{}, fmt.Errorf("archive candidate: %w", err)
		}
	case ActionTrackPrice:
		if err := s.updateCandidateTriage(ctx, a.CandidateID, "purchase_candidate", reviewerNotes); err != nil {
			return ActionResult{}, err
		}
		var itemID string
		if err := s.db.QueryRowContext(ctx, `SELECT item_id FROM scanner_matches WHERE candidate_id = ?`, a.CandidateID).Scan(&itemID); err == nil && strings.TrimSpace(itemID) != "" {
			_, _ = s.db.ExecContext(ctx, `INSERT INTO tracked_items(item_id, created_at) VALUES (?, CURRENT_TIMESTAMP) ON CONFLICT(item_id) DO NOTHING`, itemID)
		}
	case ActionAddWishlist:
		if err := s.updateCandidateTriage(ctx, a.CandidateID, "wishlisted", reviewerNotes); err != nil {
			return ActionResult{}, err
		}
		var itemID string
		if err := s.db.QueryRowContext(ctx, `SELECT item_id FROM scanner_matches WHERE candidate_id = ?`, a.CandidateID).Scan(&itemID); err == nil && strings.TrimSpace(itemID) != "" {
			var profileID, listingURL, seller, stockSignal, sourceProvider, querySetID, queryName, providerScopeJSON, observedCurrency string
			var observedPrice float64
			scanErr := s.db.QueryRowContext(ctx, `
				SELECT c.profile_id, COALESCE(NULLIF(c.source_result_url, ''), c.url), c.seller, c.stock_state, c.price, c.observed_currency, c.source, c.query_set_id, COALESCE(q.name, ''), COALESCE(q.provider_scope_json, '[]')
				FROM scanner_candidates c
				LEFT JOIN scanner_query_sets q ON q.id = c.query_set_id
				WHERE c.id = ?
			`, a.CandidateID).Scan(&profileID, &listingURL, &seller, &stockSignal, &observedPrice, &observedCurrency, &sourceProvider, &querySetID, &queryName, &providerScopeJSON)
			if scanErr != nil {
				_ = s.db.QueryRowContext(ctx, `
					SELECT COALESCE(NULLIF(source_result_url, ''), url), seller, stock_state, price, source, query_set_id
					FROM scanner_candidates
					WHERE id = ?
				`, a.CandidateID).Scan(&listingURL, &seller, &stockSignal, &observedPrice, &sourceProvider, &querySetID)
			}
			profileID = strings.TrimSpace(profileID)
			if profileID == "" {
				_ = s.db.QueryRowContext(ctx, `SELECT profile_id FROM canonical_items WHERE id = ?`, itemID).Scan(&profileID)
				profileID = strings.TrimSpace(profileID)
			}
			metadata := buildDiscoveryMetadataNote(listingURL, seller, stockSignal, observedPrice, observedCurrency, sourceProvider, querySetID, queryName, decodeStringArray(providerScopeJSON))
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
		if err := s.updateCandidateTriage(ctx, a.CandidateID, "inventory_candidate", reviewerNotes); err != nil {
			return ActionResult{}, err
		}
		if !payloadConfirmsOwnedOrPurchased(a.Payload) {
			break
		}
		var title, partNumber, profileID, listingURL, seller, stockSignal, sourceProvider, querySetID, queryName, providerScopeJSON, observedCurrency string
		var observedPrice float64
		if err := s.db.QueryRowContext(ctx, `
			SELECT c.title, m.extracted_part_number, c.profile_id, COALESCE(NULLIF(c.source_result_url, ''), c.url), c.seller, c.stock_state, c.price, c.observed_currency, c.source, c.query_set_id, COALESCE(q.name, ''), COALESCE(q.provider_scope_json, '[]')
			FROM scanner_candidates c
			JOIN scanner_matches m ON m.candidate_id = c.id
			LEFT JOIN scanner_query_sets q ON q.id = c.query_set_id
			WHERE c.id = ?
		`, a.CandidateID).Scan(&title, &partNumber, &profileID, &listingURL, &seller, &stockSignal, &observedPrice, &observedCurrency, &sourceProvider, &querySetID, &queryName, &providerScopeJSON); err == nil {
			if strings.TrimSpace(partNumber) == "" {
				partNumber = "AUTO-" + strings.ToUpper(uuid.NewString()[:8])
			}
			metadata := buildDiscoveryMetadataNote(listingURL, seller, stockSignal, observedPrice, observedCurrency, sourceProvider, querySetID, queryName, decodeStringArray(providerScopeJSON))
			sourceURLs, _ := json.Marshal([]string{strings.TrimSpace(listingURL)})
			_, _ = s.db.ExecContext(ctx, `
				INSERT INTO canonical_items(id, profile_id, brand, category, part_number, title, status, notes, source_urls_json, created_by, updated_by)
				VALUES (?, ?, ?, ?, ?, ?, 'active', ?, ?, 'discovery.service', 'discovery.service')
			`, uuid.NewString(), strings.TrimSpace(profileID), "Unknown", "Unknown", partNumber, title, metadata, string(sourceURLs))
		}
	}
	return result, nil
}

func (s *Service) enrichDiscoveryActionPayload(ctx context.Context, candidateID string, payload map[string]any) map[string]any {
	enriched := make(map[string]any, len(payload)+4)
	for key, value := range payload {
		enriched[key] = value
	}
	var sourceProvider, querySetID, queryName, providerScopeJSON string
	var listingID, listingURL, title, observedCurrency, seller, firstSeen, lastSeen, status, reviewerNotes, sourceResultURL string
	var observedPrice, confidence float64
	var needsReview int
	if err := s.db.QueryRowContext(ctx, `
		SELECT c.source, c.query_set_id, COALESCE(q.name, ''), COALESCE(q.provider_scope_json, '[]'),
			c.listing_id, c.url, c.title, c.price, c.observed_currency, c.seller,
			c.first_seen, c.last_seen, c.status, COALESCE(m.confidence, 0), COALESCE(m.needs_review, 1),
			c.reviewer_notes, COALESCE(NULLIF(c.source_result_url, ''), c.url)
		FROM scanner_candidates c
		LEFT JOIN scanner_query_sets q ON q.id = c.query_set_id
		LEFT JOIN scanner_matches m ON m.candidate_id = c.id
		WHERE c.id = ?
	`, candidateID).Scan(
		&sourceProvider, &querySetID, &queryName, &providerScopeJSON,
		&listingID, &listingURL, &title, &observedPrice, &observedCurrency, &seller,
		&firstSeen, &lastSeen, &status, &confidence, &needsReview,
		&reviewerNotes, &sourceResultURL,
	); err != nil {
		return enriched
	}
	enriched["source_provider"] = strings.TrimSpace(sourceProvider)
	enriched["query_set_id"] = strings.TrimSpace(querySetID)
	enriched["query_name"] = strings.TrimSpace(queryName)
	enriched["provider_scope"] = decodeStringArray(providerScopeJSON)
	enriched["listing_id"] = strings.TrimSpace(listingID)
	enriched["listing_url"] = strings.TrimSpace(listingURL)
	enriched["title"] = strings.TrimSpace(title)
	enriched["observed_price"] = math.Round(observedPrice*100) / 100
	enriched["observed_currency"] = strings.TrimSpace(observedCurrency)
	enriched["seller"] = strings.TrimSpace(seller)
	enriched["first_seen"] = strings.TrimSpace(firstSeen)
	enriched["last_seen"] = strings.TrimSpace(lastSeen)
	enriched["triage_status"] = strings.TrimSpace(status)
	enriched["confidence"] = confidence
	enriched["needs_review"] = needsReview != 0
	enriched["reviewer_notes"] = strings.TrimSpace(reviewerNotes)
	enriched["source_result_url"] = strings.TrimSpace(sourceResultURL)
	return enriched
}

func buildDiscoveryMetadataNote(listingURL, seller, stockSignal string, observedPrice float64, observedCurrency string, sourceProvider string, querySetID string, queryName string, providerScope []string) string {
	payload := map[string]any{
		"listing_url":       strings.TrimSpace(listingURL),
		"seller":            strings.TrimSpace(seller),
		"stock_signal":      strings.TrimSpace(stockSignal),
		"observed_price":    math.Round(observedPrice*100) / 100,
		"observed_currency": strings.TrimSpace(observedCurrency),
		"source_provider":   strings.TrimSpace(sourceProvider),
		"query_set_id":      strings.TrimSpace(querySetID),
		"query_name":        strings.TrimSpace(queryName),
		"provider_scope":    providerScope,
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

func (s *Service) updateCandidateTriage(ctx context.Context, candidateID, status, reviewerNotes string) error {
	if !validTriageStatus(status) {
		return fmt.Errorf("invalid discovery triage status %q", status)
	}
	_, err := s.db.ExecContext(ctx, `
		UPDATE scanner_candidates
		SET status = ?,
			reviewer_notes = CASE WHEN ? = '' THEN reviewer_notes ELSE ? END,
			source_result_url = CASE WHEN source_result_url = '' THEN url ELSE source_result_url END
		WHERE id = ?
	`, status, strings.TrimSpace(reviewerNotes), strings.TrimSpace(reviewerNotes), candidateID)
	if err != nil {
		return fmt.Errorf("update discovery triage status: %w", err)
	}
	return nil
}

func validTriageStatus(status string) bool {
	switch status {
	case "new", "reviewing", "wishlisted", "inventory_candidate", "purchase_candidate", "ignored", "archived":
		return true
	default:
		return false
	}
}

func payloadString(payload map[string]any, key string) string {
	if payload == nil {
		return ""
	}
	value, ok := payload[key]
	if !ok {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func payloadConfirmsOwnedOrPurchased(payload map[string]any) bool {
	for _, key := range []string{"ownership_confirmed", "owned_confirmed", "purchase_confirmed"} {
		if value, ok := payload[key].(bool); ok && value {
			return true
		}
	}
	decision := strings.ToLower(payloadString(payload, "decision") + " " + payloadString(payload, "confirmation"))
	return strings.Contains(decision, "owned") || strings.Contains(decision, "purchased") || strings.Contains(decision, "confirmed")
}

func (s *Service) ResetIgnored(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM ignored_candidates`)
	return err
}
