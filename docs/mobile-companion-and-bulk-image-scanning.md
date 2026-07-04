# Cabinet Mobile Companion and Bulk Image Scanning

Status: Planning document  
Scope: Cabinet desktop-first companion mobile app, offline capture cache, upload queue, and bulk image scan workflow  
Target implementation repository: `collectors-tech/cabinet`

## Purpose

Cabinet should remain a desktop-first, local-first collector workspace. The mobile app should be a companion capture surface, not a second full Cabinet client.

The mobile companion exists to help collectors capture items quickly while they are away from the desktop app, cache the captured content safely, and upload that content back into Cabinet for review, matching, reconciliation, and final import.

The most important mobile workflows are:

- scan barcodes, QR codes, and Cabinet handoff codes
- capture item photos and evidence photos
- capture quick inventory and wishlist notes
- bulk scan images for candidate items
- cache everything offline
- upload captured batches to Cabinet desktop
- let desktop Cabinet review, match, merge, and finalise data

## Product position

### Mobile companion

The mobile app should do fast field capture:

- quick inventory item capture
- quick wishlist item capture
- item photo capture
- barcode and QR scanning
- bulk image scan capture
- receipt or order screenshot capture
- local offline queue management
- upload to a paired Cabinet desktop profile

### Desktop Cabinet

Desktop Cabinet remains the main authority for:

- full inventory management
- wishlist management
- collections
- matching and reconciliation
- duplicate resolution
- upload inbox review
- bulk edit and import
- media storage configuration
- backup and restore
- public repository sync
- advanced verification workflows

The phone should help the user collect data. It should not silently mutate the main inventory without a review path unless the user explicitly enables a trusted auto-import mode later.

## Design principles

1. **Desktop-first authority**  
   The desktop Cabinet profile remains the source of truth for inventory, wishlist, collections, storage, backup, import, export, and repository publication.

2. **Mobile-first capture speed**  
   The phone should make capture extremely fast: scan, snap, note, queue, and move on.

3. **Offline by default**  
   Every capture must survive poor reception, convention halls, store basements, and home network interruptions.

4. **Review before truth**  
   Bulk scans create candidates, not final inventory truth.

5. **Idempotent upload**  
   Retrying a failed upload must not create duplicate inventory or duplicate media records.

6. **Private by default**  
   Mobile capture data is private local data until the user uploads it to their own Cabinet desktop workspace.

7. **No destructive mobile permissions in MVP**  
   A paired mobile device can upload capture batches, but cannot delete or overwrite desktop inventory.

## Non-goals for MVP

The first mobile release should not include:

- full desktop parity
- public Git or Radicle publication from the phone
- marketplace selling flows
- escrow or payment custody
- destructive inventory changes
- full inventory download to phone
- visual-only auto-import without review
- P2P trade desk implementation
- public seeding of private media

Those can be considered later once the mobile capture and desktop upload inbox are stable.

## High-level architecture

```text
Mobile Companion
  ├── Local capture database
  ├── Local media cache
  ├── Camera / barcode / QR / OCR capture
  ├── Bulk image scan candidate builder
  ├── Upload queue
  └── Pairing credentials

Cabinet Desktop Runtime
  ├── Mobile pairing endpoint
  ├── Mobile upload API
  ├── Mobile Upload Inbox
  ├── Candidate matching and duplicate detection
  ├── Inventory / wishlist / photo import actions
  └── Existing backup, storage, import, and media paths
```

## Pairing model

Desktop should expose a temporary **Receive from Mobile** mode.

### Pairing flow

1. User opens Cabinet desktop.
2. User opens **Settings → Mobile devices** or **Mobile Upload Inbox**.
3. User clicks **Pair mobile device**.
4. Desktop shows a short-lived QR code.
5. Mobile scans the QR code.
6. Mobile generates or reuses a device key.
7. Desktop records the device as an approved upload source.
8. Mobile can now upload capture batches to that specific Cabinet profile.

### Mobile permission scope

MVP mobile permissions should be narrow:

- upload capture batches
- upload media assets for those batches
- check upload status
- retry failed uploads
- list its own previous upload batches

MVP mobile permissions should not allow:

- deleting desktop inventory
- editing existing desktop items directly
- publishing public identity, feedback, or receipts
- exporting the full Cabinet database
- changing desktop storage settings

## Capture workflows

### 1. Quick item capture

The user can quickly capture an owned item.

Fields:

