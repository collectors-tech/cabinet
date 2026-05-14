# eBay Seller, Buyer-Interest, Purchase Capture, and Freight Forwarding Feature Plan

## Status

Planning document. This is not an implementation spec yet.

This plan expands Cabinet from a collector inventory app with marketplace search into a workflow hub for:

- eBay seller operations
- eBay buyer-interest tracking
- purchase capture when official APIs are incomplete
- package-forwarder reconciliation, initially Stackry-style workflows
- landed-cost calculation and consolidation planning

## Product goal

Cabinet should become the front end for day-to-day eBay collector commerce.

A collector should be able to manage items in Cabinet, publish and maintain listings on eBay, capture purchases made on eBay, reconcile purchased items with freight-forwarder packages, and understand true landed cost without needing to manually stitch together eBay pages, emails, package-forwarder screens, spreadsheets, and inventory records.

## Non-goals and constraints

- Cabinet must not bypass eBay login, scrape credentials, or evade platform controls.
- Any page capture must be user-authorized and user-present.
- Cabinet must not silently publish listings, send offers, delete watchlist entries, or change order/listing state without explicit confirmation.
- eBay API capability must be detected and represented honestly. Some buyer-history, cart, saved-for-later, message, or notification data may not be available through public APIs.
- Stackry public package-management API availability is not confirmed. The referenced `docs.stackery.io` URL is for Stackery/AWS serverless API Gateway documentation, not Stackry package forwarding. Cabinet should support official API integration if available, but also support browser capture, email parsing/import, CSV import, and manual entry.

## Feature areas

### 1. eBay seller account connection

Goal: Cabinet can connect to eBay as a seller account and understand the seller's marketplace context and policies.

Capabilities:

- OAuth/browser login for eBay seller account connection.
- Profile-scoped token storage and refresh.
- Marketplace selection, for example AU or US.
- Seller policy import for payment, return, and fulfilment/shipping policies.
- Health state for auth, marketplace, policy availability, and API errors.
- Clear setup guidance when permissions or policies are missing.

Primary APIs to investigate:

- eBay Sell Account API
- eBay Sell Inventory API auth requirements
- eBay seller communication/notification permissions

### 2. Listing draft workspace

Goal: create eBay-ready listing drafts from Cabinet inventory items before anything is published.

Capabilities:

- Start a listing draft from an Inventory item.
- Map Cabinet item fields to eBay listing fields:
  - title
  - description
  - category
  - item specifics/aspects
  - condition
  - price
  - quantity
  - photos
  - shipping/payment/return policies
- Validate required eBay fields before publishing.
- Show blocked/ready status with exact missing fields.
- Preview the listing before publish.
- Preserve draft history and source Cabinet item linkage.

Safety requirement:

- Publishing must require explicit confirmation and must show the exact listing target, price, quantity, marketplace, and policies.

### 3. Publish, revise, end, and relist eBay listings

Goal: Cabinet can manage live eBay listing lifecycle from inside the app.

Capabilities:

- Create/update eBay inventory item records.
- Create offers/listings from Cabinet listing drafts.
- Publish offers/listings.
- Store eBay listing IDs, offer IDs, SKU/provider IDs, and sync metadata against Cabinet items.
- Revise price, quantity, title/description, photos, and policies where eBay permits.
- End listings from Cabinet with explicit confirmation.
- Relist or sell-similar from prior Cabinet/eBay listing state.
- Detect external changes made directly on eBay and surface drift.

Primary APIs to investigate:

- eBay Sell Inventory API
- eBay taxonomy/category metadata APIs
- eBay image/media requirements

### 4. Listing sync and sales state

Goal: Cabinet reflects current eBay listing status accurately.

Capabilities:

- Sync listing status: draft, active, ended, sold, error, externally changed.
- Sync live price, quantity, watch/interested signals when available, and listing URL.
- Link sold listings to Cabinet inventory items.
- Mark Cabinet item state as listed, reserved, sold, dispatched, or returned as appropriate.
- Store listing sync events for audit/debug.

