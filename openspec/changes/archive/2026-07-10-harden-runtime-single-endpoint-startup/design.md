## Context

Cabinet currently has two startup ideas that can conflict:

- normal desktop launches should feel like reopening the same app instance
- runtime fallback and stress tooling must still support deliberate multi-instance operation

The existing runtime specs cover port fallback (`RUNTIME-CORE-011`), same-data-dir attach (`RUNTIME-CORE-012`), and explicit multi-instance orchestration (`runtime-multi-instance`). The gap is the requested-endpoint case: if `host:port` is already serving a healthy Cabinet runtime, a normal launch should reuse that endpoint instead of silently starting a second process on a fallback port. That behavior must stay compatible with explicit parallel mode and stress tooling.

## Goals / Non-Goals

**Goals:**
- make default desktop startup singleton-like at the requested endpoint
- support an explicit “restart the running Cabinet on this endpoint” path
- distinguish “Cabinet already running here” from generic “port occupied”
- preserve intentional parallel/multi-instance workflows as an explicit override
- make attach vs restart vs fallback decisions deterministic and visible in logs/startup output

**Non-Goals:**
- removing port fallback for non-Cabinet listeners
- eliminating parallel-instance support
- adding generic remote process management beyond the local requested Cabinet endpoint
- redesigning runtime setup metadata, update flow, or lifecycle logging beyond what is needed to support the startup decision

## Decisions

### 1. Cabinet-on-requested-endpoint is an attach condition, not a fallback condition
Default startup should first evaluate whether the requested endpoint is already serving a healthy Cabinet runtime. If yes, the launcher should attach/open that endpoint and exit without starting a new server process.

Why:
- matches user expectation for “start Cabinet again”
- prevents accidental review-lane drift such as `17882` silently becoming `17883`
- keeps one endpoint stable for demo/review lanes

Alternative considered:
- always fallback on any occupied port
  - rejected because it treats “Cabinet already running” the same as “some unrelated listener exists,” which is the root of the accidental duplicate-start problem

### 2. Generic port occupation still uses deterministic fallback
If the requested endpoint is occupied by a non-Cabinet listener, startup should continue using the existing fallback-port behavior.

Why:
- preserves current runtime-core behavior for local development and automation
- avoids introducing unnecessary startup failures when a different process happens to own the requested port

Alternative considered:
- fail fast on any occupied requested port
  - rejected because it would break existing local/dev workflows and reduce resilience

### 3. Parallel mode must be explicit and observable
Parallel startup should remain available only as an explicit operator/test choice, and that choice should bypass singleton attach behavior.

Why:
- prevents accidental multi-instance starts
- keeps stress tooling and isolated-lane automation viable
- aligns default product behavior with desktop expectations while preserving test infrastructure

Alternative considered:
- infer parallel intent from different data dirs alone
  - rejected because it is too implicit and easy to trigger accidentally

### 4. Restart must be an explicit startup action, not a side effect
Restarting an existing Cabinet runtime on the requested endpoint should require an explicit startup option, and that path should:

- confirm the requested endpoint is a healthy Cabinet runtime
- resolve the active PID from runtime lifecycle metadata / pid file
- stop the old process cleanly when possible, then force-terminate if needed
- wait for the requested port to become free before starting the new process

Why:
- restarting is materially different from attach/reuse and should never happen implicitly
- it keeps the requested endpoint stable for review lanes while still allowing a “replace old with new” workflow

Alternative considered:
- always restart instead of attach when Cabinet is already running
  - rejected because it is too destructive for normal desktop launches and would violate expected singleton behavior

### 5. Startup diagnostics should explain which branch won
Attach/restart/fallback decisions should remain machine-readable and clearly indicate whether startup:
- same-data-dir metadata
- requested healthy endpoint
- restarted existing endpoint
- fallback port

Why:
- makes demo/support investigations faster
- reduces ambiguity when reviewing runtime behavior from logs or startup output

## Risks / Trade-offs

- [False-positive endpoint detection] → require Cabinet-specific health/runtime checks rather than attaching/restarting against any `200 OK` listener
- [Over-constraining legitimate multi-instance workflows] → keep explicit parallel mode and orchestration specs as the supported escape hatch
- [Restart kills the wrong process] → only allow restart when the requested endpoint verifies as Cabinet and the PID/lifecycle metadata line up with that endpoint
- [Spec overlap between fallback and attach paths] → update runtime-core wording so the decision order is unambiguous: attach/restart-to-Cabinet first, fallback for non-Cabinet listeners second
- [Launcher/runtime/script drift] → include launch-script behavior in tasks so helper scripts do not reintroduce ad hoc duplicate-start logic

## Migration Plan

1. Update runtime-core and runtime-multi-instance specs to clarify the decision order and explicit parallel override.
2. Add/adjust startup tests for:
   - same-data-dir attach
   - requested-endpoint Cabinet attach
   - requested-endpoint Cabinet restart
   - non-Cabinet port fallback
   - explicit parallel bypass
3. Align launcher/runtime helpers and review-lane scripts with the clarified attach/restart behavior.
4. Roll forward on `develop`; rollback is straightforward by reverting the startup decision change if attach classification causes unexpected regressions.

## Open Questions

- Should “Cabinet already running here” detection rely on `/healthz` only, `/api/runtime`, or both?
- Should restart be exposed as a CLI flag only, or should demo/helper scripts expose a named `-Restart` wrapper switch too?
- Should helper scripts like `start-demo2.ps1` continue doing their own preflight port checks once the runtime behavior is fully spec’d and tested?
- Do we want a user-facing console line that explicitly says “Reused existing Cabinet runtime on requested endpoint” / “Restarted existing Cabinet runtime on requested endpoint” in addition to the structured attach log?