- title or search text
- category
- barcode or product code
- quantity
- condition
- collection hint
- purchase/source note
- optional price paid
- notes
- photos

Output:

- mobile capture item
- linked media assets
- upload queue entry

Desktop import options:

- create new item
- add as item instance/copy
- link to existing item
- attach photos to existing item
- ignore
- needs review

### 2. Quick wishlist capture

The user can quickly capture something they want but do not own yet.

Fields:

- title or search text
- category
- priority
- source/link/note
- estimated price
- photos or screenshots
- target collection hint

Desktop import options:

- create wishlist item
- link to existing wishlist item
- convert to inventory if already acquired
- ignore
- needs review

### 3. Photo batch capture

The user can capture many photos first and decide what they are later.

Use cases:

- shelf photos
- binder pages
- box contents
- trade night finds
- store display photos
- purchase proof
- condition evidence

Every photo should have:

- original file
- generated thumbnail
- optional compressed WebP rendition
- content hash
- local asset ID
- capture timestamp
- upload state
- optional user note

### 4. Receipt or order capture

The user can photograph or import:

- paper receipts
- invoices
- eBay order screenshots
- marketplace order screenshots
- shipping or tracking screenshots

Cabinet should extract what it can, but the desktop inbox should review and reconcile the result.

Possible extracted fields:

- source/store
- order number
- date
- item names
- quantities
- line prices
- total
- tracking/reference numbers
- linked photos

## Bulk image scanning

Bulk image scanning should be a first-class mobile companion feature, but it must feed a review queue.

The core rule is:

> Bulk image scan creates candidates, not truth.

### Supported scan inputs

The user should be able to scan from:

- newly captured camera photos
- camera roll images
- binder pages
- shelf photos
- box/tray photos
- loose item piles
- receipts
- order screenshots
- packaging photos
- barcode label photos

### Scan modes

MVP should support scan modes so Cabinet can tune the candidate builder:

| Scan mode | Purpose |
| --- | --- |
| Barcode batch | Extract one or more visible barcodes from still images. |
| QR batch | Extract Cabinet QR, handoff, or item codes. |
| OCR text scan | Extract names, model numbers, set numbers, receipt lines, and packaging text. |
| Photo batch | Store images for later manual review. |
| Manual crop | User marks regions manually when automation is weak. |
| Binder page | Detect card-like regions and crop them into candidates. |

Binder page and object detection can mature later. Barcode, QR, OCR, photo batch, and manual crop should come first.

### Candidate pipeline

```text
Image batch
  → asset hash and thumbnail generation
  → barcode / QR detection
  → OCR text extraction
  → optional region detection
  → candidate creation
  → local candidate review
  → upload to desktop
  → desktop matching and import review
```

### Candidate review

Mobile can show a lightweight review grid, but desktop remains the main review surface.

Example candidate row:

```text
[Crop image]  Porsche 911 GT3 RS   OCR match       82%   Add / Edit / Ignore
[Crop image]  088796123456         Barcode match   High  Add / Link / Ignore
[Crop image]  Unknown card          Needs review    Low   Edit / Ignore
```

### Matching confidence

Cabinet should treat signals differently.

| Signal | Confidence |
| --- | --- |
| Exact Cabinet QR/specimen code | Highest |
| Exact barcode match | Highest |
| OCR plus exact catalogue match | High |
| OCR plus category/source match | Medium |
| Visual similarity only | Low |
| AI guess only | Review required |

Visual-only matches must not auto-import in the MVP.

### Cards and binder pages

Binder page scanning should eventually support:

- 3x3 pages
- 4x3 pages
- sleeved cards
- glare warnings
- duplicate detection
- same card with multiple copies
- title OCR
- set number OCR
- edition/variant hints
- candidate crop review

For MVP, binder scanning can be implemented as photo batch plus manual crop, then improved with rectangle detection later.

### Slot car / diecast shelf scans

Shelf or boxed-item scanning is harder than cards because items are 3D, packaging varies, and model names can be partially hidden.

MVP should start with:

- barcode detection
- OCR model text
- manual crop regions
- one image with multiple candidates
- rough visual category only

Do not attempt perfect visual identification from images in the first release.

### Receipts and order screenshots

Receipt/order capture is high value because it helps reconcile purchases.

Expected desktop review outcomes:

- create purchase/order record
- link order lines to existing items
- add new inventory items
- add wishlist items
- attach receipt image as evidence
- ignore non-item lines
- flag unmatched lines

