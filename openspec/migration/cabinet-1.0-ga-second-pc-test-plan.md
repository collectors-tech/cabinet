# Cabinet 1.0 GA Second-PC Acceptance Test Plan

Issue: #1869

GA roadmap: #2546

Candidate supply: #1868 and #2034

Recovery evidence consumer: #1867

Final approval: #1864

Status: operator runbook; the formal run is blocked until the readiness gate below is complete

## Purpose

Use this plan to test one exact Cabinet 1.0 release candidate on a separate
Windows x64 PC as a collector would use it. The product must run from the
supplied portable package. A repository checkout and Node.js may be used only
to run the evidence recorder; they must not provide a development server,
test-only hooks, fixture providers, or unpublished product files.

Reserve one uninterrupted day. Live provider access, issue reporting, or a
candidate failure can extend the run. Use synthetic collection data and
non-sensitive images rather than a real personal collection.

## Formal-run readiness gate

Do not start the acceptance run until the release coordinator supplies one
exact candidate bundle from #1868 and confirms every item below:

- [ ] #2546 records the owner-approved GA scope: supported Windows versions and
  hardware, supported and Preview providers, Chat/Agent status, and the
  portable-only or signed-installer distribution decision.
- [ ] #2057 records the applicable legal, privacy, retention, licensing,
  support, and incident-ownership decisions.
- [ ] The candidate is built from one frozen full commit SHA and the associated
  candidate workflow completed successfully.
- [ ] The handoff contains the Cabinet portable ZIP, Chrome Companion ZIP, and
  Edge Companion ZIP, their separate `.sha256` files, Cabinet and Companion
  manifests, combined bundle manifest, SBOM, release notes, candidate run ID,
  and exact artifact name.
- [ ] The release coordinator confirms that this is the current candidate. Old
  beta.9, beta.10, source-build, dirty-worktree, and locally rebuilt packages
  are not acceptable substitutes.
- [ ] The operator has the repository evidence tooling at the same candidate
  commit, Node.js 22, PowerShell, current Chrome, current Edge, a screenshot
  tool, and permission to attach sanitized evidence to #1869.

### Required contract reconciliation before a 1.0 verdict

The current recorder treats all 51 rows as mandatory, including the Frontline
and Bonza rows. Its final approval row also names the older Cabinet 0.1 private
beta marker. The recommended GA scope in #2546 makes Frontline and Bonza Preview
unless they are explicitly promoted.

Before the formal run, the owner-approved GA scope and recorder must agree.
Either approve and test the broader all-row contract, or update the checklist,
recorder, tests, and approval wording under a focused test-first issue. Do not
mark an out-of-scope row blocked and then describe the candidate as passing.
Do not use the old private-beta approval text as 1.0 GA approval.

## Stop conditions

Stop the formal run and record the affected row as `fail` when any of these
conditions occurs:

- a computed checksum, manifest value, source commit, version, package target,
  protocol range, or full `/api/runtime.build_revision` does not match;
- Cabinet needs a development server, test hook, dirty source file, or package
  other than the supplied candidate;
- the app crashes, loses or corrupts data, mutates a collection without
  confirmation, reports a false provider success, or cannot recover safely;
- the product or guide contradicts the approved signing/distribution contract;
- a required supported-provider journey cannot be completed truthfully; or
- the evidence would require exposing credentials or private page content.

Use `blocked` only when an exact external condition prevents execution, and
write that condition in `--unblock`. Use `not_run` for work not attempted. Any
`fail`, `blocked`, or `not_run` row means the formal GA acceptance is not a pass.

## Candidate invalidation

Any product, packaging, manifest, or source change invalidates the candidate.
Stop, link the focused issue, wait for a new exact bundle, and initialize a new
evidence pack. The recorder archives stale evidence by candidate fingerprint;
never copy old pass states into a new candidate. If Chrome or Edge updates during
the run, record the new browser version and restart that browser's evidence pack
so the environment identity remains exact.

## Prepare the second PC

1. Install all pending Windows updates, restart, and record the exact Windows
   edition, version, build, CPU, RAM, free disk space, display scaling, Chrome
   version, and Edge version.
2. Use a Windows account that has not run this candidate. Do not reuse a real
   Cabinet data directory or a normal browser profile.
3. Create clean Chrome and Edge profiles dedicated to this run. Disable other
   unpacked extensions in those profiles.
