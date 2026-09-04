package app

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/collectors-tech/cabinet/internal/collection"
	"github.com/collectors-tech/cabinet/internal/media"
	"github.com/collectors-tech/cabinet/internal/profile"
	"github.com/collectors-tech/cabinet/internal/wishlist"
)

const inventoryFolderItemAssignmentsSettingsKey = "inventory.folder-item-assignments.v1"

type onboardingSampleSeedResult struct {
	DatasetKind             string `json:"dataset_kind"`
	DatasetLabel            string `json:"dataset_label"`
	SampleDataDisclosure    string `json:"sample_data_disclosure"`
	CreatedItems            int    `json:"created_items"`
	CreatedInstances        int    `json:"created_instances"`
	CreatedPhotos           int    `json:"created_photos"`
	CreatedWishlistEntries  int    `json:"created_wishlist_entries"`
	CreatedPurchaseOrders   int    `json:"created_purchase_orders"`
	TotalPurchaseOrders     int    `json:"total_purchase_orders"`
	TotalItems              int    `json:"total_items"`
	TotalWishlistEntries    int    `json:"total_wishlist_entries"`
	AlreadySeededForProfile bool   `json:"already_seeded_for_profile"`
}

type onboardingSampleSpec struct {
	Item                   collection.Item
	Instance               collection.Instance
	PriceHistory           []onboardingPriceSnapshot
	Wishlist               bool
	TargetPrice            float64
	WishlistPriority       string
	WishlistNotes          string
	WishlistHighlightHit   bool
	WishlistBelowTargetNow bool
	FolderName             string
}

type onboardingPriceSnapshot struct {
	SnapshotDate string
	Source       string
	MinPrice     float64
	MedianPrice  float64
	LatestPrice  float64
	StockCount   int
}

type onboardingPurchaseSampleLine struct {
	EntryID        string
	ArrivalID      string
	ItemPartNumber string
	OrderID        string
	Source         string
	Seller         string
	Tracking       string
	Quantity       int
	Amount         float64
	Status         string
	ExpectedOn     string
	DeliveredOn    string
	Notes          string
}

func generatedShowcasePriceHistory(sequence int, anchorPrice float64) []onboardingPriceSnapshot {
	if anchorPrice <= 0 {
		anchorPrice = float64(10 + sequence%50)
	}
	base := anchorPrice + float64(sequence%7)
	month := 1 + (sequence % 3)
	day := 2 + (sequence % 20)
	source := "showcase-market"
	if sequence%2 == 0 {
		source = "ebay"
	}
	return []onboardingPriceSnapshot{
		{
			SnapshotDate: fmt.Sprintf("2026-%02d-%02d", month, day),
			Source:       source,
			MinPrice:     base - 3.5,
			MedianPrice:  base,
			LatestPrice:  base - 1.25,
			StockCount:   8 + (sequence % 5),
		},
		{
			SnapshotDate: fmt.Sprintf("2026-%02d-%02d", month, day+2),
			Source:       source,
			MinPrice:     base - 2.5,
			MedianPrice:  base + 1.75,
			LatestPrice:  base + 2.25,
			StockCount:   6 + (sequence % 6),
		},
		{
			SnapshotDate: fmt.Sprintf("2026-%02d-%02d", month, day+4),
			Source:       "collector-index",
			MinPrice:     base - 1,
			MedianPrice:  base + 2.5,
			LatestPrice:  base + 1.5,
			StockCount:   10 + (sequence % 4),
		},
		{
			SnapshotDate: fmt.Sprintf("2026-%02d-%02d", month, day+6),
			Source:       source,
			MinPrice:     base,
			MedianPrice:  base + 3,
			LatestPrice:  base + 4.75,
			StockCount:   5 + (sequence % 7),
		},
	}
}

