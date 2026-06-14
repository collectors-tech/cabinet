## Purpose
Define Media workspace behavior for uploaded assets, unlinked filtering, and shared table-first management workflows.

## Requirements
### Requirement UI-SCREEN-MEDIA-001: Media workspace SHALL provide table-first default view for uploaded assets
Cabinet SHALL provide a dedicated media section with Cabinet's shared table surface as default rendering mode.

#### Scenario: Open media workspace default view
- **GIVEN** user navigates to `/media` media section
- **WHEN** workspace loads successfully
- **THEN** UI MUST render the shared table management surface by default and keep card mode available as a secondary view

### Requirement UI-SCREEN-MEDIA-002: Media workspace SHALL provide unlinked-assets filter
Cabinet SHALL support a first-class view/filter for media not linked to inventory or wishlist targets.

#### Scenario: View unlinked assets
- **GIVEN** media library contains both linked and unlinked assets
- **WHEN** user selects "Unlinked" filter/tab
- **THEN** UI MUST show only assets with no inventory/wishlist linkage and display deterministic empty state when none exist

### Requirement UI-SCREEN-MEDIA-003: Media cards SHALL expose operational metadata and quick actions
Media cards SHALL include details required for review and assignment workflows.

#### Scenario: Review media card details
- **GIVEN** media card is visible in Media workspace
- **WHEN** card renders details panel
- **THEN** card MUST include thumbnail/preview, upload timestamp, analysis status, confidence indicator (when analyzed), and quick actions (`analyze`, `assign`, `open`, `archive`)

### Requirement MEDIA-LINKAGE-001: Media linkage state SHALL be deterministic across inventory and wishlist targets
Cabinet SHALL classify each media asset into stable linkage states for filtering and assignment logic.

#### Scenario: Linkage state classification
- **GIVEN** media asset may be assigned to inventory item(s), wishlist entry(ies), both, or neither
- **WHEN** linkage state is computed for API/UI
- **THEN** runtime MUST return one of: `unlinked`, `linked_inventory`, `linked_wishlist`, `linked_both`

### Requirement MEDIA-LINKAGE-002: Assignment actions SHALL update linkage state and card indicators in same interaction cycle
Cabinet SHALL apply assignment changes and reflect updated linkage status without stale UI state.

#### Scenario: Assign unlinked media to wishlist
- **GIVEN** asset state is `unlinked` and user confirms assign-to-wishlist action
- **WHEN** assignment request succeeds
- **THEN** linkage state MUST update to `linked_wishlist` (or `linked_both` if inventory link exists) and card badges/filters MUST reflect new state immediately

### Requirement MEDIA-STORAGE-001: Upload pipeline SHALL store assets in a global media registry
Cabinet SHALL ingest uploaded binaries into a shared media library before linkage to inventory or wishlist targets.

#### Scenario: Upload inventory image into global library
- **GIVEN** user uploads image from inventory workflow
- **WHEN** upload is accepted
- **THEN** runtime MUST persist media asset in global registry and return canonical `asset_id` used for downstream linking

### Requirement MEDIA-STORAGE-002: Media ingestion SHALL enforce content-hash deduplication
Cabinet SHALL compute deterministic content hash for each uploaded binary and prevent duplicate binary storage.

#### Scenario: Re-upload existing image bytes
- **GIVEN** uploaded image bytes match an existing asset hash
- **WHEN** ingestion computes content hash
- **THEN** runtime MUST reuse existing asset record, MUST NOT store duplicate binary, and MUST return dedupe metadata (`deduped=true`, existing `asset_id`)

### Requirement MEDIA-STORAGE-003: Inventory and wishlist assignment SHALL use link records instead of binary duplication
Cabinet SHALL associate media assets to inventory/wishlist via linkage records.

#### Scenario: Assign shared asset to multiple targets
- **GIVEN** asset already exists in media registry
- **WHEN** user assigns asset to inventory item and wishlist entry
- **THEN** runtime MUST create/update link records per target and MUST NOT create duplicate media binaries

