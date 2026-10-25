import { defineConfig } from "eslint/config";
import { base } from "../../eslint.config.base.ts";

export default defineConfig(
  { ignores: ["coverage/", "dist/", ".stryker-tmp/"] },
  base,
  {
    languageOptions: {
      parserOptions: {
        project: ["./tsconfig.json", "./tsconfig.node.json"],
        tsconfigRootDir: import.meta.dirname,
      },
    },
  },
);
