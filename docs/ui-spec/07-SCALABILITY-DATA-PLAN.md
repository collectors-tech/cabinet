# 07 UI Scalability Data Plan

## Purpose
Define deterministic test datasets and generation requirements to validate UI performance and usability at scale.

## Dataset Profiles
1. `S0-Empty`
- 0 items, 0 candidates, 0 photos
- validates first-run empty UX

2. `S1-Starter`
- 100 items
- 200 instances
- 300 photos
- 150 barcode records
- 50 discovery candidates

3. `S2-Growth`
- 5,000 items
- 15,000 instances
- 20,000 photos
- 8,000 barcode records
- 2,000 discovery candidates
- 1,000 wishlist entries
- 12 months pricing history (daily snapshots)

4. `S3-Stress`
- 25,000 items
- 80,000 instances
- 150,000 photos metadata rows
- 40,000 barcode records
- 10,000 discovery candidates
- 5,000 wishlist entries
- 24 months pricing history

## Required Field Distribution
- Brand cardinality: 50+
- Category cardinality: 30+
- Tag cardinality: 200+
- Part number uniqueness: >99%
- Status mix: sealed/blister/loose/custom/on_track realistic distribution

## Interaction Performance Targets (local desktop)
1. Initial home render with cached data: <= 1.0s
2. Inventory search/filter update on S2: <= 300ms median
3. Navigation transition between top-level screens: <= 150ms median
4. Sort operation on visible list region: <= 250ms median
5. Open details panel on selected row: <= 120ms median

## UX Scalability Validation Checklist
- [ ] No layout breakage at 200% zoom and small viewport.
- [ ] No unusable controls from long text values.
- [ ] Virtualization or pagination applied for large lists.
- [ ] Loading states prevent UI freeze perception on S3.
- [ ] Filter/search operations are cancellable/debounced.

## Data Generation Requirements
1. Deterministic seed support (`seed=...`) to reproduce runs.
2. Profile-isolated generation to avoid cross-profile contamination.
3. Generation modes:
- `replace` (clean + regenerate)
- `append` (incremental growth)
4. Support JSON export snapshots for regression baselines.

## Proposed Generator Inputs
- `profile_id`
- `dataset_profile` (S0/S1/S2/S3)
- `seed`
- `date_span_months`
- `include_pricing` (bool)
- `include_discovery` (bool)

## Scalability Test Scenarios
1. `SCAL-001`: Inventory search on S2 with 20 rapid keystrokes.
2. `SCAL-002`: Filter + sort + details open loop on S3.
3. `SCAL-003`: Home attention refresh with high event volume.
4. `SCAL-004`: Discover triage list actions on 10k candidates.
5. `SCAL-005`: Reports load and export on 24-month history.

## Release Gate (Scalability)
- [ ] S2 passes all interaction targets.
- [ ] S3 completes without crashes or unrecoverable UI stalls.
- [ ] Memory growth remains bounded during 15-minute navigation soak.

