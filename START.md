# START.md - Cabinet Direct Delivery Start

Read this file and `BOOTSTRAP.md`, then start work.

## Execution mode
This file is an execution instruction, not a summary task.

After reading it:
- do **not** summarize it
- do **not** acknowledge it
- begin executing immediately
- keep chat updates visible at meaningful steps

## Active mode
**Development delivery mode.**

Use this for normal Cabinet issue work where the required flow is:

`Issue -> Spec/Governance -> Implement -> Validate -> Commit`

Do not switch to exploratory-only mode unless that is explicitly requested.
Do the work directly in this session.

## Required chat behavior
Keep the user updated with concrete progress, especially when:
- selecting or resuming an issue
- reproducing the problem
- editing code/spec/docs
- validating
- hitting a blocker
- deploying/recycling the demo lane
- moving to the next issue

Avoid vague filler.
Good updates are short and specific.

## Working loop
1. Review backlog/current state and pick exactly one open actionable issue.
2. Prefer a GitHub Project item already in `Ready` when available.
3. Update persistent GitHub state as work begins:
   - move/update the GitHub Project item to `In progress`
   - apply/update durable labels/tags where needed
   - add/update an issue comment claiming the work
4. Announce in chat: `Starting issue #<id> - <title>`
5. Bind/update the relevant spec/governance requirements before code changes.
6. Create/use one focused issue branch for that issue.
7. Implement only the scoped issue work.
8. Run required validation for the touched scope.
9. Commit with an issue-prefixed commit message and push.
10. Update the issue with concrete evidence.
11. Merge validated issue branches into `develop`.
12. Deploy/recycle the demo/review lane from `develop` unless Max explicitly says not to.
13. Report the deployed branch and commit hash in chat.
14. Update issue/project state honestly (`Blocked`, `In review`, `Done`, etc.).
15. Continue to the next issue unless genuinely blocked.

## Cabinet rules
- Verify Cabinet runtime before UI/route work.
- Use project-local `bin\cabinet.exe`.
- Use browser profile `project-cabinet`.
- Verify auth/session before authenticated work.
- Do not treat local runtime/profile/session mistakes as product bugs until preconditions are verified.

## Enforcement
Follow:

`Issue -> Spec/Governance -> Implement -> Validate -> Commit`

One active issue should keep one focused branch and one bounded implementation scope.
If follow-up work is discovered, create a focused issue instead of silently expanding scope.

## Failure handling
If a real step fails, say so clearly in chat.
Examples:
- runtime precondition failed
- validation failed
- repo is blocked
- branch/deploy follow-through failed

Do not pretend work is progressing if it is blocked.
