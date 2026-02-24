# Cabinet – Product Specification

## Vision
Cabinet is a desktop-first collector intelligence system for serious hobby collectors.
It combines structured collection tracking with market discovery and price tracking.

## Core Principles
- Local-first (SQLite database)
- Desktop Go app with embedded Web UI
- Local authentication with WebAuthn (required in v1)
- Secure local secret storage (no plaintext API keys)
- Signed installers and signed updates
- Scanner-driven discovery ("Not in my collection")
- Price tracking & trend graphs
- Optional AI assist (user provides OpenAI API key)
- Closed-source with license model

## Target Persona
Serious niche collector (starting with AFX slot cars):
- 200–2000 items
- Tracks variants & condition
- Watches resale markets weekly
- Will install desktop software

## Core Modules
1. Collection Core (Items, Instances, Photos)
2. Scanner (eBay provider v1)
3. Wishlist & Tracked Items
4. Price Tracking
5. AI Metadata Assist (OpenAI)
6. In-App Chat Copilot (collection assistant)
7. Licensing System

## Full Feature Baseline
The complete v1 + planned feature inventory is maintained in:
- `docs/FULL_FEATURE_LIST.md`
- `docs/USE_CASES_AND_SCENARIOS.md`
- `docs/UI_ENDPOINT_PARITY.md`

This file (`SPEC.md`) defines product intent and scope boundaries.
`FULL_FEATURE_LIST.md` is the implementation-level feature checklist.
`USE_CASES_AND_SCENARIOS.md` is the behavior and workflow reference.

## Out of Scope (v1)
- P2P networking
- Built-in marketplace payments
- Cloud sync
- Multiple AI providers (future)

## Monetisation
- One-time Pro license (e.g., $79)
- Optional future scanner updates subscription