### 5. eBay messages and seller notifications

Goal: bring eBay seller communication and action-needed events into Cabinet.

Capabilities:

- Pull or receive item-related buyer messages where API support allows.
- Link messages to Cabinet item, eBay listing, order, and buyer where possible.
- Reply from Cabinet where API support allows.
- Support templates/snippets for common replies.
- Normalize notification events into Cabinet:
  - buyer message
  - item sold
  - payment/order status update
  - listing ended/revised/published
  - return/case/dispute
  - auth/token/policy error
  - shipping/tracking update
- Route important events into a Seller Inbox / Action Queue.

Primary APIs to investigate:

- eBay Sell Communication APIs
- eBay Notification APIs and topic support
- eBay Fulfillment/order events

### 6. Orders and fulfilment

Goal: do seller follow-through inside Cabinet.

Capabilities:

- Fetch sold orders and order line items.
- Link each order line to Cabinet item/listing.
- Show buyer/payment/shipping status.
- Capture shipping service, tracking number, dispatch status, and delivery status where available.
- Create fulfilment tasks such as pack, label, dispatch, follow up.
- Preserve order history against the Cabinet item.

Primary APIs to investigate:

- eBay Sell Fulfillment API

### 7. Offers and negotiation assistance

Goal: use eBay signals plus Cabinet item knowledge to help make better offers.

Capabilities:

- Detect listings eligible for seller-initiated offers where eBay exposes the signal.
- Suggest offer opportunities using Cabinet context:
  - original cost
  - target margin
  - days listed
  - interested-buyer/watch signal if available
  - comparable prices
  - stock/urgency
- Draft offer amount and optional message.
- Require explicit confirmation before sending an offer.
- Track sent offers, expiry, accepted/declined/no-response outcomes.

Primary APIs to investigate:

- eBay Sell Negotiation API

Important caveat:

- Do not assume eBay exposes raw cart contents. Treat available data as sales/interest signals unless a specific API capability is confirmed.

### 8. Buyer-interest sync: watched, liked, saved, cart-like items

Goal: import eBay buyer-interest items into Cabinet and update eBay state where supported.

Capabilities:

- Sync watched/liked items from eBay into Cabinet.
- Represent imported items as Wishlist entries, Discoveries candidates, watched external listings, or price-watch records.
- Add/remove watchlist entries on eBay where supported.
- Support local-only wishlist tracking when eBay write-back is not supported.
- Use capability flags so UI only shows supported operations:
  - `can_read_watchlist`
  - `can_write_watchlist`
  - `can_read_cart`
  - `can_write_cart`
  - `can_read_save_for_later`
  - `can_write_save_for_later`

Primary APIs to investigate:

- eBay Trading API `GetMyeBayBuying`
- eBay Trading API `AddToWatchList`
- eBay Buy/Browse shopping cart APIs, noting limited-release/permission constraints

Safety requirement:

- Removing an item locally must not silently remove it from eBay. Cabinet should ask whether to update eBay too.

### 9. Browser companion purchase capture

Goal: fill the buyer purchase-history gap safely when eBay does not expose a proper purchase-history API.

Approach:

- A browser companion extension or local companion app observes eBay pages the user is actively viewing.
- It extracts structured purchase/order data with explicit permission.
- It sends records to Cabinet through a local authenticated endpoint.

Pages to detect:

- eBay item page
- cart/checkout page where available
- order confirmation page
- purchase history page
- order details page
- tracking/shipping update page
- seller message or order communication page where applicable

Captured fields:

- eBay item/listing ID
- title
- seller
- item URL
- image URL where visible
- quantity
- item price
- shipping price
- tax/import charges where visible
- total price
- currency
- order ID / transaction ID where visible
- purchase date
- estimated delivery
- destination address or freight-forwarder/suite marker where visible
- carrier
- tracking number
- order status
- seller update/status text
- return/case flags where visible
- optional saved screenshot or HTML snapshot for audit/debug

Privacy/safety rules:

