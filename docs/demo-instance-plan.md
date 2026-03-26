# Cabinet Demo / Helper Instance Plan

## Purpose
Run a separate Cabinet instance for demos, exploratory validation, and agent-driven checks **without interfering with the main dev instance**.

This plan standardizes:
- ports
- profile / instance names
- data directories
- start / stop / verify steps
- rebuild expectations

---

## Goals

1. **Never touch the active dev runtime state by accident**
   - separate port
   - separate data dir / DB
   - separate profile / instance name

2. **Keep agent work isolated**
   - helper/demo instances are safe for UI checks, login setup, route exploration, and local validation

3. **Make launches repeatable**
   - same conventions every time
   - no ad-hoc flags

4. **Allow multiple parallel lanes when needed**
   - one stable dev instance
   - one or more helper/demo instances

---

## Recommended Lane Model

### Primary dev lane
Use the normal dev runtime for active product development.

- Port: `17880`
- Data dir: normal dev location / existing dev data
- Purpose: manual development, feature verification, normal local use

### Demo/review lane (`demo2-helper`)
This is the **review lane Max inspects**. Treat it like a realistic installed-user environment that receives a new build.

- Port: `17882`
- Profile: `demo2-helper`
- Instance name: `demo2-helper`
- Data dir: `C:\projects\collectors-tech\cabinet\tmp\demo2\data`
- Purpose: seeded review/demo lane for issue walkthroughs and Max validation
- State expectation: preserve or restore believable user/demo data instead of treating it as a sterile empty DB by default

### Validator lane (existing documented pattern)
Keep the validator lane isolated and predictable.

- Port: `17881`
- Profile: `demo1-helper`
- Instance name: `demo1-helper`
- Data dir: `C:\projects\collectors-tech\cabinet\tmp\demo1\data`
- Purpose: validator / strict isolated runtime

### Optional future lanes
If more parallel runs are needed, continue the pattern:

- `demo3-helper` -> port `17883` -> `tmp\demo3\data`
- `demo4-helper` -> port `17884` -> `tmp\demo4\data`

---

## Canonical Launch Flags

Always use these base flags for isolated helper/demo instances:

```powershell
--allow-parallel
--no-open-browser
--data-dir <lane-data-dir>
--profile <lane-profile>
--instance-name <lane-instance-name>
--port <lane-port>
```

Why:
- `--allow-parallel` allows side-by-side runtime instances
- `--no-open-browser` prevents noisy browser launches during automation
- `--data-dir` guarantees separate SQLite/runtime state
- `--profile` and `--instance-name` make the lane explicit
- `--port` prevents clashes with the main dev runtime

---

## Standard Launch Commands

### Start helper/demo lane (`demo2-helper`)

```powershell
New-Item -ItemType Directory -Force -Path 'C:\projects\collectors-tech\cabinet\tmp\demo2\data' | Out-Null
& 'C:\projects\collectors-tech\cabinet\bin\cabinet.exe' `
  --allow-parallel `
  --no-open-browser `
  --data-dir 'C:\projects\collectors-tech\cabinet\tmp\demo2\data' `
  --profile demo2-helper `
  --instance-name demo2-helper `
  --port 17882
```

Expected URL:

```text
http://127.0.0.1:17882
```

### Health check

```powershell
Invoke-WebRequest -UseBasicParsing http://127.0.0.1:17882/
```

### Verify port listener

```powershell
Get-NetTCPConnection -LocalPort 17882 -State Listen
```

---

## Rebuild Policy

### Demo2 review-lane rule
For Max-facing review/handoff work, `demo2` (`17882`) should behave like an **existing user install receiving a new version**.

That means:
- rebuild/restart on `17882` is fine
- but do **not** assume empty DB / first-run setup by default
- preserve or reseed believable review data before asking Max to inspect issue work
- if a task truly requires sterile first-run validation, prefer another lane instead of wiping the review lane casually

### When rebuild is required
Rebuild before starting a demo/helper instance if:
- code changed since the last build
- embedded UI changed
- the purpose is a handoff/demo where the exact latest code matters

Use:

```powershell
./scripts/build-cabinet.ps1
```

### When rebuild is not required
A rebuild is optional if:
- the current `bin\cabinet.exe` was built very recently
- the goal is just quick isolated verification of an already-built change
- no code/UI changes landed since the last build

### Preferred rule
For anything user-facing or checkpoint-worthy, prefer:
1. rebuild
2. start isolated instance
3. verify health
4. share URL

---

## Stop / Cleanup

### Find the lane process

```powershell
Get-NetTCPConnection -LocalPort 17882 -State Listen | Select-Object OwningProcess
```

### Stop it

```powershell
Stop-Process -Id <PID>
```

### Optional state reset
If a fresh demo lane is needed, remove or archive the lane data directory:

```powershell
C:\projects\collectors-tech\cabinet\tmp\demo2\data
```

Do this only when intentionally resetting that lane's state.

---

## Safety Rules

1. **Never point a helper/demo lane at the dev data dir**
2. **Never reuse port `17880` for helper automation**
3. **Always use an explicit lane name** (`demo1-helper`, `demo2-helper`, etc.)
4. **Verify health before announcing the instance ready**
5. **Prefer project-local executable**: `C:\projects\collectors-tech\cabinet\bin\cabinet.exe`
6. **Do not assume fresh build** — state whether the lane was started from an existing build or a newly rebuilt binary
7. **Treat `demo2` as the seeded review lane** — do not wipe it to empty state unless the task explicitly calls for resetting the review environment
8. **Use another lane for sterile/first-run checks when possible** rather than repeatedly destroying Max's review state on `17882`

---

## Operational Checklist

### For agent-started helper/demo lane
- [ ] confirm target lane (`demo2-helper` unless another is needed)
- [ ] confirm unique port
- [ ] confirm unique data dir
- [ ] rebuild if required
- [ ] launch with canonical flags
- [ ] verify HTTP response
- [ ] verify listening port
- [ ] report URL + lane details

### For Max handoff update
Report:
- executable path
- port
- URL
- data dir
- whether it was freshly rebuilt or started from existing `bin\cabinet.exe`
- whether health check passed

---

## Current Recommended Default

For safe parallel agent work right now, use:

- Executable: `C:\projects\collectors-tech\cabinet\bin\cabinet.exe`
- Lane: `demo2-helper`
- Port: `17882`
- Data dir: `C:\projects\collectors-tech\cabinet\tmp\demo2\data`
- URL: `http://127.0.0.1:17882`

This should be treated as the default helper/demo lane unless a stronger reason exists to use another isolated lane.
