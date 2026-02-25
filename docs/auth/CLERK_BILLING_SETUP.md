# Clerk Billing Setup

## Purpose
Enable cloud ownership login and plan-based feature unlocks in Cabinet.

## Required Environment
- `VITE_CLERK_PUBLISHABLE_KEY=pk_...` in repo root `.env`
- `CABINET_CLERK_WEBHOOK_SECRET=...` in runtime environment for webhook signature verification

## Current Runtime Flow
1. User signs in through Clerk in the frontend shell.
2. Frontend calls `POST /api/auth/cloud/session/bootstrap` with Clerk session token.
3. Runtime reads token claims and returns plan + feature entitlement.
4. UI gates Pro features (`AI Assist`, `Price Tracking`, `Scanner Automation`) based on returned features.
5. Billing webhooks update server-side entitlement overrides immediately.

## Clerk Webhook Configuration
1. In Clerk Dashboard, create webhook endpoint:
   - URL: `http://<cabinet-host>/api/auth/cloud/clerk/webhook`
   - Method: `POST`
2. Set Cabinet runtime secret:
   - `CABINET_CLERK_WEBHOOK_SECRET=<same-signing-secret>`
3. Configure events that can change plan:
   - Subscription created/updated/canceled (or equivalent billing lifecycle events)
4. Cabinet signature header:
   - `X-Cabinet-Webhook-Signature` as HMAC-SHA256 (base64url) of raw request body.
5. Entitlement behavior:
   - Upgrade/downgrade events overwrite runtime plan for the Clerk `user_id`
   - Next bootstrap call returns updated `plan` + `features` without requiring token claim refresh.

## Plan Claim Mapping
- `plan` claim values:
  - `free` -> `collection_core`
  - `pro|paid|plus` -> `collection_core`, `ai_assist`, `price_tracking`, `scanner_automation`

## Notes
- Current bootstrap decodes JWT payload claims without signature verification.  
  This is suitable for dev wiring only.
- Production hardening is tracked in issue `#137` (webhook/signature-validated entitlement sync).
