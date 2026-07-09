# Mobile Capture Companion and Collector Differentiation Plan

Status: planning note  
Date: 2026-07-10  
Target branch: `develop`

## Purpose

This document captures product planning ideas from the iCollect Everything competitor review and follow-up Cabinet planning discussion.

It adapts those ideas to Cabinet's actual direction: a desktop-first, local-first collector workspace where inventory, wishlist, collections, media, search, scanner, matching, integrations, assistant workflows, storage, backup, import/export and maintenance form the current product spine.

The planning goal is not to copy a cloud-first collection database. Cabinet should learn from the strongest parts of existing collector apps, then apply Cabinet's own principles:

- desktop-first main workspace
- local-first source of truth
- private inventory by default
- mobile as a scanner/camera companion
- optional sync and backup targets
- complete export, including media
- trading and collector trust as future differentiators
- selling as secondary
- no escrow or payment custody in Cabinet core

## Core lesson from iCollect-style apps

Existing collector apps prove demand for broad inventory tracking, barcode capture, photos, wishlists, estimated values, custom fields, cloud sync and sharing.

Cabinet should match the everyday utility collectors expect, but win on the areas where generic apps feel weak:

- slow entry during bulk capture
- poor offline behaviour in garages, stores, swap meets and conventions
- missing niche fields for trading cards and slot cars
- weak migration paths when images and custom data do not export cleanly
- unclear data ownership and privacy
- generic collection models that do not support active hobbies such as racing, tuning, trading and set completion

Cabinet's advantage should not simply be "more fields". It should be faster capture, safer ownership, better offline workflows and niche depth for cards and slot cars.

## Product position

Cabinet should be positioned as:

> A local-first collector workspace for inventory, evidence, wishlist, trade preparation and trust.

A concise competitive contrast:

> iCollect helps you catalogue what you own. Cabinet helps you privately own, verify, prepare, trade and carry your collector reputation with you.

A launch-friendly version:

> Scan fast, organise properly, export everything, and prepare trades without relying on reception or a marketplace account.

## Recommended product shape

### Desktop Cabinet

The desktop app remains the main workspace.

Desktop responsibilities:

- review captured items
- reconcile duplicates
- match scans/photos to catalogue entries
- edit detailed metadata
- organise inventory and wishlist
- manage collections
- attach and review media/evidence
- run search, scanner and matching flows
- manage imports, exports, backups and storage
- prepare trade collections and wishlists
- later: publish selected public binders, wishlists, receipts and feedback

### Mobile Capture Companion

The native iOS/Android app should start as a companion, not as a full replacement for desktop Cabinet.

Mobile responsibilities:

- pair with the local Cabinet instance
- scan barcodes, QR codes and item identifiers
- capture item photos quickly
- apply quick category, condition and variant tags
- queue captures offline
- sync captures into the desktop Capture Inbox
- allow unresolved scans to be fixed later
- avoid requiring cloud sync for basic capture

Simple mental model:

```text
Phone = scan, photograph, tag, queue
Desktop = review, match, organise, save, prepare trades
```

## Mobile pairing flow

Preferred MVP flow:

1. User opens Cabinet desktop.
2. User opens "Pair mobile capture app".
3. Desktop displays a QR code and optional short pairing PIN.
4. Mobile scans QR / enters PIN.
5. Mobile establishes a local session with the Cabinet instance.
6. User chooses whether captures should sync over local network only, direct cable/local transfer if supported, or configured sync target later.
7. Paired device appears in Cabinet settings with last sync time and permission scope.

Pairing should be explicit. The phone should not silently attach to the user's inventory.

## Capture Inbox

The Capture Inbox is the key Cabinet workflow that turns fast phone capture into trustworthy desktop inventory.

Captured items should not immediately pollute inventory. They should land in a review queue with confidence and action states.

Each capture should include:

- captured photo(s)
- scan source: barcode, QR, photo, manual, imported CSV row
- detected category
- suggested catalogue match
- match confidence
- unresolved identifier if no match exists
- duplicate warning if relevant
- quick tags: condition, variant, boxed, graded, language, scale, etc.
- source device
- captured timestamp
- sync status

Suggested capture states:

| State | Meaning |
|---|---|
| `Captured` | Raw capture saved locally on device. |
| `Queued` | Waiting to sync to Cabinet desktop. |
| `SyncedToInbox` | Available in desktop Capture Inbox. |
| `Matched` | Has a candidate catalogue/item match. |
| `NeedsReview` | Ambiguous, incomplete or duplicate-prone. |
| `Approved` | User accepted and saved to inventory/wishlist. |
| `Rejected` | User discarded the capture. |
| `Deferred` | User kept it for later review. |

Desktop inbox actions:

- approve one
- approve all high-confidence matches
- edit before save
- merge with existing item
- create a new item
- create a wishlist item
- attach as photo/evidence to an existing item
- reject
- defer

