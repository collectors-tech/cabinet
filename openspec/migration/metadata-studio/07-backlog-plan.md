# 07 — Backlog plan for Metadata Studio

Status: planning draft  
Area: backlog planning  
Purpose: convert Metadata Studio into issue-ready epics and implementation slices for `collectors-tech/cabinet`.

## Delivery principle

Follow Cabinet's repo workflow:

```text
Issue -> OpenSpec -> implementation -> validation -> evidence -> PR -> develop
```

Do not claim completion without real validation evidence.

## Backlog strategy

Metadata Studio is too large for one issue. It should be split into a set of small, testable issues.

Recommended phases:

1. Specification and scope guardrails.
2. Data model and API foundation.
3. Item/specimen field values.
4. Template registry.
5. UI field editor.
6. Form/detail/table rendering.
7. Import/export preservation.
8. Search/filter/saved views.
9. Reports.
10. Starter templates.
11. Polish and migration helpers.

## Epic 1 — Metadata Studio foundation

### Goal

Create the underlying model for custom field definitions and typed values.

### Issues

#### Issue: Define OpenSpec for custom field definitions

Scope:

- field definition schema
- field value schema
- supported initial field types
- field lifecycle
- privacy/export flags
- profile scoping

Acceptance:

- OpenSpec validates strictly.
- Spec names field definition and value behaviours.
- Spec distinguishes custom metadata from core Cabinet concepts.

#### Issue: Add custom field definition storage

Scope:

- database migration
- profile-scoped field definitions
- field lifecycle state
- stable key and label
- type and validation metadata

Acceptance:

- Create/read/update/list field definitions via storage layer.
- Soft-delete/deprecate supported.
- Tests prove persistence and reload.

#### Issue: Add typed custom field values on specimens/items

Scope:

- field values table/model
- entity scope and entity ID
- typed value envelope
- validation against field definition

Acceptance:

- Values can be saved against item/specimen.
- Type mismatch is rejected.
- Values reload correctly.
- Deleting a field does not silently delete values.

#### Issue: Add custom field definition API

Scope:

- list/create/update/deprecate custom field definitions
- profile-scoped routes
- validation errors

Possible API shape:

```text
GET    /api/profiles/{profileID}/metadata/fields
POST   /api/profiles/{profileID}/metadata/fields
PUT    /api/profiles/{profileID}/metadata/fields/{fieldID}
DELETE /api/profiles/{profileID}/metadata/fields/{fieldID}
```

Acceptance:

- API docs update.
- Tests cover CRUD and validation failures.
- Invalid type/options rejected.

#### Issue: Add custom field value API for items/specimens

Possible API shape:

```text
GET /api/items/{itemID}/metadata-values
PUT /api/items/{itemID}/metadata-values
GET /api/items/{itemID}/instances/{instanceID}/metadata-values
PUT /api/items/{itemID}/instances/{instanceID}/metadata-values
```

Acceptance:

- Values persist through API.
- Values are profile-scoped.
- Invalid field IDs rejected.
- Type validation works.

## Epic 2 — Metadata Studio UI

### Goal

Add a configuration UI where users can create fields, manage field options, and control metadata behaviour.

### Issues

#### Issue: Add Metadata Studio navigation shell

Scope:

- Settings/Metadata Studio page
- tabs: Field Library, Templates, Layouts, Reports, Import Mappings
- empty states

Acceptance:

- UI page is discoverable.
- Empty state explains purpose.
- No broken navigation.

#### Issue: Add field library table

Scope:

- list configured fields
- show label, key, type, scope, status, usage count
- create/edit/deprecate actions

Acceptance:

- Field list loads from API.
- User can see active/deprecated fields.
- Usage count or placeholder shown safely.

#### Issue: Add field editor form

Scope:

- label
- key
- type
- scope
- description/help text
- required
- default value
- section
- searchable/sortable/filterable/reportable/exportable toggles
- privacy/default visibility

Acceptance:

- Create and edit field definitions.
- Type-specific settings render correctly.
- Save persists and reloads.
- Validation errors are visible.

#### Issue: Add enum/multi-select option editor

Scope:

- add/edit/remove/reorder options
- stable option keys
- labels and optional colour/severity

Acceptance:

- Options persist.
- Renaming label does not change stored option key.
- Removing in-use options requires confirmation or deprecation.

## Epic 3 — Category templates

### Goal

Package fields, layouts, views, and reports into reusable templates.

### Issues

#### Issue: Define OpenSpec for category templates

Scope:

- template model
- template versions
- fields association
- sections/layout defaults
- system vs user templates