- No credential capture.
- No hidden background scraping.
- No aggressive crawling.
- No bypassing eBay access controls.
- User can review before sending to Cabinet.
- Companion can be disabled per site/account.

### 10. Cabinet Purchase Inbox

Goal: provide a central queue for captured purchases and follow-up states.

States:

- captured from item page
- checkout detected
- purchased / confirmation captured
- order details enriched
- awaiting tracking
- shipped to forwarder
- arrived at forwarder
- matched to package
- consolidated
- shipped internationally
- arrived locally
- reconciled / closed
- needs review

Capabilities:

- Show captured eBay purchases grouped by purchase order.
- Treat the purchase order as the parent/container record and purchased items as child/sub-item records.
- Create a clear order folder/card for each order ID, for example eBay order `20-14595-70928`.
- Show order-level metadata on the parent:
  - order ID
  - purchase date
  - seller(s)
  - order total
  - currency
  - shipping/tax/import charges where visible
  - destination/forwarder marker where visible
  - order status
  - order detail URL
- Show item-level metadata on child records:
  - listing ID
  - variation ID
  - transaction ID
  - purchased variant/card metadata
  - item price
  - quantity
  - tracking/status where item-specific
  - seller note/action metadata where item-specific
- Support multi-item orders, one-order-many-items, and one-item orders without flattening the structure.
- Merge duplicate captures from multiple pages/events into the same order folder and child item records.
- Link purchase item to existing Cabinet item or create a new item.
- Convert purchase into Inventory item, Wishlist item, or pending incoming item.
- Show missing fields and suggested next action at both order and item level.
- Preserve capture event history.

### 11. Stackry / freight-forwarder package inbox

Goal: track incoming freight-forwarder packages and reconcile them to purchases.

Stackry API caveat:

- A public Stackry package-management API was not confirmed during planning. Cabinet should support an adapter interface so Stackry can be integrated through official API if available, but should initially support capture/import/manual workflows.

Ingestion modes:

- official API adapter if Stackry provides access
- browser companion capture from Stackry package pages
- email parser/import from Stackry arrival/ship/consolidation emails
- CSV import
- manual package entry

Captured package fields:

- forwarder account/provider
- package ID / inbox ID
- sender/merchant/seller
- received date
- warehouse/location status
- weight
- dimensions
- declared value
- photos where available
- storage fees
- handling fees
- consolidation/repack fees
- outbound shipment ID
- outbound carrier/tracking number
- final international shipping cost
- destination country
- shipment status

### 12. Purchase-to-package reconciliation

Goal: match eBay purchases to forwarder packages.

Matching signals:

- tracking number
- carrier
- seller name
- delivery date window
- purchase/order amount
- declared value
- item title keywords
- package sender text
- suite/customer ID
- package photos/labels if captured
- expected arrival date

Match states:

- unmatched purchase
- unmatched package
- suggested match
- confirmed match
- rejected match
- split package / multiple purchases
- consolidated package / multiple packages

Capabilities:

- Suggest likely matches with confidence score.
- Allow manual confirmation/override.
- Support one purchase to multiple packages and many purchases to one package.
- Keep audit trail of match decision and evidence.

### 13. Landed-cost calculation

Goal: calculate the real cost of an item after purchase price, domestic shipping, taxes, forwarder fees, international shipping, and handling.

Cost components:

- item price
- domestic shipping
- sales tax/import charge at source where visible
- currency conversion
- Stackry/forwarder handling fee
- consolidation fee
- repack fee
- storage fee
- insurance fee
- international shipping fee
- customs/duty/GST estimate where applicable
- manual adjustment

Allocation methods:

- by item value
- by weight
- by volume
- equal split
- manual override
- hybrid rule, for example shipping by weight and consolidation fee by item count

Outputs:

- landed cost per item
- landed cost per collection/category
- cost basis for future resale/listing pricing
- purchase vs resale target margin
- foreign buying viability estimate

### 14. Consolidation planner

