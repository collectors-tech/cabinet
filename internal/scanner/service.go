package scanner

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type QuerySet struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	Keywords           []string `json:"keywords"`
	Exclusions         []string `json:"exclusions"`
	ProviderScope      []string `json:"provider_scope"`
	ItemsPerPage       int      `json:"items_per_page"`
	MaxPrice           float64  `json:"max_price"`
	Region             string   `json:"region"`
	Condition          string   `json:"condition"`
	ScheduleCron       string   `json:"schedule_cron"`
	Enabled            bool     `json:"enabled"`
	RateLimitRPS       int      `json:"rate_limit_rps"`
	MaxRetryCount      int      `json:"max_retry_count"`
	CreatedAt          string   `json:"created_at"`
	UpdatedAt          string   `json:"updated_at"`
	LastRunStatus      string   `json:"last_run_status,omitempty"`
	LastRunAt          string   `json:"last_run_at,omitempty"`
	LastRunMessage     string   `json:"last_run_message,omitempty"`
	NextRunAt          string   `json:"next_run_at,omitempty"`
	LastCandidateCount int      `json:"last_candidate_count,omitempty"`
}

type RunHistoryRecord struct {
	ID             string `json:"id"`
	ProfileID      string `json:"profile_id,omitempty"`
	QuerySetID     string `json:"query_set_id"`
	Provider       string `json:"provider"`
	TriggerType    string `json:"trigger_type"`
	StartedAt      string `json:"started_at"`
	FinishedAt     string `json:"finished_at"`
	Status         string `json:"status"`
	ResultCount    int    `json:"result_count"`
	NewResultCount int    `json:"new_result_count"`
	ErrorCategory  string `json:"error_category,omitempty"`
	ErrorMessage   string `json:"error_message,omitempty"`
	RetryGuidance  string `json:"retry_guidance,omitempty"`
	NextAction     string `json:"next_action,omitempty"`
}

type retryAfterProviderError interface {
	RetryAfter() int
}

type CandidateInput struct {
	ListingID  string  `json:"listing_id"`
	Title      string  `json:"title"`
	Price      float64 `json:"price"`
	Currency   string  `json:"currency,omitempty"`
	Shipping   float64 `json:"shipping"`
	URL        string  `json:"url"`
	Image      string  `json:"image"`
	Seller     string  `json:"seller"`
	Source     string  `json:"source"`
	StockState string  `json:"stock_state"`
	StockCount int     `json:"stock_count"`
}

type Candidate struct {
	ID         string  `json:"id"`
	QuerySetID string  `json:"query_set_id"`
	ListingID  string  `json:"listing_id"`
	Title      string  `json:"title"`
	Price      float64 `json:"price"`
	Currency   string  `json:"observed_currency"`
	Shipping   float64 `json:"shipping"`
	URL        string  `json:"url"`
	Image      string  `json:"image"`
	Seller     string  `json:"seller"`
	FirstSeen  string  `json:"first_seen"`
	LastSeen   string  `json:"last_seen"`
	Status     string  `json:"status"`
	Source     string  `json:"source"`
	StockState string  `json:"stock_state"`
	StockCount int     `json:"stock_count"`
}

type CandidateListFilter struct {
	Status   string
	Provider string
	Page     int
	PageSize int
}

