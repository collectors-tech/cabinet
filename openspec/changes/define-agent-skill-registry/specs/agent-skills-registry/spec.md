## ADDED Requirements

### Requirement: Agent Skill Registry SHALL expose first-class skill metadata
Cabinet SHALL expose Agent Skills as user-facing abilities that sit above lower-level capabilities, guided workflows, UI targets, integration workflows, shell command bus commands, preview/apply handlers, Action Timeline records, and audit records.

#### Scenario: List skill registry metadata
- **GIVEN** Cabinet has built-in and imported skills available for the active profile
- **WHEN** the skill registry lists skills
- **THEN** each skill SHALL include stable skill id, version, display name, description, category, source type, status, safety level, required context, required integrations/providers, permission declaration, and source provenance
- **AND** each skill SHALL declare its bound capability ids, guided workflow ids, UI target ids, integration workflow ids, shell command ids where applicable, input schema references, output schema references, and Action Timeline/audit behavior
- **AND** the registry SHALL distinguish user-facing skills from lower-level execution machinery instead of exposing capabilities/workflows/UI targets/commands as unrelated app internals

#### Scenario: Resolve one skill
- **GIVEN** a user or agent requests a specific skill id
- **WHEN** the registry resolves the skill
- **THEN** Cabinet SHALL return the same metadata shape used by list responses
- **AND** missing, blocked, disabled, invalid, or deprecated skills SHALL return deterministic state and next-action guidance instead of being silently omitted

### Requirement: Agent Skills SHALL distinguish built-in, imported, and future marketplace sources
Cabinet SHALL model skill source type without allowing imported or future marketplace skills to override built-in safety boundaries.

#### Scenario: Built-in and imported sources are separated
- **GIVEN** Cabinet has built-in skill manifests and imported skill records
- **WHEN** the registry merges installed skills for the active profile
- **THEN** built-in skills SHALL be marked `built-in` and SHALL NOT be removable or overridden by user import flows
- **AND** imported skills SHALL be marked `archive` with installation provenance and validation state
- **AND** future `marketplace` source metadata MAY be represented as a deferred source type but SHALL NOT imply remote discovery, publishing, ratings, payments, or public marketplace browsing

#### Scenario: Duplicate built-in id is rejected
- **GIVEN** a local archive declares a skill id that matches a built-in skill id
- **WHEN** Cabinet validates or imports the archive
- **THEN** the archive SHALL be rejected unless a future signed built-in update flow explicitly allows the replacement
- **AND** the built-in skill SHALL remain the effective registry entry

### Requirement: Agent Skills SHALL publish status and safety lifecycles
Cabinet SHALL use deterministic status and safety values so users can understand availability and risk before invoking a skill.

#### Scenario: Status reflects missing requirements
- **GIVEN** a skill depends on capabilities, workflows, UI targets, integrations, selections, permissions, or setup state
- **WHEN** one or more dependencies are unavailable for the active profile/context
- **THEN** the skill SHALL report one of `available`, `preview-only`, `requires-setup`, `requires-selection`, `disabled`, `deprecated`, `blocked`, or `invalid`
- **AND** the registry SHALL include required-action guidance for statuses that prevent execution
- **AND** missing dependencies SHALL NOT be hidden by marking the skill available

#### Scenario: Safety level governs invocation
- **GIVEN** a skill declares a safety level
- **WHEN** Cabinet presents or invokes the skill
- **THEN** the safety level SHALL be one of `read-only`, `preview-only`, `confirm-required`, `external-write`, or `destructive`
- **AND** `read-only` skills SHALL NOT mutate Cabinet or external state
- **AND** `preview-only`, `confirm-required`, `external-write`, and `destructive` skills SHALL preserve Cabinet preview/confirm/apply boundaries according to their declared risk
- **AND** `external-write` and `destructive` skills SHALL require stronger confirmation copy and Action Timeline/audit evidence before apply

### Requirement: Skill permissions and context SHALL be explicit before execution
Cabinet SHALL require skills to declare the profile, route, selection, provider, permission, and data-access context needed to run safely.

#### Scenario: Missing context asks for clarification
- **GIVEN** a skill requires an active profile, selected item, selected collection, provider setup, integration credential, attachment, or route context
- **WHEN** the user invokes the skill without the required context
- **THEN** Cabinet SHALL return a clarification or setup-needed response instead of guessing a target
- **AND** the response SHALL identify the missing context and the next safe user action

