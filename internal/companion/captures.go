package companion

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/collectors-tech/cabinet/internal/ebaypurchasecapture"
	"github.com/google/uuid"
)

const (
	captureStateAccepted         = "accepted"
	captureStateValidating       = "validating"
	captureStateQueued           = "queued"
	captureStateProcessing       = "processing"
	captureStatePartial          = "partial"
	captureStateReview           = "review"
	captureStateCompleted        = "completed"
	captureStateRetryableFailed  = "retryable-failed"
	captureStatePermanentFailed  = "permanently-failed"
	captureStateCancelled        = "cancelled"
	maxCaptureRecordsPerProfile  = 10_000
	maxCaptureBytesPerProfile    = 64 << 20
	maxCaptureItemsPerSubmission = 200
	maxCaptureStringLength       = 8_192
)

type PayloadSubmission struct {
	ProtocolVersion       string         `json:"protocol_version"`
	ProfileID             string         `json:"profile_id"`
	ModuleID              string         `json:"module_id"`
	ModuleVersion         string         `json:"module_version"`
	SchemaVersion         string         `json:"schema_version"`
	IntegrationInstanceID string         `json:"integration_instance_id"`
	ProviderID            string         `json:"provider_id"`
	URL                   string         `json:"url"`
	PayloadType           string         `json:"payload_type"`
	CapturedAt            string         `json:"captured_at"`
	PageComplete          bool           `json:"page_complete"`
	Passive               bool           `json:"passive"`
	AttemptedWrite        bool           `json:"attempted_write"`
	ConfidenceScore       float64        `json:"confidence_score"`
	RedactionSummary      []string       `json:"redaction_summary"`
	PayloadHash           string         `json:"payload_hash"`
	IdempotencyKey        string         `json:"idempotency_key,omitempty"`
	Data                  map[string]any `json:"data"`
}

type AcceptedPayload struct {
	Accepted        bool           `json:"accepted"`
	Committed       bool           `json:"committed"`
	Replayed        bool           `json:"replayed"`
	CaptureID       string         `json:"capture_id"`
	ProfileID       string         `json:"profile_id"`
	ModuleID        string         `json:"module_id"`
	PayloadType     string         `json:"payload_type"`
	State           string         `json:"state"`
	SyncMode        string         `json:"sync_mode"`
	RemoteWrite     bool           `json:"remote_write"`
	Checkpoint      map[string]any `json:"checkpoint"`
	AuditTrail      []string       `json:"audit_trail"`
	ConfidenceLabel string         `json:"confidence_label"`
}

type CaptureRecord struct {
	ID                    string         `json:"id"`
	ProfileID             string         `json:"profile_id"`
	ModuleID              string         `json:"module_id"`
	ModuleVersion         string         `json:"module_version"`
	SchemaVersion         string         `json:"schema_version"`
	ProviderID            string         `json:"provider_id"`
	IntegrationInstanceID string         `json:"integration_instance_id"`
	PayloadType           string         `json:"payload_type"`
	SourceURL             string         `json:"source_url"`
	CapturedAt            string         `json:"captured_at"`
	PageComplete          bool           `json:"page_complete"`
	PayloadHash           string         `json:"payload_hash"`
	State                 string         `json:"state"`
	AttemptCount          int            `json:"attempt_count"`
	LastError             string         `json:"last_error,omitempty"`
	Checkpoint            map[string]any `json:"checkpoint"`
	CreatedAt             string         `json:"created_at"`
	UpdatedAt             string         `json:"updated_at"`
}

type CaptureInbox struct {
	Captures []CaptureRecord `json:"captures"`
	Pending  int             `json:"pending"`
	Failed   int             `json:"failed"`
	Review   int             `json:"review"`
}

type captureProviderItem struct {
	ListingID       string             `json:"listing_id"`
	VariationID     string             `json:"variation_id,omitempty"`
	Title           string             `json:"title"`
	Price           float64            `json:"price"`
	Currency        string             `json:"currency"`
	Shipping        float64            `json:"shipping,omitempty"`
	URL             string             `json:"url"`
	ImageURL        string             `json:"image_url,omitempty"`
	Seller          string             `json:"seller,omitempty"`
	StockState      string             `json:"stock_state,omitempty"`
	StockCount      int                `json:"stock_count,omitempty"`
	FirstSeen       string             `json:"first_seen,omitempty"`
	LastSeen        string             `json:"last_seen,omitempty"`
	FieldConfidence map[string]float64 `json:"field_confidence,omitempty"`
}

type captureSearchBatch struct {
	Query      string                `json:"query"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"page_size"`
	RangeStart int                   `json:"range_start"`
	RangeEnd   int                   `json:"range_end"`
	TotalPages int                   `json:"total_pages"`
	Complete   bool                  `json:"complete"`
	Items      []captureProviderItem `json:"items"`
}

type captureItemDetail struct {
	Item captureProviderItem `json:"item"`
}

type capturePurchaseItem struct {
	OrderID           string `json:"order_id,omitempty"`
	ListingID         string `json:"listing_id,omitempty"`
	VariationID       string `json:"variation_id,omitempty"`
	TransactionID     string `json:"transaction_id,omitempty"`
	Title             string `json:"title,omitempty"`
	PurchasedIdentity string `json:"purchased_identity,omitempty"`
	Quantity          int    `json:"quantity,omitempty"`
	ItemPrice         string `json:"item_price,omitempty"`
	ImageURL          string `json:"image_url,omitempty"`
	ItemURL           string `json:"item_url,omitempty"`
	Seller            string `json:"seller,omitempty"`
	OrderTotal        string `json:"order_total,omitempty"`
	Currency          string `json:"currency,omitempty"`
	Shipping          string `json:"shipping,omitempty"`
	Tax               string `json:"tax,omitempty"`
	ImportCharges     string `json:"import_charges,omitempty"`
	OrderStatus       string `json:"order_status,omitempty"`
	TrackingStatus    string `json:"tracking_status,omitempty"`
}

