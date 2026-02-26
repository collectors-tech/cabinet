# Agent Instructions for `cabinet`

This file defines the default working rules for coding agents in this repository.

## Rule Source of Truth

All rules in `.cursor/rules/*.mdc` are adopted as active instructions for this repo.  
If there is any conflict, precedence is:

1. System/developer/runtime constraints
2. This `AGENTS.md`
3. `.cursor/rules/*.mdc`

## Adopted Rule Set

Apply and follow these files:

- `.cursor/rules/gov-01-instructions.mdc`
- `.cursor/rules/gov-02-workflow.mdc`
- `.cursor/rules/gov-03-communication.mdc`
- `.cursor/rules/gov-04-quality.mdc`
- `.cursor/rules/gov-05-testing.mdc`
- `.cursor/rules/gov-06-issues.mdc`
- `.cursor/rules/gov-07-tasks.mdc`
- `.cursor/rules/gov-08-agent-verification.mdc`
- `.cursor/rules/proj-01-ticket-and-testing.mdc`
- `.cursor/rules/proj-02-test-first-no-remediation-without-test.mdc`

## Mandatory Behaviors (Operational Summary)

- Track all implementation work with a repository issue (`collectors-tech/cabinet`) using `gh`.
- Reuse existing issue scope when possible; create a new issue before code changes if none exists.
- Follow test-first workflow: create a failing test before remediation/feature implementation.
- Run tests/checks in-session and report only actual results.
- Do not hand off verification to the user unless genuinely blocked.
- Keep issue subtasks/checklists updated and only close when all acceptance criteria/subtasks are complete.
- Commit with issue reference first: `#<issue-number> <type>(<scope>): <description>`.
- Do not mark done until validation/tests pass, commit+push is complete, and issue/board are updated per rules.

## Notes

- This file is intended to make rule discovery explicit for agents that prioritize `AGENTS.md`.
- Rule content remains maintained in `.cursor/rules/`; update those files as the canonical policy documents.
- OpenSpec migration source-of-truth docs:
  - `openspec/specs/documentation-governance/spec.md`
  - `openspec/specs/README.md`
