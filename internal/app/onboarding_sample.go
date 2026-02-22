package app

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/collectors-tech/cabinet/internal/collection"
	"github.com/collectors-tech/cabinet/internal/profile"
	"github.com/collectors-tech/cabinet/internal/wishlist"
)

type onboardingSampleSeedResult struct {
	CreatedItems            int  `json:"created_items"`
	CreatedInstances        int  `json:"created_instances"`
	CreatedWishlistEntries  int  `json:"created_wishlist_entries"`
	TotalItems              int  `json:"total_items"`
	TotalWishlistEntries    int  `json:"total_wishlist_entries"`
	AlreadySeededForProfile bool `json:"already_seeded_for_profile"`
}

type onboardingSampleSpec struct {
	Item        collection.Item
	Instance    collection.Instance
	Wishlist    bool
	TargetPrice float64
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
			Wishlist:    true,
			TargetPrice: 10,
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
			Wishlist:    false,
			TargetPrice: 0,
		},
		{
			Item: collection.Item{
				Brand:       "Cabinet",
				Category:    "Diecast",
				PartNumber:  "CAB-DEMO-003",
				Title:       "Starter Ford GT40",
				Make:        "Ford",
				Model:       "GT40 Mk II",
				Year:        "1966",
				Scale:       "1:64",
				Series:      "Welcome Set",
				Description: "Starter sample item for first-time onboarding.",
				Tags:        []string{"starter", "sample", "classic"},
			},
			Instance: collection.Instance{
				Condition:        "Good",
				Status:           "loose",
				Quantity:         1,
				StorageLocation:  "Starter Shelf B1",
				AcquisitionPrice: 6.5,
				AcquisitionDate:  "2026-01-22",
				Notes:            "Included as sample onboarding data.",
			},
			Wishlist:    true,
			TargetPrice: 9,
		},
	}

	existingItems, err := collectionRepo.ListItems(ctx)
	if err != nil {
		return onboardingSampleSeedResult{}, fmt.Errorf("list items: %w", err)
	}
	itemByPart := make(map[string]collection.Item, len(existingItems))
	for _, item := range existingItems {
		itemByPart[item.PartNumber] = item
	}

	existingWishlist, err := wishlistSvc.List(ctx)
	if err != nil {
		return onboardingSampleSeedResult{}, fmt.Errorf("list wishlist: %w", err)
	}
	wishlistByItemID := make(map[string]wishlist.Entry, len(existingWishlist))
	for _, entry := range existingWishlist {
		wishlistByItemID[entry.ItemID] = entry
	}

	result := onboardingSampleSeedResult{
		AlreadySeededForProfile: alreadySeeded,
	}

	for _, spec := range specs {
		item, ok := itemByPart[spec.Item.PartNumber]
		if !ok {
			created, createErr := collectionRepo.CreateItem(ctx, spec.Item)
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

		if spec.Wishlist {
			if _, exists := wishlistByItemID[item.ID]; !exists {
				createdEntry, createErr := wishlistSvc.Create(ctx, wishlist.Entry{
					ItemID:       item.ID,
					TargetPrice:  spec.TargetPrice,
					Priority:     "normal",
					Notes:        "Sample wishlist entry created during onboarding.",
					HighlightHit: true,
				})
				if createErr != nil {
					return onboardingSampleSeedResult{}, fmt.Errorf("create wishlist for %s: %w", item.ID, createErr)
				}
				wishlistByItemID[item.ID] = createdEntry
				result.CreatedWishlistEntries++
			}
		}
	}

	totalItems, err := collectionRepo.ListItems(ctx)
	if err != nil {
		return onboardingSampleSeedResult{}, fmt.Errorf("reload items: %w", err)
	}
	totalWishlist, err := wishlistSvc.List(ctx)
	if err != nil {
		return onboardingSampleSeedResult{}, fmt.Errorf("reload wishlist: %w", err)
	}

	result.TotalItems = len(totalItems)
	result.TotalWishlistEntries = len(totalWishlist)
	if err := profiles.PutSettings(ctx, active.ID, map[string]string{"onboarding.sample_data_seeded": "1"}); err != nil {
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