type CandidateList struct {
	Candidates []Candidate `json:"candidates"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
}

type RecognitionCandidateInput struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Confidence   float64  `json:"confidence"`
	Source       string   `json:"source"`
	Provenance   string   `json:"provenance"`
	Alternates   []string `json:"alternates,omitempty"`
	MediaID      string   `json:"media_id,omitempty"`
	MediaURL     string   `json:"media_url,omitempty"`
	Target       string   `json:"target,omitempty"`
	OverrideID   string   `json:"override_id,omitempty"`
	OverrideBy   string   `json:"override_by,omitempty"`
	OverrideNote string   `json:"override_note,omitempty"`
}

type RecognitionReview struct {
	TopCandidate          RecognitionCandidateInput   `json:"top_candidate"`
	Alternates            []RecognitionCandidateInput `json:"alternates"`
	SelectedCandidate     RecognitionCandidateInput   `json:"selected_candidate"`
	ConfidenceLabel       string                      `json:"confidence_label"`
	RequiresManualReview  bool                        `json:"requires_manual_review"`
	ConfirmBeforeCreate   bool                        `json:"confirm_before_create"`
	Target                string                      `json:"target"`
	MediaEvidence         map[string]string           `json:"media_evidence"`
	Provenance            []string                    `json:"provenance"`
	ManualOverrideApplied bool                        `json:"manual_override_applied"`
}

type RunResult struct {
	RunID                 string `json:"run_id,omitempty"`
	Attempts              int    `json:"attempts"`
	Saved                 int    `json:"saved"`
	NewResults            int    `json:"new_results"`
	ItemsPerPageRequested int    `json:"items_per_page_requested"`
	ItemsPerPageEffective int    `json:"items_per_page_effective"`
	ObservedPageSize      int    `json:"observed_page_size"`
	PageCount             int    `json:"page_count"`
	ItemsPerPageWarning   string `json:"items_per_page_warning,omitempty"`
}

type Provider interface {
	Search(ctx context.Context, q QuerySet) ([]CandidateInput, error)
}

type providerIdentifier interface {
	ProviderID() string
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
	if len(in.ProviderScope) == 0 {
		in.ProviderScope = defaultProviderScope(in.Region)
	}
	in.ProviderScope = normalizeProviderScope(in.ProviderScope)
	if in.ItemsPerPage <= 0 {
		in.ItemsPerPage = 24
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
	p, _ := json.Marshal(in.ProviderScope)
	enabled := 0
	if in.Enabled {
		enabled = 1
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO scanner_query_sets(id, profile_id, name, keywords_json, exclusions_json, provider_scope_json, items_per_page, max_price, region, condition_filter, schedule_cron, enabled, rate_limit_rps, max_retry_count)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, in.ID, strings.TrimSpace(profileID), in.Name, string(k), string(e), string(p), in.ItemsPerPage, in.MaxPrice, in.Region, in.Condition, in.ScheduleCron, enabled, in.RateLimitRPS, in.MaxRetryCount)
	if err != nil {
		return QuerySet{}, fmt.Errorf("insert query set: %w", err)
	}
	return s.GetQuerySetForProfile(ctx, strings.TrimSpace(profileID), in.ID)
}

func (s *Service) GetQuerySet(ctx context.Context, id string) (QuerySet, error) {
	return s.GetQuerySetForProfile(ctx, "", id)
}

func (s *Service) UpdateQuerySet(ctx context.Context, id string, in QuerySet) (QuerySet, error) {
	return s.UpdateQuerySetForProfile(ctx, "", id, in)
}

func (s *Service) UpdateQuerySetForProfile(ctx context.Context, profileID, id string, in QuerySet) (QuerySet, error) {
	if strings.TrimSpace(id) == "" {
		return QuerySet{}, fmt.Errorf("query set id is required")
	}
	if strings.TrimSpace(in.Name) == "" || len(in.Keywords) == 0 {
		return QuerySet{}, fmt.Errorf("name and keywords are required")
	}
	if len(in.ProviderScope) == 0 {
		existing, err := s.GetQuerySetForProfile(ctx, strings.TrimSpace(profileID), strings.TrimSpace(id))
		if err != nil {
			return QuerySet{}, err
		}
		in.ProviderScope = existing.ProviderScope
	}
	if len(in.ProviderScope) == 0 {
		in.ProviderScope = defaultProviderScope(in.Region)
	}
	in.ProviderScope = normalizeProviderScope(in.ProviderScope)
	if in.ItemsPerPage <= 0 {
		in.ItemsPerPage = 24
	}
	if in.RateLimitRPS <= 0 {
		in.RateLimitRPS = 2
	}
	if in.MaxRetryCount < 0 {
		in.MaxRetryCount = 0
	}
	k, _ := json.Marshal(in.Keywords)
	e, _ := json.Marshal(in.Exclusions)
	p, _ := json.Marshal(in.ProviderScope)
	enabled := 0
	if in.Enabled {
		enabled = 1
	}
	res, err := s.db.ExecContext(ctx, `
		UPDATE scanner_query_sets
		SET name = ?, keywords_json = ?, exclusions_json = ?, provider_scope_json = ?, items_per_page = ?, max_price = ?, region = ?, condition_filter = ?, schedule_cron = ?, enabled = ?, rate_limit_rps = ?, max_retry_count = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND (? = '' OR profile_id = ?)
	`, in.Name, string(k), string(e), string(p), in.ItemsPerPage, in.MaxPrice, in.Region, in.Condition, in.ScheduleCron, enabled, in.RateLimitRPS, in.MaxRetryCount, strings.TrimSpace(id), strings.TrimSpace(profileID), strings.TrimSpace(profileID))
	if err != nil {
		return QuerySet{}, fmt.Errorf("update query set: %w", err)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return QuerySet{}, fmt.Errorf("query set not found")
	}
	return s.GetQuerySetForProfile(ctx, strings.TrimSpace(profileID), strings.TrimSpace(id))
}

func (s *Service) DeleteQuerySet(ctx context.Context, id string) error {
	return s.DeleteQuerySetForProfile(ctx, "", id)
}

func (s *Service) DeleteQuerySetForProfile(ctx context.Context, profileID, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("query set id is required")
	}
	res, err := s.db.ExecContext(ctx, `
		DELETE FROM scanner_query_sets
		WHERE id = ? AND (? = '' OR profile_id = ?)
	`, strings.TrimSpace(id), strings.TrimSpace(profileID), strings.TrimSpace(profileID))
	if err != nil {
		return fmt.Errorf("delete query set: %w", err)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("query set not found")
	}
	return nil
}

func (s *Service) GetQuerySetForProfile(ctx context.Context, profileID, id string) (QuerySet, error) {
	var q QuerySet
	var keywordsJSON, exclusionsJSON, providerScopeJSON string
	var enabled int
	err := s.db.QueryRowContext(ctx, `
		SELECT id, name, keywords_json, exclusions_json, provider_scope_json, items_per_page, max_price, region, condition_filter, schedule_cron, enabled, rate_limit_rps, max_retry_count, created_at, updated_at
		FROM scanner_query_sets WHERE id = ? AND (? = '' OR profile_id = ?)
	`, id, strings.TrimSpace(profileID), strings.TrimSpace(profileID)).Scan(&q.ID, &q.Name, &keywordsJSON, &exclusionsJSON, &providerScopeJSON, &q.ItemsPerPage, &q.MaxPrice, &q.Region, &q.Condition, &q.ScheduleCron, &enabled, &q.RateLimitRPS, &q.MaxRetryCount, &q.CreatedAt, &q.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return QuerySet{}, fmt.Errorf("query set not found")
		}
		return QuerySet{}, fmt.Errorf("get query set: %w", err)
	}
	_ = json.Unmarshal([]byte(keywordsJSON), &q.Keywords)
	_ = json.Unmarshal([]byte(exclusionsJSON), &q.Exclusions)
	_ = json.Unmarshal([]byte(providerScopeJSON), &q.ProviderScope)
	if len(q.ProviderScope) == 0 {
		q.ProviderScope = defaultProviderScope(q.Region)
	}
	q.ProviderScope = normalizeProviderScope(q.ProviderScope)
	q.Enabled = enabled == 1
	if err := s.populateQuerySetRunSnapshot(ctx, strings.TrimSpace(profileID), &q); err != nil {
		return QuerySet{}, err
	}
	return q, nil
}

func (s *Service) populateQuerySetRunSnapshot(ctx context.Context, profileID string, q *QuerySet) error {
	if q == nil || strings.TrimSpace(q.ID) == "" {
		return nil
	}
	var candidateCount int
	if err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM scanner_candidates
		WHERE query_set_id = ? AND (? = '' OR profile_id = ?)
	`, q.ID, strings.TrimSpace(profileID), strings.TrimSpace(profileID)).Scan(&candidateCount); err != nil {
		return fmt.Errorf("query set run candidate snapshot: %w", err)
	}
	q.LastCandidateCount = candidateCount
	q.LastRunStatus = "never"

	var runStatus, finishedAt, errorMessage sql.NullString
	err := s.db.QueryRowContext(ctx, `
		SELECT status, finished_at, error_message
		FROM scanner_runs
		WHERE query_set_id = ? AND (? = '' OR profile_id = ?)
		ORDER BY finished_at DESC, started_at DESC
		LIMIT 1
	`, q.ID, strings.TrimSpace(profileID), strings.TrimSpace(profileID)).Scan(&runStatus, &finishedAt, &errorMessage)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("query set durable run snapshot: %w", err)
	}
	if err == nil {
		q.LastRunStatus = normalizeRunStatus(runStatus.String)
		q.LastRunAt = strings.TrimSpace(finishedAt.String)
		q.LastRunMessage = strings.TrimSpace(errorMessage.String)
		q.NextRunAt = computeNextRunAt(q.ScheduleCron, q.LastRunAt, time.Now().UTC())
		return nil
	}

	var latestCandidateAt sql.NullString
	if err := s.db.QueryRowContext(ctx, `
		SELECT MAX(last_seen)
		FROM scanner_candidates
		WHERE query_set_id = ? AND (? = '' OR profile_id = ?)
	`, q.ID, strings.TrimSpace(profileID), strings.TrimSpace(profileID)).Scan(&latestCandidateAt); err != nil {
		return fmt.Errorf("query set legacy candidate snapshot: %w", err)
	}
	var failureMessage, failureAt string
	err = s.db.QueryRowContext(ctx, `
		SELECT message, created_at
		FROM scanner_failures
		WHERE query_set_id = ? AND (? = '' OR profile_id = ?)
		ORDER BY created_at DESC
		LIMIT 1
	`, q.ID, strings.TrimSpace(profileID), strings.TrimSpace(profileID)).Scan(&failureMessage, &failureAt)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("query set run failure snapshot: %w", err)
	}

	if latestCandidateAt.Valid && strings.TrimSpace(latestCandidateAt.String) != "" {
		q.LastRunStatus = "succeeded"
		q.LastRunAt = latestCandidateAt.String
	}
	if failureAt != "" && (q.LastRunAt == "" || failureAt >= q.LastRunAt) {
		q.LastRunStatus = "failed"
		q.LastRunAt = failureAt
		q.LastRunMessage = failureMessage
	}
	q.NextRunAt = computeNextRunAt(q.ScheduleCron, q.LastRunAt, time.Now().UTC())
	return nil
}

