# Pokemon App Capability Analysis Matrix

Updated: 2026-03-04
Scope apps:
- DexTCG (`https://dextcg.com/`)
- TCGCollector (`https://www.tcgcollector.com/`)
- PKMN.GG (`https://www.pkmn.gg/`)
- PriceCharting (`https://www.pricecharting.com/`)
- Rare Candy (`https://rarecandy.com/`, `https://get.rarecandy.com/`)

Scoring scale:
- 0 absent
- 1 minimal
- 2 usable
- 3 strong
- 4 best-in-class

## Evidence Notes (source snapshots)
- DexTCG marketing/features page and pricing tiers show scanner, notes, badges, widgets, marketplace coverage, language support, and premium split (Dex+).
- TCGCollector premium page shows large list limits, graded cards, notes, language per card, export, advanced pricing/statistics.
- PKMN.GG home/pro/help pages show deck builder, dynamic lists, value history, graded cards (manual value), TCGplayer market source, privacy toggle for value visibility.
- PriceCharting collection tracker page/FAQ/blog show free collection tracker, historic value, barcode/card scanner, photo support, folder organization (premium), grading recommendations, Android app rollout.
- Rare Candy landing/blog pages show AI-powered scanner, raw+graded+EN/JP coverage, set tracking, marketplace buy/sell/trade, mobile-first workflows, tally and social/community flows.

## Per-App Scorecards

### DexTCG
| Dimension | Score | Notes |
| --- | --- | --- |
| Onboarding/auth | 2 | Account onboarding and cross-platform sync visible. |
| Collection model | 3 | Variants, quantities, bookmarks, notes. Condition explicitly not first-class yet (help doc). |
| Price tracking | 2 | Multi-market price visibility, no deep historic stack shown publicly. |
| Discovery/search | 3 | Refined search filters by rarity/type/state/artist. |
| Portfolio analytics | 2 | Pokédex/challenges stats, but valuation analytics depth less explicit. |
| Media/scan | 3 | Camera scanner highlighted as primary flow. |
| AI/chat assist | 0 | No conversational assistant shown. |
| Integrations/export | 2 | Marketplace integrations visible; export unclear. |
| Mobile UX | 4 | Native-first emphasis, widgets/lockscreen, app icon/theme customization. |
| Collaboration | 1 | Mostly personal tracking. |

### TCGCollector
| Dimension | Score | Notes |
| --- | --- | --- |
| Onboarding/auth | 2 | Standard sign-in model. |
| Collection model | 3 | Sets/cards/lists, graded cards, language-level management (premium). |
| Price tracking | 2 | Basic/advanced pricing tiers. |
| Discovery/search | 3 | Set/card/pokedex exploration with in-collection/not-in-collection filters. |
| Portfolio analytics | 2 | Basic + advanced stats in premium. |
| Media/scan | 0 | Scanner/camera not emphasized. |
| AI/chat assist | 0 | Not present. |
| Integrations/export | 2 | Export is premium feature. |
| Mobile UX | 2 | Primarily web experience. |
| Collaboration | 1 | List sharing/use is limited relative to social apps. |

### PKMN.GG
| Dimension | Score | Notes |
| --- | --- | --- |
| Onboarding/auth | 3 | Web/app account with social profile progression. |
| Collection model | 4 | Variants, duplicates, dynamic/static lists, wishlist/trade binder patterns. |
| Price tracking | 3 | TCGplayer market pricing + value history + privacy controls. |
| Discovery/search | 4 | Full card db + strong filters + deck legality and build tools. |
| Portfolio analytics | 3 | Total estimated value + historical movement + profile-level visibility controls. |
| Media/scan | 1 | No dominant scanner-first flow; tracking is list/search centric. |
| AI/chat assist | 0 | Not present. |
| Integrations/export | 3 | Deck import/export and one-click purchase pathways highlighted. |
| Mobile UX | 3 | Works across devices with Pro feature unlocks. |
| Collaboration | 3 | Friend system + sharable lists/profile customization. |

### PriceCharting
| Dimension | Score | Notes |
| --- | --- | --- |
| Onboarding/auth | 2 | Account-based collection tracker. |
| Collection model | 3 | Quantities, grade/condition, photos, notes, folders (premium). |
| Price tracking | 4 | Historic collection value, price changes, grading recommendations. |
| Discovery/search | 3 | Broad database + scanner/barcode + integrated ownership indicators. |
| Portfolio analytics | 4 | Collection value over time + delta sorting and premium insights. |
| Media/scan | 3 | Barcode + card scanner + photo support. |
| AI/chat assist | 0 | Not present. |
| Integrations/export | 2 | Import and tooling present; marketplace/deal tooling mostly domain-specific. |
| Mobile UX | 3 | iOS and Android app support with scanner workflows. |
| Collaboration | 2 | Shared collection links and visibility. |

### Rare Candy
| Dimension | Score | Notes |
| --- | --- | --- |
| Onboarding/auth | 3 | Mobile-first onboarding across iOS/Android + web profile presence. |
| Collection model | 4 | Raw/graded, language-aware, set progress tracking, wants/trade support. |
| Price tracking | 3 | Price insight + running scanner tally + collection growth tracking. |
| Discovery/search | 3 | Set browsing, profile search/filter by rarity/type/language. |
| Portfolio analytics | 3 | Collection growth and set completion progress. |
| Media/scan | 4 | AI scanner positioned as core product behavior; high-volume scan claims. |
| AI/chat assist | 1 | AI used in scanner recognition; no broad assistant UX. |
| Integrations/export | 3 | Marketplace buy/sell/trade integrated in product. |
| Mobile UX | 4 | Strong mobile-first UX and real-time capture loop. |
| Collaboration | 4 | Feed/community/profile sharing + trading orientation. |

## Gap-to-Cabinet Mapping

### Must-have (P1)
1. Scanner confidence UX and multi-capture workflow parity (`POKEMON-COMP-001`)
2. Set completion and collection progress model with variants/language/graded overlays (`POKEMON-COMP-002`)
3. Multi-source pricing history with trend deltas and alert rules (`POKEMON-COMP-003`)
4. Marketplace-aware discovery handoff (buy links + seller trust metadata + stock signal) (`POKEMON-COMP-004`)

### Should-have (P2)
5. Shareable collection/list privacy controls (public/friends/private) (`POKEMON-COMP-005`)
6. Dynamic list templates (wishlist, trade binder, watch list) with saved filters (`POKEMON-COMP-006`)
7. Graded card workflow depth (slab metadata + valuation overrides + cert link quality checks) (`POKEMON-COMP-007`)

### Nice-to-have (P3)
8. Social/community layer (profile cards, progress sharing, compare mode) (`POKEMON-COMP-008`)
9. Gamified progress overlays for onboarding retention (badges/challenges) (`POKEMON-COMP-009`)
10. Deck-builder style query presets for collectible themes (collector goal bundles) (`POKEMON-COMP-010`)

## Priority Wave Plan
- P1 (delivery-critical): `POKEMON-COMP-001..004`
- P2 (high-value depth): `POKEMON-COMP-005..007`
- P3 (retention polish): `POKEMON-COMP-008..010`

## Cabinet Requirement Implications
- Integrations: stronger provider metadata and purchase-path quality.
- Inventory/Wishlist: explicit progress dimensions (set completion, variant completion, graded overlay).
- Pricing: richer history and alerting controls.
- Scanner: recognition confidence + batch ergonomics + deterministic fallback.
