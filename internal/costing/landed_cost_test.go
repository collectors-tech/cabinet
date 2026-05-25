package costing

import "testing"

func TestAllocateLandedCostsByValueWeightAndManualAdjustments(t *testing.T) {
	result, err := AllocateLandedCosts(AllocationRequest{
		Items: []ItemInput{
			{ID: "card-a", PurchaseCents: 10000, DomesticShippingCents: 500, TaxCents: 1000, WeightGrams: 100},
			{ID: "card-b", PurchaseCents: 30000, DomesticShippingCents: 1500, TaxCents: 3000, WeightGrams: 300},
		},
		Components: []CostComponentInput{
			{ID: "intl", Label: "International shipping", AmountCents: 8000, AllocationMethod: AllocationWeight, Provenance: "forwarder-shipment:SHIP-1"},
			{ID: "handling", Label: "Handling", AmountCents: 1200, AllocationMethod: AllocationEqual, Provenance: "forwarder-invoice:INV-1"},
			{ID: "insurance", Label: "Insurance", AmountCents: 2000, AllocationMethod: AllocationValue, Provenance: "manual-adjustment:ADJ-1"},
			{ID: "review", Label: "Manual review adjustment", AmountCents: 300, AllocationMethod: AllocationManual, ManualShares: map[string]int64{"card-a": 2, "card-b": 1}, Provenance: "manual-adjustment:ADJ-2"},
		},
	})
	if err != nil {
		t.Fatalf("AllocateLandedCosts returned error: %v", err)
	}
	if result.TotalDirectCents != 46000 || result.TotalSharedCents != 11500 || result.TotalLandedCents != 57500 {
		t.Fatalf("unexpected totals: %+v", result)
	}
	assertItem := func(index int, id string, direct, allocated, landed int64) {
		t.Helper()
		item := result.Items[index]
		if item.ItemID != id || item.DirectCostCents != direct || item.AllocatedCostCents != allocated || item.LandedCostCents != landed {
			t.Fatalf("unexpected item[%d]: %+v", index, item)
		}
		if len(item.ComponentAllocations) != 4 {
			t.Fatalf("expected 4 component allocations for %s, got %d", id, len(item.ComponentAllocations))
		}
	}
	assertItem(0, "card-a", 11500, 3300, 14800)
	assertItem(1, "card-b", 34500, 8200, 42700)
}

func TestAllocateLandedCostsRejectsInvalidManualAllocation(t *testing.T) {
	_, err := AllocateLandedCosts(AllocationRequest{
		Items: []ItemInput{{ID: "card-a", PurchaseCents: 1000}},
		Components: []CostComponentInput{{
			ID:               "manual",
			AmountCents:      100,
			AllocationMethod: AllocationManual,
			ManualShares:     map[string]int64{"missing": 1},
		}},
	})
	if err == nil {
		t.Fatal("expected unknown manual item error")
	}
}

func TestPlanConsolidationThresholdWarnings(t *testing.T) {
	plan, err := PlanConsolidation(ConsolidationRequest{
		Items: []ItemLandedCost{
			{ItemID: "card-b", LandedCostCents: 42000},
			{ItemID: "card-a", LandedCostCents: 14500},
		},
		ShipmentFeeCents:      2500,
		DestinationLimitCents: 60000,
		WarningBufferCents:    1500,
	})
	if err != nil {
		t.Fatalf("PlanConsolidation returned error: %v", err)
	}
	if plan.EstimatedValueCents != 56500 || plan.EstimatedFeeCents != 2500 || plan.EstimatedTotalCents != 59000 {
		t.Fatalf("unexpected plan totals: %+v", plan)
	}
	if plan.ThresholdState != "near_limit" || len(plan.Warnings) != 1 {
		t.Fatalf("expected near-limit warning, got %+v", plan)
	}
	if plan.Mutable {
		t.Fatal("planner result should be non-mutating until explicitly executed")
	}
	if got := plan.ItemIDs; len(got) != 2 || got[0] != "card-a" || got[1] != "card-b" {
		t.Fatalf("expected deterministic sorted item ids, got %#v", got)
	}
}
