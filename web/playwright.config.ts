import { defineConfig } from "@playwright/test";

export default defineConfig({
  testDir: "./playwright/e2e",
  timeout: 30_000,
  expect: {
    timeout: 10_000,
  },
  use: {
    baseURL: "http://127.0.0.1:8080",
    trace: "retain-on-failure",
    screenshot: "only-on-failure",
    video: "retain-on-failure",
  },
  webServer: {
    command: "npm run build && npm run serve:go",
    url: "http://127.0.0.1:8080/healthz",
    reuseExistingServer: true,
    timeout: 120_000,
  },
  reporter: [["list"], ["html", { outputFolder: "playwright-report", open: "never" }]],
});
