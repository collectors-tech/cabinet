## 1. Contract

- [x] 1.1 Define isolated WSL runner-pool requirements for issue #1949.
- [x] 1.2 Add a failing repository contract test for installer and runbook.

## 2. Implementation

- [x] 2.1 Add the Cabinet-specific WSL runner-pool installer.
- [x] 2.2 Add secure provisioning and lifecycle documentation.

## 3. Evidence

- [x] 3.1 Run the targeted contract test.
  - `node --test scripts/*.test.mjs`: 16 passed.
- [x] 3.2 Parse the PowerShell installer and run a non-mutating `-WhatIf` plan.
  - AST parse passed; the three-member plan resolved `cabinet`,
    `cabinet-02`, and `cabinet-03` without creating a distribution.
- [x] 3.3 Run strict OpenSpec validation and repository-safe regression checks.
  - Strict OpenSpec: 11 passed. Pre-push Go API docs smoke: passed.
- [x] 3.4 Commit, push, and open a pull request to `develop`.
  - Commit `eeae83ad`; pull request `#1950`.

## 4. Organisation runner groups

- [x] 4.1 Add the selected-repository organisation runner-group contract.
- [x] 4.2 Add repository-default and organisation-scope installer paths.
- [x] 4.3 Document permissions, public-repository opt-in, and safe migration.
- [x] 4.4 Validate the contract test, PowerShell and embedded Bash syntax,
  repository and organisation `WhatIf` plans, and strict OpenSpec checks.
