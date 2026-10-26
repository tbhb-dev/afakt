import { defineConfig } from "@playwright/test";
import configuration from "../playwright.config.ts";

export default defineConfig({
  ...configuration,
  testDir: "..",
  testMatch: ["all/*.spec.ts", "e2e/browser.spec.ts"],
  workers: 1,
  projects: [{ name: "all", use: { baseURL: "http://127.0.0.1:8080" } }],
  webServer: [],
});
