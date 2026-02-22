import fs from "node:fs";
import path from "node:path";
import { describe, expect, it } from "vitest";

describe("styles foundation", () => {
  it("defines themed base styles for controls and tables", () => {
    const cssPath = path.resolve(__dirname, "styles.css");
    const css = fs.readFileSync(cssPath, "utf8");

    expect(css).toContain(".cabinet-card input");
    expect(css).toContain(".cabinet-card button");
    expect(css).toContain(".cabinet-card textarea");
    expect(css).toContain(".cabinet-card select");
    expect(css).toContain(".cabinet-card table");
  });
});

