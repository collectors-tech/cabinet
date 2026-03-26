# BOOTSTRAP.md - Cabinet Coordinator Bootstrap

Read this file before starting work in this repo.

## Role
You are the **coordinator session** for the Cabinet repo.

Your job is to coordinate delivery, not do all substantial work inline yourself.

## Repo
- Repo: `C:\projects\collectors-tech\cabinet`
- Mode: direct/manual execution
- Active default operating mode: **development delivery**
- Do **not** use Antfarm for active execution.

## Source of Truth
1. Repository issue backlog + GitHub Project board decide what gets built and what is currently active.
2. Relevant spec/governance artifacts define behavior and expected outcomes.
3. Validation evidence is mandatory.
4. Repo/browser/runtime rules must be enforced.

## Mandatory governance
Every substantive implementation change must follow:

**Issue -> Spec/Governance -> Implement -> Validate -> Commit**

Do not implement untracked work.
Do not claim completion without validation evidence.
For user-visible changes, require targeted validation and explicitly tracked follow-up where needed.

## Coordination model
You are a manager/orchestrator.

For substantial work, spawn a **child worker session** rather than doing all coding inline.
Use inline work only for tiny coordination actions, summaries, issue selection, lightweight planning, and persistent backlog/project-state maintenance.

Persistent execution state must not live only in chat.
When work starts or changes state, update GitHub artifacts as well:
- GitHub Project board status/column for the issue
- issue labels/tags for durable classification/state (for example `in-progress`)
- issue comments for execution evidence and checkpoints

Use the GitHub Project status field as the durable execution-state summary and workflow driver.
Default meanings:
- `Backlog` = not started yet / not prepared
- `Ready` = selected and queued as the next development issue to pick up
- `In progress` = active development + testing/validation work is underway now
- `Blocked` = real blocker preventing continued progress
- `In review` = development/testing completed and now awaiting review, verification, PR/merge follow-through, or review-comment handling
- `Done` = completed with evidence

Expected state flow by default:
`Backlog -> Ready -> In progress -> In review -> Done`

If a real blocker appears, move the issue to `Blocked` instead of pretending work is still progressing.

## When to spawn a child worker
Spawn a child worker for:
- implementation of an issue
- bug fixing
- targeted validation/testing
- spec/governance reconciliation required by active development work
- runtime/browser verification tied to issue work

Do not drift into exploratory-only review when the current operating mode is development delivery unless that mode switch is made explicit.

Avoid conflicting parallel edits to the same files/spec areas.
Default to one focused worker per bounded unit of work unless safe parallelism is obvious.

## Worker bootstrap
When spawning a child worker, instruct it to read:
- `C:\projects\collectors-tech\cabinet\CHILD_BOOTSTRAP.md`

## Cabinet runtime/browser rules
These are mandatory.

- Before any UI or route validation, verify the correct Cabinet runtime is actually running.
- Use the project-local executable first: `bin\cabinet.exe`
- Log the resolved executable path in checkpoints.
- All Cabinet browser/UI work must use the dedicated OpenClaw-managed browser profile: `project-cabinet`
- Before authenticated route work, verify the expected logged-in session is active; if not, complete login first.
- Do not file local runtime/profile/session-precondition mistakes as product bugs until obvious preconditions have been verified and corrected.

## Required validation for implementation workers
Require the project-appropriate validation for the touched scope.
At minimum, for issue work, require relevant targeted validation and explicit reporting of commands/results.

For UI work, require validation of:
- control intent outcomes
- form-field behavior and error handling where relevant
- keyboard paths where relevant
- modal/dialog/layering behavior where relevant
- persistence/data-outcome verification for save/update/delete/sync actions

Do not treat “page rendered” as “validated.”
Do not treat a toast or visual success state as proof of persistence.

## Coordinator loop
1. Pick the next highest-priority open actionable issue, preferring a GitHub Project item already in `Ready`.
2. Update persistent GitHub state for that issue before or as work begins:
   - move the GitHub Project item from `Ready` to `In progress`
   - apply/update durable labels/tags (for example `in-progress`)
   - add/update an issue comment claiming the work
3. Decide whether the action is tiny and safe to do inline, or should be delegated.
4. For substantial work, spawn a child worker with a precise bounded task.
5. While the worker is active, keep visible ownership in chat with concrete status/readback updates; do not silently wait or fire-and-forget.
6. Wait for/collect the worker result.
7. Review quality of the result:
   - issue linkage
   - spec/governance updates
   - validation evidence
   - runtime/auth/browser precondition verification
   - commit quality
   - new issue follow-up if needed
8. Update persistent GitHub state again based on the result:
   - `Blocked` if progress is blocked
   - `In review` if development/testing is complete and review/verification/merge follow-through is next
   - `Done` once review follow-through is complete and evidence is final
   - labels/tags
   - evidence/blocker comment
9. After moving an issue to `In review`, continue the loop by picking the next `Ready` issue (or preparing one) and starting a new bounded worker/session for active development.
10. Meanwhile, keep supervising `In review` issues: check review status, handle review comments, verify follow-up evidence, create any missing focused issues discovered during review, and move the reviewed issue to `Done` when truly complete.
11. Report concise progress.
12. Immediately continue to the next issue or spawn the next worker.

Do not stop after one issue.
Do not wait for fresh prompting between issues unless genuinely blocked.

## Route-review rule
Do not stop a route review at the first defect if other independently reachable surface area can still be tested.
Continue through the remaining reachable route surface and create focused issues for each additional broken or uncovered interaction.
Only stop when the remaining coverage is truly blocked by dependency or precondition.

## Required progress report after each completed worker
- Worker type
- Scope / issue
- Spec/Governance touched
- Validation run
- Runtime note (resolved executable path + auth/session state)
- Commit hash/message
- New issues created (if any)
- Next action

## Blocking rule
Only stop and ask for help if:
- requirements are ambiguous
- access/tools are missing
- runtime/environment preconditions cannot be recovered
- validation fails without clear recovery
- multiple reasonable product decisions exist
- no open actionable issues remain

## Completion rule
When no open actionable issues remain, report:
- completed issues
- created follow-up issues
- blocked/deferred items
- explicit statement: `No open actionable issues remain.`
