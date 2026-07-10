# 01 — Field types and values

Status: planning draft  
Area: Metadata Studio / custom fields  
Purpose: define Cabinet's typed field system so custom metadata is searchable, sortable, filterable, displayable, reportable, import-safe, and export-safe.

## Product goal

Cabinet should support arbitrary collection metadata without making everything a plain text note.

Homebox-style custom fields are useful, but Cabinet needs a richer field system because it must support:

- home collectors
- high-volume collectors
- dealers
- galleries
- stores/clubs
- cards
- slot cars
- comics
- watches
- art objects
- antiques
- future category templates

## Field system principles

### Typed fields

Every configured field must have a type.

The type controls:

- input component
- validation
- stored value shape
- display formatting
- search behaviour
- sort behaviour
- filter behaviour
- report behaviour
- import parsing
- export representation

### Stable keys

Every field must have a stable internal key, separate from its display label.

Example:

```json
{
  "field_key": "grading.certificate_number",
  "label": "Certification number"
}
```

The user can rename the label without breaking saved views, imports, exports, or reports.

### Scoped fields

A field should declare what it applies to.

Suggested scopes:

| Scope | Meaning |
|---|---|
| `item` | Shared/canonical item-level metadata. |
| `specimen` | The user's physical owned copy. |
| `wishlist_item` | Wanted item planning metadata. |
| `collection` | Collection-level metadata. |
| `location` | Storage or display location metadata. |
| `acquisition` | Purchase/source metadata extension. |
| `valuation` | Valuation/appraisal metadata extension. |
| `media` | Photo/document/evidence metadata. |
| `contact` | Dealer/artist/source/custodian metadata extension, if contacts are later added. |

Initial implementation can start with `item` and `specimen`, but the schema should not block later scopes.

### Values should be profile-local

Cabinet is local-first and profile-scoped. Custom field definitions and values should belong to the active Cabinet profile/workspace, unless explicitly imported as a system/template field.

## Required field types

Cabinet should support at least these field types.

| Type | Purpose | Example |
|---|---|---|
| `text` | Short string | Serial number, artist name, card number |
| `long_text` | Notes / markdown / reports | Provenance, condition report |
| `integer` | Whole number | Issue number, edition number |
| `decimal` | Numeric precision | Grade, scale ratio |
| `quantity` | Count semantics | Duplicate count, parts count |
| `money` | Currency-aware value | Appraisal value, insured value |
| `date` | Single date | Date serviced, date signed |
| `date_range` | From/to dates | Exhibition period, loan period |
| `boolean` | Toggle/switch | Has box, autographed, graded |
| `enum` | One controlled option | Condition, language, storage type |
| `multi_select` | Multiple controlled options | Materials, genre, features |
| `country_region` | Region/origin tracking | Country of origin, market region |
| `star_rating` | 1–5 personal rating | Favourite rating, quality rating |
| `url` | External reference | Source listing, catalogue page |
| `file_reference` | Linked document | Certificate, invoice, appraisal PDF |
| `photo_reference` | Linked image/evidence | Damage photo, grading photo |
| `measurement` | Value + unit | Dimensions, weight, scale |
| `item_reference` | Link to another Cabinet object | Related item, part of set |
| `location_reference` | Link to location | Display case, shelf, box |
| `contact_reference` | Future contact/person/org link | Artist, dealer, appraiser |

## Type behaviour

### Text

Used for short identifiers and labels.

Design rules:

- max length configurable
- searchable by default
- sortable alphabetically
- can be used on card subtitles
- can be used in table columns
- can have import aliases

Examples:

- serial number
- catalogue number
- livery
- artist
- maker
- set
- card number

### Long text

Used for prose, markdown notes, provenance, condition reports, conservation notes, and public catalogue descriptions.

Design rules:

- searchable by default, but not sortable
- hidden from compact table views by default
- display in detail page sections
- export as full text
- optional markdown support

Examples:

- provenance narrative
- condition report
- restoration notes
- exhibition notes
- catalogue description

### Integer / decimal / quantity

Used for numbers that need numeric sort/filter.

Design rules:

- numeric comparisons: equals, greater than, less than, between
- optional min/max validation
- optional step value
- optional display suffix/prefix
- quantity should default to `0` or `1` depending on context

Examples:

- quantity owned
- edition number
- issue number
- card number where numeric
- motor RPM
- gear ratio as decimal

### Money

Used for values that need currency.

Design rules:

- store amount and currency separately
- display with configured profile currency by default
- support original currency
- sort by normalised value only when conversion data exists
- report totals by currency or converted value
- never publish private money fields by default

Examples:

- purchase price
- insured value
- appraisal value
- estimated current value
- replacement value

### Date / date range

Used for event dates and time windows.

Design rules:

- ISO date storage
- optional precision: year, month, day
- relative filters: overdue, next 30 days, this year
- date ranges support overlap queries

Examples:

- date serviced
- date signed
- acquisition date
- warranty expiry
- exhibition period
- loan period

### Boolean

Used for fast true/false values.

Design rules:

- render as switch/toggle in forms
- render as badge/check in tables
- filter true/false/empty
- default configurable

Examples:

- has box
- autographed
- graded
- modified
- needs maintenance
- insured

### Enum

Used when exactly one option should be selected.

Design rules:

- option list belongs to field definition
- options can have labels, colours, descriptions, sort order
- allow safe rename without changing stored option key
- optional custom options if enabled
- filter by one or more options

Examples:

- condition
- language
- raw/graded state
- box condition
- storage type
- loan status

### Multi-select

Used when many options may apply.

Design rules:

- option list belongs to field definition
- values stored as stable option keys
- filter: includes any, includes all, excludes
- display as compact chips/badges

Examples:

- materials
- genre
- features
- compatible track systems
- damage types

