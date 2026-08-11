import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import {
  evaluateAuditReport,
  evaluateDependencyTree,
} from "./check-production-dependencies.mjs";

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
  assert.match(evidence, /zero critical or high production vulnerabilities/);
  assert.match(evidence, /one low-severity `esbuild` advisory/);
  assert.match(evidence, /lock-owned optional WebAssembly bundle/);
});