func normalizeRunStatus(status string) string {
	switch strings.TrimSpace(strings.ToLower(status)) {
	case "succeeded":
		return "succeeded"
	case "failed":
		return "failed"
	default:
		return "never"
	}
}

func computeNextRunAt(scheduleCron, lastRunAt string, now time.Time) string {
	schedule, ok := parseSimpleCron(scheduleCron)
	if !ok {
		return ""
	}
	base := now.UTC()
	if parsed, ok := parseScannerTime(lastRunAt); ok && parsed.After(base) {
		base = parsed
	}
	next, ok := nextSimpleCronTime(schedule, base)
	if !ok {
		return ""
	}
	return next.Format(time.RFC3339)
}

type simpleCronSchedule struct {
	minutes []int
	hours   []int
}

func parseSimpleCron(expr string) (simpleCronSchedule, bool) {
	fields := strings.Fields(strings.TrimSpace(expr))
	if len(fields) != 5 {
		return simpleCronSchedule{}, false
	}
	if fields[2] != "*" || fields[3] != "*" || fields[4] != "*" {
		return simpleCronSchedule{}, false
	}
	minutes, ok := parseCronField(fields[0], 0, 59)
	if !ok {
		return simpleCronSchedule{}, false
	}
	hours, ok := parseCronField(fields[1], 0, 23)
	if !ok {
		return simpleCronSchedule{}, false
	}
	return simpleCronSchedule{minutes: minutes, hours: hours}, true
}

func parseCronField(field string, minValue, maxValue int) ([]int, bool) {
	field = strings.TrimSpace(field)
	if field == "*" {
		return cronRange(minValue, maxValue, 1), true
	}
	if strings.HasPrefix(field, "*/") {
		step, ok := parseCronInt(strings.TrimPrefix(field, "*/"))
		if !ok || step <= 0 {
			return nil, false
		}
		return cronRange(minValue, maxValue, step), true
	}
	value, ok := parseCronInt(field)
	if !ok || value < minValue || value > maxValue {
		return nil, false
	}
	return []int{value}, true
}

func parseCronInt(raw string) (int, bool) {
	value := 0
	for _, r := range raw {
		if r < '0' || r > '9' {
			return 0, false
		}
		value = value*10 + int(r-'0')
	}
	return value, strings.TrimSpace(raw) != ""
}

func cronRange(minValue, maxValue, step int) []int {
	out := make([]int, 0, maxValue-minValue+1)
	for value := minValue; value <= maxValue; value += step {
		out = append(out, value)
	}
	return out
}

