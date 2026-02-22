# Cabinet

Desktop-first collector intelligence app.

## Current Status
- Runtime scaffold implemented (issue `#1` in GitHub):
- Go entrypoint and graceful shutdown
- Embedded UI hosting
- SQLite initialization + baseline migrations
- Config system with update channel support
- Signed update signature verification primitives
- Installer packaging workflow for Windows and macOS artifacts (`.github/workflows/release-installers.yml`)

## Run Locally
1. Install Go 1.24+.
2. Run:

```powershell
go run ./cmd/cabinet
```

3. Open:
- `http://127.0.0.1:8080/`
- `http://127.0.0.1:8080/healthz`
- `http://127.0.0.1:8080/api/runtime`

## API (Current Scaffold)
- `GET /api/profiles`
- `POST /api/profiles` with `{ "name": "Default" }`
- `GET /api/profiles/active`
- `PUT /api/profiles/active` with `{ "profile_id": "<id>" }`
- `GET /api/profiles/{profileID}/settings`
- `PUT /api/profiles/{profileID}/settings` with `{ "settings": { "theme": "dark" } }`
- `GET /api/profiles/{profileID}/saved-filters`
- `POST /api/profiles/{profileID}/saved-filters` with `{ "name": "...", "query": {...} }`
- `PUT /api/profiles/{profileID}/saved-filters` with `{ "id": "...", "name": "...", "query": {...} }`
- `DELETE /api/profiles/{profileID}/saved-filters?id=<filterID>`
- `GET /api/profiles/{profileID}/storage`
- `PUT /api/profiles/{profileID}/secrets` with `{ "key": "...", "value": "..." }`
- `GET /api/profiles/{profileID}/secrets?key=<key>`
- `PUT /api/profiles/{profileID}/license` with `{ "license_json": "{...}" }`
- `GET /api/profiles/{profileID}/license`
- `GET /api/items`
- `POST /api/items` with canonical item payload
- `GET /api/items/{itemID}/instances`
- `POST /api/items/{itemID}/instances` with instance payload
- `GET /api/items/{itemID}/barcodes`
- `POST /api/items/{itemID}/barcodes` with `{ "barcode": "..." }`
- `GET /api/barcodes/{barcode}` for local matches
- `GET /api/search/items?q=...&brand=...&category=...&condition=...&status=...&tags=...&scale=...&sort=...&limit=...`
- `GET /api/scanner/query-sets`
- `POST /api/scanner/query-sets` with query set payload
- `POST /api/scanner/run` with `{ "query_set_id": "<id>" }`
- `POST /api/scanner/run/scheduled`
- `GET /api/scanner/candidates?query_set_id=<id>`
- `GET /api/scanner/failures`
- `GET /api/provider/health?provider=ebay`
- `POST /api/matching/run`
- `GET /api/matching/results`
- `GET /api/data/export/json`
- `GET /api/data/export/csv/items`
- `POST /api/data/import/json/dry-run` with `{ "snapshot": {...} }`
- `POST /api/data/import/json/apply` with `{ "snapshot": {...}, "options": { "default_action": "merge|create|skip", "overrides": { "PART": "merge|create|skip" } } }`
- `POST /api/data/import/csv/dry-run` with `{ "csv": "....", "mapping": { "part_number": "pn_column" } }`
- `POST /api/data/import/csv/apply` with `{ "csv_import": { "csv": "....", "mapping": {...} }, "options": {...} }`
- `POST /api/data/reindex`
- `POST /api/data/repair`
- `POST /api/backup/run`
- `GET /api/backup/list`
- `POST /api/backup/restore` with `{ "backup_path": "..." }`
- `GET /api/items/{itemID}/photos`
- `POST /api/items/{itemID}/photos` multipart form field `file`
- `DELETE /api/items/{itemID}/photos/{photoID}`
- `PUT /api/items/{itemID}/photos/{photoID}/primary`
- `POST /api/items/{itemID}/photos-rebuild`
- `POST /api/auth/webauthn/register/begin` with `{ "profile_id": "<id>" }`
- `POST /api/auth/webauthn/register/finish` with `{ "session_id": "...", "credential": {...} }`
- `GET /api/auth/requirements?profile_id=<id>`
- `POST /api/auth/webauthn/login/begin` with `{ "profile_id": "<id>" }`
- `POST /api/auth/webauthn/login/finish` with `{ "session_id": "...", "credential": {...} }` returns session token
- `POST /api/auth/recovery/passphrase` with `{ "profile_id": "<id>", "passphrase": "..." }`
- `POST /api/auth/recovery/reset/begin` with `{ "profile_id": "<id>", "passphrase": "..." }`
- `POST /api/auth/session/validate` with `{ "session_token": "..." }`
- `POST /api/auth/session/lock` with `{ "session_token": "..." }`

## Environment Variables
- `CABINET_ADDR` default: `127.0.0.1:8080`
- `CABINET_DATA_DIR` default:
  - Windows: `%LOCALAPPDATA%\\Cabinet`
  - macOS/Linux: `$HOME/.cabinet`
- `CABINET_DB_PATH` default: `${CABINET_DATA_DIR}/cabinet.db`
- `CABINET_UPDATE_CHANNEL` values: `stable`, `beta` (default `stable`)
- `CABINET_UPDATE_PUBLIC_KEY` base64 ed25519 public key for update signature validation
- `CABINET_WEBAUTHN_RP_ID` default: `127.0.0.1`
- `CABINET_WEBAUTHN_ORIGIN` default: `http://127.0.0.1:8080`
- `CABINET_WEBAUTHN_RP_NAME` default: `Cabinet`
- `CABINET_BACKUP_INTERVAL_MINUTES` default: `60`

eBay provider settings are stored per profile via `PUT /api/profiles/{profileID}/settings`:
- `ebay_bearer_token`
- `ebay_marketplace` (example: `EBAY_US`)
- `ebay_base_url` (optional override)
