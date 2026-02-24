# UI Endpoint Parity Matrix

Last updated: 2026-02-24

## Purpose
Track parity between top-level UI screens, core API workflows, and automated tests.

## Left Nav to Screen Mapping
| Nav Item | Screen Container | Status |
| --- | --- | --- |
| Dashboard | `screen-dashboard` | Implemented |
| Collection | `screen-collection` | Implemented |
| Scanner | `screen-scanner` | Implemented |
| Discoveries | `screen-discoveries` | Implemented |
| AI Assist | `screen-ai` | Implemented |
| Barcodes | `screen-barcodes` | Implemented |
| Photos | `screen-photos` | Implemented |
| Pricing | `screen-pricing` | Implemented |
| Reports | `screen-reports` | Implemented |
| Settings | `screen-settings` | Implemented |

## Endpoint Coverage by Screen
| Screen | Endpoint | UI Path |
| --- | --- | --- |
| Dashboard | `GET /api/dashboard` | Dashboard -> Refresh Dashboard |
| Collection | `GET /api/items` | Collection -> initial load / search |
| Collection | `POST /api/items` | Collection -> Add Item |
| Collection | `GET /api/items/{itemID}/instances` | Collection -> Load Instances |
| Collection | `POST /api/items/{itemID}/instances` | Collection -> Add Instance |
| Collection | `GET /api/items/{itemID}/barcodes` | Collection -> Load Barcodes |
| Collection | `POST /api/items/{itemID}/barcodes` | Collection -> Add Barcode |
| Collection | `GET /api/barcodes/{barcode}` | Collection -> Lookup Barcode |
| Collection | `GET /api/barcodes/{barcode}/external-search` | Collection -> External Search |
| Collection | `GET /api/items/{itemID}/photos` | Collection -> Photos -> Load Photos |
| Collection | `POST /api/items/{itemID}/photos` | Collection -> Photos -> Upload Photo |
| Collection | `POST /api/items/{itemID}/photos-rebuild` | Settings -> Rebuild Thumbnails |
| Scanner | `GET /api/scanner/query-sets` | Scanner -> Load Query Sets |
| Scanner | `POST /api/scanner/query-sets` | Scanner -> Create Query Set |
| Scanner | `POST /api/scanner/run` | Scanner -> Run Now |
| Scanner | `POST /api/scanner/run/scheduled` | Scanner -> Run Scheduled |
| Scanner | `GET /api/scanner/candidates` | Scanner -> Load Candidates |
| Scanner | `GET /api/scanner/failures` | Scanner -> Load Scanner Failures |
| Scanner | `POST /api/scanner/failures/retry` | Scanner -> Retry Failure `<query_set_id>` |
| Scanner | `GET /api/provider/health` | Scanner -> Check Provider Health |
| Scanner | `POST /api/matching/run` | Scanner -> Run Matching |
| Scanner | `GET /api/matching/results` | Scanner -> Matching summary |
| Scanner | `GET /api/discovery/not-in-collection` | Scanner -> Not In Collection -> Load |
| Discoveries | `GET /api/discovery/not-in-collection` | Discoveries -> Load Not In Collection (stock_state/stock_count rendered) |
| Discoveries | `POST /api/discovery/action` | Discoveries -> Ignore/Wishlist/Track/Create actions |
| AI Assist | `POST /api/ai/test` | AI Assist -> Test AI |
| AI Assist | `POST /api/ai/suggest/title` | AI Assist -> Suggest Metadata from Title |
| AI Assist | `POST /api/ai/suggest/photo` | AI Assist -> Suggest Metadata from Photo |
| Barcodes | `POST /api/items/{itemID}/barcodes` | Barcodes -> Add Barcode |
| Barcodes | `GET /api/items/{itemID}/barcodes` | Barcodes -> Load Barcodes |
| Barcodes | `GET /api/barcodes/{barcode}` | Barcodes -> Lookup Barcode |
| Photos | `GET /api/items/{itemID}/photos` | Photos -> Load Photos |
| Photos | `POST /api/items/{itemID}/photos` | Photos -> Upload Staged Photos |
| Photos | `POST /api/items/{itemID}/photos/reorder` | Photos -> Move Up/Move Down (reorder persistence) |
| Reports | `GET /api/wishlist/hits` | Reports -> Load Wishlist Report |
| Reports | `GET /api/pricing/trend` | Reports -> Load Trend Summary |
| Reports | `GET /api/pricing/stats` | Reports -> Load Trend Summary |
| Reports | `GET /api/pricing/by-source` | Reports -> Load Source Summary (stock_count rows) |
| Reports | `GET /api/pricing/history/export` | Reports -> Export Report History |
| Pricing | `GET /api/wishlist` | Pricing -> Load Wishlist |
| Pricing | `POST /api/wishlist` | Pricing -> Add Wishlist Item |
| Pricing | `DELETE /api/wishlist` | Pricing -> Remove Wishlist Item |
| Pricing | `GET /api/wishlist/hits` | Pricing -> Load Wishlist Hits |
| Pricing | `POST /api/pricing/track` | Pricing -> Track Pricing |
| Pricing | `GET /api/pricing/graph` | Pricing -> Load Pricing Graph |
| Pricing | `GET /api/pricing/by-source` | Pricing -> Load Pricing Sources |
| Pricing | `GET /api/pricing/history` | Pricing -> Load Pricing History |
| Pricing | `GET /api/pricing/stats` | Pricing -> Load Pricing Stats |
| Pricing | `GET /api/pricing/trend` | Pricing -> Load Pricing Trend |
| Pricing | `POST /api/pricing/snapshot/run` | Pricing -> Run Pricing Snapshot |
| Pricing | `GET /api/pricing/history/export` | Pricing -> Export Pricing History |
| Settings | `GET /api/profiles/{profileID}/settings` | Settings -> Load Profile Settings |
| Settings | `PUT /api/profiles/{profileID}/settings` | Settings -> Save Profile Settings |
| Settings | `PUT /api/profiles/{profileID}/secrets` | Settings -> Save Secrets |
| Settings | `POST /api/settings/reset-ignore-rules` | Settings -> Reset Ignore Rules |
| Settings | `POST /api/data/reindex` | Settings -> Reindex |
| Settings | `POST /api/data/repair` | Settings -> Repair |
| Settings | `POST /api/backup/run` | Settings -> Run Backup |
| Settings | `GET /api/backup/list` | Settings -> Load Backups |
| Settings | `POST /api/backup/restore` | Settings -> Restore Selected Backup |
| Settings | `GET /api/logs/activity` | Settings -> Load Admin Status |
| Settings | `GET /api/logs/export` | Settings -> Export Logs |
| Settings | `POST /api/logs/debug` | Settings -> Enable/Disable Debug Mode |
| Settings | `GET /api/profiles/{profileID}/license` | Settings -> Load Profile License |
| Settings | `PUT /api/profiles/{profileID}/license` | Settings -> Save Profile License |
| Settings | `POST /api/license/import` | Settings -> Import License File |
| Settings | `GET /api/license/status` | Settings -> Refresh License Status |
| Settings | `GET /api/data/export/json` | Settings -> Export JSON |
| Settings | `POST /api/data/import/json/dry-run` | Settings -> Import/Export Wizard dry run |
| Settings | `POST /api/data/import/json/apply` | Settings -> Import/Export Wizard apply |

## E2E Coverage
Primary parity workflow test:
- `web/playwright/e2e/ui-backlog.spec.ts`
  - `left navigation switches visible advanced-workspace screens (desktop + mobile)`
  - Exercises each top-level screen and key endpoint-backed actions.
