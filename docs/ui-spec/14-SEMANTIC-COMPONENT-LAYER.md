# 14 - Semantic Component Layer (Build Contract)

## Purpose
Define the semantic UI component layer needed to build the target Cabinet experience with composable, testable parts.

This is the implementation contract for moving from one large page file to reusable UI primitives.

## Principles
1. Layout-first semantics: stable shell and clear region ownership.
2. Action-first workflow: primary actions are always visible near the relevant data.
3. Progressive disclosure: advanced/technical controls are secondary.
4. Single responsibility components: each component owns one concern.
5. Testable boundaries: each semantic region has deterministic selectors and state model.

## Layer Model
```text
L0: Tokens + primitives
L1: App shell components
L2: Workspace components (home, collection, scanner, etc.)
L3: Domain blocks (cards, grids, toolbars, forms)
L4: Interaction overlays (dialogs, drawers, command palette/chat)
```

## L0 - Tokens + Primitive Components
Use these as the only shared visual primitives.

1. `UiPage`
- Root content wrapper for one route/screen.
- Owns vertical spacing rhythm.

2. `UiSectionCard`
- Standard content surface.
- Variants: `default | elevated | muted | danger`.

3. `UiButton`
- Variants: `primary | secondary | ghost | danger`.
- Sizes: `sm | md | lg`.

4. `UiInput`, `UiSelect`, `UiSearchInput`
- Form/search primitives with consistent focus/label/error handling.

5. `UiBadge`, `UiStat`, `UiChip`
- Lightweight metadata/status primitives.

6. `UiEmptyState`, `UiErrorState`, `UiLoadingState`
- Unified UX states across screens.

## L1 - App Shell Components
These replace ad-hoc shell markup.

1. `AppShell`
```text
main.app-shell
  AppPrimaryNav
  AppContextPane
  AppContent
  AppChatRail?
```

2. `AppPrimaryNav`
- Global navigation (fixed rail).
- Includes: nav items, collapse, reorder/hide editor, version/build metadata.
- Must not own page content scroll.

3. `AppContextPane`
- Collection/folder context tree and quick filtering.
- Collapsible and independently scrollable.

4. `AppPageHeader`
- Sticky content header.
- Contains only:
  - page title
  - page-scoped quick actions
- Does not show diagnostics/runtime internals.

5. `AppContentScroll`
- Sole scroll owner for main body content.

## L2 - Workspace Components

## 2.1 Home Workspace (Action Board)
1. `HomeCommandCenter`
- Container for top-level home layout.

2. `HomeHero`
- Eyebrow + headline + short summary.
- Shows next recommended action.

3. `HomeQuickActions`
- Strict CTA set:
  - `Add Item`
  - `Run Scanner Now`
  - `Open Discover`
  - `Backup Now`

4. `HomeAttentionBoard`
- Priority card grid for actionable signals.
- Child: `AttentionCard`.

5. `HomeSnapshotGrid`
- KPI summary row/cards.

6. `HomeActionQueue`
- Ordered list of pending tasks with direct deep-links.

State model:
- Loading: skeleton on hero/snapshot/attention.
- Empty: calm state with one primary CTA.
- Error: inline retry panel.

## 2.2 Collection Workspace (Core)
```text
CollectionWorkspace
  CollectionCommandRow
  CollectionSummaryStrip
  CollectionBrowser
    CollectionTreePane
    CollectionResultsPane
```

1. `CollectionCommandRow`
- Contains:
  - `Add Item`
  - `Add Folder`
  - collection search
  - sort selector
  - view toggle
  - summarize toggle

2. `CollectionSummaryStrip`
- `Folders`, `Items`, `Total Quantity`, `Total Value`.

3. `CollectionTreePane`
- Folder search + hierarchical tree + utility actions (`History`, `Trash`).

