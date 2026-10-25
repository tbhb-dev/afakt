import { defineConfig } from "vitest/config";

// Each workspace project merges these defaults into its own Vitest
// configuration. Coverage fails below 100% in any file and in total.
export const shared = defineConfig({
  test: {
    coverage: {
      provider: "v8",
      exclude: ["**/*.test.{ts,tsx}"],
      reporter: ["text"],
      thresholds: { 100: true, perFile: true },
    },
  },
});
