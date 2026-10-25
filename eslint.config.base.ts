import { defineConfig } from "eslint/config";
import tseslint from "typescript-eslint";

// ESLint runs only the type-aware rules Biome lacks. Biome implements these
// rules from the strict-type-checked set, so ESLint turns them off.
const implementedByBiome = [
  "no-base-to-string",
  "no-deprecated",
  "no-floating-promises",
  "no-for-in-array",
  "no-implied-eval",
  "no-misused-promises",
  "no-mixed-enums",
  "no-unnecessary-condition",
  "no-unnecessary-template-expression",
  "no-unnecessary-type-conversion",
  "only-throw-error",
  "prefer-reduce-type-parameter",
  "require-await",
  "restrict-plus-operands",
];

// Each workspace project's ESLint configuration extends this base and names
// its own TypeScript projects.
export const base = defineConfig(tseslint.configs.strictTypeCheckedOnly, {
  linterOptions: { reportUnusedDisableDirectives: "error" },
  rules: Object.fromEntries(
    implementedByBiome.map((rule) => [`@typescript-eslint/${rule}`, "off"]),
  ),
});