func nextSimpleCronTime(schedule simpleCronSchedule, base time.Time) (time.Time, bool) {
	start := base.UTC().Truncate(time.Minute).Add(time.Minute)
	allowedMinutes := make(map[int]struct{}, len(schedule.minutes))
	for _, minute := range schedule.minutes {
		allowedMinutes[minute] = struct{}{}
	}
	allowedHours := make(map[int]struct{}, len(schedule.hours))
	for _, hour := range schedule.hours {
		allowedHours[hour] = struct{}{}
	}
	for candidate := start; candidate.Before(start.Add(366 * 24 * time.Hour)); candidate = candidate.Add(time.Minute) {
		if _, ok := allowedMinutes[candidate.Minute()]; !ok {
			continue
		}
		if _, ok := allowedHours[candidate.Hour()]; !ok {
			continue
		}
		return candidate, true
	}
	return time.Time{}, false
}

func parseScannerTime(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05"} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed.UTC(), true
		}
	}
	return time.Time{}, false
}

func defaultProviderScope(region string) []string {
	base := []string{"ebay", "amazon"}
	if strings.EqualFold(strings.TrimSpace(region), "AU") {
		return append(base,
			"bonzaslotcars",
			"frontlinehobbies",
			"hobbytechtoys",
			"andrewshobbies",
			"voglers",
			"acercmodels",
			"mrtoys",
		)
	}
	return base
}

func normalizeProviderScope(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, raw := range values {
		normalized := strings.TrimSpace(strings.ToLower(raw))
		if normalized == "" {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	if len(out) == 0 {
		return []string{"ebay", "amazon"}
	}
	return out
}

func BuildRecognitionReview(candidates []RecognitionCandidateInput) (RecognitionReview, error) {
	normalized := make([]RecognitionCandidateInput, 0, len(candidates))
	for _, candidate := range candidates {
		candidate.ID = strings.TrimSpace(candidate.ID)
		candidate.Title = strings.TrimSpace(candidate.Title)
		candidate.Source = strings.TrimSpace(candidate.Source)
		candidate.Provenance = strings.TrimSpace(candidate.Provenance)
		candidate.MediaID = strings.TrimSpace(candidate.MediaID)
		candidate.MediaURL = strings.TrimSpace(candidate.MediaURL)
		candidate.Target = normalizeRecognitionTarget(candidate.Target)
		candidate.OverrideID = strings.TrimSpace(candidate.OverrideID)
		candidate.OverrideBy = strings.TrimSpace(candidate.OverrideBy)
		candidate.OverrideNote = strings.TrimSpace(candidate.OverrideNote)
		if candidate.Confidence < 0 {
			candidate.Confidence = 0
		}
		if candidate.Confidence > 1 {
			candidate.Confidence = 1
		}
		if candidate.ID == "" || candidate.Title == "" {
			continue
		}
		normalized = append(normalized, candidate)
	}
	if len(normalized) == 0 {
		return RecognitionReview{}, fmt.Errorf("recognition candidate is required")
	}

	top := normalized[0]
	for _, candidate := range normalized {
		if candidate.Confidence > top.Confidence {
			top = candidate
		}
	}
	selected := top
	manualOverride := false
	for _, candidate := range normalized {
		if candidate.OverrideID != "" {
			selected = candidate
			manualOverride = true
			break
		}
	}
	alternates := make([]RecognitionCandidateInput, 0, len(normalized)-1)
	for _, candidate := range normalized {
		if candidate.ID != selected.ID {
			alternates = append(alternates, candidate)
		}
	}
	target := selected.Target
	if target == "" {
		target = "inventory"
	}
	review := RecognitionReview{
		TopCandidate:          top,
		Alternates:            alternates,
		SelectedCandidate:     selected,
		ConfidenceLabel:       confidenceLabel(selected.Confidence),
		RequiresManualReview:  selected.Confidence < 0.85 || manualOverride,
		ConfirmBeforeCreate:   true,
		Target:                target,
		MediaEvidence:         map[string]string{},
		Provenance:            recognitionProvenance(normalized),
		ManualOverrideApplied: manualOverride,
	}
	if selected.MediaID != "" {
		review.MediaEvidence["media_id"] = selected.MediaID
	}
	if selected.MediaURL != "" {
		review.MediaEvidence["media_url"] = selected.MediaURL
	}
	return review, nil
}

func normalizeRecognitionTarget(target string) string {
	switch strings.ToLower(strings.TrimSpace(target)) {
	case "wishlist":
		return "wishlist"
	default:
		return "inventory"
	}
}

func confidenceLabel(confidence float64) string {
	switch {
	case confidence >= 0.85:
		return "high"
	case confidence >= 0.6:
		return "medium"
	default:
		return "low"
	}
}

func recognitionProvenance(candidates []RecognitionCandidateInput) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		for _, value := range []string{candidate.Source, candidate.Provenance} {
			value = strings.TrimSpace(value)
			if value == "" {
				continue
			}
			if _, ok := seen[value]; ok {
				continue
			}
			seen[value] = struct{}{}
			out = append(out, value)
		}
	}
	return out
}

func resolveItemsPerPage(providerScope []string, configured int) (requested int, effective int, warning string) {
	requested = configured
	if requested <= 0 {
		requested = 24
	}
	cap := itemsPerPageCap(providerScope)
	effective = requested
	if effective > cap {
		effective = cap
		warning = fmt.Sprintf("items_per_page clamped from %d to %d", requested, cap)
	}
	return requested, effective, warning
}

func itemsPerPageCap(providerScope []string) int {
	cap := 50
	for _, provider := range providerScope {
		normalized := strings.TrimSpace(strings.ToLower(provider))
		if strings.Contains(normalized, "bonza") {
			return 36
		}
	}
	return cap
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
	return s.runNowForProfile(ctx, profileID, querySetID, provider, "manual")
}