## Camera-first capture UX

The mobile capture screen should be fast and one-handed.

Core controls:

- active camera viewfinder
- barcode / QR target
- photo capture button
- flash toggle
- category quick selector
- rapid attribute toggles
- condition selector
- "log item and scan next" action
- unresolved/offline indicator

Example quick toggles:

Trading cards:

- foil / holo
- reverse holo
- first edition
- graded
- raw
- promo
- language

Slot cars:

- boxed
- unboxed
- analogue
- digital
- modified
- needs maintenance
- missing parts

The UX should optimise for "scan now, clean up later".

## Offline-first behaviour

Cabinet should treat offline capture as normal, not degraded.

Baseline requirements:

- Add, scan, photograph, edit and queue captures without internet.
- Search local inventory and cached catalogue data where available.
- Store thumbnails locally for fast browsing.
- Store original media according to configured Cabinet media rules.
- Keep a local sync queue with visible status.
- Never lose work because the network dropped.
- Resolve unknown scans later when providers or catalogue data are available.

The sync queue should target the user's Cabinet instance first. Optional cloud sync, NAS backup, Git/Radicle proof publication or provider sync should be layered on top, not treated as the primary inventory authority.

## Media and evidence model implications

Cards and slot cars both require strong media handling.

Cabinet should support:

- multiple photos per item/specimen
- primary photo
- thumbnails for fast browsing
- originals stored locally or in configured media storage
- media export with inventory data
- evidence labels such as damage, box, certificate, serial, underside, chassis, receipt, grading proof
- rebuild/repair media paths
- clear backup and restore behaviour

Media must not be trapped inside a subscription or cloud account. Export with images is a major trust feature.

## Trading card template direction

Cabinet should start with practical trading card support rather than a generic blank template.

Candidate fields:

- game/category: Pokémon, Magic, sports, other
- set
- card number
- name
- variant
- language
- edition
- foil/holo/reverse holo
- first edition flag
- promo flag
- graded/raw state
- grader: PSA, BGS, CGC, other
- grade
- certification number
- condition for raw cards
- quantity
- specimen records for high-value individual cards
- binder/deck/trade pile location
- purchase/source notes
- photos: front, back, certificate, damage

Trading card workflows:

- set completion tracking
- duplicate/quantity handling
- wishlist wanted quantity
- move wishlist item to owned inventory
- trade pile / binder view
- value by condition and variant
- unresolved card scan review

## Slot car template direction

Slot cars need to support collectors who display, race, tune, repair and restore.

Candidate fields:

- brand: Scalextric, Carrera, Slot.it, Fly, Ninco, etc.
- model
- scale
- series
- year
- livery
- catalogue number
- boxed/unboxed
- box condition
- car condition
- missing parts
- restoration notes
- chassis type
- motor type / RPM
- gear ratio
- tyres
- braid
- guide blade
- magnet setup
- digital chip / analogue state
- track compatibility
- photos: car, chassis, box, certificate, damage, underside

Slot car workflows:

- tuning log
- maintenance log
- lap time log
- track layout association
- parts wishlist
- restoration evidence
- race-ready vs display-only status

This gives Cabinet a strong wedge because generic collection apps often treat slot cars like static products rather than active hobby objects.

## Custom category templates

Cabinet should not become an overwhelming blank database builder.

Recommended approach:

- provide starter templates for core categories
- allow users to add custom fields
- support field types such as text, number, date, money, enum, multi-select, boolean, URL and file/photo reference
- allow category-specific saved views
- export custom schema with data
- preserve custom fields during import/export

This gives power users control without forcing casual collectors to build everything from scratch.

## Import and migration lessons

CSV export from existing apps creates a migration path, but text-only CSV is not enough.

Cabinet should build a migration assistant that can:

- import CSV
- detect likely source app/export shape
- map columns to Cabinet fields
- preserve unknown columns as custom fields
- identify category and template candidates
- detect duplicates
- match imported rows to catalogue entries
- fetch missing metadata where configured
- flag missing images
- allow review before apply
- produce an import report

Migration success should be measured by how safely Cabinet can reconstruct useful inventory, not merely whether rows import.

## Valuation and marketwatch direction

Estimated values are a major user hook, but Cabinet should avoid opaque value numbers.

Cabinet should aim for source-aware valuation:

- source provider
- sold vs listed price distinction
- date checked
- price range
- median / recent comparable
- confidence
- condition and variant adjustment
- currency
- stale data warning
- manual override
- value history where available

Marketwatch and Discoveries should use saved searches across integrations to surface:

- wishlist matches
- underpriced items
- price movement
- new listings
- watched categories
- trade opportunities

## Data ownership and subscription principles

Cabinet should have a clear no-data-hostage policy.

Suggested product rule:

> If a user stops paying, they should still be able to open, view, edit, back up and export their local Cabinet inventory.

Paid features should gate services, not ownership.

Better paid feature candidates:

- multi-device sync
- cloud backup convenience
- valuation refreshes
- provider integrations
- advanced AI-assisted matching
- large-scale automation
- store/club tools
- public proof/reputation automation

Features that should remain core/local:

- view local inventory
- edit local inventory
- manual item add
- local search
- local media access
- export CSV/JSON
- export media bundle
- local backup/restore

Avoid language such as "whale" in pricing or product tiers. It undermines Cabinet's collector-owned trust positioning.

## Pricing hypotheses, not commitments

Pricing needs validation, but the useful lesson is to frame paid value against real hobby costs.

Potential tiers to test:

| Tier | Purpose |
|---|---|
| Free Collector | Let users trust Cabinet with a meaningful starter collection. |
| Collector Plus | Unlimited inventory, sync/backup convenience, valuation and advanced capture. |
| Pro Collector | Power collectors with large inventories, multi-location storage, bulk workflows and advanced integrations. |
| Store / Club | Venue, club, store and event features. |

Do not over-optimise pricing before the core capture and inventory experience is proven.

## Community launch messaging lessons

The marketing lesson is to speak to collector pain directly, not generic SaaS benefits.

Good slot car angle:

> Most collection apps treat slot cars like books. Cabinet should understand cars you actually run, tune, repair and restore.

Good card angle:

> Scan now, clean up later. Cabinet should work at conventions even when reception does not.

Good ownership angle:

> Your collection data should not be held hostage by a marketplace account or subscription.

Recommended early community posture:

- ask for feedback, do not overclaim
- show concrete workflows
- avoid "ultimate app" hype
- avoid unproven performance claims
- do not claim on-device AI until shipped and validated
- lead with local-first capture, export, niche templates and ownership

## MVP roadmap implications

### Now / near-term

- Mobile Capture Companion planning
- Capture Inbox desktop surface
- barcode / QR capture
- photo capture
- quick tags
- local unresolved capture queue
- duplicate detection
- trading card starter template
- slot car starter template
- local media storage rules
- full export with media planning
- iCollect/generic CSV migration assistant planning

### Next

- bulk scan mode
- set completion for cards
- wishlist wanted quantity and conversion to inventory
- source-aware valuation
- marketwatch saved searches
- slot car maintenance log
- slot car tuning log
- slot car lap time log
- paired-device management
- optional mobile-to-desktop local sync

### Later

- AI-assisted card recognition
- AI-assisted slot car matching
- on-device model exploration
- cloud-assisted fallback matching where explicitly configured
- public binder/wishlist publishing
- QR trade handoff
- signed receipts
- signed feedback
- store/venue event workflows
- public proof/reputation automation

## Candidate backlog issues

These ideas should become issues only when they are ready to be sliced and bound to OpenSpec/spec evidence.

Suggested issue titles:

1. Plan Mobile Capture Companion pairing model
2. Add Capture Inbox product surface
3. Define capture item states and review actions
4. Add barcode/QR capture flow for mobile companion
5. Add photo capture flow for mobile companion
6. Add quick condition and variant tagging model
7. Define trading card starter category template
8. Define slot car starter category template
9. Add local unresolved scan queue
10. Add batch review actions for captured items
11. Plan full export with media bundle
12. Plan iCollect CSV migration assistant
13. Add wishlist-to-inventory acquisition workflow
14. Plan source-aware valuation model
15. Plan slot car maintenance and tuning logs
16. Plan trading card set completion tracking
17. Define no-data-hostage subscription policy

## Non-goals for this planning slice

- Do not move Cabinet to a cloud-first inventory database.
- Do not make the mobile app the primary Cabinet workspace in MVP.
- Do not introduce escrow/payment custody.
- Do not make selling the core product direction.
- Do not claim AI recognition until validated.
- Do not publish private inventory by default.

## Open questions

- Which mobile framework should be used for the first companion app: native Swift/Kotlin, React Native, Flutter, or another approach?
- Should the first mobile sync use local HTTP, WebSocket/WebRTC, file handoff, or a simpler LAN pairing model?
- How much catalogue data should be cached locally on mobile for offline matching?
- What media size/quality rules should apply to mobile thumbnails vs originals?
- Should Capture Inbox live under Inventory, Scanner, or as its own navigation item?
- What is the minimum useful card catalogue seed?
- What is the minimum useful slot car catalogue seed?
- How should custom fields be stored so imports, exports and future public templates remain safe?

## Summary

The strongest competitor lesson is not that Cabinet needs to become a bigger iCollect Everything clone.

Cabinet should become a collector-owned workspace where capture is fast, offline work is safe, media is exportable, categories fit real hobbies, and future trading/reputation workflows grow from trustworthy local inventory.
