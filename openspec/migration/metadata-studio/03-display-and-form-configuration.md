# 03 — Display and form configuration

Status: planning draft  
Area: Metadata Studio / UX  
Purpose: define how configured metadata controls Cabinet item forms, detail pages, cards, tables, badges, and saved layouts.

## Product goal

Cabinet should let users control how metadata appears without requiring code changes.

This is where Cabinet should match and exceed iCollect-style collection display customisation:

- choose what fields show on cards
- choose table columns
- hide irrelevant default fields
- group fields into meaningful sections
- make advanced metadata available without cluttering basic item entry
- set default values for fast capture
- show badges/labels for important states

## Design principle

The same configured field should be reusable across multiple surfaces:

| Surface | Field behaviour |
|---|---|
| Create form | Input control, required/default state, help text. |
| Edit form | Input control and validation. |
| Detail view | Readable formatted value in sections. |
| Card/tile | Subtitle line or badge. |
| Table | Column value with sort/filter. |
| Saved view | Filter/sort/group rule. |
| Report | Column, metric, or grouping. |
| Import mapping | Target field. |
| Export | Data column/schema entry. |

## Layout configuration objects

Cabinet should support these configurable layouts:

| Layout | Purpose |
|---|---|
| Form layout | Controls create/edit field sections and order. |
| Detail layout | Controls item detail page sections. |
| Card layout | Controls card subtitles, labels, badges, and thumbnail behaviour. |
| Table layout | Controls default columns, widths, order, and pinned fields. |
| Quick capture layout | Controls fast entry/mobile capture fields. |
| Report layout | Controls fields included in specific reports. |

## Form layout

A form layout should define sections.

Example:

```json
{
  "schema": "cabinet.form-layout.v1",
  "layout_id": "layout_trading_card_form_v1",
  "template_id": "tpl_trading_card_default",
  "sections": [
    {
      "key": "basic",
      "label": "Card details",
      "mode": "basic",
      "fields": ["core.name", "card.set", "card.number", "card.variant"]
    },
    {
      "key": "grading",
      "label": "Grading",
      "mode": "advanced",
      "fields": ["grading.provider", "grading.grade", "grading.certificate_number"]
    },
    {
      "key": "evidence",
      "label": "Photos and evidence",
      "mode": "basic",
      "fields": ["media.front_photo", "media.back_photo", "media.damage_photo"]
    }
  ]
}
```

### Basic and advanced fields

Cabinet needs a clean way to support rich metadata without overwhelming casual users.

Use modes:

| Mode | Behaviour |
|---|---|
| `basic` | Shows in normal create/edit. |
| `advanced` | Hidden behind Advanced toggle. |
| `audit` | Used mostly for reports/review. |
| `system` | Shown only in technical/admin surfaces. |

Example:

A trading card create form should show:

- name
- set
- number
- variant
- condition
- location
- photos

Advanced can show:

- grading provider
- certification number
- purchase details
- valuation source
- public/private notes

## Detail layout

Detail pages should be readable, not a long unstructured field dump.

Example sections for gallery object:

```text
Object identity
  Artist: Grace Cossington Smith
  Title: Still Life
  Date: c. 1934
  Medium: Oil on canvas
  Dimensions: 45 x 60 cm

Provenance
  Acquired from: ...
  Previous owner: ...
  Notes: ...

Condition and conservation
  Condition report: Available
  Last conservation: 2025-11-02

Value and insurance
  Insurance value: AUD 12,000
  Appraisal date: 2026-04-12
```

## Card/tile layout

Cards are quick visual summaries.

A configurable card layout should support:

- primary image
- title
- up to 3 subtitle rows
- up to 2 primary labels
- small badges/chips
- optional value/location line
- missing metadata warning icon

### Example: trading card card layout

```text
[thumbnail]
Charizard
Base Set • #4/102
Holo • PSA 8
Binder A / Page 3
[Graded] [High value]
```

### Example: slot car card layout

```text
[thumbnail]
Carrera Ferrari 512S
1:32 • Red #23 livery
Boxed • Digital
Shelf 2 / Case B
[Needs tyres] [Modified]
```

### Example: gallery object card layout

```text
[thumbnail]
Still Life
Grace Cossington Smith • c. 1934
Oil on canvas • 45 x 60 cm
Gallery Wall 2
[On display] [Insured]
```

## Badge and label rules

Badges should come from configured fields or core state.

Badge types:

| Badge type | Source |
|---|---|
| Status badge | Core item/specimen/wishlist/trade state. |
| Metadata badge | Enum/boolean custom field. |
| Warning badge | Missing required metadata, stale value, invalid field. |
| Value badge | Configured money/valuation field, private by default. |
| Evidence badge | Media/certificate/condition report present/missing. |

Badge configuration:

```json
{
  "field_key": "slotcar.needs_maintenance",
  "show_when": true,
  "label": "Needs maintenance",
  "severity": "warning"
}
```

## Table layout

Tables need configurable columns and filters.

Table layout should include:

- visible columns
- column order
- column widths
- pinned columns
- default sort
- default grouping
- quick filters
- page size
- row density

Example:

```json
{
  "schema": "cabinet.table-layout.v1",
  "layout_id": "table_gallery_insurance_v1",
  "name": "Insurance schedule",
  "columns": [
    "core.title",
    "gallery.artist",
    "location.current",
    "valuation.insurance_value",
    "valuation.appraisal_date",
    "media.condition_report"
  ],
  "sort": [{"field": "valuation.insurance_value", "direction": "desc"}],
  "filters": [{"field": "valuation.insurance_value", "op": "exists"}]
}
```

## Hidden default fields

Homebox/iCollect-style customisation includes hiding unneeded default fields.

Cabinet should allow templates to hide core fields from specific forms/views when they are irrelevant, but not delete the underlying model.

Example:

- Hide warranty section for trading cards by default.
- Hide purchase details in basic create form.
- Hide trade state fields for gallery template if trading is disabled.
- Hide motor/tuning fields for unmodified display-only slot cars unless advanced is enabled.

## Default values

Templates should support defaults.

Examples:

| Template | Default |
|---|---|
| Trading cards | Currency = AUD, condition = Unknown, visibility = Private |
| Slot cars | Scale = 1:32, boxed = Unknown, needs maintenance = false |
| Gallery objects | Visibility = Private, display status = Stored, appraisal required = true |

Defaults should be visible during creation and overridable.

## Conditional fields

Cabinet should eventually support conditional display rules.

Examples:

- Show grading provider only if `graded = true`.
- Show box condition only if `boxed_state != unboxed`.
- Show loan period only if `display_status = on_loan`.
- Show conservation notes only if `has_conservation_history = true`.

Initial implementation can defer this, but the model should allow later rules.

Rule example:

```json
{
  "field_key": "grading.provider",
  "visible_when": {
    "field": "grading.state",
    "op": "equals",
    "value": "graded"
  }
}
```

## UX flows

### Create item from template

```text
Inventory > New item
  1. Choose category/template
  2. Basic form opens with template fields
  3. User fills required fields
  4. Optional Advanced section
  5. Save
```

### Edit layout

```text
Collection settings > Layout
  1. Choose template/view
  2. Drag fields into sections
  3. Choose card subtitle lines
  4. Choose badges
  5. Choose table columns
  6. Preview with sample item
  7. Save layout
```

### Configure card display

```text
Card layout
  Title: core.name
  Subtitle 1: gallery.artist + gallery.date
  Subtitle 2: gallery.medium + gallery.dimensions
  Subtitle 3: location.current
  Badges: display_status, insured
```

## Accessibility and usability rules

- Do not rely only on colour for badges.
- Field labels must stay visible or accessible.
- Required fields must be clear.
- Validation messages must explain the problem.
- Advanced sections should be expandable/collapsible.
- Drag/drop layout should have non-drag alternatives.
- Table configuration should support reset to template defaults.

## Acceptance criteria

Display/form configuration is acceptable when:

- Templates can define form sections.
- Templates can define default table columns.
- Templates can define card subtitles and badges.
- Users can hide/show fields in layout config.
- Users can set default values.
- Basic and advanced field modes work.
- Layout changes persist and reload correctly.
- Layouts do not delete field values.
- Item create/edit forms use configured metadata.
- Table/card/detail views display configured fields correctly.

## Suggested backlog slices

1. Define OpenSpec for form/display layout configuration.
2. Add form section model for templates.
3. Add display settings to field definitions.
4. Add default table column config to templates.
5. Add card subtitle/badge config model.
6. Add Metadata Studio layout editor UI.
7. Render item create/edit form from template layout.
8. Render item detail sections from template layout.
9. Render card subtitles and badges from config.
10. Add table column configuration and persistence.
11. Add basic/advanced field mode toggle.
12. Add default values in item creation.
13. Add conditional field display rules later.
