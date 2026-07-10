# 02 — Category templates

Status: planning draft  
Area: Metadata Studio / collection templates  
Purpose: define how Cabinet should package reusable fields, layouts, defaults, saved views, and reports for different collection categories.

## Product goal

Cabinet should not require users to build every collection schema from scratch.

A new user should be able to choose a sensible starter template, add items immediately, and customise later.

Templates let Cabinet support Homebox/iCollect-style flexibility while still feeling purpose-built for serious collectors, dealers, and galleries.

## Template concept

A Cabinet category template should include:

- template name
- category or object type
- target scope: item/specimen/location/collection/etc.
- included field definitions
- default field values
- form layout sections
- card/tile display layout
- table columns
- saved views
- quick filters
- default reports
- import mapping aliases
- export behaviour
- version number

A template is more than fields. It should describe how metadata is entered, shown, searched, filtered, reported, imported, and exported.

## Template types

| Type | Owner | Purpose |
|---|---|---|
| System template | Cabinet | Built-in starter templates. |
| User template | User/profile | User-created or modified templates. |
| Cloned template | User/profile | Copy of a system template for customisation. |
| Imported template | User/profile | Created during migration/import. |
| Community template | Later | Shared public template, potentially signed/published. |

Initial implementation can support system and user templates only.

## Template lifecycle

| State | Meaning |
|---|---|
| `draft` | Editable, not applied by default. |
| `active` | Available for normal item creation. |
| `deprecated` | Existing items remain, new use discouraged. |
| `archived` | Hidden but preserved. |

Template changes should be versioned.

If a template changes, existing records should not break. Cabinet should show whether an item was created from:

```text
Template: Trading Card v1
Current template: Trading Card v3
```

Then allow optional migration.

## Template definition schema draft

```json
{
  "schema": "cabinet.category-template.v1",
  "template_id": "tpl_trading_card_default",
  "profile_id": "system",
  "name": "Trading Card",
  "description": "Starter template for trading cards, including set, variant, grading, condition and binder metadata.",
  "category_key": "trading_card",
  "status": "active",
  "version": 1,
  "field_ids": [
    "fld_card_set",
    "fld_card_number",
    "fld_card_variant",
    "fld_card_grade"
  ],
  "sections": [
    {
      "key": "identity",
      "label": "Card details",
      "field_keys": ["card.set", "card.number", "card.variant", "card.language"]
    },
    {
      "key": "grading",
      "label": "Grading",
      "field_keys": ["grading.provider", "grading.grade", "grading.certificate_number"]
    }
  ],
  "display": {
    "card_subtitles": ["card.set", "card.number", "card.variant"],
    "badges": ["condition.raw", "grading.provider"],
    "default_table_columns": ["name", "card.set", "card.number", "condition", "location"]
  },
  "saved_views": [
    {
      "name": "Needs grading details",
      "filter": {"missing_any": ["grading.provider", "grading.grade"]}
    }
  ],
  "reports": ["missing_metadata", "collection_completeness", "valuation"],
  "created_at": "2026-07-10T00:00:00+10:00",
  "updated_at": "2026-07-10T00:00:00+10:00"
}
```

## Starter templates

### Trading card template

Candidate fields:

| Field | Type | Scope | Notes |
|---|---|---|---|
| Game/category | enum | item | Pokémon, Magic, sports, other. |
| Set | text/enum | item | Can become catalogue-linked later. |
| Card number | text | item | Text because many card numbers contain letters/symbols. |
| Name | core field | item | Do not duplicate if core item name exists. |
| Variant | enum/multi-select | item/specimen | Holo, reverse holo, promo, etc. |
| Language | enum/country_region | specimen | Language can differ by specimen. |
| Edition | text/enum | specimen | First edition, unlimited, etc. |
| Foil/holo | boolean/enum | specimen | Depending on category. |
| Promo flag | boolean | specimen | Quick filter. |
| Graded/raw state | enum | specimen | Raw, graded, slabbed, other. |
| Grader | enum | specimen | PSA, BGS, CGC, other. |
| Grade | decimal/enum | specimen | Numeric grade with scale later. |
| Certification number | text | specimen | Searchable. |
| Raw condition | enum | specimen | If not graded. |
| Quantity | core/specimen | specimen | Quantity model must be first-class. |
| Binder/deck/trade pile location | location_reference or core location | specimen | Location should remain first-class. |
| Purchase/source notes | acquisition/core + long_text | acquisition | Use acquisition event where possible. |
| Photos | core media | media | Front/back/cert/damage labels. |

