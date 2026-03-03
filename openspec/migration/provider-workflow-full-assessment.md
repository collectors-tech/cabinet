# Provider Workflow Full Implementation Assessment (Refined)

## Purpose
Drive production-grade completion for each integration provider using one consistent workflow contract and proof gates.

## Providers in Scope
1. Bonza / Default Site Search
2. eBay
3. Amazon
4. AU Webshops Pack
5. OpenAI Assistant Lane

---

## Standard Workflow Contract (required for each provider)
A provider is only complete when all stages below are implemented and proven:

1) Connect/Auth
2) Validate Token/Credentials
3) Search/Query
4) Candidate Mapping (normalized schema)
5) Apply/Import (confirm-before-apply where relevant)
6) Sync/Refresh
7) Deterministic Error/Retry/Rate-limit behavior
8) Mock lane (deterministic CI/dev)
9) Live lane (real provider verification)
10) User setup docs (step-by-step)

---

## Definition of Done (binary)
For each provider issue:
- [ ] All 10 stages complete
- [ ] Traceability IDs moved to `implemented` only with executable evidence
- [ ] Cypress/API tests pass
- [ ] `openspec validate --all` passes
- [ ] Setup + troubleshooting docs published
- [ ] Issue evidence comment includes commit hash and test links

No partial close-outs.

---

## Current Assessment Snapshot

| Provider | Current State | Primary Gaps | Priority |
|---|---|---|---|
| Bonza/default-site-search | Partial | Saved searches, selective import depth, mock/live parity hardening | P1 |
| eBay | Partial | Full stage parity + deterministic error matrix | P1 |
| Amazon | Partial | Full stage parity + apply/import proof + docs | P1 |
| AU webshops pack | Partial | Config-driven domain source + parity + proofs | P1 |
| OpenAI assistant lane | Partial | Chat create/update confirm flow + mobile image + structured tool proofs | P1 |

---

## Critical Risks to Track
1) Provider auth/token expiry and inconsistent scope permissions
2) Non-deterministic mapping between provider payloads and candidate schema
3) Hidden background sync behavior causing unexpected imports
4) Hardcoded domain/provider source configuration
5) Legal/compliance constraints for source access and data reuse

---

## Execution Order
1. Bonza/default-site-search
2. eBay
3. Amazon
4. AU webshops pack
5. OpenAI assistant lane

Reason: maximize value fast while reducing highest integration uncertainty first.

---

## Required Artifacts Per Provider
- `openspec/migration/provider-<name>-workflow-summary.md`
- `openspec/migration/provider-<name>-workflow-changed-files.txt`
- Test evidence links (Cypress/API)
- Issue evidence comment with verified pushed commit hash

---

## Orchestration Linkage
- Meta issue: **Provider workflow full review + delivery orchestration**
- One execution issue per provider (already created)
- Backlog priority labels govern next selection

---

## Notes
This assessment is governance + execution control. It is not a close-out report.
Use it to ensure every provider reaches the same production-grade bar.
