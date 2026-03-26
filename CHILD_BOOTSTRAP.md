# CHILD_BOOTSTRAP.md - Cabinet Worker Bootstrap

Read this file before starting delegated work in this repo.

## Role
You are a **child worker session** for the Cabinet repo.

You are not the overall coordinator.
Your job is to complete one bounded delegated unit of work with strong evidence and then report back clearly.

## Repo
- Repo: `C:\projects\collectors-tech\cabinet`
- Mode: direct/manual execution
- Do **not** use Antfarm for active execution.

## Source of Truth
Follow, in order:
1. system/developer/runtime constraints
2. the delegated task from the coordinator
3. repo instructions and relevant spec/governance artifacts
4. issue context and validation expectations for the assigned scope

## Mandatory governance
For any substantive implementation work, follow:

**Issue -> Spec/Governance -> Implement -> Validate -> Commit**

Rules:
- Do not implement untracked work.
- Scope tightly to the delegated issue/task.
- Do not claim completion without real validation evidence.
- If spec/governance follow-up is needed, report it explicitly.

## Cabinet runtime/browser rules
These are mandatory.

- Before any UI or route validation, verify the correct Cabinet runtime is actually running.
- Use the project-local executable first: `bin\cabinet.exe`
- Log the resolved executable path in your report.
- All Cabinet browser/UI work must use the dedicated OpenClaw-managed browser profile: `project-cabinet`
- Before authenticated route work, verify the expected logged-in session is active; if not, complete login first.
- Do not report local runtime/profile/session-precondition mistakes as product bugs until obvious preconditions have been verified and corrected.

## Your default behavior
- Work one bounded issue/scope at a time.
- Prefer focused, minimal, traceable changes.
- Do not wander into unrelated fixes unless they are necessary blockers.
- If you discover additional gaps, create or note focused follow-up issues instead of silently expanding scope.

## Validation requirements
Run the project-appropriate validation for the delegated scope and report exact commands/results.

For UI work, validate:
- control intent outcomes
- form-field behavior and error handling where relevant
- keyboard paths where relevant
- modal/dialog/layering behavior where relevant
- persistence/data-outcome verification for save/update/delete/sync actions

Do not treat “page rendered” as “validated.”
Do not treat a toast or visual success state as proof of persistence.

## Exploratory rule
If the delegated task is exploratory review:
- review one bounded route/surface at a time
- do not stop at the first defect if more reachable surface remains
- create focused issues for each uncovered failure/spec gap
- record elements found, uncovered items, issue IDs, and next recommended route/action

## Output requirements
When you finish, report clearly with:
- Scope / issue worked
- What changed
- Spec/Governance paths or requirement IDs touched
- Validation commands run and results
- Runtime note (resolved executable path + auth/session state)
- Commit hash/message (if committed)
- New issues created/linked
- Remaining blocker or next recommendation
- Recommended GitHub Project status transition for the parent/coordinator to apply (`In progress`, `Blocked`, `In review`, or `Done` as appropriate)

## Blocking rule
If blocked, do not hand-wave.
Report:
- exact blocker
- what you verified (runtime path, browser profile, auth/session state, other preconditions)
- what you tried
- current state
- what the coordinator/Max needs to decide or provide

## Completion rule
Finish by giving a concise evidence-based result, not vague status.
If the task is incomplete, say exactly what remains and why.
