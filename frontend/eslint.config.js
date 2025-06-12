import js from "@eslint/js";
import pluginVue from "eslint-plugin-vue";
import vueTsEslintConfig from "@vue/eslint-config-typescript";
import prettierConfig from "@vue/eslint-config-prettier";

export default [
  js.configs.recommended,
  pluginVue.configs["flat/recommended"],
  vueTsEslintConfig,
  prettierConfig,
  {
    files: ["**/*.vue", "**/*.js", "**/*.jsx", "**/*.cjs", "**/*.mjs", "**/*.ts", "**/*.tsx", "**/*.cts", "**/*.mts"],
    rules: {
      semi: "error",
      "prefer-const": "error"
    }
  }
];