4. Create this folder layout, replacing `<sha12>` with the first 12 characters
   of the candidate commit:

   ```text
   C:\Cabinet-GA-Acceptance\<sha12>\
     candidate\
     data-clean\
     data-restore\
     data-relocated\
     evidence\
       00-identity\
       01-onboarding\
       02-collector\
       03-chrome\
       04-edge\
       05-recovery\
       06-errors\
       07-closeout\
     recorder\
       chrome\
       edge\
   ```

5. Copy the supplied candidate files into `candidate` without renaming them.
   Keep a second untouched copy of the bundle outside the test folders.
6. Prepare synthetic data with recognizable names, for example profile `GA Test`,
   inventory item `GA Test Item 001`, wishlist item `GA Wish 001`, and collection
   `GA Collection 001`. Use a non-sensitive JPEG or PNG for media checks.

## Verify the candidate before opening Cabinet

From PowerShell in the acceptance root, save the output of these checks under
`evidence\00-identity`:

```powershell
Get-ChildItem .\candidate
Get-FileHash .\candidate\*.zip -Algorithm SHA256 | Format-Table -AutoSize
Get-Content .\candidate\*.sha256
```

Compare every computed value with its `.sha256` file and the corresponding
manifest. Confirm that the Cabinet, Chrome, Edge, and combined manifests all
name the same full source commit and immutable version. Confirm the SBOM and
both release-note files named by the manifests are present.

Extract each ZIP into its own new folder. Do not copy over an earlier install.
Start the packaged Cabinet executable with `data-clean`, using the package guide
to set the isolated data directory. Record the port and process ID. Read the
runtime identity at `http://127.0.0.1:<port>/api/runtime` and save the sanitized
response:

```powershell
Invoke-RestMethod 'http://127.0.0.1:<port>/api/runtime' |
  ConvertTo-Json -Depth 8 |
  Tee-Object .\evidence\00-identity\runtime.json
```

The runtime `app_version`, full `build_revision`, and build date must match the
Cabinet manifest exactly. A short SHA match is insufficient.

## Initialize the evidence recorder

Create separate Chrome and Edge evidence packs because browser name, version,
profile, and Companion package are part of the test environment. The `init`
command verifies all three packages and manifests even though each output is
bound to one browser environment.

Run this command from the candidate-matched repository checkout, replacing all
angle-bracket placeholders with observed values and actual paths:

```powershell
node scripts/record-beta-acceptance.mjs init `
  --cabinet-manifest '<acceptance-root>\candidate\<cabinet-manifest>.json' `
  --companion-manifest '<acceptance-root>\candidate\<companion-manifest>.json' `
  --bundle-manifest '<acceptance-root>\candidate\<bundle-manifest>.json' `
  --candidate-run-id '<candidate-run-id>' `
  --candidate-artifact '<exact-artifact-name>' `
  --os-version '<Windows edition, version, and build>' `
  --host-profile 'clean-second-pc' `
  --browser-name 'Google Chrome' `
  --browser-version '<full Chrome version>' `
  --isolated-profile 'Cabinet GA Chrome' `
  --data-directory '<acceptance-root>\data-clean' `
  --app-version '<runtime app_version>' `
  --build-revision '<full runtime build_revision>' `
  --build-date '<runtime build date>' `
  --runtime-port '<port>' `
  --runtime-pid '<pid>' `
  --json '<acceptance-root>\recorder\chrome\beta-acceptance.json' `
  --markdown '<acceptance-root>\recorder\chrome\beta-acceptance.md'
```

Repeat it for Microsoft Edge with the exact Edge version, isolated Edge profile,
and `recorder\edge` output paths. Record browser-independent collector and
recovery evidence in both packs only when it came from the same exact runtime
and data set. Execute and record browser-specific rows independently.

The recorder has these stable row groups:

| Group | Rows | Purpose |
| --- | --- | --- |
| Candidate identity | `IDENTITY-01..11` | Host, artifacts, manifests, versions, checksums, runtime, and notes |
| Collector journey | `COLLECTOR-01..10` | Onboarding, records, media, backup/restore, Discovery, and errors |
| Provider/Companion | `PROVIDER-01..15` | Both browsers, supported providers, provenance, idempotency, and recovery |
| Cross-cutting | `CROSS-01..06` | Restart, isolation, truthful UI, safety, recovery, and exact version proof |
| Failure handling | `FAILURE-01..05` | Issue routing, reruns, verdict, and approval boundary |
| Prohibited shortcuts | `SHORTCUT-01..04` | Package-only, clean, non-publishing execution |

