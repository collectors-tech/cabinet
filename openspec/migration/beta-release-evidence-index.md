# Cabinet 0.1 beta release evidence index

Issue: #1864 `epic(beta): ship Cabinet 0.1 private beta`
Status as of 2026-07-12: source-level beta critical-path evidence is mostly merged on `develop`; final release readiness remains gated on approved packaged Windows acceptance and release publication.

This index is the #1864 release-governance handoff point. It links the merged child-issue evidence without claiming private beta completion before the packaged candidate is approved and exercised.

## Child issue state

| Issue | Current release evidence | Release state |
| --- | --- | --- |
| #1865 Develop CI and release-candidate quality gates | PR #1879 merged. Develop Quality Gate and manual release-candidate gate exist; demo2 evidence is attached to #1864. | Done |
| #1866 Truthful local authentication | Closed with merged evidence on GitHub before this index slice. | Done |
| #1867 Data upgrade, backup, export and restore round trip | Source-level evidence is summarized in `openspec/migration/beta-data-safety-evidence-matrix.md`, covering merged PRs #1896, #1897, and #1899-#1905. | In review pending packaged candidate proof |
| #1870 Free/Plus/Pro entitlement and plan alignment | Closed with merged evidence on GitHub before this index slice. | Done |
| #1871 Live Market Watch provider path | PR #1908 merged; #1864 contains the Voglers live provider proof summary and demo2 runtime evidence. | Done |
| #1868 Versioned Windows beta artifact and GitHub release lane | PR #1910 and PR #1913 merged. `release/windows-portable-artifact-validation.md` records the non-publishing package checklist; current artifact proof includes package name, checksum, release notes, ZIP contents, and packaged runtime smoke. | In review; prerelease publication requires #1864 approval |
| #1869 Packaged core-workflow acceptance suite | PR #1912 merged. `openspec/migration/beta-packaged-core-workflow-acceptance.md` encodes the required clean Windows packaged journey and failure-handling rules. | In progress pending approved packaged artifact execution |
| #1872 OpenSpec/task/backlog governance reconciliation | PR #1891 merged; stale active changes reconciled to archive, closed issue evidence, or explicit deferred holders. | Done |

## Current develop/runtime anchor

- Current `develop` evidence anchor at this index slice: `c4ba8087f8c8482d0e268f07d54a1c6f4d2bbd45`.
- Latest validated demo2 evidence before this index slice came from #1868 / PR #1913: `http://127.0.0.1:17882`, `/healthz` 200 `ok`, `/api/runtime.app_version=rev-c4ba8087f8c8`.
- This index does not deploy demo2 or publish artifacts; it records the release evidence map only.

## Remaining release gates

1. #1864 approval is required before attaching or publishing the versioned Windows beta artifact to a GitHub prerelease.
2. #1869 must run the encoded packaged core-workflow checklist against the exact approved artifact, checksum, commit SHA, and app version.
3. #1867 can only move from source-level data-safety review to release-ready after the packaged Windows acceptance evidence includes data upgrade/export/backup/restore proof.
4. `develop` must not merge to `main` until Max explicitly approves the tested release candidate.

## Operator notes

- Do not count this index as release approval.
- Do not claim #1864 done while #1868 publication and #1869 packaged acceptance remain incomplete.
- If #1869 finds a blocker, create or link a focused Cabinet issue with repro steps, artifact identity, expected behavior, actual behavior, evidence path, requirement link, and rerun target.
