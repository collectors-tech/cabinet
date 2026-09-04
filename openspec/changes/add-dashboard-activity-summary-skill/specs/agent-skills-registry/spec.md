## MODIFIED Requirements

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

#### Scenario: Dashboard activity summary skill is read-only and profile-scoped
- **GIVEN** the active profile has Dashboard data available
- **WHEN** Cabinet lists, resolves, previews, or executes `cabinet.dashboard.summarise_activity`
- **THEN** the skill SHALL be registered as built-in, read-only, profile-scoped, and available without external provider setup
- **AND** direct execution SHALL return structured current totals, attention signals, recent items, destination links, record identifiers, source context, and non-secret evidence
- **AND** the skill SHALL NOT create mutation previews, confirmation tokens, external writes, database writes outside non-secret audit/workflow evidence, or applied mutations
- **AND** profile mismatch or unavailable Dashboard dependencies SHALL return deterministic guidance without leaking another profile's data
