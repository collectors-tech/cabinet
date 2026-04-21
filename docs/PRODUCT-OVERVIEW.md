# Cabinet Product Overview

Cabinet is a desktop-first collector intelligence app.

Its job is to help a collector keep a trustworthy working system for what they own, what they want, how items are grouped, what external providers know, and what actions they should take next.

## Product intent

Cabinet should feel like one coherent collector workspace, not a pile of disconnected admin screens.

A working Cabinet experience should let a user:
- get into the app quickly
- choose or create the right data/profile context
- see useful information immediately
- add, edit, find, and organise items without friction
- manage wanted items separately from owned items
- group items into meaningful collections
- attach photos, barcodes, and other identifying data
- connect external providers and validate those connections
- use chat/assistant flows with real app context
- trust that changes persist, reload correctly, and survive app restarts

## Main feature map

### 1. Onboarding, sign-in, and profile context

This area should:
- let the user sign in or unlock the app cleanly
- let the user create/select the active database profile
- support starter/sample-data flows for first-run exploration
- make the active profile/database context obvious

When this is working well:
- first-run setup is clear
- switching profile changes the whole app context correctly
- auth/session state is stable and understandable
- sample data creates a usable demo state without confusing the real data model

### 2. Dashboard / home surface

This area should:
- give a quick summary of the current workspace
- surface the most useful next information for the active profile
- act as a reliable landing page after sign-in

When this is working well:
- the dashboard loads with correct copy and layout
- key counts, summaries, and entry points are understandable
- it helps the user move into inventory, wishlist, collections, chats, or integrations without hunting around

### 3. Inventory

Inventory is the owned-items management surface.

This area should:
- create new owned items
- edit existing items reliably
- browse items in rows/cards/folder views
- support search, filter, sort, and saved views
- show item details and related metadata
- support item instances, barcodes, and photos
- preserve changes through the API-backed save path

When this is working well:
- creating and editing items always persists correctly
- folder tree and list hierarchy are readable
- item details stay in sync with list state
- photos and identifiers attach to the correct item
- the user can trust inventory as the source of truth for owned items

### 4. Wishlist

Wishlist is the planning surface for items not yet owned.

This area should:
- create and manage wanted items
- track status, priority, and planning state
- support practical decision-making, not just a flat list
- allow conversion of acquired items into owned inventory state
- support movement between wishlist and owned states when item status changes

When this is working well:
- the wishlist acts like a real planning tool
- users can see what matters next
- conversion to owned inventory is explicit and reliable
- the boundary between wanted and owned remains clear

### 5. Collections

Collections are grouping buckets across the app.

This area should:
- create, edit, and manage collection buckets
- support assigning items into collections
- support moving items between collections
- work across both inventory and wishlist contexts where appropriate
- use a clear management surface rather than hidden or one-off controls

When this is working well:
- users can organise their items around real collection structures
- collection assignment is easy from item flows
- collection moves are predictable and persist correctly
- the collections screen feels like a first-class management area

### 6. Chats / assistant

Chats provide contextual assistant/copilot support inside Cabinet.

This area should:
- open chat threads reliably
- preserve thread history
- let the assistant operate with app context where appropriate
- support workspace-aware assistant panels and defaults

When this is working well:
- chat threads are readable and stable
- the assistant surface respects the active workspace/profile context
- saved/default assistant settings load correctly
- the user can move between app work and chat work without context confusion

### 7. Integrations

Integrations connect Cabinet to outside providers.

This area should:
- connect and disconnect providers
- store and validate credentials/settings safely
- run provider sync/status checks
- surface health and failure states clearly

When this is working well:
- provider setup is understandable
- token/config validation gives clear pass/fail feedback
- sync actions produce visible outcomes
- integration failures do not silently poison the rest of the app

### 8. Photos, media, and item evidence

This area should:
- attach photos to items
- mark a primary photo
- rebuild photo state when needed
- respect configured storage/media paths

When this is working well:
- media uploads are reliable
- item imagery is easy to review
- storage configuration is understandable
- media state stays attached to the correct item and profile context

### 9. Search, matching, and scanner flows

This area should:
- search local items well
- support barcode lookups and matching flows
- manage scanner query sets and candidate results
- help identify or reconcile item data from external/provider signals

When this is working well:
- users can find things quickly
- scanner/matching results are actionable, not noisy
- candidate/failure states are inspectable
- item identification improves instead of creating ambiguity

### 10. Settings, storage, and maintenance

This area should:
- manage app preferences and profile settings
- manage storage/media path configuration
- expose diagnostics, repair, backup, restore, export, and import flows where enabled
- make maintenance tasks explicit and safe

When this is working well:
- settings are understandable and scoped correctly
- storage paths are visible and correct
- backup/restore/import/export flows are trustworthy
- maintenance operations help recover the app instead of creating fear

## Core product expectations

Across all feature areas, Cabinet should be doing these things well:

- **Persistence works**: create, edit, move, convert, assign, and settings changes actually save
- **State is coherent**: profile, session, workspace, and screen state do not drift or contradict each other
- **Navigation is clear**: the user can move between main areas without getting lost
- **UI controls match intent**: buttons look clickable, dialogs layer correctly, forms behave correctly
- **Data boundaries are clear**: owned vs wishlist, collection vs profile, local state vs provider state
- **Reload/restart is safe**: app restarts do not destroy trust in saved work
- **Errors are visible and actionable**: failures should explain what went wrong and what the user can do next

## Core vs secondary features

### Core app spine
These should feel solid first:
- onboarding / auth / profile selection
- dashboard / landing experience
- inventory
- wishlist
- collections
- chats / assistant
- integrations
- settings/storage

### Supporting platform capabilities
These support the spine and should make it stronger, not distract from it:
- photos/media management
- scanner/query-set flows
- local search and matching
- import/export
- backup/restore
- diagnostics and repair

## Product quality bar

A feature in Cabinet should be considered properly working when it:
- is discoverable in the UI
- communicates its purpose clearly
- performs the intended action successfully
- persists the result correctly
- reloads with the same correct state
- behaves correctly with the active profile/session context
- has validation coverage appropriate to its risk

## Short version

Cabinet should be a dependable collector workspace where users can:
- manage owned items
- plan wanted items
- organise everything into collections
- attach supporting media and identifiers
- connect external providers
- use assistant/chat workflows with context
- trust that the app saves, restores, and explains state correctly
