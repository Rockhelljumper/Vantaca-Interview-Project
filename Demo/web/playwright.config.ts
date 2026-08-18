import { defineConfig, devices } from "@playwright/test";

const webBaseURL = process.env.WEB_BASE_URL ?? "http://localhost:13000";

export default defineConfig({
  testDir: "./tests",
  fullyParallel: false,
  workers: 1,
  timeout: 90_000,
  expect: {
    timeout: 12_000,
  },
  outputDir: "test-results/playwright",
  reporter: [
    ["list"],
    ["html", { outputFolder: "playwright-report", open: "never" }],
  ],
  use: {
    baseURL: webBaseURL,
    trace: "retain-on-failure",
    screenshot: "only-on-failure",
    video: "retain-on-failure",
  },
  projects: [
    {
      name: "api",
      testMatch: /api\/.*\.spec\.ts/,
    },
    {
      name: "ui-chromium",
      testMatch: /ui\/.*\.spec\.ts/,
      use: {
        ...devices["Desktop Chrome"],
      },
    },
  ],
});
