package app

import (
	"context"
	"encoding/json"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/collectors-tech/cabinet/internal/config"
	"github.com/collectors-tech/cabinet/internal/db"
	"github.com/collectors-tech/cabinet/internal/profile"
	"github.com/collectors-tech/cabinet/internal/update"
)

func TestStartupSampleDataBootstrapSeedsWishlistForActiveProfile(t *testing.T) {
	t.Setenv("CABINET_SEED_SAMPLE_DATA", "1")

	base := t.TempDir()
	cfg := config.Config{
		Addr:           "127.0.0.1:0",
		DataDir:        base,
		DBPath:         filepath.Join(base, "cabinet.db"),
		UpdateChannel:  update.ChannelStable,
		WebAuthnRPID:   "127.0.0.1",
		WebAuthnOrigin: "http://127.0.0.1:8080",
		WebAuthnName:   "Cabinet Test",
		BackupInterval: 60,
	}

	ctx := context.Background()
	conn, err := db.OpenAndMigrate(ctx, cfg.DBPath)
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	profiles := profile.NewRepository(conn)
	created, err := profiles.Create(ctx, "Showcase DB")
	if err != nil {
		t.Fatalf("Create() profile error = %v", err)
	}
	if err := profiles.SetActiveProfile(ctx, created.ID); err != nil {
		t.Fatalf("SetActiveProfile() error = %v", err)
	}
	if err := conn.Close(); err != nil {
		t.Fatalf("close seed connection: %v", err)
	}

	a := newTestAppWithConfig(t, cfg)
	resp := doRequest(t, a, http.MethodGet, "/api/wishlist", nil, nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("wishlist status=%d body=%s", resp.Code, resp.Body.String())
	}
	var payload struct {
		Items []struct {
			Priority       string `json:"priority"`
			BelowTargetNow bool   `json:"below_target_now"`
		} `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode wishlist payload: %v", err)
	}
	if len(payload.Items) < 3 {
		t.Fatalf("expected startup sample seed to create wishlist rows, got %+v", payload.Items)
	}
}

func TestStartupSampleDataBootstrapSeedsInventoryPriceHistory(t *testing.T) {
	t.Setenv("CABINET_SEED_SAMPLE_DATA", "1")

	base := t.TempDir()
	cfg := config.Config{
		Addr:           "127.0.0.1:0",
		DataDir:        base,
		DBPath:         filepath.Join(base, "cabinet.db"),
		UpdateChannel:  update.ChannelStable,
		WebAuthnRPID:   "127.0.0.1",
		WebAuthnOrigin: "http://127.0.0.1:8080",
		WebAuthnName:   "Cabinet Test",
		BackupInterval: 60,
	}

	ctx := context.Background()
	conn, err := db.OpenAndMigrate(ctx, cfg.DBPath)
	if err != nil {
		t.Fatalf("OpenAndMigrate() error = %v", err)
	}
	profiles := profile.NewRepository(conn)
	created, err := profiles.Create(ctx, "Showcase DB")
	if err != nil {
		t.Fatalf("Create() profile error = %v", err)
	}
	if err := profiles.SetActiveProfile(ctx, created.ID); err != nil {
		t.Fatalf("SetActiveProfile() error = %v", err)
	}
	if err := conn.Close(); err != nil {
		t.Fatalf("close seed connection: %v", err)
	}

	a := newTestAppWithConfig(t, cfg)
	itemsResp := doRequest(t, a, http.MethodGet, "/api/items", nil, nil)
	if itemsResp.Code != http.StatusOK {
		t.Fatalf("items status=%d body=%s", itemsResp.Code, itemsResp.Body.String())
	}
	var itemsPayload struct {
		Items []struct {
			ID         string `json:"id"`
			PartNumber string `json:"part_number"`
		} `json:"items"`
	}
	if err := json.NewDecoder(itemsResp.Body).Decode(&itemsPayload); err != nil {
		t.Fatalf("decode items payload: %v", err)
	}
	var itemID string
	for _, item := range itemsPayload.Items {
		if item.PartNumber == "CAB-DEMO-006" {
			itemID = item.ID
			break
		}
	}
	if itemID == "" {
		t.Fatalf("expected CAB-DEMO-006 in seeded showcase inventory items, got %+v", itemsPayload.Items)
	}

	historyResp := doRequest(t, a, http.MethodGet, "/api/pricing/history?item_id="+itemID, nil, nil)
	if historyResp.Code != http.StatusOK {
		t.Fatalf("pricing history status=%d body=%s", historyResp.Code, historyResp.Body.String())
	}
	var historyPayload struct {
		History []struct {
			SnapshotDate string  `json:"snapshot_date"`
			LatestPrice  float64 `json:"latest_price"`
		} `json:"history"`
	}
	if err := json.NewDecoder(historyResp.Body).Decode(&historyPayload); err != nil {
		t.Fatalf("decode pricing history payload: %v", err)
	}
	if len(historyPayload.History) < 4 {
		t.Fatalf("expected seeded showcase price history with at least 4 points, got %+v", historyPayload.History)
	}
}
