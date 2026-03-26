# START.md - Cabinet Chat-Orchestrated ACP Start

Read this file and `BOOTSTRAP.md`, then start work.

## Execution mode
This file is an execution instruction, not a summary or acknowledgment task.

After reading this file:
- do **not** summarize it
- do **not** explain it
- do **not** acknowledge that you read it
- begin executing immediately in chat
- keep visible chat narration active throughout the run

If you only describe the plan, give a summary, or go quiet during waits, execution has failed.

## Mode
**Active execution mode: Development mode.**

This file currently governs **Development-mode execution only**.
Use it for issue-driven implementation/fix delivery where the expected loop is:
`Issue -> Spec/Governance -> Implement -> Validate -> Commit`

Do **not** treat this file as a generic exploratory-review loop.
If the work is exploratory-only, route-audit-only, or analysis-only, that is a different operating mode and should be announced explicitly instead of silently using this Development loop.

Stay in the current chat session as the planner/orchestrator.
Do **not** use an ACP coordinator session.
Do **not** use nested ACP coordinator -> ACP child orchestration.
Do **not** use Antfarm.

## Core visible-chat rule
If you are doing anything related to this run, you must say what you are doing in chat.

The user must never be left wondering whether the session is:
- alive
- selecting an issue
- spawning a worker
- waiting
- checking status
- reading results
- validating
- blocked
- moving to the next issue

Silence is failure.
Sparse milestone-only updates are failure.
After-the-fact summaries are failure.

Default to being noisier than feels necessary.
If you must choose between slightly too many short operational messages and too little visibility, choose more visibility.

## Verbosity rule
Visible orchestration updates are not just boundary markers. They are the execution trace.

During active work, narrate:
- what exact issue you are selecting
- what exact worker action you are taking
- what exact readback/status check you are performing
- what exact state the worker is currently in
- what exact next step you will do after the current wait/check completes

Prefer concrete operational lines over vague filler.

Good:
- `reviewing Cabinet backlog to choose the next issue`
- `choosing issue #370 now`
- `spawning Codex worker for issue #370 now`
- `worker for issue #370 id: <sessionKey>`
- `checking child-session history for issue #370`
- `worker for issue #370 still running; no final result yet`
- `reading latest worker output for issue #370`
- `worker for issue #370 reached validation step`
- `waiting 10s then checking issue #370 again`

Bad:
- `working on it`
- `still going`
- `please wait`
- any silent wait with no visible operational check

## Maximum silence rule
During an active wait, do not remain silent for longer than 10 seconds.

If the worker is still running, emit another short visible status/readback message in chat.
If the run is especially active or ambiguous, use even shorter intervals.

## Wait-loop rule
While a worker is active, every wait cycle should usually produce a visible message pair:
1. what you are checking now
2. what happened / what you will check next

Default wait-loop examples:
- `checking worker for issue #370`
- `worker for issue #370 still running; checking again in 10s`
- `reading latest output for issue #370`
- `latest output shows validation in progress; waiting 10s`

Do not compress an extended wait into one message and then disappear.

## Ownership and follow-through rule
Spawning a worker is not progress by itself.
The orchestrator owns the issue until a visible result or explicit blocker has been reported in chat.

Fire-and-forget behavior is failure.
If you spawn a worker and then stop checking, stop reading, or stop narrating, execution has failed even if the child eventually finishes.

The orchestrator must visibly prove ownership by:
- confirming the worker was actually created
- reporting the worker identifier/session key
- checking readback/status after spawn
- continuing readback/status checks until completion or blocker
- reporting final outcome or explicit stop condition in chat

## Forced post-spawn cadence
Immediately after spawning a worker, do not drift into a long silent wait.

Required early cadence after every spawn:
1. spawn message
2. worker id/session key message
3. first visible readback/status check within 60 seconds of spawn
4. at least one more visible follow-up check during the next 60 seconds if the worker is still not complete

If readback/status cannot be obtained, say so explicitly in chat and treat that as a real failure state, not invisible waiting.

## Long-run follow-up cadence
If a worker remains active beyond the initial post-spawn period, the orchestrator must continue visible ownership checks.

Required minimum cadence:
- continue short operational checks during active/uncertain early execution
- after that, post a visible ownership/status check at least every 5 minutes until completion or blocker

Each 5-minute follow-up must say:
- which issue is still owned
- approximate elapsed runtime since spawn/resume
- whether readback/status was successfully checked
- the latest known worker state
- the next check plan or escalation step

Example:
- `issue #370 still owned here; running about 10m now; checked child session just now, latest output still in validation, checking again in 5m unless it finishes sooner`

If five minutes pass without a visible ownership check while the worker is still active, execution has failed.

## Elapsed-time tracking rule
The parent/orchestrator must track how long the worker has been running.

Minimum requirement:
- record or retain the spawn/resume time
- include approximate elapsed runtime in long-run follow-up updates
- use elapsed time to decide when to escalate concern, not just whether the worker is technically still alive

Good examples:
- `worker for issue #370 has been running about 6m; latest readback still shows validation`
- `issue #370 has been running about 17m; still owned here, no final result yet`

## Threshold and escalation rule
Long-running delegated work must trigger explicit parent-thread supervision, not passive waiting.

