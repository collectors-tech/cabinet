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

function parseLockedVersion(packageName, packagePath, version) {
  const match = /^(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z.-]+))?$/.exec(version ?? "");
  if (!match) {
    throw new Error(
      `${packageName} at ${packagePath} has an invalid locked version: ${String(version)}`,
    );
  }
  return {
    major: Number(match[1]),
    minor: Number(match[2]),
    patch: Number(match[3]),
    prerelease: match[4] !== undefined,
  };
}

function isBefore(version, floor) {
  for (const field of ["major", "minor", "patch"]) {
    if (version[field] !== floor[field]) {
      return version[field] < floor[field];
    }
  }
  return version.prerelease;
}

function nanoidAdvisory(version) {
  if (version.major < 3 || (version.major === 3 && isBefore(version, { major: 3, minor: 3, patch: 17 }))) {
    return "GHSA-2v37-7h3g-55p8 (patched in 3.3.17)";
  }
  if (version.major === 4 || (version.major === 5 && isBefore(version, { major: 5, minor: 1, patch: 6 }))) {
    return "GHSA-2v37-7h3g-55p8 (patched in 5.1.6)";
  }
  if (version.major === 5 && isBefore(version, { major: 5, minor: 1, patch: 16 })) {
    return "GHSA-28wg-ghj8-5hjv (patched in 5.1.16)";
  }
  return null;
}

function jsYamlAdvisory(version) {
  if (version.major === 3 && isBefore(version, { major: 3, minor: 15, patch: 1 })) {
    return "GHSA-5p4m-2wfm-xmqj (patched in 3.15.1)";
  }
  if (version.major === 4 && isBefore(version, { major: 4, minor: 3, patch: 1 })) {
    return "GHSA-5p4m-2wfm-xmqj (patched in 4.3.1)";
  }
  return null;
}

function postcssAdvisory(version) {
  if (isBefore(version, { major: 8, minor: 5, patch: 23 })) {
    return "GHSA-fxqj-rqcc-2cmp (patched in 8.5.23)";
  }
  return null;
}

function extractZipAdvisory(version) {
  if (isBefore(version, { major: 2, minor: 0, patch: 2 })) {
    return "GHSA-jmr9-qjv8-65gv (no patched release; package must remain absent)";
  }
  return null;
}

export function evaluateKnownHighAdvisoryGraph(packageLock) {
  if (!packageLock?.packages || typeof packageLock.packages !== "object") {
    throw new Error("package lock does not contain a packages graph");
  }

  const findings = [];
  const counts = { nanoid: 0, "js-yaml": 0, postcss: 0, "extract-zip": 0 };
  for (const packageName of ["nanoid", "js-yaml", "postcss", "extract-zip"]) {
    const suffix = `/node_modules/${packageName}`;
    for (const [rawPackagePath, entry] of Object.entries(packageLock.packages)) {
      const packagePath = rawPackagePath.replaceAll("\\", "/");
      if (packagePath !== `node_modules/${packageName}` && !packagePath.endsWith(suffix)) {
        continue;
      }
      counts[packageName] += 1;
      const version = parseLockedVersion(packageName, packagePath, entry?.version);
      const advisory =
        packageName === "nanoid"
          ? nanoidAdvisory(version)
          : packageName === "js-yaml"
            ? jsYamlAdvisory(version)
            : packageName === "postcss"
              ? postcssAdvisory(version)
              : extractZipAdvisory(version);
      if (advisory) {
        findings.push(`${packageName} ${entry.version} at ${packagePath}: ${advisory}`);
      }
    }
  }

  if (findings.length > 0) {
    throw new Error(`lock graph contains known high dependency advisories:\n${findings.join("\n")}`);
  }
  return counts;
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
  const graphCounts = evaluateKnownHighAdvisoryGraph(packageLock);
  evaluateDependencyTree(runNpmJson(["ls", "--depth=0", "--json"]), packageLock);
  const counts = evaluateAuditReport(runNpmJson(["audit", "--omit=dev", "--json"]));
  console.log(
    `Known release advisory graph passed: nanoid=${graphCounts.nanoid} js-yaml=${graphCounts["js-yaml"]} postcss=${graphCounts.postcss} extract-zip=${graphCounts["extract-zip"]}`,
  );
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
