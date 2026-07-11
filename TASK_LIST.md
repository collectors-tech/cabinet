# TASK_LIST

Last reconciled from live GitHub/OpenSpec state: 2026-07-11 13:48 +10:00

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

1. [x] #1865 [DONE] Add develop CI and release-candidate quality gates.
   - Evidence: PR #1879 merged to `develop`; #1864 contains Develop Quality Gate and demo2 runtime evidence.
2. [x] #1866 [DONE] Replace simulated local sign-in and passkey flows.
3. [ ] #1867 [IN REVIEW] Validate database upgrade, backup, export and restore round trip.
   - Source-level data-safety proof is summarized in `openspec/migration/beta-data-safety-evidence-matrix.md`; packaged Windows evidence remains owned by #1868/#1869 before beta release closure.
4. [x] #1870 [DONE] Align Free, Plus and Pro entitlements and user-visible plan state.
5. [x] #1871 [DONE] Prove one live Market Watch provider and fail closed for others.
   - Evidence: PR #1908 merged to `develop`; #1864 contains the Voglers live Market Watch provider proof summary.
6. [ ] #1868 [OPEN] Package versioned Windows beta artefact and GitHub release.
7. [ ] #1869 [OPEN] Run packaged core-workflow release acceptance suite.
8. [x] #1872 [DONE] Reconcile OpenSpec active changes and stale release tracking.
   - Evidence: PR #1891 merged to `develop`; all active OpenSpec tasks reconciled to archive, closed issue evidence, or explicit deferred follow-ups.

## Current Release Governance Work

- [x] #1872: reconciled every unchecked OpenSpec task into completed/archive evidence, linked focused issue, or explicit deferred/post-beta reason.
- [ ] Correct stale labels/status on release issues after each live issue is verified.
- [ ] Keep #1864 as the single private beta release summary and evidence index.
- [ ] Do not merge `develop` into `main` until Max explicitly approves the tested release candidate.

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
- Broad retailer/provider expansion beyond the one proven beta Market Watch provider path.
- Store/venue nodes, event infrastructure, escrow, or payment processing.

## Validation Expectations

- OpenSpec changes: `openspec validate --all --strict --no-interactive`
- Backend/runtime changes: targeted `go test` package(s), then broader gates when merging.
- UI changes: targeted Cypress spec(s) through `cypress.ps1`, with live persistence/state verification for mutating flows.
- Release candidate: exact commit gate, packaged Windows acceptance evidence, checksums, release notes, and demo/runtime evidence from `develop`.