### Requirement MEDIA-STORAGE-004: Derived variants SHALL be generated once per canonical asset
Cabinet SHALL generate preview/thumbnail variants at asset level and reuse them across linked targets.

#### Scenario: Variant reuse across links
- **GIVEN** asset is linked to multiple inventory/wishlist targets
- **WHEN** target views request preview/thumbnail
- **THEN** runtime MUST serve shared derived variants for same `asset_id` without per-link regeneration

### Requirement MEDIA-STORAGE-005: Upload response SHALL expose deterministic dedupe and linkage envelope
Cabinet SHALL return a stable upload response contract for UI and E2E assertions.

#### Scenario: Upload response contract
- **GIVEN** upload request completes (new or deduped)
- **WHEN** API returns upload result
- **THEN** response MUST include `asset_id`, `content_hash`, `deduped`, variant availability, and current linkage summary

### Requirement MEDIA-STORAGE-006: Media storage root path SHALL be user-configurable from Settings
Cabinet SHALL allow users to configure where media binaries are physically stored via Settings.

#### Scenario: Configure media root path in Settings
- **GIVEN** user opens **Settings > Storage** and enters a valid writable folder path
- **WHEN** user saves configuration
- **THEN** runtime MUST persist `storage.media_root_path` and subsequent uploads MUST store files under that configured root

#### Scenario: Display configured media root path in Settings
- **GIVEN** media root path is already configured
- **WHEN** user opens **Settings > Storage**
- **THEN** UI MUST display current configured path, validation status, and last successful write-check result

### Requirement MEDIA-STORAGE-007: Media root path validation SHALL be deterministic and actionable
Cabinet SHALL validate configured storage path and return clear errors for invalid/unwritable locations.

#### Scenario: Invalid media root path
- **GIVEN** configured media root path is invalid, missing, or not writable
- **WHEN** upload or variant generation is attempted
- **THEN** runtime MUST fail with deterministic error code and guidance to update storage settings

### Requirement MEDIA-STORAGE-008: Media path changes SHALL support controlled migration
Cabinet SHALL support optional migration of existing media binaries when storage root path is changed.

#### Scenario: Move existing media to new root
- **GIVEN** user updates media root path and chooses migrate-existing option
- **WHEN** migration is executed
- **THEN** runtime MUST move/copy existing assets safely, preserve linkage integrity, and report migration summary with success/failure counts

### Requirement MEDIA-STORAGE-009: Assets SHALL support human-readable download filenames after analysis
Cabinet SHALL generate user-friendly file names from resolved item metadata (for example part number + title) without breaking canonical storage identity.

#### Scenario: Generate friendly filename from analyzed metadata
- **GIVEN** asset has resolved metadata such as `part_number` and `title`
- **WHEN** user requests download or share URL
- **THEN** runtime MUST provide human-readable filename (slug-safe) while preserving canonical `asset_id` and content-hash identity internally

### Requirement MEDIA-STORAGE-010: URL-safe naming SHALL be deterministic and collision-safe
Cabinet SHALL generate deterministic URL-safe names and handle collisions predictably.

#### Scenario: Friendly-name collision handling
- **GIVEN** two assets resolve to same friendly base name
- **WHEN** download/share URLs are generated
- **THEN** runtime MUST keep deterministic unique output naming (for example suffix with short hash/token) and MUST NOT overwrite existing binaries

### Requirement MEDIA-STORAGE-011: Media section SHALL support bulk and selective download workflows
Cabinet SHALL allow users to download one, many, or all media assets from current scope/filter.

#### Scenario: Download selected assets
- **GIVEN** user selects subset of media cards
- **WHEN** user triggers download action
- **THEN** runtime MUST stream downloadable package containing selected files with friendly filenames

#### Scenario: Download all filtered assets
- **GIVEN** user applies filter (for example `unlinked`)
- **WHEN** user triggers `download all`
- **THEN** runtime MUST package all matching assets for download and include summary metadata (count, skipped, failures)

### Requirement UI-SCREEN-MEDIA-004: Mobile UI SHALL provide one-tap camera capture from any app screen
Cabinet SHALL expose a global one-tap mobile camera action that is reachable from any authenticated screen and routes capture into media library.