Acceptance:

- OpenSpec validates strictly.
- Template lifecycle defined.
- Existing item safety defined.

#### Issue: Add template storage and API

Possible API shape:

```text
GET  /api/profiles/{profileID}/metadata/templates
POST /api/profiles/{profileID}/metadata/templates
PUT  /api/profiles/{profileID}/metadata/templates/{templateID}
POST /api/profiles/{profileID}/metadata/templates/{templateID}/clone
```

Acceptance:

- System/user templates list correctly.
- User templates can be created and updated.
- System templates cannot be edited directly.
- Clone works.

#### Issue: Add trading card starter template

Acceptance:

- Template includes set, card number, variant, language, grading, condition, certificate, location/display defaults.
- Template can be used to create an item/specimen.
- Template can be exported as schema.

#### Issue: Add slot car starter template

Acceptance:

- Template includes brand, model, scale, livery, boxed state, box condition, technical setup, maintenance/restoration metadata.
- Template can be used to create an item/specimen.

#### Issue: Add gallery object starter template

Acceptance:

- Template includes artist, title/date, medium, dimensions, provenance, condition report, conservation, appraisal/insurance, loan/display status.
- Sensitive fields default private.

## Epic 4 — Form, detail, card, and table layout configuration

### Goal

Use templates/fields to control item create/edit/detail/card/table surfaces.

### Issues

#### Issue: Define OpenSpec for metadata-driven form layouts

Acceptance:

- Sections, sort order, basic/advanced mode, default values defined.

#### Issue: Render item create/edit form from template sections

Acceptance:

- Selecting template renders configured fields.
- Required fields validate.
- Values save and reload.
- Advanced mode toggles advanced fields.

#### Issue: Render item detail metadata sections

Acceptance:

- Detail panel/page groups fields by configured section.
- Empty values are handled cleanly.
- Deprecated fields are hidden unless requested.

#### Issue: Add card subtitle and badge configuration

Acceptance:

- Template can define subtitle fields and badge fields.
- Inventory cards show configured fields.
- Missing or hidden fields do not break layout.

#### Issue: Add table column configuration for custom fields

Acceptance:

- User can show/hide custom field columns.
- Column config persists.
- Custom field values render correctly.

## Epic 5 — Import/export and migration

### Goal

Make metadata migration safe and complete.

### Issues

#### Issue: Define OpenSpec for metadata import mapping

Acceptance:

- Unknown column preservation defined.
- Type inference rules defined.
- Dry-run/apply workflow defined.

#### Issue: Add import batch model

Acceptance:

- Import batch stores source file info, mapping, created/updated records, warnings/errors.
- Import batch can be queried after apply.

#### Issue: Add CSV analyser for unknown/custom columns

Acceptance:

- CSV columns are detected.
- Known core fields are suggested.
- Unknown fields are proposed as custom fields.
- Type inference suggestions are produced.

#### Issue: Add import mapping UI

Acceptance:

- User can map columns to core fields, existing custom fields, new custom fields, ignore.
- User can change inferred type.
- Mapping persists for dry run/apply.

#### Issue: Add metadata import dry run

Acceptance:

- Shows creates/updates/skips.
- Shows fields to create.
- Shows invalid values.
- Shows duplicate candidates.

#### Issue: Apply CSV import with custom field creation

Acceptance:

- Unknown columns can become field definitions.
- Values are stored on imported items/specimens.
- Import batch records created field/value IDs.

#### Issue: Add Homebox CSV import preset

Acceptance:

- Recognises `HB.import_ref`, `HB.location`, `HB.label`, `HB.field.*`.
- Maps `HB.field.*` columns to custom field proposals.
- Dry-run report shows mapping.

#### Issue: Export custom schema with data

Acceptance:

- JSON export includes custom field definitions, values, templates, and layouts.
- CSV export includes custom field columns.
- Export can be re-imported without losing custom metadata.

#### Issue: Plan media bundle export/import

Acceptance:

- Media manifest format defined.
- Export options for originals/thumbnails defined.
- Missing media import warnings defined.

## Epic 6 — Search, filters, saved views, and reports

### Goal

Make configured metadata operational.

### Issues

#### Issue: Define OpenSpec for custom field query behaviour

Acceptance:

- Searchable/sortable/filterable/reportable behaviour defined per type.

#### Issue: Add custom field query projection

Acceptance:

- Custom field values are indexed/projected for search/filter.
- Reindex command rebuilds projection.

#### Issue: Add custom fields to inventory search

Acceptance:

