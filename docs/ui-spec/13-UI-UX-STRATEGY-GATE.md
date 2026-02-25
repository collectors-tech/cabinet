# 13 - UI/UX Strategy Gate (Strict)

## Purpose
Convert UI/UX direction into release gates that must pass before a screen is considered done.

## Scope
Applies to all Cabinet desktop screens, with strict priority on `Home` and `Collection`.

## Mandatory Gates
1. Information hierarchy gate
- One primary outcome per screen above the fold.
- Secondary/technical actions must be visually and semantically de-prioritized.
- No mixed-purpose hero blocks with more than 4 primary actions.

2. Action clarity gate
- Primary action row must be visible without scrolling on desktop.
- Labels must be task language, not implementation language.
- Each critical status signal includes at least one direct action.

3. Layout behavior gate
- Left navigation remains fixed.
- Page header remains sticky.
- Body content owns vertical scroll.
- No clipping of footer/meta controls inside side rails.

4. Progressive disclosure gate
- Diagnostic and admin-heavy controls are not in default first-run surface.
- Advanced tools are behind dedicated screens/expanders, not mixed with first action path.

5. Test gate
- Each screen has:
  - one structure test (semantic regions),
  - one primary action test (happy path),
  - one state test (empty/loading/error).
- PR cannot close issue scope unless targeted tests pass in-session.

## Sortly Alignment Checklist (Used in Reviews)
1. Page title + command row are obvious and stable.
2. Search and view controls are colocated with item results.
3. Summary strip appears before result content.
4. Action wording is concise and operational.
5. Visual density is progressive, not uniformly heavy.

## Required Review Output for Each Remediation Wave
1. Before/after screenshot references.
2. Gap list mapped to this gate checklist.
3. Test evidence for new/changed behavior.
4. Follow-up issues for unresolved gaps.

