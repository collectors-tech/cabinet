import { readFile } from "node:fs/promises";

const githubHeaders = (token) => ({
  Accept: "application/vnd.github+json",
  "X-GitHub-Api-Version": "2022-11-28",
  ...(token ? { Authorization: `Bearer ${token}` } : {}),
});

const enabled = (value) => value === true || value?.enabled === true;
const values = (value) => (Array.isArray(value) ? value : []);

export const hasExactApprovalLine = (body, marker) =>
  typeof body === "string" &&
  body.split(/\r?\n/).some((line) => line.trim() === marker);

const fetchJson = async (url, { token, fetchImpl, allowNotFound = false }) => {
  const response = await fetchImpl(url, { headers: githubHeaders(token) });
  if (allowNotFound && response.status === 404) return null;
  if (!response.ok) {
    const body = await response.text();
    throw new Error(
      `github_api_${response.status}:${url}:${body.slice(0, 300)}`,
    );
  }
  return response.json();
};

export const loadProtectionPolicy = async (path) => {
  const policy = JSON.parse(await readFile(path, "utf8"));
  if (policy.schema_version !== 1)
    throw new Error("branch_protection_policy_schema_unsupported");
  if (!/^[^/]+\/[^/]+$/.test(policy.repository ?? ""))
    throw new Error("branch_protection_policy_repository_invalid");
  if (!Number.isSafeInteger(policy.required_check_app_id))
    throw new Error("branch_protection_policy_app_id_invalid");
  for (const branch of ["develop", "main"]) {
    if (
      !Array.isArray(policy.branches?.[branch]?.required_checks) ||
      policy.branches[branch].required_checks.length === 0
    ) {
      throw new Error(`branch_protection_policy_checks_invalid:${branch}`);
    }
  }
  return policy;
};

export const readRepositoryProtection = async (
  policy,
  {
    token = process.env.GITHUB_TOKEN || process.env.GH_TOKEN,
    fetchImpl = globalThis.fetch,
    apiBase = "https://api.github.com",
  } = {},
) => {
  if (typeof fetchImpl !== "function")
    throw new Error("github_fetch_unavailable");
  const repositoryPath = `repos/${policy.repository}`;
  const repository = await fetchJson(`${apiBase}/${repositoryPath}`, {
    token,
    fetchImpl,
  });
  const protections = {};
  for (const branch of Object.keys(policy.branches)) {
    protections[branch] = await fetchJson(
      `${apiBase}/${repositoryPath}/branches/${encodeURIComponent(branch)}/protection`,
      { token, fetchImpl, allowNotFound: true },
    );
  }

  const summaries = await fetchJson(
    `${apiBase}/${repositoryPath}/rulesets?per_page=100`,
    { token, fetchImpl },
  );
  const rulesets = [];
  for (const ruleset of values(summaries)) {
    rulesets.push(
      await fetchJson(`${apiBase}/${repositoryPath}/rulesets/${ruleset.id}`, {
        token,
        fetchImpl,
      }),
    );
  }
  return { repository, protections, rulesets };
};

const evaluateBranch = (branch, expected, actual, appID) => {
  if (!actual) return [`${branch}:branch_unprotected`];
  const errors = [];
  const checks = values(actual.required_status_checks?.checks);
  const contexts = new Set([
    ...values(actual.required_status_checks?.contexts),
    ...checks.map((check) => check.context),
  ]);
  for (const required of expected.required_checks) {
    if (!contexts.has(required)) {
      errors.push(`${branch}:required_check_missing:${required}`);
      continue;
    }
    if (
      !checks.some(
        (check) => check.context === required && check.app_id === appID,
      )
    ) {
      errors.push(`${branch}:required_check_app_mismatch:${required}`);
    }
  }
  if (actual.required_status_checks?.strict !== true)
    errors.push(`${branch}:required_status_checks_not_strict`);

  const reviews = actual.required_pull_request_reviews;
  if (!reviews) {
    errors.push(`${branch}:pull_request_reviews_not_required`);
  } else {
    const actualApprovalCount = reviews.required_approving_review_count ?? 0;
    if (actualApprovalCount !== expected.required_approving_review_count) {
      errors.push(
        `${branch}:required_approvals_mismatch:${expected.required_approving_review_count}:${actualApprovalCount}`,
      );
    }
    if (
      expected.dismiss_stale_reviews &&
      reviews.dismiss_stale_reviews !== true
    ) {
      errors.push(`${branch}:stale_reviews_not_dismissed`);
    }
    const actualCodeOwnerReviews = reviews.require_code_owner_reviews === true;
    if (actualCodeOwnerReviews !== expected.require_code_owner_reviews) {
      errors.push(
        `${branch}:code_owner_review_mismatch:${expected.require_code_owner_reviews}:${actualCodeOwnerReviews}`,
      );
    }
    const bypass = reviews.bypass_pull_request_allowances ?? {};
    for (const actorType of ["users", "teams", "apps"]) {
      for (const actor of values(bypass[actorType])) {
        errors.push(
          `${branch}:pull_request_bypass_actor:${actorType}:${actor.login ?? actor.slug ?? actor.name ?? actor.id}`,
        );
      }
    }
  }

  if (expected.enforce_admins && !enabled(actual.enforce_admins))
    errors.push(`${branch}:administrators_not_enforced`);
  if (
    expected.required_linear_history &&
    !enabled(actual.required_linear_history)
  )
    errors.push(`${branch}:linear_history_not_required`);
  if (
    expected.required_conversation_resolution &&
    !enabled(actual.required_conversation_resolution)
  ) {
    errors.push(`${branch}:conversation_resolution_not_required`);
  }
  if (!expected.allow_force_pushes && enabled(actual.allow_force_pushes))
    errors.push(`${branch}:force_push_allowed`);
  if (!expected.allow_deletions && enabled(actual.allow_deletions))
    errors.push(`${branch}:deletion_allowed`);
  return errors;
};

export const evaluateRepositoryProtection = (policy, state) => {
  const errors = [];
  if (
    policy.repository_settings?.allow_auto_merge === false &&
    state.repository?.allow_auto_merge !== false
  ) {
    errors.push("repository:auto_merge_enabled");
  }
  for (const [branch, expected] of Object.entries(policy.branches)) {
    errors.push(
      ...evaluateBranch(
        branch,
        expected,
        state.protections?.[branch],
        policy.required_check_app_id,
      ),
    );
  }
  for (const ruleset of values(state.rulesets)) {
    if (ruleset.target !== "branch" || ruleset.enforcement !== "active")
      continue;
    for (const actor of values(ruleset.bypass_actors)) {
      errors.push(
        `ruleset:${ruleset.id}:persistent_bypass_actor:${actor.actor_type}:${actor.actor_id}:${actor.bypass_mode}`,
      );
    }
  }
  return {
    repository: policy.repository,
    compliant: errors.length === 0,
    errors,
  };
};

export const verifyRepositoryProtection = async (policyPath, options = {}) => {
  const policy = await loadProtectionPolicy(policyPath);
  const state = await readRepositoryProtection(policy, options);
  return evaluateRepositoryProtection(policy, state);
};
