# Runtime #387 Test Results

Date: 2026-03-19
Issue: #387 `[Runtime] Persist lifecycle metadata in cabinet.json and write structured runtime logs to file`
Tester: MegaMind

## Test target
Fresh isolated runtime lane using an empty data directory.

- Executable: `C:\projects\collectors-tech\cabinet\bin\cabinet.exe`
- Port: `17883`
- Data dir: `C:\projects\collectors-tech\cabinet\tmp\demo3\data`
- Profile flag: `demo3-helper`
- Instance flag: `demo3-helper`
- URL: `http://127.0.0.1:17883`

## Scope executed
Executed against a fresh empty data directory to capture true first-run behavior, including setup path and follow-up runtime activity.

### Steps run
1. Created fresh empty data dir `tmp\\demo3\\data`
2. Started isolated instance on `17883`
3. Verified `/healthz` returned `ok`
4. Verified `/api/runtime/setup-status` reported `setup_required=true`
5. Inspected startup-created files
6. Completed runtime setup via `POST /api/runtime/setup-complete`
7. Inspected resulting `cabinet.json`
8. Created a profile via `POST /api/profiles`
9. Probed auth/setup state via `GET /api/auth/requirements?...`
10. Checked for durable runtime log files after real runtime activity
11. Force-killed the process to simulate unclean termination
12. Relaunched same lane and verified recovery/startup behavior
13. Sent Ctrl+C to stop the relaunched instance and verified file cleanup/state

---

## Results

## Phase 1 - Fresh startup on empty data dir

### Observed files immediately after first launch
- `cabinet.db`
- `cabinet.pid`

### Not observed
- `cabinet.json`
- any runtime log file
- any JSONL log file

### Observed setup-status response
`GET /api/runtime/setup-status` returned:
- `setup_required: true`
- `config_path: C:\projects\collectors-tech\cabinet\tmp\demo3\data\cabinet.json`

### Verdict
Fresh startup does **not** produce the lifecycle metadata file or any durable runtime log file before setup completes.

---

## Phase 2 - Setup completion on empty DB instance

### Request executed
`POST /api/runtime/setup-complete`

Payload used:
- `instance_name = Demo3 Runtime Test`
- `profile_key = demo3-runtime-test`
- `auth_mode = local`
- `runtime_port_mode = auto`
- `bootstrap_workspace = Local Workspace`
- `bootstrap_database_ref = Primary DB`

### Result
Setup succeeded and created:
- `cabinet.json`

### Observed `cabinet.json`
It contains sections:
- `instance`
- `storage`
- `runtime`
- `auth`
- `bootstrap`
- `features`
- `meta`

### Observed `meta` payload after setup
```json
{
  "createdAt": "2026-03-19T00:17:05Z",
  "updatedAt": "2026-03-19T00:17:05Z",
  "wizardVersion": "1",
  "currentUrl": "http://0.0.0.0:17883"
}
```

### Missing lifecycle fields required by #387
Not present:
- `meta.startedAt`
- `meta.startedBy`
- `meta.launchSource` / `meta.launchCommand`
- `meta.lastKnownPid`
- `meta.lastKnownUrl`
- `meta.lastHeartbeatAt`
- `meta.lastShutdownAt`
- `meta.lastShutdownReason`
- `meta.lastRunClean`

### Verdict
`cabinet.json` exists only after setup and currently persists setup/config metadata, **not** the richer runtime lifecycle provenance required by #387.

---

## Phase 3 - Runtime activity after setup

### Activity executed
- `GET /healthz`
- `GET /api/profiles`
- `POST /api/profiles`
- `GET /api/auth/requirements?profile_id=<created profile>`

### Additional auth-related note
Profile creation succeeded. A later attempt to activate that profile via `/api/profiles/active` returned `invalid_profile_id`, which looks unrelated to #387 and should not be treated as a runtime logging/lifecycle result.

### Log file check after runtime activity
Checked recursively under:
- `C:\projects\collectors-tech\cabinet\tmp\demo3\data`

### Result
No runtime log file found.
No JSONL file found.
No durable structured log artifact found.

### Verdict
Even after real runtime traffic and setup activity, #387 durable log-file behavior is **not present**.

---

## Phase 4 - Unclean termination

### Action
Force-killed the running process.

### Result after kill
Files present:
- `cabinet.db`
- `cabinet.json`
- `cabinet.pid`

### Verdict
After unclean termination, no new lifecycle metadata was written to indicate:
- unclean last run
- shutdown timestamp
- shutdown reason

---

## Phase 5 - Relaunch after unclean termination

### Action
Relaunched the same instance against the same data dir.

### Observed console output
Relaunch correctly picked up configured identity from `cabinet.json`:
- `Instance: Demo3 Runtime Test`
- `Profile: demo3-runtime-test`

### Observed `cabinet.json` after relaunch
```json
{
  "createdAt": "2026-03-19T00:17:05Z",
  "updatedAt": "2026-03-19T00:18:43Z",
  "wizardVersion": "1",
  "currentUrl": "http://127.0.0.1:17883"
}
```

### What changed
- `meta.updatedAt` changed
- `meta.currentUrl` changed from `http://0.0.0.0:17883` to `http://127.0.0.1:17883`

### What still did not exist
- no `startedAt`
- no `startedBy`
- no `lastKnownPid`
- no `lastShutdownAt`
- no `lastShutdownReason`
- no `lastRunClean`
- no durable runtime log file

### Verdict
Recovery/relaunch works at a basic URL-sync level, but the richer #387 lifecycle contract is still missing.

---

## Phase 6 - Shutdown / PID cleanup

### Action
Stopped the relaunched process from the attached console session.

### Result
After shutdown:
- listener on `17883` was gone
- `cabinet.pid` was removed
- `cabinet.json` remained

### Observed `cabinet.json` state after shutdown
No shutdown metadata was added.
`meta.updatedAt` remained at the relaunch timestamp and did not record a shutdown event.

### Verdict
PID cleanup works, but shutdown provenance required by #387 is not persisted.

---

## Acceptance criteria assessment for #387

### 1. Cabinet writes a durable runtime log file during execution
**FAIL** - no durable log file observed.
