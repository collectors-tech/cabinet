import { defineConfig } from "cypress";

export default defineConfig({
  e2e: {
    baseUrl: "http://127.0.0.1:17880",
    specPattern: "cypress/e2e/**/*.cy.ts",
    supportFile: false,
  },
  video: true,
  screenshotsFolder: "cypress/artifacts/screenshots",
  videosFolder: "cypress/artifacts/videos",
});
