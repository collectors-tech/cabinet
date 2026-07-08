# Agent Skill Archive Template

Issue: #1671
Related OpenSpec requirements: AGENT-SKILLS-REGISTRY-005 and AGENT-SKILLS-REGISTRY-010

Use this guide for local Cabinet Agent Skill archive and development-folder imports. It defines the template target for `.cabinet-skill.zip` packages before Cabinet has a public marketplace or external template repository.

## Local Import Scope

Supported local sources:

- `.cabinet-skill.zip` archive
- unpacked development folder with the same structure

Deferred sources:

- marketplace discovery
- remote repository install
- publishing, payments, ratings, reviews, and public search

Local import must not imply marketplace availability. Marketplace or remote repository support needs a separate signed-source design before Cabinet can trust remote packages.

## Archive Layout

Archive roots must contain `cabinet.skill.json` at the root. The archive root may be the zip root itself or one top-level folder such as `cabinet-skill-template/`.
Key template paths include `schemas/input.schema.json`, `schemas/output.schema.json`, `workflows/guided-workflow.json`, `ui-targets/ui-targets.json`, `tests/skill-validation.json`, `examples/sample-input.json`, and `examples/sample-output.json`.

```text
cabinet-skill-template/
  cabinet.skill.json
  README.md
  CHANGELOG.md
  LICENSE
  schemas/
    input.schema.json
    output.schema.json
  workflows/
    guided-workflow.json
  ui-targets/
    ui-targets.json
  tests/
    skill-validation.json
  examples/
    sample-input.json
    sample-output.json
```

Allowed files are declarative skill metadata, Markdown documentation, JSON schemas, JSON workflow descriptors, UI target declarations, examples, validation fixtures, and license/changelog text. Archives must not include executable or native code such as `.exe`, `.dll`, `.so`, `.dylib`, `.bat`, `.cmd`, `.ps1`, `.sh`, `.js`, `.ts`, `.go`, `.py`, or compiled binaries unless a future design explicitly adds a signed execution model.

## Manifest Fields

`cabinet.skill.json` is required. The initial schema name is `https://collectors.tech/cabinet/schemas/agent-skill.v1.json`.

Required fields:

- `schema`
- `id`
- `version`
- `displayName`
- `description`
- `category`
- `author`
- `source`
- `license`
- `safetyLevel`
- `status`
- `modes`
- `permissions`
- `compatibility`
- `audit`

Optional fields:

- `homepage`
- `capabilities`
- `guidedWorkflows`
- `uiTargets`
- `integrationRequirements`
- `inputSchemaRef`
- `outputSchemaRef`
- `checksums`

Field rules:

- `id` must be stable, lower-case, dot-separated, and prefixed with an owned namespace such as `cabinet.example.open_inventory_help`.
- `version` must be semantic version text such as `1.0.0`.
- `source.type` must be `archive` for local archives and may include a future-ready provenance object, but it must not claim marketplace origin.
- `safetyLevel` must be one of `read-only`, `preview-only`, `confirm-required`, `external-write`, or `destructive`.
- `status` must be one of `available`, `preview-only`, `requires-setup`, `requires-selection`, `disabled`, `deprecated`, `blocked`, or `invalid`.
- `permissions` must explicitly separate Cabinet-local reads, Cabinet-local writes, external reads, external writes, secret access, and destructive operations.
- `audit.actionTimeline` must describe the non-secret Action Timeline evidence the skill creates or depends on.
- `checksums` must use paths inside the archive and digest values that match the corresponding file bytes.

## Complete Manifest Example

```json
{
  "schema": "https://collectors.tech/cabinet/schemas/agent-skill.v1.json",
  "id": "cabinet.example.open_inventory_help",
  "version": "1.0.0",
  "displayName": "Open inventory help",
  "description": "Explains the Inventory surface and opens the supported help target without changing data.",
  "category": "inventory",
  "author": {
    "name": "Example Skill Author",
    "url": "https://example.invalid/cabinet-skills"
  },
  "homepage": "https://example.invalid/cabinet-skills/open-inventory-help",
  "source": {
    "type": "archive",
    "provenance": {
      "importedFrom": "open-inventory-help.cabinet-skill.zip"
    }
  },
  "license": "MIT",
  "safetyLevel": "read-only",
  "status": "available",
  "modes": ["in-app", "assistant"],
  "capabilities": ["cabinet.navigate.open_surface"],
  "guidedWorkflows": [],
  "uiTargets": ["settings.help-center", "inventory.items"],
  "integrationRequirements": [],
  "permissions": {
    "cabinetReads": ["inventory.help", "ui.route"],
    "cabinetWrites": [],
    "externalReads": [],
    "externalWrites": [],
    "secretAccess": false,
    "destructive": false
  },
  "inputSchemaRef": "schemas/input.schema.json",
  "outputSchemaRef": "schemas/output.schema.json",
  "audit": {
    "actionTimeline": "records non-secret help route and source surface",
    "requiresConfirmation": false
  },
  "checksums": {
    "schemas/input.schema.json": "sha256:REPLACE_WITH_FILE_DIGEST",
    "schemas/output.schema.json": "sha256:REPLACE_WITH_FILE_DIGEST"
  },
  "compatibility": {
    "cabinetMinVersion": "0.1.0",
    "schemaVersion": "v1"
  }
}
```