Record one row at a time. A human-observed pass or fail requires
`--operator-confirmed`; ranges such as `COLLECTOR-01..10` are not valid row IDs.

```powershell
node scripts/record-beta-acceptance.mjs record `
  --json '<acceptance-root>\recorder\chrome\beta-acceptance.json' `
  --markdown '<acceptance-root>\recorder\chrome\beta-acceptance.md' `
  --row COLLECTOR-01 --status pass `
  --evidence 'evidence/01-onboarding/onboarding-complete.png' `
  --notes 'Clean first run completed with the GA Test profile.' `
  --operator-confirmed
```

For a failure, use `--status fail` with non-secret evidence and notes. For an
external blocker, use `--status blocked --unblock '<exact condition>' --notes
'<context>'`. Never turn an unavailable step into a pass.

## Operator run sheet

### 1. Identity and clean first run

- [ ] Record all `IDENTITY-01..11` evidence after independent checksum,
  manifest, release-note, browser, and runtime comparison.
- [ ] With networking available, launch only the packaged app against
  `data-clean`. Complete first-run onboarding and create profile `GA Test`.
- [ ] Confirm the app opens to a useful collector screen without a development
  tool, hidden setup step, raw translation key, or unexplained error.
- [ ] Close Cabinet completely, confirm the owned process exits, reopen the
  same package and data directory, and verify the profile persists.
- [ ] Record `COLLECTOR-01`, `CROSS-01`, `CROSS-03`, `CROSS-05`, `CROSS-06`,
  and the applicable `SHORTCUT` rows with screenshots plus observed values.

### 2. Core collector journey

- [ ] Create `GA Test Item 001` with category, identifier, quantity, condition,
  value, notes, and the synthetic media file. Save it.
- [ ] Edit at least two fields. Find the item by search, identifier/barcode, and
  one filter. Reload and restart Cabinet; confirm the edit remains.
- [ ] Mark the media as primary. Confirm the correct primary image and asset
  remain after reload and restart.
- [ ] Create `GA Wish 001`, change priority and status, then mark it purchased.
  Confirm exactly one Inventory item is created and the Wishlist relationship
  is truthful.
- [ ] Create and rename `GA Collection 001`; add the inventory item, move it to
  another collection and back, safely soft-delete the test collection, and
  confirm **All Items** cannot be deleted.
- [ ] Export the synthetic profile and open the export enough to confirm it is
  non-empty and represents the visible records.
- [ ] Create a backup and record its filename, size, time, integrity result, and
  non-secret record/media counts.
- [ ] Exercise the documented media migration or maintenance path and record
  discovered, migrated, already-migrated, duplicate, skipped, failed, and
  orphan counts. Do not invent zeroes when the product did not report them.
- [ ] Record `COLLECTOR-02..07`, `CROSS-01`, and relevant screenshots or
  sanitized exports in both browser packs.

### 3. Chrome Companion and supported providers

- [ ] In the isolated Chrome profile, install the exact extracted Chrome
  Companion using only the candidate guide. Record the extension version,
  source commit, target, protocol range, origin/ID, and browser version.
- [ ] Pair through Cabinet's normal six-digit flow. Do not record the pairing
  code, session token, cookies, or authenticated page contents.
- [ ] Verify connect, browser restart/reconnect, extension service-worker
  restart, credential rotation, revoke-one, revoke-all, and re-pair behavior.
- [ ] Enable only the providers in the owner-approved supported contract.
  Confirm Companion reflects the active profile without an extension rebuild.
- [ ] Run a real saved Market Watch for Voglers, review a Discovery, verify
  source/transport/module/schema provenance, and hand the reviewed result to
  Wishlist or Inventory exactly once.
- [ ] Repeat independently for Hobbytech.
- [ ] If Frontline is supported, complete the lawful user-present search and
  reviewed handoff in `PROVIDER-08`, plus its bounded timeout/recovery row.
- [ ] If Bonza is supported, complete the normal user-present interaction,
  search, reviewed handoff in `PROVIDER-09`, plus its bounded timeout/recovery
  row. Do not automate challenges or export browser credentials.
- [ ] Replay one capture and confirm there is no duplicate item or media. Save
  one permitted provider image and verify its asset manifest, hash, renditions,
  provenance, and primary-image behavior.
- [ ] Force one supported provider into a safe unavailable or disconnected
  state. Confirm it reports no false success, returns within the bounded time,
  and leaves the next provider usable without corrupting observations.