Goal: recommend outbound freight-forwarder shipments that balance cost, timing, risk, and destination value thresholds.

Inputs:

- unmatched/available forwarder packages
- item/package declared values
- weight/dimensions
- destination country
- country value threshold, for example USD 600 per parcel to AU
- storage deadline/fees
- urgency
- prohibited/fragile/oversized constraints
- preferred carrier/service
- expected future arrivals

Capabilities:

- Recommend parcel groupings under a configured declared-value limit.
- Explain why items are grouped or split.
- Warn when a package would exceed country threshold.
- Suggest waiting for soon-arriving packages when economical.
- Suggest shipping now when storage fees or deadlines make waiting worse.
- Estimate landed cost impact per consolidation option.
- Produce a consolidation plan that can be marked as executed once the forwarder shipment is created.

Example output pattern:

- Parcel A: items 1, 2, and 5; declared value USD 587; under AU threshold.
- Parcel B: item 3 alone; oversized.
- Hold item 4; expected paired package arriving within 3 days.

### 15. Inventory integration

Goal: make purchase and freight-forwarding workflows part of Cabinet's item lifecycle.

Capabilities:

- Create an incoming inventory item from a captured eBay purchase.
- Attach eBay source URL, order ID, seller, purchase price, and evidence snapshots.
- Track incoming status before the item physically arrives.
- Mark item arrived when local delivery is reconciled.
- Store landed cost as item cost basis.
- Use landed cost when planning resale/listing price.
- Preserve purchase/order/forwarder timeline on the item detail page.

## Proposed Cabinet surfaces

### eBay Setup

- seller account connection
- buyer-interest capability detection
- marketplace selection
- policy import
- health/status

### Listings

- local drafts
- active eBay listings
- revise/end/relist
- publish confirmation

### Seller Inbox

- buyer messages
- listing notifications
- order/action-needed events
- cases/returns/disputes where supported

### Orders & Offers

- sold orders
- fulfilment state
- offer opportunities
- sent offer outcomes

### Buyer Interest

- watched/saved/cart-like items where supported
- wishlist handoff
- price watch
- local-only tracking fallback

### Purchase Inbox

- browser-captured purchases
- enriched order details
- unmatched/needs-review purchases

### Forwarding / Stackry Inbox

- incoming forwarder packages
- package status
- fees and shipment data

### Reconciliation Workspace

- purchase-to-package matching
- manual confirmation
- match audit trail

### Consolidation Planner

- recommended parcel groups
- country/value threshold checks
- landed-cost estimates

### Item Timeline / Cost Basis

- listing events
- purchase events
- package events
- landed cost
- resale planning

## Data model candidates

- `provider_accounts`
- `provider_account_capabilities`
- `ebay_seller_profiles`
- `ebay_business_policies`
- `listing_drafts`
- `provider_listings`
- `listing_sync_events`
- `seller_messages`
- `seller_notifications`
- `orders`
- `order_items`
- `fulfillment_tasks`
- `buyer_interest_items`
- `purchase_orders`
- `purchase_order_items`
- `purchase_capture_events`
- `freight_forwarder_accounts`
- `forwarder_packages`
- `forwarder_shipments`
- `package_purchase_matches`
- `landed_cost_components`
- `landed_cost_allocations`
- `consolidation_plans`
- `consolidation_plan_items`

## Existing extension review notes

Reviewed local Chrome extension source:

`C:\projects\collectors-tech\extensions\dhccpfcjgmlajnnoigjhokbfgpaamhpe\3.8_0`

Extension identity:

- Name: `Ebay Purchase History Downloader`
- Version: `3.8`
- Manifest: MV3
- Primary files: `manifest.json`, `popup.html`, `popup.js`, `content.js`
- Bundled libraries: jQuery, SheetJS, FileSaver

What it does well:

