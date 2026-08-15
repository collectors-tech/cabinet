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
    assert.equal(payload.version, "0.1.0-beta.9");
    assert.match(payload.channel, /^private-beta$/);
  });

  it("builds only the truthful Windows portable package and checksum", () => {
    const script = readRepoFile("scripts/package-installers.ps1");

    assert.match(script, /cabinet-beta-version\.json/);
    assert.match(script, /\$ExpectedCommit/);
    assert.match(script, /git.+status.+--porcelain/s);
    assert.match(script, /Expected commit.+checked out/s);
    assert.match(script, /windows-amd64-portable\.zip/);
    assert.match(script, /cabinet-mcp\.exe/);
    assert.match(script, /\.\/cmd\/cabinet-mcp/);
    assert.doesNotMatch(script, /macos-(amd64|arm64)\.zip/);
    assert.match(script, /Get-FileHash.+SHA256/s);
    assert.match(script, /\.sha256/);
    assert.match(script, /portable package/i);
    assert.match(script, /cabinet-release-manifest\.json/);
    assert.match(script, /publication_state/);
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
    assert.match(workflow, /commit_sha:/);
    assert.match(workflow, /ref: \${{ inputs\.commit_sha }}/);
    assert.match(workflow, /git status --porcelain/);
    assert.match(workflow, /Build Windows portable package/);
    assert.match(workflow, /verify-cabinet-release-package\.mjs/);
    assert.match(workflow, /cabinet-\*-windows-amd64-portable\.zip/);
    assert.match(workflow, /cabinet-\*-windows-amd64-portable\.zip\.sha256/);
    assert.match(workflow, /cabinet-\*-release-notes\.md/);
    assert.match(workflow, /cabinet-release-manifest\.json/);
    assert.doesNotMatch(workflow, /push:\s*\n\s*tags:/);
    assert.doesNotMatch(workflow, /Build installers/i);
    assert.doesNotMatch(workflow, /name:\s*installers/i);
  });

  it("creates one non-publishing exact Cabinet and Browser Companion candidate bundle", () => {
    const workflow = readRepoFile(".github/workflows/beta-release-candidate.yml");

    assert.match(workflow, /package-installers\.ps1.+-ExpectedCommit/s);
    assert.match(workflow, /verify-cabinet-release-package\.mjs/);
    assert.match(workflow, /create-beta-candidate-bundle\.mjs/);
    assert.match(workflow, /cabinet-release-manifest\.json/);
    assert.match(workflow, /browser-companion-release-manifest\.json/);
    assert.match(workflow, /beta-candidate-bundle-manifest\.json/);
    assert.match(workflow, /dist\/cabinet\/cabinet-\*-windows-amd64-portable\.zip/);
    assert.match(workflow, /dist\/browser-companion\/cabinet-browser-companion-\*\.zip/);
    assert.doesNotMatch(workflow, /softprops\/action-gh-release|gh release|create-release/i);
  });

  it("builds and verifies a real Windows portable package in develop and main gates", () => {
    for (const path of [".github/workflows/develop-quality-gate.yml", ".github/workflows/main-gate.yml"]) {
      const workflow = readRepoFile(path);
      assert.match(workflow, /Windows portable package verification/);
      assert.match(workflow, /runs-on: windows-latest/);
      assert.match(workflow, /package-installers\.ps1.+-ExpectedCommit/s);
      assert.match(workflow, /verify-cabinet-release-package\.mjs/);
    }
  });

  it("publishes only an explicitly approved successful exact candidate as a prerelease", () => {
    const workflow = readRepoFile(".github/workflows/publish-beta-prerelease.yml");

    assert.match(workflow, /workflow_dispatch:/);
    assert.match(workflow, /commit_sha:/);
    assert.match(workflow, /candidate_run_id:/);
    assert.match(workflow, /approval_comment_id:/);
    assert.match(workflow, /issues\.getComment/);
    assert.match(workflow, /APPROVE CABINET 0\.1 PRIVATE BETA/);
    assert.match(workflow, /author_association/);
    assert.match(workflow, /Beta Release Candidate Gate/);
    assert.match(workflow, /conclusion.+success/s);
    assert.match(workflow, /actions\/download-artifact@v4/);
    assert.match(workflow, /verify-cabinet-release-package\.mjs/);
    assert.match(workflow, /verify-browser-companion-package\.mjs/);
    assert.match(
      workflow,
      /candidate\/dist\/browser-companion\/browser-companion-candidate-summary\.md/,
    );
    assert.match(workflow, /softprops\/action-gh-release@v2/);
    assert.match(workflow, /prerelease:\s*true/);
    assert.match(workflow, /target_commitish: \${{ inputs\.commit_sha }}/);
    assert.doesNotMatch(workflow, /git push|develop.*main|main.*develop/i);
  });

  it("does not create an automatic release after main succeeds", () => {
    const workflow = readRepoFile(".github/workflows/release-on-main-success.yml");
    assert.doesNotMatch(workflow, /softprops\/action-gh-release|gh release|create-release/i);
    assert.match(workflow, /does not publish/i);
  });

  it("keeps the non-publishing artifact validation checklist bound to the beta release gate", () => {
    const checklist = readRepoFile("release/windows-portable-artifact-validation.md");

    assert.match(checklist, /cabinet-0\.1\.0-beta\.9-windows-amd64-portable\.zip/);
    assert.match(checklist, /cabinet-0\.1\.0-beta\.9-windows-amd64-portable\.zip\.sha256/);
    assert.match(checklist, /cabinet-0\.1\.0-beta\.9-release-notes\.md/);
    assert.match(checklist, /WINDOWS-PORTABLE-BETA\.md/);
    assert.match(checklist, /cabinet-mcp\.exe/);
    assert.match(checklist, /\/healthz/);
    assert.match(checklist, /\/api\/runtime/);
    assert.match(checklist, /app_version=0\.1\.0-beta\.9/);
    assert.match(checklist, /build_revision.*source_commit/i);
    assert.match(checklist, /#1864 approval/);
    assert.match(checklist, /must not publish/i);
  });

  it("keeps the install and existing-data upgrade validation checklist bound to the beta release gate", () => {
    const checklist = readRepoFile("release/windows-portable-upgrade-validation.md");

    assert.match(checklist, /cabinet-0\.1\.0-beta\.9-windows-amd64-portable\.zip/);
    assert.match(checklist, /\.sha256/);
    assert.match(checklist, /clean install and start/i);
    assert.match(checklist, /existing data directory upgrade/i);
    assert.match(checklist, /backup before replacing or reusing the existing data directory/i);
    assert.match(checklist, /\/healthz/);
    assert.match(checklist, /\/api\/runtime/);
    assert.match(checklist, /app_version=0\.1\.0-beta\.9/);
    assert.match(checklist, /build_revision.*source_commit/i);
    assert.match(checklist, /inventory item count/i);
    assert.match(checklist, /wishlist item count/i);
    assert.match(checklist, /collection membership count/i);
    assert.match(checklist, /saved filter\/view count/i);
    assert.match(checklist, /rollback instructions/i);
    assert.match(checklist, /#1864 approval/);
    assert.match(checklist, /Do not publish/i);
  });
});
