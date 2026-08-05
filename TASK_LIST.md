# TASK_LIST

Last reconciled from live GitHub/OpenSpec state: 2026-08-06 +10:00

This file is the Cabinet 0.1 private beta execution summary. Historical issue
snapshots from April 2026 have been retired from this top-level queue; use
GitHub issues and OpenSpec change/task files for detailed execution history.

## Status Legend

- `OPEN`: issue is open and needs focused implementation, validation, release evidence, or governance work.
- `IN REVIEW`: issue implementation evidence is merged, but release-lane validation or final review still owns closure.
- `DONE`: issue is closed with linked evidence on GitHub.
- `POST-BETA`: issue/epic is intentionally outside Cabinet 0.1 private beta unless a release test exposes it as a direct blocker.
- Work is executed one focused issue branch at a time, validated, merged into `develop`, then demo/release lanes are rebuilt from `develop`.

## Cabinet 0.1 Private Beta Critical Path

Parent release epic: #1864 `epic(beta): ship Cabinet 0.1 private beta`

Completed foundations: #1865, #1866, #1870, #1871, #1872, #1936 and #1937.

1. [x] #2037 [DONE] Enforce the complete runtime/OpenAPI route parity gate.
   - Evidence: PR #2038 and Develop Quality Gate run `31018831485` passed on the exact PR head.
2. [ ] #2036 [IN REVIEW] Keep the release dependency graph, evidence and approval sequence acyclic.
3. [ ] #2033 [OPEN] Replace predictable companion access with secure pairing/session transport.
4. [ ] #2035 [OPEN] Build the shared Chromium MV3 host, configuration and readiness UI.
5. [ ] #2032 [OPEN] Persist idempotent provider item/image observations through Cabinet.
6. [ ] #1944 and #1945 [OPEN] Prove Frontline and Bonza browser-assisted live source readiness; #1943 Hobbytech is already source/live ready.
7. [ ] #2034 [OPEN] Produce verified Chrome/Edge packages, checksums and a release manifest.
8. [ ] #1868 [IN REVIEW] Build the exact private/internal candidate for Cabinet + Browser Companion.
   - Internal candidate creation is allowed before final approval; external prerelease publication is not.
9. [ ] #1869 [IN PROGRESS] Run packaged collector, four-provider, companion and media acceptance.
10. [ ] #1867 [IN REVIEW] Attach exact-candidate upgrade, backup, export, restore and relocation evidence.
11. [ ] #1864 [IN PROGRESS] Decide final approval after all exact-candidate evidence is linked.
   - Only explicit final approval permits external prerelease publication and `develop` to `main` promotion.

## Current Release Governance Work

- [x] #1872 reconciled every unchecked OpenSpec task into completed/archive evidence, linked focused issue, or explicit deferred/post-beta reason.
- [x] #2036 defines provider source-ready separately from packaged evidence and removes the candidate/approval cycle.
- [ ] Keep labels/status aligned after each live issue is verified.
- [ ] Keep #1864 as the single private beta release summary, final evidence parent and approval decision.
  - Current repo evidence index: `openspec/migration/beta-release-evidence-index.md`.
- [ ] Start the temporary release freeze after every source/live-ready prerequisite is merged.
  - During the freeze, accept only a P0 release blocker, directly required test/evidence repair or release-document correction through a focused issue and green PR.
  - Every accepted fix creates a new exact candidate commit and invalidates older candidate acceptance.
- [ ] Do not publish externally or merge `develop` into `main` until Max explicitly approves the tested release candidate.

## Active OpenSpec Change Inventory

Captured with `openspec list` on 2026-07-11. Updated in #1872 to archive completed Agent-channel, Agent-skill registry, inventory grading, runtime startup, parity route-contract, inventory accessibility selector, and Bonza URL changes, then reconcile the final active changes to explicit closed/deferred issue holders.

- `finalize-onboarding-and-collector-ux`: 14/14 tasks reconciled. Product remainder is post-beta and tracked by #1889 unless #1864/#1869 release acceptance exposes one task as a beta blocker.
- `complete-screen-api-parity-audits`: 15/15 tasks reconciled. Original audit holders #143, #144, and #145 are closed; renewed parity failures should be filed from #1869 evidence as focused issues.
- `stabilize-inventory-runtime-regressions`: 11/11 tasks reconciled. Legacy inventory non-500 holder #149 is closed; remaining concrete runtime/startup regressions are deferred to #1890 unless release acceptance exposes a beta blocker.
Archived during #1872 reconciliation:

- `ingest-bonza-product-urls`: archived into canonical provider-family, AU-webshop, and inventory specs after reconciling closed #811 live verification and #1077 QA evidence.
- `define-universal-agent-channel-contracts`: archived into `openspec/specs/agent-universal-channels/spec.md`; broad Agent/Telegram implementation remains post-beta unless #1864 acceptance exposes a release blocker.
- `define-agent-skill-registry`: archived into `openspec/specs/agent-skills-registry/spec.md`; Agent skill registry product breadth remains post-beta unless #1864 acceptance exposes a release blocker.
- `inventory-item-type-condition-scales`: archived after confirming its requirements already exist in canonical inventory specs and traceability rows.
- `harden-runtime-single-endpoint-startup`: archived after confirming its requirements already exist in canonical runtime specs and traceability rows.
- `stabilize-startup-nfr-gates`: archived into `openspec/specs/fresh-runtime-startup/spec.md`; #446 and #448 are closed with PR/demo evidence and package-level release acceptance remains tracked by #1869.
- `stabilize-parity-components-route-contracts`: archived into `openspec/specs/parity-components-route-contracts/spec.md`; broader packaged release acceptance remains tracked by #1869.
- `stabilize-inventory-a11y-keyboard-selectors`: archived into `openspec/specs/inventory-a11y-keyboard-selectors/spec.md`; broader packaged release acceptance remains tracked by #1869.

## Post-Beta Scope Guardrails

The following broad work remains visible but is outside Cabinet 0.1 private beta unless release testing proves a direct blocker:

- Metadata Studio breadth and Homebox/iCollect parity.
- Public identity, signed receipts, signed feedback, reputation, Git/Radicle/P2P ledgers, and community governance.
- Telegram as a general Agent channel and universal Agent entry points across every surface.
- eBay seller/listing/fulfilment command-centre scope.
- Broad retailer/provider expansion beyond required beta providers Voglers, Hobbytech, Frontline and Bonza.
- Store/venue nodes, event infrastructure, escrow, or payment processing.

## Validation Expectations

- OpenSpec changes: `openspec validate --all --strict --no-interactive`
- Backend/runtime changes: targeted `go test` package(s), then broader gates when merging.
- UI changes: targeted Cypress spec(s) through `cypress.ps1`, with live persistence/state verification for mutating flows.
- Release candidate: exact commit gate, packaged Windows acceptance evidence, checksums, release notes, and demo/runtime evidence from `develop`.
