## Context

Cabinet already stores flexible inventory category options in profile settings and exposes grading enums through `/api/inventory/grading/enums`. The current model treats condition values as a flat list, which breaks down when different collector domains use different grading language.

## Goals / Non-Goals

**Goals:**
- Keep Category as flexible multi-select metadata.
- Add Item Type as the single source used to select a condition scale.
- Store item type condition scales in profile settings so demo/local profiles can evolve independently.
- Seed sensible Slot Cars and Trading Cards defaults.
- Reuse existing settings/profile APIs where practical.
- Restore Inventory Category and Condition filters using the existing compact filter control style.

**Non-Goals:**
- Redesign canonical item persistence beyond the minimum item type field needed by Inventory UI.
- Replace all legacy grading fields in one pass.
- Build valuation workflows that infer condition from photos or market data.
- Force migrate existing category values into item types automatically.

## Decisions

### 1. Item Type is separate from Category
Item Type is single-select because a condition scale needs one deterministic source. Category remains multi-select/free-add because collectors use categories more like flexible labels.

Alternative considered: use the first selected category to drive condition. Rejected because category is intentionally multi-select and would create ambiguous condition choices.

### 2. Condition scales are profile settings first
Store type definitions and condition values in profile settings as JSON, alongside existing category options and grading defaults. This keeps the first implementation low-risk and avoids schema churn while the taxonomy model is still settling.

Alternative considered: create normalized database tables immediately. Rejected for this slice because the current product already manages user-facing inventory lists through profile settings and the feature needs UI confidence before schema hardening.

### 3. Existing grading enum API remains compatible
The existing flat condition list remains available for older callers. New item-type scale data can be returned by the same grading/taxonomy surface so forms can scope options without breaking current tests or API consumers.

Alternative considered: replace `/api/inventory/grading/enums`. Rejected because several screens/tests already rely on it.

### 4. Unknown item types fall back safely
If an item has no item type or references a removed type, forms SHALL fall back to the configured default item type or the flat condition list. Existing records must remain editable.

## Risks / Trade-offs

- [Settings JSON grows complex] -> keep the payload small and validate/normalize server-side.
- [Existing items have no item type] -> default to a configured or seeded type while still displaying current condition text.
- [Category and item type labels look similar] -> label UI explicitly and keep Category multi-select while Item Type is single-select.
- [Filter semantics drift] -> use shared compact filter controls and test expected filtering behavior.

## Migration Plan

1. Add OpenSpec requirements for item type condition scales and inventory filter/form behavior.
2. Add failing API/UI tests for settings persistence and scoped condition choices.
3. Extend grading/taxonomy settings payloads with item type condition scales and defaults.
4. Wire Inventory item forms and filters to item type scoped condition choices.
5. Roll forward on `develop`; rollback is reverting the UI/settings change because persisted settings are additive JSON values.
