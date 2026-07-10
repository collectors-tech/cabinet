# TASK_LIST

Last reconciled from live GitHub/OpenSpec state: 2026-07-11 05:55 +10:00

This file is the Cabinet 0.1 private beta execution summary. Historical issue
snapshots from April 2026 have been retired from this top-level queue; use
GitHub issues and OpenSpec change/task files for detailed execution history.

## Status Legend

- `OPEN`: issue is open and needs focused implementation, validation, release evidence, or governance work.
- `DONE`: issue is closed with linked evidence on GitHub.
- `POST-BETA`: issue/epic is intentionally outside Cabinet 0.1 private beta unless a release test exposes it as a direct blocker.
- Work is executed one focused issue branch at a time, validated, merged into `develop`, then demo/release lanes are rebuilt from `develop`.

## Cabinet 0.1 Private Beta Critical Path

Parent release epic: #1864 `epic(beta): ship Cabinet 0.1 private beta`

1. [x] #1865 [DONE] Add develop CI and release-candidate quality gates.
   - Evidence: PR #1879 merged to `develop`; #1864 contains Develop Quality Gate and demo2 runtime evidence.
2. [x] #1866 [DONE] Replace simulated local sign-in and passkey flows.
3. [ ] #1867 [OPEN] Validate database upgrade, backup, export and restore round trip.
4. [x] #1870 [DONE] Align Free, Plus and Pro entitlements and user-visible plan state.
5. [ ] #1871 [OPEN] Prove one live Market Watch provider and fail closed for others.
6. [ ] #1868 [OPEN] Package versioned Windows beta artefact and GitHub release.
7. [ ] #1869 [OPEN] Run packaged core-workflow release acceptance suite.
8. [ ] #1872 [OPEN] Reconcile OpenSpec active changes and stale release tracking.

## Current Release Governance Work

- [ ] #1872: reconcile every unchecked OpenSpec task into one of:
  - completed and archive-ready with validation evidence
  - linked to an open focused issue
  - explicitly deferred/post-beta with reason
- [ ] Correct stale labels/status on release issues after each live issue is verified.
- [ ] Keep #1864 as the single private beta release summary and evidence index.
- [ ] Do not merge `develop` into `main` until Max explicitly approves the tested release candidate.

## Active OpenSpec Change Inventory

Captured with `openspec list` on 2026-07-11:

- `ingest-bonza-product-urls`: 21/24 tasks complete. Remainder is live/manual provider verification, optional HTML fallback, and final #811 delivery cleanup.
- `stabilize-startup-nfr-gates`: 7/9 tasks complete. Remainder is blocked validation-chain rerun and issue/PR closure evidence.
- `stabilize-parity-components-route-contracts`: 7/8 tasks complete. Remainder is feeding targeted validation back into the broader regression gate.
- `stabilize-inventory-a11y-keyboard-selectors`: 4/5 tasks complete. Remainder is feeding targeted validation back into the broader regression gate.
- `finalize-onboarding-and-collector-ux`: 0/14 tasks complete. Keep as post-beta unless #1864 release acceptance exposes one task as a beta blocker.
- `complete-screen-api-parity-audits`: 0/15 tasks complete. Keep as post-beta unless #1864 release acceptance exposes one task as a beta blocker.
- `stabilize-inventory-runtime-regressions`: 0/11 tasks complete. Triage against current release smoke evidence before treating old startup/runtime subtasks as active beta blockers.
- `define-universal-agent-channel-contracts`: complete, archive candidate after confirming linked Agent issues remain post-beta.
- `define-agent-skill-registry`: complete, archive candidate after confirming linked skill registry issues remain post-beta.
- `inventory-item-type-condition-scales`: complete, archive candidate after final validation evidence is linked.
- `harden-runtime-single-endpoint-startup`: complete, archive candidate after final validation evidence is linked.

## Post-Beta Scope Guardrails

The following broad work remains visible but is outside Cabinet 0.1 private beta unless release testing proves a direct blocker:

- Metadata Studio breadth and Homebox/iCollect parity.
- Public identity, signed receipts, signed feedback, reputation, Git/Radicle/P2P ledgers, and community governance.
- Telegram as a general Agent channel and universal Agent entry points across every surface.
- eBay seller/listing/fulfilment command-centre scope.
- Broad retailer/provider expansion beyond the one proven beta Market Watch provider path.
- Store/venue nodes, event infrastructure, escrow, or payment processing.

## Validation Expectations

- OpenSpec changes: `openspec validate --all --strict --no-interactive`
- Backend/runtime changes: targeted `go test` package(s), then broader gates when merging.
- UI changes: targeted Cypress spec(s) through `cypress.ps1`, with live persistence/state verification for mutating flows.
- Release candidate: exact commit gate, packaged Windows acceptance evidence, checksums, release notes, and demo/runtime evidence from `develop`.