Default sections:

- Card details
- Variant and condition
- Grading
- Ownership and location
- Acquisition
- Media and evidence

Default views:

- All cards
- Needs photos
- Needs condition
- Graded cards
- Raw cards
- Duplicates
- Trade pile
- Wishlist gaps
- High value

Default reports:

- Set completion
- Missing grading metadata
- Valuation by set
- Binder audit
- Trade-ready list

### Slot car template

Candidate fields:

| Field | Type | Scope | Notes |
|---|---|---|---|
| Brand | enum/text | item | Scalextric, Carrera, Slot.it, Fly, Ninco, etc. |
| Model | text | item | Searchable. |
| Scale | enum/measurement | item | 1:32, 1:64, etc. |
| Series | text | item | Optional. |
| Year | integer/date precision | item | Allow approximate year. |
| Livery | text | item | Important search/filter field. |
| Catalogue number | text | item | Searchable identifier. |
| Boxed/unboxed | enum/boolean | specimen | Use enum if partial box states matter. |
| Box condition | enum | specimen | Separate from car condition. |
| Car condition | enum | specimen | First-class condition may cover this. |
| Missing parts | multi-select/long_text | specimen | Structured + notes. |
| Restoration notes | long_text | specimen | Could later become restoration events. |
| Chassis type | text/enum | specimen | Technical metadata. |
| Motor type/RPM | text/number | specimen | Tuning metadata. |
| Gear ratio | text/decimal | specimen | Depends on format. |
| Tyres | text | specimen | Current setup. |
| Braid | text | specimen | Current setup. |
| Guide blade | text | specimen | Current setup. |
| Magnet setup | text/enum | specimen | Tuning metadata. |
| Digital chip / analogue | enum | specimen | Compatibility. |
| Track compatibility | multi-select | item/specimen | Carrera, Scalextric, routed, etc. |
| Photos | core media | media | Car/chassis/box/damage/underside. |

Default sections:

- Car details
- Box and condition
- Technical setup
- Maintenance/restoration
- Location and ownership
- Photos/evidence

Default views:

- Boxed cars
- Needs maintenance
- Modified cars
- Digital cars
- Race-ready
- Display-only
- Missing parts
- High value

Default reports:

- Maintenance schedule
- Missing parts
- Box condition audit
- Race-ready list
- Valuation by brand/series

### Gallery / art object template

Candidate fields:

| Field | Type | Scope | Notes |
|---|---|---|---|
| Artist/maker | text/contact_reference | item | Contact reference later. |
| Title | core item name | item | Should use core name/title. |
| Date/period | text/date/date_range | item | Art often needs approximate dates. |
| Medium/material | text/multi-select | item | Paint, bronze, textile, etc. |
| Dimensions | measurement | specimen/item | Multiple dimensions needed. |
| Edition | text/integer | specimen/item | Edition number/size. |
| Signature/markings | long_text/photo_reference | specimen | Include evidence photo. |
| Provenance | long_text/event later | specimen | Critical for galleries. |
| Exhibition history | long_text/date_range later | specimen | Could become structured events later. |
| Publication history | long_text/url | specimen | Source/reference. |
| Condition report | long_text/file_reference | specimen | Could link documents/photos. |
| Conservation notes | long_text | specimen | Later restoration/conservation event. |
| Acquisition method | enum | acquisition | Purchase, gift, consignment, etc. |
| Appraisal value | money | valuation | Reportable, private by default. |
| Insurance value | money | valuation | Private by default. |
| Rights/licence notes | long_text | item/specimen | Public/private warning. |
| Public catalogue description | long_text | item | Allow public visibility. |
| Current display status | enum | specimen | Displayed, stored, loaned, transit. |
| Loan period | date_range | specimen/event | Could be first-class later. |

