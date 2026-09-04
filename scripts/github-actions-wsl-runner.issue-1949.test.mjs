import assert from "node:assert/strict";
import { existsSync, readFileSync } from "node:fs";
import { join } from "node:path";
import test from "node:test";
import { fileURLToPath } from "node:url";

const repoRoot = fileURLToPath(new URL("../", import.meta.url));
const installerRelative = "scripts/github-actions-runner/install-wsl-runner.ps1";
const docsRelative = "openspec/changes/add-wsl-runner-pool/design.md";

function readRequired(relativePath) {
  const absolutePath = join(repoRoot, relativePath);
  assert.ok(existsSync(absolutePath), `${relativePath} must exist`);
  return readFileSync(absolutePath, "utf8");
}

test("issue #1949 defines an isolated Cabinet WSL runner pool", () => {
  const installer = readRequired(installerRelative);
  const docs = readRequired(docsRelative);

  assert.match(installer, /\[string\]\$DistroName = "cabinet"/);
  assert.match(installer, /\[string\]\$InstallLocation = "C:\\WSL\\cabinet"/);
  assert.match(installer, /\[ValidateRange\(1, 16\)\]\s*\[int\]\$RunnerCount = 1/);
  assert.match(installer, /\[string\]\$LinuxUser = "cabinet"/);
  assert.match(installer, /https:\/\/github\.com\/collectors-tech\/cabinet/);
  assert.match(installer, /cabinet-ci,docker,node22,go,wsl/);
  assert.match(installer, /node_22\.x/);
  assert.match(installer, /go\.mod/);
  assert.match(installer, /go\.dev\/dl\/go\$\{go_version\}\.linux-\$\{go_arch\}\.tar\.gz/);
  assert.match(installer, /"storage-driver": "overlay2"/);
  assert.match(installer, /"max-size": "10m"/);
  assert.match(installer, /function Test-WslRunnerConfigured/);
  assert.match(installer, /cabinet-runner-\{1:D2\}/);
  assert.match(installer, /a supplied GitHub runner registration token can only be used once/);
  assert.match(installer, /Get-GhRunnerRegistrationToken[\s\S]*?-Scope \$RunnerScope[\s\S]*?-RepoUrl \$RepositoryUrl/);
  assert.match(installer, /Runner\.Worker/);
  assert.match(installer, /docker volume prune --all --force/);
  assert.match(installer, /fstrim \//);
  assert.match(installer, /function Register-WslAutostartTask/);
  assert.doesNotMatch(installer, /Environment=.*registration.*token/i);

  assert.match(docs, /-RunnerCount 3/);
  assert.match(docs, /cabinet-02/);
  assert.match(docs, /C:\\WSL\\cabinet-03/);
  assert.match(docs, /wsl --unregister/);
  assert.match(docs, /fstrim/);
});

test("issue #1949 supports selected-repository organisation runner groups", () => {
  const installer = readRequired(installerRelative);
  const docs = readRequired(docsRelative);
  const spec = readRequired("openspec/changes/add-wsl-runner-pool/specs/wsl-runner-pool/spec.md");

  assert.match(installer, /\[ValidateSet\("Repository", "Organization"\)\]/);
  assert.match(installer, /\[string\]\$RunnerScope = "Repository"/);
  assert.match(installer, /\[string\]\$RunnerGroupName = "cabinet-wsl"/);
  assert.match(installer, /\[string\[\]\]\$RunnerGroupRepositories = @\(\)/);
  assert.match(installer, /\[switch\]\$AllowPublicRepositories/);
  assert.match(installer, /orgs\/\$OrganizationName\/actions\/runner-groups/);
  assert.match(installer, /visibility=selected/);
  assert.match(installer, /repositories\/\$\(\$repository\.id\)/);
  assert.match(installer, /orgs\/\$OrganizationName\/actions\/runners\/registration-token/);
  assert.match(installer, /--runnergroup/);
  assert.doesNotMatch(installer, /\$\{repo_url\}/);
  assert.match(installer, /does not silently convert an existing repository-scoped registration/i);
  assert.match(docs, /organisation runner group/i);
  assert.match(docs, /selected repositories/i);
  assert.match(spec, /organisation-scoped runner group/i);
});