func generatedShowcaseInventorySpecs() []onboardingSampleSpec {
	type folderPlan struct {
		Name  string
		Count int
	}

	plans := []folderPlan{
		{Name: "Watch List", Count: 4},
		{Name: "Wishlist Focus", Count: 3},
		{Name: "Store 1", Count: 5},
		{Name: "Store 2", Count: 3},
		{Name: "Store 3", Count: 6},
		{Name: "Store 4", Count: 4},
		{Name: "Store 5", Count: 7},
		{Name: "Store 6", Count: 2},
		{Name: "Store 7", Count: 5},
		{Name: "Store 8", Count: 3},
		{Name: "Store 9", Count: 6},
		{Name: "Store 10", Count: 2},
		{Name: "Warehouse 1", Count: 5},
		{Name: "Warehouse 2", Count: 6},
		{Name: "Warehouse 3", Count: 4},
		{Name: "Archive A", Count: 4},
		{Name: "Archive B", Count: 2},
		{Name: "Archive C", Count: 1},
		{Name: "Archive D", Count: 5},
		{Name: "Archive E", Count: 2},
		{Name: "Archive F", Count: 3},
		{Name: "Archive G", Count: 3},
		{Name: "Archive H", Count: 2},
		{Name: "Archive I", Count: 4},
		{Name: "Archive J", Count: 2},
		{Name: "Archive K", Count: 3},
	}
	categories := []string{
		"Diecast",
		"Slot Car",
		"Trading Card",
		"Action Figure",
		"Comic",
		"Model Kit",
		"Video Game",
		"Vinyl",
	}
	brands := []string{
		"Hot Wheels",
		"Matchbox",
		"Pokemon",
		"Marvel",
		"Bandai",
		"Nintendo",
		"Hasbro",
		"Cabinet",
	}
	conditions := []string{"Mint", "Near Mint", "Excellent", "Very Good", "Good"}
	statuses := []string{"sealed", "blister", "loose", "custom", "on_track"}

	specs := make([]onboardingSampleSpec, 0, 96)
	sequence := 1
	for _, plan := range plans {
		for index := 1; index <= plan.Count; index++ {
			category := categories[(sequence-1)%len(categories)]
			brand := brands[(sequence-1)%len(brands)]
			condition := conditions[(sequence-1)%len(conditions)]
			status := statuses[(sequence-1)%len(statuses)]
			specs = append(specs, onboardingSampleSpec{
				Item: collection.Item{
					Brand:       brand,
					Category:    category,
					PartNumber:  fmt.Sprintf("CAB-SHOW-%03d", sequence),
					Title:       fmt.Sprintf("%s Showcase Item %02d", plan.Name, index),
					Make:        brand,
					Model:       fmt.Sprintf("%s Sample %02d", category, index),
					Year:        fmt.Sprintf("%d", 1980+(sequence%45)),
					Scale:       "Sample",
					Series:      "Showcase Inventory",
					Description: fmt.Sprintf("Distinct showcase inventory item assigned only to %s.", plan.Name),
					Tags:        []string{"showcase", "sample", "inventory"},
				},
				Instance: collection.Instance{
					Condition:        condition,
					Status:           status,
					Quantity:         1 + (sequence % 3),
					StorageLocation:  fmt.Sprintf("%s Bin %02d", plan.Name, index),
					AcquisitionPrice: float64(5 + (sequence % 60)),
					AcquisitionDate:  fmt.Sprintf("2026-02-%02d", 1+(sequence%27)),
					Notes:            "Generated showcase inventory row with exclusive folder membership.",
				},
				PriceHistory: generatedShowcasePriceHistory(sequence, float64(5+(sequence%60))),
				FolderName:   plan.Name,
			})
			sequence++
		}
	}

	return specs
}

func showcasePhotoTargetCount(item collection.Item) int {
	seed := sha256.Sum256([]byte(item.PartNumber + "|" + item.Title))
	switch {
	case seed[0]%17 == 0:
		return 3
	case seed[0]%4 == 0:
		return 2
	default:
		return 1
	}
}