#### Scenario: One-tap capture from non-media screen
- **GIVEN** user is on any authenticated route on mobile viewport
- **WHEN** user taps global camera action button
- **THEN** camera capture flow MUST open without requiring navigation to media section

#### Scenario: Save captured photo to media library
- **GIVEN** camera capture succeeds
- **WHEN** user confirms capture
- **THEN** runtime MUST ingest photo into global media registry with dedupe/hash pipeline and show success feedback with quick link to media item

### Requirement UI-SCREEN-MEDIA-005: Mobile media flow SHALL support rapid multi-photo intake
Cabinet SHALL support high-speed intake on mobile via batch picker and capture-next loop.

#### Scenario: Batch select multiple photos
- **GIVEN** user opens media upload on mobile
- **WHEN** user selects multiple images from device gallery/files
- **THEN** runtime MUST enqueue all selected assets, process dedupe/hash per asset, and show per-file progress/results

#### Scenario: Capture-next loop for rapid camera intake
- **GIVEN** user completes a mobile camera capture
- **WHEN** user chooses `capture next`
- **THEN** flow MUST return to camera without leaving context and append each accepted photo to same upload queue

### Requirement UI-SCREEN-MEDIA-006: Media workspace shell SHALL be discoverable from authenticated navigation
Cabinet SHALL expose a dedicated authenticated `/media` workspace from primary navigation with a page title, shared table asset surface, unlinked filter, operational metadata, and visible action controls before backend ingestion and assignment slices are complete.

#### Scenario: Open Media workspace shell from navigation
- **GIVEN** user is signed in on an authenticated Cabinet route
- **WHEN** user opens the primary Media navigation item
- **THEN** Cabinet MUST navigate to `/media`, set document title `Cabinet - Media`, render a shared table media workspace, show all/unlinked filter controls, show deterministic asset metadata, and expose open/analyze/assign/archive plus upload/download controls without mutating media links.

### Requirement UI-SCREEN-MEDIA-007: Media workspace API SHALL expose profile-scoped assets and preview-only assignment/download contracts
Cabinet SHALL expose a deterministic backend contract for the Media workspace that scopes media assets to the active profile, classifies linkage state, supports the unlinked filter, and previews assignment/download actions without mutating linkage records until confirmed implementation slices exist.

#### Scenario: List profile-scoped Media workspace assets
- **GIVEN** the active profile has inventory-linked photos and unlinked media evidence
- **WHEN** UI requests `/api/media/assets`
- **THEN** runtime MUST return only active-profile assets with stable `linkage_state`, source, upload metadata, thumbnail/download hints, and summary counts for total, unlinked, and linked states.

#### Scenario: Filter unlinked Media workspace assets
- **GIVEN** the active profile has linked and unlinked media assets
- **WHEN** UI requests `/api/media/assets?filter=unlinked`
- **THEN** runtime MUST return only unlinked assets while retaining summary counts for the full active-profile scope.

#### Scenario: Preview assignment and download actions
- **GIVEN** a media asset exists in the active profile
- **WHEN** UI requests assignment or download preview contracts
- **THEN** runtime MUST return deterministic projected linkage, confirmation, blocker/audit, and filename details without creating links, duplicating binaries, or streaming files in the preview request.

### Requirement UI-SCREEN-MEDIA-008: Media workspace UI SHALL render API-backed asset states and download previews
Cabinet SHALL wire the authenticated Media workspace to the profile-scoped media API instead of static UI fixtures.

#### Scenario: Render API-backed media list
- **GIVEN** the active profile has linked and unlinked media assets
- **WHEN** the user opens `/media`
- **THEN** the UI MUST request `/api/media/assets`, render returned asset cards, summary counts, linkage badges, filenames, thumbnails where available, and keep the unlinked tab backed by `/api/media/assets?filter=unlinked`.

#### Scenario: Handle empty and unavailable media API states
- **GIVEN** the media API returns no assets or a recoverable error
- **WHEN** the user opens or retries the Media workspace
- **THEN** the UI MUST show deterministic empty, error, and retry states without falling back to stale fixture cards.

