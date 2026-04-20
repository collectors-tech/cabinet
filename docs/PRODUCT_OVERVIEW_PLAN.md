# Cabinet Product Overview Plan

This document turns [PRODUCT-OVERVIEW.md](./PRODUCT-OVERVIEW.md) into a practical completion plan for the main collector workflows:

- Inventory
- Wishlist
- Collections
- Photos, media, and item evidence
- Settings, storage, and maintenance

The intent is to finish these areas as one coherent collector workspace, not as isolated screens.

## Planning principles

These feature areas should be completed in a sequence that strengthens trust in the app:

1. Owned-item source of truth before cross-feature convenience
2. Clear data boundaries before automation
3. Reliable persistence before visual polish
4. Restart-safe behavior before advanced workflows
5. Storage and maintenance confidence before large media or import-heavy use

Across every phase, a feature should only be considered complete when it:

- is discoverable in the UI
- communicates its purpose clearly
- performs the intended action successfully
- persists the result correctly
- reloads with the same correct state
- behaves correctly with the active profile/session context
- has validation coverage appropriate to its risk

## Recommended delivery order

The recommended order is:

1. Inventory foundation and persistence
2. Wishlist planning and owned-item conversion
3. Collections as shared organization model
4. Photos, media, and item evidence
5. Settings, storage, and maintenance hardening

This order keeps the owned-item model stable first, then adds planning, grouping, evidence, and recovery around it.

## Delivery waves

### Wave 1: Inventory as source of truth

Inventory is the core operational surface and should become the most trustworthy part of the app first.

#### Goals

- Create and edit owned items reliably
- Keep list, detail, and folder/tree views in sync
- Make search, filter, sort, and saved-view behavior dependable
- Preserve item identifiers, instances, and detail state through the API-backed save path

#### Main workstreams

- Inventory create/edit/save reliability
- Folder tree, list hierarchy, and detail synchronization
- Search, filter, sort, and saved-view consistency
- Item instance, barcode, and identifier support
- State restoration after reload and restart

#### Completion criteria

- New owned items save correctly and reappear after reload/restart
- Existing items update without stale state between list and detail surfaces
- Folder tree and list organization are readable and persist correctly
- Item details remain aligned with the selected item
- Search/filter/sort flows produce stable and understandable results
- API-backed save flows are the only trusted persistence path

#### Suggested milestone slices

- M1.1: Item create/edit persistence and reload coverage
- M1.2: Folder tree and hierarchy stability
- M1.3: Search/filter/sort and saved views
- M1.4: Item identifiers, barcodes, and instances

### Wave 2: Wishlist as planning surface

Wishlist should become a real planning tool, not just a holding list.

#### Goals

- Create and manage wanted items cleanly
- Support status, priority, and planning state
- Keep wanted vs owned boundaries clear
- Allow explicit conversion into inventory when items are acquired

#### Main workstreams

- Wishlist item creation and planning metadata
- Status and priority model clarity
- Conversion flow from wanted to owned
- Boundary rules between wishlist state and inventory state
- Wishlist summaries and actionable views

#### Completion criteria

- Wanted items persist correctly and survive reload/restart
- Priority and status fields drive useful planning behavior
- Conversion to owned inventory is explicit, traceable, and reliable
- Users can understand what is wanted, what is owned, and what changed state
- Wishlist views help users decide what matters next

#### Suggested milestone slices

- M2.1: Wishlist create/edit/persist
- M2.2: Priority, status, and planning views
- M2.3: Acquisition and conversion to inventory

### Wave 3: Collections as shared organization layer

Collections should sit above stable inventory and wishlist behavior as the grouping model across the app.

#### Goals

- Create, edit, and manage collection buckets
- Assign and move items predictably
- Support collection management from clear surfaces
- Use collections consistently across the relevant parts of the app

#### Main workstreams

- Collection CRUD and management UI
- Collection assignment and move flows
- Cross-surface collection consistency
- Persistence and reload behavior for collection state
- Clear user-facing collection semantics

#### Completion criteria

- Collections can be created, renamed, and managed without hidden control paths
- Items can be assigned and moved between collections reliably
- Collection changes persist and reload correctly
- Inventory and wishlist use collection rules consistently where intended
- The collections screen feels first-class rather than supplemental

#### Suggested milestone slices

