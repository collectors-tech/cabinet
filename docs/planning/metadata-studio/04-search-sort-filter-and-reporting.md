# 04 — Search, sort, filter, saved views, and reporting

Status: planning draft  
Area: Metadata Studio / query and reporting  
Purpose: define how custom metadata becomes useful in Cabinet's search, saved views, dashboards, reports, and export flows.

## Product goal

Custom metadata is only valuable if users can use it to find, sort, filter, group, report, and act.

Cabinet should make configured metadata operational.

A gallery should be able to ask:

- Which objects are missing condition reports?
- Which items are on loan?
- Which objects have no insurance value?
- Which pieces were acquired this year?

A collector should be able to ask:

- Which cards are missing back photos?
- Which slot cars need maintenance?
- Which items are boxed and high value?
- Which items are trade-ready?

## Field behaviour flags

Each field definition should declare whether it is:

| Flag | Meaning |
|---|---|
| `searchable` | Included in keyword/full-text search. |
| `sortable` | Can be used as sort key. |
| `filterable` | Can be used in filters. |
| `groupable` | Can group report/table results. |
| `reportable` | Available in report builder. |
| `visible_in_table` | Available as table column. |
| `visible_on_card` | Available in card layout. |
| `exportable` | Included in exports unless disabled. |

Default behaviour should depend on field type.

## Search design

### Search scope

Cabinet search should cover:

- core item fields
- specimen fields
- configured searchable custom fields
- tags/labels
- barcodes/identifiers
- catalogue references
- media labels
- acquisition references
- notes/long text where enabled

### Field-aware search

Support simple search first:

```text
"charizard psa 8 binder a"
```

Then add advanced field queries later:

```text
set:"Base Set" grade:8 location:"Binder A"
boxed:true needs_maintenance:true
artist:"Grace Cossington Smith" display_status:on_loan
```

### Search indexing

For SQLite/local-first:

- maintain a search projection table
- include text fields and selected custom field values
- use FTS for searchable text/long text
- use typed indexes/projections for sortable/filterable fields
- rebuild index via maintenance action

Suggested projection:

```text
item_search_projection
  profile_id
  entity_id
  entity_scope
  title_text
  keyword_text
  custom_text_blob
  updated_at
```

Typed filter projection:

```text
custom_field_filter_projection
  profile_id
  field_id
  entity_id
  entity_scope
  text_value
  number_value
  money_amount
  money_currency
  date_value
  boolean_value
  option_key
```

Do not over-optimise initially, but avoid a design that requires scanning every JSON blob for every table filter.

## Sorting

Sort behaviour should be type-aware.

| Type | Sort behaviour |
|---|---|
| text | locale-aware alphabetical. |
| long_text | not sortable by default. |
| integer/decimal/quantity | numeric. |
| money | amount within currency; converted amount only where rate exists. |
| date | chronological. |
| boolean | false/true/missing order configurable. |
| enum | option sort order, not alphabetical by label unless configured. |
| multi_select | not default sortable; may sort by first selected or count. |
| star_rating | numeric. |
| measurement | by normalised unit if compatible. |
| reference | by referenced display name. |

## Filtering

Filter operations should be determined by field type.

| Type | Filter examples |
|---|---|
| text | contains, equals, starts with, is empty. |
| long_text | contains, is empty. |
| number | equals, greater than, less than, between, is empty. |
| money | between, currency, greater than, missing. |
| date | before, after, between, next 30 days, overdue, missing. |
| boolean | true, false, missing. |
| enum | is, is not, in list, missing. |
| multi_select | includes any, includes all, excludes, missing. |
| country_region | is, region group later. |
| star_rating | at least, exactly, below. |
| measurement | greater than, less than, unit-compatible. |
| reference | is, is not, missing, target archived. |

## Saved views

A saved view combines:

- filters
- sort
- grouping
- columns
- card/table mode
- report mode optional
- template/category scope

Example:

```json
{
  "schema": "cabinet.saved-view.v1",
  "view_id": "view_missing_gallery_metadata",
  "name": "Missing gallery metadata",
  "template_id": "tpl_gallery_object_default",
  "filters": [
    {"field": "gallery.provenance", "op": "missing"},
    {"field": "media.condition_report", "op": "missing"}
  ],
  "columns": [
    "core.title",
    "gallery.artist",
    "location.current",
    "gallery.provenance",
    "media.condition_report"
  ],
  "sort": [{"field": "core.title", "direction": "asc"}]
}
```

## Report builder

Reports should be metadata-driven, not hard-coded one-offs.

### Report definition

