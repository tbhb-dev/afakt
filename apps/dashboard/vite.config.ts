import react from "@vitejs/plugin-react";
import { defineConfig, mergeConfig } from "vitest/config";
import { shared } from "../../vitest.shared.ts";

export default mergeConfig(
  shared,
  defineConfig({
    plugins: [react()],
    server: { host: "127.0.0.1", port: 5173, strictPort: true },
    preview: { host: "127.0.0.1", port: 4173, strictPort: true },
    test: {
      include: ["src/**/*.test.{ts,tsx}"],
      environment: "jsdom",
      setupFiles: [`${import.meta.dirname}/vitest.setup.ts`],
      coverage: { include: ["src/**/*.{ts,tsx}"] },
    },
  }),
);
