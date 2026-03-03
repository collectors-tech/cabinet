# Clerk Integration + Permissions Test Matrix (Cabinet)

## Objective
Validate Clerk-based auth and plan-gated permissions using multiple seeded test accounts.

## Required test accounts
- `mvp_user@example.com` -> Plan: MVP
- `creator_user@example.com` -> Plan: Creator
- `teams_owner@example.com` -> Plan: Teams
- `teams_member@example.com` -> Plan: Teams (non-owner role)
- `admin@example.com` -> Admin/ops override (if supported)

## Feature gate checklist
For each account level, verify:
1. Add User gating
2. Premium/Teams-only actions (review/team workflows)
3. API token and integration management gates
4. Webhooks/advanced capabilities
5. Chat/AI mutation permissions (confirm-before-apply)

## Validation lanes
- UI Cypress lane per account profile
- API contract lane with auth context
- Effective permission diagnostics snapshot

## Pass criteria
- Ineligible accounts blocked with deterministic error and upgrade guidance
- Eligible accounts can complete gated flows
- No privilege leakage across account switches
