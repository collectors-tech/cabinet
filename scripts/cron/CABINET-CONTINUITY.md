# Cabinet Continuity Prompt

This file is the source of truth for Cabinet cron-turn behavior in this session.

## Mission
Continue Cabinet work in the same Telegram group session that scheduled the cron. Keep reducing the active Cabinet backlog through direct, visible, tightly bounded execution.

## Hard rules
- Continue work in this same session.
- Do the work directly in this session.
- Do NOT delegate.
- Do NOT use ACP.
- Do NOT use subagents.
- Do NOT use background child sessions.
- Do NOT act as a planner/coordinator instead of doing the work.
- Stay on one current issue until it is closed or clearly blocked with concrete evidence.
- Send exactly one normal visible Telegram update every cron turn.
- Keep each cron turn tightly bounded so it finishes reliably.

## Cabinet-specific execution rules
- All normal Cabinet project rules still apply.
- Follow Issue -> Spec -> Validate -> Commit for implementation work.
- Keep work on one focused issue branch at a time.
- When a cron turn picks a new eligible issue from `Ready`, `Backlog`, or another non-active state, move the linked GitHub Project item Status to `In progress` before starting implementation.
- If project status cannot be updated, record the attempted command/error on the issue and treat the pickup as blocked unless the issue was already active.
- Do not claim completion without concrete evidence.
- For every turn, make bounded, reliable progress that can be seen in the Telegram update.
- If blocked, capture exact blocker evidence, state what was checked, and either advance the same issue to the next concrete unblock step or report the blocker plainly.

## Turn shape
Each cron turn must be small enough to finish reliably. Pick exactly one bounded action that moves the current issue forward, such as:
- inspect one file or one behavior needed for the current issue
- make one focused code/spec/test edit set for the current issue
- run one bounded validation command
- prepare one issue/PR evidence update
- complete one deploy/checkpoint step required by Cabinet rules

Do not chain a large multi-phase plan into a single cron turn. Finish one bounded action, then checkpoint.

## Required Telegram update every turn
Send one normal visible Telegram message every cron turn. It must be a real progress checkpoint, not silent bookkeeping.

Include:
- active issue
- exact bounded action completed this turn
- concrete evidence/result
- next exact bounded action

If blocked, include:
- active issue
- exact blocker
- evidence checked
- next attempted unblock step or why human input is required

## Continuity policy
- Keep working the same current issue across turns until it is closed or blocked with concrete evidence.
- Do not switch issues just to look busy.
- Do not spend a turn only narrating intent.
- Do not end a turn without either evidence of progress or concrete blocker evidence.

## Reliability policy
- Prefer bounded commands and bounded edits.
- Avoid long fragile waits inside a single cron turn.
- If a task is too large for one turn, split it into the next smallest meaningful step and report that step.
- Optimize for reliable completion and visible momentum, not for hidden background activity.

## Response policy
Every cron-triggered response should look like live hands-on work is happening in-session right now.
