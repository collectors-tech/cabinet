# Cabinet Beta MVP Plan

## Purpose

This document defines the immediate Cabinet Beta MVP, the Free/paid packaging model, and the first commercial validation path.

The goal is to make Cabinet sellable as a useful local-first collector workspace before the full decentralised trust network is implemented.

## Product thesis

Cabinet should first sell as a desktop-first, local-first collector workspace.

The immediate value proposition is:

> Cabinet helps collectors organise what they own, track what they want, find better buying opportunities, and prepare trades while keeping their collection private and portable.

Cabinet Free should create trust and usage habit. Paid Cabinet should charge for collector intelligence: saved searches, discoveries, matching, evidence, guided workflows, backup confidence, import/export polish, and trade preparation.

## Strategic split

### Immediate beta product

The Beta should focus on a useful app a collector can install, use, and pay for now:

- private inventory
- wishlist
- collections
- photos, media, and evidence
- search and scanner flows
- Market Watch
- Discoveries
- integrations
- assistant-guided workflows
- backup, restore, import, export, and maintenance

### Long-term moat

The long-term Cabinet architecture remains the differentiator, but should not block the first paid beta:

- collector-owned identity
- signed receipts
- signed feedback vault
- portable reputation
- public Git/GitHub/GitLab/Codeberg/Radicle proof ledgers
- store, venue, custodian, and master collector attestations
- Radicle and P2P support
- community verification

These should be framed as the trust roadmap, not as required beta functionality.

## Beta positioning

Suggested beta positioning:

> Cabinet is your private collector workspace: inventory, wishlist, collections, evidence, saved searches, discoveries, and trade preparation in one desktop-first app.

Suggested paid positioning:

> Cabinet Free organises your collection. Cabinet Plus helps you find the things you want before someone else does.

## Beta MVP scope

### Must-have Beta scope

| Area | Beta scope | Commercial reason |
|---|---|---|
| Onboarding | Create/select profile, starter checklist, sample data | Gets user to value quickly |
| Inventory | Add, edit, search, filter, view rows/cards, attach photos and identifiers | Core daily use |
| Wishlist | Add wanted items, priority/status, mark purchased into inventory | Turns collecting intent into action |
| Collections | Create/manage collections, assign/move items | Makes Cabinet feel like a real workspace |
| Media/evidence | Item photos, primary image, condition/evidence notes | Builds trust in the collection record |
| Import/export | CSV/JSON import/export, backup/restore | Reduces data-lock-in fear |
| Market Watch | Saved searches across integrations, wishlist matching, run history | Main paid conversion hook |
| Discoveries | Inbox/dashboard for watched results, wishlist matches, good-price candidates | Gives users a reason to return |
| Assistant | Guided workflows for common tasks | Reduces learning friction |
| Licensing | Free/Plus/Pro feature gates | Enables pricing tests |

### Preview-only or later Beta scope

These should be visible as roadmap items or paper prototypes only:

| Area | Beta treatment |
|---|---|
| Signed trade receipts | Local/QR concept or JSON prototype only |
| Feedback vault | Concept preview or local private record only |
| Public Git/Radicle identity | Experimental setup screen or roadmap copy |
| P2P local trading | Paper prototype or fake local demo flow |
| Store/venue nodes | Not in immediate MVP |
| Community governance | Not in immediate MVP |
| Escrow/payment | Explicitly out of Cabinet core |

## Packaging model

### Cabinet Free

Purpose: acquisition, trust building, local-first proof.

Include:

- one local profile
- manual inventory, wishlist, and collections
- basic item photos/media
- basic search and filters
- basic CSV/JSON export
- manual backup
- one saved Market Watch, manual run only
- sample data and guided onboarding

Candidate soft limits:

- 250 inventory items
- 100 wishlist items
- 100 media attachments
- 3 custom collections plus All Items

Free should never trap user data. Export should remain available.

### Cabinet Plus Beta

Target user: serious collectors.

Suggested test pricing:

- AUD 9/month
- AUD 79/year founding beta
- later normal annual price: AUD 99/year

Unlock:

- higher or unlimited local inventory limits
- unlimited collections/wishlists
- more media/evidence capacity
- 10 active Market Watches
- scheduled watch runs
- Discoveries dashboard
- wishlist matching
- import/export templates
- stronger backup/restore workflows
- guided assistant workflows