#### Scenario: Preview selected downloads from the UI
- **GIVEN** one or more media cards are selected
- **WHEN** the user previews selected downloads
- **THEN** the UI MUST call `/api/media/downloads/preview` with the selected asset IDs and current filter, then display returned human-readable filenames without streaming files or mutating media state.

### Requirement UI-SCREEN-MEDIA-009: Media workspace download API SHALL stream selected profile-scoped assets
Cabinet SHALL provide an authenticated Media workspace download endpoint that returns selected asset bytes from the active profile with human-readable filenames.

#### Scenario: Download selected media payload
- **GIVEN** the active profile has inventory-linked media and unlinked evidence assets with stored bytes
- **WHEN** the user requests `/api/media/downloads` with selected asset IDs and the current filter
- **THEN** the API MUST reject assets outside the active profile or current filter, return a single selected asset with its human-readable filename and content type, and return multiple selected assets as a zip archive preserving each human-readable filename.

### Requirement UI-SCREEN-MEDIA-010: Media workspace assignment API SHALL persist scoped media links
Cabinet SHALL provide a confirmed assignment endpoint for linking Media workspace assets to inventory or wishlist targets without duplicating source binaries.

#### Scenario: Assign unlinked media to wishlist
- **GIVEN** the active profile has an unlinked media asset and a wishlist entry
- **WHEN** the user confirms assignment through `/api/media/assignments`
- **THEN** the API MUST persist a profile-scoped media link, preserve the original media asset provenance, update the asset linkage state to `linked_wishlist`, and update unlinked/wishlist summary counts on the next `/api/media/assets` response.

#### Scenario: Reject out-of-scope media assignment
- **GIVEN** a media asset or assignment target belongs to another profile
- **WHEN** the user confirms assignment through `/api/media/assignments`
- **THEN** the API MUST reject the request and MUST NOT create a media link.

### Requirement UI-SCREEN-MEDIA-011: Media workspace UI SHALL confirm assignments through the API
Cabinet SHALL route Media workspace assignment actions through a preview-first confirmation UI backed by the confirmed assignment API.

#### Scenario: Confirm media assignment from card action
- **GIVEN** the active profile has an unlinked media asset and a valid assignment target
- **WHEN** the user opens the card assignment action, previews the target, and confirms assignment
- **THEN** the UI MUST call `/api/media/assignments/preview`, display the projected linkage and audit/provenance summary, call `/api/media/assignments` only after confirmation, refresh `/api/media/assets`, and render the updated linkage state without stale unlinked controls.

#### Scenario: Assignment preview or apply fails
- **GIVEN** the assignment target is invalid or the assignment endpoint rejects the request
- **WHEN** the user previews or confirms assignment
- **THEN** the UI MUST show a deterministic error state and MUST NOT update card linkage state until the refreshed API state confirms the assignment.

### Requirement UI-SCREEN-MEDIA-012: Media cards SHALL use compact responsive density
Cabinet SHALL render Media workspace asset cards in a compact grid that substantially reduces the desktop card footprint while preserving readable metadata, visible thumbnails, usable quick actions, and responsive mobile behavior.

#### Scenario: Render compact desktop Media cards
- **GIVEN** the Media workspace has multiple returned assets
- **WHEN** the user opens `/media` on the primary desktop review viewport
- **THEN** the card grid MUST render compact cards at approximately one third of the previous visual footprint, keep thumbnails and metadata readable, and keep open/analyze/assign/archive controls usable without clipping.

#### Scenario: Preserve responsive Media card behavior
- **GIVEN** the Media workspace has multiple returned assets
- **WHEN** the user opens `/media` at tablet or mobile widths
- **THEN** the card grid MUST reflow without horizontal overflow, overlapping content, or clipped metadata/action controls.

### Requirement UI-SCREEN-MEDIA-013: Media workspace SHALL provide persisted Cards and Rows view modes
Cabinet SHALL expose explicit Cards and Rows view controls on the authenticated Media workspace, default to Rows, preserve the existing card behavior as a secondary view, and persist the selected view mode using Cabinet's view-state convention.

