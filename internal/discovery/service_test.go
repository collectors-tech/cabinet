package discovery

import (
	"context"
	"encoding/json"
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
	if err := svc.ApplyAction(context.Background(), Action{CandidateID: "c1", Type: ActionAddWishlist}); err != nil {
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
}