- M3.1: Collection management surface completion
- M3.2: Assignment and move flows from item workflows
- M3.3: Cross-surface consistency and persistence

### Wave 4: Photos, media, and item evidence

Once the item model is solid, evidence attachment should become reliable and understandable.

#### Goals

- Attach photos and evidence to the correct item and profile
- Support primary-photo selection
- Rebuild or repair photo state when needed
- Make configured media/storage paths understandable

#### Main workstreams

- Media upload and attachment reliability
- Primary-photo behavior
- Media-path and storage awareness
- Rebuild/repair flows for evidence state
- Item evidence review experience

#### Completion criteria

- Uploading media is dependable and tied to the correct item/profile
- Primary photo state is visible and persists
- Media survives reload/restart and remains attached to the right item
- Users can understand where media is stored and what happens when storage changes
- Repair or rebuild flows recover broken media state without guesswork

#### Suggested milestone slices

- M4.1: Item photo attachment and primary photo behavior
- M4.2: Media path visibility and storage-aware state
- M4.3: Media repair/rebuild flows

### Wave 5: Settings, storage, and maintenance

This wave turns the app from merely usable into trustworthy over time.

#### Goals

- Make settings understandable and properly scoped
- Expose storage configuration clearly
- Provide safe maintenance and recovery workflows
- Build confidence in backup, restore, import, export, and repair paths

#### Main workstreams

- Profile settings vs app settings clarity
- Storage path configuration and visibility
- Backup and restore flows
- Import/export reliability and guardrails
- Diagnostics, repair, and maintenance operations

#### Completion criteria

- Users can understand which settings affect the app, the profile, or storage
- Storage paths are visible, correct, and easy to validate
- Backup/restore/import/export flows are explicit and trustworthy
- Maintenance operations explain risk and outcome clearly
- Recovery workflows reduce fear instead of introducing more uncertainty

#### Suggested milestone slices

- M5.1: Settings scoping and information architecture
- M5.2: Storage configuration and validation
- M5.3: Backup/restore/import/export
- M5.4: Diagnostics and repair

## Cross-cutting requirements

These requirements apply to every wave and should be tracked in parallel:

### 1. Persistence and restart safety

Every major action should prove:

- save succeeds
- reload shows the same result
- restart preserves the same result

### 2. Profile/session correctness

Every workflow should prove:

- data stays inside the active profile
- switching profile changes the whole app context correctly
- screen state does not leak across profiles

### 3. Data-boundary clarity

The app should make these boundaries explicit:

- owned vs wanted
- collection vs profile
- local state vs provider state
- UI draft state vs persisted state

### 4. UI trust and interaction quality

Controls should match intent:

- buttons look clickable
- dialogs layer correctly
- forms validate clearly
- empty states explain what to do next
- failure states explain what went wrong

### 5. Validation strategy

Each main feature area should have:

- one stable happy-path smoke
- one meaningful persistence/reload test
- one active-profile isolation check
- one failure-path test for the highest-risk interaction

## Dependency map

The main dependency chain should be treated like this:

```text
Inventory foundation
    ->
Wishlist conversion rules
    ->
Collections as shared organization
    ->
Photos/media attached to stable item state
    ->
Settings/storage/maintenance to preserve and recover all of the above
```

This means:

- do not finalize media flows before item identity and persistence are trustworthy
- do not finalize collection movement before inventory and wishlist boundaries are clear
- do not present backup/restore as trustworthy until the underlying save paths are stable

## Suggested issue structure

To keep delivery focused, this plan should be executed through five epics:

- Epic A: Inventory completion and persistence hardening
- Epic B: Wishlist planning and owned-item conversion
- Epic C: Collections management and shared organization
- Epic D: Photos, media, and item evidence reliability
- Epic E: Settings, storage, and maintenance trust

Each epic should contain:

- workflow completion issues
- persistence/restart verification issues
- profile-context verification issues
- targeted UI trust/polish issues only after behavior is stable

## Definition of overall success

These feature areas should feel complete when a collector can:

- add and maintain owned items confidently
- plan wanted items without confusing them with owned items
- group items into collections in a way that persists and stays understandable
- attach and review media/evidence without losing trust in storage state
- recover, back up, restore, and maintain the workspace safely

At that point, Cabinet will feel less like a collection of screens and more like a dependable collector workspace.