Minimum escalation behavior:
- if the worker is still running at the next 5-minute checkpoint, say so explicitly
- if elapsed time is growing with weak/no meaningful progress, say that concern explicitly
- if readback/status is stale, unavailable, or ambiguous, escalate that as a real orchestration problem
- if runtime exceeds what feels normal for the current scope, state the concern and the next supervisory action in chat

Example escalation lines:
- `issue #370 has been running about 15m; latest output is stale, so I am treating this as a supervision concern and checking again now`
- `issue #370 has been running about 25m; no meaningful new readback, so this is no longer normal waiting and I am escalating/recovering`

## Required visible message pattern per issue
For every issue, visibly send short operational chat updates in this shape:
1. issue selection/start message
2. worker spawned message
3. worker identifier/session key message
4. repeated waiting/status/readback messages while the worker is active
5. issue finished message
6. concise result message
7. next-issue selection message or an explicit stop/blocker message

Minimum expectation:
- do not emit only one wait message per issue
- do not disappear between spawn and finish
- keep the wait/readback loop visibly alive until the worker is actually done

Good examples:
- `reviewing backlog for next issue`
- `starting issue #370 - Wishlist new collection flow lacks validation`
- `spawning worker for issue #370`
- `worker for issue #370 id: <sessionKey>`
- `checking worker for issue #370`
- `worker for issue #370 still running, waiting for result`
- `reading worker result for issue #370`
- `issue #370 finished`
- `issue #370 result: fixed, validated, commit abc1234`
- `blocked on issue #370: runtime would not start`

Bad examples:
- vague filler like `still working on it`
- large reflective commentary
- silent waiting with no visible checks
- only posting a final summary after hidden work

## Working loop
1. Review backlog/current state and pick exactly one open actionable issue.
2. Selection rule: prefer an issue whose GitHub Project status is `Ready`. If no `Ready` issue exists, select the next explicitly prioritized actionable issue and move it to `Ready` before starting if appropriate.
3. Update persistent GitHub state for that issue as work begins:
   - move/update the GitHub Project item from `Ready` to `In progress`
   - apply/update durable labels/tags (for example `in-progress`)
   - add/update an issue comment claiming the work
4. Announce in chat that issue selection is happening.
5. Announce in chat: `Starting issue #<id> - <title>`
6. Spawn exactly one bounded ACP/Codex worker for that issue only.
7. Make the worker read `CHILD_BOOTSTRAP.md`.
8. Immediately post a worker-spawned message in chat.
9. In a separate chat message, report the worker identifier/session key.
10. Wait for/read the worker result back through child-session history/readback.
11. While waiting, visibly narrate meaningful status/readback checks in chat at least every 10 seconds.
12. Make each wait-cycle update concrete: check action, current worker state, and next check timing.
13. Announce in chat: `Issue #<id> finished`
14. Report the result concisely with concrete evidence if available.
15. Update persistent GitHub state based on the outcome:
   - `Blocked` if a real blocker prevents continued progress
   - `In review` when development + validation are complete and the issue is ready for review/verification/merge follow-through
   - `Done` when review follow-through is complete and the issue is completed with evidence
   - labels/tags
   - evidence/blocker comment
16. After moving an issue to `In review`, pick the next `Ready` issue (or prepare the next one), start a new bounded worker/session for that next development unit, and in parallel coordinator mode keep watching `In review` items for review comments, verification needs, follow-up issues, and completion transitions.
17. If an `In review` item reveals missing work, bugs, or follow-up scope, create the missing issue(s), link/comment appropriately, and keep statuses honest rather than silently folding unrelated work into the reviewed issue.
18. Use the GitHub Project status as the durable "where this issue is at" signal:
   - `Backlog` = not prepared yet
   - `Ready` = next-up queued work
   - `In progress` = active development/testing now
   - `Blocked` = real blocker
   - `In review` = development/testing complete, under review/verification/merge follow-through
   - `Done` = completed with evidence
19. Then pick the next issue and repeat.

## Worker rule
One worker owns one issue only.
The worker may do all work related to that issue:
- implementation
- spec/governance updates
- tests
- validation
- follow-up issues discovered during the work
- blocker reporting
- final summary

But the worker must not start unrelated issues or become a long-running backlog manager.

## Cabinet rules
- Before UI/route validation, verify Cabinet runtime is running.
- Use project-local `bin\cabinet.exe`.
- Use browser profile `project-cabinet`.
- Verify auth/session before authenticated work.
- Do not treat local runtime/profile/session mistakes as product bugs until preconditions are verified.

## Enforcement
Follow:
Issue -> Spec/Governance -> Implement -> Validate -> Commit

Use one bounded ACP worker at a time and keep orchestration in this chat session.

## Failure handling
If a real execution step fails, report it visibly in chat immediately.

Examples:
- worker spawn failed
- no worker identifier returned
- first post-spawn check was missed
- readback failed
- ownership follow-up check was missed
- runtime precondition failed
- validation failed
- GitHub Project/label state could not be updated
- repo is blocked

If a failure happens:
- send one short explicit failure/blocker message in chat
- stop pretending the issue is progressing
- update persistent GitHub state if possible to reflect the blocker/failure
- either retry transparently with visible narration or report the stop condition clearly
