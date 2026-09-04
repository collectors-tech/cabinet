#!/usr/bin/env node
import fs from "node:fs";
import path from "node:path";
import process from "node:process";

function readArg(name, fallback = "") {
  const index = process.argv.indexOf(name);
  if (index === -1) {
    return fallback;
  }
  return process.argv[index + 1] ?? "";
}

function fail(message) {
  console.error(message);
  process.exit(1);
}

function normalizeSpec(value) {
  return value.replaceAll("\\", "/").replace(/^\.?\//, "").trim();
}

function parseOverride(value) {
  if (!value || !value.trim()) {
    return [];
  }
  return value
    .split(",")
    .map(normalizeSpec)
    .filter(Boolean);
}

const repoRoot = process.cwd();
const manifestPath = path.resolve(repoRoot, readArg("--manifest", "release/beta-core-cypress-pack.json"));
const outputPath = readArg("--output", "");
const overrideSpecs = parseOverride(readArg("--spec-override", ""));

const manifest = JSON.parse(fs.readFileSync(manifestPath, "utf8"));
if (!Number.isInteger(manifest.version) || manifest.version < 1) {
  fail("Beta core Cypress pack manifest must contain a positive integer version.");
}
if (!Array.isArray(manifest.required_categories) || manifest.required_categories.length === 0) {
  fail("Beta core Cypress pack manifest must declare required_categories.");
}
if (!Array.isArray(manifest.specs) || manifest.specs.length === 0) {
  fail("Beta core Cypress pack manifest must declare at least one Cypress spec.");
}

const requiredCategories = new Set(manifest.required_categories);
const specs = manifest.specs.map((entry) => {
  if (!entry || typeof entry !== "object") {
    fail("Beta core Cypress pack specs must be objects with category and path.");
  }
  const category = String(entry.category ?? "").trim();
  const specPath = normalizeSpec(String(entry.path ?? ""));
  if (!requiredCategories.has(category)) {
    fail(`Beta core Cypress pack spec ${specPath} uses undeclared category ${category}.`);
  }
  if (!specPath.startsWith("cypress/e2e/") || !specPath.endsWith(".cy.ts")) {
    fail(`Beta core Cypress pack spec ${specPath} must be a Cypress e2e TypeScript spec.`);
  }
  const absoluteSpecPath = path.join(repoRoot, "ui.web", ...specPath.split("/"));
  if (!fs.existsSync(absoluteSpecPath)) {
    fail(`Beta core Cypress pack spec does not exist: ${specPath}`);
  }
  return { category, path: specPath };
});

const seenSpecs = new Set();
for (const spec of specs) {
  if (seenSpecs.has(spec.path)) {
    fail(`Beta core Cypress pack contains duplicate spec: ${spec.path}`);
  }
  seenSpecs.add(spec.path);
}

for (const requiredCategory of requiredCategories) {
  if (!specs.some((spec) => spec.category === requiredCategory)) {
    fail(`Beta core Cypress pack is missing required category: ${requiredCategory}`);
  }
}

if (overrideSpecs.length > 0) {
  const overrideSet = new Set();
  for (const spec of overrideSpecs) {
    if (overrideSet.has(spec)) {
      fail(`Cypress pack override contains duplicate spec: ${spec}`);
    }
    overrideSet.add(spec);
  }
  if (overrideSet.size !== seenSpecs.size) {
    fail("Cypress pack override must contain the exact fixed beta core pack; under-scoped or extra overrides are rejected.");
  }
  for (const spec of seenSpecs) {
    if (!overrideSet.has(spec)) {
      fail(`Cypress pack override is missing required spec: ${spec}`);
    }
  }
}

const result = {
  version: manifest.version,
  issue: manifest.issue,
  spec_count: specs.length,
  required_categories: [...requiredCategories],
  specs: specs.map((spec) => spec.path),
  manual_packaged_steps: manifest.manual_packaged_steps ?? []
};

const resultJson = `${JSON.stringify(result, null, 2)}\n`;
if (outputPath) {
  const resolvedOutput = path.resolve(repoRoot, outputPath);
  fs.mkdirSync(path.dirname(resolvedOutput), { recursive: true });
  fs.writeFileSync(resolvedOutput, resultJson);
}

console.log(result.specs.join(","));