```json
{
  "schema": "cabinet.report-definition.v1",
  "report_id": "report_insurance_schedule",
  "name": "Insurance schedule",
  "template_scope": "gallery_object",
  "description": "Items with insurance values grouped by location.",
  "filters": [
    {"field": "valuation.insurance_value", "op": "exists"}
  ],
  "group_by": ["location.current"],
  "columns": [
    "core.title",
    "gallery.artist",
    "valuation.insurance_value",
    "valuation.appraisal_date",
    "location.current"
  ],
  "summary": [
    {"field": "valuation.insurance_value", "op": "sum", "grouped": true}
  ],
  "export_formats": ["csv", "json", "pdf-later"]
}
```

Initial implementation should focus on table-based reports and CSV/JSON export. PDF can come later.

## Report categories

### Metadata quality reports

Purpose: help users clean up inventory.

Examples:

- Missing required fields
- Missing photos
- Missing condition
- Missing location
- Missing provenance
- Missing appraisal value
- Missing certificate number
- Invalid field values
- Deprecated fields still in use

### Insurance and value reports

Purpose: support serious collection management.

Examples:

- Insurance schedule
- High-value items
- Value by collection
- Value by location
- Value missing appraisal date
- Stale valuation report

### Acquisition reports

Purpose: track purchase/source history.

Examples:

- Acquired this year
- Acquisition by source/dealer
- Purchase price by month
- Missing receipt
- Items without acquisition source

### Maintenance/restoration/conservation reports

Purpose: support active hobby and gallery workflows.

Examples:

- Slot cars needing maintenance
- Items serviced in last 12 months
- Restoration projects in progress
- Conservation required
- Missing condition reports

### Location audit reports

Purpose: verify physical storage/display.

Examples:

- Items by location
- Empty locations
- Items with unknown location
- Location inventory checklist
- Gallery wall audit
- Binder audit

### Collection completeness reports

Purpose: cards, comics, sets, series.

Examples:

- Missing items in set
- Duplicate items
- Wishlist gaps
- Completion by set
- Trade-ready duplicates

### Gallery catalogue reports

Purpose: present collection data.

Examples:

- Public catalogue export
- Exhibition checklist
- Loan object list
- Wall labels/source metadata
- Artist inventory list

## Dashboards and widgets

Once reports exist, dashboard widgets can be configured from saved views/reports.

Examples:

```text
Dashboard widgets
  - Items missing photos: 23
  - Items needing valuation refresh: 8
  - Slot cars needing maintenance: 5
  - Wishlist matches from Marketwatch: 12
  - Gallery objects on loan: 3
```

## Privacy and reporting

Reports can include sensitive metadata.

Report definitions should specify visibility:

| Visibility | Use |
|---|---|
| `private` | Local/internal only. |
| `shareable` | Can export manually. |
| `public` | Safe for public catalogue/binder. |

Money fields, storage locations, private notes, full collection value, and contact details should default to private.

## Performance expectations

Cabinet is desktop-first and local-first. Reports and filters should feel fast for common collection sizes.

Targets for planning:

- thousands of items: instant/near-instant filters
- tens of thousands: acceptable with indexes/projections
- media-heavy reports: avoid loading originals unless required
- import report generation: can take longer but must show progress

## UI design

### Saved view bar

```text
Inventory
[All] [Cards] [Trade pile] [Needs photos] [High value] [Missing metadata] [+ Save view]
```

### Filter builder

```text
Filter
  Field: Box condition
  Operator: is
  Value: Poor

+ Add filter
[Apply] [Save as view]
```

### Report library

```text
Reports
  Metadata quality
    Missing required metadata
    Missing photos
    Deprecated fields in use

  Insurance and value
    Insurance schedule
    High-value items
    Stale valuations

  Gallery
    Public catalogue
    Loan report
    Exhibition checklist
```

### Report result actions

- export CSV
- export JSON
- save as view
- bulk edit selected items
- open item side panel
- print/PDF later

## Acceptance criteria

Search/reporting is acceptable when:

- Custom fields can be marked searchable/sortable/filterable/reportable.
- Search includes configured searchable metadata.
- Table filters support typed custom fields.
- Saved views can include custom field filters and columns.
- Reports can be defined from configured fields.
- Missing metadata reports work.
- Report results can export CSV/JSON.
- Sensitive fields are private by default.
- Reindex/repair can rebuild metadata projections.

## Suggested backlog slices

1. Define OpenSpec for custom field search/filter behaviour.
2. Add custom field query projection model.
3. Add metadata reindex command.
4. Add custom fields to inventory search index.
5. Add typed custom field filter API.
6. Add table custom field columns.
7. Add saved views with custom field filters.
8. Add report definition model.
9. Add missing metadata report.
10. Add insurance/value report prototype.
11. Add location audit report prototype.
12. Add report CSV/JSON export.
13. Add dashboard widgets from saved views later.