- Runs on eBay domains using content scripts and active-tab scripting.
- Supports multiple eBay domains including `ebay.com`, `ebay.co.uk`, `ebay.com.au`, and `ebay.com.ca`.
- Uses the user's existing eBay logged-in browser session instead of requiring separate credentials.
- Calls eBay's purchase-history AJAX endpoint from the page context:
  - `/mye/myebay/ajax/v2/purchase/mp/get`
  - year filter such as `CURRENT_YEAR`, `LAST_YEAR`, etc.
  - paginates up to 300 pages.
- Extracts itemized order data from the JSON response rather than only scraping rendered DOM.
- Captures useful fields:
  - order number
  - order date
  - item ID / listing ID
  - seller
  - item name
  - item price
  - currency
  - quantity
  - order total
  - order notes
  - tracking number where present
  - image URL
  - order detail URL
  - optional ISBN by fetching item detail pages
- Exports the captured records to an Excel workbook.

Important implementation observations:

- `content.js` is very small and only returns the visible selected date filter from `.expand-btn__cell`.
- Most extraction logic is injected from `popup.js` using `chrome.scripting.executeScript`.
- The extension does not use a long-running background service worker.
- The extension does not maintain a structured local database or sync target; it only produces a file download.
- Data extraction is tightly coupled to eBay's internal purchase-history JSON shape, especially `modules.RIVER[0].data.items`, `itemCards`, and `__myb` fields.
- The year filter mapping is hard-coded and currently supports years by relative labels.
- The current host matching includes `*.ebay.com.ca`, which may not be the canonical Canadian eBay domain; Cabinet should verify all host patterns.
- Error handling is light: failed AJAX or selector drift can degrade silently unless logs are inspected.
- There is no schema versioning, confidence score, partial-capture warning model, or merge/de-duplication logic.

Useful lessons for Cabinet:

- The best base idea is not the Excel export; it is the browser-session capture technique.
- eBay purchase history can be captured more reliably by calling eBay's logged-in internal JSON endpoint than by reading only visible DOM rows.
- Cabinet should keep this endpoint-capture strategy as one eBay module implementation option, but it must be treated as unofficial and drift-prone.
- Cabinet should replace file export with direct structured sync into a local authenticated Cabinet endpoint.
- Cabinet should preserve the useful captured field set and extend it with capture metadata, confidence, source URL, source module, and raw/normalized payload separation.
- Cabinet should support pagination/range reporting so the user knows whether a full year/range was captured.
- Optional enrichment such as ISBN/item-page fetches should be module capabilities, not hard-wired into the purchase-history importer.
- Tracking/order-detail extraction should be represented as enrichment with confidence and should not be assumed complete.
- Selector/internal-JSON drift must be handled with fixtures and module tests.

How to use this as a base:

- Keep the idea of a Chrome MV3 extension.
- Keep eBay-domain content-script/module permissions.
- Reuse the field map and internal endpoint discovery as a reference.
- Do not keep the single-purpose popup/download architecture as-is.
- Refactor into a Cabinet extension host plus modules.
- Replace Excel generation with Cabinet capture envelopes.
- Add module registry, payload schemas, validation, confidence, retry, and local sync.

The existing extension should therefore be treated as a prototype/reference for the first `ebay.purchase-history` module, not as the final Cabinet extension app architecture.

## Companion architecture

Cabinet should provide a **Cabinet Browser Extension app** rather than a one-off eBay extension.

The extension should be a small host/runtime with separately registered capture modules. Each module owns one external site or workflow and emits normalized capture payloads into Cabinet.

### Browser extension host responsibilities

- Manage connection to the local Cabinet app.
- Authenticate with Cabinet using a pairing token or local approval flow.
- Display global extension status: connected, disconnected, capture available, capture failed.
- Load registered site modules based on URL match patterns and user-enabled permissions.
- Provide shared UI primitives: capture button, review panel, confidence warnings, send-to-Cabinet action, error display.
- Provide shared services: DOM snapshot helpers, URL normalization, currency/number/date parsing, local queue/retry, structured logging, module version reporting.
- Prevent modules from sending destructive actions directly; modules only submit captures or user-confirmed intent payloads.

### Capture module responsibilities

