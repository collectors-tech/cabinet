package matching

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/collectors-tech/cabinet/internal/db"
)

func TestClassifyCandidatesAndPersistConfidence(t *testing.T) {
	t.Parallel()

	conn, err := db.OpenAndMigrate(context.Background(), filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	if _, err := conn.Exec(`INSERT INTO canonical_items (id, brand, category, part_number, title) VALUES ('i1','AFX','Slot','P-1','AFX P-1')`); err != nil {
		t.Fatalf("seed item: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO scanner_query_sets (id, name, keywords_json, exclusions_json) VALUES ('q1','q','["afx"]','[]')`); err != nil {
		t.Fatalf("seed query set: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO scanner_candidates (id, query_set_id, listing_id, title, price, shipping, url, image, seller, first_seen, last_seen, status, source) VALUES ('c1','q1','L1','AFX P-1 rare',50,0,'http://x/1','','s1',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP,'new','ebay')`); err != nil {
		t.Fatalf("seed candidate c1: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO scanner_candidates (id, query_set_id, listing_id, title, price, shipping, url, image, seller, first_seen, last_seen, status, source) VALUES ('c2','q1','L2','Unknown title',20,0,'http://x/2','','s2',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP,'new','ebay')`); err != nil {
		t.Fatalf("seed candidate c2: %v", err)
	}

	svc := NewService(conn)
	if err := svc.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	out, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(out))
	}
	seenMatched := false
	seenNotInCollection := false
	for _, m := range out {
		if m.State == StateMatched && m.Confidence >= 0.95 && !m.NeedsReview {
			seenMatched = true
		}
		if m.State == StateNotInCollection && m.Confidence == 0 && m.NeedsReview {
			seenNotInCollection = true
		}
	}
	if !seenMatched || !seenNotInCollection {
		t.Fatalf("unexpected classification output: %+v", out)
	}
}
