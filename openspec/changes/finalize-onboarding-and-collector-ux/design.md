## Context

The app has route structure and baseline components, but onboarding and core collector interactions are not fully standardized. This creates inconsistency, slows new users, and increases regression risk.

## Goals / Non-Goals

**Goals:**
- Make first-run onboarding clear and short.
- Ensure inventory/wishlist interactions are predictable and performant.
- Make dashboard the primary action surface for daily collector operations.

**Non-Goals:**
- Rebuild all visual styling from scratch.
- Introduce non-essential advanced customization during this phase.

## Decisions

1. Use a guided onboarding sequence with explicit completion gates.
   - Rationale: prevents immediate overwhelm and aligns with first-value UX.
   - Alternative: expose full forms at first login. Rejected due to complexity for new users.

2. Adopt a split interaction model for list rows and thumbnails.
   - Rationale: row click opens details, thumbnail opens lightbox carousel, bulk via checkboxes only.
   - Alternative: single-click behavior for all interactions. Rejected due to accidental actions.

3. Keep dashboard focused on actionable attention cards.
   - Rationale: dashboard should guide immediate work, not display static filler metrics.
   - Alternative: broad overview-only dashboard. Rejected due to low operational value.

## Risks / Trade-offs

- [Risk] Standardizing interactions may require refactoring current route components.  
  Mitigation: introduce shared handlers/primitives and migrate per screen.

- [Risk] Onboarding gate logic can conflict with returning-user paths.  
  Mitigation: explicit state machine and persisted completion checkpoints.

## Migration Plan

1. Define onboarding state model and completion checks.
2. Implement inventory/wishlist interaction standards with tests.
3. Refactor dashboard cards/actions to attention-center model.
4. Validate end-to-end journey from first run to active workspace.

Rollback:
- Feature-flag onboarding/workspace interaction changes to allow targeted rollback while preserving data.

## Open Questions

- Should onboarding completion state be editable/resettable in settings for support workflows?
- Should dashboard card ordering be static in v1 or user-configurable in v1.1?