#### Scenario: Permissions are visible and auditable
- **GIVEN** a skill can read, write, import, export, delete, configure, or call an external provider
- **WHEN** the registry or Skills page displays the skill
- **THEN** Cabinet SHALL display a permission declaration that separates Cabinet-local reads, Cabinet-local writes, external reads, external writes, secret access, and destructive operations
- **AND** any applied write or external operation SHALL create Action Timeline/audit evidence with non-secret payload references

### Requirement: Skill archives SHALL use a bounded local import structure
Cabinet SHALL support local `.cabinet-skill.zip` archives and local development folders as import sources without creating a public marketplace.

#### Scenario: Archive manifest structure is validated
- **GIVEN** a local skill archive or folder is submitted for validation
- **WHEN** Cabinet inspects the archive
- **THEN** it SHALL require a `cabinet.skill.json` manifest
- **AND** the manifest SHALL include schema, id, version, displayName, description, category, source, safetyLevel, status, modes, permissions, compatibility, and audit declarations
- **AND** optional bindings MAY include capabilities, guidedWorkflows, uiTargets, integrationRequirements, inputSchemaRef, outputSchemaRef, workflows, examples, and checksums

#### Scenario: Archive safety constraints are enforced
- **GIVEN** a submitted archive contains unsupported files, executable/native code, unsafe paths, too many files, excessive size, invalid checksums, unknown schema, or unsupported manifest values
- **WHEN** Cabinet validates the archive
- **THEN** validation SHALL fail before installation
- **AND** failure output SHALL identify the unsafe or invalid condition without extracting files outside the intended root or enabling the skill
- **AND** imported skills SHALL NOT execute arbitrary code unless a future explicit design adds that capability with separate safety requirements

### Requirement: Skill import SHALL produce safe validation and install states
Cabinet SHALL separate validation from installation and keep imported skills disabled or blocked until their safety and dependencies are clear.

#### Scenario: Import result states are deterministic
- **GIVEN** a local archive or folder import is requested
- **WHEN** Cabinet validates and installs it
- **THEN** the result SHALL be one of `valid-ready-to-install`, `valid-with-warnings`, `blocked-missing-dependency`, `blocked-invalid-manifest`, `blocked-unsafe-archive`, `installed-disabled`, or `installed-enabled`
- **AND** invalid, unsafe, or dependency-blocked archives SHALL NOT appear as enabled or executable skills

#### Scenario: Imported skill installs disabled by default when risk is not read-only
- **GIVEN** an archive validates successfully
- **WHEN** the skill safety level is `preview-only`, `confirm-required`, `external-write`, or `destructive`
- **THEN** Cabinet SHALL install the skill disabled by default or require explicit enable confirmation according to product policy
- **AND** enabling a skill SHALL preserve status, permission, and audit evidence

### Requirement: Installed skill state SHALL be profile-scoped and reversible
Cabinet SHALL persist installed skill state separately from immutable built-in manifests and imported archive provenance.

#### Scenario: Enable or disable imported skill
- **GIVEN** an imported skill is installed for the active profile
- **WHEN** the user enables or disables the skill
- **THEN** Cabinet SHALL persist profile-scoped enabled state
- **AND** disabled skills SHALL remain visible with status `disabled` but SHALL NOT be executable
- **AND** built-in skills SHALL NOT be removed by import/delete flows

#### Scenario: Invalid installed skill remains safe
- **GIVEN** a previously installed skill becomes invalid due to a schema, dependency, compatibility, or permission problem
- **WHEN** the registry refreshes state
- **THEN** Cabinet SHALL mark the skill `invalid` or `blocked`
- **AND** the skill SHALL remain non-executable until validation passes again
- **AND** the user SHALL receive clear repair or remove guidance where removal is allowed

### Requirement: Skills page SHALL manage installed skills without marketplace browsing
Cabinet SHALL provide a Skills page for local skill management while clearly deferring public marketplace behavior.

#### Scenario: Skills page list and summary
- **GIVEN** a user opens the Cabinet Skills page
- **WHEN** the page loads registry data
- **THEN** it SHALL show summary counts for installed skills, enabled skills, needs-setup skills, and blocked/invalid skills
- **AND** it SHALL list skill name, category, source, status, safety level, required setup/context, version, and available actions
- **AND** built-in and imported skills SHALL be visually distinguishable without requiring separate hardcoded cards

#### Scenario: Skills page detail panel
- **GIVEN** a user selects a skill from the Skills page
- **WHEN** the detail panel opens
- **THEN** it SHALL show description, permissions, safety level, bound capabilities, guided workflows, UI targets, integrations required, audit/Action Timeline behavior, validation warnings/errors, source provenance, and enable/disable affordance where allowed
- **AND** icon-only actions SHALL have accessible names

