/* Minimal error-prevention lint pass for the Vue 3 + TS codebase. Formatting is
 * left to the editor; deeper type-aware rules are a follow-up (docs/adr/0002). */
module.exports = {
  root: true,
  env: { browser: true, es2022: true, node: true },
  parser: 'vue-eslint-parser',
  parserOptions: {
    parser: '@typescript-eslint/parser',
    ecmaVersion: 2022,
    sourceType: 'module'
  },
  extends: ['eslint:recommended', 'plugin:vue/vue3-essential'],
  rules: {
    'no-unused-vars': 'off',
    'no-undef': 'off',
    // Route views are legitimately single-word (Login, Reader, Register).
    'vue/multi-word-component-names': 'off'
  },
  overrides: [
    {
      files: ['*.ts', '*.tsx'],
      parser: '@typescript-eslint/parser'
    }
  ],
  ignorePatterns: ['dist/', 'node_modules/', '*.d.ts']
}