4. `CollectionResultsPane`
- `CollectionResultsToolbar`
- `CollectionResultsGrid` and/or `CollectionResultsTable`
- `CollectionPagination`

5. `CollectionSelectionBar`
- Appears only when selection count > 0.
- Drives bulk operations.

6. `CollectionBulkEditPanel`
- Preview + confirm + apply workflow.

7. `CollectionItemEditor`
- Inline quick edit and detail edit drawer/modal.

State model:
- Loading: skeleton rows/cards.
- Empty: friendly CTA to add item/folder.
- Error: retry + error detail.

## L3 - Domain Blocks
Reusable blocks inside multiple workspaces.

1. `CommandToolbar`
- Horizontal action + filter row.

2. `SummaryStrip`
- compact metrics row.

3. `EntityCardGrid`
- Card-based item/folder layout.

4. `EntityTable`
- Dense tabular mode for bulk ops.

5. `StatusPanel`
- Health/diagnostic summary card (non-primary by default).

6. `MediaPicker`
- Upload, camera, preview, reorder, set primary.

7. `ProviderHealthBadge`
- Scanner/source health indicator.

## L4 - Overlays
1. `ConfirmDialog`
- Destructive confirmation and bulk apply guardrails.

2. `ItemDetailDrawer`
- In-context edit without route jump.

3. `GlobalCommandPalette` (future)
- Quick navigation and command execution.

4. `ChatCopilotRail`
- Context-aware assistant with explicit open/close and bounded actions.

## Accessibility Contract (Mandatory)
1. Landmarks:
- `main`, `nav`, `aside`, `header`, labeled `section[role=region]` as needed.
2. Buttons:
- Action labels must be verb-first and unique within visible region.
3. Inputs:
- Every input has associated label or `aria-label`.
4. Tables:
- Use semantic `table/thead/tbody/th`.
5. State messaging:
- Use `role=status`/`role=alert` appropriately for async updates/errors.

## State Contract Per Major Component
Each L2 component must expose these states in deterministic order:
1. `loading`
2. `empty`
3. `error`
4. `ready`

## Shadcn/Radix Mapping Plan
Use shadcn/Radix as implementation base for primitives; keep semantic names above them.

1. `UiButton` -> shadcn `Button`
2. `UiSectionCard` -> shadcn `Card`
3. `UiInput/UiSelect` -> shadcn `Input/Select`
4. `ConfirmDialog` -> Radix/shadcn `Dialog`
5. `ItemDetailDrawer` -> Radix `Dialog` (sheet pattern)
6. `Tabs/ViewToggle` -> Radix `Tabs` or shadcn `Tabs`
7. `Command` surfaces -> shadcn `Command` where suitable

Rule:
- Business/screen code imports semantic wrapper components, not raw primitive library components directly.

## Suggested File Structure
```text
web/src/components/semantic/
  shell/
    AppShell.tsx
    AppPrimaryNav.tsx
    AppContextPane.tsx
    AppPageHeader.tsx
  home/
    HomeCommandCenter.tsx
    HomeQuickActions.tsx
    HomeAttentionBoard.tsx
    HomeSnapshotGrid.tsx
    HomeActionQueue.tsx
  collection/
    CollectionWorkspace.tsx
    CollectionCommandRow.tsx
    CollectionSummaryStrip.tsx
    CollectionTreePane.tsx
    CollectionResultsPane.tsx
    CollectionSelectionBar.tsx
    CollectionBulkEditPanel.tsx
```

## Delivery Sequence
1. Extract L0 primitives.
2. Extract L1 shell (no behavior changes).
3. Extract Home L2 components.
4. Extract Collection L2 components.
5. Move remaining screens into same pattern.

## Definition of Done (for each extracted component)
1. Component has explicit props contract.
2. Component has loading/empty/error/ready handling if data-bound.
3. Unit test covers main rendering and primary action.
4. E2E path still passes for affected workflow.
5. No regression to scroll ownership or sticky header behavior.
