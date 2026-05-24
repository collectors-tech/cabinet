## Purpose
Define the first deliverable slice for landed-cost allocation and consolidation planning.

## Requirements
### Requirement COMMERCE-LANDED-COST-001: Landed cost allocation SHALL produce explainable item cost basis
Cabinet SHALL calculate item landed cost from purchase price, domestic shipping, taxes, forwarder fees, international shipping, handling, consolidation fees, and manual adjustments.

#### Scenario: Allocate shared costs deterministically
- **GIVEN** a set of purchased items with direct costs and shared forwarding cost components
- **WHEN** Cabinet allocates landed costs by equal, value, weight, or manual allocation rules
- **THEN** Cabinet MUST return per-item direct cost, allocated cost, landed cost, allocation method, and provenance for each allocated component without mutating inventory state.

### Requirement COMMERCE-LANDED-COST-002: Manual adjustments SHALL preserve audit provenance
Cabinet SHALL require manual landed-cost adjustments to preserve source/provenance metadata and deterministic allocation shares.

#### Scenario: Reject invalid manual allocation
- **GIVEN** a manual allocation references an item outside the allocation request
- **WHEN** Cabinet validates the landed-cost request
- **THEN** Cabinet MUST reject the request instead of silently reallocating or dropping the manual adjustment.

### Requirement COMMERCE-LANDED-COST-003: Consolidation planning SHALL warn near or over destination thresholds
Cabinet SHALL produce non-mutating consolidation recommendations with threshold states for destination value limits.

#### Scenario: Warn when consolidation plan nears destination limit
- **GIVEN** landed-cost items, a shipment fee estimate, a destination value limit, and a warning buffer
- **WHEN** Cabinet estimates the consolidation plan total
- **THEN** Cabinet MUST return sorted item IDs, estimated value, estimated fee, estimated total, threshold state, and warnings while leaving execution state unchanged.
