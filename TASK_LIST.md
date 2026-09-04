# TASK_LIST

Reconciled on 2026-08-11 AEST from exact `origin/develop`
`f7853669e4586e6f50a2316cb8ca7082e019ce79`, live GitHub issues, GitHub
Project 2, OpenSpec, and isolated prepared-worktree evidence.

This file is the high-level Cabinet 0.1 private-beta execution summary.
Repository issues and Project 2 remain authoritative for detailed checklists,
State, Project Priority, and Order. Refresh those sources before acting when
this snapshot and GitHub differ.

## Ship Verdict

**Not yet shippable.** The merged source baseline now contains the
request-boundary, Browser Companion critical-path, dependency-security,
branch-protection, product-documentation, and packaged-acceptance-recorder
source work. The active Project 2 queue has no Ready cards; every remaining
release child is either Done or Blocked, while #1864 stays In progress as the
final evidence and approval parent. No exact Cabinet plus Browser Companion
candidate exists, and user-present provider proof, packaged acceptance,
same-candidate data safety, owner/legal decisions, and final approval remain
incomplete.

Earlier integrated dirty-worktree rehearsals remain compatibility evidence only.
They are not reviewed commits, hosted results, a release freeze, exact
candidates, or approval records.

## Completed Merged Evidence

- #2033 [DONE], #2035 [DONE], and #2032 [DONE]: Browser Companion pairing,
  data-driven modules, durable capture, and fail-closed protocol foundations.
- #2036 [DONE]: acyclic release and evidence graph.
- #2037 [DONE]: runtime/OpenAPI route parity gate.
- #2047 [DONE]: canonical Cabinet 0.1 capability and limitation disclosure.
- #2050 [DONE]: complete Go contract coverage restored in the release gates.
- #2051 [DONE]: production dependency security gate merged; default-branch
  critical/high Dependabot findings reconciled to zero.
- #2052 [DONE]: request-host, origin, LAN authorization, verified session, and
  entitlement boundaries merged.
- #2053 [DONE]: recursive diagnostic redaction merged.
- #2054 [DONE]: bounded provider HTTP calls and fail-closed timeout behavior
  merged.
- #2055 [DONE]: fixed, versioned beta Cypress source pack merged.
- #2056 [DONE]: branch-protection policy, read-only verifier, and exact #1864
  promotion approval guard merged; repository settings were later applied and
  verified compliant.
- #2057 [SOURCE MERGED / BLOCKED]: README, privacy, terms, Help Center, and
  generated release-note guidance are aligned with the private-beta product
  surface; owner/legal decisions and exact-candidate verification remain open.
- #2048 [DONE]: 51-row resumable packaged-acceptance evidence recorder merged.
- #2062 [DONE]: Add Integration hand-off, persistence, hosted Edge loading, and
  compact pairing accessibility merged.
- #2064 [DONE]: canonical Browser Companion capture-run success status merged.
- #2065 [DONE]: captured discovery review and Wishlist hand-off persistence
  merged.
- #2066 [DONE]: prior task-list reconciliation merged; this document is now the
  follow-up #1864 reconciliation from current `develop`.

These are evidence inputs, not executable READY work.

## Gate A - Source and Governance Work

The status below matches the live issue and Project state at reconciliation.
"Staged-only" means an explicitly authorized index exists but there is no
commit, push, PR, merge, or hosted result. "Local unstaged" means isolated
worktree evidence only. Neither state proves a merge or hosted gate.

| Issue | Project state | Current truth and remaining work |
| ----- | ------------- | -------------------------------- |
| #2057 | Blocked | Source/docs PR #2075 merged at `f7853669e4586e6f50a2316cb8ca7082e019ce79`. Owner/legal decisions and exact-candidate release-note/package verification remain open. |
| #2048 | Done | PR #2074 merged at `edbacd78c3bdf8469aeba43ce1d2a9d786ac0ace`; recorder source is available for #1869 but does not run or satisfy packaged acceptance by itself. |
| #2066 | Done | PR #2073 merged at `94311bd6be42fc5f23ab6e942588fa9b7ff4a66b`; this #1864 update supersedes that snapshot for the current baseline. |

Gate A source landing is complete for the currently merged pre-candidate source
set, but the release programme is still blocked until the human/legal and
candidate-bound rows below are satisfied. Any new P0 release blocker or evidence
correction must land through its own focused issue and hosted checks before
candidate nomination.

## Gate B - User-Present Live Provider Proof

- #1944 [BLOCKED]: run a genuine Frontline user-present live proof through the
  Browser Companion and persist source/module/schema provenance, review state,
  and confirmed hand-off.