func (s *Service) runNowForProfile(ctx context.Context, profileID, querySetID string, provider Provider, triggerType string) (RunResult, error) {
	qs, err := s.GetQuerySetForProfile(ctx, strings.TrimSpace(profileID), querySetID)
	if err != nil {
		return RunResult{}, err
	}
	if provider == nil {
		return RunResult{}, fmt.Errorf("provider is required")
	}
	providerID := providerHealthID(provider, qs)
	runID := uuid.NewString()
	startedAt := time.Now().UTC()
	requestedItemsPerPage, effectiveItemsPerPage, itemsPerPageWarning := resolveItemsPerPage(
		qs.ProviderScope,
		qs.ItemsPerPage,
	)
	qs.ItemsPerPage = effectiveItemsPerPage

	maxAttempts := qs.MaxRetryCount + 1
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	var items []CandidateInput
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		items, lastErr = provider.Search(ctx, qs)
		if lastErr == nil {
			s.recordProviderHealth(ctx, providerID, "ok", "")
			result, persistErr := s.persistCandidatesForProfile(
				ctx,
				strings.TrimSpace(profileID),
				qs.ID,
				items,
				attempt,
				requestedItemsPerPage,
				effectiveItemsPerPage,
				itemsPerPageWarning,
			)
			result.RunID = runID
			if persistErr == nil {
				if err := s.recordRun(ctx, runRecord{
					ID:             runID,
					ProfileID:      strings.TrimSpace(profileID),
					QuerySetID:     qs.ID,
					Provider:       providerID,
					TriggerType:    triggerType,
					StartedAt:      startedAt,
					Status:         "succeeded",
					ResultCount:    result.Saved,
					NewResultCount: result.NewResults,
				}); err != nil {
					return RunResult{}, err
				}
			}
			return result, persistErr
		}
		s.recordProviderHealth(ctx, providerID, "error", lastErr.Error(), retryAfterSecondsFromError(lastErr))
		s.logFailure(ctx, strings.TrimSpace(profileID), qs.ID, providerID, lastErr.Error())
		if attempt < maxAttempts {
			sleep := time.Duration(1000/qs.RateLimitRPS) * time.Millisecond
			if sleep < 100*time.Millisecond {
				sleep = 100 * time.Millisecond
			}
			time.Sleep(sleep * time.Duration(attempt))
		}
	}
	result := RunResult{
		RunID:                 runID,
		Attempts:              maxAttempts,
		ItemsPerPageRequested: requestedItemsPerPage,
		ItemsPerPageEffective: effectiveItemsPerPage,
		ObservedPageSize:      effectiveItemsPerPage,
		PageCount:             0,
		ItemsPerPageWarning:   itemsPerPageWarning,
	}
	if err := s.recordRun(ctx, runRecord{
		ID:             runID,
		ProfileID:      strings.TrimSpace(profileID),
		QuerySetID:     qs.ID,
		Provider:       providerID,
		TriggerType:    triggerType,
		StartedAt:      startedAt,
		Status:         "failed",
		ResultCount:    0,
		NewResultCount: 0,
		ErrorCategory:  "provider_error",
		ErrorMessage:   lastErr.Error(),
		RetryGuidance:  retryGuidanceForProviderFailure(providerID),
	}); err != nil {
		return RunResult{}, err
	}
	return result, fmt.Errorf("run failed: %w", lastErr)
}

func providerHealthID(provider Provider, qs QuerySet) string {
	if identified, ok := provider.(providerIdentifier); ok {
		if id := strings.TrimSpace(strings.ToLower(identified.ProviderID())); id != "" {
			return id
		}
	}
	for _, candidate := range qs.ProviderScope {
		if id := strings.TrimSpace(strings.ToLower(candidate)); id != "" {
			return id
		}
	}
	return "ebay"
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
	var runErrs []error
	for _, qs := range sets {
		if !qs.Enabled || strings.TrimSpace(qs.ScheduleCron) == "" {
			continue
		}
		if _, err := s.runNowForProfile(ctx, strings.TrimSpace(profileID), qs.ID, provider, "scheduled"); err != nil {
			runErrs = append(runErrs, fmt.Errorf("scheduled scanner run failed for query set %s: %w", qs.ID, err))
			continue
		}
		ran++
	}
	return ran, errors.Join(runErrs...)
}

type runRecord struct {
	ID             string
	ProfileID      string
	QuerySetID     string
	Provider       string
	TriggerType    string
	StartedAt      time.Time
	Status         string
	ResultCount    int
	NewResultCount int
	ErrorCategory  string
	ErrorMessage   string
	RetryGuidance  string
}

