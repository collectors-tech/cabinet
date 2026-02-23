package scanner

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type QuerySet struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Keywords      []string `json:"keywords"`
	Exclusions    []string `json:"exclusions"`
	MaxPrice      float64  `json:"max_price"`
	Region        string   `json:"region"`
	Condition     string   `json:"condition"`
	ScheduleCron  string   `json:"schedule_cron"`
	Enabled       bool     `json:"enabled"`
	RateLimitRPS  int      `json:"rate_limit_rps"`
	MaxRetryCount int      `json:"max_retry_count"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
}

type CandidateInput struct {
	ListingID string  `json:"listing_id"`
	Title     string  `json:"title"`
	Price     float64 `json:"price"`
	Shipping  float64 `json:"shipping"`
	URL       string  `json:"url"`
	Image     string  `json:"image"`
	Seller    string  `json:"seller"`
	Source    string  `json:"source"`
}

type Candidate struct {
	ID         string  `json:"id"`
	QuerySetID string  `json:"query_set_id"`
	ListingID  string  `json:"listing_id"`
	Title      string  `json:"title"`
	Price      float64 `json:"price"`
	Shipping   float64 `json:"shipping"`
	URL        string  `json:"url"`
	Image      string  `json:"image"`
	Seller     string  `json:"seller"`
	FirstSeen  string  `json:"first_seen"`
	LastSeen   string  `json:"last_seen"`
	Status     string  `json:"status"`
	Source     string  `json:"source"`
}

type RunResult struct {
	Attempts int `json:"attempts"`
	Saved    int `json:"saved"`
}

type Provider interface {
	Search(ctx context.Context, q QuerySet) ([]CandidateInput, error)
}

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) CreateQuerySet(ctx context.Context, in QuerySet) (QuerySet, error) {
	return s.CreateQuerySetForProfile(ctx, "", in)
}

func (s *Service) CreateQuerySetForProfile(ctx context.Context, profileID string, in QuerySet) (QuerySet, error) {
	if strings.TrimSpace(in.Name) == "" || len(in.Keywords) == 0 {
		return QuerySet{}, fmt.Errorf("name and keywords are required")
	}
	in.ID = uuid.NewString()
	if in.RateLimitRPS <= 0 {
		in.RateLimitRPS = 2
	}
	if in.MaxRetryCount < 0 {
		in.MaxRetryCount = 0
	}
	k, _ := json.Marshal(in.Keywords)
	e, _ := json.Marshal(in.Exclusions)
	enabled := 0
	if in.Enabled {
		enabled = 1
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO scanner_query_sets(id, profile_id, name, keywords_json, exclusions_json, max_price, region, condition_filter, schedule_cron, enabled, rate_limit_rps, max_retry_count)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, in.ID, strings.TrimSpace(profileID), in.Name, string(k), string(e), in.MaxPrice, in.Region, in.Condition, in.ScheduleCron, enabled, in.RateLimitRPS, in.MaxRetryCount)
	if err != nil {
		return QuerySet{}, fmt.Errorf("insert query set: %w", err)
	}
	return s.GetQuerySetForProfile(ctx, strings.TrimSpace(profileID), in.ID)
}

func (s *Service) GetQuerySet(ctx context.Context, id string) (QuerySet, error) {
	return s.GetQuerySetForProfile(ctx, "", id)
}

func (s *Service) GetQuerySetForProfile(ctx context.Context, profileID, id string) (QuerySet, error) {
	var q QuerySet
	var keywordsJSON, exclusionsJSON string
	var enabled int
	err := s.db.QueryRowContext(ctx, `
		SELECT id, name, keywords_json, exclusions_json, max_price, region, condition_filter, schedule_cron, enabled, rate_limit_rps, max_retry_count, created_at, updated_at
		FROM scanner_query_sets WHERE id = ? AND (? = '' OR profile_id = ?)
	`, id, strings.TrimSpace(profileID), strings.TrimSpace(profileID)).Scan(&q.ID, &q.Name, &keywordsJSON, &exclusionsJSON, &q.MaxPrice, &q.Region, &q.Condition, &q.ScheduleCron, &enabled, &q.RateLimitRPS, &q.MaxRetryCount, &q.CreatedAt, &q.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return QuerySet{}, fmt.Errorf("query set not found")
		}
		return QuerySet{}, fmt.Errorf("get query set: %w", err)
	}
	_ = json.Unmarshal([]byte(keywordsJSON), &q.Keywords)
	_ = json.Unmarshal([]byte(exclusionsJSON), &q.Exclusions)
	q.Enabled = enabled == 1
	return q, nil
}

func (s *Service) ListQuerySets(ctx context.Context) ([]QuerySet, error) {
	return s.ListQuerySetsByProfile(ctx, "")
}

func (s *Service) ListQuerySetsByProfile(ctx context.Context, profileID string) ([]QuerySet, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id FROM scanner_query_sets WHERE (? = '' OR profile_id = ?) ORDER BY created_at ASC`, strings.TrimSpace(profileID), strings.TrimSpace(profileID))
	if err != nil {
		return nil, fmt.Errorf("list query sets: %w", err)
	}
	defer rows.Close()
	var out []QuerySet
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan query set id: %w", err)
		}
		q, err := s.GetQuerySetForProfile(ctx, strings.TrimSpace(profileID), id)
		if err != nil {
			return nil, err
		}
		out = append(out, q)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate query sets: %w", err)
	}
	return out, nil
}

