## 1. Specification

- [x] 1.1 Define Agent Skill Registry identity, source, status, safety, permissions, and binding requirements.
- [x] 1.2 Define local skill archive structure, manifest fields, validation, import result states, and safe failure requirements.
- [x] 1.3 Define installed skill enable/disable lifecycle and Skills page list/detail/import behavior.
- [x] 1.4 Define marketplace-deferred boundary so local import work does not imply public discovery, publishing, ratings, payments, or remote execution.

## 2. Traceability

- [x] 2.1 Link #1667 to planned registry/import/UI/docs implementation issues.
- [x] 2.2 Add planned verification targets for #1668, #1669, #1670, #1671, and #1672.

## 3. Validation

- [x] 3.1 Run `openspec validate --changes --strict --no-interactive`.
- [x] 3.2 Capture validation output under `.work-agent/logs/issue-1667/`.
- [x] 3.3 Commit and push the OpenSpec change set.