Each module should declare:

- `module_id`, for example `ebay.purchase-history`, `ebay.order-detail`, `stackry.package-detail`.
- version.
- supported hostnames and URL patterns.
- required browser permissions.
- capture types it can emit.
- extractor functions.
- confidence scoring rules.
- redaction/privacy rules.
- schema version for emitted payloads.
- tests/fixtures for supported page shapes.

Each module should be easy to add without changing the whole extension runtime.

### Initial capture modules

#### `ebay.item-page`

Purpose: capture a listing the user is viewing.

Fields:

- listing ID
- title
- seller
- current price
- shipping estimate where visible
- item URL
- image URL
- condition
- quantity/availability where visible
- watched/saved state where visible

Cabinet action:

- add to Wishlist
- add to Buyer Interest
- create price-watch record
- compare against Inventory

#### `ebay.purchase-history`

Purpose: capture visible purchase-history rows/ranges.

Fields:

- order date
- seller
- listing/sale title
- purchased variant/card metadata
- item/listing ID
- variation ID where present
- quantity
- item price
- order total
- shipping status
- estimated delivery
- thumbnail/image URL
- order detail URL
- notes where visible
- tracking number if visible and confidence is high

Cabinet action:

- import/update Purchase Inbox records.

Important behavior:

- detect visible-row/range limits.
- warn when only current page/year/range is captured.
- support repeated captures and de-duplicate in Cabinet.
- preserve the sale/listing title separately from purchased item metadata.
- parse eBay purchase card aspect metadata from the item card body, not just the title.
- for collectible/card purchases, use aspect rows such as `Choose Single Or Playset`, `Card`, and `Quantity` to identify the actual purchased card/variant.
- treat metadata such as `Card: Accompanying Flute TWM 142 (142/167)` as the purchased item identity candidate even when the listing title is a broad sale/deck-builder title.
- preserve original aspect key/value pairs as structured metadata so Cabinet can map them into inventory fields later.
- capture seller identity from the card body, including seller username and seller profile URL, for example `nearmintormeta` and `/usr/nearmintormeta`.
- capture the per-item eBay note field, including existing note value, note textarea identity where useful, and note edit/save/delete capability metadata if Cabinet later supports note write-back.
- capture note action names and button/action metadata where visible, for example `CANCEL_EDIT_NOTE`, `DELETE_NOTE`, `SAVE_NOTE`, and related `data-action-name`, `data-action`, `aria-labelledby`, and tracking metadata.
- capture note action endpoint/capability hints separately, for example `https://www.ebay.com.au/myb/DeleteNote`, `https://www.ebay.com.au/myb/SaveNote`, and action params such as `note`, `itemId`, and `variationId`, without treating null params as valid target identifiers.
- capture whether note actions are currently enabled/disabled, for example `SAVE_NOTE` may render with `aria-disabled="true"` until note text changes.
- keep note capture separate from item identity metadata so user notes do not pollute item title/card fields.
- capture purchase-card CTA/action metadata separately from item identity, for example `LEAVE_FEEDBACK_FOR_SELLER`, feedback URL, `item_id`, and `transaction_id` from the feedback link.
- capture seller-discovery CTA metadata separately, for example `VIEW_SELLERS_OTHER_ITEMS`, seller storefront/search URL such as `/sch/<seller>/m.html`, and linked seller username.
- capture overflow/menu action affordances, for example `More actions` buttons, menu IDs, tracking metadata, `aria-expanded`, `aria-controls`, and whether the menu is collapsed/open, without opening menus during passive capture unless the user explicitly requests enrichment.
- when menu content is already present in the DOM, capture menu actions as passive metadata without clicking them, including actions such as `VIEW_ITEM` / `Buy again`, `CONTACT_SELLER`, `START_RETURN`, `EDIT_NOTE` / `Add note`, `HIDE` / `Hide order`, and `HELP_AND_REPORT`.
- preserve menu action URLs and parsed identifiers as reconciliation hints, for example listing URL with `var`, contact-seller URL with `item_id`, `requested` seller username, `transId`, return URL with `itemId` and `transactionId`, edit-note action metadata, and hide-order endpoint with `orderId`.
- preserve transaction/order/action identifiers from CTAs as reconciliation hints where visible, including order identifiers embedded in URLs such as `orderId=20-14595-70928`.
- classify destructive or visibility-changing actions separately, for example `HIDE` / `Hide order`, and never present them as passive metadata-only display links.
- do not execute note edit/delete/save actions, CTA/navigation actions, return/contact/buy-again actions, hide-order actions, help/report actions, or overflow menu actions during passive capture; any future note write-back, feedback action, seller navigation, buyer action, return action, hide action, help/report action, or menu action must be an explicit confirmed Cabinet workflow and must resolve concrete item/listing/variation/transaction/order identifiers before calling eBay endpoints or navigating.