- #1945 [BLOCKED]: run the equivalent genuine Bonza user-present live proof after
  normal browser interaction, including truthful Sucuri/challenge handling.
- #1929 [BLOCKED]: complete the four-provider source/live checklist only after
  both live child issues pass.

Fixture tests, mocked responses, extension-load checks, screenshots, or dirty
rehearsal evidence do not satisfy #1944 or #1945. Do not export cookies,
automate login, bypass challenges, or use hidden crawling. The recorded merged
baseline now contains #2052, #2062, #2064, and #2065, but only genuine
user-present live proof can satisfy these provider gates. Candidate nomination
still waits for all Gate A source and hosted evidence.

## Gate C - Freeze and Exact Private/Internal Candidate

1. Nominate one exact clean `develop` commit after Gates A and B pass.
2. Start the temporary release freeze. Any accepted blocker fix or evidence
   correction invalidates the older candidate and requires a new exact commit.
3. #1868 [BLOCKED]: run the non-publishing exact private/internal candidate
   workflow and retain its commit-bound validation evidence.
4. #2034 [BLOCKED]: retain the exact Chrome and Edge packages, versions,
   manifests, checksums, install identity, and pairing evidence from that same
   commit.

Internal candidate creation does not require final #1864 approval. It does not
publish a release or promote `develop` to `main`.

## Gate D - Exact Packaged Windows Acceptance

- #2048 must already be merged with the final acceptance inventory.
- #1869 [BLOCKED]: run exact packaged Windows acceptance against the Gate C
  Cabinet and Browser Companion files, not a source runtime or rebuilt files.
- Cover onboarding, inventory, media, Wishlist/Collections, all four provider
  journeys, review hand-off, restart/persistence, profile isolation, failure
  recovery, versions, manifests, and checksums.
- Record pass, focused blocker, or not-run for every row. A release-blocking fix
  returns the programme to Gate C with a new candidate.

## Gate E - Same-Candidate Data Safety

- #1867 [BLOCKED]: prove prior-data upgrade, backup, export, restore, media
  relocation, relationship/link preservation, and zero-data-loss recovery
  against the same exact candidate accepted in Gate D.

Source tests are prerequisites but do not replace same-candidate data safety.

## Gate F - Legal Decisions and Final Approval

- #2057 [IN PROGRESS]: obtain the owner/legal decisions for entity/controller
  and contact details, effective date and acceptance, governing law and
  consumer/age/licence/IP terms, retention/processors, operator
  responsibilities, and support/SLA wording; verify the final guidance in the
  exact candidate.
- #1864 [IN PROGRESS]: review the complete same-candidate evidence set, known
  limitations, checksums, and focused blockers. External prerelease publication
  or `develop` to `main` promotion requires trusted explicit #1864 approval for
  that exact commit.

Owner/legal decisions and final approval are distinct from source landing,
candidate creation, packaged acceptance, and recovery proof.

## Work the Dev Agent Can Continue

Until operator-present work is available, the development agent can:

1. prepare and review the remaining Gate A patches in the recorded order;
2. keep each remaining issue manifest isolated and unstaged until separately
   authorized;
3. re-run focused source contracts after each reconciliation;
4. keep Project 2 and issue checklists truthful without checking hosted or human
   rows from local evidence; and
5. prepare exact commands and evidence templates for Gates B-E without claiming
   those gates passed.

The dev agent cannot substitute for #1944/#1945 user-present live proof, #1869
packaged acceptance, #1867 same-candidate recovery proof, legal decisions
requiring owner authority, or #1864 approval. Branch-protection source and
settings evidence is complete as of #2056, but future drift must still be
verified before final promotion.

## Validation Expectations

At source landing and again for the exact candidate, use the repository-owned
gates rather than copying volatile pass counts here:

- strict OpenSpec validation;
- complete Go repository tests and OpenAPI parity;
- clean UI dependency install, dependency-security gate, and production build;
- fixed beta Cypress source pack;
- Browser Companion protocol/package contracts; and
- Cabinet release-package contracts and exact manifest/checksum verifiers.

No local rehearsal, documentation update, or green source test authorizes
release publication or `develop` to `main` promotion.

## Post-Beta Guardrail

Unless packaged acceptance exposes a direct blocker, defer broad Agent/Telegram
coverage, provider expansion, eBay seller command-centre breadth, Metadata
Studio/parity expansion, broad table/UI refactors, community governance,
payments, SBOM/attestation expansion, and OpenSpec archive hygiene until after
Cabinet 0.1.