func (s *Service) recordRun(ctx context.Context, record runRecord) error {
	triggerType := strings.TrimSpace(record.TriggerType)
	if triggerType == "" {
		triggerType = "manual"
	}
	startedAt := record.StartedAt
	if startedAt.IsZero() {
		startedAt = time.Now().UTC()
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO scanner_runs(id, profile_id, query_set_id, provider, trigger_type, started_at, finished_at, status, result_count, new_result_count, error_category, error_message, retry_guidance)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, ?, ?, ?, ?, ?, ?)
	`, strings.TrimSpace(record.ID), strings.TrimSpace(record.ProfileID), strings.TrimSpace(record.QuerySetID), strings.TrimSpace(record.Provider), triggerType, startedAt.Format(time.RFC3339Nano), strings.TrimSpace(record.Status), record.ResultCount, record.NewResultCount, strings.TrimSpace(record.ErrorCategory), strings.TrimSpace(record.ErrorMessage), strings.TrimSpace(record.RetryGuidance))
	if err != nil {
		return fmt.Errorf("record scanner run: %w", err)
	}
	return nil
}

func (s *Service) persistCandidates(ctx context.Context, querySetID string, items []CandidateInput, attempts int) (RunResult, error) {
	return s.persistCandidatesForProfile(ctx, "", querySetID, items, attempts, 0, 0, "")
}

func (s *Service) PersistCandidatesForProfile(
	ctx context.Context,
	profileID,
	querySetID string,
	items []CandidateInput,
	attempts int,
	requestedItemsPerPage int,
	effectiveItemsPerPage int,
	itemsPerPageWarning string,
) (RunResult, error) {
	return s.persistCandidatesForProfile(
		ctx,
		strings.TrimSpace(profileID),
		strings.TrimSpace(querySetID),
		items,
		attempts,
		requestedItemsPerPage,
		effectiveItemsPerPage,
		itemsPerPageWarning,
	)
}

func (s *Service) persistCandidatesForProfile(
	ctx context.Context,
	profileID,
	querySetID string,
	items []CandidateInput,
	attempts int,
	requestedItemsPerPage int,
	effectiveItemsPerPage int,
	itemsPerPageWarning string,
) (RunResult, error) {
	saved := 0
	newResults := 0
	for _, it := range items {
		listingID := strings.TrimSpace(it.ListingID)
		if listingID == "" && strings.TrimSpace(it.URL) != "" {
			listingID = "url:" + strings.TrimSpace(it.URL)
		}
		if listingID == "" {
			continue
		}
		source := strings.TrimSpace(it.Source)
		if source == "" {
			source = "ebay"
		}
		observedCurrency := strings.ToUpper(strings.TrimSpace(it.Currency))
		stockCount := it.StockCount
		if stockCount < -1 {
			stockCount = -1
		}
		stockState := normalizeStockState(it.StockState, stockCount)
		res, err := s.db.ExecContext(ctx, `
			UPDATE scanner_candidates
			SET title = ?, price = ?, shipping = ?, url = ?, image = ?, seller = ?, last_seen = CURRENT_TIMESTAMP,
				status = CASE WHEN status IN ('ignored', 'archived', 'wishlisted', 'purchase_candidate', 'inventory_candidate', 'watching', 'dismissed', 'purchased', 'added_to_wishlist', 'added_to_inventory', 'expired', 'duplicate', 'failed_to_refresh') THEN status ELSE 'seen' END,
				source = ?, stock_state = ?, stock_count = ?, observed_currency = ?
			WHERE query_set_id = ? AND source = ? AND listing_id = ? AND (? = '' OR profile_id = ?)
		`, it.Title, it.Price, it.Shipping, it.URL, it.Image, it.Seller, source, stockState, stockCount, observedCurrency, querySetID, source, listingID, strings.TrimSpace(profileID), strings.TrimSpace(profileID))
		if err != nil {
			return RunResult{}, fmt.Errorf("update candidate: %w", err)
		}
		affected, _ := res.RowsAffected()
		var candidateID string
		if affected == 0 {
			candidateID = uuid.NewString()
			_, err := s.db.ExecContext(ctx, `
				INSERT INTO scanner_candidates(id, profile_id, query_set_id, listing_id, title, price, shipping, url, image, seller, first_seen, last_seen, status, source, stock_state, stock_count, observed_currency)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'new', ?, ?, ?, ?)
			`, candidateID, strings.TrimSpace(profileID), querySetID, listingID, it.Title, it.Price, it.Shipping, it.URL, it.Image, it.Seller, source, stockState, stockCount, observedCurrency)
			if err != nil {
				return RunResult{}, fmt.Errorf("insert candidate: %w", err)
			}
			newResults++
		} else {
			if err := s.db.QueryRowContext(ctx, `
				SELECT id
				FROM scanner_candidates
				WHERE query_set_id = ? AND source = ? AND listing_id = ? AND (? = '' OR profile_id = ?)
			`, querySetID, source, listingID, strings.TrimSpace(profileID), strings.TrimSpace(profileID)).Scan(&candidateID); err != nil {
				return RunResult{}, fmt.Errorf("load candidate id for discovery match: %w", err)
			}
		}
		if strings.TrimSpace(candidateID) != "" {
			if err := s.ensureDiscoveryMatch(ctx, candidateID, it); err != nil {
				return RunResult{}, err
			}
		}
		saved++
	}
	return RunResult{
		Attempts:              attempts,
		Saved:                 saved,
		NewResults:            newResults,
		ItemsPerPageRequested: requestedItemsPerPage,
		ItemsPerPageEffective: effectiveItemsPerPage,
		ObservedPageSize:      effectiveItemsPerPage,
		PageCount:             calculatePageCount(saved, effectiveItemsPerPage),
		ItemsPerPageWarning:   itemsPerPageWarning,
	}, nil
}

func (s *Service) ensureDiscoveryMatch(ctx context.Context, candidateID string, it CandidateInput) error {
	partNumber := strings.TrimSpace(it.ListingID)
	if partNumber == "" {
		partNumber = strings.TrimSpace(it.Title)
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO scanner_matches(candidate_id, item_id, state, confidence, needs_review, extracted_part_number, updated_at)
		VALUES (?, '', 'not_in_collection', 0, 1, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(candidate_id) DO NOTHING
	`, strings.TrimSpace(candidateID), partNumber)
	if err != nil {
		return fmt.Errorf("ensure discovery match: %w", err)
	}
	return nil
}

func calculatePageCount(saved, pageSize int) int {
	if saved <= 0 || pageSize <= 0 {
		return 0
	}
	count := saved / pageSize
	if saved%pageSize != 0 {
		count++
	}
	if count <= 0 {
		return 1
	}
	return count
}

func (s *Service) ListCandidates(ctx context.Context, querySetID string) ([]Candidate, error) {
	return s.ListCandidatesByProfile(ctx, "", querySetID)
}

func (s *Service) ListCandidatesByProfile(ctx context.Context, profileID, querySetID string) ([]Candidate, error) {
	list, err := s.ListCandidatesByProfileFiltered(ctx, profileID, querySetID, CandidateListFilter{})
	if err != nil {
		return nil, err
	}
	return list.Candidates, nil
}

