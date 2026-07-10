# 06 — Core domain boundaries

Status: planning draft  
Area: Metadata Studio / domain modelling  
Purpose: define what must stay first-class in Cabinet, and what belongs in configurable custom metadata.

## Product goal

Cabinet should use custom metadata to fill gaps, not to replace the product's real model.

If everything becomes a custom field, Cabinet loses the ability to build reliable workflows around inventory, collections, locations, evidence, acquisition, valuation, wishlist, trade, and trust.

## Rule

> Metadata is for category-specific details. Workflows need first-class models.

## First-class Cabinet concepts

These should not be implemented only as custom fields.

| Concept | Why it must be first-class |
|---|---|
| Item | Core object identity and catalogue matching. |
| Specimen | Physical owned copy, condition, ownership, movement, evidence. |
| Collection | First-class grouping/workspace surface. |
| Location/storage | Physical audit, storage hierarchy, movement history. |
| Media/evidence | Photos, receipts, certificates, condition proof, export bundles. |
| Condition | Used for valuation, trade, disputes, reports, and trust. |
| Acquisition | Source, purchase, receipt, provenance, valuation basis. |
| Valuation | Money, source, date, confidence, stale data warnings. |
| Wishlist | Wanted items, priority, planning, conversion to inventory. |
| Trade state | Reservation, proposal, handoff, receipt, feedback. |
| Deleted/archive lifecycle | Safe data retention and filters. |
| Public/private visibility | Privacy, sharing, public binder/catalogue safety. |
| Import batch | Auditable import/migration history. |
| Backup/restore | Local-first trust and recoverability. |

## Custom metadata concepts

These are good custom metadata candidates.

| Category | Examples |
|---|---|
| Category-specific identifiers | Card number, catalogue number, issue number, reference number. |
| Variant details | Foil, livery, colourway, edition, cover variant. |
| Technical specs | Motor type, tyres, movement, scale, material. |
| Gallery details | Medium, dimensions, signature notes, public description. |
| Hobby status | Painted, modified, race-ready, display-only. |
| App migration fields | Old app ID, legacy category, imported tags. |
| Report extensions | Insurance class, appraisal source, audit flag. |
| Display hints | Subtitle fields, badge fields, layout sections. |

## Boundary examples

### Location

Bad design:

```text
Custom field: Storage Location = Cabinet A / Shelf 2
```

Better design:

```text
First-class location tree:
  Studio / Cabinet A / Shelf 2

Optional custom fields on location:
  humidity_controlled = true
  display_case_material = glass
```

Why:

- location needs hierarchy
- movement history needs location references
- reports need item counts by location
- galleries need audits
- future trade/packing workflows need physical location

### Condition

Bad design:

```text
Custom field: Condition = Good
```

Better design:

```text
First-class specimen condition:
  current_condition = good
  condition_updated_at = date
  condition_notes = notes

Optional custom fields:
  box_condition = poor
  dust_jacket_condition = fair
  chassis_condition = excellent
```

Why:

- condition affects valuation
- condition affects trade trust
- condition affects disputes
- condition needs history/evidence

### Acquisition

Bad design:

```text
Custom fields:
  Purchase Date
  Purchase Price
  Seller
  Receipt
```

Better design:

```text
First-class acquisition event:
  acquired_at
  source
  method
  price
  currency
  receipt_media_id
  notes

Optional custom fields:
  auction_lot_number
  dealer_invoice_reference
  import_permit_number
```

Why:

- acquisition feeds provenance
- reports need purchase totals
- receipts/media need links
- selling/trading history needs event timeline

### Valuation

Bad design:

```text
Custom field: Value = 300
```

Better design:

```text
First-class valuation record:
  amount
  currency
  source
  checked_at
  confidence
  stale_after
  manual_override

Optional custom fields:
  appraisal_method
  insurance_class
```

Why:

- value changes over time
- source matters
- listed vs sold matters
- reports need current/stale logic

### Media/evidence

Bad design:

```text
Custom field: Certificate File = /Users/max/Desktop/cert.pdf
```

Better design:

```text
First-class media asset:
  media_id
  hash
  type
  path/provider
  thumbnail
  label = certificate
  linked_entity = specimen

Custom field value:
  certificate_reference = media_id
```

Why:

- media needs storage rules
- export needs media manifest
- evidence may be private
- path repair/rebuild matters

## Metadata versus events

Some data starts as metadata but may become an event workflow later.

| Initially configurable | Later first-class candidate |
|---|---|
| Maintenance notes | Maintenance event log |
| Restoration notes | Restoration/conservation event |
| Exhibition history text | Exhibition event records |
| Loan status/date range | Loan event workflow |
| Appraisal value | Valuation event history |
| Source listing URL | Acquisition/source record |

Design should allow migration from custom fields to first-class events.

Do this by:

- preserving field IDs
- storing import source
- supporting field deprecation
- supporting data migration maps
- not hard-deleting field values

## Privacy boundaries

Field definitions and core models must include privacy metadata.

Private by default:

- purchase price
- storage location
- private notes
- full collection value
- contact details
- home address
- unreleased dispute notes
- internal appraisal notes

Potentially public/shareable:

- public title
- public description
- artist/maker
- public photos selected by user
- public condition summary
- public trade binder fields
- public wishlist fields
- signed receipts where user explicitly publishes

## Avoiding metadata soup

Cabinet should prevent custom fields from becoming a junk drawer.

Design controls:

- starter templates
- field library
- field descriptions/help text
- duplicate field detection
- merge fields action
- deprecate fields action
- missing metadata report
- unused field report
- imported field review
- template governance
- field naming conventions

## Duplicate field detection

Warn when creating fields like:

```text
Serial Number
Serial No.
serial_number
SN
```

Suggest merge or alias.

Field aliases should help imports without creating multiple duplicate fields.

## Field merge flow

```text
Merge fields
  Source field: Serial No.
  Target field: Serial Number

Affected values: 246
Conflicts: 12

Conflict action:
  - keep target
  - overwrite target
  - keep both in notes
  - review one by one
```

## Public proof / future trust boundary

Cabinet's future public ledgers, receipts, feedback vault, and identity/reputation layers should not automatically include arbitrary metadata.

Custom metadata must declare whether it is safe for:

- local only
- private export
- public catalogue
- public trade binder
- public wishlist
- signed receipt
- signed feedback context
- community catalogue contribution

Default should be private/local only.

## Acceptance criteria

Domain boundaries are acceptable when:

- Core Cabinet workflow concepts are not reduced to arbitrary fields.
- Custom fields can extend item/specimen/location/collection records.
- Sensitive metadata defaults to private.
- Field definitions support future migration to first-class models.
- Import preserves metadata without corrupting domain boundaries.
- Users can detect, merge, deprecate, and clean up duplicate fields.
- Public exports do not leak private metadata by default.

## Suggested backlog slices

1. Define OpenSpec for core-vs-custom metadata boundary.
2. Add field scope rules.
3. Add privacy/export visibility flags to field definitions.
4. Add duplicate field detection by key/label/alias.
5. Add deprecate/merge field planning spec.
6. Add sensitive field warning rules.
7. Add public export allowlist model.
8. Add missing/unused/duplicate metadata reports.
9. Add migration path from custom field to first-class model later.
