# 05 — Import, export, and migration

Status: planning draft  
Area: Metadata Studio / import-export  
Purpose: define how Cabinet imports metadata from other apps, preserves unknown fields, maps data into templates, and exports complete custom schema + values.

## Product goal

Cabinet should make migration safe.

Users coming from Homebox, iCollect, spreadsheets, Airtable, gallery systems, or custom CSVs should not lose data because Cabinet does not recognise a column.

The rule:

> Unknown columns are not rubbish. They are candidate metadata.

## Core requirements

Cabinet import/export should support:

- CSV import
- JSON import
- dry-run preview
- field mapping
- type inference
- unknown column preservation
- custom field creation during import
- template suggestion
- duplicate detection
- import refs
- media folder matching
- error report
- reversible/import batch audit
- export of custom field definitions
- export of custom field values
- export with media manifest

## Homebox migration lesson

Homebox supports a useful CSV convention for custom fields:

```text
HB.field.Serial Number
HB.field.Condition Notes
HB.field.Appraisal Value
```

Cabinet should support Homebox-style imports, but it should not require Homebox-specific prefixes.

Cabinet should accept:

```text
HB.field.Serial Number
custom.Serial Number
Serial Number
Certificate Number
PSA Cert No
```

Then map them through the import UI.

## Import flow

### Step 1 — Select source

```text
Import data
  Source type:
    - Cabinet export
    - Homebox CSV
    - iCollect CSV/export
    - Generic CSV/TSV
    - Folder of images + CSV
    - JSON bundle
```

### Step 2 — Upload/choose files

Inputs:

- CSV/TSV file
- optional media folder/ZIP
- optional JSON export
- optional template/schema file

### Step 3 — Analyse

Cabinet detects:

- columns
- likely source app
- row count
- likely item name/title column
- likely category/template
- likely custom fields
- likely field types
- import refs
- duplicate risk
- media references
- invalid values

### Step 4 — Mapping screen

Each source column gets mapped to:

- core field
- existing custom field
- new custom field
- ignore
- notes field
- media reference
- relationship/reference

Example:

```text
Source column              Cabinet target                         Action
Name                       core.name                              map
HB.location                core.location                          map
HB.field.Serial Number     custom.serial_number                   create text field
Appraisal Value            custom.valuation.appraisal_value        create money field
Photo File                 media.original                          match media file
Random Old Column          custom.import.random_old_column         preserve as text
```

### Step 5 — Type inference

Cabinet should infer types cautiously.

| Input pattern | Suggested type |
|---|---|
| `true/false`, `yes/no`, `1/0` | boolean |
| ISO dates | date |
| currency symbols or amount+currency | money |
| all integers | integer |
| decimals | decimal |
| repeated finite values | enum |
| comma-separated repeated values | multi-select |
| URLs | url |
| image/document filenames | file/photo reference |
| otherwise | text |

The user must be able to override type inference before apply.

### Step 6 — Dry-run report

Dry run should show:

- rows to create
- rows to update/merge
- rows to skip
- duplicate candidates
- fields to create
- values with parse errors
- media files matched
- media files missing
- columns ignored
- warnings about private/sensitive metadata

### Step 7 — Apply

Apply creates an import batch.

Import batch stores:

- source file hash
- source app/type
- mapping definition
- created item IDs
- updated item IDs
- created field definitions
- parse errors/warnings
- media files imported
- timestamp

### Step 8 — Review

After apply:

- show import result report
- provide saved view: “Imported batch YYYY-MM-DD”
- allow rollback where feasible
- allow bulk edit from imported batch
- allow unresolved rows to remain in review queue

## Import batch schema draft

```json
{
  "schema": "cabinet.import-batch.v1",
  "import_batch_id": "imp_01H...",
  "profile_id": "profile_default",
  "source_kind": "homebox_csv",
  "source_files": [
    {
      "name": "homebox-export.csv",
      "hash": "sha256:..."
    }
  ],
  "mapping": {
    "Name": "core.name",
    "HB.location": "core.location",
    "HB.field.Serial Number": "custom.serial_number"
  },
  "created_field_ids": ["fld_serial_number"],
  "created_item_ids": ["item_001"],
  "updated_item_ids": [],
  "skipped_rows": [],
  "warnings": [],
  "created_at": "2026-07-10T00:00:00+10:00"
}
```

## Import refs and deduplication

Cabinet should support stable import references.

Import refs prevent duplicate items when importing updated files later.

Possible sources:

- Homebox `HB.import_ref`
- iCollect internal ID/export ID
- spreadsheet ID column
- generated source row hash
- barcode/serial/certification number when configured

Deduplication modes:

| Mode | Behaviour |
|---|---|
| create only | Always creates new records. |
| merge by import ref | Updates existing records with same import ref. |
| merge by barcode | Updates if barcode match exists. |
| merge by selected fields | Uses configured match rule. |
| review duplicates | Does not auto-merge; sends to review queue. |

## Unknown column preservation

Default rule:

> Preserve unknown columns as custom fields unless the user explicitly ignores them.

Column handling options:

```text
[Map to existing field]
[Create new custom field]
[Preserve as import note]
[Ignore column]
```

If a user ignores a column, Cabinet should include it in the import report so the decision is auditable.

## Media import

CSV-only import is not enough.

Cabinet should support:

- image/document folder import
- ZIP import
- filename matching by column
- media labels
- primary photo detection
- missing media warnings
- media hash calculation
- media manifest export

Example columns:

```text
Photo
Front Photo
Back Photo
Certificate File
Receipt File
Condition Report
```

Media matching rules:

- exact filename
- relative path
- case-insensitive fallback
- common extension fallback
- duplicate file warning
- unmatched media report

## Export requirements

Cabinet exports must not trap custom metadata.

Export should support:

| Export | Contents |
|---|---|
| CSV | Human-readable table with custom field columns. |
| JSON | Full schema, field definitions, values, templates, layout metadata. |
| ZIP bundle | JSON + CSV + media manifest + media files/thumbnails/originals depending options. |
| Template export | Field definitions + layout + views + reports. |

## Export schema structure

```text
cabinet-export.zip
  manifest.json
  data/
    items.json
    specimens.json
    collections.json
    locations.json
    custom-field-definitions.json
    custom-field-values.json
    templates.json
    saved-views.json
    reports.json
  csv/
    items.csv
    specimens.csv
  media/
    manifest.json
    originals/
    thumbnails/
```

## CSV export column naming

Use stable and readable names.

Options:

```text
core.name
core.location
field.serial_number
field.grading.certificate_number
field.gallery.artist
```

For Homebox compatibility exports, optionally support:

```text
HB.field.Serial Number
```

But Cabinet's canonical export should include both a schema and values, not just CSV headers.

## Privacy and export

Exports should make visibility explicit.

Export options:

- full private backup
- selected collection export
- public catalogue export
- trade binder export
- insurance report export
- media included/excluded
- originals or thumbnails only

Sensitive fields should warn before public/shareable export:

- storage locations
- purchase price
- insured value
- full collection value
- private notes
- contacts
- home address
- unreleased disputes

## Migration assistant presets

Cabinet should include source presets:

### Homebox preset

Recognise:

- `HB.import_ref`
- `HB.location`
- `HB.label`
- `HB.field.*`
- standard purchase/warranty/sold columns

Map:

- location to Cabinet location
- labels to Cabinet tags/labels
- fields to custom fields
- attachments flagged as missing if not in export

### iCollect preset

Recognise from user mapping/export patterns:

- collection type
- custom field labels
- value/currency fields
- rating fields
- country/region fields
- toggles
- display-related metadata where exported

Because iCollect export formats may vary, this should start as a configurable preset rather than a fragile parser.

### Generic CSV preset

- infer title/name
- infer location
- infer quantity
- preserve unknown columns
- ask user for template/category

## Acceptance criteria

Import/export is acceptable when:

- Unknown columns are preserved as custom fields by default.
- Users can map columns before applying an import.
- Cabinet can create field definitions during import.
- Cabinet can infer field types and let the user override.
- Dry-run preview shows creates/updates/skips/warnings.
- Import batch is auditable.
- Custom field definitions and values export together.
- Media references are handled through a manifest.
- Public exports warn for sensitive metadata.
- Re-importing with import refs does not create duplicates unless requested.

## Suggested backlog slices

1. Define OpenSpec for metadata import mapping.
2. Add import batch model.
3. Add CSV column analyser.
4. Add unknown-column-to-custom-field proposal model.
5. Add type inference for custom fields.
6. Add import mapping UI.
7. Add import dry-run report.
8. Add apply import with custom field creation.
9. Add Homebox CSV preset.
10. Add generic CSV preset.
11. Add media folder/ZIP matching plan.
12. Add JSON export of field definitions and values.
13. Add CSV export with custom field columns.
14. Add ZIP export with schema and media manifest.
15. Add import batch saved view and audit UI.
16. Add rollback/undo planning later.
