# agent-skills-registry Specification

## Purpose
Define how Cabinet presents, resolves, validates, and executes Agent Skills as user-facing capabilities with explicit metadata, safety levels, permissions, context requirements, and audit behavior.

## Requirements
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

### Requirement: AGENT-SKILLS-REGISTRY-013 Users administration SHALL derive authority only from the server session
Cabinet Agent SHALL bind Users/admin work to the authenticated server-side actor and active profile membership rather than caller-authored context.

#### Scenario: Users administration derives authority only from the server session
- **GIVEN** Cabinet Agent receives a Users/admin request through an in-app or future external entry point
- **WHEN** Cabinet authorizes planner, read, preview, apply, cancel, replay, and audit access
- **THEN** it SHALL derive the actor, role, and profile scope from a server-validated session and active profile membership
- **AND** client-supplied `admin_session`, role, permission, or authority values SHALL never grant or expand access
- **AND** a local session SHALL match the active profile and its protected owner/admin membership
- **AND** a remote session SHALL carry the validated admin claim and match an active admin membership in that profile
- **AND** missing, expired, wrong-profile, inactive, or non-admin authority SHALL fail closed before user lists, mutation targets, or admin workflow evidence are returned
- **AND** every Users/admin mutation SHALL revalidate authority at execution time while protected owner safeguards remain enforced
- **AND** responses and evidence SHALL omit the opaque session token, cookie, secrets, and client-asserted authority values while retaining non-secret server-derived actor and decision evidence

### Requirement: AGENT-SKILLS-REGISTRY-014 Destructive Agent skills SHALL require action-specific strong confirmation
Cabinet Agent SHALL require a fresh server-issued confirmation for each exact destructive target instead of accepting caller-authored booleans or freeform confirmation text.

#### Scenario: Review and confirm one exact destructive action
- **GIVEN** Chat has created a durable preview for user removal or backup restore
- **WHEN** the authorized user requests the destructive impact review
- **THEN** Cabinet SHALL revalidate the current server session, profile, skill, preview expiry, protected-owner rules, backup compatibility, and exact current target
- **AND** the review SHALL show the action, exact non-secret target, impacts, recovery path, and expiry in both full and contextual Chat
- **AND** Cabinet SHALL issue an action-specific, profile-scoped, five-minute, single-use token whose stored representation is hashed
- **AND** only the dedicated visible confirmation control SHALL submit that token with the same opaque preview id

#### Scenario: Destructive confirmation fails closed across lifecycle boundaries
- **GIVEN** Cabinet issued a strong confirmation for an exact destructive preview and target
- **WHEN** a caller supplies only `confirm=true`, replays or supersedes a token, changes the target, changes profile, cancels the preview, or waits past expiry
- **THEN** Cabinet SHALL reject the mutation without applying it
- **AND** protected or last-owner removal SHALL remain blocked before a confirmation can be issued
- **AND** cancellation, expiry, confirmation issuance, and successful apply SHALL retain non-secret audit evidence without the bearer token or raw sensitive payload

#### Scenario: Restore revalidates compatible backup identity and preserves recovery
- **GIVEN** a destructive restore preview names a backup in Cabinet's managed backup directory
- **WHEN** Cabinet issues and consumes strong confirmation
- **THEN** it SHALL revalidate the archive format, backup format, database integrity, selected database hash, compatible schema, and active profile scope immediately before replacement
- **AND** it SHALL create a pre-restore recovery backup before applying the selected backup
- **AND** the terminal durable preview and used-confirmation receipt SHALL survive the restored database state without persisting the raw token

### Requirement: AGENT-SKILLS-REGISTRY-015 Integration configure-provider Chat previews SHALL carry canonical setup context
Cabinet Agent SHALL let natural-language Chat configure integration providers through the governed Agent Skill preview lifecycle without requiring clients to own provider setup authority or echo secrets.

#### Scenario: Configure an integration provider through Chat preview and token apply
- **GIVEN** Chat selects `cabinet.integrations.configure_provider` from a user-friendly provider/setup request
- **WHEN** the planner returns friendly provider name, catalogue, setup metadata, or optional credential fields
- **THEN** Cabinet SHALL expose planner schema fields for provider identity and non-secret setup payload, keep provider secrets optional and write-only, and normalize friendly values into canonical `provider_id` and setup parameters before preview creation
- **AND** the durable preview SHALL persist enough non-secret canonical context for token-only apply to revalidate authority and execute without `missing_context`
- **AND** confirmed apply SHALL persist enabled provider setup metadata for the active profile, omit secrets from response/log evidence, and reject replay or cross-profile confirmation without a second mutation
- **AND** read-only integration provider search and explanation skills SHALL remain non-mutating and unchanged by configure-provider normalization

#### Scenario: Canonicalize live Browser Auth setup prose before preview persistence
- **GIVEN** a Browser Auth provider selects `cabinet.integrations.configure_provider` but returns a friendly setup sentence containing profile context or a negated API-key instruction
- **WHEN** that sentence clearly requests a provider's public catalogue
- **THEN** Cabinet SHALL canonicalize it to `setup_payload=public_catalogue`, `setup_step=public_catalogue`, and `marketplace=public` before durable preview creation
- **AND** profile identifiers, prompt prose, API-key wording, and secret wording SHALL NOT enter persisted provider setup fields
- **AND** confirmed apply and replay protection SHALL retain the same governed token-only behavior