func (s *Service) ListCandidatesByProfileFiltered(ctx context.Context, profileID, querySetID string, filter CandidateListFilter) (CandidateList, error) {
	querySetID = strings.TrimSpace(querySetID)
	if querySetID == "" {
		return CandidateList{}, fmt.Errorf("query set id is required")
	}
	status, statusAlias := normalizeCandidateStatusFilter(filter.Status)
	provider := strings.ToLower(strings.TrimSpace(filter.Provider))
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}
	if pageSize > 100 {
		pageSize = 100
	}
	var total int
	if err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(1)
		FROM scanner_candidates
		WHERE query_set_id = ? AND (? = '' OR profile_id = ?) AND (? = '' OR status = ? OR status = ?) AND (? = '' OR LOWER(source) = ?)
	`, querySetID, strings.TrimSpace(profileID), strings.TrimSpace(profileID), status, status, statusAlias, provider, provider).Scan(&total); err != nil {
		return CandidateList{}, fmt.Errorf("count candidates: %w", err)
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, query_set_id, listing_id, title, price, observed_currency, shipping, url, image, seller, first_seen, last_seen, status, source, stock_state, stock_count
		FROM scanner_candidates
		WHERE query_set_id = ? AND (? = '' OR profile_id = ?) AND (? = '' OR status = ? OR status = ?) AND (? = '' OR LOWER(source) = ?)
		ORDER BY last_seen DESC
		LIMIT ? OFFSET ?
	`, querySetID, strings.TrimSpace(profileID), strings.TrimSpace(profileID), status, status, statusAlias, provider, provider, pageSize, (page-1)*pageSize)
	if err != nil {
		return CandidateList{}, fmt.Errorf("list candidates: %w", err)
	}
	defer rows.Close()
	var out []Candidate
	for rows.Next() {
		var c Candidate
		if err := rows.Scan(&c.ID, &c.QuerySetID, &c.ListingID, &c.Title, &c.Price, &c.Currency, &c.Shipping, &c.URL, &c.Image, &c.Seller, &c.FirstSeen, &c.LastSeen, &c.Status, &c.Source, &c.StockState, &c.StockCount); err != nil {
			return CandidateList{}, fmt.Errorf("scan candidate: %w", err)
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return CandidateList{}, fmt.Errorf("iterate candidates: %w", err)
	}
	return CandidateList{Candidates: out, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *Service) UpdateCandidateStatusForProfile(ctx context.Context, profileID, candidateID, status string) (Candidate, error) {
	candidateID = strings.TrimSpace(candidateID)
	nextStatus := normalizeCandidateStatusForUpdate(status)
	if candidateID == "" {
		return Candidate{}, fmt.Errorf("candidate id is required")
	}
	if nextStatus == "" {
		return Candidate{}, fmt.Errorf("unsupported candidate status")
	}
	res, err := s.db.ExecContext(ctx, `
		UPDATE scanner_candidates
		SET status = ?, last_seen = last_seen
		WHERE id = ? AND (? = '' OR profile_id = ?)
	`, nextStatus, candidateID, strings.TrimSpace(profileID), strings.TrimSpace(profileID))
	if err != nil {
		return Candidate{}, fmt.Errorf("update candidate status: %w", err)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return Candidate{}, fmt.Errorf("candidate not found")
	}
	var c Candidate
	if err := s.db.QueryRowContext(ctx, `
		SELECT id, query_set_id, listing_id, title, price, observed_currency, shipping, url, image, seller, first_seen, last_seen, status, source, stock_state, stock_count
		FROM scanner_candidates
		WHERE id = ? AND (? = '' OR profile_id = ?)
	`, candidateID, strings.TrimSpace(profileID), strings.TrimSpace(profileID)).Scan(&c.ID, &c.QuerySetID, &c.ListingID, &c.Title, &c.Price, &c.Currency, &c.Shipping, &c.URL, &c.Image, &c.Seller, &c.FirstSeen, &c.LastSeen, &c.Status, &c.Source, &c.StockState, &c.StockCount); err != nil {
		return Candidate{}, fmt.Errorf("load candidate: %w", err)
	}
	return c, nil
}

func normalizeCandidateStatusFilter(status string) (string, string) {
	normalized := normalizeCandidateStatusForUpdate(status)
	switch normalized {
	case "dismissed":
		return normalized, "ignored"
	case "expired":
		return normalized, "archived"
	case "added_to_wishlist":
		return normalized, "wishlisted"
	case "purchased":
		return normalized, "purchase_candidate"
	case "added_to_inventory":
		return normalized, "inventory_candidate"
	default:
		return normalized, normalized
	}
}

func normalizeCandidateStatusForUpdate(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "":
		return ""
	case "new", "seen", "watching", "dismissed", "purchased", "added_to_wishlist", "added_to_inventory", "expired", "duplicate", "failed_to_refresh":
		return strings.ToLower(strings.TrimSpace(status))
	case "ignored":
		return "dismissed"
	case "archived":
		return "expired"
	case "wishlisted":
		return "added_to_wishlist"
	case "purchase_candidate":
		return "purchased"
	case "inventory_candidate":
		return "added_to_inventory"
	default:
		return ""
	}
}

