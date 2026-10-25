import { defineConfig } from "@playwright/test";
import configuration from "../playwright.config.ts";

const baseUrl = "http://127.0.0.1:4183/dashboard/";

export default defineConfig({
  ...configuration,
  testDir: "../e2e",
  workers: 1,
  projects: [{ name: "binary", use: { baseURL: baseUrl } }],
  webServer: {
    cwd: `${import.meta.dirname}/..`,
    command: "bash ../../scripts/serve-dashboard-test.sh",
    url: baseUrl,
  },
});
