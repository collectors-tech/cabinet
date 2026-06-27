package app

import (
	"context"
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/collectors-tech/cabinet/internal/config"
	"github.com/collectors-tech/cabinet/internal/db"
	"github.com/collectors-tech/cabinet/internal/profile"
	"github.com/collectors-tech/cabinet/internal/update"
)

func TestStartupSampleDataBootstrapSeedsShowcaseArtifacts(t *testing.T) {
	t.Setenv("CABINET_SEED_SAMPLE_DATA", "1")
	t.Setenv("CABINET_STARTUP_TIMEOUT_SECONDS", "180")

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
	assertStartupSampleWishlistSeeded(t, a)
	assertStartupSamplePurchasesSeeded(t, a)
	assertStartupSamplePriceHistorySeeded(t, a)
	assertStartupSamplePhotosSeeded(t, a)
}

func assertStartupSampleWishlistSeeded(t *testing.T, a *App) {
	t.Helper()

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

func assertStartupSamplePurchasesSeeded(t *testing.T, a *App) {
	t.Helper()

	resp := doRequest(t, a, http.MethodGet, "/api/commerce/purchase-orders?status=all&page=1&page_size=10", nil, nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("purchase orders status=%d body=%s", resp.Code, resp.Body.String())
	}
	var payload struct {
		Total  int `json:"total"`
		Orders []struct {
			OrderID         string `json:"order_id"`
			Source          string `json:"source"`
			Seller          string `json:"seller"`
			Tracking        string `json:"tracking"`
			Status          string `json:"status"`
			LineItemCount   int    `json:"line_item_count"`
			ReceivedCount   int    `json:"received_count"`
			UnreceivedCount int    `json:"unreceived_count"`
			LineItems       []struct {
				Title  string `json:"title"`
				Status string `json:"status"`
			} `json:"line_items"`
		} `json:"orders"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode purchase order payload: %v", err)
	}
	if payload.Total < 3 {
		t.Fatalf("expected at least three seeded sample purchase orders, got %+v", payload.Orders)
	}

	var foundMultiLinePartial, foundReceived, foundShipped bool
	for _, order := range payload.Orders {
		switch order.OrderID {
		case "EBAY-SAMPLE-1001":
			foundMultiLinePartial = order.Source == "ebay" &&
				order.Seller == "seller-one" &&
				order.Tracking == "TRACK-SAMPLE-1001" &&
				order.LineItemCount == 2 &&
				order.ReceivedCount == 1 &&
				order.UnreceivedCount == 1 &&
				order.Status == "active"
		case "MANUAL-SAMPLE-2001":
			foundReceived = order.Source == "manual" &&
				order.ReceivedCount == 1 &&
				order.UnreceivedCount == 0 &&
				order.Status == "received"
		case "SHOP-SAMPLE-3001":
			foundShipped = order.Source == "storefront" &&
				order.Tracking == "AUSPOST-SAMPLE-3001" &&
				order.UnreceivedCount == 1
		}
	}
	if !foundMultiLinePartial {
		t.Fatalf("expected seeded eBay multi-line partial purchase order, got %+v", payload.Orders)
	}
	if !foundReceived {
		t.Fatalf("expected seeded manual fully received purchase order, got %+v", payload.Orders)
	}
	if !foundShipped {
		t.Fatalf("expected seeded shipped/unreceived purchase order, got %+v", payload.Orders)
	}

	reviews := doRequest(t, a, http.MethodGet, "/api/commerce/purchase-orders?status=reviews", nil, nil)
	if reviews.Code != http.StatusOK || !strings.Contains(reviews.Body.String(), "EBAY-SAMPLE-1001") {
		t.Fatalf("expected review filter to include partial sample purchase, status=%d body=%s", reviews.Code, reviews.Body.String())
	}
	received := doRequest(t, a, http.MethodGet, "/api/commerce/purchase-orders?status=received", nil, nil)
	if received.Code != http.StatusOK || !strings.Contains(received.Body.String(), "MANUAL-SAMPLE-2001") {
		t.Fatalf("expected received filter to include manual received sample, status=%d body=%s", received.Code, received.Body.String())
	}
}

func assertStartupSamplePriceHistorySeeded(t *testing.T, a *App) {
	t.Helper()

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

func assertStartupSamplePhotosSeeded(t *testing.T, a *App) {
	t.Helper()

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
	if len(itemsPayload.Items) < 10 {
		t.Fatalf("expected seeded showcase inventory items, got %+v", itemsPayload.Items)
	}

	missingPhotos := []string{}
	multiPhotoItems := 0
	for _, item := range itemsPayload.Items {
		photosResp := doRequest(t, a, http.MethodGet, "/api/items/"+item.ID+"/photos", nil, nil)
		if photosResp.Code != http.StatusOK {
			t.Fatalf("photos status=%d for %s body=%s", photosResp.Code, item.PartNumber, photosResp.Body.String())
		}
		var photosPayload struct {
			Photos []struct {
				ID           string `json:"id"`
				Filename     string `json:"filename"`
				IsPrimary    bool   `json:"is_primary"`
				DisplayOrder int    `json:"display_order"`
			} `json:"photos"`
		}
		if err := json.NewDecoder(photosResp.Body).Decode(&photosPayload); err != nil {
			t.Fatalf("decode photos payload for %s: %v", item.PartNumber, err)
		}
		if len(photosPayload.Photos) == 0 {
			missingPhotos = append(missingPhotos, item.PartNumber)
			continue
		}
		if len(photosPayload.Photos) > 1 {
			multiPhotoItems++
		}
		if !photosPayload.Photos[0].IsPrimary || photosPayload.Photos[0].DisplayOrder != 1 {
			t.Fatalf("expected first photo for %s to be primary display order 1, got %+v", item.PartNumber, photosPayload.Photos[0])
		}
		fileResp := doRequest(t, a, http.MethodGet, "/api/items/"+item.ID+"/photos/"+photosPayload.Photos[0].ID+"/file?variant=thumbnail", nil, nil)
		if fileResp.Code != http.StatusOK {
			t.Fatalf("thumbnail status=%d for %s body=%s", fileResp.Code, item.PartNumber, fileResp.Body.String())
		}
	}
	if len(missingPhotos) > 0 {
		t.Fatalf("expected every seeded showcase item to have generated photos, missing=%v", missingPhotos)
	}
	if multiPhotoItems == 0 {
		t.Fatalf("expected at least one seeded showcase item with multiple generated photos")
	}
}
