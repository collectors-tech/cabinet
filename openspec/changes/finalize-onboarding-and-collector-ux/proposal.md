## Why

Cabinet has foundational screens but still needs a production-quality first-run experience and consistent collector interactions across dashboard, inventory, and wishlist workflows.

## What Changes

- Complete onboarding wizard flow from identity through first usable workspace.
- Standardize collector interactions (row click details, thumbnail lightbox, bulk mode behavior).
- Finalize dashboard action model around "what needs attention now."
- Define acceptance and regression expectations for these high-use workflows.

## Capabilities

### New Capabilities

- `onboarding-wizard-completion`: End-to-end starter flow with identity, starter data, first item, and preferences.
- `collector-workspace-interactions`: Standard interaction model for inventory and wishlist list/grid/detail behaviors.
- `dashboard-attention-center`: Action-driven dashboard requirements for discovery, pricing, and health signals.

### Modified Capabilities

- None.

## Impact

- Affected code: onboarding components/routes, dashboard screen, inventory/wishlist interactions, shared UI primitives.
- Affected tests: onboarding and collector E2E suites.
- Related issues: `#154` and subsequent UX completion tickets from `docs/APP_COMPLETION_ANALYSIS.md`.
