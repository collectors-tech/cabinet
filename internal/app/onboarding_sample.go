package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/collectors-tech/cabinet/internal/collection"
	"github.com/collectors-tech/cabinet/internal/profile"
	"github.com/collectors-tech/cabinet/internal/wishlist"
)

const inventoryFolderItemAssignmentsSettingsKey = "inventory.folder-item-assignments.v1"

type onboardingSampleSeedResult struct {
	CreatedItems            int  `json:"created_items"`
	CreatedInstances        int  `json:"created_instances"`
	CreatedWishlistEntries  int  `json:"created_wishlist_entries"`
	TotalItems              int  `json:"total_items"`
	TotalWishlistEntries    int  `json:"total_wishlist_entries"`
	AlreadySeededForProfile bool `json:"already_seeded_for_profile"`
}

type onboardingSampleSpec struct {
	Item                   collection.Item
	Instance               collection.Instance
	Wishlist               bool
	TargetPrice            float64
	WishlistPriority       string
	WishlistNotes          string
	WishlistHighlightHit   bool
	WishlistBelowTargetNow bool
	FolderName             string
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
				FolderName: plan.Name,
			})
			sequence++
		}
	}

	return specs
}

func seedOnboardingSampleData(
	ctx context.Context,
	profiles *profile.Repository,
	collectionRepo *collection.Repository,
	wishlistSvc *wishlist.Service,
	dbConn *sql.DB,
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
			Wishlist:    false,
			TargetPrice: 0,
			FolderName:  "Warehouse 2",
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
		inventoryFolderItemAssignmentsSettingsKey: string(folderAssignmentsJSON),
	}); err != nil {
		return onboardingSampleSeedResult{}, fmt.Errorf("mark onboarding seeded: %w", err)
	}

	if _, err := dbConn.ExecContext(ctx, `
		INSERT INTO activity_logs(id, level, action, details)
		VALUES (lower(hex(randomblob(16))), 'info', 'onboarding_sample_seeded', ?)
	`, fmt.Sprintf("profile_id=%s created_items=%d created_wishlist_entries=%d", active.ID, result.CreatedItems, result.CreatedWishlistEntries)); err != nil {
		return onboardingSampleSeedResult{}, fmt.Errorf("log onboarding seed activity: %w", err)
	}

	return result, nil
}
