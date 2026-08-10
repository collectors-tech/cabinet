# TASK_LIST

Last reconciled from clean `develop` commit
`d676ac02cdbc466402ba38f14a741292f9165320`, live GitHub state, OpenSpec,
source inspection, and local validation on 2026-08-10 AEST.

This file is the high-level Cabinet 0.1 private-beta execution summary.
Repository issues are the detailed source of truth. Historical implementation
evidence remains in GitHub, OpenSpec migration notes, and traceability files.

## Ship Verdict

**Not yet shippable.** Cabinet has broad source implementation, but the exact
Windows private-beta candidate has not been created or accepted. The current
release lane is also blocked by confirmed source-gate, security, dependency,
provider-timeout, and candidate-pack gaps.

Do not start the release freeze or run #1868 until every P0 preflight issue in
this file is merged and revalidated and the Frontline/Bonza live-source rows are
complete.

## Verified Product State

### Implemented with merged source evidence

- Desktop-first Go/SQLite runtime with embedded React UI and local MCP launcher.
- Inventory, Wishlist, Collections, media, search/matching, Discoveries,
  Market Watch, import/export, backup/restore, settings, local/ZITADEL modes,
  Assistant/Agent preview surfaces, and deployment/runtime foundations.
- Browser Companion secure pairing, rotation/revocation, profile scoping,
  exact optional origins, durable/idempotent item and media capture,
  fail-closed protected-provider handling, and deterministic Chrome/Edge
  package controls (#2033, #2035, and #2032).
- Source-level upgrade, restore, export, relationship, media, saved-filter, and
  secret-exclusion data-safety tests.
- Canonical Cabinet 0.1 capability/limitation disclosure in UI and generated
  release notes (#2047, PR #2049).
- Acyclic release/evidence graph (#2036 [DONE]) and complete runtime/OpenAPI
  route parity gate (#2037 [DONE]).

### Verified evidence snapshot

- PR #2049: all seven Develop Quality Gate jobs passed.
- `go run ./cmd/openapi-parity-gate`: passed during the 2026-08-10 audit.
- `npx --yes @fission-ai/openspec@latest validate --all --strict --no-interactive`:
  138 passed, 0 failed.
- `npm run test:release-package`: 14 passed, 0 failed.
- `npm run test:companion-package`: 6 passed, 0 failed.
- `npm run build` in `ui.web`: passed; two large-chunk warnings remain.
- Test inventory: 192 Go test files, 127 Cypress specs, plus extension and Node
  contract suites.
- All 17 remaining OpenSpec change folders report complete tasks; they are
  archive hygiene, not unfinished beta features.

### Confirmed red or unproven evidence

- `go test ./... -count=1` fails three root contract tests; #2050 owns the
  repairs and the missing Develop-gate coverage.
- `npm audit --omit=dev` reports 1 critical and 6 high production package
  families; GitHub reports 45 open Dependabot alerts, including 1 critical and
  21 high; #2051 owns remediation and adjudication.
- Local/LAN/cloud request boundaries, loopback origin protection, pseudo-login
  credentials, and unchecked entitlement claims require P0 remediation (#2052).
- Diagnostic redaction is not complete before persistence/export/remote send
  (#2053).
- Beta provider HTTP calls can hang without bounded timeouts (#2054).
- The candidate workflow can default to an under-scoped Cypress spec instead of
  a fixed beta core pack (#2055).
- Frontline and Bonza still lack genuine user-present live-source evidence
  (#1944/#1945).
- No exact Cabinet + Browser Companion beta candidate, packaged acceptance pack,
  same-candidate recovery evidence, or beta prerelease exists.
- `develop` is 1,766 commits ahead of `main`; promotion remains prohibited until
  exact-candidate approval.

## Cabinet 0.1 Private-Beta Critical Path

Parent release epic: #1864 `epic(beta): ship Cabinet 0.1 private beta`.

### Gate A - P0 source, security, and release preflight

These issues are executable dev-agent work and may run in parallel. #2050 is the
first handoff because it restores the complete gate that the other work must
pass.

1. [ ] #2050 [READY] Restore `go test ./...`, repair the three red root
   contracts, and make Develop Quality Gate include the root `tests` package.
2. [ ] #2051 [READY] Remove or adjudicate all critical/high production
   dependency vulnerabilities and add repeatable dependency security gating.
3. [ ] #2052 [READY] Enforce local/LAN/cloud authentication and request
   boundaries, loopback origin/CSRF protection, truthful setup responses, and
   verified entitlement claims.
4. [ ] #2053 [READY] Redact credentials, cookies, tokens, secrets, private
   content, session identifiers, and sensitive paths before diagnostics are
   persisted, exported, or sent remotely.
5. [ ] #2054 [READY] Give every beta provider request bounded timeouts and
   deterministic fail-closed behavior without cross-provider corruption.
6. [ ] #2055 [READY] Bind the exact candidate workflow to a fixed, versioned
   beta core Cypress pack and reject under-scoped dispatch.

Gate A exit criteria:

- all six issues are merged through green focused PRs;
- `go test ./... -count=1`, strict OpenSpec, OpenAPI parity, production UI
  build, release/companion package contracts, dependency security gate, and the
  fixed source-level Cypress release pack pass on one clean `develop` commit;
- no unresolved candidate-blocking security or data-loss issue remains.

### Gate B - Human live-provider source proof

These rows require normal user-present browser/operator interaction. They are
not substitutable with fixtures, CI extension loading, screenshots, hidden
crawling, cookie export, or challenge solving.

1. [ ] #1944 [BLOCKED] A real Frontline search persists candidates with
   transport/module/schema provenance and completes confirmed hand-off.
2. [ ] #1945 [BLOCKED] A real Bonza search after normal browser interaction
   persists candidates with provenance and completes confirmed hand-off.
3. [ ] #1929 [BLOCKED] Mark the four-provider source/live checklist ready only
   after Frontline and Bonza evidence passes. Voglers and Hobbytech source proof
   is already complete.

Gate B can proceed while dev agents execute Gate A.

### Gate C - Freeze and exact private/internal candidate

1. [ ] Nominate one exact clean `develop` commit after Gates A and B pass.
2. [ ] Start the temporary release freeze. Accept only a focused P0 blocker,
   required evidence repair, or release-document correction through a green PR.
   Any accepted change invalidates older candidate evidence.
3. [ ] #1868 [BLOCKED] Run the non-publishing exact candidate workflow.
4. [ ] #2034 [BLOCKED] Retain the exact Chrome and Edge packages, versions,
   manifests, checksums, install identity, and pairing evidence from the same
   commit.

Internal candidate creation does not require final #1864 approval. External
publication and `develop` to `main` promotion do.

### Gate D - Exact packaged Windows acceptance

1. [ ] #2048 [READY] Build the resumable, fail-closed evidence recorder before
   the candidate run. It must never auto-pass human/browser/provider steps.
2. [ ] #1869 [BLOCKED] Run the complete packaged collector, provider,
   companion, media, isolation, restart, error, and recovery checklist against
   the exact Gate C files.
3. [ ] Re-run Voglers, Hobbytech, Frontline, and Bonza packaged journeys with
   exact versions, provenance, hand-off, idempotency, and failure isolation.
4. [ ] Record pass, fail with focused blockers, or not-run for every checklist
   row. A release-blocking fix returns the programme to Gate C with a new commit.

### Gate E - Data safety, decision, and publication

1. [ ] #1867 [BLOCKED] Attach same-candidate database upgrade, backup, export,
   restore, media relocation, manifest/link, and zero-data-loss evidence.
2. [ ] #2056 [READY/P1] Enforce required develop checks and approval-controlled
   main promotion before final release approval.
3. [ ] #2057 [READY/P1] Align README, privacy, terms, Help Center, data-path,
   diagnostics, Browser Companion, and support claims with verified behavior.
4. [ ] #1864 [IN PROGRESS] Review all exact-candidate evidence and known
   limitations. Only Max's trusted exact-commit approval may publish the beta
   prerelease or promote `develop` to `main`.

## Dev-Agent Allocation Before the Operator Critical Path

Recommended order:

1. **Primary dev agent: #2050.** The release-candidate gate is currently
   guaranteed to fail and the develop PR gate is blind to the root contracts.
2. **Security/dependency lane: #2051 and #2052.** These are independent of live
   provider proof and must be resolved before a candidate can be trusted.
3. **Reliability lane: #2053 and #2054.** Harden diagnostic data handling and
   provider failure behavior while the operator gathers live proof.
4. **Release-tooling lane: #2055, then #2048.** Fix the source release pack,
   then build the resumable packaged-evidence recorder.
5. **Release-governance/docs lane: #2056 and #2057.** Complete after security
   behavior stabilizes and before final approval.

The dev agent cannot complete #1944/#1945 live provider proof, #1869 packaged
Windows acceptance, #1867 same-candidate recovery proof, or #1864 approval.

## Backlog Cleanup Applied on 2026-08-10

- Closed #2047 after PR #2049 merge, seven green PR jobs, focused source
  revalidation, checked acceptance criteria, and #1864 evidence linkage.
- Created #2050-#2057 as focused owners for newly verified shippability gaps.
- Added missing release issues to the Cabinet project and aligned active
  statuses: #1864 is the only programme item in progress; executable work is
  Ready; operator/candidate-dependent work is Blocked.
- Paused post-beta #1701 Agent breadth and moved #1939 UI table refactoring to
  deferred Backlog.
- Closed empty legacy milestones M1-M4 and created the `0.1 private beta`
  milestone for the active release chain.
- Preserved #1943 and #2034 as open/blocked because their source work is merged
  but exact packaged evidence is still missing.

## Post-Beta Scope Guardrails

Unless Gate D exposes a direct beta blocker, keep these outside Cabinet 0.1:

- Metadata Studio breadth and Homebox/iCollect parity.
- Universal Agent/Telegram entry points and live Telegram acceptance.
- eBay seller/listing/fulfilment command-centre breadth.
- Broad retailer/provider expansion, Shopify/generic crawling, and Hobbytech
  Parts Finder breadth beyond the required beta journey.
- Inventory bulk/lightbox breadth, dashboard attention-centre breadth, and
  shared table/action refactoring.
- Public identity, signed receipts/feedback, reputation, Git/Radicle/P2P
  ledgers, community governance, venues/events, escrow, and payments.

Post-beta quality hygiene also includes archiving the 17 completed OpenSpec
change folders, classifying 13 intentionally skipped Purchases Cypress tests,
adding coverage trends/thresholds, and considering SBOM/attestation and broader
CodeQL/gosec/govulncheck automation. These do not waive any P0 gate above.

## Validation Expectations

- OpenSpec: `npx --yes @fission-ai/openspec@latest validate --all --strict --no-interactive`
- Full Go source and contracts: `go test ./... -count=1`
- OpenAPI/runtime parity: `go run ./cmd/openapi-parity-gate`
- UI: production build plus the fixed beta Cypress source pack.
- Release packages: Cabinet and Browser Companion contract/verifier suites,
  exact manifests, versions, separate SHA-256 values, and combined candidate
  identity.
- Packaged release: #2048/#1869 evidence against the exact Windows files, then
  #1867 recovery evidence and #1864 approval.
