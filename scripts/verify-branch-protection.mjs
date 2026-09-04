import { resolve } from "node:path";

import { verifyRepositoryProtection } from "./lib/branch-protection-contract.mjs";

const option = (name) => {
  const index = process.argv.indexOf(name);
  return index >= 0 ? process.argv[index + 1] : undefined;
};

if (process.argv.includes("--help")) {
  console.log(
    "Usage: node scripts/verify-branch-protection.mjs [--policy <path>] [--json]",
  );
  console.log(
    "Reads GitHub protection/ruleset state only. Authentication uses GH_TOKEN or GITHUB_TOKEN.",
  );
  process.exit(0);
}

const policyPath = resolve(
  option("--policy") ?? "release/github-branch-protection-policy.json",
);
const result = await verifyRepositoryProtection(policyPath);
if (process.argv.includes("--json")) {
  console.log(JSON.stringify(result, null, 2));
} else if (result.compliant) {
  console.log(`Branch protection compliant: ${result.repository}`);
} else {
  console.error(`Branch protection drift: ${result.repository}`);
  for (const error of result.errors) console.error(`- ${error}`);
}
if (!result.compliant) process.exitCode = 1;