#### `ebay.order-detail`

Purpose: enrich an existing purchase with full order details.

Fields:

- eBay order ID / transaction ID where visible
- tracking carrier
- tracking number
- delivery estimate
- delivery status
- shipping address / forwarder suite marker where visible
- itemized charges
- seller update/status messages
- return/case flags where visible

Cabinet action:

- enrich Purchase Inbox record.
- update shipment state.
- create matching hints for forwarder reconciliation.

#### `ebay.checkout-confirmation`

Purpose: capture a purchase immediately after checkout while the richest confirmation data is visible.

Fields:

- purchased item(s)
- final paid price
- shipping/tax/import charges
- destination/forwarder address marker
- order confirmation ID where visible
- estimated delivery

Cabinet action:

- create purchase records immediately.
- mark as bought/awaiting seller shipment.

#### `stackry.package-detail`

Purpose: capture package information from Stackry pages if no official API is available.

Fields:

- package ID
- sender/merchant
- received date
- weight/dimensions
- declared value
- package photos where visible
- storage/handling/consolidation/repack fees
- outbound shipment/tracking IDs
- status

Cabinet action:

- import/update Forwarding Inbox package records.

#### `stackry.shipment-detail`

Purpose: capture outbound consolidated shipment details.

Fields:

- shipment ID
- included packages
- outbound carrier
- tracking number
- shipping cost
- insurance/services
- declared value
- destination country
- shipment status

Cabinet action:

- update consolidation plan execution.
- allocate forwarding costs to matched purchases/items.

### Cabinet local endpoint responsibilities

- Authenticate extension requests.
- Validate capture payload schema.
- Store module ID, module version, source URL, capture timestamp, and confidence.
- De-duplicate captures.
- Merge updates into existing purchase/package records.
- Preserve raw capture event for audit/debug where allowed.
- Reject unknown module/schema versions unless explicitly allowed.
- Avoid executing destructive actions from capture payloads.

### Extension module registry

Cabinet should maintain a registry of installed/known extension modules.

Registry fields:

- module ID
- display name
- provider/site
- enabled state
- version
- host permissions
- supported capture types
- last successful capture
- last error
- schema version
- minimum Cabinet version

The registry should let the user enable/disable modules individually.

### Module payload envelope

All modules should emit a consistent payload envelope:

```json
{
  "source": "browser-extension",
  "module_id": "ebay.purchase-history",
  "module_version": "0.1.0",
  "schema_version": "capture.purchase.v1",
  "captured_at": "2026-05-12T00:00:00Z",
  "source_url": "https://www.ebay.com.au/...",
  "confidence": 0.92,
  "warnings": ["Only visible purchase-history rows were captured"],
  "records": []
}
```

### Test strategy for modules

Each module should ship fixtures:

- sanitized HTML snippets for supported page states.
- expected normalized payload JSON.
- selector-drift tests.
- locale/currency/date variants.
- missing-field cases.

A module should be considered broken if it silently emits low-quality data after selector drift. Low confidence or missing required fields must create a visible warning instead.

## Implementation phases

### Phase 0: Research and API capability map

Deliverables:

- eBay seller API capability matrix.
- eBay buyer-interest/watchlist capability matrix.
- eBay purchase-history capture feasibility notes.
- Stackry/forwarder API availability check.
- Companion privacy/security design.

### Phase 1: Companion capture to Purchase Inbox

Deliverables:

- Cabinet local companion endpoint.
- Browser companion proof-of-concept for eBay item and order confirmation pages.
- Purchase Inbox data model and UI.
- De-duplication and capture event audit trail.

Why first:

- This proves the hardest purchase-history gap before building forwarder matching or consolidation logic.

### Phase 2: Forwarder package inbox

Deliverables:

- Forwarder package model.
- Manual package entry.
- CSV/email/browser import path.
- Stackry adapter placeholder with explicit capability flags.
- Package status UI.

### Phase 3: Purchase-to-package matching

Deliverables:

- Matching engine.
- Suggested match UI.
- Confirm/reject/override workflow.
- One-to-many and many-to-one matching support.

### Phase 4: Landed-cost engine

Deliverables:

- Cost component model.
- Allocation methods.
- Item-level landed cost display.
- Manual override/audit trail.

### Phase 5: Consolidation planner

Deliverables:

- Country/value threshold rules.
- Parcel grouping recommendations.
- Fee/landed-cost estimates per plan.
- Mark consolidation plan as executed.

### Phase 6: eBay seller listing operations

Deliverables:

- Seller OAuth/account connection.
- Listing draft workspace.
- Publish/revise/end/relist workflows.
- Listing sync and drift detection.

### Phase 7: Seller inbox, orders, offers

Deliverables:

- Messages/notifications action queue.
- Sold order sync.
- Fulfilment task workflow.
- Negotiation/offer suggestions and confirm-before-send.

## Backlog issue split

Recommended epic:

- `epic(commerce): eBay seller and purchase-forwarding command centre`

Recommended child issues:

1. `feat(companion): capture eBay item and order confirmation data into Cabinet`
2. `feat(purchases): add Purchase Inbox for captured external purchases`
3. `feat(forwarding): add freight-forwarder package inbox with Stackry-ready adapter boundary`
4. `feat(reconciliation): match eBay purchases to forwarder packages`
5. `feat(costing): calculate item landed cost from purchase and forwarding fees`
6. `feat(consolidation): recommend forwarder parcel groupings under country value limits`
7. `feat(ebay): connect seller account and import seller policies`
8. `feat(ebay): create listing drafts from Cabinet inventory items`
9. `feat(ebay): publish revise end and relist eBay listings from Cabinet`
10. `feat(ebay): sync eBay listing status and external drift into Cabinet`
11. `feat(ebay): import seller messages and notifications into Seller Inbox`
12. `feat(ebay): sync sold orders and fulfilment tasks`
13. `feat(ebay): suggest and send seller offers with confirmation`
14. `feat(ebay): sync watched and buyer-interest items into Cabinet Wishlist`
15. `feat(companion): define privacy security and permission model for page capture`

## Open questions

- Which eBay marketplaces are first-class for the first release?
- Is seller work or buyer purchase capture the first priority?
- Does Stackry provide private/partner API access for package data?
- Which country threshold rules are required first besides AU USD 600 per parcel?
- Should Cabinet support multiple forwarders or start with a Stackry-labelled generic forwarder adapter?
- How much raw page evidence should Cabinet store: structured fields only, screenshots, sanitized HTML snapshots, or all optional?
- Should browser companion be Chrome-only first, or browser-extension plus local companion from the start?
- What is the minimum viable landed-cost calculation: equal split, value split, or manual allocation first?

## Recommended first slice

Start with:

`feat(companion): capture eBay item and order confirmation data into Cabinet`

Rationale:

- eBay buyer purchase history is the least reliable through official APIs.
- Captured purchase records are the foundation for Stackry reconciliation, landed cost, and consolidation planning.
- It can be built safely as user-present capture without touching live eBay seller state.
- It creates immediate value before the larger seller/listing integration is complete.
