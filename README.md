# Cabinet

Desktop-first collector intelligence app.

## 🚧 Current Delivery Mode

This repo is currently worked **direct/manual**.

### Current source of truth
- GitHub issue backlog + project board decide what gets worked next
- OpenSpec + traceability define the required behavior
- Validation evidence is mandatory before completion claims
- Branch/deploy workflow is enforced through repo rules

### How work flows now
1. Pick the next real backlog issue.
2. Claim the issue with a comment and durable backlog/project state.
3. Bind or update the relevant spec/governance requirement(s).
4. Implement the issue on **one focused issue branch**.
5. Validate with the required checks for the touched scope.
6. Commit with an issue-prefixed message.
7. Push and update the issue with evidence.
8. Open a PR from the issue branch into `develop`.
9. Manually run the local pipeline on the local dev instance, deploy locally, and run the full regression suite.
10. Comment in the PR with the validation/deploy report and upload the report artifact there.
11. Merge the PR into `develop`.
12. Pull/update local `develop` and deploy demo/review lanes from `develop`.
13. Merge `develop` into `main` only after Max explicitly approves.

### Workflow policy (enforced)
- Issue -> Spec -> Validate -> Commit
- One focused branch per issue/fix whenever possible
- UI checks must verify control intent outcomes, form-field behavior, dialog/layering behavior, and persistence/data outcomes where relevant
- OpenSpec validation gate required for implementation work
- No “done” claims without command/test evidence
- Do the active work directly in-repo using the current manual delivery flow

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
2. Optional: copy `.env.example` to `.env` and adjust runtime URL/port.
3. Build local binary (includes `ui.web` build -> `internal/ui/static`):

```powershell
./scripts/build-cabinet.ps1
```

4. Run:

```powershell
./bin/cabinet.exe
```

5. Open:
- `http://127.0.0.1:17880/`
- `http://127.0.0.1:17880/healthz`
- `http://127.0.0.1:17880/api/runtime`

### Isolated demo / helper instance
- Runbook: `references/demo-instance-plan.md`
- One-command helper launcher: `./scripts/runtime/start-demo2.ps1`

## Branch / Demo Promotion Workflow
- Create one focused branch per issue/fix.
- Validate on that issue branch first.
- Merge validated issue branches into `develop`.
- Deploy demo/review lanes from `develop`, not from ad hoc branch heads or dirty working trees.
- Every demo checkpoint should state the deployed branch and commit hash.
- Merge `develop` into `main` only after Max explicitly says testing is complete and approves the merge.

## Frontend Development
- Frontend source: `ui.web/` (shadcn-admin aligned)
- Embedded output served by Go: `internal/ui/static/`

Build commands:

```powershell
cd ui.web
npm install
npm run build
npm run e2e:foundation
```

Then run Cabinet again:

```powershell
./scripts/build-cabinet.ps1
./bin/cabinet.exe
```

Notes:
- `./scripts/build-cabinet.ps1` always builds `ui.web` and refreshes `internal/ui/static` before `go build`.
- Use `./scripts/build-ui-static.ps1` only for explicit UI-only rebuild workflows.

## API Documentation
- OpenAPI source: `docs/api/openapi.yaml`
- Runtime docs UI: `http://127.0.0.1:17880/apidocs`
- Runtime OpenAPI spec: `http://127.0.0.1:17880/api/openapi.yaml`
- Generated static docs output: `docs/api/index.html`
- Validate spec: `./scripts/validate-openapi.ps1`
- Generate static docs: `./scripts/generate-api-docs.ps1`

## OpenSpec (Spec-First Workflow)
- OpenSpec workspace: `openspec/`
- Workflow guide: `openspec/WORKFLOW.md`
- Validate active changes:

```powershell
openspec validate --changes --strict --no-interactive
```

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
- `POST /api/auth/cloud/session/bootstrap` with `{ "provider": "clerk", "token": "<jwt>" }`
- `POST /api/onboarding/sample-data`

## Environment Variables
- `.env` in repo root is loaded automatically when present.
- Process env vars override `.env` values.
- `CABINET_ADDR` default: `127.0.0.1:17880`
- `CABINET_DATA_DIR` default:
  - Windows: `%LOCALAPPDATA%\\Cabinet`
  - macOS/Linux: `$HOME/.cabinet`
- `CABINET_DB_PATH` default: `${CABINET_DATA_DIR}/cabinet.db`
- `CABINET_UPDATE_CHANNEL` values: `stable`, `beta` (default `stable`)
- `CABINET_UPDATE_PUBLIC_KEY` base64 ed25519 public key for update signature validation
- `CABINET_WEBAUTHN_RP_ID` default: `127.0.0.1`
- `CABINET_WEBAUTHN_ORIGIN` default: `http://127.0.0.1:17880`
- `CABINET_WEBAUTHN_RP_NAME` default: `Cabinet`
- `CABINET_BACKUP_INTERVAL_MINUTES` default: `60`
- `VITE_CLERK_PUBLISHABLE_KEY` enables Clerk sign-in gate and cloud entitlement bootstrap in the web UI

eBay provider settings are stored per profile via `PUT /api/profiles/{profileID}/settings`:
- `ebay_bearer_token`
- `ebay_marketplace` (example: `EBAY_US`)
- `ebay_base_url` (optional override)
