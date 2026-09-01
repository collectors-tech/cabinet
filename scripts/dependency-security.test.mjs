import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import {
  evaluateAuditReport,
  evaluateDependencyTree,
  evaluateKnownHighAdvisoryGraph,
} from "./check-production-dependencies.mjs";

test("known release advisory graph rejects vulnerable nanoid, js-yaml, postcss and extract-zip lines", () => {
  assert.throws(
    () =>
      evaluateKnownHighAdvisoryGraph({
        packages: {
          "node_modules/nanoid": { version: "3.3.16" },
          "node_modules/tool/node_modules/nanoid": { version: "5.1.15" },
          "node_modules/js-yaml": { version: "4.3.0", dev: true },
          "node_modules/postcss": { version: "8.5.22" },
          "node_modules/extract-zip": { version: "2.0.1", dev: true },
        },
      }),
    /nanoid.*3\.3\.16.*GHSA-2v37-7h3g-55p8[\s\S]*nanoid.*5\.1\.15.*GHSA-28wg-ghj8-5hjv[\s\S]*js-yaml.*4\.3\.0.*GHSA-5p4m-2wfm-xmqj[\s\S]*postcss.*8\.5\.22.*GHSA-fxqj-rqcc-2cmp[\s\S]*extract-zip.*2\.0\.1.*GHSA-jmr9-qjv8-65gv/,
  );

  assert.deepEqual(
    evaluateKnownHighAdvisoryGraph({
      packages: {
        "node_modules/nanoid": { version: "3.3.17" },
        "node_modules/tool/node_modules/nanoid": { version: "5.1.16" },
        "node_modules/js-yaml": { version: "4.3.1", dev: true },
        "node_modules/postcss": { version: "8.5.23" },
        "node_modules/esbuild": { version: "0.28.1" },
      },
    }),
    { nanoid: 2, "js-yaml": 1, postcss: 1, "extract-zip": 0, esbuild: 1 },
  );
});

test("known release advisory graph rejects the vulnerable esbuild Windows development server", () => {
  assert.throws(
    () =>
      evaluateKnownHighAdvisoryGraph({
        packages: {
          "node_modules/esbuild": { version: "0.27.3" },
        },
      }),
    /esbuild.*0\.27\.3.*GHSA-g7r4-m6w7-qqqr.*0\.28\.1/,
  );
});

test("current lock graph is above every governed advisory floor or omits the package", async () => {
  const packageLock = JSON.parse(
    await readFile("ui.web/package-lock.json", "utf8"),
  );
  const counts = evaluateKnownHighAdvisoryGraph(packageLock);
  assert.ok(counts.nanoid > 0, "the nanoid advisory contract must exercise a lock path");
  assert.ok(counts["js-yaml"] > 0, "the js-yaml advisory contract must exercise a lock path");
  assert.ok(counts.postcss > 0, "the postcss advisory contract must exercise a lock path");
  assert.equal(counts["extract-zip"], 0, "the unpatched extract-zip package must be absent");
  assert.ok(counts.esbuild > 0, "the esbuild advisory contract must exercise a lock path");
});

test("production dependency audit rejects unresolved critical and high findings", () => {
  assert.throws(
    () =>
      evaluateAuditReport({
        metadata: {
          vulnerabilities: { critical: 1, high: 2, moderate: 3, low: 4 },
        },
      }),
    /1 critical and 2 high/,
  );
});

test("production dependency audit rejects invalid severity counts", () => {
  assert.throws(
    () =>
      evaluateAuditReport({
        metadata: {
          vulnerabilities: { critical: 0, high: "0", moderate: 0, low: 0 },
        },
      }),
    /invalid high vulnerability count/,
  );
});

test("production dependency audit permits lower-severity findings", () => {
  assert.deepEqual(
    evaluateAuditReport({
      metadata: {
        vulnerabilities: { critical: 0, high: 0, moderate: 2, low: 1 },
      },
    }),
    { critical: 0, high: 0, moderate: 2, low: 1 },
  );
});

test("production dependency tree rejects clean-install drift", () => {
  assert.throws(
    () =>
      evaluateDependencyTree({
        problems: ["extraneous: stale-package@1.0.0 node_modules/stale-package"],
      }),
    /extraneous: stale-package/,
  );
  assert.doesNotThrow(() => evaluateDependencyTree({ problems: [] }));
});

test("production dependency tree permits lock-owned optional bundle packages", () => {
  assert.doesNotThrow(() =>
    evaluateDependencyTree(
      {
        problems: [
          "extraneous: @napi-rs/wasm-runtime@1.1.1 node_modules/@napi-rs/wasm-runtime",
        ],
      },
      {
        packages: {
          "node_modules/@napi-rs/wasm-runtime": { optional: true },
        },
      },
    ),
  );
});

test("develop, candidate, main and package workflows enforce the security gate", async () => {
  const packageManifest = JSON.parse(await readFile("package.json", "utf8"));
  assert.equal(
    packageManifest.scripts["security:dependencies"],
    "node scripts/check-production-dependencies.mjs",
  );

  for (const workflow of [
    ".github/workflows/develop-quality-gate.yml",
    ".github/workflows/beta-release-candidate.yml",
    ".github/workflows/main-gate.yml",
    ".github/workflows/release-installers.yml",
  ]) {
    const source = await readFile(workflow, "utf8");
    assert.match(
      source,
      /name: Enforce production dependency security[\s\S]*npm run security:dependencies/,
      workflow,
    );
  }

  const candidateWorkflow = await readFile(
    ".github/workflows/beta-release-candidate.yml",
    "utf8",
  );
  assert.match(
    candidateWorkflow,
    /security:dependencies[^\n]*Tee-Object[^\n]*\n\s+if \(\$LASTEXITCODE -ne 0\)/,
    "candidate evidence logging must preserve a failing dependency-gate exit",
  );
});

test("dependency security gate remains bound to OpenSpec and release evidence", async () => {
  const runtimeSpec = await readFile(
    "openspec/specs/general/runtime-core/spec.md",
    "utf8",
  );
  const traceability = await readFile("openspec/traceability.md", "utf8");
  const evidence = await readFile(
    "openspec/migration/beta-dependency-security-evidence.md",
    "utf8",
  );

  assert.match(runtimeSpec, /Requirement RUNTIME-CORE-025:/);
  assert.match(traceability, /RUNTIME-CORE-025[\s\S]*#2051/);
  assert.match(evidence, /1 critical and 6 high/);
  assert.match(evidence, /zero production vulnerabilities at every severity/);
  assert.match(evidence, /`esbuild` 0\.28\.2/);
  assert.match(evidence, /#2559/);
  assert.match(evidence, /lock-owned optional WebAssembly bundle/);
  assert.match(evidence, /#56-#60/);
  assert.match(evidence, /no `extract-zip` entry/);
  assert.match(evidence, /`postcss` 8\.5\.26/);
});
