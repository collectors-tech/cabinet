# Cabinet

Cabinet is a desktop-first collector workspace. Cabinet 0.1 is currently a private beta for invited validation, with local collection data, inventory and wishlist workflows, provider-assisted discovery, backup/export, and optional Assistant/Agent surfaces.

The governed **Cabinet 0.1 Private Beta Disclosure** in the in-app Help Center is the source for supported, preview, limited, browser-assisted, action-required, packaged-unproven, and excluded capability claims.

## Private-beta distribution

- Cabinet is distributed as an unsigned Windows portable ZIP. It is not an installer, Microsoft Store package, or automatic-update channel.
- Verify the supplied SHA-256 file, extract the ZIP into a writable, stable folder, and run `cabinet.exe`.
- The candidate includes `cabinet-mcp.exe`, this README, and `WINDOWS-PORTABLE-BETA.md`.
- Browser Companion candidates are separate Chrome and Edge ZIPs loaded through developer mode. They are not browser-store releases and do not update automatically.
- A package is identified by its version, exact source commit, release manifest, and checksum. Do not substitute a mutable branch download for an accepted candidate.

See [Windows portable install, upgrade, rollback, and removal](WINDOWS-PORTABLE-BETA.md), which is supplied at the top level of the extracted package. After Cabinet starts, open **Help Center > Integrations** for Browser Companion install, pairing, provider capture, revocation, and recovery guidance. The companion candidate may also be supplied with separate target-specific release notes and a manifest.

## Start Cabinet on Windows

1. Verify and extract the supplied portable ZIP.
2. Run `cabinet.exe` from the extracted folder.
3. Open the printed URL if Cabinet does not open it automatically. The normal loopback URL is `http://127.0.0.1:17880/`.
4. Complete first-run setup and select local mode, or use ZITADEL only when the beta deployment was configured for that authority.
5. Back up from Settings before replacing a build or reusing an existing data directory.

### Data paths

For a normal portable launch, Cabinet creates a `data` directory beside `cabinet.exe`. This executable-local directory is the default runtime root and normally contains `cabinet.db`, `cabinet.json`, media, backups, and runtime logs.

`CABINET_DATA_DIR` can override the runtime root before startup, and `CABINET_DB_PATH` can override the database file. First-run custom storage can also record another writable storage location. The `data_dir` returned by `/api/runtime` is authoritative for the running process; inspect it before backup, upgrade, rollback, or deletion instead of assuming a path.

Deleting only `cabinet.exe` does not reliably remove workspace data. With the default portable layout, deleting the whole extracted folder also deletes its `data` subdirectory. Keep a verified backup outside that folder before upgrade, relocation, or removal.

## Product and data boundaries

- Local mode keeps Cabinet identity and workspace operations on the local runtime. ZITADEL mode is available only when its issuer, client, audience, and redirect configuration are supplied by the deployment operator.
- Provider integrations run only when configured or invoked. Direct provider calls disclose the request and normal network metadata to that provider under its own terms.
- The optional Browser Companion uses paired, profile-scoped loopback access. It captures supported page observations from a user-present Chrome or Edge tab, does not export cookies or tokens, does not bypass challenges, and does not perform provider writes.
- Remote diagnostics are disabled by default. Local diagnostics remain in the Cabinet data directory; an explicitly configured remote diagnostics endpoint is contacted only after opt-in, using the runtime redaction boundary.
- Settings exposes profile data JSON/CSV exports, backup/restore, and redacted diagnostic-log export. Review exported files before sharing them.

Read the in-app Privacy Policy, Terms of Service, Help Center, and the canonical beta disclosure before using external providers or Browser Companion.

## Support for this beta

Use the beta coordinator who supplied the candidate as the support route. Include the Cabinet version, exact source commit, package checksum, the failing action, and a redacted diagnostics export when useful. Do not send passwords, provider credentials, Browser Companion credentials, cookies, authorization headers, raw private page content, or an unreviewed database/backup.

This private beta has no support service-level commitment. Provider availability, ZITADEL operation, and optional remote diagnostics endpoints may be controlled by separate deployment operators or third parties.

## Developer quick start

Requirements: Go 1.24+, Node.js/npm, and PowerShell on Windows.

```powershell
./scripts/build-cabinet.ps1
./bin/cabinet.exe
```

`scripts/build-cabinet.ps1` builds `ui.web`, refreshes the embedded output in `internal/ui/static`, and builds the Go runtime. Useful runtime endpoints are:

- `http://127.0.0.1:17880/healthz`
- `http://127.0.0.1:17880/api/runtime`
- `http://127.0.0.1:17880/apidocs`
- `http://127.0.0.1:17880/api/openapi.yaml`

For UI-only work:

```powershell
cd ui.web
npm install
npm run build
```

## Engineering sources of truth

- GitHub issues and the project board define tracked delivery work.
- `openspec/specs/` and `openspec/traceability.md` define required behavior and evidence.
- `docs/api/openapi.yaml` is the API contract; validate it with `./scripts/validate-openapi.ps1` and `go run ./cmd/openapi-parity-gate`.
- `docs/help-center/` is published user-facing guidance embedded in the app.
- `release/cabinet-beta-disclosure.json` is the governed capability disclosure rendered into Help Center content and release notes.

Validate OpenSpec changes with:

```powershell
openspec validate --all --strict --no-interactive
```

Release branches are validated into `develop`. Promotion from `develop` to `main`, external publication, or an immutable release requires explicit approval under the repository release workflow.
