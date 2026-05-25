package costing

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type AllocationMethod string

const (
	AllocationEqual  AllocationMethod = "equal"
	AllocationValue  AllocationMethod = "value"
	AllocationWeight AllocationMethod = "weight"
	AllocationManual AllocationMethod = "manual"
)

type ItemInput struct {
	ID                    string `json:"id"`
	PurchaseCents         int64  `json:"purchase_cents"`
	DomesticShippingCents int64  `json:"domestic_shipping_cents"`
	TaxCents              int64  `json:"tax_cents"`
	WeightGrams           int64  `json:"weight_grams"`
	Quantity              int    `json:"quantity,omitempty"`
}

type CostComponentInput struct {
	ID               string           `json:"id"`
	Label            string           `json:"label,omitempty"`
	AmountCents      int64            `json:"amount_cents"`
	AllocationMethod AllocationMethod `json:"allocation_method"`
	ManualShares     map[string]int64 `json:"manual_shares,omitempty"`
	Provenance       string           `json:"provenance,omitempty"`
}

type AllocationRequest struct {
	Items      []ItemInput          `json:"items"`
	Components []CostComponentInput `json:"components"`
}

type ItemLandedCost struct {
	ItemID                 string                `json:"item_id"`
	DirectCostCents        int64                 `json:"direct_cost_cents"`
	AllocatedCostCents     int64                 `json:"allocated_cost_cents"`
	LandedCostCents        int64                 `json:"landed_cost_cents"`
	ComponentAllocations   []ComponentAllocation `json:"component_allocations,omitempty"`
	AllocationProvenanceID []string              `json:"allocation_provenance_id,omitempty"`
}

type ComponentAllocation struct {
	ComponentID      string           `json:"component_id"`
	Label            string           `json:"label,omitempty"`
	Method           AllocationMethod `json:"method"`
	AmountCents      int64            `json:"amount_cents"`
	Provenance       string           `json:"provenance,omitempty"`
	AllocationBasis  int64            `json:"allocation_basis"`
	AllocationWeight int64            `json:"allocation_weight"`
}

type AllocationResult struct {
	Items             []ItemLandedCost             `json:"items"`
	TotalDirectCents  int64                        `json:"total_direct_cents"`
	TotalSharedCents  int64                        `json:"total_shared_cents"`
	TotalLandedCents  int64                        `json:"total_landed_cents"`
	AllocationSummary []ComponentAllocationSummary `json:"allocation_summary,omitempty"`
}

type ComponentAllocationSummary struct {
	ComponentID string           `json:"component_id"`
	Label       string           `json:"label,omitempty"`
	Method      AllocationMethod `json:"method"`
	AmountCents int64            `json:"amount_cents"`
	Provenance  string           `json:"provenance,omitempty"`
}

type ConsolidationRequest struct {
	Items                 []ItemLandedCost `json:"items"`
	ShipmentFeeCents      int64            `json:"shipment_fee_cents"`
	DestinationLimitCents int64            `json:"destination_limit_cents"`
	WarningBufferCents    int64            `json:"warning_buffer_cents"`
}

type ConsolidationPlan struct {
	ItemIDs             []string `json:"item_ids"`
	EstimatedValueCents int64    `json:"estimated_value_cents"`
	EstimatedFeeCents   int64    `json:"estimated_fee_cents"`
	EstimatedTotalCents int64    `json:"estimated_total_cents"`
	ThresholdState      string   `json:"threshold_state"`
	Warnings            []string `json:"warnings,omitempty"`
	Mutable             bool     `json:"mutable"`
}

func AllocateLandedCosts(req AllocationRequest) (AllocationResult, error) {
	if len(req.Items) == 0 {
		return AllocationResult{}, errors.New("at least one item is required")
	}

	items := make([]ItemLandedCost, len(req.Items))
	itemByID := make(map[string]int, len(req.Items))
	for i, item := range req.Items {
		id := strings.TrimSpace(item.ID)
		if id == "" {
			return AllocationResult{}, fmt.Errorf("item at index %d requires id", i)
		}
		if _, exists := itemByID[id]; exists {
			return AllocationResult{}, fmt.Errorf("duplicate item id %q", id)
		}
		if item.PurchaseCents < 0 || item.DomesticShippingCents < 0 || item.TaxCents < 0 || item.WeightGrams < 0 {
			return AllocationResult{}, fmt.Errorf("item %q has negative cost input", id)
		}
		direct := item.PurchaseCents + item.DomesticShippingCents + item.TaxCents
		items[i] = ItemLandedCost{
			ItemID:          id,
			DirectCostCents: direct,
			LandedCostCents: direct,
		}
		itemByID[id] = i
	}

	var result AllocationResult
	result.Items = items
	for _, item := range items {
		result.TotalDirectCents += item.DirectCostCents
	}

	for _, component := range req.Components {
		if component.AmountCents < 0 {
			return AllocationResult{}, fmt.Errorf("component %q has negative amount", component.ID)
		}
		if strings.TrimSpace(component.ID) == "" {
			return AllocationResult{}, errors.New("component requires id")
		}
		method := component.AllocationMethod
		if method == "" {
			method = AllocationEqual
		}
		allocations, bases, err := allocateComponent(req.Items, component, method)
		if err != nil {
			return AllocationResult{}, err
		}
		for i, amount := range allocations {
			allocation := ComponentAllocation{
				ComponentID:      component.ID,
				Label:            component.Label,
				Method:           method,
				AmountCents:      amount,
				Provenance:       component.Provenance,
				AllocationBasis:  bases[i],
				AllocationWeight: amount,
			}
			result.Items[i].AllocatedCostCents += amount
			result.Items[i].LandedCostCents += amount
			result.Items[i].ComponentAllocations = append(result.Items[i].ComponentAllocations, allocation)
			if component.Provenance != "" {
				result.Items[i].AllocationProvenanceID = append(result.Items[i].AllocationProvenanceID, component.Provenance)
			}
		}
		result.TotalSharedCents += component.AmountCents
		result.AllocationSummary = append(result.AllocationSummary, ComponentAllocationSummary{
			ComponentID: component.ID,
			Label:       component.Label,
			Method:      method,
			AmountCents: component.AmountCents,
			Provenance:  component.Provenance,
		})
	}
	result.TotalLandedCents = result.TotalDirectCents + result.TotalSharedCents
	return result, nil
}

