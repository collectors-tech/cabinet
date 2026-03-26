# ACP-TROUBLESHOOT.md - Cabinet

Read this file and do only ACP troubleshooting for this Cabinet chat session.

## Purpose
Determine whether this Cabinet chat session can correctly start one bounded ACP worker and read its result back.

## Rules
- Do **not** do Cabinet issue work.
- Do **not** plan backlog work.
- Do **not** use Antfarm.
- Do **not** use local Codex CLI fallback.
- Do **not** use nested ACP coordinator -> ACP child orchestration.
- One ACP worker only.

## Preflight
Before running the test, report these diagnostics first:
- session/chat identifier if known
- repo path
- whether `BOOTSTRAP.md`, `CHILD_BOOTSTRAP.md`, and `RESUME.md` are present
- whether this session is acting as the top-level chat orchestrator
- whether ACP spawn appears available
- whether one bounded ACP worker at a time will be used
- whether local Codex CLI fallback is disabled
- whether nested ACP coordinator -> ACP child orchestration is disabled
- whether any obvious blocker exists before testing

Also write the same preflight into the markdown result file.

## Test
Run the ACP chat loop test from:
`C:\Users\maxbarrass\.openclaw\workspace\ACP-CHAT-LOOP-TEST.md`

## What to report
Report only which of these is true:
- ACP spawn failed
- ACP worker started but result was not readable
- ACP worker returned correctly

Then briefly say what appears broken in this session.

## Write result to markdown
Before you stop, write the diagnosis into:

`C:\projects\collectors-tech\cabinet\ACP-TROUBLESHOOT-RESULT.md`

Include:
- session/chat identifier if known
- outcome
- what you tried
- what happened
- what appears broken in this session
