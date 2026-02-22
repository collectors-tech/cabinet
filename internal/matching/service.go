package matching

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
)

type State string

const (
	StateMatched         State = "matched"
	StateSuggested       State = "suggested"
	StateNotInCollection State = "not_in_collection"
)

type Result struct {
	CandidateID string  `json:"candidate_id"`
	ItemID      string  `json:"item_id"`
	State       State   `json:"state"`
	Confidence  float64 `json:"confidence"`
	NeedsReview bool    `json:"needs_review"`
	PartNumber  string  `json:"part_number"`
	LastUpdated string  `json:"last_updated"`
}

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

var partRE = regexp.MustCompile(`(?i)\b[A-Z0-9]+(?:-[A-Z0-9]+)+\b`)

func extractPartNumber(title string) string {
	m := partRE.FindString(strings.ToUpper(title))
	return strings.TrimSpace(m)
}

func (s *Service) Run(ctx context.Context) error {
	rows, err := s.db.QueryContext(ctx, `SELECT id, title FROM scanner_candidates`)
	if err != nil {
		return fmt.Errorf("list candidates: %w", err)
	}
	defer rows.Close()
	type candidate struct {
		id, title string
	}
	var list []candidate
	for rows.Next() {
		var c candidate
		if err := rows.Scan(&c.id, &c.title); err != nil {
			return fmt.Errorf("scan candidate: %w", err)
		}
		list = append(list, c)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate candidates: %w", err)
	}

	for _, c := range list {
		part := extractPartNumber(c.title)
		state := StateNotInCollection
		confidence := 0.0
		itemID := ""
		if part != "" {
			err := s.db.QueryRowContext(ctx, `SELECT id FROM canonical_items WHERE UPPER(part_number) = ?`, strings.ToUpper(part)).Scan(&itemID)
			if err == nil {
				state = StateMatched
				confidence = 1.0
			} else if err != sql.ErrNoRows {
				return fmt.Errorf("lookup part number: %w", err)
			} else {
				state = StateSuggested
				confidence = 0.65
			}
		}
		needsReview := state != StateMatched || confidence < 0.75
		_, err := s.db.ExecContext(ctx, `
			INSERT INTO scanner_matches(candidate_id, item_id, state, confidence, needs_review, extracted_part_number, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
			ON CONFLICT(candidate_id) DO UPDATE SET
				item_id = excluded.item_id,
				state = excluded.state,
				confidence = excluded.confidence,
				needs_review = excluded.needs_review,
				extracted_part_number = excluded.extracted_part_number,
				updated_at = CURRENT_TIMESTAMP
		`, c.id, itemID, string(state), confidence, boolToInt(needsReview), part)
		if err != nil {
			return fmt.Errorf("upsert match: %w", err)
		}
	}
	return nil
}

func (s *Service) List(ctx context.Context) ([]Result, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT candidate_id, item_id, state, confidence, needs_review, extracted_part_number, updated_at
		FROM scanner_matches
		ORDER BY updated_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("list matches: %w", err)
	}
	defer rows.Close()
	var out []Result
	for rows.Next() {
		var r Result
		var st string
		var needs int
		if err := rows.Scan(&r.CandidateID, &r.ItemID, &st, &r.Confidence, &needs, &r.PartNumber, &r.LastUpdated); err != nil {
			return nil, fmt.Errorf("scan match: %w", err)
		}
		r.State = State(st)
		r.NeedsReview = needs == 1
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate matches: %w", err)
	}
	return out, nil
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