func (s *Service) RunNow(ctx context.Context, querySetID string, provider Provider) (RunResult, error) {
	return s.RunNowForProfile(ctx, "", querySetID, provider)
}

func (s *Service) RunNowForProfile(ctx context.Context, profileID, querySetID string, provider Provider) (RunResult, error) {
	qs, err := s.GetQuerySetForProfile(ctx, strings.TrimSpace(profileID), querySetID)
	if err != nil {
		return RunResult{}, err
	}
	if provider == nil {
		return RunResult{}, fmt.Errorf("provider is required")
	}

	maxAttempts := qs.MaxRetryCount + 1
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	var items []CandidateInput
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		items, lastErr = provider.Search(ctx, qs)
		if lastErr == nil {
			s.recordProviderHealth(ctx, "ebay", "ok", "")
			return s.persistCandidatesForProfile(ctx, strings.TrimSpace(profileID), qs.ID, items, attempt)
		}
		s.recordProviderHealth(ctx, "ebay", "error", lastErr.Error())
		s.logFailure(ctx, "ebay", lastErr.Error())
		if attempt < maxAttempts {
			sleep := time.Duration(1000/qs.RateLimitRPS) * time.Millisecond
			if sleep < 100*time.Millisecond {
				sleep = 100 * time.Millisecond
			}
			time.Sleep(sleep * time.Duration(attempt))
		}
	}
	return RunResult{Attempts: maxAttempts}, fmt.Errorf("run failed: %w", lastErr)
}

func (s *Service) RunScheduled(ctx context.Context, provider Provider) (int, error) {
	return s.RunScheduledForProfile(ctx, "", provider)
}

func (s *Service) RunScheduledForProfile(ctx context.Context, profileID string, provider Provider) (int, error) {
	sets, err := s.ListQuerySetsByProfile(ctx, strings.TrimSpace(profileID))
	if err != nil {
		return 0, err
	}
	ran := 0
	for _, qs := range sets {
		if !qs.Enabled || strings.TrimSpace(qs.ScheduleCron) == "" {
			continue
		}
		if _, err := s.RunNowForProfile(ctx, strings.TrimSpace(profileID), qs.ID, provider); err != nil {
			return ran, err
		}
		ran++
	}
	return ran, nil
}

func (s *Service) persistCandidates(ctx context.Context, querySetID string, items []CandidateInput, attempts int) (RunResult, error) {
	return s.persistCandidatesForProfile(ctx, "", querySetID, items, attempts)
}