## Confirm-Required Example

```json
{
  "schema": "https://collectors.tech/cabinet/schemas/agent-skill.v1.json",
  "id": "cabinet.example.update_item_guided",
  "version": "1.0.0",
  "displayName": "Guided item update",
  "description": "Previews an item update and requires confirmation before changing inventory data.",
  "category": "inventory",
  "author": {
    "name": "Example Skill Author"
  },
  "source": {
    "type": "archive"
  },
  "license": "MIT",
  "safetyLevel": "confirm-required",
  "status": "preview-only",
  "modes": ["in-app", "assistant"],
  "capabilities": ["cabinet.inventory.update_item"],
  "guidedWorkflows": ["cabinet.guided.inventory.update_item"],
  "uiTargets": ["inventory.item.detail"],
  "integrationRequirements": [],
  "permissions": {
    "cabinetReads": ["inventory.item"],
    "cabinetWrites": ["inventory.item"],
    "externalReads": [],
    "externalWrites": [],
    "secretAccess": false,
    "destructive": false
  },
  "inputSchemaRef": "schemas/input.schema.json",
  "outputSchemaRef": "schemas/output.schema.json",
  "audit": {
    "actionTimeline": "records preview id, confirmation state, item id, and non-secret changed fields",
    "requiresConfirmation": true
  },
  "compatibility": {
    "cabinetMinVersion": "0.1.0",
    "schemaVersion": "v1"
  }
}
```

## Validation Rules

Cabinet validation must reject an archive before installation when any of these are true:

- `cabinet.skill.json` is missing or not valid JSON.
- `schema`, `id`, `version`, `safetyLevel`, `status`, `permissions`, `compatibility`, or `audit` is missing or unrecognised.
- `id` matches a built-in skill id.
- declared capabilities, guided workflows, UI targets, or integrations are unknown and not marked as blocked or missing.
- permissions are implicit, broad, or inconsistent with `safetyLevel`.
- any path is absolute, contains `..`, uses unsafe separators, or escapes the extraction root.
- archive size, file count, or individual file size exceeds Cabinet bounds.
- unsupported files, executable files, native code, or hidden payloads are present.
- a declared checksum is missing, malformed, or does not match the file bytes.

Validation output should use the import result states from OpenSpec: `valid-ready-to-install`, `valid-with-warnings`, `blocked-missing-dependency`, `blocked-invalid-manifest`, `blocked-unsafe-archive`, `installed-disabled`, and `installed-enabled`.

## Safety Boundaries

Template skills may bind to Cabinet capabilities, guided workflows, UI targets, and integration requirements. They must not bypass:

- preview/confirm/apply boundaries for mutations
- Action Timeline and audit records
- integration setup checks
- UI target allowlists
- destructive-action confirmations
- archive validation
- profile-scoped enabled/disabled state

Valid imported skills are installed disabled by default unless they are read-only and the active product policy explicitly permits auto-enable. Invalid archives must never appear as enabled or executable.

## Packaging Checklist

Before packaging a `.cabinet-skill.zip`:

- Put `cabinet.skill.json` at the archive root or inside one top-level root folder.
- Keep all paths relative and inside the archive root.
- Include README, changelog, license, schemas, examples, and workflow/UI target declarations when referenced.
- Remove executable/native files and unsupported sources.
- Set explicit permissions and safety level.
- Mark missing dependencies as blocked or requires setup instead of claiming availability.
- Calculate checksums after the final file contents are in place.
- Validate that marketplace publishing, payments, ratings, reviews, and remote discovery are not documented or implied.
