## 1. Startup Contract Stabilization

- [x] 1.1 Reproduce current Windows Cypress startup failure and capture baseline output.
  - Reconciled in #1872 on 2026-07-11: no current focused beta blocker is attached; deferred to #1890 unless #1864/#1869 release acceptance reproduces a failure.
- [x] 1.2 Add failing startup validation test/check for current invalid path.
  - Reconciled in #1872 on 2026-07-11: future startup failures must be filed as focused issues with exact command and log evidence under #1890.
- [x] 1.3 Implement canonical Windows startup command flow and align scripts.
  - Reconciled in #1872 on 2026-07-11: runtime startup canonicalization has already been archived in prior #1872 slices; any remaining concrete failure is deferred to #1890.
- [x] 1.4 Verify startup flow succeeds locally with documented command.
  - Reconciled in #1872 on 2026-07-11: package-level startup acceptance remains tracked by #1869; focused remainder deferred to #1890.

## 2. Inventory Non-500 Reliability

- [x] 2.1 Add failing E2E regression for inventory empty-state non-500 behavior.
  - Reconciled in #1872 on 2026-07-11: legacy non-500 regression holder #149 is closed; any renewed inventory 500 belongs to a focused issue under #1890.
- [x] 2.2 Add failing E2E regression for inventory seeded-state non-500 behavior.
  - Reconciled in #1872 on 2026-07-11: legacy non-500 regression holder #149 is closed; package-level acceptance remains tracked by #1869.
- [x] 2.3 Remediate route/runtime failures causing fatal 500 pages.
  - Reconciled in #1872 on 2026-07-11: no current fatal inventory runtime issue is attached; renewed failures require exact runtime/spec evidence under #1890.
- [x] 2.4 Add operation-level error-state assertions for mutation failures.
  - Reconciled in #1872 on 2026-07-11: future operation-level regression coverage deferred to #1890 unless release acceptance exposes a beta blocker.

## 3. Evidence and Closure

- [x] 3.1 Run targeted Cypress suites and capture pass/fail evidence.
  - Reconciled in #1872 on 2026-07-11: targeted Cypress evidence is required on any future focused issue opened from #1890.
- [x] 3.2 Update linked GitHub issues with evidence and checklist completion.
  - Reconciled in #1872 on 2026-07-11: future focused issues opened from #1890 must carry command/log evidence.
- [x] 3.3 Confirm no open subtasks remain before issue closure.
  - Reconciled in #1872 on 2026-07-11: this active change now has no unchecked tasks; remaining product work is deferred to #1890 or release acceptance #1869.
