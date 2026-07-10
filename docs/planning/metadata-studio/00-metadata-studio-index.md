# Cabinet Metadata Studio — planning index

Status: planning draft  
Target repo: `collectors-tech/cabinet`  
Target branch: `develop`  
Purpose: plan Cabinet's configurable metadata, templates, display, reporting, and migration foundation so Cabinet can cover Homebox/iCollect-style metadata gaps without becoming a messy blank database builder.

## Why this exists

Cabinet needs to be better than Homebox as a base collection manager, before the future trust, trading, receipts, feedback vault, Radicle, and P2P layers matter.

The practical finding from reviewing Homebox and the supplied iCollect notes is simple:

- Homebox's advantage is mostly **metadata, locations, attachments, import/export, and reporting-style organisation**.
- iCollect's advantage is mostly **custom field types, collection-specific display control, sorting/filtering, pre-filled defaults, and user-configurable layouts**.
- Cabinet's advantage should be turning those into a serious local-first **Metadata Studio** that supports collectors, dealers, galleries, clubs, stores, and future trust/trade workflows.

Cabinet should not hard-code every competitor field. It should provide a configuration layer that can represent most metadata patterns safely, then keep important Cabinet concepts first-class.

## Document set

| File | Purpose |
|---|---|
| `01-field-types-and-values.md` | Defines the typed custom field system, field value storage, validation, privacy flags, and behaviour per type. |
| `02-category-templates.md` | Defines category templates such as trading cards, slot cars, gallery objects, comics, books, watches, and other collection types. |
| `03-display-and-form-configuration.md` | Defines configurable item forms, card/tile subtitles, table columns, badges, sections, advanced fields, and layout rules. |
| `04-search-sort-filter-and-reporting.md` | Defines how configured metadata becomes searchable, sortable, filterable, saved-view capable, and reportable. |
| `05-import-export-and-migration.md` | Defines CSV/JSON/media import/export, unknown column preservation, Homebox/iCollect migration, and export schema rules. |
| `06-core-domain-boundaries.md` | Defines what must remain first-class Cabinet data instead of being hidden inside arbitrary custom fields. |
| `07-backlog-plan.md` | Converts the plan into epics, issue slices, dependencies, acceptance criteria, and suggested backlog sequencing. |

## Key product decision

Build **Metadata Studio**, not a loose custom-fields widget.

Metadata Studio should let a user/admin configure:

- field definitions
- field types
- category templates
- default values
- form sections
- card/tile display
- table columns
- saved views
- filters
- reports
- import mappings
- export schema
- public/private visibility

This means Cabinet can support fields such as:

- serial number
- model number
- artist
- maker
- livery
- set name
- card number
- edition number
- acquisition method
- appraisal value
- insured value
- warranty expiry
- conservation notes
- condition report
- exhibition history
- loan state
- storage reference
- grading certificate
- box condition
- track compatibility

without turning the whole product into custom field soup.

## Design principles

### 1. First-class where workflow matters

If Cabinet has meaningful behaviour attached to a concept, that concept should be first-class.

Examples:

- item/specimen identity
- collection membership
- location/storage
- media/evidence
- condition
- acquisition
- valuation
- movement history
- wishlist state
- trade state
- deletion/archive lifecycle
- public/private visibility

Custom metadata should fill **category-specific and app-specific gaps**, not replace Cabinet's real domain model.

### 2. Typed fields, not just text blobs

Every custom field should have a type and typed behaviour.

A money field should sort as money. A date field should filter by date. A toggle should render as a switch. An enum should provide controlled options. A measurement should know its unit.

### 3. Templates over blank databases

Casual users should be able to start with a sensible template. Power users should be able to edit or clone templates.

Cabinet should ship with starter templates for obvious collection types, then allow user-created templates later.

### 4. Import must preserve unknown data

When importing from Homebox, iCollect, a spreadsheet, or a gallery system, Cabinet must not drop unknown columns.

Unknown columns should become proposed custom fields during the import mapping flow, with type inference and review before apply.

### 5. Export must include schema and values

Users should never lose custom metadata because Cabinet's export is too shallow.

Exports should include:

- field definitions
- template definitions
- field values
- import refs
- media references
- reportable metadata
- visibility flags where safe

### 6. Search and reports should be metadata-aware

Custom metadata is valuable only if it can be used.

Fields should be configurable as:

- searchable
- sortable
- filterable
- groupable
- visible in table
- visible on cards
- reportable
- exportable

### 7. Local-first and private by default

Metadata must follow Cabinet's local-first and private-by-default posture.

Publishing a public catalogue, public binder, receipt, proof bundle, or future Radicle/Git manifest must not accidentally leak private metadata.

## Suggested product name

Use **Metadata Studio** in product/design/backlog language.

Alternative lower-key labels:

- Metadata settings
- Collection templates
- Field templates
- Collection schema

Recommended UI navigation:

```text
Settings
  Metadata Studio
    Field Library
    Templates
    Form Layouts
    Card & Table Views
    Reports
    Import Mappings
```

For normal users, expose a gentler entry from a collection:

```text
Collection settings
  Fields
  Layout
  Reports
```

## Non-goals

- Do not make Cabinet a cloud-first collection database.
- Do not replace core item/specimen/location/media/trade concepts with arbitrary fields.
- Do not force users to build templates before adding items.
- Do not require custom field knowledge for basic inventory use.
- Do not publish private metadata by default.
- Do not create a marketplace/escrow feature as part of metadata work.

## Source inputs used

This planning set synthesises:

- Max's supplied iCollect Everything metadata/customisation notes.
- Homebox custom field, entity type, and CSV import patterns.
- Cabinet product direction: desktop-first, local-first collector workspace.
- Cabinet planning direction around custom category templates, trading card fields, slot car fields, import/migration, export with media, and no-data-hostage principles.