func (s *Service) ListRunHistoryByProfile(ctx context.Context, profileID, querySetID string, limit int) ([]RunHistoryRecord, error) {
	profileID = strings.TrimSpace(profileID)
	querySetID = strings.TrimSpace(querySetID)
	if limit <= 0 {
		limit = 25
	}
	if limit > 100 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, profile_id, query_set_id, provider, trigger_type, started_at, finished_at, status,
			result_count, new_result_count, error_category, error_message, retry_guidance
		FROM scanner_runs
		WHERE (? = '' OR profile_id = ?) AND (? = '' OR query_set_id = ?)
		ORDER BY finished_at DESC, started_at DESC
		LIMIT ?
	`, profileID, profileID, querySetID, querySetID, limit)
	if err != nil {
		return nil, fmt.Errorf("list scanner run history: %w", err)
	}
	defer rows.Close()
	var out []RunHistoryRecord
	for rows.Next() {
		var record RunHistoryRecord
		var profile, errorCategory, errorMessage, retryGuidance sql.NullString
		if err := rows.Scan(
			&record.ID,
			&profile,
			&record.QuerySetID,
			&record.Provider,
			&record.TriggerType,
			&record.StartedAt,
			&record.FinishedAt,
			&record.Status,
			&record.ResultCount,
			&record.NewResultCount,
			&errorCategory,
			&errorMessage,
			&retryGuidance,
		); err != nil {
			return nil, fmt.Errorf("scan scanner run history: %w", err)
		}
		record.ProfileID = strings.TrimSpace(profile.String)
		record.Status = normalizeRunStatus(record.Status)
		record.ErrorCategory = strings.TrimSpace(errorCategory.String)
		record.ErrorMessage = strings.TrimSpace(errorMessage.String)
		record.RetryGuidance = strings.TrimSpace(retryGuidance.String)
		if record.RetryGuidance == "" && record.Status == "failed" {
			record.RetryGuidance = retryGuidanceForProviderFailure(record.Provider)
		}
		if record.Status == "failed" {
			record.NextAction = nextActionForProviderFailure(record.Provider)
		}
		out = append(out, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate scanner run history: %w", err)
	}
	return out, nil
}

func (s *Service) ListFailures(ctx context.Context) ([]map[string]string, error) {
	return s.ListFailuresByProfile(ctx, "")
}

func (s *Service) ListFailuresByProfile(ctx context.Context, profileID string) ([]map[string]string, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT query_set_id, provider, message, created_at
		FROM scanner_failures
		WHERE (? = '' OR profile_id = ?)
		ORDER BY created_at DESC
		LIMIT 100
	`, strings.TrimSpace(profileID), strings.TrimSpace(profileID))
	if err != nil {
		return nil, fmt.Errorf("list failures: %w", err)
	}
	defer rows.Close()
	var out []map[string]string
	for rows.Next() {
		var q, p, m, c string
		if err := rows.Scan(&q, &p, &m, &c); err != nil {
			return nil, fmt.Errorf("scan failure: %w", err)
		}
		out = append(out, map[string]string{
			"query_set_id":   q,
			"provider":       p,
			"message":        m,
			"created_at":     c,
			"reason":         m,
			"last_error_at":  c,
			"retry_guidance": retryGuidanceForProviderFailure(p),
			"next_action":    nextActionForProviderFailure(p),
		})
	}
	return out, rows.Err()
}

func retryGuidanceForProviderFailure(provider string) string {
	switch strings.TrimSpace(strings.ToLower(provider)) {
	case "ebay":
		return "Check provider health, credentials, and retry the operation."
	default:
		return "Review provider status and retry the operation."
	}
}

func nextActionForProviderFailure(provider string) string {
	switch strings.TrimSpace(strings.ToLower(provider)) {
	case "ebay":
		return "check_provider_health_and_credentials"
	default:
		return "review_provider_status"
	}
}

func (s *Service) ProviderHealth(ctx context.Context, provider string) (map[string]string, error) {
	var status, msg, updated string
	var retryAfterSeconds int
	err := s.db.QueryRowContext(ctx, `
		SELECT status, message, updated_at, retry_after_seconds
		FROM provider_health WHERE provider = ?
	`, provider).Scan(&status, &msg, &updated, &retryAfterSeconds)
	if err != nil {
		if err == sql.ErrNoRows {
			return map[string]string{"provider": provider, "status": "unknown", "message": ""}, nil
		}
		return nil, fmt.Errorf("provider health: %w", err)
	}
	return map[string]string{"provider": provider, "status": status, "message": msg, "updated_at": updated, "retry_after_seconds": fmt.Sprintf("%d", retryAfterSeconds)}, nil
}

func (s *Service) recordProviderHealth(ctx context.Context, provider, status, message string, retryAfterSeconds ...int) {
	retryAfter := 0
	if len(retryAfterSeconds) > 0 && retryAfterSeconds[0] > 0 {
		retryAfter = retryAfterSeconds[0]
	}
	_, _ = s.db.ExecContext(ctx, `
		INSERT INTO provider_health(provider, status, message, retry_after_seconds, updated_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(provider) DO UPDATE SET
			status = excluded.status,
			message = excluded.message,
			retry_after_seconds = excluded.retry_after_seconds,
			updated_at = CURRENT_TIMESTAMP
	`, provider, status, message, retryAfter)
}

func retryAfterSecondsFromError(err error) int {
	var providerErr retryAfterProviderError
	if errors.As(err, &providerErr) {
		return providerErr.RetryAfter()
	}
	return 0
}

func (s *Service) logFailure(ctx context.Context, profileID, querySetID, provider, message string) {
	_, _ = s.db.ExecContext(ctx, `
		INSERT INTO scanner_failures(id, profile_id, query_set_id, provider, message, created_at)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, uuid.NewString(), strings.TrimSpace(profileID), strings.TrimSpace(querySetID), provider, message)
}

func normalizeStockState(state string, count int) string {
	normalized := strings.TrimSpace(strings.ToLower(state))
	switch normalized {
	case "in_stock", "low_stock", "out_of_stock", "unknown":
		return normalized
	}
	if count == 0 {
		return "out_of_stock"
	}
	if count > 0 && count <= 3 {
		return "low_stock"
	}
	if count > 3 {
		return "in_stock"
	}
	return "unknown"
}
