import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import test from "node:test";

import {
  evaluateRepositoryProtection,
  hasExactApprovalLine,
  loadProtectionPolicy,
} from "./lib/branch-protection-contract.mjs";

const repoRoot = resolve(import.meta.dirname, "..");
const policyPath = resolve(
  repoRoot,
  "release",
  "github-branch-protection-policy.json",
);
const read = (path) => readFileSync(resolve(repoRoot, path), "utf8");

const protectedBranch = (
  requiredChecks,
  { approvals = 0, codeOwners = false } = {},
) => ({
  required_status_checks: {
    strict: true,
    contexts: requiredChecks,
    checks: requiredChecks.map((context) => ({ context, app_id: 15368 })),
  },
  required_pull_request_reviews: {
    dismiss_stale_reviews: true,
    require_code_owner_reviews: codeOwners,
    required_approving_review_count: approvals,
    bypass_pull_request_allowances: { users: [], teams: [], apps: [] },
  },
  enforce_admins: { enabled: true },
  required_linear_history: { enabled: true },
  required_conversation_resolution: { enabled: true },
  allow_force_pushes: { enabled: false },
  allow_deletions: { enabled: false },
  block_creations: { enabled: false },
  lock_branch: { enabled: false },
  allow_fork_syncing: { enabled: false },
});

test("declares the exact develop and main release protection contract", async () => {
  const policy = await loadProtectionPolicy(policyPath);

  assert.equal(policy.schema_version, 1);
  assert.equal(policy.repository, "collectors-tech/cabinet");
  assert.deepEqual(policy.emergency_bypass.persistent_bypass_actors, []);
  assert.equal(policy.emergency_bypass.authority, "wildone");
  assert.match(
    policy.release_approval.marker,
    /APPROVE CABINET 0\.1 PRIVATE BETA <exact-commit>/,
  );

  assert.deepEqual(policy.branches.develop.required_checks, [
    "Workflow contract",
    "OpenSpec strict validation",
    "UI production build",
    "Go runtime package tests",
    "OpenAPI parity and docs",
    "Cypress login/profile/runtime smoke",
    "Windows portable package verification",
  ]);
  assert.deepEqual(policy.branches.main.required_checks, [
    "Go CI (ubuntu-latest)",
    "Go CI (windows-latest)",
    "Go CI (macos-latest)",
    "OpenAPI",
    "UI Build",
    "Cypress",
    "Windows portable package verification",
    "Exact #1864 promotion approval",
  ]);
  assert.equal(policy.branches.develop.required_approving_review_count, 0);
  assert.equal(policy.branches.main.required_approving_review_count, 0);
  assert.equal(policy.branches.main.require_code_owner_reviews, false);
});

test("accepts the exact protected branch state and no repository ruleset bypass", async () => {
  const policy = await loadProtectionPolicy(policyPath);
  const state = {
    repository: { allow_auto_merge: false },
    protections: {
      develop: protectedBranch(policy.branches.develop.required_checks),
      main: protectedBranch(policy.branches.main.required_checks),
    },
    rulesets: [],
  };

  assert.deepEqual(evaluateRepositoryProtection(policy, state), {
    repository: "collectors-tech/cabinet",
    compliant: true,
    errors: [],
  });
});

test("reports unprotected branches and every missing required quality check", async () => {
  const policy = await loadProtectionPolicy(policyPath);
  const result = evaluateRepositoryProtection(policy, {
    repository: { allow_auto_merge: false },
    protections: { develop: null, main: null },
    rulesets: [],
  });

  assert.equal(result.compliant, false);
  assert.ok(result.errors.includes("develop:branch_unprotected"));
  assert.ok(result.errors.includes("main:branch_unprotected"));
});

test("rejects weakened merge controls and any persistent ruleset bypass actor", async () => {
  const policy = await loadProtectionPolicy(policyPath);
  const develop = protectedBranch(
    policy.branches.develop.required_checks.slice(1),
  );
  develop.required_status_checks.strict = false;
  develop.required_pull_request_reviews = null;
  develop.enforce_admins.enabled = false;
  develop.allow_force_pushes.enabled = true;
  develop.required_status_checks.checks =
    develop.required_status_checks.checks.map((check) => ({
      ...check,
      app_id: null,
    }));
  const main = protectedBranch(policy.branches.main.required_checks, {
    approvals: 1,
    codeOwners: true,
  });

  const result = evaluateRepositoryProtection(policy, {
    repository: { allow_auto_merge: true },
    protections: {
      develop,
      main,
    },
    rulesets: [
      {
        id: 12,
        name: "unsafe bypass",
        target: "branch",
        enforcement: "active",
        bypass_actors: [
          { actor_type: "Integration", actor_id: 15368, bypass_mode: "always" },
        ],
      },
    ],
  });

  for (const expected of [
    "develop:required_check_missing:Workflow contract",
    "develop:required_check_app_mismatch:OpenSpec strict validation",
    "develop:required_status_checks_not_strict",
    "develop:pull_request_reviews_not_required",
    "develop:administrators_not_enforced",
    "develop:force_push_allowed",
    "main:required_approvals_mismatch:0:1",
    "main:code_owner_review_mismatch:false:true",
    "repository:auto_merge_enabled",
    "ruleset:12:persistent_bypass_actor:Integration:15368:always",
  ]) {
    assert.ok(
      result.errors.includes(expected),
      `missing drift error ${expected}`,
    );
  }
});