#### Scenario: Switch from Cards to Rows view
- **GIVEN** the Media workspace has returned profile-scoped media assets
- **WHEN** the user opens `/media`
- **THEN** the UI MUST default to Rows mode and render the media shared table.
- **WHEN** the user switches to Cards mode
- **THEN** the UI MUST render the preserved compact card grid.

#### Scenario: Persist Media view mode
- **GIVEN** the user has selected Cards or Rows mode on `/media`
- **WHEN** the workspace reloads
- **THEN** Cabinet MUST restore the selected Media view mode from `cabinet.viewMode.media` without changing the active media filter or API query.

### Requirement UI-SCREEN-MEDIA-015: Media workspace table SHALL use Cabinet shared table affordances
Cabinet SHALL render the Media page primary content with the shared table surface used by other Cabinet management pages, including toolbar search/filtering, sortable scan-friendly columns, stable row selection, row actions, and a header-level Add new asset action. The Media workspace SHALL reclaim working space by omitting the former top summary-card row above the table workflow.

#### Scenario: Manage media from shared table surface
- **GIVEN** the Media workspace has returned profile-scoped media assets
- **WHEN** the user opens `/media`
- **THEN** the primary content MUST render a `data-table-surface` table with thumbnail/title identity, analysis status, linkage state, upload timestamp, source/context, filename, row selection, and open/analyze/assign/archive actions.
- **AND** the page header MUST expose an accessible `Add new asset` action.
- **AND** the workspace MUST NOT render the former top summary cards for Assets, Unlinked, or Ready for review above the table workflow.
- **AND** the table toolbar MUST provide search/filtering over media identity, status, linkage, source, and filename fields without switching away from the table.

### Requirement UI-SCREEN-MEDIA-014: Media workspace SHALL support page-wide image drop and add-media metadata dialog
Cabinet SHALL let authenticated users add unlinked media assets from the Media workspace by dragging supported image files anywhere over the page or by opening an explicit add-media dialog from a `+` action.

#### Scenario: Start add-media flow from page-wide drop
- **GIVEN** the user is on `/media`
- **WHEN** the user drags and drops a supported image file anywhere over the Media workspace
- **THEN** Cabinet MUST open the Add media dialog, preserve the dropped image, show drag/drop feedback, and allow metadata entry before save.

#### Scenario: Save dropped media with metadata
- **GIVEN** the Add media dialog contains a supported image and metadata fields
- **WHEN** the user saves the media
- **THEN** the UI MUST submit the same multipart `/api/media/assets` create path used by the explicit `+` dialog, refresh `/api/media/assets`, and show the newly created unlinked asset without losing metadata on success.

#### Scenario: Reject unsupported media files without losing metadata
- **GIVEN** the user has entered Add media metadata
- **WHEN** the user drops or chooses an unsupported file type
- **THEN** Cabinet MUST block save, show deterministic unsupported-file feedback, and preserve the entered metadata for correction.

#### Scenario: Preserve add-media metadata on save failure
- **GIVEN** the Add media dialog contains a supported image and metadata
- **WHEN** `/api/media/assets` rejects the save
- **THEN** Cabinet MUST keep the dialog open, preserve the selected file and metadata values, and show an actionable error.

### Requirement UI-SCREEN-MEDIA-016: Media assets SHALL open a double-click metadata edit modal
Cabinet SHALL let authenticated users double-click any Media workspace asset row or card to edit asset metadata and review thumbnail variants in one place.

#### Scenario: Edit media asset metadata from double-click modal
- **GIVEN** the Media workspace has returned profile-scoped media assets
- **WHEN** the user double-clicks a media row or card
- **THEN** Cabinet MUST open an edit modal for that asset with a thumbnail preview, visible thumbnail variation options, and editable metadata fields for title, filename, source, download filename, and notes.
- **WHEN** the user saves the modal
- **THEN** the UI MUST submit the edited metadata through the Media asset update API, refresh `/api/media/assets`, and show the updated asset metadata without requiring a page reload.
