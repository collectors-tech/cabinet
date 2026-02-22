# Cabinet

Desktop-first collector intelligence app.

## Current Status
- Runtime scaffold implemented (issue `#1` in GitHub):
- Go entrypoint and graceful shutdown
- Embedded UI hosting
- SQLite initialization + baseline migrations
- Config system with update channel support
- Signed update signature verification primitives

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
- `GET /api/items`
- `POST /api/items` with canonical item payload
- `GET /api/items/{itemID}/instances`
- `POST /api/items/{itemID}/instances` with instance payload
- `GET /api/items/{itemID}/barcodes`
- `POST /api/items/{itemID}/barcodes` with `{ "barcode": "..." }`
- `GET /api/barcodes/{barcode}` for local matches
- `GET /api/search/items?q=...&brand=...&category=...&condition=...&status=...&tags=...&scale=...&sort=...&limit=...`
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
- `POST /api/auth/webauthn/login/begin` with `{ "profile_id": "<id>" }`
- `POST /api/auth/webauthn/login/finish` with `{ "session_id": "...", "credential": {...} }`

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
