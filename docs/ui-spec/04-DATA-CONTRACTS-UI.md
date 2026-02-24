# 04 UI Data Contracts (Strict)

## Purpose
Map UI widgets to endpoints, required fields, and fallback behavior.

## Home Dashboard
1. Collection Snapshot
- Endpoint: `GET /api/dashboard`
- Required fields: `total_items`, `total_instances`, `estimated_value`
- Fallback: default zero values

2. Watchlist Hits
- Endpoint: `GET /api/wishlist/hits`
- Required fields: `item_id`, `listing_id`, `title`, `price`
- Fallback: empty list

3. Price Changes
- Endpoint: `GET /api/pricing/history`, `GET /api/pricing/trend`
- Required fields: `date/day`, `latest`, `previous`
- Fallback: hide trend chart, show `data unavailable`

4. Discoveries
- Endpoint: `GET /api/discovery/not-in-collection`
- Required fields: `candidate_id`, `title`, `price`, `last_seen`
- Optional stock fields: `stock_status`, `stock_count`, `stock_observed_at`
- Fallback: empty list

5. Scanner Failures
- Endpoint: `GET /api/scanner/failures`
- Required fields: `query_set_id`, `reason`, `attempts`, `last_error_at`
- Fallback: empty list

6. Recovery Alerts
- Endpoint: `GET /api/runtime/recovery`
- Required field: `recovery_required`
- Fallback: treat as unknown and show non-blocking warning

## Inventory Tabs
1. Items
- Endpoint: `GET /api/items`, `POST /api/items`
- Required fields: `id`, `part_number`, `title`

2. Photos
- Endpoint: `GET /api/photos`, `POST /api/photos`, `POST /api/photos/primary`, `DELETE /api/photos`
- Required fields: `id`, `item_id`, `filename`, `is_primary`

3. Barcodes
- Endpoint: `GET /api/items/{id}/barcodes`, `POST /api/items/{id}/barcodes`, `GET /api/barcodes/lookup`
- Required fields: `barcode`

4. AI Assist
- Endpoint: `POST /api/ai/*`
- Required fields: `suggestion`, `confidence`

## Contract Rules
1. UI never crashes on missing optional fields.
2. UI validates required fields before rendering interactive rows.
3. Unknown/missing fields produce deterministic placeholders.
4. All endpoint failures return user-visible error + retry action.
5. Stock fields are optional but when present must render as explicit state (`in stock`, `low stock`, `out of stock`, `unknown`).

## Acceptance Criteria
- [ ] Every critical widget has endpoint and required field contract.
- [ ] Fallback behavior defined for missing/failed data.
- [ ] Contracts align with OpenAPI and runtime handlers.