func (s *Service) persistCandidatesForProfile(ctx context.Context, profileID, querySetID string, items []CandidateInput, attempts int) (RunResult, error) {
	saved := 0
	for _, it := range items {
		if strings.TrimSpace(it.ListingID) == "" {
			continue
		}
		source := strings.TrimSpace(it.Source)
		if source == "" {
			source = "ebay"
		}
		res, err := s.db.ExecContext(ctx, `
			UPDATE scanner_candidates
			SET title = ?, price = ?, shipping = ?, url = ?, image = ?, seller = ?, last_seen = CURRENT_TIMESTAMP, status = 'seen', source = ?
			WHERE listing_id = ? AND (? = '' OR profile_id = ?)
		`, it.Title, it.Price, it.Shipping, it.URL, it.Image, it.Seller, source, it.ListingID, strings.TrimSpace(profileID), strings.TrimSpace(profileID))
		if err != nil {
			return RunResult{}, fmt.Errorf("update candidate: %w", err)
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			_, err := s.db.ExecContext(ctx, `
				INSERT INTO scanner_candidates(id, profile_id, query_set_id, listing_id, title, price, shipping, url, image, seller, first_seen, last_seen, status, source)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'new', ?)
			`, uuid.NewString(), strings.TrimSpace(profileID), querySetID, it.ListingID, it.Title, it.Price, it.Shipping, it.URL, it.Image, it.Seller, source)
			if err != nil {
				return RunResult{}, fmt.Errorf("insert candidate: %w", err)
			}
		}
		saved++
	}
	return RunResult{Attempts: attempts, Saved: saved}, nil
}

func (s *Service) ListCandidates(ctx context.Context, querySetID string) ([]Candidate, error) {
	return s.ListCandidatesByProfile(ctx, "", querySetID)
}

func (s *Service) ListCandidatesByProfile(ctx context.Context, profileID, querySetID string) ([]Candidate, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, query_set_id, listing_id, title, price, shipping, url, image, seller, first_seen, last_seen, status, source
		FROM scanner_candidates
		WHERE query_set_id = ? AND (? = '' OR profile_id = ?)
		ORDER BY last_seen DESC
	`, querySetID, strings.TrimSpace(profileID), strings.TrimSpace(profileID))
	if err != nil {
		return nil, fmt.Errorf("list candidates: %w", err)
	}
	defer rows.Close()
	var out []Candidate
	for rows.Next() {
		var c Candidate
		if err := rows.Scan(&c.ID, &c.QuerySetID, &c.ListingID, &c.Title, &c.Price, &c.Shipping, &c.URL, &c.Image, &c.Seller, &c.FirstSeen, &c.LastSeen, &c.Status, &c.Source); err != nil {
			return nil, fmt.Errorf("scan candidate: %w", err)
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate candidates: %w", err)
	}
	return out, nil
}

func (s *Service) ListFailures(ctx context.Context) ([]map[string]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT provider, message, created_at FROM scanner_failures ORDER BY created_at DESC LIMIT 100`)
	if err != nil {
		return nil, fmt.Errorf("list failures: %w", err)
	}
	defer rows.Close()
	var out []map[string]string
	for rows.Next() {
		var p, m, c string
		if err := rows.Scan(&p, &m, &c); err != nil {
			return nil, fmt.Errorf("scan failure: %w", err)
		}
		out = append(out, map[string]string{"provider": p, "message": m, "created_at": c})
	}
	return out, rows.Err()
}

func (s *Service) ProviderHealth(ctx context.Context, provider string) (map[string]string, error) {
	var status, msg, updated string
	err := s.db.QueryRowContext(ctx, `
		SELECT status, message, updated_at
		FROM provider_health WHERE provider = ?
	`, provider).Scan(&status, &msg, &updated)
	if err != nil {
		if err == sql.ErrNoRows {
			return map[string]string{"provider": provider, "status": "unknown", "message": ""}, nil
		}
		return nil, fmt.Errorf("provider health: %w", err)
	}
	return map[string]string{"provider": provider, "status": status, "message": msg, "updated_at": updated}, nil
}

func (s *Service) recordProviderHealth(ctx context.Context, provider, status, message string) {
	_, _ = s.db.ExecContext(ctx, `
		INSERT INTO provider_health(provider, status, message, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(provider) DO UPDATE SET
			status = excluded.status,
			message = excluded.message,
			updated_at = CURRENT_TIMESTAMP
	`, provider, status, message)
}

func (s *Service) logFailure(ctx context.Context, provider, message string) {
	_, _ = s.db.ExecContext(ctx, `
		INSERT INTO scanner_failures(id, provider, message, created_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	`, uuid.NewString(), provider, message)
}