## Local storage model

### Mobile local database

Suggested tables:

```text
paired_cabinets
capture_batches
capture_items
capture_assets
scan_batches
scan_candidates
sync_queue
sync_attempts
local_catalogue_cache
```

### Mobile local files

Suggested app-private file structure:

```text
mobile-cache/
  originals/
    yyyy/mm/<asset_id>.<ext>
  thumbnails/
    yyyy/mm/<asset_id>.webp
  crops/
    yyyy/mm/<candidate_id>.webp
  manifests/
    <batch_id>.json
```

Every file should be content-hashed. Hashes should be checked before and after upload.

## Upload queue states

| State | Meaning |
| --- | --- |
| Draft | Capture started but not ready. |
| Ready | User marked it ready for upload. |
| PendingUpload | Stored locally and waiting for a paired Cabinet desktop. |
| Uploading | Transfer in progress. |
| Uploaded | Desktop received the raw batch/assets. |
| Imported | Desktop accepted the content into inventory/wishlist/media. |
| NeedsReview | Desktop needs user decision. |
| Conflict | Desktop found duplicate or conflicting state. |
| Failed | Transfer failed and needs retry. |

Upload retries must be safe. The desktop should use batch IDs, item IDs, asset hashes, and idempotency keys to prevent duplicates.

## Desktop Mobile Upload Inbox

Desktop should have a **Mobile Upload Inbox**.

This is where mobile batches become real Cabinet data.

Inbox sections:

- capture batches
- bulk image scan batches
- barcode matches
- OCR candidates
- possible duplicates
- missing required fields
- failed asset uploads
- ready to import
- ignored/imported history

Desktop actions:

- import as new inventory item
- import as wishlist item
- link to existing item
- add as item instance/copy
- add photos to existing item
- attach receipt/order evidence
- batch apply collection/source/condition
- ignore selected candidates
- mark batch resolved

## Data packet examples

### Mobile capture batch

```json
{
  "schema": "cabinet.mobile-capture-batch.v1",
  "batchId": "mobile_batch_001",
  "deviceId": "cabinet:device:iphone-max",
  "profileId": "profile_default",
  "createdAt": "2026-07-04T10:00:00+10:00",
  "items": [
    {
      "clientItemId": "mobile_item_001",
      "target": "inventory",
      "title": "Hot Wheels Porsche 911",
      "category": "diecast",
      "barcode": "1234567890",
      "quantity": 1,
      "condition": "Good",
      "collectionHint": "Loose cars",
      "notes": "Found at swap meet",
      "assetHashes": ["sha256:abc..."]
    }
  ],
  "assets": [
    {
      "assetId": "asset_001",
      "hash": "sha256:abc...",
      "mimeType": "image/jpeg",
      "filename": "IMG_0012.jpeg",
      "role": "item-photo"
    }
  ],
  "idempotencyKey": "mobile_batch_001",
  "signature": "device-signature"
}
```

### Bulk image scan batch

```json
{
  "schema": "cabinet.mobile-bulk-image-scan.v1",
  "scanBatchId": "scan_batch_001",
  "deviceId": "cabinet:device:iphone-max",
  "scanMode": "binder-page",
  "createdAt": "2026-07-04T10:00:00+10:00",
  "images": [
    {
      "assetId": "asset_001",
      "hash": "sha256:abc...",
      "mimeType": "image/jpeg",
      "width": 4032,
      "height": 3024,
      "thumbnailHash": "sha256:def..."
    }
  ],
  "candidates": [
    {
      "candidateId": "candidate_001",
      "assetId": "asset_001",
      "region": {
        "x": 120,
        "y": 380,
        "width": 640,
        "height": 880
      },
      "signals": {
        "barcode": null,
        "ocrText": "Porsche 911 GT3 RS",
        "visualLabel": "diecast car"
      },
      "suggestedMatch": {
        "catalogueItemId": "catalogue_hotwheels_001",
        "confidence": 0.82
      },
      "userDecision": "pending"
    }
  ],
  "syncStatus": "PendingUpload"
}
```

## Desktop API additions

Suggested MVP API surface:

```text
POST   /api/mobile/pairing/begin
POST   /api/mobile/pairing/finish
GET    /api/mobile/devices
DELETE /api/mobile/devices/{deviceID}

POST   /api/mobile/batches
POST   /api/mobile/batches/{batchID}/assets
POST   /api/mobile/batches/{batchID}/complete
GET    /api/mobile/batches/{batchID}/status

GET    /api/mobile/inbox
POST   /api/mobile/inbox/{batchID}/import
POST   /api/mobile/inbox/{batchID}/ignore
POST   /api/mobile/inbox/{batchID}/resolve
```

The mobile upload API should receive raw capture data and media first. Existing inventory, wishlist, barcode, item instance, import, and photo APIs can be used internally once the desktop user approves the batch.

## Desktop database additions

Suggested tables:

```text
mobile_devices
mobile_pairing_sessions
mobile_upload_batches
mobile_upload_items
mobile_upload_assets
mobile_scan_batches
mobile_scan_candidates
mobile_upload_events
mobile_import_decisions
```

## Sync transport phases

### Phase 1: LAN-first

Desktop exposes the upload API to the local network only while **Receive from Mobile** is enabled.

Mobile connects to the paired desktop address and uploads directly.

### Phase 2: temporary remote tunnel

Later, Cabinet may support temporary remote access through a tunnel provider.

Rules:

- tunnel must be explicit
- tunnel must be temporary
- pairing token must be short-lived
- permissions remain upload-scoped
- tunnel provider is transport only, not trust authority

### Phase 3: local/P2P event support

Later, mobile can support local trading and P2P workflows:

- nearby peer discovery
- QR pairing
- wishlist/binder exchange
- trade desk
- signed handoff receipt
- connection history
- offline receipt queue

This should not block the first mobile capture release.

## Tech recommendation

Given Cabinet currently has a React/Vite web UI and a Go/SQLite desktop runtime, the recommended first implementation is:

- **Mobile shell:** Capacitor + React/Vite
- **Mobile local DB:** SQLite
- **Mobile secure storage:** native keychain/keystore
- **Mobile media storage:** app-private filesystem
- **Desktop runtime:** existing Go runtime
- **Desktop persistence:** existing SQLite database
- **Desktop UI:** existing React/Vite UI patterns

A later native rewrite can be considered only if the capture, camera, OCR, or background upload requirements outgrow Capacitor.

## Privacy and safety rules

Mobile capture must follow these rules:

- do not publish anything to Git/Radicle directly in MVP
- do not expose the full desktop inventory to phone by default
- do not seed private media
- do not upload private storage locations unless the user explicitly enters them
- do not allow phone-side delete of desktop inventory
- do not silently import visual-only matches
- do not upload originals to public storage automatically
- keep originals local/private unless user imports or exports them
- show clear review states before desktop import

## MVP acceptance criteria

The mobile companion MVP is acceptable when:

- phone can pair with Cabinet desktop by QR
- desktop can revoke a paired mobile device
- phone can capture item metadata and photos offline
- phone can capture wishlist notes offline
- phone can scan barcode/QR into a capture item
- phone can create bulk image scan batches
- phone can cache originals, thumbnails, crops, and manifests
- uploads retry safely without duplicate inventory
- desktop receives uploads into Mobile Upload Inbox
- desktop verifies asset hashes
- desktop can import as inventory, wishlist, item instance, or item photos
- desktop can ignore or resolve batches
- duplicate candidates are visible before import
- visual-only candidates require review
- mobile cannot delete or overwrite desktop inventory in MVP

## Suggested issue breakdown

1. Spec: Mobile companion capture and upload workflow
2. Backend: mobile pairing and device authorisation
3. Backend: mobile upload batch API and DB tables
4. Desktop UI: Mobile Upload Inbox
5. Mobile: Capacitor app shell and pairing flow
6. Mobile: offline capture cache and upload queue
7. Mobile: camera/photo asset pipeline with hashes and thumbnails
8. Mobile: barcode/QR scan into capture item
9. Mobile: OCR extraction from still images
10. Mobile: bulk image scan candidate packets
11. Backend: scan candidate matching and duplicate detection
12. Inventory: import scan candidate as item / instance / wishlist
13. Media: link original image and crop assets to imported items
14. Validation: idempotent retry and duplicate upload tests

## First implementation slice

The best first slice is:

> Desktop pairing + upload inbox API skeleton.

Reason:

- it creates the contract the mobile app will use
- it fits Cabinet's desktop-first authority model
- it does not require a finished mobile UI first
- it lets local tests verify pairing, upload, idempotency, and inbox state early

After that, implement the mobile shell and photo/barcode capture against the stable contract.
