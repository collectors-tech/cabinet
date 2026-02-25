# Clerk Billing Setup

## Purpose
Enable cloud ownership login and plan-based feature unlocks in Cabinet.

## Required Environment
- `VITE_CLERK_PUBLISHABLE_KEY=pk_...` in repo root `.env`

## Current Runtime Flow
1. User signs in through Clerk in the frontend shell.
2. Frontend calls `POST /api/auth/cloud/session/bootstrap` with Clerk session token.
3. Runtime reads token claims and returns plan + feature entitlement.
4. UI gates Pro features (`AI Assist`, `Price Tracking`, `Scanner Automation`) based on returned features.

## Plan Claim Mapping
- `plan` claim values:
  - `free` -> `collection_core`
  - `pro|paid|plus` -> `collection_core`, `ai_assist`, `price_tracking`, `scanner_automation`

## Notes
- Current bootstrap decodes JWT payload claims without signature verification.  
  This is suitable for dev wiring only.
- Production hardening is tracked in issue `#137` (webhook/signature-validated entitlement sync).
