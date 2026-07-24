import assert from "node:assert/strict";
import { describe, it } from "node:test";
import { existsSync, readFileSync } from "node:fs";
import { join } from "node:path";

const root = new URL("..", import.meta.url).pathname.replace(/^\/([A-Za-z]:)/, "$1");

function readRepoFile(path) {
  return readFileSync(join(root, path), "utf8");
}

describe("Cabinet beta packaging contract", () => {
  it("defines one canonical beta version source", () => {
    const versionPath = join(root, "release", "cabinet-beta-version.json");
    assert.equal(existsSync(versionPath), true);

    const payload = JSON.parse(readFileSync(versionPath, "utf8"));
    assert.equal(payload.version, "0.1.0-beta.1");
    assert.match(payload.channel, /^private-beta$/);
  });

  it("builds only the truthful Windows portable package and checksum", () => {
    const script = readRepoFile("scripts/package-installers.ps1");

    assert.match(script, /cabinet-beta-version\.json/);
    assert.match(script, /windows-amd64-portable\.zip/);
    assert.match(script, /cabinet-mcp\.exe/);
    assert.match(script, /\.\/cmd\/cabinet-mcp/);
    assert.doesNotMatch(script, /macos-(amd64|arm64)\.zip/);
    assert.match(script, /Get-FileHash.+SHA256/s);
    assert.match(script, /\.sha256/);
    assert.match(script, /portable package/i);
    assert.doesNotMatch(script, /Installers packaged under/);
  });

  it("embeds semantic version, commit, and build date in the runtime binary", () => {
    const script = readRepoFile("scripts/package-installers.ps1");

    assert.match(script, /internal\/app\.buildVersion=\$resolvedVersion/);
    assert.match(script, /internal\/app\.buildRevision=\$buildRevision/);
    assert.match(script, /internal\/app\.buildDate=\$buildDate/);
  });

  it("uploads portable package evidence from the release workflow", () => {
    const workflow = readRepoFile(".github/workflows/release-installers.yml");

    assert.match(workflow, /Release Portable Packages/);
    assert.match(workflow, /Build Windows portable package/);
    assert.match(workflow, /cabinet-\*-windows-amd64-portable\.zip/);
    assert.match(workflow, /cabinet-\*-windows-amd64-portable\.zip\.sha256/);
    assert.match(workflow, /cabinet-\*-release-notes\.md/);
    assert.doesNotMatch(workflow, /Build installers/i);
    assert.doesNotMatch(workflow, /name:\s*installers/i);
  });

  it("keeps the non-publishing artifact validation checklist bound to the beta release gate", () => {
    const checklist = readRepoFile("release/windows-portable-artifact-validation.md");

    assert.match(checklist, /cabinet-0\.1\.0-beta\.1-windows-amd64-portable\.zip/);
    assert.match(checklist, /cabinet-0\.1\.0-beta\.1-windows-amd64-portable\.zip\.sha256/);
    assert.match(checklist, /cabinet-0\.1\.0-beta\.1-release-notes\.md/);
    assert.match(checklist, /WINDOWS-PORTABLE-BETA\.md/);
    assert.match(checklist, /cabinet-mcp\.exe/);
    assert.match(checklist, /\/healthz/);
    assert.match(checklist, /\/api\/runtime/);
    assert.match(checklist, /app_version=0\.1\.0-beta\.1/);
    assert.match(checklist, /#1864 approval/);
    assert.match(checklist, /must not publish/i);
  });

  it("keeps the install and existing-data upgrade validation checklist bound to the beta release gate", () => {
    const checklist = readRepoFile("release/windows-portable-upgrade-validation.md");

    assert.match(checklist, /cabinet-0\.1\.0-beta\.1-windows-amd64-portable\.zip/);
    assert.match(checklist, /\.sha256/);
    assert.match(checklist, /clean install and start/i);
    assert.match(checklist, /existing data directory upgrade/i);
    assert.match(checklist, /backup before replacing or reusing the existing data directory/i);
    assert.match(checklist, /\/healthz/);
    assert.match(checklist, /\/api\/runtime/);
    assert.match(checklist, /app_version=0\.1\.0-beta\.1/);
    assert.match(checklist, /inventory item count/i);
    assert.match(checklist, /wishlist item count/i);
    assert.match(checklist, /collection membership count/i);
    assert.match(checklist, /saved filter\/view count/i);
    assert.match(checklist, /rollback instructions/i);
    assert.match(checklist, /#1864 approval/);
    assert.match(checklist, /Do not publish/i);
  });
});