type capturePurchaseBatch struct {
	Cards []capturePurchaseItem `json:"cards"`
}

type captureReadiness struct {
	State          string   `json:"state"`
	EvidenceIDs    []string `json:"evidence_ids"`
	FixtureVersion string   `json:"fixture_version"`
}

type captureUserIntent struct {
	Intent    string `json:"intent"`
	TargetKey string `json:"target_key"`
	Confirmed bool   `json:"confirmed"`
}

type normalisedCapture struct {
	Search    *captureSearchBatch
	Items     []captureProviderItem
	Purchases []capturePurchaseItem
	Readiness *captureReadiness
	Intent    *captureUserIntent
}

func PayloadDigest(data map[string]any) string {
	raw, _ := json.Marshal(data)
	digest := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func (s *Service) AcceptPayload(ctx context.Context, in PayloadSubmission, authorization string, metadata RequestMetadata) (AcceptedPayload, error) {
	profileID := strings.TrimSpace(in.ProfileID)
	if profileID == "" {
		return AcceptedPayload{}, protocolError("companion_profile_required")
	}
	session, err := s.Authenticate(ctx, authorization, metadata, CapabilityCapturesSubmit)
	if err != nil {
		return AcceptedPayload{}, err
	}
	if session.ProfileID != profileID {
		return AcceptedPayload{}, protocolError("companion_profile_mismatch")
	}
	release, err := s.acquireSession(session.ID)
	if err != nil {
		return AcceptedPayload{}, err
	}
	defer release()

	module, err := s.validateCaptureEnvelope(ctx, session, &in, metadata)
	if err != nil {
		return AcceptedPayload{}, err
	}
	if _, err := normalisePayloadData(in, module); err != nil {
		return AcceptedPayload{}, err
	}

	capture, replayed, err := s.persistCapture(ctx, session, in, metadata)
	if err != nil {
		return AcceptedPayload{}, err
	}
	if !replayed && capture.State != captureStateCancelled {
		_, _ = s.db.ExecContext(ctx, `UPDATE companion_captures SET state = ?, updated_at = ? WHERE id = ? AND state = ?`, captureStateQueued, formatTime(s.options.Now().UTC()), capture.ID, captureStateAccepted)
		capture, err = s.processCapture(ctx, capture.ID)
		if err != nil {
			capture, _ = s.captureByID(ctx, profileID, capture.ID)
		}
	}
	s.recordAudit(ctx, session.ProfileID, session.ID, "capture.transport.accepted", capture.State, metadata)
	return acceptedPayload(capture, in.ConfidenceScore, replayed), nil
}

func (s *Service) validateCaptureEnvelope(ctx context.Context, session Session, in *PayloadSubmission, metadata RequestMetadata) (Module, error) {
	moduleID := boundedCaptureText(in.ModuleID, 128)
	payloadType := boundedCaptureText(in.PayloadType, 128)
	integrationID := boundedCaptureText(in.IntegrationInstanceID, 128)
	providerID := boundedCaptureText(in.ProviderID, 128)
	if moduleID == "" || payloadType == "" || integrationID == "" || providerID == "" {
		return Module{}, protocolError("companion_capture_contract_required")
	}
	if in.ProtocolVersion != ProtocolVersionV1 || in.ProtocolVersion != session.ProtocolVersion {
		return Module{}, protocolError("companion_protocol_version_unsupported")
	}
	registry, err := s.registryForProfile(ctx, session.ProfileID)
	if err != nil {
		return Module{}, err
	}
	var module Module
	for _, candidate := range registry.Modules {
		if candidate.ID == moduleID && candidate.IntegrationInstanceID == integrationID {
			module = candidate
			break
		}
	}
	if module.ID == "" {
		return Module{}, protocolError("companion_module_not_registered")
	}
	if module.ProviderID != providerID || module.ModuleVersion != strings.TrimSpace(in.ModuleVersion) || module.FixtureVersion != strings.TrimSpace(in.SchemaVersion) {
		return Module{}, protocolError("companion_capture_contract_mismatch")
	}
	schemaFound := false
	for _, schema := range module.CaptureSchemas {
		if schema.PayloadType == payloadType {
			schemaFound = true
			break
		}
	}
	if !schemaFound {
		return Module{}, protocolError("companion_payload_type_unsupported")
	}
	if !captureURLAllowed(module, in.URL) {
		return Module{}, protocolError("companion_capture_url_rejected")
	}
	in.URL = sanitiseCaptureURL(in.URL)
	if !in.Passive || in.AttemptedWrite {
		return Module{}, protocolError("companion_payload_must_be_passive")
	}
	if in.ConfidenceScore < 0 || in.ConfidenceScore > 1 {
		return Module{}, protocolError("companion_confidence_score_out_of_range")
	}
	capturedAt, parseErr := time.Parse(time.RFC3339, strings.TrimSpace(in.CapturedAt))
	if parseErr != nil || capturedAt.After(s.options.Now().Add(5*time.Minute)) {
		return Module{}, protocolError("companion_captured_at_invalid")
	}
	in.CapturedAt = capturedAt.UTC().Format(time.RFC3339)
	if len(in.RedactionSummary) == 0 || len(in.RedactionSummary) > 20 || !redactionSummaryAllowed(module.RedactionRules, in.RedactionSummary) {
		return Module{}, protocolError("companion_redaction_summary_invalid")
	}
	if hasForbiddenCaptureField(in.Data) {
		return Module{}, protocolError("companion_payload_redaction_failed")
	}
	if in.PayloadHash != PayloadDigest(in.Data) {
		return Module{}, protocolError("companion_payload_hash_mismatch")
	}
	idempotencyKey := boundedCaptureText(metadata.IdempotencyKey, 128)
	if idempotencyKey == "" {
		idempotencyKey = boundedCaptureText(in.IdempotencyKey, 128)
	}
	if idempotencyKey == "" {
		return Module{}, protocolError("companion_idempotency_key_required")
	}
	in.IdempotencyKey = idempotencyKey
	return module, nil
}

func (s *Service) persistCapture(ctx context.Context, session Session, in PayloadSubmission, metadata RequestMetadata) (CaptureRecord, bool, error) {
	raw, err := json.Marshal(in)
	if err != nil || len(raw) > companionJSONPayloadLimit {
		return CaptureRecord{}, false, protocolError("companion_payload_too_large")
	}
	var count int
	var bytesUsed int64
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*), COALESCE(SUM(LENGTH(raw_payload_json)), 0) FROM companion_captures WHERE profile_id = ?`, session.ProfileID).Scan(&count, &bytesUsed); err != nil {
		return CaptureRecord{}, false, protocolError("companion_capture_store_failed")
	}
	if count >= maxCaptureRecordsPerProfile || bytesUsed+int64(len(raw)) > maxCaptureBytesPerProfile {
		return CaptureRecord{}, false, protocolError("companion_capture_quota_exceeded")
	}
	if existing, lookupErr := s.captureByIdempotency(ctx, session.ProfileID, in.IdempotencyKey); lookupErr == nil {
		if existing.PayloadHash != in.PayloadHash {
			return CaptureRecord{}, false, protocolError("companion_idempotency_conflict")
		}
		return existing, true, nil
	} else if lookupErr != sql.ErrNoRows {
		return CaptureRecord{}, false, protocolError("companion_capture_store_failed")
	}
	now := formatTime(s.options.Now().UTC())
	id := uuid.NewString()
	redactions, _ := json.Marshal(in.RedactionSummary)
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO companion_captures(
			id, profile_id, session_id, module_id, module_version, schema_version, provider_id,
			integration_instance_id, payload_type, source_url, captured_at, page_complete,
			payload_hash, idempotency_key, redaction_summary_json, raw_payload_json,
			state, created_at, updated_at
		) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, id, session.ProfileID, session.ID, strings.TrimSpace(in.ModuleID), strings.TrimSpace(in.ModuleVersion), strings.TrimSpace(in.SchemaVersion),
		strings.TrimSpace(in.ProviderID), strings.TrimSpace(in.IntegrationInstanceID), strings.TrimSpace(in.PayloadType), sanitiseCaptureURL(in.URL),
		in.CapturedAt, boolInt(in.PageComplete), in.PayloadHash, in.IdempotencyKey, string(redactions), string(raw), captureStateAccepted, now, now)
	if err != nil {
		if existing, lookupErr := s.captureByIdempotency(ctx, session.ProfileID, in.IdempotencyKey); lookupErr == nil && existing.PayloadHash == in.PayloadHash {
			return existing, true, nil
		}
		return CaptureRecord{}, false, protocolError("companion_capture_store_failed")
	}
	s.recordAudit(ctx, session.ProfileID, session.ID, "capture.persisted", "accepted", metadata)
	record, err := s.captureByID(ctx, session.ProfileID, id)
	return record, false, err
}

const companionJSONPayloadLimit = 1 << 20

func (s *Service) processCapture(ctx context.Context, captureID string) (CaptureRecord, error) {
	var profileID, rawJSON, state string
	if err := s.db.QueryRowContext(ctx, `SELECT profile_id, raw_payload_json, state FROM companion_captures WHERE id = ?`, captureID).Scan(&profileID, &rawJSON, &state); err != nil {
		return CaptureRecord{}, protocolError("companion_capture_not_found")
	}
	if state == captureStateCompleted || state == captureStatePartial || state == captureStateReview || state == captureStateCancelled {
		return s.captureByID(ctx, profileID, captureID)
	}
	now := formatTime(s.options.Now().UTC())
	if _, err := s.db.ExecContext(ctx, `UPDATE companion_captures SET state = ?, attempt_count = attempt_count + 1, updated_at = ? WHERE id = ? AND state != ?`, captureStateValidating, now, captureID, captureStateCancelled); err != nil {
		return CaptureRecord{}, protocolError("companion_capture_process_failed")
	}
	var in PayloadSubmission
	if err := json.Unmarshal([]byte(rawJSON), &in); err != nil {
		s.failCapture(ctx, captureID, captureStatePermanentFailed, "invalid_persisted_payload")
		return CaptureRecord{}, protocolError("companion_capture_process_failed")
	}
	registry, err := s.registryForProfile(ctx, profileID)
	if err != nil {
		s.failCapture(ctx, captureID, captureStateRetryableFailed, "module_registry_unavailable")
		return CaptureRecord{}, err
	}
	var module Module
	for _, candidate := range registry.Modules {
		if candidate.ID == in.ModuleID && candidate.IntegrationInstanceID == in.IntegrationInstanceID {
			module = candidate
			break
		}
	}
	if module.ID == "" {
		s.failCapture(ctx, captureID, captureStatePermanentFailed, "module_not_registered")
		return CaptureRecord{}, protocolError("companion_module_not_registered")
	}
	normalised, err := normalisePayloadData(in, module)
	if err != nil {
		s.failCapture(ctx, captureID, captureStatePermanentFailed, ErrorCode(err))
		return CaptureRecord{}, err
	}
	if _, err := s.db.ExecContext(ctx, `UPDATE companion_captures SET state = ?, updated_at = ? WHERE id = ? AND state != ?`, captureStateProcessing, formatTime(s.options.Now().UTC()), captureID, captureStateCancelled); err != nil {
		return CaptureRecord{}, protocolError("companion_capture_process_failed")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		s.failCapture(ctx, captureID, captureStateRetryableFailed, "database_busy")
		return CaptureRecord{}, protocolError("companion_capture_process_failed")
	}
	defer tx.Rollback()
	observations, candidates, purchases := 0, 0, 0
	if normalised.Search != nil || len(normalised.Items) > 0 {
		items := normalised.Items
		query := ""
		if normalised.Search != nil {
			items = normalised.Search.Items
			query = normalised.Search.Query
		}
		observations, candidates, err = dispatchProviderItems(ctx, tx, captureID, in, items, query)
	}
	if err == nil && len(normalised.Purchases) > 0 {
		purchases, err = dispatchPurchaseItems(ctx, tx, captureID, in, normalised.Purchases)
	}
	if err == nil && normalised.Readiness != nil {
		observations, err = dispatchDiagnosticObservation(ctx, tx, captureID, in, "readiness", normalised.Readiness, in.ConfidenceScore)
	}
	if err == nil && normalised.Intent != nil {
		observations, err = dispatchDiagnosticObservation(ctx, tx, captureID, in, "intent:"+normalised.Intent.TargetKey, normalised.Intent, in.ConfidenceScore)
	}
	if err != nil {
		s.failCapture(ctx, captureID, captureStateRetryableFailed, "dispatch_failed")
		return CaptureRecord{}, protocolError("companion_capture_process_failed")
	}
	finalState := captureStateCompleted
	if !in.PageComplete || (normalised.Search != nil && !normalised.Search.Complete) {
		finalState = captureStatePartial
	} else if captureNeedsReview(in, normalised) {
		finalState = captureStateReview
	}
	checkpoint := map[string]any{
		"committed": true, "observations": observations, "candidates": candidates,
		"purchase_items": purchases, "page_complete": in.PageComplete,
	}
	checkpointJSON, _ := json.Marshal(checkpoint)
	result, err := tx.ExecContext(ctx, `UPDATE companion_captures SET state = ?, checkpoint_json = ?, last_error = '', updated_at = ? WHERE id = ? AND state = ?`, finalState, string(checkpointJSON), now, captureID, captureStateProcessing)
	if err != nil {
		s.failCapture(ctx, captureID, captureStateRetryableFailed, "checkpoint_failed")
		return CaptureRecord{}, protocolError("companion_capture_process_failed")
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return CaptureRecord{}, protocolError("companion_capture_cancelled")
	}
	if err := tx.Commit(); err != nil {
		s.failCapture(ctx, captureID, captureStateRetryableFailed, "commit_failed")
		return CaptureRecord{}, protocolError("companion_capture_process_failed")
	}
	return s.captureByID(ctx, profileID, captureID)
}

func captureNeedsReview(in PayloadSubmission, normalised normalisedCapture) bool {
	if in.ConfidenceScore < 0.8 || len(normalised.Purchases) > 0 || normalised.Intent != nil {
		return true
	}
	items := normalised.Items
	if normalised.Search != nil {
		items = normalised.Search.Items
	}
	for _, item := range items {
		if lowFieldConfidence(item.FieldConfidence) {
			return true
		}
	}
	return false
}

func dispatchProviderItems(ctx context.Context, tx *sql.Tx, captureID string, in PayloadSubmission, items []captureProviderItem, query string) (int, int, error) {
	now := strings.TrimSpace(in.CapturedAt)
	querySetID := deterministicID("companion-query", in.ProfileID, in.IntegrationInstanceID, query)
	keywords, _ := json.Marshal([]string{query})
	providers, _ := json.Marshal([]string{in.ProviderID})
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO scanner_query_sets(id, profile_id, name, keywords_json, exclusions_json, provider_scope_json, enabled, created_at, updated_at)
		VALUES(?, ?, ?, ?, '[]', ?, 1, ?, ?)
		ON CONFLICT(id) DO UPDATE SET name=excluded.name, keywords_json=excluded.keywords_json,
			provider_scope_json=excluded.provider_scope_json, updated_at=excluded.updated_at
	`, querySetID, in.ProfileID, "Browser Companion: "+in.ProviderID, string(keywords), string(providers), now, now); err != nil {
		return 0, 0, err
	}
	for _, item := range items {
		listingKey := item.ListingID
		if item.VariationID != "" {
			listingKey += ":" + item.VariationID
		}
		normalizedJSON, _ := json.Marshal(item)
		review := in.ConfidenceScore < 0.8 || lowFieldConfidence(item.FieldConfidence)
		observationID := deterministicID("companion-observation", in.ProfileID, in.IntegrationInstanceID, listingKey, in.PayloadType)
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO companion_observations(id, capture_id, profile_id, provider_id, integration_instance_id,
				listing_key, payload_type, normalized_json, confidence, review_required, first_seen, last_seen, created_at, updated_at)
			VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(profile_id, integration_instance_id, listing_key, payload_type) DO UPDATE SET
				capture_id=excluded.capture_id, normalized_json=excluded.normalized_json, confidence=excluded.confidence,
				review_required=excluded.review_required, last_seen=excluded.last_seen, updated_at=excluded.updated_at
		`, observationID, captureID, in.ProfileID, in.ProviderID, in.IntegrationInstanceID, listingKey, in.PayloadType,
			string(normalizedJSON), in.ConfidenceScore, boolInt(review), firstNonEmpty(item.FirstSeen, now), firstNonEmpty(item.LastSeen, now), now, now); err != nil {
			return 0, 0, err
		}
		candidateID := deterministicID("companion-candidate", in.ProfileID, querySetID, in.ProviderID, listingKey)
		status := "new"
		if review {
			status = "review"
		}
		stockState := firstNonEmpty(item.StockState, "unknown")
		stockCount := item.StockCount
		if item.StockState == "" && item.StockCount == 0 {
			stockCount = -1
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO scanner_candidates(id, profile_id, query_set_id, listing_id, title, price, shipping, url,
				image, seller, first_seen, last_seen, status, source, observed_currency, source_result_url, stock_state, stock_count)
			VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(profile_id, query_set_id, source, listing_id) DO UPDATE SET
				title=excluded.title, price=excluded.price, shipping=excluded.shipping, url=excluded.url,
				image=excluded.image, seller=excluded.seller, last_seen=excluded.last_seen,
				observed_currency=excluded.observed_currency, source_result_url=excluded.source_result_url,
				stock_state=excluded.stock_state, stock_count=excluded.stock_count
		`, candidateID, in.ProfileID, querySetID, listingKey, item.Title, item.Price, item.Shipping, item.URL,
			item.ImageURL, item.Seller, firstNonEmpty(item.FirstSeen, now), firstNonEmpty(item.LastSeen, now), status,
			in.ProviderID, item.Currency, sanitiseCaptureURL(in.URL), stockState, stockCount); err != nil {
			return 0, 0, err
		}
	}
	runID := deterministicID("companion-run", captureID)
	if _, err := tx.ExecContext(ctx, `
		INSERT OR IGNORE INTO scanner_runs(id, profile_id, query_set_id, provider, trigger_type, started_at, finished_at, status, result_count, new_result_count)
		VALUES(?, ?, ?, ?, 'browser_companion', ?, ?, 'completed', ?, ?)
	`, runID, in.ProfileID, querySetID, in.ProviderID, now, now, len(items), len(items)); err != nil {
		return 0, 0, err
	}
	return len(items), len(items), nil
}

