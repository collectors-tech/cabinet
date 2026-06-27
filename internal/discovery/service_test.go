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
	var createdNotes, createdSourceURLs string
	if err := conn.QueryRow(`SELECT COUNT(*), COALESCE(MAX(notes), ''), COALESCE(MAX(source_urls_json), '') FROM canonical_items WHERE part_number = 'CREATE-901' AND title = 'AFX Create Candidate'`).Scan(&createdCount, &createdNotes, &createdSourceURLs); err != nil {
		t.Fatalf("load created item count: %v", err)
	}
	if createdCount != 1 {
		t.Fatalf("expected create item side effect, got %d", createdCount)
	}
	if !strings.Contains(createdNotes, `"source_provider":"ebay"`) || !strings.Contains(createdNotes, `"query_set_id":"q1"`) || !strings.Contains(createdNotes, `"query_name":"AFX saved search"`) {
		t.Fatalf("expected created inventory item notes to preserve saved-search provenance, got %q", createdNotes)
	}
	if !strings.Contains(createdSourceURLs, "https://example.test/listing") {
		t.Fatalf("expected created inventory item source URL to preserve listing URL, got %q", createdSourceURLs)
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

func TestDiscoveryCandidateContractIncludesStatusAndSourceResultAuditLink(t *testing.T) {
	t.Parallel()

	conn, err := db.OpenAndMigrate(context.Background(), filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	if _, err := conn.Exec(`INSERT INTO scanner_query_sets(id, name, keywords_json, exclusions_json, provider_scope_json) VALUES ('q1','Audit Query','["audit"]','[]','["ebay"]')`); err != nil {
		t.Fatalf("seed query set: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO scanner_candidates(
			id, query_set_id, listing_id, title, price, shipping, url, image, seller,
			first_seen, last_seen, status, source, observed_currency, reviewer_notes, source_result_url,
			stock_state, stock_count
		) VALUES (
			'c-audit','q1','listing-audit','Audit Candidate',19.95,0,'https://example.test/item','','seller-a',
			'2026-06-01T00:00:00Z','2026-06-02T00:00:00Z','reviewing','ebay','AUD','check photos','https://provider.test/result/listing-audit',
			'in_stock',4
		)
	`); err != nil {
		t.Fatalf("seed candidate: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO scanner_matches(candidate_id, item_id, state, confidence, needs_review, extracted_part_number, updated_at) VALUES ('c-audit','','not_in_collection',0.82,1,'AUD-001',CURRENT_TIMESTAMP)`); err != nil {
		t.Fatalf("seed match: %v", err)
	}

	items, err := NewService(conn).ListNotInCollection(context.Background(), Filter{})
	if err != nil {
		t.Fatalf("ListNotInCollection() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one item, got %d", len(items))
	}
	got := items[0]
	if got.SourceProvider != "ebay" || got.QuerySetID != "q1" || got.QueryName != "Audit Query" || got.ListingID != "listing-audit" {
		t.Fatalf("expected source/query/listing provenance, got %+v", got)
	}
	if got.ObservedCurrency != "AUD" || got.Seller != "seller-a" || got.FirstSeen == "" || got.LastSeen == "" {
		t.Fatalf("expected observed marketplace metadata, got %+v", got)
	}
	if got.Status != "reviewing" || got.Confidence != 0.82 || !got.NeedsReview || got.ReviewerNotes != "check photos" {
		t.Fatalf("expected review status/confidence/notes, got %+v", got)
	}
	if got.SourceResultURL != "https://provider.test/result/listing-audit" || got.ExtractedPart != "AUD-001" {
		t.Fatalf("expected source-result audit link and extracted part, got %+v", got)
	}
}

func TestListNotInCollectionDashboardFieldsAndArchivedOptIn(t *testing.T) {
	t.Parallel()

	conn, err := db.OpenAndMigrate(context.Background(), filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	if _, err := conn.Exec(`INSERT INTO provider_health(provider, status, message) VALUES ('market_watch', 'auth attention', 'reauth required')`); err != nil {
		t.Fatalf("seed provider health: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO canonical_items(id, brand, category, part_number, title, status) VALUES ('wish-item','AFX','Cars','WISH-1533','Wishlist Target','wishlist')`); err != nil {
		t.Fatalf("seed wishlist item: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO wishlist_entries(id, item_id, target_price, priority) VALUES ('wish-1533','wish-item',55,'high')`); err != nil {
		t.Fatalf("seed wishlist entry: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO price_snapshots(id, item_id, snapshot_date, source, median_price, latest_price) VALUES ('price-1533','wish-item','2026-06-25','market',70,68)`); err != nil {
		t.Fatalf("seed price snapshot: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO scanner_query_sets(id, name, keywords_json, exclusions_json, provider_scope_json) VALUES ('q1','AFX deals','["afx"]','[]','["market_watch"]')`); err != nil {
		t.Fatalf("seed query set: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO scanner_candidates(
			id, query_set_id, listing_id, title, price, shipping, url, image, seller,
			first_seen, last_seen, status, source, observed_currency, source_result_url,
			stock_state, stock_count
		) VALUES (
			'c-deal','q1','listing-deal','AFX Wishlist Deal',42,0,'https://example.test/deal','https://example.test/deal.jpg','Hobby store',
			'2026-06-20T00:00:00Z','2026-06-26T00:00:00Z','new','market_watch','AUD','https://provider.test/deal',
			'in_stock',2
		)
	`); err != nil {
		t.Fatalf("seed deal candidate: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO scanner_matches(candidate_id, item_id, state, confidence, needs_review, extracted_part_number, updated_at) VALUES ('c-deal','wish-item','not_in_collection',0.96,0,'WISH-1533',CURRENT_TIMESTAMP)`); err != nil {
		t.Fatalf("seed deal match: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO scanner_candidates(id, query_set_id, listing_id, title, price, shipping, url, seller, first_seen, last_seen, status, source, observed_currency, stock_state, stock_count)
		VALUES ('c-archived','q1','listing-archived','Archived Discovery',12,0,'https://example.test/archived','Seller','2026-06-18T00:00:00Z','2026-06-19T00:00:00Z','archived','market_watch','AUD','out_of_stock',0)
	`); err != nil {
		t.Fatalf("seed archived candidate: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO scanner_matches(candidate_id, item_id, state, confidence, needs_review, extracted_part_number, updated_at) VALUES ('c-archived','','not_in_collection',0.4,1,'ARCH-1533',CURRENT_TIMESTAMP)`); err != nil {
		t.Fatalf("seed archived match: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO ignored_candidates(candidate_id) VALUES ('c-archived')`); err != nil {
		t.Fatalf("seed ignored candidate: %v", err)
	}

	svc := NewService(conn)
	defaultItems, err := svc.ListNotInCollection(context.Background(), Filter{})
	if err != nil {
		t.Fatalf("ListNotInCollection(default) error = %v", err)
	}
	if len(defaultItems) != 1 || defaultItems[0].CandidateID != "c-deal" {
		t.Fatalf("expected default list to hide archived candidate, got %+v", defaultItems)
	}

	items, err := svc.ListNotInCollection(context.Background(), Filter{IncludeArchived: true})
	if err != nil {
		t.Fatalf("ListNotInCollection(include archived) error = %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected archived opt-in to return 2 items, got %d: %+v", len(items), items)
	}
	deal := items[0]
	if deal.CandidateID != "c-deal" {
		t.Fatalf("expected newest deal first, got %+v", deal)
	}
	if deal.Currency != "AUD" || deal.TriageStatus != "new" || deal.SellerLabel != "Hobby store" || deal.ThumbnailURL == "" {
		t.Fatalf("expected dashboard display aliases, got %+v", deal)
	}
	if deal.MatchType != "wishlist_match" || deal.MatchReason != "Wishlist match below target" {
		t.Fatalf("expected wishlist match fields, got %+v", deal)
	}
	if deal.WishlistID != "wish-1533" || deal.WishlistItemID != "wish-item" {
		t.Fatalf("expected wishlist linkage, got %+v", deal)
	}
	if deal.TargetPrice != 55 || deal.MarketBaseline != 70 || deal.PriceDeltaAmount != 13 || deal.PriceDeltaPct != 23.64 {
		t.Fatalf("expected target/baseline/delta fields, got %+v", deal)
	}
	if deal.DealScore <= 70 || deal.SourceTrust != "auth attention" || !strings.Contains(deal.Availability, "2 available") {
		t.Fatalf("expected deal score, trust, and availability, got %+v", deal)
	}
	var archived Item
	for _, item := range items {
		if item.CandidateID == "c-archived" {
			archived = item
		}
	}
	if archived.CandidateID == "" || archived.TriageStatus != "archived" {
		t.Fatalf("expected archived candidate in opt-in response, got %+v", items)
	}
}

func TestListNotInCollectionRanksWishlistDealsAndSuppressesHandledStatuses(t *testing.T) {
	t.Parallel()

	conn, err := db.OpenAndMigrate(context.Background(), filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	if _, err := conn.Exec(`INSERT INTO canonical_items(id, brand, category, part_number, title, status) VALUES ('wish-item','AFX','Cars','WISH-1547','Wishlist Target','wishlist')`); err != nil {
		t.Fatalf("seed wishlist item: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO wishlist_entries(id, item_id, target_price, priority) VALUES ('wish-1547','wish-item',80,'high')`); err != nil {
		t.Fatalf("seed wishlist entry: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO scanner_query_sets(id, name, keywords_json, exclusions_json, provider_scope_json) VALUES ('q1','AFX ranked deals','["afx"]','[]','["market_watch"]')`); err != nil {
		t.Fatalf("seed query set: %v", err)
	}
	seedCandidate := func(id, listingID, itemID, title, status, lastSeen string, price, confidence float64) {
		t.Helper()
		if _, err := conn.Exec(`
			INSERT INTO scanner_candidates(id, query_set_id, listing_id, title, price, shipping, url, image, seller, first_seen, last_seen, status, source, observed_currency, stock_state, stock_count)
			VALUES (?, 'q1', ?, ?, ?, 0, 'https://example.test/source', '', 'seller', '2026-06-20T00:00:00Z', ?, ?, 'market_watch', 'AUD', 'in_stock', 3)
		`, id, listingID, title, price, lastSeen, status); err != nil {
			t.Fatalf("seed candidate %s: %v", id, err)
		}
		if _, err := conn.Exec(`INSERT INTO scanner_matches(candidate_id, item_id, state, confidence, needs_review, extracted_part_number, updated_at) VALUES (?, ?, 'not_in_collection', ?, 0, ?, CURRENT_TIMESTAMP)`, id, itemID, confidence, listingID); err != nil {
			t.Fatalf("seed match %s: %v", id, err)
		}
	}
	seedCandidate("c-low-recent", "LOW-1547", "", "Recent low-signal result", "new", "2026-06-27T00:00:00Z", 20, 0.2)
	seedCandidate("c-deal-older", "DEAL-1547", "wish-item", "Older wishlist deal", "new", "2026-06-21T00:00:00Z", 40, 0.8)
	seedCandidate("c-promoted", "PROMOTED-1547", "wish-item", "Already promoted wishlist deal", "wishlisted", "2026-06-28T00:00:00Z", 35, 0.9)
	seedCandidate("c-archived-status", "ARCH-1547", "wish-item", "Archived status-only deal", "archived", "2026-06-29T00:00:00Z", 30, 0.9)

	svc := NewService(conn)
	items, err := svc.ListNotInCollection(context.Background(), Filter{})
	if err != nil {
		t.Fatalf("ListNotInCollection(default) error = %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("expected default list to hide archived status-only candidate, got %d: %+v", len(items), items)
	}
	if items[0].CandidateID != "c-deal-older" {
		t.Fatalf("expected wishlist deal to rank ahead of newer low-signal result, got %+v", items)
	}
	if items[0].DealScore <= items[1].DealScore || items[0].MatchType != "wishlist_match" || items[0].PriceDeltaAmount != 40 {
		t.Fatalf("expected deterministic wishlist deal score and metadata, got first=%+v second=%+v", items[0], items[1])
	}
	if items[2].CandidateID != "c-promoted" || items[2].DealScore != 0 || items[2].DestinationLink != "/wishlist/?item_id=wish-item" {
		t.Fatalf("expected already-promoted candidate to keep destination state without actionable deal ranking, got %+v", items[2])
	}

	withArchived, err := svc.ListNotInCollection(context.Background(), Filter{IncludeArchived: true})
	if err != nil {
		t.Fatalf("ListNotInCollection(include archived) error = %v", err)
	}
	var archived Item
	for _, item := range withArchived {
		if item.CandidateID == "c-archived-status" {
			archived = item
			break
		}
	}
	if archived.CandidateID == "" || archived.DealScore != 0 || archived.TriageStatus != "archived" {
		t.Fatalf("expected archived status-only candidate only through opt-in with no deal ranking, got %+v", withArchived)
	}
}

func TestApplyActionPersistsDestinationStatusesAndWishlistDoesNotClaimOwnership(t *testing.T) {
	t.Parallel()

	conn, err := db.OpenAndMigrate(context.Background(), filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	if _, err := conn.Exec(`INSERT INTO canonical_items(id, brand, category, part_number, title, status) VALUES ('i1','AFX','Cars','WISH-001','Wishlist Target','active')`); err != nil {
		t.Fatalf("seed wishlist target: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO scanner_query_sets(id, name, keywords_json, exclusions_json) VALUES ('q1','Destination Query','["dest"]','[]')`); err != nil {
		t.Fatalf("seed query set: %v", err)
	}
	seedCandidate := func(id, listingID, itemID, title, partNumber string) {
		t.Helper()
		if _, err := conn.Exec(`INSERT INTO scanner_candidates(id, query_set_id, listing_id, title, price, shipping, url, image, seller, first_seen, last_seen, status, source) VALUES (?, 'q1', ?, ?, 10, 0, 'https://example.test/source', '', 'seller', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 'new', 'ebay')`, id, listingID, title); err != nil {
			t.Fatalf("seed candidate %s: %v", id, err)
		}
		if _, err := conn.Exec(`INSERT INTO scanner_matches(candidate_id, item_id, state, confidence, needs_review, extracted_part_number, updated_at) VALUES (?, ?, 'not_in_collection', 0.7, 1, ?, CURRENT_TIMESTAMP)`, id, itemID, partNumber); err != nil {
			t.Fatalf("seed match %s: %v", id, err)
		}
	}
	seedCandidate("c-review", "L-review", "", "Review Candidate", "REV-001")
	seedCandidate("c-wishlist", "L-wishlist", "i1", "Wishlist Candidate", "WISH-001")
	seedCandidate("c-create-blocked", "L-create-blocked", "", "Blocked Create Candidate", "BLOCK-001")
	seedCandidate("c-create-owned", "L-create-owned", "", "Owned Create Candidate", "OWN-001")
	seedCandidate("c-archive", "L-archive", "", "Archive Candidate", "ARCH-001")

	svc := NewService(conn)
	actions := []Action{
		{CandidateID: "c-review", Type: ActionReview, Payload: map[string]any{"reviewer_notes": "needs manual pricing"}},
		{CandidateID: "c-wishlist", Type: ActionAddWishlist},
		{CandidateID: "c-create-blocked", Type: ActionCreateItem},
		{CandidateID: "c-create-owned", Type: ActionCreateItem, Payload: map[string]any{"ownership_confirmed": true}},
		{CandidateID: "c-archive", Type: ActionArchive, Payload: map[string]any{"notes": "duplicate source result"}},
	}
	for _, action := range actions {
		if err := svc.ApplyAction(context.Background(), action); err != nil {
			t.Fatalf("ApplyAction(%s/%s) error = %v", action.CandidateID, action.Type, err)
		}
	}

	expectedStatuses := map[string]string{
		"c-review":         "reviewing",
		"c-wishlist":       "wishlisted",
		"c-create-blocked": "inventory_candidate",
		"c-create-owned":   "inventory_candidate",
		"c-archive":        "archived",
	}
	for candidateID, want := range expectedStatuses {
		var got, sourceResultURL string
		if err := conn.QueryRow(`SELECT status, source_result_url FROM scanner_candidates WHERE id = ?`, candidateID).Scan(&got, &sourceResultURL); err != nil {
			t.Fatalf("load status for %s: %v", candidateID, err)
		}
		if got != want {
			t.Fatalf("candidate %s status=%q, want %q", candidateID, got, want)
		}
		if sourceResultURL != "https://example.test/source" {
			t.Fatalf("candidate %s source_result_url=%q", candidateID, sourceResultURL)
		}
	}

	var owned, quantity int
	if err := conn.QueryRow(`SELECT owned, quantity FROM wishlist_entries WHERE item_id = 'i1'`).Scan(&owned, &quantity); err != nil {
		t.Fatalf("load wishlist row: %v", err)
	}
	if owned != 0 || quantity != 0 {
		t.Fatalf("wishlist promotion must not claim ownership, owned=%d quantity=%d", owned, quantity)
	}
	var instanceCount int
	if err := conn.QueryRow(`SELECT COUNT(1) FROM instances WHERE item_id = 'i1'`).Scan(&instanceCount); err != nil {
		t.Fatalf("load instance count: %v", err)
	}
	if instanceCount != 0 {
		t.Fatalf("wishlist promotion must not create owned instances, got %d", instanceCount)
	}

	var blockedCreateCount, ownedCreateCount int
	if err := conn.QueryRow(`SELECT COUNT(1) FROM canonical_items WHERE part_number = 'BLOCK-001'`).Scan(&blockedCreateCount); err != nil {
		t.Fatalf("load blocked create count: %v", err)
	}
	if blockedCreateCount != 0 {
		t.Fatalf("unconfirmed inventory promotion created %d item(s)", blockedCreateCount)
	}
	if err := conn.QueryRow(`SELECT COUNT(1) FROM canonical_items WHERE part_number = 'OWN-001'`).Scan(&ownedCreateCount); err != nil {
		t.Fatalf("load owned create count: %v", err)
	}
	if ownedCreateCount != 1 {
		t.Fatalf("confirmed inventory promotion created %d item(s), want 1", ownedCreateCount)
	}
}

func TestReviewActionRestoresIgnoredCandidateForDefaultQueue(t *testing.T) {
	t.Parallel()

	conn, err := db.OpenAndMigrate(context.Background(), filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	if _, err := conn.Exec(`INSERT INTO scanner_query_sets(id, name, keywords_json, exclusions_json) VALUES ('q1','Restore Query','["restore"]','[]')`); err != nil {
		t.Fatalf("seed query set: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO scanner_candidates(id, query_set_id, listing_id, title, price, shipping, url, image, seller, first_seen, last_seen, status, source) VALUES ('c-restore','q1','RESTORE-001','Restore Candidate',25,0,'https://example.test/restore','','seller',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP,'ignored','ebay')`); err != nil {
		t.Fatalf("seed candidate: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO scanner_matches(candidate_id, item_id, state, confidence, needs_review, extracted_part_number, updated_at) VALUES ('c-restore','','not_in_collection',0.7,1,'RESTORE-001',CURRENT_TIMESTAMP)`); err != nil {
		t.Fatalf("seed match: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO ignored_candidates(candidate_id) VALUES ('c-restore')`); err != nil {
		t.Fatalf("seed ignored candidate: %v", err)
	}

	svc := NewService(conn)
	defaultItems, err := svc.ListNotInCollection(context.Background(), Filter{})
	if err != nil {
		t.Fatalf("ListNotInCollection(default) error = %v", err)
	}
	if len(defaultItems) != 0 {
		t.Fatalf("expected ignored candidate hidden by default, got %+v", defaultItems)
	}

	if err := svc.ApplyAction(context.Background(), Action{
		CandidateID: "c-restore",
		Type:        ActionReview,
		Payload: map[string]any{
			"reviewer_notes": "restore for follow-up",
		},
	}); err != nil {
		t.Fatalf("ApplyAction(review) error = %v", err)
	}

	restored, err := svc.ListNotInCollection(context.Background(), Filter{})
	if err != nil {
		t.Fatalf("ListNotInCollection(restored) error = %v", err)
	}
	if len(restored) != 1 || restored[0].CandidateID != "c-restore" || restored[0].TriageStatus != "reviewing" {
		t.Fatalf("expected restored reviewing candidate in default queue, got %+v", restored)
	}
	var ignoredCount int
	if err := conn.QueryRow(`SELECT COUNT(1) FROM ignored_candidates WHERE candidate_id = 'c-restore'`).Scan(&ignoredCount); err != nil {
		t.Fatalf("load ignored count: %v", err)
	}
	if ignoredCount != 0 {
		t.Fatalf("expected review restore to remove ignored marker, got %d", ignoredCount)
	}
}
