# GitHub branch protection for Cabinet 0.1

Issue: #2056. Parent approval gate: #1864.

The machine-readable source of truth is `release/github-branch-protection-policy.json`. Check the live repository without changing it:

```powershell
$env:GH_TOKEN = gh auth token
node scripts/verify-branch-protection.mjs --json
```

The command is read-only. It reads the repository, classic branch-protection and repository-ruleset APIs, emits only non-secret configuration drift and exits nonzero for drift. Do not attach the token or raw authenticated API headers as evidence.

## Required remote configuration

Configure classic branch protection for both branches. Bind every required status check to GitHub Actions app ID `15368`, require the branch to be current before merge, require pull requests, require linear history and resolved conversations, enforce rules for administrators, and disallow force pushes and deletion. Set required approving reviews to zero because the repository currently has one maintainer and GitHub rejects self-review; the exact release-approval check below provides the human promotion gate. Leave pull-request bypass allowances empty. Keep repository auto-merge disabled and do not configure an active branch ruleset with a bypass actor.

`develop` requires these checks from **Develop Quality Gate**:

- `Workflow contract`
- `OpenSpec strict validation`
- `UI production build`
- `Go runtime package tests`
- `OpenAPI parity and docs`
- `Cypress login/profile/runtime smoke`
- `Windows portable package verification`

`main` requires these checks from **Main Gate**:

- `Go CI (ubuntu-latest)`
- `Go CI (windows-latest)`
- `Go CI (macos-latest)`
- `OpenAPI`
- `UI Build`
- `Cypress`
- `Windows portable package verification`
- `Exact #1864 promotion approval` from **Main Promotion Approval**

The approval job runs without checkout and has read-only repository permissions. It only accepts a same-repository `develop` to `main` pull request whose exact head SHA has an exact marker authored by `wildone` on #1864. A changed head gets a new check run and cannot reuse approval for an earlier commit.

The settings write is intentionally not automated by this repository. A repository administrator must review the desired state, apply it in GitHub, and preserve the before/after verifier output and settings audit event. Re-run the verifier immediately afterward.

### Bootstrap the new required check

The REST API can configure a required check by its declared context and GitHub Actions app ID. If the GitHub settings UI will not offer `Exact #1864 promotion approval` until it has observed the context, use this one-time sequence after this repository-owned preparation is on `develop`:

1. With separate authorization to create a pull request, open a draft same-repository `develop` to `main` pull request labelled as protection bootstrap only.
2. Let **Main Promotion Approval** report the `Exact #1864 promotion approval` check. It is expected to fail because no release approval marker exists; do not add a fake marker, merge, publish or promote.
3. Close the bootstrap pull request unmerged.
4. With separate authorization to change repository settings, configure both branch protections from the policy file, including the newly observed check context and app ID `15368`.
5. Capture the before/after state and audit event, then rerun the read-only verifier. It must report compliant before the release-promotion path is used.

## Release approval and promotion

Internal candidate creation and packaged acceptance may happen before final approval. External prerelease publication and `develop` to `main` promotion may not.

The exact accepted commit must have successful #1868/#1869/#1867 evidence linked to #1864. Max (`wildone`) then posts this exact marker on #1864:

```text
APPROVE CABINET 0.1 PRIVATE BETA <exact-commit>
```

The promotion pull request must have `develop` as its head, `main` as its base and the approved commit as its head SHA. The required **Exact #1864 promotion approval** check is the explicit promotion authorization. Post the marker before opening the promotion pull request; if approval arrives later, rerun the failed approval check. A new push requires a new exact-commit #1864 marker. Workflows and GitHub Apps have no bypass allowance and cannot directly push or approve their way around this sequence. The prerelease publisher separately verifies the issue number, exact marker, candidate run and commit before publication.

## Emergency/P0 path

There is no persistent emergency bypass. Only Max (`wildone`), acting as repository administrator, may temporarily change protection for a confirmed P0 incident. The P0 issue must exist first and record:

- incident impact and why the normal protected path cannot resolve it in time;
- actor, UTC start/end time, exact pre-change and post-restoration verifier output;
- exact commit(s), changed setting, reason and checks run;
- GitHub audit-log/settings evidence and follow-up review or rollback result.

Restore the normal policy immediately after the bounded action and rerun the verifier. An emergency bypass does not authorize external publication or `develop` to `main` promotion without the exact #1864 approval marker. If the emergency commit changes a release candidate, rebuild and reaccept the candidate.
