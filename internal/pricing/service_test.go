package pricing

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/collectors-tech/cabinet/internal/db"
)

func TestDailySnapshotAndExport(t *testing.T) {
	t.Parallel()

	conn, err := db.OpenAndMigrate(context.Background(), filepath.Join(t.TempDir(), "cabinet.db"))
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if _, err := conn.Exec(`INSERT INTO canonical_items(id, brand, category, part_number, title) VALUES ('i1','AFX','Slot','P-1','AFX P-1')`); err != nil {
		t.Fatalf("seed item: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO scanner_query_sets(id, name, keywords_json, exclusions_json) VALUES ('q1','Q','["afx"]','[]')`); err != nil {
		t.Fatalf("seed query set: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO scanner_candidates(id, query_set_id, listing_id, title, price, shipping, url, image, seller, first_seen, last_seen, status, source) VALUES 
		('c1','q1','L1','AFX P-1',10,0,'http://x/1','','s',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP,'new','ebay'),
		('c2','q1','L2','AFX P-1',20,0,'http://x/2','','s',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP,'new','ebay'),
		('c3','q1','L3','AFX P-1',30,0,'http://x/3','','s',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP,'new','ebay')
	`); err != nil {
		t.Fatalf("seed candidates: %v", err)
	}

	svc := NewService(conn)
	if err := svc.TrackItem(context.Background(), "i1"); err != nil {
		t.Fatalf("TrackItem() error = %v", err)
	}
	if err := svc.RunDailySnapshot(context.Background()); err != nil {
		t.Fatalf("RunDailySnapshot() error = %v", err)
	}
	history, err := svc.History(context.Background(), "i1")
	if err != nil {
		t.Fatalf("History() error = %v", err)
	}
	if len(history) == 0 {
		t.Fatal("expected snapshot history")
	}
	if history[0].MinPrice != 10 || history[0].MedianPrice != 20 || history[0].LatestPrice != 30 {
		t.Fatalf("unexpected snapshot values: %+v", history[0])
	}
	bySource, err := svc.BySource(context.Background(), "i1")
	if err != nil {
		t.Fatalf("BySource() error = %v", err)
	}
	if len(bySource) == 0 {
		t.Fatal("expected source breakdown")
	}
	csv, err := svc.ExportCSV(context.Background(), "i1")
	if err != nil {
		t.Fatalf("ExportCSV() error = %v", err)
	}
	if !strings.Contains(csv, "snapshot_date,min_price,median_price,latest_price,source") {
		t.Fatalf("unexpected csv export: %s", csv)
	}
	trend, err := svc.Trend(context.Background(), "i1")
	if err != nil {
		t.Fatalf("Trend() error = %v", err)
	}
	if len(trend) == 0 || trend[0].Latest <= 0 {
		t.Fatalf("unexpected trend output: %+v", trend)
	}
}