func dispatchPurchaseItems(ctx context.Context, tx *sql.Tx, captureID string, in PayloadSubmission, items []capturePurchaseItem) (int, error) {
	now := strings.TrimSpace(in.CapturedAt)
	for _, item := range items {
		card := purchaseCard(item)
		itemKey := ebaypurchasecapture.PurchaseItemKey(card)
		if itemKey == "" {
			return 0, fmt.Errorf("purchase item key is required")
		}
		orderKey := firstNonEmpty(item.OrderID, "orderless:"+itemKey)
		cardJSON, _ := json.Marshal(card)
		id := deterministicID("companion-purchase", in.ProfileID, in.ProviderID, itemKey)
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO companion_purchase_inbox(id, capture_id, profile_id, provider_id, order_key, item_key, card_json, state, first_seen, last_seen, created_at, updated_at)
			VALUES(?, ?, ?, ?, ?, ?, ?, 'review', ?, ?, ?, ?)
			ON CONFLICT(profile_id, provider_id, item_key) DO UPDATE SET
				capture_id=excluded.capture_id, order_key=excluded.order_key, card_json=excluded.card_json,
				last_seen=excluded.last_seen, updated_at=excluded.updated_at
		`, id, captureID, in.ProfileID, in.ProviderID, orderKey, itemKey, string(cardJSON), now, now, now, now); err != nil {
			return 0, err
		}
	}
	return len(items), nil
}

func dispatchDiagnosticObservation(ctx context.Context, tx *sql.Tx, captureID string, in PayloadSubmission, key string, value any, confidence float64) (int, error) {
	now := strings.TrimSpace(in.CapturedAt)
	normalizedJSON, _ := json.Marshal(value)
	id := deterministicID("companion-observation", in.ProfileID, in.IntegrationInstanceID, key, in.PayloadType)
	_, err := tx.ExecContext(ctx, `
		INSERT INTO companion_observations(id, capture_id, profile_id, provider_id, integration_instance_id,
			listing_key, payload_type, normalized_json, confidence, review_required, first_seen, last_seen, created_at, updated_at)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(profile_id, integration_instance_id, listing_key, payload_type) DO UPDATE SET
			capture_id=excluded.capture_id, normalized_json=excluded.normalized_json, confidence=excluded.confidence,
			review_required=excluded.review_required, last_seen=excluded.last_seen, updated_at=excluded.updated_at
	`, id, captureID, in.ProfileID, in.ProviderID, in.IntegrationInstanceID, key, in.PayloadType, string(normalizedJSON),
		confidence, boolInt(confidence < 0.8 || strings.HasPrefix(key, "intent:")), now, now, now, now)
	return 1, err
}

func normalisePayloadData(in PayloadSubmission, module Module) (normalisedCapture, error) {
	schema, ok := moduleCaptureSchema(module, in.PayloadType)
	if !ok {
		return normalisedCapture{}, protocolError("companion_payload_type_unsupported")
	}
	if err := validateTopLevelFields(in.Data, schema); err != nil {
		return normalisedCapture{}, err
	}
	switch in.PayloadType {
	case "search_results", "provider_search_results":
		var data captureSearchBatch
		if err := strictDecode(in.Data, &data); err != nil || len(data.Items) == 0 || len(data.Items) > maxCaptureItemsPerSubmission || data.Page < 1 || data.PageSize < 1 || data.PageSize > maxCaptureItemsPerSubmission {
			return normalisedCapture{}, protocolError("companion_payload_schema_invalid")
		}
		if data.RangeStart < 0 || data.RangeEnd < data.RangeStart || data.RangeEnd-data.RangeStart+1 < len(data.Items) {
			return normalisedCapture{}, protocolError("companion_payload_schema_invalid")
		}
		for index := range data.Items {
			if err := validateProviderItem(&data.Items[index], module); err != nil {
				return normalisedCapture{}, err
			}
		}
		return normalisedCapture{Search: &data}, nil
	case "item_detail":
		var data captureItemDetail
		if err := strictDecode(in.Data, &data); err != nil || validateProviderItem(&data.Item, module) != nil {
			return normalisedCapture{}, protocolError("companion_payload_schema_invalid")
		}
		return normalisedCapture{Items: []captureProviderItem{data.Item}}, nil
	case "purchase_order":
		var items []capturePurchaseItem
		if _, hasCards := in.Data["cards"]; hasCards {
			var batch capturePurchaseBatch
			if err := strictDecode(in.Data, &batch); err != nil {
				return normalisedCapture{}, protocolError("companion_payload_schema_invalid")
			}
			items = batch.Cards
		} else {
			var item capturePurchaseItem
			if err := strictDecode(in.Data, &item); err != nil {
				return normalisedCapture{}, protocolError("companion_payload_schema_invalid")
			}
			items = []capturePurchaseItem{item}
		}
		if len(items) == 0 || len(items) > maxCaptureItemsPerSubmission {
			return normalisedCapture{}, protocolError("companion_payload_schema_invalid")
		}
		for index := range items {
			if err := validatePurchaseItem(&items[index], module); err != nil {
				return normalisedCapture{}, err
			}
		}
		return normalisedCapture{Purchases: items}, nil
	case "readiness_diagnostic":
		var data captureReadiness
		if err := strictDecode(in.Data, &data); err != nil || !containsString([]string{"ready", "logged_out", "challenge", "unsupported"}, data.State) || data.FixtureVersion != module.FixtureVersion || len(data.EvidenceIDs) > 20 {
			return normalisedCapture{}, protocolError("companion_payload_schema_invalid")
		}
		return normalisedCapture{Readiness: &data}, nil
	case "user_intent":
		var data captureUserIntent
		if err := strictDecode(in.Data, &data); err != nil || !data.Confirmed || boundedCaptureText(data.Intent, 128) == "" || boundedCaptureText(data.TargetKey, 256) == "" {
			return normalisedCapture{}, protocolError("companion_payload_schema_invalid")
		}
		return normalisedCapture{Intent: &data}, nil
	default:
		return normalisedCapture{}, protocolError("companion_payload_type_unsupported")
	}
}

func moduleCaptureSchema(module Module, payloadType string) (CaptureSchema, bool) {
	for _, schema := range module.CaptureSchemas {
		if schema.PayloadType == payloadType {
			return schema, true
		}
	}
	return CaptureSchema{}, false
}

func validateTopLevelFields(data map[string]any, schema CaptureSchema) error {
	if data == nil {
		return protocolError("companion_payload_schema_invalid")
	}
	allowed := map[string]struct{}{}
	for _, field := range append(append([]string{}, schema.Fields...), schema.MediaFields...) {
		allowed[field] = struct{}{}
	}
	for field := range data {
		if _, ok := allowed[field]; !ok {
			return protocolError("companion_payload_field_rejected")
		}
	}
	return nil
}

func strictDecode(data map[string]any, target any) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func validateProviderItem(item *captureProviderItem, module Module) error {
	item.ListingID = boundedCaptureText(item.ListingID, 256)
	item.VariationID = boundedCaptureText(item.VariationID, 256)
	item.Title = boundedCaptureText(item.Title, 2_048)
	item.Currency = strings.ToUpper(boundedCaptureText(item.Currency, 3))
	item.Seller = boundedCaptureText(item.Seller, 512)
	item.StockState = boundedCaptureText(item.StockState, 64)
	if item.ListingID == "" || item.Title == "" || len(item.Currency) != 3 || item.Price < 0 || item.Shipping < 0 || !captureURLAllowed(module, item.URL) {
		return protocolError("companion_payload_schema_invalid")
	}
	item.URL = sanitiseCaptureURL(item.URL)
	if item.ImageURL != "" {
		if !captureURLAllowed(module, item.ImageURL) {
			return protocolError("companion_payload_cross_provider")
		}
		item.ImageURL = sanitiseCaptureURL(item.ImageURL)
	}
	for field, confidence := range item.FieldConfidence {
		if boundedCaptureText(field, 128) == "" || confidence < 0 || confidence > 1 || len(item.FieldConfidence) > 32 {
			return protocolError("companion_payload_schema_invalid")
		}
	}
	return nil
}

func validatePurchaseItem(item *capturePurchaseItem, module Module) error {
	values := []*string{&item.OrderID, &item.ListingID, &item.VariationID, &item.TransactionID, &item.Title,
		&item.PurchasedIdentity, &item.ItemPrice, &item.Seller, &item.OrderTotal, &item.Currency,
		&item.Shipping, &item.Tax, &item.ImportCharges, &item.OrderStatus, &item.TrackingStatus}
	for _, value := range values {
		*value = boundedCaptureText(*value, 2_048)
	}
	if item.Quantity < 0 || (item.ListingID == "" && item.TransactionID == "" && item.PurchasedIdentity == "" && item.Title == "") {
		return protocolError("companion_payload_schema_invalid")
	}
	if item.ItemURL != "" {
		if !captureURLAllowed(module, item.ItemURL) {
			return protocolError("companion_payload_cross_provider")
		}
		item.ItemURL = sanitiseCaptureURL(item.ItemURL)
	}
	if item.ImageURL != "" {
		if !captureURLAllowed(module, item.ImageURL) {
			return protocolError("companion_payload_cross_provider")
		}
		item.ImageURL = sanitiseCaptureURL(item.ImageURL)
	}
	return nil
}

func purchaseCard(item capturePurchaseItem) ebaypurchasecapture.PurchaseCard {
	return ebaypurchasecapture.PurchaseCard{
		OrderID: item.OrderID, ListingID: item.ListingID, VariationID: item.VariationID, TransactionID: item.TransactionID,
		ListingTitle: item.Title, PurchasedIdentity: item.PurchasedIdentity, Quantity: item.Quantity, ItemPrice: item.ItemPrice,
		ImageURL: item.ImageURL, ItemURL: item.ItemURL, SellerUsername: item.Seller, OrderTotal: item.OrderTotal,
		Currency: item.Currency, Shipping: item.Shipping, Tax: item.Tax, ImportCharges: item.ImportCharges,
		OrderStatus: item.OrderStatus, TrackingStatus: item.TrackingStatus,
	}
}

func (s *Service) ListCaptures(ctx context.Context, authorization string, metadata RequestMetadata, state string, limit int) (CaptureInbox, error) {
	session, err := s.Authenticate(ctx, authorization, metadata, CapabilityCapturesSubmit)
	if err != nil {
		return CaptureInbox{}, err
	}
	if limit < 1 || limit > 200 {
		limit = 50
	}
	query := `SELECT id, profile_id, module_id, module_version, schema_version, provider_id, integration_instance_id,
		payload_type, source_url, captured_at, page_complete, payload_hash, state, attempt_count, last_error,
		checkpoint_json, created_at, updated_at FROM companion_captures WHERE profile_id = ?`
	args := []any{session.ProfileID}
	if state = strings.TrimSpace(state); state != "" {
		query += ` AND state = ?`
		args = append(args, state)
	}
	query += ` ORDER BY created_at DESC LIMIT ?`
	args = append(args, limit)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return CaptureInbox{}, protocolError("companion_capture_list_failed")
	}
	defer rows.Close()
	inbox := CaptureInbox{Captures: []CaptureRecord{}}
	for rows.Next() {
		record, scanErr := scanCapture(rows)
		if scanErr != nil {
			return CaptureInbox{}, protocolError("companion_capture_list_failed")
		}
		inbox.Captures = append(inbox.Captures, record)
	}
	_ = s.db.QueryRowContext(ctx, `SELECT
		COALESCE(SUM(CASE WHEN state IN ('accepted','validating','queued','processing','retryable-failed') THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN state IN ('retryable-failed','permanently-failed') THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN state = 'review' THEN 1 ELSE 0 END), 0)
		FROM companion_captures WHERE profile_id = ?`, session.ProfileID).Scan(&inbox.Pending, &inbox.Failed, &inbox.Review)
	return inbox, nil
}

func (s *Service) CancelCapture(ctx context.Context, authorization string, metadata RequestMetadata, captureID string) (CaptureRecord, error) {
	session, err := s.Authenticate(ctx, authorization, metadata, CapabilityCapturesSubmit)
	if err != nil {
		return CaptureRecord{}, err
	}
	now := formatTime(s.options.Now().UTC())
	result, err := s.db.ExecContext(ctx, `UPDATE companion_captures SET state = ?, cancelled_at = ?, updated_at = ?
		WHERE id = ? AND profile_id = ? AND state IN ('accepted','validating','queued','processing','retryable-failed')`, captureStateCancelled, now, now, strings.TrimSpace(captureID), session.ProfileID)
	if err != nil {
		return CaptureRecord{}, protocolError("companion_capture_cancel_failed")
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return CaptureRecord{}, protocolError("companion_capture_not_cancellable")
	}
	return s.captureByID(ctx, session.ProfileID, captureID)
}

func (s *Service) ResumePendingCaptures(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, `UPDATE companion_captures SET state = ?, last_error = 'cabinet_restarted', updated_at = CURRENT_TIMESTAMP WHERE state IN (?, ?)`, captureStateRetryableFailed, captureStateValidating, captureStateProcessing); err != nil {
		return err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id FROM companion_captures WHERE state IN (?, ?, ?) ORDER BY created_at ASC LIMIT 100`, captureStateAccepted, captureStateQueued, captureStateRetryableFailed)
	if err != nil {
		return err
	}
	ids := []string{}
	for rows.Next() {
		var id string
		if rows.Scan(&id) == nil {
			ids = append(ids, id)
		}
	}
	rows.Close()
	for _, id := range ids {
		_, _ = s.processCapture(ctx, id)
	}
	return nil
}

