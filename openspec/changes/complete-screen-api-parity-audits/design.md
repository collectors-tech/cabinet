## Context

Cabinet has a strong API foundation, but parity risk remains when routed screens are scaffolded from templates and not fully connected to runtime data/mutations. This change formalizes parity audits as a repeatable delivery standard.

## Goals / Non-Goals

**Goals:**
- Ensure Users, Settings, and Chat screens are fully API wired.
- Eliminate production-path sample/static placeholders.
- Ensure parity checks are automatable and linked to issue acceptance.

**Non-Goals:**
- Introduce new business features beyond parity completion.
- Redesign visual system or navigation taxonomy.

## Decisions

1. Parity is defined at route+endpoint level.
   - Rationale: prevents "mostly connected" ambiguity.
   - Alternative: component-level assertions only. Rejected due to poor traceability.

2. Each parity issue requires test evidence before closure.
   - Rationale: removes manual-only verification gaps.
   - Alternative: screenshots/manual checks only. Rejected as insufficient for regression control.

3. Placeholder content is treated as release-blocking on production routes.
   - Rationale: avoids user confusion and confidence loss.
   - Alternative: leave placeholders until later hardening. Rejected due to recurring drift.

## Risks / Trade-offs

- [Risk] Tight parity checks may initially surface many failures.  
  Mitigation: phase in route groups with explicit issue sequencing.

- [Risk] Additional test coverage increases execution time.  
  Mitigation: maintain targeted suites plus full regression gate.

## Migration Plan

1. Audit Users, Settings, Chat route by route against API map.
2. Add failing tests for missing/incorrect bindings.
3. Implement wiring and remove placeholder paths.
4. Validate parity matrix and close issues with evidence.

Rollback:
- Revert route-level changes per screen if a regression is introduced; preserve failing tests to retain visibility.

## Open Questions

- Should parity checks be codified in a generated contract test from `docs/UI_ENDPOINT_PARITY.md`?
- Should UI build fail on detected placeholder copy markers?