#### Scenario: Discard empty optional secret fields from Browser Auth plans
- **GIVEN** a Browser Auth provider selects `cabinet.integrations.configure_provider` and includes an empty or null optional secret-shaped field
- **WHEN** Cabinet normalizes the provider plan before durable preview creation
- **THEN** Cabinet SHALL discard the empty field so a credential-free Browser Auth preview does not invoke secure secret storage
- **AND** any non-empty secret value SHALL remain governed, write-only, redacted from evidence, and stored only through the secure secret boundary

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

#### Scenario: Wishlist skill confirmed apply persists governed state
- **GIVEN** a confirmed Wishlist Agent Skill apply request includes the required profile and target context
- **WHEN** Cabinet applies create, update, mark-purchased, soft-delete, restore, or read-only search Wishlist skills
- **THEN** create/update/delete/restore actions SHALL persist through the Wishlist service instead of bypassing the product data model
- **AND** read-only search SHALL return matching Wishlist entries without mutating Cabinet state
- **AND** mark-purchased SHALL report purchase lifecycle and inventory quantity sync evidence while preserving the Wishlist service rule that repeated confirmed applies do not duplicate inventory quantity increments
- **AND** source surface, source channel, source thread id, and source message id SHALL remain visible in the preview/apply response when supplied by the invoking boundary

#### Scenario: Collections skill confirmed apply persists governed workspace state
- **GIVEN** a confirmed Collections Agent Skill apply request includes the required profile and collection or item context
- **WHEN** Cabinet applies create, update-metadata, assign-item, soft-delete, move-items-on-delete, or read-only search Collections skills
- **THEN** create/update/delete/move actions SHALL persist the profile-scoped Collections workspace state instead of returning preview-only claims
- **AND** assign-item SHALL verify the target inventory item belongs to the active profile before adding workspace collection membership
- **AND** read-only search SHALL return matching collections and workspace item memberships without mutating Cabinet state
- **AND** `All Items` SHALL remain protected from deletion or unsafe rename during confirmed apply
- **AND** source surface, source channel, source thread id, and source message id SHALL remain visible in the preview/apply response when supplied by the invoking boundary

#### Scenario: Inventory skill surfaces expose governed preview boundaries
- **GIVEN** Inventory is registered as a Cabinet Agent skill surface
- **WHEN** Cabinet lists or previews its built-in skills
- **THEN** Inventory SHALL expose `cabinet.inventory.search_items`, `cabinet.inventory.create_item`, `cabinet.inventory.update_item`, `cabinet.inventory.attach_media`, and `cabinet.inventory.assign_to_collection` with read-only search and confirmation-gated mutation safety
- **AND** previews SHALL block missing item details, missing selected item, missing explicit media, missing collection, and invalid deleted/trash collection targets before any mutation is applied
- **AND** the guided `cabinet.guided.inventory.update_item` skill SHALL remain non-executable until the guided update walkthrough dependency is validated

#### Scenario: Inventory skill confirmed apply persists governed state
- **GIVEN** a confirmed Inventory Agent Skill apply request includes the required profile and target context
- **WHEN** Cabinet applies create-item, update-item, attach-media, assign-to-collection, or read-only search Inventory skills
- **THEN** create and update actions SHALL persist through the Inventory item repository instead of returning preview-only claims
- **AND** read-only search SHALL return matching active-profile Inventory items without mutating Cabinet state
- **AND** media attach SHALL require an explicit existing media asset and persist a profile-scoped inventory media link with provenance
- **AND** collection assignment SHALL verify the item belongs to the active profile and persist the workspace collection membership
- **AND** source surface, source channel, source thread id, and source message id SHALL remain visible in the preview/apply response when supplied by the invoking boundary

### Requirement: Mutating Agent Skill previews SHALL use durable single-use confirmation tokens
Cabinet SHALL represent each actionable confirmation-required Agent Skill preview as a server-owned, profile-scoped record identified by an opaque token rather than treating a client-carried apply payload as mutation authority.

#### Scenario: Confirm a durable Agent Skill preview exactly once
- **GIVEN** Cabinet has created an actionable mutating Agent Skill preview for a profile, skill, source surface, source channel, source thread, source message, and bounded target
- **WHEN** the owning profile confirms the opaque preview id before expiry
- **THEN** Cabinet SHALL reload the server-owned request, re-evaluate current Agent authority and target validity, atomically claim the pending preview, and execute the bound skill at most once
- **AND** a duplicate confirmation, concurrent replay, cross-profile request, mismatched supplied source context, stale target, or changed authority SHALL fail closed without a second mutation
- **AND** the applied response and Action Timeline evidence SHALL preserve the preview id, skill id, bounded source provenance, confirmation state, and non-secret result summary

#### Scenario: Cancel or expire a durable Agent Skill preview
- **GIVEN** a durable Agent Skill preview is pending
- **WHEN** the owning profile cancels it or its expiry time passes before confirmation
- **THEN** Cabinet SHALL transition it to a terminal cancelled or expired state and reject every later apply attempt
- **AND** the rejection SHALL provide clear recoverable guidance to create a fresh preview without mutating Cabinet state

#### Scenario: Pending preview secrets remain write-only
- **GIVEN** a governed provider or settings preview contains credential parameters
- **WHEN** Cabinet persists the durable preview
- **THEN** the preview row, API response, logs, workflow evidence, and UI SHALL contain only redacted parameters, bounded target provenance, and an opaque secret reference
- **AND** pending credential material SHALL use Cabinet's encrypted profile secret storage and SHALL be removed when the preview is applied, cancelled, expired, or fails after claim
- **AND** legacy Inventory Chat action previews SHALL retain their established preview/apply API compatibility while generic Agent Skill consumers migrate to the durable token contract