test("binds policy check names to the workflow jobs that actually report them", async () => {
  const policy = await loadProtectionPolicy(policyPath);
  const developWorkflow = read(".github/workflows/develop-quality-gate.yml");
  const mainWorkflow = read(".github/workflows/main-gate.yml");
  const approvalWorkflow = read(
    ".github/workflows/main-promotion-approval.yml",
  );

  for (const check of policy.branches.develop.required_checks) {
    assert.ok(
      developWorkflow.includes(`name: ${check}`),
      `develop workflow does not report ${check}`,
    );
  }
  for (const check of policy.branches.main.required_checks) {
    if (check === "Exact #1864 promotion approval") {
      assert.ok(approvalWorkflow.includes(`name: "${check}"`));
      continue;
    }
    const matrixCheck = check.match(/^Go CI \((.+)\)$/);
    if (matrixCheck) {
      assert.match(mainWorkflow, /name: Go CI/);
      assert.ok(
        mainWorkflow.includes(matrixCheck[1]),
        `main workflow has no ${matrixCheck[1]} matrix lane`,
      );
    } else {
      assert.ok(
        mainWorkflow.includes(`name: ${check}`),
        `main workflow does not report ${check}`,
      );
    }
  }
});

test("keeps verification read-only and documents approval and emergency evidence", () => {
  const verifier = `${read("scripts/verify-branch-protection.mjs")}\n${read("scripts/lib/branch-protection-contract.mjs")}`;
  const documentation = read("openspec/migration/github-branch-protection.md");
  const publisherWorkflow = read(
    ".github/workflows/publish-beta-prerelease.yml",
  );

  assert.doesNotMatch(
    verifier,
    /method\s*:\s*['"](?:POST|PUT|PATCH|DELETE)['"]/,
  );
  const approvalWorkflow = read(
    ".github/workflows/main-promotion-approval.yml",
  );
  for (const fragment of [
    "pull_request:",
    "branches: [main]",
    "pr.head.ref !== 'develop'",
    "comment.user?.login === 'wildone'",
    "APPROVE CABINET 0.1 PRIVATE BETA",
    "issues: read",
  ]) {
    assert.ok(
      approvalWorkflow.includes(fragment),
      `approval workflow missing ${fragment}`,
    );
  }
  assert.doesNotMatch(
    approvalWorkflow,
    /actions\/checkout|contents:\s*write|git push|pull-requests:\s*write/,
  );
  assert.ok(
    publisherWorkflow.includes("comment.user?.login !== 'wildone'"),
    "prerelease publisher must reject approval from anyone other than the release owner",
  );
  for (const workflow of [approvalWorkflow, publisherWorkflow]) {
    assert.ok(
      workflow.includes(
        "split(/\\r?\\n/).some((line) => line.trim() === marker)",
      ),
      "approval workflow must require the marker as one exact trimmed line",
    );
    assert.doesNotMatch(workflow, /body\.includes\(marker\)/);
  }
  for (const fragment of [
    "APPROVE CABINET 0.1 PRIVATE BETA <exact-commit>",
    "There is no persistent emergency bypass.",
    "GitHub audit-log/settings evidence",
    "does not authorize external publication",
    "Bootstrap the new required check",
  ]) {
    assert.ok(
      documentation.includes(fragment),
      `protection guide missing ${fragment}`,
    );
  }
});

test("rejects embedded or suffixed release approval markers", () => {
  const marker =
    "APPROVE CABINET 0.1 PRIVATE BETA 0123456789abcdef0123456789abcdef01234567";
  assert.equal(hasExactApprovalLine(marker, marker), true);
  assert.equal(hasExactApprovalLine(`  ${marker}  `, marker), true);
  assert.equal(
    hasExactApprovalLine(`Context\n${marker}\nEvidence`, marker),
    true,
  );
  assert.equal(hasExactApprovalLine(`prefix ${marker}`, marker), false);
  assert.equal(hasExactApprovalLine(`${marker} suffix`, marker), false);
  assert.equal(hasExactApprovalLine(`\`${marker}\``, marker), false);
});

test("traces protected promotion governance into canonical OpenSpec", () => {
  const spec = read("openspec/specs/general/non-functional/spec.md");
  const traceability = read("openspec/traceability.md");

  for (const fragment of [
    "Requirement NON-FUNCTIONAL-004:",
    "required GitHub Actions checks",
    "explicit exact-commit #1864 approval",
    "no persistent bypass actor",
  ]) {
    assert.ok(
      spec.includes(fragment),
      `non-functional spec missing ${fragment}`,
    );
  }
  assert.match(
    traceability,
    /NON-FUNCTIONAL-004.*#2056.*branch-protection-contract\.test\.mjs/,
  );
});