func PlanConsolidation(req ConsolidationRequest) (ConsolidationPlan, error) {
	if len(req.Items) == 0 {
		return ConsolidationPlan{}, errors.New("at least one landed-cost item is required")
	}
	if req.ShipmentFeeCents < 0 || req.DestinationLimitCents < 0 || req.WarningBufferCents < 0 {
		return ConsolidationPlan{}, errors.New("consolidation thresholds and fees cannot be negative")
	}
	plan := ConsolidationPlan{
		EstimatedFeeCents: req.ShipmentFeeCents,
		ThresholdState:    "under_limit",
		Mutable:           false,
	}
	for _, item := range req.Items {
		if strings.TrimSpace(item.ItemID) == "" {
			return ConsolidationPlan{}, errors.New("landed-cost item requires id")
		}
		if item.LandedCostCents < 0 {
			return ConsolidationPlan{}, fmt.Errorf("item %q has negative landed cost", item.ItemID)
		}
		plan.ItemIDs = append(plan.ItemIDs, item.ItemID)
		plan.EstimatedValueCents += item.LandedCostCents
	}
	sort.Strings(plan.ItemIDs)
	plan.EstimatedTotalCents = plan.EstimatedValueCents + plan.EstimatedFeeCents
	if req.DestinationLimitCents > 0 {
		switch {
		case plan.EstimatedTotalCents > req.DestinationLimitCents:
			plan.ThresholdState = "over_limit"
			plan.Warnings = append(plan.Warnings, "estimated total exceeds destination value limit")
		case req.WarningBufferCents > 0 && plan.EstimatedTotalCents+req.WarningBufferCents > req.DestinationLimitCents:
			plan.ThresholdState = "near_limit"
			plan.Warnings = append(plan.Warnings, "estimated total is within warning buffer of destination value limit")
		}
	}
	return plan, nil
}

func allocateComponent(items []ItemInput, component CostComponentInput, method AllocationMethod) ([]int64, []int64, error) {
	bases := make([]int64, len(items))
	switch method {
	case AllocationEqual:
		for i := range bases {
			bases[i] = 1
		}
	case AllocationValue:
		for i, item := range items {
			bases[i] = item.PurchaseCents + item.DomesticShippingCents + item.TaxCents
		}
	case AllocationWeight:
		for i, item := range items {
			bases[i] = item.WeightGrams
		}
	case AllocationManual:
		for i, item := range items {
			bases[i] = component.ManualShares[item.ID]
		}
	default:
		return nil, nil, fmt.Errorf("unsupported allocation method %q", method)
	}
	if method == AllocationManual {
		for itemID := range component.ManualShares {
			found := false
			for _, item := range items {
				if item.ID == itemID {
					found = true
					break
				}
			}
			if !found {
				return nil, nil, fmt.Errorf("manual share references unknown item %q", itemID)
			}
		}
	}
	return allocateByBasis(component.AmountCents, bases)
}

func allocateByBasis(amount int64, bases []int64) ([]int64, []int64, error) {
	var totalBasis int64
	for _, basis := range bases {
		if basis < 0 {
			return nil, nil, errors.New("allocation basis cannot be negative")
		}
		totalBasis += basis
	}
	if totalBasis <= 0 {
		return nil, nil, errors.New("allocation basis must be positive")
	}
	allocations := make([]int64, len(bases))
	type remainder struct {
		index int
		value int64
	}
	remainders := make([]remainder, len(bases))
	var allocated int64
	for i, basis := range bases {
		numerator := amount * basis
		allocations[i] = numerator / totalBasis
		allocated += allocations[i]
		remainders[i] = remainder{index: i, value: numerator % totalBasis}
	}
	sort.SliceStable(remainders, func(i, j int) bool {
		if remainders[i].value == remainders[j].value {
			return remainders[i].index < remainders[j].index
		}
		return remainders[i].value > remainders[j].value
	})
	for remaining := amount - allocated; remaining > 0; remaining-- {
		allocations[remainders[0].index]++
		remainders = append(remainders[1:], remainders[0])
	}
	return allocations, bases, nil
}