### Cabinet Pro Beta

Target user: power collectors, trade-heavy users, and small seller-style collectors.

Suggested test pricing:

- AUD 19/month
- AUD 179/year founding beta
- later normal annual price: AUD 199/year

Unlock:

- 50+ Market Watches
- advanced integrations
- purchase/reconciliation workflows
- bulk import/export
- advanced media/evidence workflows
- assistant do-it-with-me workflows
- early trade binder / receipt features
- priority beta support

### Future Store/Event tier

Not immediate Beta.

Potential later pricing:

- AUD 49-99/month per store, club, or event host

Potential unlocks:

- event check-in
- local catalogue cache
- store attestations
- event trade room
- local discovery support

## Pricing experiment

Run the first beta as a willingness-to-pay test.

### Test offers

| Offer | Purpose |
|---|---|
| Free Beta | Prove install, trust, and repeat usage |
| Founding Collector | AUD 49 for 12 months, limited first cohort |
| Cabinet Plus Beta | AUD 79/year or AUD 9/month |
| Cabinet Pro Beta | AUD 179/year or AUD 19/month |

Do not only ask users whether they would pay. Ask for payment, deposit, or explicit upgrade commitment.

### Success thresholds

For an initial invited cohort of 20-30 collectors:

- 12+ install and create or import real data
- 8+ use Cabinet twice in one week
- 5+ pay for Founding Collector, Plus, or Pro
- 3+ say Market Watch or wishlist matching is the reason they paid
- zero data-loss incidents
- all users can export their data without help

### Strong paid signals

- User imports a real collection.
- User adds more than 25 real wishlist items.
- User creates more than one Market Watch.
- User asks for more providers.
- User checks Discoveries more than once.
- User asks about backup/restore before paying.
- User wants to use Cabinet at an event, swap meet, or store night.

## Immediate backlog shape

Create or link an umbrella issue:

```text
epic(beta): package Cabinet Free and paid beta MVP
```

Suggested child issues:

1. Define Cabinet Free/Plus/Pro edition matrix.
2. Implement local licence/entitlement gate using existing profile licence support.
3. Build first-run Beta onboarding and starter checklist.
4. Harden Inventory for beta save/edit/search/photo flows.
5. Harden Wishlist for beta save, purchased-to-inventory, soft delete, and deleted filter flows.
6. Harden Collections for beta metadata, soft delete, All Items protection, and item reassignment flows.
7. Finish Market Watch as the paid conversion hook.
8. Build Discoveries dashboard for watched results and wishlist matches.
9. Add beta feedback and pricing-capture prompts.
10. Prepare beta installer/release lane and release notes.
11. Add assistant beta coverage for common workflows.

## Market Watch as the paid spine

Market Watch should be the strongest paid feature because it directly maps to money saved, missed opportunities avoided, and better buying decisions.

User-facing promise:

> Tell Cabinet what you are hunting for. Cabinet watches connected sources, matches results against your wishlist, and shows discoveries worth reviewing.

Minimum beta behaviour:

- create saved watch
- choose provider/source
- set keywords
- set target price or notes
- choose cadence where paid
- run manually where free
- show provider health
- show run history
- show discoveries/results inbox
- hand off result to wishlist, purchase review, or inventory

## Ready for private paid beta

Cabinet is ready for a private paid beta when:

- Inventory save/edit/search works reliably.
- Wishlist purchased-to-inventory flow works reliably.
- Collections are safe to manage.
- Photos attach to correct items.
- Backup/export are trustworthy.
- Market Watch can produce visible useful discoveries.
- Licensing gates do not block export or destroy user trust.

## Non-goals

- Do not turn Cabinet into an escrow/payment app.
- Do not make selling the core beta promise.
- Do not require public identity to use Cabinet Free.
- Do not make Radicle/P2P a blocker for basic inventory and Market Watch value.
- Do not hide export behind paid lockout.
- Do not claim production integration reliability without validation evidence.

## Acceptance criteria for this plan

This plan is useful when:

- Cabinet has a clear Free/paid packaging model.
- Market Watch and Discoveries are treated as the immediate paid value spine.
- The beta can be validated with real users and real payment intent.
- The long-term trust architecture remains aligned without blocking the first sellable beta.