- Search can find items by searchable custom text fields.
- Tests prove field values are included/excluded based on flag.

#### Issue: Add typed custom field filters

Acceptance:

- Boolean, enum, text, number, money, date filters work.
- Invalid operator/type combinations are rejected.

#### Issue: Add saved views with custom fields

Acceptance:

- Saved view can store custom field filters and columns.
- Saved view reloads and applies correctly.

#### Issue: Add missing metadata report

Acceptance:

- Report identifies missing required/configured fields.
- User can filter by template/category.
- Report can export CSV.

#### Issue: Add insurance/value report prototype

Acceptance:

- Uses money fields and location/category grouping.
- Sensitive/private defaults are respected.

#### Issue: Add location audit report prototype

Acceptance:

- Groups items by location.
- Includes configured fields.
- Supports export.

## Epic 7 — Domain boundary and safety controls

### Goal

Prevent Metadata Studio from becoming metadata soup or leaking sensitive data.

### Issues

#### Issue: Define OpenSpec for core-vs-custom metadata boundary

Acceptance:

- First-class concepts documented.
- Custom field scope rules documented.
- Sensitive defaults documented.

#### Issue: Add privacy/export visibility flags

Acceptance:

- Field definitions include visibility settings.
- Public/shareable export respects visibility.
- Sensitive field warning shown.

#### Issue: Add duplicate field detection

Acceptance:

- Creating similar field labels/keys warns user.
- Import mapping suggests existing fields by alias/similarity.

#### Issue: Add field deprecation and merge plan

Acceptance:

- Fields can be deprecated.
- Deprecated fields remain exportable.
- Merge conflicts are identified for future UI.

#### Issue: Add unused/deprecated field report

Acceptance:

- Shows fields with no values.
- Shows deprecated fields with values.
- Helps clean up imported metadata.

## Suggested delivery sequence

### Phase 0 — Spec and planning

1. Define OpenSpec for custom fields.
2. Define OpenSpec for templates.
3. Define OpenSpec for import/export preservation.
4. Define OpenSpec for query/report behaviour.

### Phase 1 — Data foundation

5. Field definitions storage/API.
6. Field values storage/API.
7. Field lifecycle and privacy flags.

### Phase 2 — Minimal UI

8. Metadata Studio shell.
9. Field library table.
10. Field editor.
11. Item/specimen custom fields render/save.

### Phase 3 — Templates

12. Template storage/API.
13. Create item from template.
14. Trading card template.
15. Slot car template.
16. Gallery object template.

### Phase 4 — Layouts

17. Form sections.
18. Detail sections.
19. Table columns.
20. Card subtitles/badges.

### Phase 5 — Import/export

21. CSV analyser.
22. Import mapping UI.
23. Unknown column preservation.
24. Homebox preset.
25. Export schema + values.

### Phase 6 — Search/reports

26. Query projection.
27. Custom field search/filter.
28. Saved views.
29. Missing metadata report.
30. Insurance/location reports.

### Phase 7 — Cleanup and polish

31. Duplicate field detection.
32. Deprecated/unused fields report.
33. Template version migration preview.
34. Media bundle import/export.
35. Public/private export safety.

## Labels to use

Suggested labels:

- `metadata`
- `templates`
- `inventory`
- `import-export`
- `reports`
- `ux`
- `openspec`
- `local-first`
- `privacy`

## Validation expectations

Each implementation issue should include:

- OpenSpec validation where relevant.
- API tests for persistence and validation.
- UI tests for form behaviour and reload persistence.
- Import/export tests where relevant.
- Search/report tests where relevant.
- Evidence logs attached to PR/issue.

## First issue to create

Recommended first issue:

```text
Define Metadata Studio custom field model and OpenSpec contract
```

Body outline:

```markdown
## Goal
Define the OpenSpec contract for Cabinet Metadata Studio custom field definitions and typed field values.

## Scope
- Field definition schema
- Field value schema
- Supported initial field types
- Scope rules
- Field lifecycle
- Privacy/export flags
- Core-vs-custom boundary

## Acceptance criteria
- OpenSpec validates strictly.
- Spec supports typed fields, not text-only values.
- Spec defines item/specimen value attachment.
- Spec defines export/import preservation expectations.
- Spec defines sensitive/private defaults.

## Out of scope
- Full UI implementation
- Reports
- Template migration
- Public template sharing
```

## Second issue to create

```text
Add custom field definition and typed value storage
```

## Third issue to create

```text
Add Metadata Studio field library and field editor UI
```

These three issues give Cabinet the foundation needed before templates, import/export, and reports can be built cleanly.