func (s *Service) PurchaseCards(ctx context.Context, profileID string) ([]ebaypurchasecapture.PurchaseCard, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT card_json FROM companion_purchase_inbox WHERE profile_id = ? AND state = 'review' ORDER BY last_seen DESC`, strings.TrimSpace(profileID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cards := []ebaypurchasecapture.PurchaseCard{}
	for rows.Next() {
		var raw string
		var card ebaypurchasecapture.PurchaseCard
		if rows.Scan(&raw) != nil || json.Unmarshal([]byte(raw), &card) != nil {
			continue
		}
		cards = append(cards, card)
	}
	return cards, rows.Err()
}

func (s *Service) MarkPurchaseHandOff(ctx context.Context, profileID, itemKey, state, itemID string) error {
	if state != "linked" && state != "converted" {
		return fmt.Errorf("invalid purchase hand-off state")
	}
	result, err := s.db.ExecContext(ctx, `UPDATE companion_purchase_inbox SET state = ?, linked_item_id = ?, updated_at = CURRENT_TIMESTAMP WHERE profile_id = ? AND item_key = ?`, state, strings.TrimSpace(itemID), strings.TrimSpace(profileID), strings.TrimSpace(itemKey))
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Service) CaptureMediaContext(ctx context.Context, session Session, captureID, fieldName string) (CaptureRecord, error) {
	record, err := s.captureByID(ctx, session.ProfileID, strings.TrimSpace(captureID))
	if err != nil {
		return CaptureRecord{}, protocolError("companion_capture_not_found")
	}
	if record.State == captureStateCancelled || record.State == captureStatePermanentFailed {
		return CaptureRecord{}, protocolError("companion_capture_media_rejected")
	}
	registry, err := s.registryForProfile(ctx, session.ProfileID)
	if err != nil {
		return CaptureRecord{}, err
	}
	fieldName = normaliseMediaField(fieldName)
	fieldAllowed := false
	for _, module := range registry.Modules {
		if module.ID != record.ModuleID || module.IntegrationInstanceID != record.IntegrationInstanceID {
			continue
		}
		for _, schema := range module.CaptureSchemas {
			if schema.PayloadType != record.PayloadType {
				continue
			}
			for _, allowed := range schema.MediaFields {
				if fieldName == normaliseMediaField(allowed) {
					fieldAllowed = true
				}
			}
		}
	}
	if !fieldAllowed {
		return CaptureRecord{}, protocolError("companion_capture_media_field_rejected")
	}
	return record, nil
}

func normaliseMediaField(value string) string {
	value = strings.TrimSpace(value)
	var result strings.Builder
	for index := 0; index < len(value); index++ {
		if value[index] == '[' {
			end := index + 1
			for end < len(value) && value[end] >= '0' && value[end] <= '9' {
				end++
			}
			if end > index+1 && end < len(value) && value[end] == ']' {
				index = end
				continue
			}
		}
		result.WriteByte(value[index])
	}
	return result.String()
}

func (s *Service) captureByID(ctx context.Context, profileID, id string) (CaptureRecord, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id, profile_id, module_id, module_version, schema_version, provider_id, integration_instance_id,
		payload_type, source_url, captured_at, page_complete, payload_hash, state, attempt_count, last_error,
		checkpoint_json, created_at, updated_at FROM companion_captures WHERE profile_id = ? AND id = ?`, profileID, id)
	return scanCapture(row)
}