Default sections:

- Object identity
- Maker/artist
- Materials and dimensions
- Provenance
- Condition and conservation
- Acquisition and value
- Exhibition and loan
- Rights and public catalogue
- Media/evidence

Default views:

- On display
- In storage
- On loan
- Missing condition report
- Missing provenance
- Missing insurance value
- Recently acquired
- Needs photography

Default reports:

- Gallery catalogue
- Insurance schedule
- Loan report
- Condition report register
- Missing provenance
- Acquisition register

### Comic template

Candidate fields:

- publisher
- series
- issue number
- volume
- cover date
- variant cover
- artist/writer
- grade/raw condition
- slab/grading provider
- certification number
- first appearance flags
- signed flag
- signature authentication
- bag/board status
- storage box

### Watch template

Candidate fields:

- brand
- model
- reference number
- serial number
- movement
- case material
- case size
- dial colour
- bracelet/strap
- box/papers status
- service history
- warranty expiry
- purchase dealer
- appraisal value

### Book/rare book template

Candidate fields:

- author
- title
- publisher
- publication year
- edition
- printing
- ISBN
- signed/inscribed
- dust jacket condition
- binding condition
- provenance
- shelf location

### Miniatures template

Candidate fields:

- game/system
- faction
- unit/type
- scale
- painted status
- painter
- paint scheme
- basing style
- conversion/modification notes
- storage tray
- points value

## Template UI

### Template library

Show a list of templates:

```text
Metadata Studio > Templates

[Trading Card]      System   Active     Used by 128 items
[Slot Car]          System   Active     Used by 42 items
[Gallery Object]    System   Active     Used by 0 items
[My Pokemon Cards]  User     Active     Used by 613 items
```

Actions:

- preview template
- create from template
- clone
- edit clone
- deactivate
- export template
- import template

System templates should be read-only. Users can clone them.

### Template preview

Preview should show:

- fields
- sections
- default card display
- default table columns
- saved views
- reports
- import aliases

### Applying templates

Use cases:

- create new item with selected template
- apply template to existing item/specimen
- bulk apply to selected items
- create collection with default template
- import creates suggested template

Applying a template should not destroy existing data. It should add fields/layout defaults and optionally map existing values.

## Template migration

When a template version changes:

- Keep old values.
- Show new fields as missing/default.
- Keep deprecated fields hidden but exportable.
- Offer migration mapping if a field was renamed/split/merged.

Example:

```text
Template update available: Trading Card v2 -> v3

Changes:
+ Added Language
+ Added Certification number
- Deprecated Foil flag; replaced by Variant multi-select

Actions:
[Preview] [Apply to new items only] [Migrate existing items]
```

## Acceptance criteria

Category templates are acceptable when:

- Cabinet ships with starter templates for trading cards, slot cars, and gallery objects.
- Users can create items from a template without manually building fields first.
- Users can clone and modify a system template.
- Template field definitions are versioned.
- Existing item values remain safe when templates change.
- Templates include form layout and display defaults, not just fields.
- Templates can be exported with field definitions.
- Templates can be used by import mapping.

## Suggested backlog slices

1. Define OpenSpec for category templates.
2. Add system template registry.
3. Add template-field association model.
4. Add template sections/layout defaults.
5. Add template CRUD API for user templates.
6. Add clone system template behaviour.
7. Add trading card starter template.
8. Add slot car starter template.
9. Add gallery object starter template.
10. Add template preview UI.
11. Add create item from template flow.
12. Add apply template to existing item/specimen flow.
13. Add template export/import support.
14. Add template versioning/migration preview.
