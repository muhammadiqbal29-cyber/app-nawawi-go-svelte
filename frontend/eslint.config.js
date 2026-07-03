import eslint from '@eslint/js';
import sveltePlugin from 'eslint-plugin-svelte';
import globals from 'globals';
import prettierConfig from 'eslint-config-prettier';

export default [
  eslint.configs.recommended,
  ...sveltePlugin.configs['flat/recommended'],
  prettierConfig,
  {
    languageOptions: {
      globals: {
        ...globals.browser,
        ...globals.node
      }
    },
    rules: {
      'svelte/no-navigation-without-resolve': 'off'
    }
  },
  {
    ignores: [
      'node_modules/',
      'dist/',
      'build/',
      '.svelte-kit/',
      'playwright-report/',
      'test-results/'
    ]
  }
];