func (s *Service) captureByIdempotency(ctx context.Context, profileID, key string) (CaptureRecord, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id, profile_id, module_id, module_version, schema_version, provider_id, integration_instance_id,
		payload_type, source_url, captured_at, page_complete, payload_hash, state, attempt_count, last_error,
		checkpoint_json, created_at, updated_at FROM companion_captures WHERE profile_id = ? AND idempotency_key = ?`, profileID, key)
	return scanCapture(row)
}

type rowScanner interface{ Scan(...any) error }

func scanCapture(row rowScanner) (CaptureRecord, error) {
	var record CaptureRecord
	var complete int
	var checkpoint string
	err := row.Scan(&record.ID, &record.ProfileID, &record.ModuleID, &record.ModuleVersion, &record.SchemaVersion,
		&record.ProviderID, &record.IntegrationInstanceID, &record.PayloadType, &record.SourceURL, &record.CapturedAt,
		&complete, &record.PayloadHash, &record.State, &record.AttemptCount, &record.LastError, &checkpoint,
		&record.CreatedAt, &record.UpdatedAt)
	if err != nil {
		return CaptureRecord{}, err
	}
	record.PageComplete = complete != 0
	record.Checkpoint = map[string]any{}
	_ = json.Unmarshal([]byte(checkpoint), &record.Checkpoint)
	return record, nil
}

func (s *Service) failCapture(ctx context.Context, captureID, state, code string) {
	_, _ = s.db.ExecContext(ctx, `UPDATE companion_captures SET state = ?, last_error = ?, updated_at = ? WHERE id = ? AND state != ?`, state, boundedCaptureText(code, 256), formatTime(s.options.Now().UTC()), captureID, captureStateCancelled)
}

func acceptedPayload(capture CaptureRecord, confidence float64, replayed bool) AcceptedPayload {
	return AcceptedPayload{
		Accepted: true, Committed: capture.State != captureStateAccepted && capture.State != captureStateProcessing,
		Replayed: replayed, CaptureID: capture.ID, ProfileID: capture.ProfileID, ModuleID: capture.ModuleID,
		PayloadType: capture.PayloadType, State: capture.State, SyncMode: SyncModePassiveCapture, RemoteWrite: false,
		Checkpoint: capture.Checkpoint, ConfidenceLabel: confidenceLabel(confidence),
		AuditTrail: []string{"companion_module=" + capture.ModuleID, "capture_id=" + capture.ID,
			"protocol_version=" + ProtocolVersionV1, "sync_mode=" + SyncModePassiveCapture, "remote_write=false"},
	}
}

func captureURLAllowed(module Module, rawURL string) bool {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed.Scheme != "https" || parsed.Hostname() == "" || parsed.User != nil || parsed.Fragment != "" {
		return false
	}
	for _, origin := range module.Browser.Origins {
		allowed, parseErr := url.Parse(strings.TrimSuffix(origin, "/*"))
		if parseErr == nil && parsed.Scheme == allowed.Scheme && parsed.Host == allowed.Host {
			return true
		}
	}
	return false
}

func sanitiseCaptureURL(rawURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return ""
	}
	parsed.User = nil
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String()
}

func redactionSummaryAllowed(rules, supplied []string) bool {
	allowed := map[string]struct{}{}
	for _, rule := range rules {
		allowed[rule] = struct{}{}
	}
	for _, rule := range supplied {
		if _, ok := allowed[rule]; !ok {
			return false
		}
	}
	return true
}

func hasForbiddenCaptureField(value any) bool {
	switch typed := value.(type) {
	case map[string]any:
		for key, nested := range typed {
			lower := strings.ToLower(key)
			if strings.Contains(lower, "secret") || strings.Contains(lower, "token") || strings.Contains(lower, "password") ||
				strings.Contains(lower, "cookie") || strings.Contains(lower, "authorization") || strings.Contains(lower, "credential") ||
				strings.Contains(lower, "raw_html") || strings.Contains(lower, "raw_page") || strings.Contains(lower, "raw_dom") {
				return true
			}
			if hasForbiddenCaptureField(nested) {
				return true
			}
		}
	case []any:
		for _, nested := range typed {
			if hasForbiddenCaptureField(nested) {
				return true
			}
		}
	}
	return false
}

func lowFieldConfidence(values map[string]float64) bool {
	for _, value := range values {
		if value < 0.8 {
			return true
		}
	}
	return false
}

func deterministicID(namespace string, values ...string) string {
	digest := sha256.Sum256([]byte(namespace + "\x00" + strings.Join(values, "\x00")))
	return namespace + "-" + hex.EncodeToString(digest[:16])
}

func boundedCaptureText(value string, maximum int) string {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > maximum || strings.ContainsAny(value, "\r\n\x00") {
		return ""
	}
	return value
}

func confidenceLabel(score float64) string {
	switch {
	case score >= 0.8:
		return "high"
	case score >= 0.5:
		return "medium"
	default:
		return "low"
	}
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