- [ ] Record the applicable `PROVIDER-01..15`, `COLLECTOR-09`, `CROSS-02`, and
  `CROSS-04` rows in the Chrome pack.

### 4. Edge Companion

- [ ] Fully revoke and close the Chrome Companion session before switching.
- [ ] In the isolated Edge profile, install the exact Edge ZIP using only the
  candidate guide and repeat install, pair, reconnect, rotation, revoke, and
  recovery checks.
- [ ] Repeat each supported provider's real Market Watch, Discovery provenance,
  reviewed handoff, replay/idempotency, image, failure-isolation, and restart
  checks required by the approved contract.
- [ ] Record browser-specific `PROVIDER-01..15`, `CROSS-02`, `CROSS-04`, and
  `CROSS-06` evidence in the Edge pack. Reuse browser-independent evidence only
  where it proves the same exact runtime and data set.

### 5. Backup, restore, relocation, and restart

- [ ] With Cabinet stopped, preserve the completed `data-clean` directory and
  the backup outside the extracted application folder.
- [ ] Deliberately change the test data after the backup: edit the inventory
  item, remove its primary media choice, and move it between collections.
- [ ] Restore the backup into `data-restore`, not over the only good data copy.
  Verify profile, Inventory, Wishlist-to-Inventory relationship, Collections,
  Market Watches, Discoveries, the contractually expected Companion state
  (including a truthful re-pair requirement when sessions are intentionally not
  restored), media hashes, primary media, and record counts against the
  pre-mutation evidence.
- [ ] Copy the restored directory to `data-relocated`, start the same candidate
  against that location, and repeat the relationship, media, provider-image,
  reload, and full application-restart checks.
- [ ] Submit one intentionally invalid restore/import file. Confirm no active
  data is mutated, the error is actionable, and a subsequent valid operation
  still works.
- [ ] Record `COLLECTOR-08`, `COLLECTOR-10`, `PROVIDER-14`, `PROVIDER-15`,
  `CROSS-01`, `CROSS-02`, and `CROSS-05` with before/after counts and evidence.

### 6. Closeout and result handback

- [ ] Review every screenshot and text file before upload. Do not record credentials, tokens, cookies,
  pairing codes, private page content, personal
  collection data, or unnecessary local paths.
- [ ] Record every applicable `FAILURE-01..05` and `SHORTCUT-01..04` row. If a
  row does not match the approved 1.0 contract, stop for contract reconciliation
  rather than forcing a status.
- [ ] Validate and deterministically re-render both packs:

  ```powershell
  node scripts/record-beta-acceptance.mjs validate --json '<acceptance-root>\recorder\chrome\beta-acceptance.json'
  node scripts/record-beta-acceptance.mjs render --json '<acceptance-root>\recorder\chrome\beta-acceptance.json' --markdown '<acceptance-root>\recorder\chrome\beta-acceptance.md'
  node scripts/record-beta-acceptance.mjs validate --json '<acceptance-root>\recorder\edge\beta-acceptance.json'
  node scripts/record-beta-acceptance.mjs render --json '<acceptance-root>\recorder\edge\beta-acceptance.json' --markdown '<acceptance-root>\recorder\edge\beta-acceptance.md'
  ```

- [ ] Generate SHA-256 values for the two JSON files, two Markdown files, and
  the sanitized evidence archive. Keep the unredacted working folders private.
- [ ] Attach the sanitized JSON/Markdown packs and concise outcome to #1869.
  Link same-candidate recovery evidence to #1867 and the candidate decision to
  #1864. Every failure must link a focused issue with expected behavior, actual
  behavior, reproduction steps, evidence, requirement, severity, and rerun target.
- [ ] Stop the owned Cabinet process and remove or revoke test Companion
  sessions without deleting the retained evidence.

## Verdict rules

`PASS` is permitted only when the owner-approved GA scope matches the recorder,
both Chrome and Edge evidence packs validate, every required row is `pass`, the
same candidate passes recovery, and no open P0/P1 affects the approved contract.

`FAIL_WITH_BLOCKERS` applies when either pack reports `fail_with_blockers`, a
required provider or recovery step fails, a required row is blocked, or the
scope/approval contract is unresolved. `NOT_RUN` applies while required steps
remain unattempted.

This plan records evidence only. It does not publish a release, approve GA, and
does not promote `develop` to `main`. Those remain separate exact-candidate
decisions.
