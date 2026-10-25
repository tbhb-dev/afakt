import { defineConfig, devices } from "@playwright/test";

const devUrl = "http://127.0.0.1:5173";
const previewUrl = "http://127.0.0.1:4173";

// Each spec runs against the development server and against a preview of the
// production build.
export default defineConfig({
  testDir: "e2e",
  forbidOnly: true,
  retries: 0,
  reporter: "list",
  use: devices["Desktop Chrome"],
  projects: [
    { name: "dev", use: { baseURL: devUrl } },
    { name: "preview", use: { baseURL: previewUrl } },
  ],
  webServer: [
    { command: "pnpm exec vite", url: devUrl },
    { command: "pnpm exec vite preview", url: previewUrl },
  ],
});
