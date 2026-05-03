import assert from "node:assert/strict";
import { describe, it } from "node:test";
import { readFileSync } from "node:fs";
import { join } from "node:path";

const root = new URL("..", import.meta.url).pathname.replace(/^\/([A-Za-z]:)/, "$1");

function readRepoFile(path) {
  return readFileSync(join(root, path), "utf8");
}

describe("Cabinet console output standard", () => {
  it("provides a shared PowerShell console helper", () => {
    const helper = readRepoFile("scripts/lib/cabinet-console.ps1");

    for (const functionName of [
      "Write-CabinetBanner",
      "Write-CabinetSection",
      "Write-CabinetStatus",
      "Write-CabinetKeyValue",
      "Write-CabinetHint",
    ]) {
      assert.match(helper, new RegExp(`function\\s+${functionName}\\b`));
    }

    assert.match(helper, /NO_COLOR/);
    assert.match(helper, /CI/);
  });

  it("documents the terminal output style guide", () => {
    const docs = readRepoFile("docs/CONSOLE-OUTPUT-STANDARD.md");

    assert.match(docs, /Cabinet Console Output Standard/);
    assert.match(docs, /Status states/);
    assert.match(docs, /NO_COLOR/);
    assert.match(docs, /CI/);
  });

  it("uses the shared helper in high-traffic PowerShell scripts", () => {
    for (const script of [
      "scripts/build-cabinet.ps1",
      "scripts/build-ui-static.ps1",
      "scripts/validate-openapi.ps1",
      "scripts/runtime/start-demo2.ps1",
    ]) {
      const contents = readRepoFile(script);

      assert.match(
        contents,
        /\. .+cabinet-console\.ps1/,
        `${script} should dot-source the shared helper`
      );
      assert.match(
        contents,
        /Write-Cabinet(Banner|Section|Status|KeyValue|Hint)/,
        `${script} should use shared console output functions`
      );
    }
  });
});