### Country / region

Used for origin or market tracking.

Design rules:

- ISO country code internally where possible
- label localised in UI
- optional region/state/subregion later

Examples:

- country of origin
- market region
- language region
- manufacture country

### Star rating

Used for personal subjective scoring.

Design rules:

- 1–5 stars
- optional half-stars later
- not the same as Cabinet trust/reputation rating
- private by default

Examples:

- personal favourite rating
- display quality rating
- race performance rating

### URL

Used for external references.

Design rules:

- validate URL format
- display domain preview
- clickable from detail view
- optional label/title

Examples:

- source listing
- auction result
- maker page
- catalogue reference

### File/photo reference

Used for linking custom fields to Cabinet-managed media/evidence assets.

Design rules:

- value references an existing Cabinet media asset ID
- field does not store the raw file path directly
- supports public/private visibility per media asset
- export includes media manifest reference

Examples:

- certificate file
- appraisal PDF
- invoice scan
- damage photo
- signature photo

### Measurement

Used for dimensions, weight, size, scale, and other units.

Design rules:

- store value and unit
- support single measurement and dimension set
- filter numerically when compatible units exist
- display with unit

Examples:

- height
- width
- depth
- weight
- scale
- wheelbase

### References

Used to connect metadata to other Cabinet entities.

Design rules:

- store referenced Cabinet ID, not label
- if target is deleted/archived, show broken/archived reference state
- support reference search picker

Examples:

- related item
- part of set
- replacement part
- display location
- appraiser contact

## Field definition schema draft

```json
{
  "schema": "cabinet.custom-field-definition.v1",
  "field_id": "fld_01H...",
  "profile_id": "profile_default",
  "scope": "specimen",
  "field_key": "grading.certificate_number",
  "label": "Certification number",
  "description": "Certificate number from the grading provider.",
  "type": "text",
  "required": false,
  "default_value": null,
  "validation": {
    "max_length": 80,
    "pattern": null
  },
  "display": {
    "show_in_detail": true,
    "show_in_table_default": false,
    "show_on_card": false,
    "section": "Grading",
    "sort_order": 40
  },
  "behaviour": {
    "searchable": true,
    "sortable": true,
    "filterable": true,
    "reportable": true,
    "exportable": true
  },
  "privacy": {
    "default_visibility": "private",
    "allow_public": true,
    "warn_on_public": false
  },
  "import": {
    "aliases": ["Cert No", "Certificate Number", "grading_cert"]
  },
  "status": "active",
  "version": 1,
  "created_at": "2026-07-10T00:00:00+10:00",
  "updated_at": "2026-07-10T00:00:00+10:00"
}
```

## Field value schema draft

Use a typed value envelope so values remain queryable and exportable.

```json
{
  "schema": "cabinet.custom-field-value.v1",
  "value_id": "cfv_01H...",
  "field_id": "fld_01H...",
  "entity_scope": "specimen",
  "entity_id": "specimen_123",
  "typed_value": {
    "type": "text",
    "text": "PSA12345678"
  },
  "source": {
    "kind": "manual",
    "import_batch_id": null
  },
  "created_at": "2026-07-10T00:00:00+10:00",
  "updated_at": "2026-07-10T00:00:00+10:00"
}
```

Money example:

```json
{
  "typed_value": {
    "type": "money",
    "amount": "1200.00",
    "currency": "AUD"
  }
}
```

Enum example:

```json
{
  "typed_value": {
    "type": "enum",
    "option_key": "near_mint"
  }
}
```

Measurement example:

```json
{
  "typed_value": {
    "type": "measurement",
    "value": "30.5",
    "unit": "cm"
  }
}
```

## Empty values

Cabinet should distinguish:

| State | Meaning |
|---|---|
| Missing value | Field does not have a value for this item. |
| Explicit empty | User intentionally cleared the value. |
| Unknown | Imported or unresolved value is unknown. |
| Not applicable | Field does not apply to this item/specimen. |

This matters for missing-metadata reports.

## Field lifecycle

| Status | Meaning |
|---|---|
| `draft` | Field exists but is not active for normal forms. |
| `active` | Field is usable. |
| `deprecated` | Field remains readable but should not be used for new data. |
| `archived` | Hidden from normal use but preserved for export/history. |
| `deleted` | Soft-deleted; values remain until explicit purge/export policy. |

Deleting a field must not silently delete values. Cabinet should require a migration action:

- keep values hidden
- move values to another field
- export before removal
- permanently delete after confirmation

## UI requirements

Field editor should support:

- label
- key/slug
- type
- scope
- section
- help text
- required flag
- default value
- options for enum/multi-select
- import aliases
- searchable/sortable/filterable/reportable/exportable toggles
- privacy/default visibility
- table/card/detail display toggles

Field values in item forms should:

- render the correct input for type
- validate before save
- save through API-backed persistence
- reload correctly after app restart
- show source/import information where relevant

## Acceptance criteria

A field types implementation is acceptable when:

- Users can create typed field definitions.
- Users can attach values to items/specimens.
- Values persist and reload correctly.
- Fields can be soft-deleted without losing values silently.
- Field values can be exported with definitions.
- Unknown imported columns can be represented as field definitions and values.
- Search/filter/report code can determine behaviour from the field type.
- Public/private visibility exists at field level.

## Suggested backlog slices

1. Define OpenSpec for custom field definitions and typed values.
2. Add database tables for field definitions and typed values.
3. Add API endpoints for field definition CRUD.
4. Add API endpoints for item/specimen custom field values.
5. Add validation for each initial field type.
6. Add basic Metadata Studio field editor UI.
7. Add item detail custom field rendering.
8. Add import alias support.
9. Add export of custom field definitions and values.
10. Add soft-delete/deprecate field lifecycle.
