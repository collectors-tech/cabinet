import { spawnSync } from "node:child_process";
import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import path from "node:path";

const repositoryRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const uiRoot = path.join(repositoryRoot, "ui.web");

function vulnerabilityCount(report, severity) {
  const value = report?.metadata?.vulnerabilities?.[severity];
  if (!Number.isInteger(value) || value < 0) {
    throw new Error(`npm audit returned an invalid ${severity} vulnerability count`);
  }
  return value;
}

export function evaluateAuditReport(report) {
  if (!report?.metadata?.vulnerabilities) {
    throw new Error("npm audit did not return vulnerability metadata");
  }

  const counts = {
    critical: vulnerabilityCount(report, "critical"),
    high: vulnerabilityCount(report, "high"),
    moderate: vulnerabilityCount(report, "moderate"),
    low: vulnerabilityCount(report, "low"),
  };
  if (counts.critical > 0 || counts.high > 0) {
    throw new Error(
      `production dependency audit found ${counts.critical} critical and ${counts.high} high unresolved vulnerabilities`,
    );
  }
  return counts;
}

function isLockOwnedOptionalProblem(problem, packageLock) {
  if (!problem.startsWith("extraneous:")) {
    return false;
  }
  const normalized = problem.replaceAll("\\", "/");
  const marker = "node_modules/";
  const packagePathIndex = normalized.lastIndexOf(marker);
  if (packagePathIndex < 0) {
    return false;
  }
  const packagePath = normalized.slice(packagePathIndex);
  return packageLock?.packages?.[packagePath]?.optional === true;
}

export function evaluateDependencyTree(tree, packageLock = {}) {
  const problems = Array.isArray(tree?.problems)
    ? tree.problems
        .filter(Boolean)
        .filter((problem) => !isLockOwnedOptionalProblem(problem, packageLock))
    : [];
  if (problems.length > 0) {
    throw new Error(`npm dependency tree contains drift:\n${problems.join("\n")}`);
  }
  if (tree?.error) {
    throw new Error(`npm dependency tree inspection failed: ${JSON.stringify(tree.error)}`);
  }
  return { problems: 0 };
}

function runNpmJson(args) {
  const npmCli = process.env.npm_execpath;
  if (!npmCli) {
    throw new Error("npm_execpath is unavailable; run this check through npm run");
  }
  const result = spawnSync(process.execPath, [npmCli, ...args], {
    cwd: uiRoot,
    encoding: "utf8",
    maxBuffer: 20 * 1024 * 1024,
    windowsHide: true,
  });
  if (result.error) {
    throw result.error;
  }

  try {
    return JSON.parse(result.stdout);
  } catch {
    const detail = result.stderr.trim() || result.stdout.trim() || `exit ${result.status}`;
    throw new Error(`npm ${args.join(" ")} did not return JSON: ${detail}`);
  }
}

export function main() {
  const packageLock = JSON.parse(readFileSync(path.join(uiRoot, "package-lock.json"), "utf8"));
  evaluateDependencyTree(runNpmJson(["ls", "--depth=0", "--json"]), packageLock);
  const counts = evaluateAuditReport(runNpmJson(["audit", "--omit=dev", "--json"]));
  console.log(
    `Production dependency audit passed: critical=${counts.critical} high=${counts.high} moderate=${counts.moderate} low=${counts.low}`,
  );
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  try {
    main();
  } catch (error) {
    console.error(error instanceof Error ? error.message : String(error));
    process.exitCode = 1;
  }
}