#### Scenario: Skills page import flow
- **GIVEN** a user chooses `Import skill`
- **WHEN** the import modal or drawer opens
- **THEN** it SHALL support choosing a `.cabinet-skill.zip` archive or local development folder where supported
- **AND** it SHALL show validation progress, validation warnings/errors, metadata preview before install, and install result state
- **AND** it SHALL state that local import is available and marketplace browsing is not available yet

### Requirement: Skill execution SHALL preserve existing Cabinet governance
Agent Skills SHALL never bypass the existing governed assistant execution model.

#### Scenario: Skill invokes lower-level machinery through policy
- **GIVEN** a user invokes a skill that binds to a capability, guided workflow, UI target, integration workflow, or shell command
- **WHEN** Cabinet prepares execution
- **THEN** it SHALL route through the existing governed preview/apply, command bus, provider readiness, Action Timeline, and audit boundaries declared by the binding
- **AND** direct mutation SHALL be blocked unless the skill safety level and bound capability both permit a confirmed apply
- **AND** execution evidence SHALL reference the skill id, lower-level binding id, profile/thread/context, confirmation state, and non-secret result or error evidence

#### Scenario: Skill invocation preserves source context
- **GIVEN** a supported in-app surface, Inbox review item, Chat thread, or authorized external channel invokes an Agent skill
- **WHEN** Cabinet creates a skill preview or confirmed apply response
- **THEN** the response SHALL preserve the non-secret source surface, source channel, source thread id, and source message id supplied by the invoking boundary
- **AND** preview responses SHALL keep `mutation_applied=false` until a confirmed apply request is accepted
- **AND** confirmed apply responses SHALL retain the same source context while reporting whether the mutation was applied or blocked

### Requirement: Marketplace behavior SHALL remain explicitly deferred
Cabinet SHALL not treat local skill import support as a public marketplace implementation.

#### Scenario: Marketplace unavailable state
- **GIVEN** a user views the Skills page or import flow
- **WHEN** they look for public discovery, publishing, ratings, payments, reviews, remote marketplace search, or marketplace installation
- **THEN** Cabinet SHALL show those behaviors as unavailable or deferred
- **AND** no remote marketplace content SHALL be fetched or implied by this registry/import contract

### Requirement: Agent skill coverage SHALL stay mapped by Cabinet surface
Cabinet SHALL maintain a durable per-surface Agent skill coverage matrix so broad "Agent can do the work" claims remain traceable to exact surfaces, skill ids, safety levels, dependencies, statuses, issue links, and validation evidence.

#### Scenario: Maintain per-surface coverage matrix
- **GIVEN** Agent skill implementation is spread across inventory, wishlist, collections, media, discoveries, market watch, purchases, integrations, inbox, users, settings, storage, data-management, chats, and external channels
- **WHEN** Cabinet reports Agent skill coverage
- **THEN** the coverage matrix SHALL list each major Cabinet surface with the expected skill ids, user-facing request examples, safety level, required context or selection, required integration/provider setup, bound capability ids, bound guided workflow ids, implementation status, missing issue or PR links, validation evidence, and blocked/deferred reason where applicable
- **AND** the matrix SHALL distinguish in-app main Chat, side-panel Chat, Inbox review, and Telegram/external channels instead of treating all Agent entry points as equivalent
- **AND** planned, blocked, and deferred entries SHALL stay explicit until their linked issues provide implementation, validation, and closure evidence

#### Scenario: Wishlist and Collections skill surfaces expose governed preview boundaries
- **GIVEN** Wishlist and Collections are registered as Cabinet Agent skill surfaces
- **WHEN** Cabinet lists or previews their built-in skills
- **THEN** Wishlist SHALL expose `cabinet.wishlist.search_entries`, `cabinet.wishlist.create_entry`, `cabinet.wishlist.update_entry`, `cabinet.wishlist.mark_purchased`, `cabinet.wishlist.soft_delete_entry`, and `cabinet.wishlist.restore_entry` with read-only search and confirmation-gated mutation safety
- **AND** Collections SHALL expose `cabinet.collections.search`, `cabinet.collections.create`, `cabinet.collections.update_metadata`, `cabinet.collections.assign_item`, `cabinet.collections.soft_delete`, and `cabinet.collections.move_items_on_delete` with read-only search and confirmation-gated mutation safety
- **AND** previews SHALL block missing wishlist item or entry context, missing collection item or destination context, and attempts to delete `All Items` before any mutation is applied
- **AND** the Wishlist purchased preview contract SHALL identify purchase lifecycle and inventory quantity sync evidence as required before the runtime apply path can be treated as complete
