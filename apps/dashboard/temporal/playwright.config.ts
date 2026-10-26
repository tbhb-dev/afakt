import { defineConfig } from "@playwright/test";
import configuration from "../playwright.config.ts";

export default defineConfig({
  ...configuration,
  testDir: "..",
  testMatch: ["temporal/*.spec.ts", "e2e/browser.spec.ts"],
  workers: 1,
  projects: [{ name: "temporal", use: { baseURL: "http://127.0.0.1:4186" } }],
  webServer: {
    cwd: `${import.meta.dirname}/..`,
    command: "bash ../../scripts/serve-temporal-ui-test.sh",
    url: "http://127.0.0.1:4186/readyz",
    timeout: 120_000,
  },
});
