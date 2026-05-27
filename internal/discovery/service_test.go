package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/collectors-tech/cabinet/internal/db"
)

func TestListNotInCollectionAndActions(t *testing.T) {
	t.Parallel()

	conn, err := db.OpenAndMigrate(context.Background(), filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	if _, err := conn.Exec(`INSERT INTO scanner_query_sets(id, name, keywords_json, exclusions_json) VALUES ('q1','Q','["afx"]','[]')`); err != nil {
		t.Fatalf("seed query set: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO scanner_candidates(id, query_set_id, listing_id, title, price, shipping, url, image, seller, first_seen, last_seen, status, source, stock_state, stock_count) VALUES ('c1','q1','L1','AFX P-2',10,0,'http://x/1','','s',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP,'new','ebay','low_stock',2)`); err != nil {
		t.Fatalf("seed candidate: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO scanner_matches(candidate_id, item_id, state, confidence, needs_review, extracted_part_number, updated_at) VALUES ('c1','','not_in_collection',0,1,'P-2',CURRENT_TIMESTAMP)`); err != nil {
		t.Fatalf("seed match: %v", err)
	}

	svc := NewService(conn)
	items, err := svc.ListNotInCollection(context.Background(), Filter{})
	if err != nil {
		t.Fatalf("ListNotInCollection() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].StockState != "low_stock" || items[0].StockCount != 2 {
		t.Fatalf("expected stock fields in discovery item, got %+v", items[0])
	}

	if err := svc.ApplyAction(context.Background(), Action{
		CandidateID: "c1",
		Type:        ActionIgnore,
	}); err != nil {
		t.Fatalf("ApplyAction(ignore) error = %v", err)
	}
	afterIgnore, err := svc.ListNotInCollection(context.Background(), Filter{})
	if err != nil {
		t.Fatalf("ListNotInCollection() after ignore error = %v", err)
	}
	if len(afterIgnore) != 0 {
		t.Fatalf("expected 0 items after ignore, got %d", len(afterIgnore))
	}
	if err := svc.ResetIgnored(context.Background()); err != nil {
		t.Fatalf("ResetIgnored() error = %v", err)
	}
	afterReset, err := svc.ListNotInCollection(context.Background(), Filter{})
	if err != nil {
		t.Fatalf("ListNotInCollection() after reset error = %v", err)
	}
	if len(afterReset) != 1 {
		t.Fatalf("expected 1 item after reset, got %d", len(afterReset))
	}
}

func TestApplyActionAddWishlistRetainsMarketplaceMetadata(t *testing.T) {
	t.Parallel()

	conn, err := db.OpenAndMigrate(context.Background(), filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	if _, err := conn.Exec(`INSERT INTO canonical_items(id, brand, category, part_number, title) VALUES ('i1','AFX','Cars','E2E-PN-900','Seed Item')`); err != nil {
		t.Fatalf("seed item: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO scanner_query_sets(id, name, keywords_json, exclusions_json, provider_scope_json) VALUES ('q1','AFX saved search','["afx"]','[]','["ebay","bonzaslotcars"]')`); err != nil {
		t.Fatalf("seed query set: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO scanner_candidates(id, query_set_id, listing_id, title, price, shipping, url, image, seller, first_seen, last_seen, status, source, stock_state, stock_count) VALUES ('c1','q1','L1','AFX P-2',44.95,0,'https://example.test/listing','','seller-1',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP,'new','ebay','low_stock',2)`); err != nil {
		t.Fatalf("seed candidate: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO scanner_matches(candidate_id, item_id, state, confidence, needs_review, extracted_part_number, updated_at) VALUES ('c1','i1','not_in_collection',0.9,0,'E2E-PN-900',CURRENT_TIMESTAMP)`); err != nil {
		t.Fatalf("seed match: %v", err)
	}

	svc := NewService(conn)
	if err := svc.ApplyAction(context.Background(), Action{
		CandidateID: "c1",
		Type:        ActionAddWishlist,
		Payload: map[string]any{
			"source": "market_watch",
		},
	}); err != nil {
		t.Fatalf("ApplyAction(add_to_wishlist) error = %v", err)
	}

	var notes string
	if err := conn.QueryRow(`SELECT notes FROM wishlist_entries WHERE item_id = 'i1'`).Scan(&notes); err != nil {
		t.Fatalf("load wishlist notes: %v", err)
	}
	marker := "[discovery_metadata]"
	idx := strings.Index(notes, marker)
	if idx < 0 {
		t.Fatalf("expected discovery metadata marker in notes, got %q", notes)
	}
	metadataJSON := strings.TrimSpace(notes[idx+len(marker):])
	var metadata map[string]any
	if err := json.Unmarshal([]byte(metadataJSON), &metadata); err != nil {
		t.Fatalf("parse metadata json: %v", err)
	}
	for _, field := range []string{"listing_url", "seller", "stock_signal", "observed_price", "source_provider", "query_set_id", "query_name", "provider_scope"} {
		if _, ok := metadata[field]; !ok {
			t.Fatalf("expected metadata field %q in %v", field, metadata)
		}
	}
	if metadata["source_provider"] != "ebay" {
		t.Fatalf("expected source_provider ebay, got %v", metadata["source_provider"])
	}
	if metadata["query_set_id"] != "q1" || metadata["query_name"] != "AFX saved search" {
		t.Fatalf("expected query set provenance, got %v", metadata)
	}
	providerScope, ok := metadata["provider_scope"].([]any)
	if !ok || len(providerScope) != 2 || providerScope[0] != "ebay" || providerScope[1] != "bonzaslotcars" {
		t.Fatalf("expected provider scope provenance, got %v", metadata["provider_scope"])
	}

	var actionPayloadJSON string
	if err := conn.QueryRow(`SELECT payload_json FROM discovery_actions WHERE candidate_id = 'c1' AND action_type = 'add_to_wishlist'`).Scan(&actionPayloadJSON); err != nil {
		t.Fatalf("load discovery action payload: %v", err)
	}
	var actionPayload map[string]any
	if err := json.Unmarshal([]byte(actionPayloadJSON), &actionPayload); err != nil {
		t.Fatalf("parse action payload json: %v", err)
	}
	if actionPayload["source"] != "market_watch" {
		t.Fatalf("expected existing action payload source to survive, got %v", actionPayload)
	}
	for _, field := range []string{"source_provider", "query_set_id", "query_name", "provider_scope"} {
		if _, ok := actionPayload[field]; !ok {
			t.Fatalf("expected action payload audit field %q in %v", field, actionPayload)
		}
	}
	if actionPayload["source_provider"] != "ebay" || actionPayload["query_set_id"] != "q1" || actionPayload["query_name"] != "AFX saved search" {
		t.Fatalf("expected saved-search provenance in action payload, got %v", actionPayload)
	}
}

func TestApplySavedSearchActionsRetainAuditProvenance(t *testing.T) {
	t.Parallel()

	conn, err := db.OpenAndMigrate(context.Background(), filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	if _, err := conn.Exec(`INSERT INTO canonical_items(id, brand, category, part_number, title) VALUES ('i1','AFX','Cars','E2E-PN-901','Seed Item')`); err != nil {
		t.Fatalf("seed item: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO scanner_query_sets(id, name, keywords_json, exclusions_json, provider_scope_json) VALUES ('q1','AFX saved search','["afx"]','[]','["ebay","amazon","bonzaslotcars"]')`); err != nil {
		t.Fatalf("seed query set: %v", err)
	}
	seedCandidate := func(id, listingID, itemID, title, partNumber string) {
		t.Helper()
		if _, err := conn.Exec(`INSERT INTO scanner_candidates(id, query_set_id, listing_id, title, price, shipping, url, image, seller, first_seen, last_seen, status, source, stock_state, stock_count) VALUES (?, 'q1', ?, ?, 44.95, 0, 'https://example.test/listing', '', 'seller-1', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'new', 'ebay', 'in_stock', 7)`, id, listingID, title); err != nil {
			t.Fatalf("seed candidate %s: %v", id, err)
		}
		if _, err := conn.Exec(`INSERT INTO scanner_matches(candidate_id, item_id, state, confidence, needs_review, extracted_part_number, updated_at) VALUES (?, ?, 'not_in_collection', 0.9, 0, ?, CURRENT_TIMESTAMP)`, id, itemID, partNumber); err != nil {
			t.Fatalf("seed match %s: %v", id, err)
		}
	}
	seedCandidate("c-ignore", "L-ignore", "", "AFX Ignore Candidate", "IGNORE-901")
	seedCandidate("c-track", "L-track", "i1", "AFX Track Candidate", "TRACK-901")
	seedCandidate("c-create", "L-create", "", "AFX Create Candidate", "CREATE-901")

	svc := NewService(conn)
	for _, action := range []Action{
		{CandidateID: "c-ignore", Type: ActionIgnore, Payload: map[string]any{"decision": "not relevant"}},
		{CandidateID: "c-track", Type: ActionTrackPrice, Payload: map[string]any{"decision": "monitor"}},
		{CandidateID: "c-create", Type: ActionCreateItem, Payload: map[string]any{"decision": "create owned item"}},
	} {
		if err := svc.ApplyAction(context.Background(), action); err != nil {
			t.Fatalf("ApplyAction(%s) error = %v", action.Type, err)
		}
	}

	var ignoredCount int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM ignored_candidates WHERE candidate_id = 'c-ignore'`).Scan(&ignoredCount); err != nil {
		t.Fatalf("load ignored candidate count: %v", err)
	}
	if ignoredCount != 1 {
		t.Fatalf("expected ignore action side effect, got %d", ignoredCount)
	}
	var trackedCount int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM tracked_items WHERE item_id = 'i1'`).Scan(&trackedCount); err != nil {
		t.Fatalf("load tracked item count: %v", err)
	}
	if trackedCount != 1 {
		t.Fatalf("expected track action side effect, got %d", trackedCount)
	}
	var createdCount int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM canonical_items WHERE part_number = 'CREATE-901' AND title = 'AFX Create Candidate'`).Scan(&createdCount); err != nil {
		t.Fatalf("load created item count: %v", err)
	}
	if createdCount != 1 {
		t.Fatalf("expected create item side effect, got %d", createdCount)
	}

	rows, err := conn.Query(`SELECT action_type, payload_json FROM discovery_actions WHERE candidate_id IN ('c-ignore', 'c-track', 'c-create')`)
	if err != nil {
		t.Fatalf("load discovery action payloads: %v", err)
	}
	defer rows.Close()

	seen := map[string]bool{}
	for rows.Next() {
		var actionType, raw string
		if err := rows.Scan(&actionType, &raw); err != nil {
			t.Fatalf("scan discovery action payload: %v", err)
		}
		var payload map[string]any
		if err := json.Unmarshal([]byte(raw), &payload); err != nil {
			t.Fatalf("parse %s payload json: %v", actionType, err)
		}
		if payload["source_provider"] != "ebay" || payload["query_set_id"] != "q1" || payload["query_name"] != "AFX saved search" {
			t.Fatalf("expected saved-search provenance in %s payload, got %v", actionType, payload)
		}
		providerScope, ok := payload["provider_scope"].([]any)
		if !ok || len(providerScope) != 3 || providerScope[0] != "ebay" || providerScope[1] != "amazon" || providerScope[2] != "bonzaslotcars" {
			t.Fatalf("expected provider scope in %s payload, got %v", actionType, payload["provider_scope"])
		}
		if strings.TrimSpace(fmt.Sprint(payload["decision"])) == "" {
			t.Fatalf("expected existing decision payload to survive in %s payload, got %v", actionType, payload)
		}
		seen[actionType] = true
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate discovery action payloads: %v", err)
	}
	for _, actionType := range []string{string(ActionIgnore), string(ActionTrackPrice), string(ActionCreateItem)} {
		if !seen[actionType] {
			t.Fatalf("expected action payload for %s, got %v", actionType, seen)
		}
	}
}
