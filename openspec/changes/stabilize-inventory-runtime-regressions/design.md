## Context

Cabinet currently has broad API coverage, but UI reliability regressions are still reaching runtime. Inventory route failures and unstable Cypress startup reduce confidence in day-to-day delivery and make issue verification inconsistent across Windows developer environments.

## Goals / Non-Goals

**Goals:**
- Eliminate known inventory/wishlist 500 regressions.
- Make Cypress startup deterministic on Windows.
- Enforce regression tests for known defect classes before issue closure.

**Non-Goals:**
- Large visual redesign of inventory/wishlist pages.
- New inventory feature expansion unrelated to reliability.

## Decisions

1. Treat runtime reliability as a first-class capability.
   - Rationale: existing features are less valuable if critical routes fail.
   - Alternative considered: defer reliability until after new features. Rejected due to repeated user-impacting regressions.

2. Add explicit non-500 regression tests mapped to issue IDs.
   - Rationale: direct prevention of recurrence for known failures.
   - Alternative considered: rely on manual smoke testing. Rejected due to inconsistent reproducibility.

3. Standardize Cypress startup command path on Windows.
   - Rationale: local and CI must share deterministic startup semantics.
   - Alternative considered: use separate platform-specific manual procedures. Rejected due to maintenance overhead.

## Risks / Trade-offs

- [Risk] Tightening startup rules may require script refactoring across existing commands.  
  Mitigation: keep compatibility wrappers and document canonical command.

- [Risk] Additional E2E regressions increase runtime in CI.  
  Mitigation: tag targeted suites and run focused subsets on issue scope, full regression on merge.

## Migration Plan

1. Fix startup contract and verify on Windows.
2. Add failing non-500 tests for known inventory cases.
3. Implement remediations to turn tests green.
4. Run affected E2E + regression subset and document evidence in issue closure.

Rollback:
- Revert startup script/config changes and test files if failures are systemic; keep issue open until a stable contract is restored.

## Open Questions

- Should Cypress startup be enforced via a single repo script alias used by both local and CI runners?
- Should inventory non-500 suite be promoted to mandatory pre-merge checks for all UI PRs?