func generateShowcaseIdenticonPNG(item collection.Item, variant int) ([]byte, error) {
	seed := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|%d", item.PartNumber, item.Title, variant)))
	// These generated assets are shown as compact inventory thumbnails. Keeping
	// the source at that display scale avoids encoding two oversized renditions
	// for every row while onboarding is waiting for the synchronous response.
	img := image.NewRGBA(image.Rect(0, 0, 128, 128))

	bg := color.RGBA{
		R: 16 + seed[1]%28,
		G: 22 + seed[2]%34,
		B: 38 + seed[3]%46,
		A: 255,
	}
	accent := color.RGBA{
		R: 72 + seed[4]%150,
		G: 88 + seed[5]%150,
		B: 118 + seed[6]%130,
		A: 255,
	}
	highlight := color.RGBA{
		R: 186 + seed[7]%58,
		G: 200 + seed[8]%42,
		B: 220 + seed[9]%34,
		A: 255,
	}

	draw.Draw(img, img.Bounds(), &image.Uniform{C: bg}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(7, 7, 121, 121), &image.Uniform{C: color.RGBA{R: bg.R + 6, G: bg.G + 8, B: bg.B + 10, A: 255}}, image.Point{}, draw.Src)

	const cell = 18
	const gap = 2
	const origin = 19
	for y := 0; y < 5; y++ {
		for x := 0; x < 3; x++ {
			if seed[10+y*3+x]%2 == 0 {
				continue
			}
			for _, mirroredX := range []int{x, 4 - x} {
				left := origin + mirroredX*(cell+gap)
				top := origin + y*(cell+gap)
				fill := accent
				if (x+y+variant)%3 == 0 {
					fill = highlight
				}
				draw.Draw(img, image.Rect(left, top, left+cell, top+cell), &image.Uniform{C: fill}, image.Point{}, draw.Src)
			}
		}
	}

	var out bytes.Buffer
	if err := png.Encode(&out, img); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func seedShowcaseItemPhotos(ctx context.Context, tx *sql.Tx, mediaSvc *media.Service, item collection.Item, createdAssetDirs *[]string) (int, error) {
	if mediaSvc == nil {
		return 0, nil
	}
	existingCount, err := mediaSvc.CountByItemTx(ctx, tx, item.ID)
	if err != nil {
		return 0, fmt.Errorf("list photos for %s: %w", item.ID, err)
	}

	target := showcasePhotoTargetCount(item)
	created := 0
	for variant := existingCount + 1; variant <= target; variant++ {
		data, err := generateShowcaseIdenticonPNG(item, variant)
		if err != nil {
			return created, fmt.Errorf("generate showcase photo for %s: %w", item.PartNumber, err)
		}
		filename := fmt.Sprintf("%s-identicon-%02d.png", item.PartNumber, variant)
		photo, err := mediaSvc.UploadTx(ctx, tx, item.ID, filename, bytes.NewReader(data))
		if err != nil {
			return created, fmt.Errorf("upload showcase photo for %s: %w", item.PartNumber, err)
		}
		*createdAssetDirs = append(*createdAssetDirs, filepath.Dir(filepath.Dir(photo.OriginalPath)))
		created++
	}
	return created, nil
}

func onboardingPurchaseSamples() []onboardingPurchaseSampleLine {
	return []onboardingPurchaseSampleLine{
		{
			EntryID:        "sample-purchase-life-ebay-watch-001",
			ArrivalID:      "sample-purchase-arrival-ebay-watch-001",
			ItemPartNumber: "CAB-DEMO-001",
			OrderID:        "EBAY-SAMPLE-1001",
			Source:         "ebay",
			Seller:         "seller-one",
			Tracking:       "TRACK-SAMPLE-1001",
			Quantity:       2,
			Amount:         17.98,
			Status:         "expected",
			ExpectedOn:     "2026-02-08",
			Notes:          "source_url=https://example.test/orders/EBAY-SAMPLE-1001 review_state=pending_reconciliation",
		},
		{
			EntryID:        "sample-purchase-life-ebay-watch-002",
			ArrivalID:      "sample-purchase-arrival-ebay-watch-002",
			ItemPartNumber: "CAB-DEMO-003",
			OrderID:        "EBAY-SAMPLE-1001",
			Source:         "ebay",
			Seller:         "seller-one",
			Tracking:       "TRACK-SAMPLE-1001",
			Quantity:       1,
			Amount:         15.25,
			Status:         "delivered",
			ExpectedOn:     "2026-02-08",
			DeliveredOn:    "2026-02-07",
			Notes:          "source_url=https://example.test/orders/EBAY-SAMPLE-1001 review_state=needs_review received_state=partial",
		},
		{
			EntryID:        "sample-purchase-life-manual-2001",
			ArrivalID:      "sample-purchase-arrival-manual-2001",
			ItemPartNumber: "CAB-DEMO-006",
			OrderID:        "MANUAL-SAMPLE-2001",
			Source:         "manual",
			Seller:         "Local-collectors-fair",
			Tracking:       "PICKUP-SAMPLE-2001",
			Quantity:       1,
			Amount:         19.95,
			Status:         "delivered",
			ExpectedOn:     "2026-02-02",
			DeliveredOn:    "2026-02-02",
			Notes:          "source_url=https://example.test/orders/MANUAL-SAMPLE-2001 review_state=ready received_state=complete",
		},
		{
			EntryID:        "sample-purchase-life-shop-3001",
			ArrivalID:      "sample-purchase-arrival-shop-3001",
			ItemPartNumber: "CAB-DEMO-004",
			OrderID:        "SHOP-SAMPLE-3001",
			Source:         "storefront",
			Seller:         "Toy-Shop-AU",
			Tracking:       "AUSPOST-SAMPLE-3001",
			Quantity:       1,
			Amount:         24.99,
			Status:         "expected",
			ExpectedOn:     "2026-02-12",
			Notes:          "source_url=https://example.test/orders/SHOP-SAMPLE-3001 review_state=tracking_in_transit received_state=unreceived",
		},
	}
}

func seedOnboardingPurchaseSamples(ctx context.Context, dbConn *sql.Tx, profileID string, itemByPart map[string]collection.Item) (int, int, error) {
	samples := onboardingPurchaseSamples()
	createdOrderKeys := map[string]struct{}{}
	for _, sample := range samples {
		item, ok := itemByPart[sample.ItemPartNumber]
		if !ok {
			return 0, 0, fmt.Errorf("purchase sample item %s missing", sample.ItemPartNumber)
		}

		notes := fmt.Sprintf(
			"seller=%s tracking=%s %s",
			sample.Seller,
			sample.Tracking,
			sample.Notes,
		)
		result, err := dbConn.ExecContext(ctx, `
			INSERT OR IGNORE INTO commerce_lifecycle_entries(
				id, profile_id, item_id, state, source, external_ref, quantity, amount, currency, notes
			) VALUES (?, ?, ?, 'purchase', ?, ?, ?, ?, 'AUD', ?)
		`, sample.EntryID, profileID, item.ID, sample.Source, sample.OrderID, sample.Quantity, sample.Amount, notes)
		if err != nil {
			return 0, 0, fmt.Errorf("insert purchase sample lifecycle %s: %w", sample.EntryID, err)
		}
		if rows, _ := result.RowsAffected(); rows > 0 {
			createdOrderKeys[strings.ToLower(sample.Source)+":"+sample.OrderID] = struct{}{}
		}
		if _, err := dbConn.ExecContext(ctx, `
			INSERT OR IGNORE INTO expected_arrivals(
				id, profile_id, item_id, lifecycle_entry_id, source, external_ref, quantity, amount, currency, status, expected_on, delivered_on, notes
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'AUD', ?, ?, ?, ?)
		`, sample.ArrivalID, profileID, item.ID, sample.EntryID, sample.Source, sample.OrderID, sample.Quantity, sample.Amount, sample.Status, sample.ExpectedOn, sample.DeliveredOn, notes); err != nil {
			return 0, 0, fmt.Errorf("insert purchase sample arrival %s: %w", sample.ArrivalID, err)
		}
	}

	var totalOrders int
	if err := dbConn.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM (
			SELECT source, external_ref
			FROM commerce_lifecycle_entries
			WHERE profile_id = ? AND state = 'purchase'
			GROUP BY source, external_ref
		)
	`, profileID).Scan(&totalOrders); err != nil {
		return 0, 0, fmt.Errorf("count purchase sample orders: %w", err)
	}
	return len(createdOrderKeys), totalOrders, nil
}

func seedOnboardingSampleData(
	ctx context.Context,
	profiles *profile.Repository,
	collectionRepo *collection.Repository,
	wishlistSvc *wishlist.Service,
	mediaSvc *media.Service,
	dbConn *sql.DB,
) (onboardingSampleSeedResult, error) {
	tx, err := dbConn.BeginTx(ctx, nil)
	if err != nil {
		return onboardingSampleSeedResult{}, fmt.Errorf("begin onboarding sample transaction: %w", err)
	}
	createdAssetDirs := make([]string, 0)
	committed := false
	defer func() {
		if committed {
			return
		}
		_ = tx.Rollback()
		for _, assetDir := range createdAssetDirs {
			_ = os.RemoveAll(assetDir)
		}
	}()

	result, err := seedOnboardingSampleDataTx(
		ctx,
		profiles.WithTx(tx),
		collectionRepo.WithTx(tx),
		wishlistSvc.WithTx(tx),
		mediaSvc,
		tx,
		&createdAssetDirs,
	)
	if err != nil {
		return onboardingSampleSeedResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return onboardingSampleSeedResult{}, fmt.Errorf("commit onboarding sample transaction: %w", err)
	}
	committed = true
	return result, nil
}

func seedOnboardingSampleDataTx(
	ctx context.Context,
	profiles *profile.Repository,
	collectionRepo *collection.Repository,
	wishlistSvc *wishlist.Service,
	mediaSvc *media.Service,
	dbConn *sql.Tx,
	createdAssetDirs *[]string,
) (onboardingSampleSeedResult, error) {
	active, err := profiles.GetActiveProfile(ctx)
	if err != nil {
		return onboardingSampleSeedResult{}, fmt.Errorf("active profile required: %w", err)
	}

	settings, err := profiles.GetSettings(ctx, active.ID)
	if err != nil {
		return onboardingSampleSeedResult{}, fmt.Errorf("load profile settings: %w", err)
	}
	alreadySeeded := settings["onboarding.sample_data_seeded"] == "1"

	specs := []onboardingSampleSpec{
		{
			Item: collection.Item{
				Brand:       "Cabinet",
				Category:    "Diecast",
				PartNumber:  "CAB-DEMO-001",
				Title:       "Starter Skyline GT-R",
				Make:        "Nissan",
				Model:       "Skyline GT-R R34",
				Year:        "1999",
				Scale:       "1:64",
				Series:      "Welcome Set",
				Description: "Starter sample item for first-time onboarding.",
				Tags:        []string{"starter", "sample", "jdm"},
			},
			Instance: collection.Instance{
				Condition:        "Near Mint",
				Status:           "blister",
				Quantity:         1,
				StorageLocation:  "Starter Shelf A1",
				AcquisitionPrice: 8.99,
				AcquisitionDate:  "2026-01-15",
				Notes:            "Included as sample onboarding data.",
			},
			PriceHistory:         generatedShowcasePriceHistory(1, 8.99),
			Wishlist:             true,
			TargetPrice:          10,
			WishlistPriority:     "high",
			WishlistNotes:        "Sample grail chase: buy when a clean card lands near target.",
			WishlistHighlightHit: true,
			FolderName:           "Watch List",
		},
		{
			Item: collection.Item{
				Brand:       "Cabinet",
				Category:    "Slot Car",
				PartNumber:  "CAB-DEMO-002",
				Title:       "Starter Porsche 911 GT3",
				Make:        "Porsche",
				Model:       "911 GT3",
				Year:        "2023",
				Scale:       "1:64",
				Series:      "Welcome Set",
				Description: "Starter sample item for first-time onboarding.",
				Tags:        []string{"starter", "sample", "euro"},
			},
			Instance: collection.Instance{
				Condition:        "Excellent",
				Status:           "sealed",
				Quantity:         1,
				StorageLocation:  "Starter Shelf A2",
				AcquisitionPrice: 11.49,
				AcquisitionDate:  "2026-01-20",
				Notes:            "Included as sample onboarding data.",
			},
			PriceHistory:         generatedShowcasePriceHistory(2, 11.49),
			Wishlist:             true,
			TargetPrice:          9,
			WishlistPriority:     "low",
			WishlistNotes:        "Sample steady watch: useful comparison item, not urgent.",
			WishlistHighlightHit: false,
			FolderName:           "Store 1",
		},
		{
			Item: collection.Item{
				Brand:       "Cabinet",
				Category:    "Trading Card",
				PartNumber:  "CAB-DEMO-003",
				Title:       "Starter Charizard Holo",
				Make:        "Pokemon",
				Model:       "Base Set",
				Year:        "1999",
				Scale:       "Card",
				Series:      "Welcome Set",
				Description: "Starter trading card sample for first-time onboarding.",
				Tags:        []string{"starter", "sample", "tcg"},
			},
			Instance: collection.Instance{
				Condition:        "Very Good",
				Status:           "custom",
				Quantity:         1,
				StorageLocation:  "Starter Binder 1",
				AcquisitionPrice: 15.25,
				AcquisitionDate:  "2026-01-22",
				Notes:            "Included as sample onboarding data.",
			},
			PriceHistory:           generatedShowcasePriceHistory(3, 15.25),
			Wishlist:               true,
			TargetPrice:            18,
			WishlistPriority:       "high",
			WishlistNotes:          "Sample price-drop candidate: below target should bubble up.",
			WishlistHighlightHit:   true,
			WishlistBelowTargetNow: true,
			FolderName:             "Wishlist Focus",
		},
		{
			Item: collection.Item{
				Brand:       "Cabinet",
				Category:    "Action Figure",
				PartNumber:  "CAB-DEMO-004",
				Title:       "Starter Mandalorian Figure",
				Make:        "Star Wars",
				Model:       "The Mandalorian",
				Year:        "2021",
				Scale:       "6in",
				Series:      "Welcome Set",
				Description: "Starter action figure sample for first-time onboarding.",
				Tags:        []string{"starter", "sample", "display"},
			},
			Instance: collection.Instance{
				Condition:        "Mint",
				Status:           "custom",
				Quantity:         1,
				StorageLocation:  "Display Shelf C1",
				AcquisitionPrice: 24.99,
				AcquisitionDate:  "2026-01-24",
				Notes:            "Included as sample onboarding data.",
			},
			PriceHistory:         generatedShowcasePriceHistory(4, 24.99),
			Wishlist:             true,
			TargetPrice:          22,
			WishlistPriority:     "medium",
			WishlistNotes:        "Sample display target: watch for boxed examples.",
			WishlistHighlightHit: true,
			FolderName:           "Store 2",
		},
		{
			Item: collection.Item{
				Brand:       "Cabinet",
				Category:    "Comic",
				PartNumber:  "CAB-DEMO-005",
				Title:       "Starter Amazing Fantasy #15",
				Make:        "Marvel",
				Model:       "Spider-Man Debut",
				Year:        "1962",
				Scale:       "Issue",
				Series:      "Welcome Set",
				Description: "Starter comic sample for first-time onboarding.",
				Tags:        []string{"starter", "sample", "silver-age"},
			},
			Instance: collection.Instance{
				Condition:        "Fine",
				Status:           "custom",
				Quantity:         1,
				StorageLocation:  "Comic Box D2",
				AcquisitionPrice: 12.75,
				AcquisitionDate:  "2026-01-26",
				Notes:            "Included as sample onboarding data.",
			},
			PriceHistory:         generatedShowcasePriceHistory(5, 12.75),
			Wishlist:             true,
			TargetPrice:          20,
			WishlistPriority:     "medium",
			WishlistNotes:        "Sample silver-age target: review condition before buying.",
			WishlistHighlightHit: true,
			FolderName:           "Warehouse 1",
		},
		{
			Item: collection.Item{
				Brand:       "Cabinet",
				Category:    "Model Kit",
				PartNumber:  "CAB-DEMO-006",
				Title:       "Starter RX-78-2 Gundam",
				Make:        "Bandai",
				Model:       "RX-78-2",
				Year:        "1979",
				Scale:       "1:144",
				Series:      "Welcome Set",
				Description: "Starter model kit sample for first-time onboarding.",
				Tags:        []string{"starter", "sample", "mecha"},
			},
			Instance: collection.Instance{
				Condition:        "New",
				Status:           "sealed",
				Quantity:         1,
				StorageLocation:  "Workshop Shelf E1",
				AcquisitionPrice: 19.95,
				AcquisitionDate:  "2026-01-28",
				Notes:            "Included as sample onboarding data.",
			},
			PriceHistory: generatedShowcasePriceHistory(6, 19.95),
			Wishlist:     false,
			TargetPrice:  0,
			FolderName:   "Warehouse 2",
		},
	}
	specs = append(specs, generatedShowcaseInventorySpecs()...)

	existingItems, err := collectionRepo.ListItemsByProfile(ctx, active.ID)
	if err != nil {
		return onboardingSampleSeedResult{}, fmt.Errorf("list items: %w", err)
	}
	itemByPart := make(map[string]collection.Item, len(existingItems))
	for _, item := range existingItems {
		itemByPart[item.PartNumber] = item
	}

	existingWishlist, err := wishlistSvc.ListByProfile(ctx, active.ID)
	if err != nil {
		return onboardingSampleSeedResult{}, fmt.Errorf("list wishlist: %w", err)
	}
	wishlistByItemID := make(map[string]wishlist.Entry, len(existingWishlist))
	for _, entry := range existingWishlist {
		wishlistByItemID[entry.ItemID] = entry
	}

	folderAssignments := map[string]string{}
	if rawAssignments := settings[inventoryFolderItemAssignmentsSettingsKey]; rawAssignments != "" {
		_ = json.Unmarshal([]byte(rawAssignments), &folderAssignments)
	}

	result := onboardingSampleSeedResult{
		DatasetKind:             "sample_showcase",
		DatasetLabel:            "Cabinet sample showcase data",
		SampleDataDisclosure:    "Seeded example records for onboarding and demos; replace or delete before using this profile as a real working collection.",
		AlreadySeededForProfile: alreadySeeded,
	}

	for _, spec := range specs {
		item, ok := itemByPart[spec.Item.PartNumber]
		if !ok {
			created, createErr := collectionRepo.CreateItemForProfile(ctx, active.ID, spec.Item)
			if createErr != nil {
				return onboardingSampleSeedResult{}, fmt.Errorf("create sample item %s: %w", spec.Item.PartNumber, createErr)
			}
			item = created
			itemByPart[item.PartNumber] = item
			result.CreatedItems++
		}

		instances, listErr := collectionRepo.ListInstancesByItemID(ctx, item.ID)
		if listErr != nil {
			return onboardingSampleSeedResult{}, fmt.Errorf("list instances for %s: %w", item.ID, listErr)
		}
		if len(instances) == 0 {
			instanceInput := spec.Instance
			instanceInput.ItemID = item.ID
			if _, createErr := collectionRepo.CreateInstance(ctx, instanceInput); createErr != nil {
				return onboardingSampleSeedResult{}, fmt.Errorf("create instance for %s: %w", item.ID, createErr)
			}
			result.CreatedInstances++
		}

		if spec.FolderName != "" {
			folderAssignments[item.ID] = spec.FolderName
		}

		createdPhotos, photoErr := seedShowcaseItemPhotos(ctx, dbConn, mediaSvc, item, createdAssetDirs)
		if photoErr != nil {
			return onboardingSampleSeedResult{}, photoErr
		}
		result.CreatedPhotos += createdPhotos

		priceHistory := spec.PriceHistory
		if len(priceHistory) == 0 {
			priceHistory = generatedShowcasePriceHistory(result.CreatedItems+result.CreatedInstances+1, spec.Instance.AcquisitionPrice)
		}
		for index, snapshot := range priceHistory {
			var existingSnapshotID string
			if err := dbConn.QueryRowContext(ctx, `
				SELECT id
				FROM price_snapshots
				WHERE item_id = ? AND snapshot_date = ? AND source = ?
				LIMIT 1
			`, item.ID, snapshot.SnapshotDate, snapshot.Source).Scan(&existingSnapshotID); err != nil && err != sql.ErrNoRows {
				return onboardingSampleSeedResult{}, fmt.Errorf("check price history for %s: %w", item.ID, err)
			} else if err == nil {
				continue
			}

			if _, insertErr := dbConn.ExecContext(ctx, `
				INSERT INTO price_snapshots(id, item_id, snapshot_date, source, min_price, median_price, latest_price, stock_count, created_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
			`,
				fmt.Sprintf("sample-price-%s-%02d", item.ID, index+1),
				item.ID,
				snapshot.SnapshotDate,
				snapshot.Source,
				snapshot.MinPrice,
				snapshot.MedianPrice,
				snapshot.LatestPrice,
				snapshot.StockCount,
			); insertErr != nil {
				return onboardingSampleSeedResult{}, fmt.Errorf("insert price history for %s: %w", item.ID, insertErr)
			}
		}

		if spec.Wishlist {
			if _, exists := wishlistByItemID[item.ID]; !exists {
				priority := spec.WishlistPriority
				if priority == "" {
					priority = "medium"
				}
				notes := spec.WishlistNotes
				if notes == "" {
					notes = "Sample wishlist entry created during onboarding."
				}
				createdEntry, createErr := wishlistSvc.CreateForProfile(ctx, active.ID, wishlist.Entry{
					ItemID:         item.ID,
					TargetPrice:    spec.TargetPrice,
					Priority:       priority,
					Notes:          notes,
					HighlightHit:   spec.WishlistHighlightHit,
					BelowTargetNow: spec.WishlistBelowTargetNow,
				})
				if createErr != nil {
					return onboardingSampleSeedResult{}, fmt.Errorf("create wishlist for %s: %w", item.ID, createErr)
				}
				wishlistByItemID[item.ID] = createdEntry
				result.CreatedWishlistEntries++
			}
		}
	}

	createdPurchaseEntries, totalPurchaseOrders, err := seedOnboardingPurchaseSamples(ctx, dbConn, active.ID, itemByPart)
	if err != nil {
		return onboardingSampleSeedResult{}, err
	}
	result.CreatedPurchaseOrders = createdPurchaseEntries
	result.TotalPurchaseOrders = totalPurchaseOrders

	totalItems, err := collectionRepo.ListItemsByProfile(ctx, active.ID)
	if err != nil {
		return onboardingSampleSeedResult{}, fmt.Errorf("reload items: %w", err)
	}
	totalWishlist, err := wishlistSvc.ListByProfile(ctx, active.ID)
	if err != nil {
		return onboardingSampleSeedResult{}, fmt.Errorf("reload wishlist: %w", err)
	}

	result.TotalItems = len(totalItems)
	result.TotalWishlistEntries = len(totalWishlist)
	folderAssignmentsJSON, err := json.Marshal(folderAssignments)
	if err != nil {
		return onboardingSampleSeedResult{}, fmt.Errorf("marshal folder assignments: %w", err)
	}
	if err := profiles.PutSettings(ctx, active.ID, map[string]string{
		"onboarding.sample_data_seeded":           "1",
		"onboarding.sample_data.dataset_kind":     result.DatasetKind,
		"onboarding.sample_data.dataset_label":    result.DatasetLabel,
		"onboarding.sample_data.disclosure":       result.SampleDataDisclosure,
		inventoryFolderItemAssignmentsSettingsKey: string(folderAssignmentsJSON),
	}); err != nil {
		return onboardingSampleSeedResult{}, fmt.Errorf("mark onboarding seeded: %w", err)
	}

	if _, err := dbConn.ExecContext(ctx, `
		INSERT INTO activity_logs(id, level, action, details)
		VALUES (lower(hex(randomblob(16))), 'info', 'onboarding_sample_seeded', ?)
	`, fmt.Sprintf("profile_id=%s created_items=%d created_photos=%d created_wishlist_entries=%d", active.ID, result.CreatedItems, result.CreatedPhotos, result.CreatedWishlistEntries)); err != nil {
		return onboardingSampleSeedResult{}, fmt.Errorf("log onboarding seed activity: %w", err)
	}

	return result, nil
}
