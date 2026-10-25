import { defineConfig } from "eslint/config";
import { base } from "./eslint.config.base.ts";

// The workspace root's own configuration files.
export default defineConfig({ ignores: ["apps/"] }, base, {
  languageOptions: {
    parserOptions: {
      project: ["./tsconfig.json"],
      tsconfigRootDir: import.meta.dirname,
    },
  },
});
